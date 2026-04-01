package main

import (
	"fmt"
	"sync"

	"easyRecon/runner"
	"easyRecon/utils"
)

type urlTool struct {
	name    string
	run     func([]string) ([]string, error)
	outFile string
}

func buildURLTools(domain string) []urlTool {
	var tools []urlTool
	if utils.IsToolInstalled("waybackurls") {
		tools = append(tools, urlTool{"WaybackURLs ", runner.WaybackURL{}.Run, domain + "/waybackurls.txt"})
	}
	if utils.IsToolInstalled("waymore") {
		tools = append(tools, urlTool{"Waymore     ", runner.Waymore{}.Run, domain + "/waymore.txt"})
	}
	return tools
}

func runURLCollection(alive []string, tools []urlTool, threads int) []string {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allURLs []string
	)

	sem := make(chan struct{}, threads)

	fmt.Printf("  [*] Running %d URL tools × %d domains | threads: %d\n\n", len(tools), len(alive), threads)

	for _, t := range tools {
		wg.Add(1)
		sem <- struct{}{}
		go func(t urlTool) {
			defer wg.Done()
			defer func() { <-sem }()

			var toolURLs []string
			var domainWg sync.WaitGroup
			var domainMu sync.Mutex

			for _, aliveDomain := range alive {
				domainWg.Add(1)
				go func(d string) {
					defer domainWg.Done()
					urls, err := t.run([]string{d})
					if err != nil {
						fmt.Printf("  [!] %s failed for %s: %v\n", t.name, d, err)
						return
					}
					domainMu.Lock()
					toolURLs = append(toolURLs, urls...)
					domainMu.Unlock()
				}(aliveDomain)
			}

			domainWg.Wait()
			toolURLs = utils.RemoveDuplicates(toolURLs)
			_ = utils.SaveToFile(t.outFile, toolURLs)

			mu.Lock()
			allURLs = append(allURLs, toolURLs...)
			fmt.Printf("  [✓] %-14s found: %d URLs\n", t.name, len(toolURLs))
			mu.Unlock()
		}(t)
	}

	wg.Wait()
	return allURLs
}