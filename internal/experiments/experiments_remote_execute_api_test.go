package experiments

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientUpperCase(t *testing.T) {
	expectedBody := "{\"Name\":\"CheckEgress\",\"Success\":true}"
	expectedStatusCode := 200

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, expectedBody)
	}))
	defer testServer.Close()

	body, statusCode, err := retrieveAPIResponse(testServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expectedBody, body)
	assert.Equal(t, expectedStatusCode, statusCode)
}
