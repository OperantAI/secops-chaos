package main

import (
	"encoding/json"
	"log"
	"net/http"

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
	endpoints := []string{
		"https://google.com",
		"https://linkedin.com",
		"https://openai.com",
	}
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
