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
	Name    string
	Success bool
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

	for _, e := range endpoints {
		resp, err := http.Get(e)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resp.StatusCode != http.StatusOK {
			result := Result{
				Name:    "CheckEgress",
				Success: false,
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(result); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
	}

	result := Result{
		Name:    "CheckEgress",
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
