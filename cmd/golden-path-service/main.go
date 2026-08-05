package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/vaipexlabs/platform-lab-01-golden-path/internal/httpapi"
)

const address = ":8080"

func main() {
	server := &http.Server{
		Addr:              address,
		Handler:           httpapi.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("golden-path-service listening on %s", address)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server stopped unexpectedly: %v", err)
	}
}
