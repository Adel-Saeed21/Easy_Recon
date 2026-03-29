package main

import (
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

    if *domain == "" {
        fmt.Println("Usage: easyRecon -d domain.com")
        os.Exit(1)
    }

    if err := run(*domain); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}



func run(domain string) error {
    domain = cleanDomain(domain)
    fmt.Printf("[*] Starting recon on: %s\n", domain)

    if err := os.MkdirAll(domain, 0755); err != nil {
        return fmt.Errorf("failed to create output folder: %w", err)
    }

    if err := utils.CheckTools("subfinder", "httpx"); err != nil {
        return err
    }

    // Run subfinder
    fmt.Println("[*] Running subfinder...")
    subs, err := runner.Subfinder{}.Run([]string{domain})
    if err != nil {
        return fmt.Errorf("subfinder failed: %w", err)
    }
    subs = utils.RemoveDuplicates(subs)
    fmt.Printf("[+] Found %d unique subdomains\n", len(subs))

    if err := utils.SaveToFile(domain+"/subdomains.txt", subs); err != nil {
        return fmt.Errorf("failed to save subdomains: %w", err)
    }

    // Run httpx
    fmt.Println("[*] Running httpx...")
    alive, err := runner.Httpx{}.Run(subs)
    if err != nil {
        return fmt.Errorf("httpx failed: %w", err)
    }
    fmt.Printf("[+] Found %d alive domains\n", len(alive))

    if err := utils.SaveToFile(domain+"/alive.txt", alive); err != nil {
        return fmt.Errorf("failed to save alive domains: %w", err)
    }

    fmt.Printf("[✓] Recon completed! Results saved in: %s/\n", domain)
    return nil
}


func cleanDomain(domain string) string {
    domain = strings.TrimPrefix(domain, "https://")
    domain = strings.TrimPrefix(domain, "http://")
    return strings.TrimSuffix(domain, "/")
}