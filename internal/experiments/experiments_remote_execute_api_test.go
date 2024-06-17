package experiments

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/operantai/secops-chaos/internal/executor"
	"github.com/stretchr/testify/assert"
)

func TestRetrieveAPIResponse(t *testing.T) {
	testBody := executor.Result{
		Name: "CheckEgress",
		URLResult: []executor.URLResult{
			{
				URL:     "google.com",
				Success: true,
			},
			{
				URL:     "blah.com",
				Success: false,
			},
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(testBody)
	}))
	defer testServer.Close()

	config := RemoteExecuteAPIExperiment{}
	result, err := config.retrieveAPIResponse(testServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, result, &testBody)
}
