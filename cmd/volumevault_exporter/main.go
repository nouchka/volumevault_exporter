package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nouchka/volumevault_exporter/internal/exporter"
	"github.com/nouchka/volumevault_exporter/internal/volumevault"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	listenAddress := flag.String("web.listen-address", envOrDefault("LISTEN_ADDRESS", ":9780"), "Address to listen on for web interface and telemetry.")
	metricsPath := flag.String("web.telemetry-path", envOrDefault("METRICS_PATH", "/metrics"), "Path under which to expose metrics.")
	apiURL := flag.String("volumevault.api-url", envOrDefault("VOLUMEVAULT_API_URL", ""), "VolumeVault API base URL, e.g. https://volumevault.example.com/api/v1")
	apiToken := flag.String("volumevault.api-token", envOrDefault("VOLUMEVAULT_API_TOKEN", ""), "VolumeVault API bearer token")
	timeout := flag.Duration("volumevault.timeout", envDurationOrDefault("VOLUMEVAULT_TIMEOUT", 10*time.Second), "VolumeVault API HTTP timeout")
	flag.Parse()

	if *apiURL == "" {
		log.Fatal("volumevault.api-url (or VOLUMEVAULT_API_URL) is required")
	}
	if *apiToken == "" {
		log.Fatal("volumevault.api-token (or VOLUMEVAULT_API_TOKEN) is required")
	}

	client := volumevault.NewClient(*apiURL, *apiToken, *timeout)
	collector := exporter.NewCollector(client)

	registry := prometheus.NewRegistry()
	if err := registry.Register(collector); err != nil {
		log.Fatalf("failed to register collector: %v", err)
	}

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body><h1>VolumeVault Exporter</h1><p><a href=\"" + *metricsPath + "\">Metrics</a></p></body></html>"))
	})

	log.Printf("starting volumevault_exporter on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDurationOrDefault(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return def
}
