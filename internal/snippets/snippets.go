package snippets

import (
	"fmt"

	embedComponents "github.com/operantai/secops-chaos/components"
	embedExperiments "github.com/operantai/secops-chaos/experiments"
)

func GetComponentTemplate(component string) ([]byte, error) {
	content, err := embedComponents.EmbeddedComponents.ReadFile(fmt.Sprintf("%s.yaml", component))
	if err != nil {
		return nil, err
	}
	return content, nil
}

func GetExperimentTemplate(experiment string) ([]byte, error) {
	content, err := embedExperiments.EmbeddedExperiments.ReadFile(fmt.Sprintf("%s.yaml", experiment))
	if err != nil {
		return nil, err
	}
	return content, nil
}
