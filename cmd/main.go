package main

import (
	"github.com/gorilla/mux"
	v1 "kv-store/internal/controller/v1"
	"log"
	"net/http"
)

func main() {
	r := mux.NewRouter()

	v1.Init(r)

	log.Fatal(http.ListenAndServe(":8080", r))
}
