package runner

import (
	"os/exec"
	"strings"
)

type Httpx struct{}

func (h Httpx) Name() string {
	return "httpx"
}

func (h Httpx) Run(input []string) ([]string, error) {
	cmd := exec.Command("httpx", "-silent")

	stdin, _ := cmd.StdinPipe()

	go func() {
		defer stdin.Close()
		for _, sub := range input {
			stdin.Write([]byte(sub + "\n"))
		}
	}()

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