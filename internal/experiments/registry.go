package experiments

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// ExperimentsRegistry is a list of all experiments
var ExperimentsRegistry = []Experiment{
	&PrivilegedContainerExperiment{},
	&HostPathMountExperiment{},
	&ClusterAdminBindingExperiment{},
	&ContainerSecretsExperiment{},
	&RemoteExecuteAPIExperiment{},
	&ExecuteAPIExperiment{},
	&ListK8sSecrets{},
	&LLMDataLeakageExperiment{},
	&LLMDataPoisoningExperiment{},
	&KubeExec{},
}

func ListExperiments() map[string]string {
	experiments := make(map[string]string)
	for _, experiment := range ExperimentsRegistry {
		experiments[experiment.Type()] = experiment.Description()
	}
	return experiments
}

func ExperimentFactory(config *ExperimentConfig) (Experiment, error) {
	switch config.Metadata.Type {
	case "host-path-mount":
		exp := &HostPathMountExperiment{Metadata: config.Metadata}
		err := mapstructure.Decode(config.Parameters, &exp.Parameters)
		if err != nil {
			return nil, fmt.Errorf("Error decoding parameters for experiment %s", config.Metadata.Name)
		}
		return exp, nil
	case "list-kubernetes-secrets":
		exp := &ListK8sSecrets{Metadata: config.Metadata}
		err := mapstructure.Decode(config.Parameters, &exp.Parameters)
		if err != nil {
			return nil, fmt.Errorf("Error decoding parameters for experiment %s", config.Metadata.Name)
		}
		return exp, nil
	case "remote-execute-api":
		exp := &RemoteExecuteAPIExperiment{Metadata: config.Metadata}
		err := mapstructure.Decode(config.Parameters, &exp.Parameters)
		if err != nil {
			return nil, fmt.Errorf("Error decoding parameters for experiment %s", config.Metadata.Name)
		}
		return exp, nil
	case "execute-api":
		exp := &ExecuteAPIExperiment{Metadata: config.Metadata}
		err := mapstructure.Decode(config.Parameters, &exp.Parameters)
		if err != nil {
			return nil, fmt.Errorf("Error decoding parameters for experiment %s", config.Metadata.Name)
		}
		return exp, nil
	case "credential-access-container-secrets":
		exp := &ContainerSecretsExperiment{Metadata: config.Metadata}
		err := mapstructure.Decode(config.Parameters, &exp.Parameters)
		if err != nil {
			return nil, fmt.Errorf("Error decoding parameters for experiment %s", config.Metadata.Name)
		}
		return exp, nil
	case "cluster-admin-binding":
		exp := &ClusterAdminBindingExperiment{Metadata: config.Metadata}
		err := mapstructure.Decode(config.Parameters, &exp.Parameters)
		if err != nil {
			return nil, fmt.Errorf("Error decoding parameters for experiment %s", config.Metadata.Name)
		}
		return exp, nil
	case "privileged-container":
		exp := &PrivilegedContainerExperiment{Metadata: config.Metadata}
		err := mapstructure.Decode(config.Parameters, &exp.Parameters)
		if err != nil {
			return nil, fmt.Errorf("Error decoding parameters for experiment %s", config.Metadata.Name)
		}
		return exp, nil
	default:
		return nil, fmt.Errorf("Uknown experiment type: %s", config.Metadata.Type)
	}
}
