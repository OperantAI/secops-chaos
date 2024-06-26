package verifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifier_Success(t *testing.T) {
	tests := []struct {
		testName string
		test     string
		expect   *Outcome
	}{
		{
			testName: "Set outcome to true for experiment when test is empty",
			test:     "",
			expect: &Outcome{
				Experiment:  "experiment_name",
				Description: "experiment_description",
				Framework:   "experiment_framework",
				Tactic:      "experiment_tactic",
				Technique:   "experiment_technique",
				Result: map[string]string{
					"experiment_name": Success,
				},
				ResultOutputs: make(map[string][]interface{}),
			},
		},
		{
			testName: "Set outcome to true for a specific test",
			test:     "test_name",
			expect: &Outcome{
				Experiment:  "experiment_name",
				Description: "experiment_description",
				Framework:   "experiment_framework",
				Tactic:      "experiment_tactic",
				Technique:   "experiment_technique",
				Result: map[string]string{
					"test_name": Success,
				},
				ResultOutputs: make(map[string][]interface{}),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			verifier := New("experiment_name", "experiment_description", "experiment_framework", "experiment_tactic", "experiment_technique")
			verifier.Success(test.test)
			assert.Equal(t, test.expect, verifier.GetOutcome())
		})
	}
}

func TestVerifier_Fail(t *testing.T) {
	tests := []struct {
		testName string
		test     string
		expect   *Outcome
	}{
		{
			testName: "Set outcome to false for experiment when test is empty",
			test:     "",
			expect: &Outcome{
				Experiment:  "experiment_name",
				Description: "experiment_description",
				Framework:   "experiment_framework",
				Tactic:      "experiment_tactic",
				Technique:   "experiment_technique",
				Result: map[string]string{
					"experiment_name": Fail,
				},
				ResultOutputs: make(map[string][]interface{}),
			},
		},
		{
			testName: "Set outcome to false for a specific test",
			test:     "test_name",
			expect: &Outcome{
				Experiment:  "experiment_name",
				Description: "experiment_description",
				Framework:   "experiment_framework",
				Tactic:      "experiment_tactic",
				Technique:   "experiment_technique",
				Result: map[string]string{
					"test_name": Fail,
				},
				ResultOutputs: make(map[string][]interface{}),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			verifier := New("experiment_name", "experiment_description", "experiment_framework", "experiment_tactic", "experiment_technique")
			verifier.Fail(test.test)
			assert.Equal(t, test.expect, verifier.GetOutcome())
		})
	}
}

func TestOutcome_GetResultString(t *testing.T) {
	tests := []struct {
		testName string
		outcome  *Outcome
		expected string
	}{
		{
			testName: "Generate result string for a successful experiment",
			outcome: &Outcome{
				Experiment: "experiment_name",
				Result: map[string]string{
					"experiment_name": Success,
				},
			},
			expected: "experiment_name: success\n",
		},
		{
			testName: "Generate result string for a failed experiment",
			outcome: &Outcome{
				Experiment: "experiment_name",
				Result: map[string]string{
					"experiment_name": Fail,
				},
			},
			expected: "experiment_name: fail\n",
		},
		{
			testName: "Generate result string for multiple experiments",
			outcome: &Outcome{
				Experiment: "experiment_name",
				Result: map[string]string{
					"test_name1": Success,
					"test_name2": Fail,
				},
			},
			expected: "test_name1: success\ntest_name2: fail\n OR test_name2: fail\ntest_name1: success\n",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			resultString := test.outcome.GetResultString()
			assert.Contains(t, test.expected, resultString)
		})
	}
}
