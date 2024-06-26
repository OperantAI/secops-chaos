package snippets

import (
	"fmt"

	embeded "github.com/operantai/secops-chaos"
)

func GetComponentTemplate(component string) ([]byte, error) {
	content, err := embeded.EmbeddedComponents.ReadFile(fmt.Sprintf("components/%s.yaml", component))
	if err != nil {
		return nil, err
	}
	return content, nil
}

func GetExperimentTemplate(experiment string) ([]byte, error) {
	content, err := embeded.EmbeddedExperiments.ReadFile(fmt.Sprintf("experiments/%s.yaml", experiment))
	if err != nil {
		return nil, err
	}
	return content, nil
}
