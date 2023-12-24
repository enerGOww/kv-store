package main

import (
	"context"
	"github.com/gorilla/mux"
	"kv-store/internal/closer"
	v1 "kv-store/internal/controller/v1"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	r := mux.NewRouter()

	err := v1.Init(r)
	if err != nil {
		log.Fatal("initializing error: %w", err)
	}

	server := http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		log.Println("can not stop http server")
	}

	if err = closer.Closer.Close(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
}
