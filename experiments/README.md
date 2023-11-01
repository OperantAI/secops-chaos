# Experiments

**secops-chaos** experiments are driven by **experiment** files, they all follow a common format, made up of a `metadata` and `parameters` section.

``` yaml
experiments:
  - metadata:
      name: my-experiment # A meaningful name for your experiment
      type: prvileged_container # The type of experiment, see table below for a list of valid types
      namespace: my-namespace # What namespace to apply the experiment to
    parameters: # Parameters holds the settings for your experiment, tweak them to suit your needs.
        host_pid: true 
```


## Available Experiments

| Type                   | Description                                                                         | Framework |
| `privileged_container` | This experiment attempts to run a privileged container in a namespace               | MITRE     |
| `host_path_mount`      | This experiment attempts to mount a sensitive host filesystem path into a container | MITRE     |


## Implementing a new Experiment

Each experiment within `secops-chaos` adheres to a shared interface, this allows for a common set of functionality to be used across all experiments.

When implementing a new experiment you should create a new file starting with `experiment_` within the [internal/experiments](https://github.com/OperantAI/secops-chaos/blob/main/internal/experiments/) directory, and implement the `Experiment` interface.

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
	Run(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) error
	// Verify verifies the experiment, returning an error if it fails
	Verify(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) (*Outcome, error)
	// Cleanup cleans up the experiment, returning an error if it fails
	Cleanup(ctx context.Context, client *kubernetes.Clientset, experimentConfig *ExperimentConfig) error
}
```
