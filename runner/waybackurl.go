package runner

import (
    "os/exec"
    "strings"
)

type WaybackURL struct{}

func (w WaybackURL) Name() string {
    return "waybackurl"
}

func (w WaybackURL) Run(input []string) ([]string, error) {
    domain := input[0]

    cmd := exec.Command("waybackurls", domain)

    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    lines := strings.Split(strings.TrimSpace(string(output)), "\n")

    var results []string
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" {
            results = append(results, line)
        }
    }

    return results, nil
}