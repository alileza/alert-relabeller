package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
)

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests by status code and method.",
	}, []string{"code", "method", "path"})

	relabelingMatchesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relabelling_matches_total",
		Help: "Total number of requests that matched relabelling rules.",
	})
)

func main() {
	var alertmanagerURL string
	var configPath string
	var port string
	flag.StringVar(&configPath, "config", "config.yml", "destination of config file")
	flag.StringVar(&port, "port", ":9999", "port to listen on")
	flag.StringVar(&alertmanagerURL, "alertmanager-url", "http://localhost:9093", "alertmanager url")
	flag.Parse()

	uam, err := url.Parse(alertmanagerURL)
	if err != nil {
		log.Fatal(err)
	}
	amproxy := httputil.NewSingleHostReverseProxy(uam)

	var config Config
	if err := config.Load(configPath); err != nil {
		log.Fatal(err)
	}
	config.ConfigLastUpdatedAt = time.Now().Format(time.RFC3339)

	log.Printf("alert-relabelling running on %s", port)
	err = http.ListenAndServe(port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/config", http.StatusFound)
			return
		}

		if r.URL.Path == "/metrics" {
			promhttp.Handler().ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/favicon.ico" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.URL.Path == "/config" {
			config.Handler(w, r)
			return
		}

		if r.URL.Path == "/-/ready" || r.URL.Path == "/-/healthy" {
			amproxy.ServeHTTP(w, r)
			return
		}

		// alerts handling
		var incomingAlerts []model.Alert
		if err := json.NewDecoder(r.Body).Decode(&incomingAlerts); err != nil {
			log.Printf("[ERR] failed to decode incoming alerts: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			requestsTotal.WithLabelValues(toString(http.StatusInternalServerError), r.Method, r.URL.Path).Inc()
			return
		}

		for key := range incomingAlerts {
			relabelling(&config, &incomingAlerts[key])
		}

		payload, err := json.Marshal(incomingAlerts)
		if err != nil {
			log.Printf("[ERR] failed to marshal incoming alerts: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			requestsTotal.WithLabelValues(toString(http.StatusInternalServerError), r.Method, r.URL.Path).Inc()
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(payload))
		r.ContentLength = int64(len(payload))
		amproxy.ServeHTTP(w, r)
		requestsTotal.WithLabelValues(toString(http.StatusOK), r.Method, r.URL.Path).Inc()
	}))

	if err != nil {
		log.Fatal(err)
	}
}

func toString(i int) string {
	return fmt.Sprintf("%d", i)
}
