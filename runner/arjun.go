package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Arjun struct{}

func (a Arjun) Name() string {
    return "arjun"
}

func (a Arjun) Run(input []string) ([]string, error) {
    url := input[0]

    outFile := strings.ReplaceAll(url, "/", "_") + "_arjun.json"

    ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
    defer cancel()

    var stderr bytes.Buffer
    cmd := exec.CommandContext(ctx, "arjun",
        "-u", url,
        "-oJ", outFile,  
        "-q",            
    )
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("arjun error: %w — stderr: %s", err, stderr.String())
    }

    // نقرأ الـ JSON
    results, err := parseArjunOutput(outFile)
    if err != nil {
        return nil, err
    }

    return results, nil
}

type arjunResult struct {
    URL    string   `json:"url"`
    Params []string `json:"params"`
}

func parseArjunOutput(path string) ([]string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read arjun output: %w", err)
    }
    defer os.Remove(path) 

    var result arjunResult
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, fmt.Errorf("failed to parse arjun output: %w", err)
    }

    var results []string
    for _, param := range result.Params {
        results = append(results, fmt.Sprintf("%s?%s=", result.URL, param))
    }

    return results, nil
}