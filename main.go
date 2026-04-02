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
	threads := flag.Int("t", 10, "Number of threads")
	noTelegram := flag.Bool("no-tg", false, "Disable Telegram bot (use terminal only)")
	flag.Parse()

	printBanner()

	if *domain == "" {
		fmt.Println("Usage: easyRecon -d domain.com [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -d      string   Target domain")
		fmt.Println("  -t      int      Number of threads (default 10)")
		fmt.Println("  -no-tg           Disable Telegram bot, use terminal only")
		os.Exit(1)
	}

	// ── Telegram Setup ────────────────────────────────────────────────────────
	var tio *TelegramIO

	if !*noTelegram {
		_, telegramIO, err := SetupBot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [!] Telegram setup failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "  [!] Falling back to terminal mode.\n\n")
		} else {
			tio = telegramIO
			fmt.Println("  [✓] Telegram bot connected — you can now interact from your phone!")
			fmt.Println()
		}
	}

	if err := run(*domain, *threads, tio); err != nil {
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

func askUser(question string, tio *TelegramIO) bool {
	if tio != nil {
		return tio.AskYesNo(question)
	}
	// original terminal logic
	fmt.Println(question)
	fmt.Print("  Your choice [y/n]: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes"
}

func emit(tio *TelegramIO, format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	fmt.Print(line)
	if tio != nil && strings.TrimSpace(stripANSI(line)) != "" {
		tio.bot.OutputCh <- stripANSI(line)
	}
}

func stripANSI(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++
		} else {
			out.WriteByte(s[i])
			i++
		}
	}
	return out.String()
}

func run(domain string, threads int, tio *TelegramIO) error {
	green := "\033[32m"
	reset := "\033[0m"

	domain = cleanDomain(domain)
	emit(tio, "  [*] Target  : %s\n", domain)
	emit(tio, "  [*] Threads : %d\n\n", threads)

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
	emit(tio, "\n")

	if len(allSubs) == 0 {
		return fmt.Errorf("no subdomains found")
	}

	allSubs = utils.RemoveDuplicates(allSubs)
	emit(tio, "  %s[+]%s Total unique subdomains : %d\n\n", green, reset, len(allSubs))

	if err := utils.SaveToFile(domain+"/subdomains.txt", allSubs); err != nil {
		return fmt.Errorf("failed to save subdomains: %w", err)
	}

	// ── Httpx ─────────────────────────────────────────────────────────────────
	emit(tio, "  [*] Running httpx...\n")
	alive, err := runner.Httpx{}.Run(allSubs)
	if err != nil {
		return fmt.Errorf("httpx failed: %w", err)
	}
	alive = utils.RemoveDuplicates(alive)
	emit(tio, "  %s[✓]%s Alive domains : %d\n\n", green, reset, len(alive))

	if err := utils.SaveToFile(domain+"/alive.txt", alive); err != nil {
		return fmt.Errorf("failed to save alive domains: %w", err)
	}

	// ── URL Collection ────────────────────────────────────────────────────────
	var allURLs []string
	urlTools := buildURLTools(domain)

	if len(urlTools) == 0 {
		emit(tio, "  [!] No URL collection tools found (waybackurls / waymore), skipping...\n\n")
	} else {
		collect := askUser(
			"  ┌────────────────────────────────────────────────┐\n"+
				"  │  Collect URLs from Wayback Machine?            │\n"+
				"  │  (waybackurls / waymore)                       │\n"+
				"  │                                                │\n"+
				"  │  [y] Yes, collect URLs                         │\n"+
				"  │  [n] No, skip                                  │\n"+
				"  └────────────────────────────────────────────────┘",
			tio,
		)

		if collect {
			allURLs = runURLCollection(alive, urlTools, threads)
			emit(tio, "\n")

			allURLs = utils.RemoveDuplicates(allURLs)
			emit(tio, "  %s[+]%s Total unique URLs collected : %d\n\n", green, reset, len(allURLs))

			if err := utils.SaveToFile(domain+"/urls.txt", allURLs); err != nil {
				return fmt.Errorf("failed to save URLs: %w", err)
			}
		}
	}

	if len(allURLs) > 0 {
		discoverParams := askUser(
			"\n"+
				"    Run parameter discovery?                      \n"+
				"                                                  \n"+
				"    [y] Yes                                      \n"+
				"    [n] No, skip                                  \n",
			tio,
		)

		if discoverParams {
			params := runParamDiscovery(allURLs, domain, threads)
			emit(tio, "\n")

			params = utils.RemoveDuplicates(params)
			emit(tio, "  %s[+]%s Total unique parameters found : %d\n\n", green, reset, len(params))

			if err := utils.SaveToFile(domain+"/params.txt", params); err != nil {
				return fmt.Errorf("failed to save params: %w", err)
			}
		}
	}

	cleanup := askUser(
		"  ┌────────────────────────────────────────────────┐\n"+
			"  │  Cleanup intermediate files?                   │\n"+
			"  │  (alive.txt & urls.txt & params.txt stay safe) │\n"+
			"  │                                                │\n"+
			"  │  [y] Yes, delete them                          │\n"+
			"  │  [n] No, keep them                             │\n"+
			"  └────────────────────────────────────────────────┘",
		tio,
	)

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
				emit(tio, "  %s[✓]%s Deleted: %s\n", green, reset, f)
			}
		}
		emit(tio, "  %s[✓]%s Cleanup done — only alive.txt kept.\n", green, reset)
	} else {
		emit(tio, "  %s[✓]%s Files kept.\n", green, reset)
	}

	emit(tio, "\n  %s[✓] Recon completed! Results in: ./%s/%s\n\n", green, domain, reset)
	return nil
}

func cleanDomain(domain string) string {
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	return strings.TrimSuffix(domain, "/")
}
