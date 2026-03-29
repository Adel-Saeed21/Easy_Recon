package utils

import "os/exec"
import "fmt"

func IsToolInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func CheckTools(tools ...string) error {
    for _, tool := range tools {
        if !IsToolInstalled(tool) {
            return fmt.Errorf("required tool not found: %s", tool)
        }
    }
    return nil
}