package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"easyRecon/runner"
	"easyRecon/utils"
)

func main() {
	domain := flag.String("d", "", "Target domain")
	flag.Parse()

	printBanner()

	if *domain == "" {
		fmt.Println("Usage: easyRecon -d domain.com")
		os.Exit(1)
	}

	if err := run(*domain); err != nil {
		fmt.Fprintf(os.Stderr, "\n[!] Error: %v\n", err)
		os.Exit(1)
	}
}

func printBanner() {
	cyan := "\033[36m"
	orange := "\033[33m"
	reset := "\033[0m"

	fmt.Println()
	fmt.Printf("%s  ███████╗ █████╗ ███████╗██╗   ██╗%s\n", cyan, reset)
	fmt.Printf("%s  ██╔════╝██╔══██╗██╔════╝╚██╗ ██╔╝%s\n", cyan, reset)
	fmt.Printf("%s  █████╗  ███████║███████╗ ╚████╔╝ %s\n", cyan, reset)
	fmt.Printf("%s  ██╔══╝  ██╔══██║╚════██║  ╚██╔╝  %s\n", cyan, reset)
	fmt.Printf("%s  ███████╗██║  ██║███████║   ██║   %s\n", cyan, reset)
	fmt.Printf("%s  ╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝   %s\n", cyan, reset)
	fmt.Println()
	fmt.Printf("%s  ██████╗ ███████╗ ██████╗ ██████╗ ███╗   ██╗%s\n", orange, reset)
	fmt.Printf("%s  ██╔══██╗██╔════╝██╔════╝██╔═══██╗████╗  ██║%s\n", orange, reset)
	fmt.Printf("%s  ██████╔╝█████╗  ██║     ██║   ██║██╔██╗ ██║%s\n", orange, reset)
	fmt.Printf("%s  ██╔══██╗██╔══╝  ██║     ██║   ██║██║╚██╗██║%s\n", orange, reset)
	fmt.Printf("%s  ██║  ██║███████╗╚██████╗╚██████╔╝██║ ╚████║%s\n", orange, reset)
	fmt.Printf("%s  ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝%s\n", orange, reset)
	fmt.Println()
	fmt.Printf("  %sSubdomain Enumeration & Recon Tool%s\n", "\033[90m", reset)
	fmt.Println()
}

func renderBar(current, total int, toolName string) {
	barWidth := 40
	filled := 0
	if total > 0 {
		filled = (current * barWidth) / total
	}
	percent := 0
	if total > 0 {
		percent = (current * 100) / total
	}

	green := "\033[32m"
	gray := "\033[90m"
	cyan := "\033[36m"
	reset := "\033[0m"

	bar := green + strings.Repeat("█", filled) + gray + strings.Repeat("░", barWidth-filled) + reset

	fmt.Printf("\r  %s  %s[%s]%s %3d%%  ",
		cyan+toolName+reset,
		"\033[90m", bar, "\033[90m]",
		percent,
	)
	fmt.Printf("%-10s", "")
}

