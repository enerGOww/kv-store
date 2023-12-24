package v1

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"kv-store/internal/closer"
	domainError "kv-store/internal/error"
	"kv-store/internal/event"
	"kv-store/internal/repository"
	kvStore "kv-store/internal/service"
	"net/http"
)

//TODO ограничение в размерах ключей и значений

var transactionRepository repository.TransactionLogger

func Init(r *mux.Router) error {
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

	return initTransactionLog()
}

func initTransactionLog() error {
	var err error

	transactionRepository, err = repository.NewTransactionLogger("transaction.log")
	if err != nil {
		return fmt.Errorf("failed to create event logger: %w", err)
	}

	closer.Closer.Add(transactionRepository.Close)

	events, eventErrors := transactionRepository.ReadEvent()
	e, ok := event.TransactionEvent{}, true

	for ok && err != nil {
		select {
		case err, ok = <-eventErrors:
		case e, ok = <-events:
			switch e.EventType {
			case event.EventDelete:
				err = kvStore.Delete(e.Key)
			case event.EventPut:
				err = kvStore.Put(e.Key, e.Value)
			}
		}
	}

	transactionRepository.Run()

	return err
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
	transactionRepository.WritePut(key, string(value))

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
	transactionRepository.WriteDelete(key)

	w.WriteHeader(http.StatusNoContent)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello w#rld!\n"))
}
