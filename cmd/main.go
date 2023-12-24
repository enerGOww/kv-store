package main

import (
	"github.com/gorilla/mux"
	v1 "kv-store/internal/controller/v1"
	"log"
	"net/http"
)

func main() {
	r := mux.NewRouter()

	err := v1.Init(r)
	if err != nil {
		log.Fatal("initializing error: %w", err)
	}

	log.Fatal(http.ListenAndServe(":8080", r))
}
