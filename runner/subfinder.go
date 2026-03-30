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
	cmd := exec.Command("subfinder", "-d", domain, "-silent")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var results []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || isLogLine(line) {
			continue
		}
		results = append(results, line)
	}
	return results, nil
}

func isLogLine(line string) bool {
	logPrefixes := []string{"[INF]", "[WRN]", "[ERR]", "[DBG]", "[FTL]"}
	for _, prefix := range logPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}