func askUser(question string) bool {
	fmt.Println(question)
	fmt.Print("  Your choice [y/n]: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes"
}

func run(domain string) error {
	green := "\033[32m"
	reset := "\033[0m"

	domain = cleanDomain(domain)
	fmt.Printf("  [*] Target : %s\n\n", domain)

	if err := os.MkdirAll(domain, 0755); err != nil {
		return fmt.Errorf("failed to create output folder: %w", err)
	}
	if err := utils.CheckTools("httpx"); err != nil {
		return err
	}

	// ── Subdomain Enumeration ─────────────────────────────────────────────────
	type toolEntry struct {
		name    string
		run     func([]string) ([]string, error)
		outFile string
	}

	var tools []toolEntry
	if utils.IsToolInstalled("subfinder") {
		tools = append(tools, toolEntry{"Subfinder   ", runner.Subfinder{}.Run, domain + "/subfinder.txt"})
	}
	if utils.IsToolInstalled("assetfinder") {
		tools = append(tools, toolEntry{"Assetfinder ", runner.Assetfinder{}.Run, domain + "/assetfinder.txt"})
	}
	if utils.IsToolInstalled("findomain") {
		tools = append(tools, toolEntry{"Findomain   ", runner.Findomain{}.Run, domain + "/findomain.txt"})
	}
	if utils.IsToolInstalled("amass") {
		tools = append(tools, toolEntry{"Amass       ", runner.Amass{}.Run, domain + "/amass.txt"})
	}

	if len(tools) == 0 {
		return fmt.Errorf("no enumeration tools found")
	}

	total := len(tools)
	var allSubs []string

	for i, t := range tools {
		renderBar(i*100/total, 100, t.name)

		subs, err := t.run([]string{domain})
		if err != nil {
			fmt.Printf("\n  [!] %s failed: %v\n", t.name, err)
			continue
		}

		subs = utils.RemoveDuplicates(subs)
		_ = utils.SaveToFile(t.outFile, subs)
		allSubs = append(allSubs, subs...)
	}

	renderBar(100, 100, "Done        ")
	fmt.Println()
	fmt.Println()

	if len(allSubs) == 0 {
		return fmt.Errorf("no subdomains found")
	}

	allSubs = utils.RemoveDuplicates(allSubs)
	fmt.Printf("  %s[+]%s Total unique subdomains : %d\n", green, reset, len(allSubs))
	if err := utils.SaveToFile(domain+"/subdomains.txt", allSubs); err != nil {
		return fmt.Errorf("failed to save subdomains: %w", err)
	}

	// ── Httpx ─────────────────────────────────────────────────────────────────
	fmt.Printf("  [*] Running httpx...\n")
	alive, err := runner.Httpx{}.Run(allSubs)
	if err != nil {
		return fmt.Errorf("httpx failed: %w", err)
	}
	alive = utils.RemoveDuplicates(alive)
	fmt.Printf("  %s[✓]%s Alive domains : %d\n\n", green, reset, len(alive))

	if err := utils.SaveToFile(domain+"/alive.txt", alive); err != nil {
		return fmt.Errorf("failed to save alive domains: %w", err)
	}

	// ── URL Collection ────────────────────────────────────────────────────────
	var allURLs []string
	urlToolsAvailable := utils.IsToolInstalled("waybackurls") || utils.IsToolInstalled("waymore")

	if urlToolsAvailable {
		collectURLs := askUser("  ┌────────────────────────────────────────────────┐\n  │  Collect URLs from Wayback Machine?            │\n  │  (waybackurls / waymore)                       │\n  │                                                │\n  │  [y] Yes, collect URLs                         │\n  │  [n] No, skip                                  │\n  └────────────────────────────────────────────────┘")

		if collectURLs {
			type urlTool struct {
				name    string
				run     func([]string) ([]string, error)
				outFile string
			}

			var urlTools []urlTool
			if utils.IsToolInstalled("waybackurls") {
				urlTools = append(urlTools, urlTool{"WaybackURLs ", runner.WaybackURL{}.Run, domain + "/waybackurls.txt"})
			}
			if utils.IsToolInstalled("waymore") {
				urlTools = append(urlTools, urlTool{"Waymore     ", runner.Waymore{}.Run, domain + "/waymore.txt"})
			}

			totalURL := len(urlTools)
			for i, t := range urlTools {
				renderBar(i*100/totalURL, 100, t.name)

				for _, aliveDomain := range alive {
					urls, err := t.run([]string{aliveDomain})
					if err != nil {
						fmt.Printf("\n  [!] %s failed for %s: %v\n", t.name, aliveDomain, err)
						continue
					}
					allURLs = append(allURLs, urls...)
				}

				_ = utils.SaveToFile(t.outFile, utils.RemoveDuplicates(allURLs))
			}

			renderBar(100, 100, "Done        ")
			fmt.Println()
			fmt.Println()

			allURLs = utils.RemoveDuplicates(allURLs)
			fmt.Printf("  %s[+]%s Total unique URLs collected : %d\n\n", green, reset, len(allURLs))

			if err := utils.SaveToFile(domain+"/urls.txt", allURLs); err != nil {
				return fmt.Errorf("failed to save URLs: %w", err)
			}
		}
	} else {
		fmt.Printf("  [!] No URL collection tools found (waybackurls / waymore), skipping...\n\n")
	}

	cleanup := askUser("  ┌────────────────────────────────────────────────┐\n  │  Cleanup intermediate files?                   │\n  │  (alive.txt & urls.txt will stay safe)         │\n  │                                                │\n  │  [y] Yes, delete them                          │\n  │  [n] No, keep them                             │\n  └────────────────────────────────────────────────┘")

	if cleanup {
		toDelete := []string{
			domain + "/subfinder.txt",
			domain + "/assetfinder.txt",
			domain + "/findomain.txt",
			domain + "/amass.txt",
			domain + "/subdomains.txt",
			domain + "/waybackurls.txt",
			domain + "/waymore.txt",
		}
		for _, f := range toDelete {
			if err := os.Remove(f); err == nil {
				fmt.Printf("  %s[✓]%s Deleted: %s\n", green, reset, f)
			}
		}
		fmt.Printf("  %s[✓]%s Cleanup done — alive.txt & urls.txt are safe.\n", green, reset)
	} else {
		fmt.Printf("  %s[✓]%s Files kept.\n", green, reset)
	}

	fmt.Printf("\n  %s[✓] Recon completed! Results in: ./%s/%s\n\n", green, domain, reset)
	return nil
}

func cleanDomain(domain string) string {
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	return strings.TrimSuffix(domain, "/")
}
