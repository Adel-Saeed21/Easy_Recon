package runner

import(
	"os/exec"
	"strings"
)

type Amass struct{}

func (a Amass) Name() string{
	return "amass"
}

func (a Amass) Run(input []string)([] string,error){
	domain:=input[0]
	cmd:= exec.Command("amass","enum", "-d",domain)

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