package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

type Result struct {
	Name      string      `json:"name"`
	URLResult []URLResult `json:"url_result"`
}

type URLResult struct {
	URL     string `json:"url"`
	Success bool   `json:"success"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/experiment/CheckEgress/", CheckEgress)

	// Start the experiment server
	log.Print("starting server on :4000")
	err := http.ListenAndServe(":4000", r)
	if err != nil {
		log.Fatal(err)
	}

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
