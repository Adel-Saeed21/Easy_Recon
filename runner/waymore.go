package runner

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"
)

type Waymore struct{}

func (w Waymore) Name() string {
    return "waymore"
}

func (w Waymore) Run(input []string) ([]string, error) {
    domain := input[0]

    ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
    defer cancel()

    var stdout, stderr bytes.Buffer
    cmd := exec.CommandContext(ctx, "waymore", "-i", domain, "-mode", "U")
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("waymore error: %w — stderr: %s", err, stderr.String())
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