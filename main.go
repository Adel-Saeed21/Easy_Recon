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
	domain  := flag.String("d", "", "Target domain")
	threads := flag.Int("t", 10, "Number of threads")
	flag.Parse()

	printBanner()

	if *domain == "" {
		fmt.Println("Usage: easyRecon -d domain.com [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -d  string   Target domain")
		fmt.Println("  -t  int      Number of threads (default 10)")
		os.Exit(1)
	}

	if err := run(*domain, *threads); err != nil {
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

func askUser(question string) bool {
	fmt.Println(question)
	fmt.Print("  Your choice [y/n]: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes"
}

func run(domain string, threads int) error {
	green := "\033[32m"
	reset := "\033[0m"

	domain = cleanDomain(domain)
	fmt.Printf("  [*] Target  : %s\n", domain)
	fmt.Printf("  [*] Threads : %d\n\n", threads)

	if err := os.MkdirAll(domain, 0755); err != nil {
		return fmt.Errorf("failed to create output folder: %w", err)
	}
	if err := utils.CheckTools("httpx"); err != nil {
		return err
	}

	// ── Subdomain Enumeration ─────────────────────────────────────────────────
	subTools := buildSubTools(domain)
	if len(subTools) == 0 {
		return fmt.Errorf("no enumeration tools found")
	}

	allSubs := runSubdomainEnum(domain, subTools)
	fmt.Println()

	if len(allSubs) == 0 {
		return fmt.Errorf("no subdomains found")
	}

	allSubs = utils.RemoveDuplicates(allSubs)
	fmt.Printf("  %s[+]%s Total unique subdomains : %d\n\n", green, reset, len(allSubs))

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
	urlTools := buildURLTools(domain)

	if len(urlTools) == 0 {
		fmt.Printf("  [!] No URL collection tools found (waybackurls / waymore), skipping...\n\n")
	} else {
		collect := askUser(
			"  ┌────────────────────────────────────────────────┐\n" +
			"  │  Collect URLs from Wayback Machine?            │\n" +
			"  │  (waybackurls / waymore)                       │\n" +
			"  │                                                │\n" +
			"  │  [y] Yes, collect URLs                         │\n" +
			"  │  [n] No, skip                                  │\n" +
			"  └────────────────────────────────────────────────┘")

		if collect {
			allURLs = runURLCollection(alive, urlTools, threads)
			fmt.Println()

			allURLs = utils.RemoveDuplicates(allURLs)
			fmt.Printf("  %s[+]%s Total unique URLs collected : %d\n\n", green, reset, len(allURLs))

			if err := utils.SaveToFile(domain+"/urls.txt", allURLs); err != nil {
				return fmt.Errorf("failed to save URLs: %w", err)
			}
		}
	}

	// ── Parameter Discovery ───────────────────────────────────────────────────
	if len(allURLs) > 0 {
		discoverParams := askUser(
			"  ┌────────────────────────────────────────────────┐\n" +
			"  │  Run parameter discovery? (arjun)              │\n" +
			"  │                                                │\n" +
			"  │  [y] Yes                                       │\n" +
			"  │  [n] No, skip                                  │\n" +
			"  └────────────────────────────────────────────────┘")

		if discoverParams {
			params := runParamDiscovery(allURLs, domain, threads)
			fmt.Println()

			params = utils.RemoveDuplicates(params)
			fmt.Printf("  %s[+]%s Total unique parameters found : %d\n\n", green, reset, len(params))

			if err := utils.SaveToFile(domain+"/params.txt", params); err != nil {
				return fmt.Errorf("failed to save params: %w", err)
			}
		}
	}

	// ── Cleanup ───────────────────────────────────────────────────────────────
	cleanup := askUser(
		"  ┌────────────────────────────────────────────────┐\n" +
		"  │  Cleanup intermediate files?                   │\n" +
		"  │  (alive.txt & urls.txt & params.txt stay safe) │\n" +
		"  │                                                │\n" +
		"  │  [y] Yes, delete them                          │\n" +
		"  │  [n] No, keep them                             │\n" +
		"  └────────────────────────────────────────────────┘")

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
		fmt.Printf("  %s[✓]%s Cleanup done — alive.txt, urls.txt & params.txt are safe.\n", green, reset)
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