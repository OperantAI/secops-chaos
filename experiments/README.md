# Experiments

**secops-chaos** experiments are driven by **experiment** files, they all follow a common format, made up of a `metadata` and `parameters` section.

``` yaml
experiments:
  - metadata:
      name: my-experiment # A meaningful name for your experiment
      type: prvileged-container # The type of experiment, see table below for a list of valid types
      namespace: my-namespace # What namespace to apply the experiment to
    parameters: # Parameters holds the settings for your experiment, tweak them to suit your needs.
        host-pid: true 
```


## Available Experiments

| Type                                                  | Description                                                                                                                | Framework |
|-------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------|-----------|
<<<<<<< Updated upstream
| [privileged_container](run_privileged_container.yaml) | This experiment attempts to run a privileged container in a namespace                                                      | MITRE     |
| [host_path_mount](host_path_volume.yaml)              | This experiment attempts to mount a sensitive host filesystem path into a container                                        | MITRE     |
| [cluster_admin_binding](cluster_admin_binding.yaml)   | This experiment attempts to create a container with the cluster-admin role binding attached                                | MITRE     |
| [remote_execute_api](remote_execute_api.yaml)         | This experiment attempts to create a deployment with a configurable image and verifies based off of API calls to the image | MITRE     |
| [execute_api](execute_api.yaml)                       | This experiment attempts to call a service with a payload                                                                  | MITRE     |
=======
| [privileged-container](run-privileged-container.yaml) | This experiment attempts to run a privileged container in a namespace                                                      | MITRE     |
| [host-path-mount](host-path-volume.yaml)              | This experiment attempts to mount a sensitive host filesystem path into a container                                        | MITRE     |
| [cluster-admin-binding](cluster-admin-binding.yaml)   | This experiment attempts to create a container with the cluster-admin role binding attached                                | MITRE     |
| [remote-execute-api](remote-execute-api.yaml)         | This experiment attempts to create a deployment with a configurable image and verifies based off of API calls to the image | MITRE     |
>>>>>>> Stashed changes

## Implementing a new Experiment

Each experiment within `secops-chaos` adheres to a shared interface, this allows for a common set of functionality to be used across all experiments.

When implementing a new experiment you should create a new file starting with `experiment-` within the [internal/experiments](https://github.com/OperantAI/secops-chaos/blob/main/internal/experiments/) directory, and implement the `Experiment` interface.

```go
type Experiment interface {
	// Type returns the type of the experiment
	Type() string
	// Description describes the experiment in a brief sentence
	Description() string
	// Framework returns the attack framework e.g., MITRE/OWASP
	Framework() string
	// Tactic returns the attack tactic category
	Tactic() string
	// Technique returns the attack method
	Technique() string
	// Run runs the experiment, returning an error if it fails
	Run(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error
	// Verify verifies the experiment, returning an error if it fails
	Verify(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) (*verifier.Outcome, error)
	// Cleanup cleans up the experiment, returning an error if it fails
	Cleanup(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error
}
```

Additionally, you need some way of configuring your experiment, this is done via the `ExperimentConfig` struct, which is passed to the `Run`, `Verify` and `Cleanup` methods for your experiment.

```go
type ExperimentsConfig struct {
	ExperimentConfigs []ExperimentConfig `yaml:"experiments"`
}

type ExperimentConfig struct {
    // Metadata for the experiment
	Metadata ExperimentMetadata `yaml:"metadata"`
	// Parameters for the experiment
	Parameters interface{} `yaml:"parameters"`
}

type ExperimentMetadata struct {
	// Name of the experiment
	Name string `yaml:"name"`
	// Namespace to apply the experiment to
	Namespace string `yaml:"namespace"`
	// Type of the experiment
	Type string `yaml:"type"`
}
```

Inside your experiment file, you also need to add Parameter configuration as parameters will differ per experiment. These params get parsed into the `Parameters` field of the `ExperimentConfig` struct.

``` go
type HostPathMountExperimentConfig struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters HostPathMount      `yaml:"parameters"`
}

type HostPathMount struct {
	HostPath HostPath `yaml:"host-path"`
}

type HostPath struct {
	Path string `yaml:"path"`
}

func (p *HostPathMountExperimentConfig) Run(ctx context.Context, client *k8s.client, experimentConfig *ExperimentConfig) error {
	var hostPathMountExperimentConfig HostPathMountExperimentConfig
	yamlObj, - := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &hostPathMountExperimentConfig)
	if err != nil {
		return err
	}
	params := hostPathMountExperimentConfig.Parameters
    ...rest of your experiment implementation...
}
```

`Verify`, and `Clean` also follow the same pattern, and are passed the same `ExperimentConfig` struct.

Finally, add your Experiment to the Experiment registry in [internal/experiments/registry.go](https://github.com/operantai/secops-chaos/blob/main/internal/experiments/registry.go), and create a new experiment file in the [experiments](https://github.com/operantai/secops-chaos/blob/main/experiments) directory.

Now you're set to cause some chaos! 🎉

``` go
secops-chaos run -f experiments/my-experiment.yaml
```
