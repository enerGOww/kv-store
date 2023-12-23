package v1

import (
	"errors"
	"github.com/gorilla/mux"
	"io"
	domainError "kv-store/internal/error"
	kvStore "kv-store/internal/service"
	"net/http"
)

func Init(r *mux.Router) {
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

	err = kvStore.Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	res, err := kvStore.Get(key)
	if errors.Is(err, domainError.ErrorNoSuchKey) {
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

	err := kvStore.Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello w#rld!\n"))
}
