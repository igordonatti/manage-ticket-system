package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/igordonatti/sistema-gestao-ingressos/backend/internal/httpapi"
	"github.com/igordonatti/sistema-gestao-ingressos/backend/internal/redisstore"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient := redisstore.New(redisAddr)

	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("error closing Redis client redis: %v", err)
		}
	}()

	healthHandler := httpapi.NewHealthHandler(
		redisClient,
		500*time.Millisecond,
	)

	// isso aqui cria um "multiplexer" (http router)
	mux := http.NewServeMux()
	// registry the health route
	mux.HandleFunc("/health", healthHandler)

	// servidor na porta 8080
	// timeouts evitam que o servidor fique preso por reqs lentas ou maliciosas
	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Println("server running at :8080")

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case sig := <-quit:
		log.Printf("finishing the server after signal %s...", sig)
	case err := <-serverErr:
		if err != nil {
			log.Printf("error runing the server: %v", err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("erro ao encerrar servidor: %v", err)
	}

	if err := <-serverErr; err != nil {
		log.Printf("servidor finalizado com erro: %v", err)
	}
}
