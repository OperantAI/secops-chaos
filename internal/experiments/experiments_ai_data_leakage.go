package experiments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/operantai/woodpecker/internal/categories"
	"github.com/operantai/woodpecker/internal/verifier"
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

func (p *LLMDataLeakageExperiment) Run(ctx context.Context, experimentConfig *ExperimentConfig) error {
	var config LLMDataLeakageExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return err
	}

	results := make(map[string]ExecuteAIAPIResult)
	verifierAddr, appAddr, err := getAIComponentAddrs(ctx, experimentConfig)
	if err != nil {
		return err
	}

	for _, api := range config.Parameters.Apis {
		appURL := url.URL{
			Scheme: "http",
			Host:   appAddr,
			Path:   "/chat",
		}

		var appRequestBody []byte
		if &api.Payload == nil {
			return errors.New("Payload may not be empty")
		}

		aiAppPayload := AIAppPayload{
			Model:        api.Payload.Model,
			SystemPrompt: api.Payload.SystemPrompt,
			Prompt:       api.Payload.Prompt,
		}

		appRequestBody, err = json.Marshal(aiAppPayload)
		if err != nil {
			return err
		}
		appReq, err := http.NewRequest("POST", appURL.String(), bytes.NewBuffer(appRequestBody))
		if err != nil {
			return err
		}

		appReq.Header.Add("Content-type", "application/json")

		appResponse, err := http.DefaultClient.Do(appReq)
		if err != nil || appResponse.StatusCode != 200 {
			return err
		}
		defer appResponse.Body.Close()

		var ar AIAppResponse
		err = json.NewDecoder(appResponse.Body).Decode(&ar)
		if err != nil {
			return err
		}

		verifierURL := url.URL{
			Scheme: "http",
			Host:   verifierAddr,
			Path:   "/v1/ai-experiments",
		}

		var verifierRequestBody []byte
		verifierRequest := AIAPIPayload{
			Model:                aiAppPayload.Model,
			SystemPrompt:         aiAppPayload.SystemPrompt,
			Prompt:               aiAppPayload.Prompt,
			AIApi:                api.Payload.AIApi,
			Response:             ar.Message,
			VerifyPromptChecks:   api.Payload.VerifyPromptChecks,
			VerifyResponseChecks: api.Payload.VerifyResponseChecks,
		}

		verifierRequestBody, err = json.Marshal(verifierRequest)
		if err != nil {
			return err
		}
		verifierReq, err := http.NewRequest("POST", verifierURL.String(), bytes.NewBuffer(verifierRequestBody))
		if err != nil {
			return err
		}

		verifierReq.Header.Add("Content-type", "application/json")

		verifierResponse, err := http.DefaultClient.Do(verifierReq)
		if err != nil || appResponse.StatusCode != 200 {
			return err
		}
		defer verifierResponse.Body.Close()

		var vr AIVerifierAPIResponse
		err = json.NewDecoder(verifierResponse.Body).Decode(&vr)
		if err != nil {
			return err
		}

		results[api.Description] = ExecuteAIAPIResult{
			Description:    api.Description,
			ExperimentName: config.Metadata.Name,
			Timestamp:      time.Now(),
			Status:         appResponse.StatusCode,
			Response:       vr,
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

// Updated Verify method using generics - THIS IS THE KEY CHANGE
func (p *LLMDataLeakageExperiment) Verify(ctx context.Context, experimentConfig *ExperimentConfig) (*verifier.LegacyOutcome, error) {
	var config LLMDataLeakageExperiment
	yamlObj, _ := yaml.Marshal(experimentConfig)
	err := yaml.Unmarshal(yamlObj, &config)
	if err != nil {
		return nil, err
	}

	// Use the NEW AI-specific verifier with generics
	v := verifier.NewAIVerifier(
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

			// Create properly typed AIVerifierOutcome
			aiVerifierOutcome := verifier.AIVerifierOutcome{
				Model:       result.Response.Model,
				AIApi:       result.Response.AIApi,
				Prompt:      result.Response.Prompt,
				APIResponse: result.Response.APIResponse,
				// Convert the API response types to verifier types
				VerifiedPromptChecks:   convertToVerifierResults(result.Response.VerifiedPromptChecks),
				VerifiedResponseChecks: convertToVerifierResults(result.Response.VerifiedResponseChecks),
			}

			fail := false
			if api.ExpectedResponse.VerifiedResponseChecks != nil {
				for range api.ExpectedResponse.VerifiedResponseChecks {
					if len(aiVerifierOutcome.VerifiedResponseChecks) > 0 {
						fail = true
						v.Fail(api.Description)
						break
					}
				}
			}
			if !fail {
				v.Success(api.Description)
			}

			// Store the strongly typed result - NO MORE NULL FIELDS!
			v.StoreResultOutputs(api.Description, aiVerifierOutcome)
		}
	}

	// Convert to legacy format for backward compatibility
	typedOutcome := v.GetOutcome()
	legacyOutcome := &verifier.LegacyOutcome{
		Experiment:    typedOutcome.Experiment,
		Description:   typedOutcome.Description,
		Framework:     typedOutcome.Framework,
		Tactic:        typedOutcome.Tactic,
		Technique:     typedOutcome.Technique,
		Result:        typedOutcome.Result,
		ResultOutputs: convertToInterfaceMap(typedOutcome.ResultOutputs),
	}

	return legacyOutcome, nil
}

// Helper function to convert typed results to interface{} for legacy compatibility
func convertToInterfaceMap(typedResults map[string][]verifier.AIVerifierOutcome) map[string][]interface{} {
	result := make(map[string][]interface{})
	for key, values := range typedResults {
		interfaceSlice := make([]interface{}, len(values))
		for i, value := range values {
			interfaceSlice[i] = value
		}
		result[key] = interfaceSlice
	}
	return result
}

// Convert API response types to verifier types
func convertToVerifierResults(apiResults []AIVerifierResult) []verifier.AIVerifierResult {
	if apiResults == nil {
		return []verifier.AIVerifierResult{} // Return empty slice instead of nil
	}

	results := make([]verifier.AIVerifierResult, len(apiResults))
	for i, apiResult := range apiResults {
		results[i] = verifier.AIVerifierResult{
			Check:      apiResult.Check,
			Detected:   apiResult.Detected,
			EntityType: apiResult.EntityType,
			Score:      apiResult.Score,
		}
	}
	return results
}

func (p *LLMDataLeakageExperiment) Cleanup(ctx context.Context, experimentConfig *ExperimentConfig) error {
	var config LLMDataLeakageExperiment
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
