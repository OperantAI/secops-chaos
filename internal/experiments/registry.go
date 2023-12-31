package experiments

// ExperimentsRegistry is a list of all experiments
var ExperimentsRegistry = []Experiment{
	&PrivilegedContainerExperimentConfig{},
	&HostPathMountExperimentConfig{},
	&ClusterAdminBindingExperimentConfig{},
	&ContainerSecretsExperimentConfig{},
	&RemoteExecuteAPIExperimentConfig{},
	&ListK8sSecretsConfig{},
}
