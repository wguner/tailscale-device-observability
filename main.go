package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wguner/tailscale-exporter/collector"
	"github.com/wguner/tailscale-exporter/tailscale"
)

func main() {
	apiKey := os.Getenv("TAILSCALE_API_KEY")
	tailnet := os.Getenv("TAILSCALE_TAILNET")
	listenAddr := os.Getenv("LISTEN_ADDR")

	if apiKey == "" || tailnet == "" {
		log.Fatal("TAILSCALE_API_KEY and TAILSCALE_TAILNET environment variables are required")
	}
	if listenAddr == "" {
		listenAddr = ":9100"
	}

	client := tailscale.NewClient(apiKey, tailnet)
	deviceCollector := collector.NewDeviceCollector(client)

	prometheus.MustRegister(deviceCollector)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Printf("tailscale-exporter listening on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
