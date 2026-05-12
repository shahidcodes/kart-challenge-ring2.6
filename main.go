package main

import (
	"context"
	"fmt"
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

// Print coupon info at startup for verification
func init() {
	fmt.Println("=== Order Food API Server ===")
	fmt.Println("Valid coupon codes (appear in ≥2 coupon files):")
	for _, c := range []string{"HAPPYHRS", "BUYGETON", "FIFTYOFF", "SIXTYOFF", "BIRTHDAY", "GNULINUX", "OVER9000", "FREEZAAA"} {
		info := coupon.Info(c)
		fmt.Printf("  %s → %v\n", c, info)
	}
}