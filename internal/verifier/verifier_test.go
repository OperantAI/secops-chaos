package verifier

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifier_AssertEqual(t *testing.T) {
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
		wantPass bool
	}{
		{
			"TestEqual",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod-1",
				},
			},
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod-1",
				},
			},
			true,
		},
		{
			"TestNotEqual",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod-1",
				},
			},
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "pod-2",
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := New("test_experiment", "test_category")

			result := verifier.AssertEqual(
				tt.actual.(*v1.Pod).ObjectMeta.Name,
				tt.expected.(*v1.Pod).ObjectMeta.Name,
			)

			require.Equal(t, tt.wantPass, result)
			outcome := verifier.GetOutcome()
			assert.Equal(t, tt.wantPass, outcome.Result.Successful == outcome.Result.Total)
		})
	}
}
