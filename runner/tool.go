package runner

type Tool interface {
	Name() string
	Run(input []string) ([]string, error)
}