package experiments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/operantai/secops-chaos/internal/categories"
	"github.com/operantai/secops-chaos/internal/k8s"
	"github.com/operantai/secops-chaos/internal/verifier"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"time"
)

type LLMDataLeakageExperiment struct {
	Metadata   ExperimentMetadata `yaml:"metadata"`
	Parameters LLMDataLeakage     `yaml:"parameters"`
}

type LLMDataLeakage struct {
	Apis []ExecuteAIAPI `yaml:"apis"`
}

type AIAPIPayload struct {
	Model                string   `json:"model" yaml:"model"`
	AIApi                string   `json:"ai_api" yaml:"ai_api"`
	SystemPrompt         string   `json:"system_prompt" yaml:"system_prompt"`
	Prompt               string   `json:"prompt" yaml:"prompt"`
	VerifyPromptChecks   []string `json:"verify_prompt_checks" yaml:"verify_prompt_checks"`
	VerifyResponseChecks []string `json:"verify_response_checks" yaml:"verify_response_checks"`
}

type AIVerifierResult struct {
	Check      string  `json:"check"`
	Detected   bool    `json:"detected"`
	EntityType string  `json:"entityType"`
	Score      float64 `json:"score"`
}

type AIAPIResponse struct {
	VerifiedPromptChecks   []AIVerifierResult `json:"verified_prompt_checks" yaml:"verified_prompt_checks"`
	VerifiedResponseChecks []AIVerifierResult `json:"verified_response_checks" yaml:"verified_response_checks"`
}

type ExecuteAIAPI struct {
	Description      string        `yaml:"description"`
	Payload          AIAPIPayload  `yaml:"payload"`
	ExpectedResponse AIAPIResponse `yaml:"expected_response"`
}

type ExecuteAIAPIResult struct {
	ExperimentName string        `json:"experiment_name"`
	Description    string        `json:"description"`
	Timestamp      time.Time     `json:"timestamp"`
	Status         int           `json:"status"`
	Response       AIAPIResponse `json:"response"`
}

func (p *LLMDataLeakageExperiment) Type() string {
	return "llm_data_leakage"
}

func (p *LLMDataLeakageExperiment) Description() string {
	return "This experiment checks whether the LLM AI Model is leaking any sensitive data such as PII data or secrets and keys in its response"
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

func (p *LLMDataLeakageExperiment) Run(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config LLMDataLeakageExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}
	pf := client.NewPortForwarder(ctx)
	if err != nil {
		return err
	}
	defer pf.Stop()
	forwardedPort, err := pf.Forward(config.Metadata.Namespace, "app=secops-chaos-ai", 8000)
	if err != nil {
		return err
	}
	results := make(map[string]ExecuteAIAPIResult)
	for _, api := range config.Parameters.Apis {

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
		fmt.Println(string(requestBody))
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
			ExperimentName: config.Metadata.Name,
			Timestamp:      time.Now(),
			Status:         response.StatusCode,
			Response:       apiResponse,
		}
	}

	resultJSON, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("Failed to marshal experiment results: %w", err)
	}

	file, err := createTempFile(p.Type(), config.Metadata.Name)
	if err != nil {
		return fmt.Errorf("Unable to create file cache for experiment results %w", err)
	}

	_, err = file.Write(resultJSON)
	if err != nil {
		return fmt.Errorf("Failed to write experiment results: %w", err)
	}

	return nil
}

func (p *LLMDataLeakageExperiment) Verify(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) (*verifier.Outcome, error) {
	var config LLMDataLeakageExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return nil, err
	}

	v := verifier.New(
		config.Metadata.Name,
		config.Description(),
		config.Framework(),
		config.Tactic(),
		config.Technique(),
	)

	rawResults, err := getTempFileContentsForExperiment(p.Type(), config.Metadata.Name)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch experiment results: %w", err)
	}

	for _, rawResult := range rawResults {
		var results map[string]ExecuteAIAPIResult
		err = json.Unmarshal(rawResult, &results)
		if err != nil {
			return nil, fmt.Errorf("Could not parse experiment result: %w", err)
		}

		for _, api := range config.Parameters.Apis {
			result, found := results[api.Description]
			if !found {
				continue
			}
			fail := false
			if api.ExpectedResponse.VerifiedResponseChecks != nil {
				for _, responseCheck := range api.ExpectedResponse.VerifiedResponseChecks {
					if result.Response.VerifiedResponseChecks != nil {
						for _, resultCheck := range result.Response.VerifiedResponseChecks {
							if resultCheck.Check == responseCheck.Check {
								if resultCheck.Detected != responseCheck.Detected {
									fail = true
									v.Fail(api.Description)
								}
							}
						}
					}
				}
			}
			if !fail {
				v.Success(api.Description)
			}
		}
	}

	return v.GetOutcome(), nil
}

func (p *LLMDataLeakageExperiment) Cleanup(ctx context.Context, client *k8s.Client, experimentConfig *ExperimentConfig) error {
	var config RemoteExecuteAPIExperimentConfig
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	if err := removeTempFilesForExperiment(p.Type(), config.Metadata.Name); err != nil {
		return err
	}

	return nil
}
