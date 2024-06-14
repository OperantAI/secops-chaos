package experiments

import "fmt"

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
		return &HostPathMountExperiment{config}, nil
	case "list-kubernetes-secrets":
		return &ListK8sSecrets{config}, nil
	case "remote-execute-api":
		return &RemoteExecuteAPIExperiment{config}, nil
	case "execute-api":
		return &ExecuteAPIExperiment{config}, nil
	case "credential-access-container-secrets":
		return &ContainerSecretsExperiment{config}, nil
	case "cluster-admin-binding":
		return &ClusterAdminBindingExperiment{config}, nil
	case "privileged-container":
		return &PrivilegedContainerExperiment{config}, nil
	default:
		return nil, fmt.Errorf("Uknown experiment type: %w", config.Metadata.Type)
	}
}
