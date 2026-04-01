package main

import (
	"fmt"
	"sync"

	"easyRecon/runner"
	"easyRecon/utils"
)

type toolEntry struct {
	name    string
	run     func([]string) ([]string, error)
	outFile string
}

func buildSubTools(domain string) []toolEntry {
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
	return tools
}

func runSubdomainEnum(domain string, tools []toolEntry) []string {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allSubs []string
	)

	fmt.Printf("  [*] Running %d tools concurrently...\n\n", len(tools))

	for _, t := range tools {
		wg.Add(1)
		go func(t toolEntry) {
			defer wg.Done()

			subs, err := t.run([]string{domain})
			if err != nil {
				fmt.Printf("  [!] %s failed: %v\n", t.name, err)
				return
			}

			subs = utils.RemoveDuplicates(subs)
			_ = utils.SaveToFile(t.outFile, subs)

			mu.Lock()
			allSubs = append(allSubs, subs...)
			fmt.Printf("  [✓] %-14s found: %d\n", t.name, len(subs))
			mu.Unlock()
		}(t)
	}

	wg.Wait()
	return allSubs
}