package experiments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LLMDataLeakageExperiment struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters LLMDataLeakage     `yaml:"parameters"`
}

type LLMDataLeakage struct {
	Apis []ExecuteAIAPI `yaml:"apis"`
}

func (p *LLMDataLeakageExperiment) Name() string {
	return p.Metadata.Name
}

func (p *LLMDataLeakageExperiment) Type() string {
	return "llm-data-leakage"
}

func (p *LLMDataLeakageExperiment) Description() string {
	return "Check whether the LLM AI Model is leaking any sensitive data such as PII data or secrets and keys in its response"
}
func (p *LLMDataLeakageExperiment) Technique() string {
	return categories.MITREATLAS.Exfiltration.LLMDataLeakage.Technique
}
func (p *LLMDataLeakageExperiment) Tactic() string {
	return categories.MITREATLAS.Exfiltration.LLMDataLeakage.Tactic
}
func (p *LLMDataLeakageExperiment) Framework() string {
	return string(categories.MitreAtlas)
}

func (p *LLMDataLeakageExperiment) DependsOn() []string {
	return p.Metadata.DependsOn
}

func (p *LLMDataLeakageExperiment) Run(ctx context.Context, client *k8s.Client) error {
	_, err := client.Clientset.AppsV1().Deployments(p.Metadata.Namespace).Get(ctx, SecopsChaosAi, metav1.GetOptions{})
	if err != nil {
		return errors.New("Error in checking for Secops Chaos AI component to run AI experiments. Is it deployed? Deploy with secops-chaos component install command.")
	}
	pf := client.NewPortForwarder(ctx)
	if err != nil {
		return err
	}
	defer pf.Stop()
	forwardedPort, err := pf.Forward(p.Metadata.Namespace, fmt.Sprintf("app=%s", SecopsChaosAi), 8000)
	if err != nil {
		return err
	}
	results := make(map[string]ExecuteAIAPIResult)
	for _, api := range p.Parameters.Apis {

		url := url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", pf.Addr(), forwardedPort.Local),
			Path:   "/ai-experiments",
		}

		var requestBody []byte
		if &api.Payload != nil {
			requestBody, err = json.Marshal(api.Payload)
			if err != nil {
				return err
			}
		}
		req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(requestBody))
		if err != nil {
			return err
		}

		req.Header.Add("Content-type", "application/json")

		response, err := http.DefaultClient.Do(req)
		if err != nil || response.StatusCode != 200 {
			return err
		}
		defer response.Body.Close()
		var apiResponse AIAPIResponse
		err = json.NewDecoder(response.Body).Decode(&apiResponse)
		if err != nil {
			return err
		}

		results[api.Description] = ExecuteAIAPIResult{
			Description:    api.Description,
			ExperimentName: p.Metadata.Name,
			Timestamp:      time.Now(),
			Status:         response.StatusCode,
			Response:       apiResponse,
		}
	}

	resultJSON, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("Failed to marshal experiment results: %w", err)
	}

	file, err := createTempFile(p.Type(), p.Metadata.Name)
	if err != nil {
		return fmt.Errorf("Unable to create file cache for experiment results %w", err)
	}

	_, err = file.Write(resultJSON)
	if err != nil {
		return fmt.Errorf("Failed to write experiment results: %w", err)
	}
	return nil
}

func (p *LLMDataLeakageExperiment) Verify(ctx context.Context, client *k8s.Client) (*verifier.Outcome, error) {
	v := verifier.New(
		p.Metadata.Name,
		p.Description(),
		p.Framework(),
		p.Tactic(),
		p.Technique(),
	)

	rawResults, err := getTempFileContentsForExperiment(p.Type(), p.Metadata.Name)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch experiment results: %w", err)
	}

	var aiVerifierOutcome *verifier.AIVerifierOutcome
	for _, rawResult := range rawResults {
		var results map[string]ExecuteAIAPIResult
		err = json.Unmarshal(rawResult, &results)
		if err != nil {
			return nil, fmt.Errorf("Could not parse experiment result: %w", err)
		}

		for _, api := range p.Parameters.Apis {
			result, found := results[api.Description]
			if !found {
				continue
			}
			fail := false
			if api.ExpectedResponse.VerifiedResponseChecks != nil {
				for _, responseCheck := range api.ExpectedResponse.VerifiedResponseChecks {
					if result.Response.VerifiedResponseChecks != nil {
						fail = true
						v.Fail(api.Description)
						if aiVerifierOutcome == nil {
							aiVerifierOutcome = &verifier.AIVerifierOutcome{
								Model:                  result.Response.Model,
								AIApi:                  result.Response.AIApi,
								Prompt:                 result.Response.Prompt,
								APIResponse:            result.Response.APIResponse,
								VerifiedResponseChecks: nil,
							}
						}
						for _, resultCheck := range result.Response.VerifiedResponseChecks {
							if resultCheck.Check == responseCheck.Check {
								aiVerifierOutcome.VerifiedResponseChecks = append(aiVerifierOutcome.VerifiedResponseChecks, verifier.AIVerifierResult{
									Check:      resultCheck.Check,
									Detected:   resultCheck.Detected,
									Score:      resultCheck.Score,
									EntityType: resultCheck.EntityType,
								})
							}
						}

					}
				}
			}
			if !fail {
				v.Success(api.Description)
			}
			v.StoreResultOutputs(api.Description, aiVerifierOutcome)
		}
	}

	return v.GetOutcome(), nil
}

func (p *LLMDataLeakageExperiment) Cleanup(ctx context.Context, client *k8s.Client) error {
	if err := removeTempFilesForExperiment(p.Type(), p.Metadata.Name); err != nil {
		return err
	}

	return nil
}
