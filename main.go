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
	cyan   := "\033[36m"
	orange := "\033[33m"
	reset  := "\033[0m"

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

	green  := "\033[32m"
	gray   := "\033[90m"
	cyan   := "\033[36m"
	reset  := "\033[0m"

	bar := green + strings.Repeat("█", filled) + gray + strings.Repeat("░", barWidth-filled) + reset

	fmt.Printf("\r  %s  %s[%s]%s %3d%%  ",
		cyan + toolName + reset,
		"\033[90m", bar, "\033[90m]",
		percent,
	)
	fmt.Printf("%-10s", "")
}

func run(domain string) error {
	domain = cleanDomain(domain)
	fmt.Printf("  [*] Target : %s\n\n", domain)

	if err := os.MkdirAll(domain, 0755); err != nil {
		return fmt.Errorf("failed to create output folder: %w", err)
	}
	if err := utils.CheckTools("httpx"); err != nil {
		return err
	}

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

	total    := len(tools)
	var allSubs []string
	green  := "\033[32m"
	reset  := "\033[0m"

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

	// bar 100%
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

	// ── httpx ─────────────────────────────────────────────────────────────────
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

	fmt.Println("  ┌────────────────────────────────────────────────┐")
	fmt.Println("  │  Keep per-tool files?                          │")
	fmt.Println("  │  subfinder.txt / assetfinder.txt / ...         │")
	fmt.Println("  │                                                │")
	fmt.Println("  │  [y] Yes, keep them                            │")
	fmt.Println("  │  [n] No, delete  (alive.txt stays safe)        │")
	fmt.Println("  └────────────────────────────────────────────────┘")
	fmt.Print("  Your choice [y/n]: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))

	if answer == "n" || answer == "no" {
		toDelete := []string{
			domain + "/subfinder.txt",
			domain + "/assetfinder.txt",
			domain + "/findomain.txt",
			domain + "/amass.txt",
			domain + "/subdomains.txt",
		}
		for _, f := range toDelete {
			if err := os.Remove(f); err == nil {
				fmt.Printf("  %s[✓]%s Deleted: %s\n", green, reset, f)
			}
		}
		fmt.Printf("  %s[✓]%s Cleanup done — alive.txt is safe.\n", green, reset)
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