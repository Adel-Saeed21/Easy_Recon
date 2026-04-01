package main

import (
	"fmt"
	"sync"

	"easyRecon/runner"
	"easyRecon/utils"
)

func runParamDiscovery(urls []string, domain string, threads int) []string {
	if !utils.IsToolInstalled("arjun") {
		fmt.Printf("  [!] arjun not found, skipping...\n\n")
		return nil
	}

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		allParams []string
	)

	sem := make(chan struct{}, threads)

	fmt.Printf("  [*] Running arjun on %d URLs | threads: %d\n\n", len(urls), threads)

	for _, url := range urls {
		wg.Add(1)
		sem <- struct{}{}
		go func(u string) {
			defer wg.Done()
			defer func() { <-sem }()

			params, err := runner.Arjun{}.Run([]string{u})
			if err != nil {
				fmt.Printf("  [!] arjun failed for %s: %v\n", u, err)
				return
			}
			if len(params) == 0 {
				return
			}

			mu.Lock()
			allParams = append(allParams, params...)
			fmt.Printf("  [✓] Found %d params in: %s\n", len(params), u)
			mu.Unlock()
		}(url)
	}

	wg.Wait()
	return allParams
}
