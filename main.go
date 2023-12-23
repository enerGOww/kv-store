package main

import (
	"errors"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"sync"
)

var store = struct {
	m map[string]string
	sync.RWMutex
}{m: make(map[string]string)}

var ErrorNoSuchKey = errors.New("No such key")

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", helloHandler)

	r.HandleFunc("/v1/key/{key}", keyValuePutHandler).
		Methods("PUT").
		Schemes("http", "https")

	r.HandleFunc("/v1/key/{key}", keyValueGetHandler).
		Methods("GET").
		Schemes("http", "https")

	r.HandleFunc("/v1/key/{key}", keyValueDeleteHandler).
		Methods("DELETE").
		Schemes("http", "https")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func keyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	res, err := Get(key)
	if errors.Is(err, ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(res))
}

func keyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello w#rld!\n"))
}

func Put(key string, value string) error {
	store.Lock()
	store.m[key] = value
	store.Unlock()

	return nil
}

func Get(key string) (string, error) {
	store.RLock()
	value, ok := store.m[key]
	store.RUnlock()
	if !ok {
		return "", ErrorNoSuchKey
	}

	return value, nil
}

func Delete(key string) error {
	store.Lock()
	delete(store.m, key)
	store.Unlock()

	return nil
}
