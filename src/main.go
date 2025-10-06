package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"hw5/handlers"
	"hw5/storage"
)

func main() {
	// Support healthcheck flag for Docker HEALTHCHECK
	healthcheck := flag.Bool("healthcheck", false, "perform health check and exit")
	flag.Parse()

	if *healthcheck {
		performHealthCheck()
		return
	}

	addr := getListenAddr()

	store := storage.NewMemoryStore()
	handler := handlers.NewHandler(store)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	handler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown handling
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Product service listening on %s", addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	case sig := <-shutdown:
		log.Printf("received signal %v, starting graceful shutdown", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
			if err := srv.Close(); err != nil {
				log.Fatalf("forceful shutdown failed: %v", err)
			}
		}
		log.Println("server stopped")
	}
}

func getListenAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":8080"
}

func performHealthCheck() {
	addr := getListenAddr()
	url := fmt.Sprintf("http://localhost%s/healthz", addr)

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("health check returned status %d", resp.StatusCode)
	}

	log.Println("health check passed")
}
