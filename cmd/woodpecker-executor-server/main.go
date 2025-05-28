package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
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
	r.HandleFunc("/experiment/listKubernetesSecrets/{namespace}", ListK8sSecrets)

	// Start the experiment server
	log.Print("starting server on :4000")
	err := http.ListenAndServe(":4000", r)
	if err != nil {
		log.Fatal(err)
	}
}
