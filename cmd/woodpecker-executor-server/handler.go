package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/operantai/woodpecker/internal/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	"strings"
)

func ListK8sSecrets(w http.ResponseWriter, r *http.Request) {
	client, err := k8s.NewClientInContainer()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := context.Background()
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	result := false
	_, err = client.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(result)
	}
	result = true
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

func CheckEgress(w http.ResponseWriter, r *http.Request) {
	urls, exists := os.LookupEnv("URLS")
	if !exists {
		http.Error(w, "No URLS found in environment", http.StatusInternalServerError)
	}

	endpoints := strings.Split(urls, ",")

	var urlResult []URLResult

	for _, e := range endpoints {
		resp, err := http.Get(e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if resp.StatusCode != http.StatusOK {
			urlResult = append(urlResult, URLResult{
				URL:     e,
				Success: false,
			})
		} else if resp.StatusCode == http.StatusOK {
			urlResult = append(urlResult, URLResult{
				URL:     e,
				Success: true,
			})
		}
	}

	result := Result{
		Name:      "CheckEgress",
		URLResult: urlResult,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
