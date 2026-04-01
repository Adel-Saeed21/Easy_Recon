package runner

import (
    "bytes"
    "os/exec"
    "strings"
)

type Httpx struct{}

func (h Httpx) Name() string {
    return "httpx"
}

func (h Httpx) Run(input []string) ([]string, error) {
    cmd := exec.Command("httpx", "-silent")

    cmd.Stdin = strings.NewReader(strings.Join(input, "\n"))

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return nil, err
    }

    lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
    var results []string
    for _, line := range lines {
        if line = strings.TrimSpace(line); line != "" {
            results = append(results, line)
        }
    }
    return results, nil
}