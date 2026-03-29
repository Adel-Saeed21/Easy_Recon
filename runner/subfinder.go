package runner

import (
	"os/exec"
	"strings"
)

type Subfinder struct{}

func (s Subfinder) Name() string {
	return "subfinder"
}

func (s Subfinder) Run(input []string) ([]string, error) {
	domain := input[0]

	cmd := exec.Command("subfinder", "-d", domain)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")

	var results []string
	for _, line := range lines {
		if line != "" {
			results = append(results, line)
		}
	}

	return results, nil
}