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

	"github.com/anshulpatel25/termux-api-exporter/collector"
	"github.com/anshulpatel25/termux-api-exporter/executor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress = flag.String("listen-address", ":9797", "Address to listen on for HTTP requests")
	metricsPath   = flag.String("metrics-path", "/metrics", "Path under which to expose metrics")
	timeout       = flag.Duration("timeout", 5*time.Second, "Timeout for executing termux commands")
)

func main() {
	flag.Parse()

	log.Printf("Starting Termux API Exporter")
	log.Printf("Listen address: %s", *listenAddress)
	log.Printf("Metrics path: %s", *metricsPath)
	log.Printf("Command timeout: %s", *timeout)

	// Create the executor
	exec := executor.NewTermuxExecutor(*timeout)

	// Create collectors
	batteryCollector := collector.NewBatteryCollector(exec)
	wifiCollector := collector.NewWiFiCollector(exec)

	// Create a new Prometheus registry
	registry := prometheus.NewRegistry()

	// Register collectors
	if err := registry.Register(batteryCollector); err != nil {
		log.Fatalf("Failed to register battery collector: %v", err)
	}

	if err := registry.Register(wifiCollector); err != nil {
		log.Fatalf("Failed to register WiFi collector: %v", err)
	}

	// Setup HTTP handlers
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      log.Default(),
		ErrorHandling: promhttp.ContinueOnError,
	}))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK\n")
	})

	// Root endpoint with information
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html>
<head><title>Termux API Exporter</title></head>
<body>
<h1>Termux API Exporter</h1>
<p><a href="%s">Metrics</a></p>
<p><a href="/health">Health</a></p>
</body>
</html>`, *metricsPath)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         *listenAddress,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", *listenAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
