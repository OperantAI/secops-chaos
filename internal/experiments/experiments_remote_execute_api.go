package experiments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/executor"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"
)

// RemoteExecuteAPI is an experiment that uses the remote executor to check a remote output
// The image must be created independently -- the current default is `alconen/egress_server`, which runs a simple web app on port 4000 that checks http connectivity to
// a few domains ("https://google.com", "https://linkedin.com", "https://openai.com/") and responds with a success based on the success of those calls.
// The source can be found at cmd/executor-server
type RemoteExecuteAPIExperiment struct {
	Metadata   ExperimentMetadata
	Parameters executor.RemoteExecuteAPI
}

func (p *RemoteExecuteAPIExperiment) Name() string {
	return p.Metadata.Name
}

func (p *RemoteExecuteAPIExperiment) Type() string {
	return "remote-execute-api"
}

func (p *RemoteExecuteAPIExperiment) Description() string {
	return "Runs a deployment based on a configurable image and then verifies based off of API calls to that image"
}
func (p *RemoteExecuteAPIExperiment) Technique() string {
	return categories.MITRE.Execution.NewContainer.Technique
}
func (p *RemoteExecuteAPIExperiment) Tactic() string {
	return categories.MITRE.Execution.NewContainer.Tactic
}
func (p *RemoteExecuteAPIExperiment) Framework() string {
	return string(categories.Mitre)
}

func (p *RemoteExecuteAPIExperiment) DependsOn() []string {
	return p.Metadata.DependsOn
}

func (p *RemoteExecuteAPIExperiment) Run(ctx context.Context, client *k8s.Client) error {
	executorConfig := executor.NewExecutorConfig(
		p.Metadata.Name,
		p.Metadata.Namespace,
		p.Parameters.Image,
		p.Parameters.ImageParameters,
		p.Parameters.ServiceAccountName,
		p.Parameters.Target.Port,
	)

	err := executorConfig.Deploy(ctx, client.Clientset)
	if err != nil {
		return err
	}

	return nil
}

func (p *RemoteExecuteAPIExperiment) Verify(ctx context.Context, client *k8s.Client) (*verifier.Outcome, error) {

	pf := client.NewPortForwarder(ctx)
	defer pf.Stop()
	forwardedPort, err := pf.Forward(p.Metadata.Namespace, fmt.Sprintf("app=%s", p.Metadata.Name), int(p.Parameters.Target.Port))
	if err != nil {
		return nil, err
	}

	path := p.Parameters.Target.Path
	url := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", pf.Addr(), int32(forwardedPort.Local)),
		Path:   path,
	}

	v := verifier.New(
		p.Metadata.Name,
		p.Description(),
		p.Framework(),
		p.Tactic(),
		p.Technique(),
	)

	result, err := p.retrieveAPIResponse(url.String())
	if err != nil {
		return nil, err
	}

	for _, r := range result.URLResult {
		if r.Success {
			v.Success(r.URL)
		} else if !r.Success {
			v.Fail(r.URL)
		}
	}

	return v.GetOutcome(), nil
}

func (p *RemoteExecuteAPIExperiment) retrieveAPIResponse(url string) (*executor.Result, error) {
	var result executor.Result
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *RemoteExecuteAPIExperiment) Cleanup(ctx context.Context, client *k8s.Client) error {
	clientset := client.Clientset
	executorConfig := executor.NewExecutorConfig(
		p.Metadata.Name,
		p.Metadata.Namespace,
		p.Parameters.Image,
		p.Parameters.ImageParameters,
		p.Parameters.ServiceAccountName,
		p.Parameters.Target.Port,
	)

	err := executorConfig.Cleanup(ctx, clientset)
	if err != nil {
		return err
	}

	return nil
}
