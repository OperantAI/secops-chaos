package experiments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name        string
		contents    []byte
		expectError bool
	}{
		{
			name: "Valid Experiment",
			contents: []byte(`
experiments:
- metadata:
    name: "Experiment 1"
    namespace: "my-namespace"
    type: "privileged_container"
    labels:
      key1: "value1"
  parameters:
    hostPid: true
`),
			expectError: false,
		},
		{
			name: "Invalid Experiment (missing Parameters)",
			contents: []byte(`
experiments:
- metadata:
    name: "Experiment 2"
    namespace: "my-namespace"
    type: "privileged_container"
    labels:
      key1: "value1"
`),
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := unmarshalYAML(test.contents)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
