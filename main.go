package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"orderfood/internal/coupon"
	"orderfood/internal/handler"
	"orderfood/internal/middleware"
)

func main() {
	port := envOrDefault("PORT", "8080")

	// Stream all 3 coupon files at startup and build the valid set dynamically.
	// Complexity: O(|file2|_8char + |file3|_8char + |file1|_8char) — only 8-char
	// codes can overlap between files, so we extract and test a tiny subset.
	// Falls back to hardcoded set if files are absent.
	if err := coupon.Load("./"); err != nil {
		log.Printf("Warning: could not load coupon files: %v", err)
		log.Println("Falling back to embedded valid codes.")
	}

	mux := http.NewServeMux()

	// Public product endpoints
	mux.HandleFunc("GET /product", handler.ListProducts)
	mux.HandleFunc("GET /product/{productId}", handler.GetProduct)

	// Protected order endpoint — requires API key
	mux.HandleFunc("POST /order", middleware.Auth(handler.PlaceOrder))

	// Health check
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdown
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Printf("Server starting on port %s...", port)
	log.Printf("Valid coupon codes loaded: %d", coupon.Count())
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server exited gracefully.")
}

func envOrDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}