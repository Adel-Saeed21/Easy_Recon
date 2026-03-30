package runner

import (
	"os/exec"
	"strings"
)

type Assetfinder struct{}

func (a Assetfinder) Name() string {
	return "assetfinder"
}

func (a Assetfinder) Run(input []string) ([]string, error) {
	domain := input[0]

	cmd := exec.Command("assetfinder", "--subs-only", domain)

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
