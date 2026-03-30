package runner
import(
	"os/exec"
	"strings"
)

type Findomain struct{}

func (f Findomain) Name()string{
	return "findomain"
}

func (f Findomain) Run(input []string) ([]string,error){
	domain :=input[0]
	cmd:= exec.Command("findomain","-t",domain)

	output,err := cmd.CombinedOutput()

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