package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/common/model"
)

func main() {
	var alertmanagerURL string
	var configPath string
	var port string
	flag.StringVar(&configPath, "config", "config.yml", "destination of config file")
	flag.StringVar(&port, "port", ":9999", "port to listen on")
	flag.StringVar(&alertmanagerURL, "alertmanager-url", "http://localhost:9093", "alertmanager url")
	flag.Parse()

	if !strings.HasSuffix(alertmanagerURL, "/api/v1/alerts") {
		alertmanagerURL += "/api/v1/alerts"
	}

	var config Config
	if err := config.Load(configPath); err != nil {
		log.Fatal(err)
	}
	config.ConfigLastUpdatedAt = time.Now().Format(time.RFC3339)

	client := &http.Client{}
	log.Printf("alert-relabelling running on %s", port)
	err := http.ListenAndServe(port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config" {
			config.Handler(w, r)
			return
		}

		// alerts handling
		var incomingAlerts []model.Alert
		if err := json.NewDecoder(r.Body).Decode(&incomingAlerts); err != nil {
			log.Printf("[ERR] failed to decode incoming alerts: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for key := range incomingAlerts {
			relabelling(&config, &incomingAlerts[key])
		}

		payload, err := json.Marshal(incomingAlerts)
		if err != nil {
			log.Printf("[ERR] failed to marshal incoming alerts: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := client.Post(alertmanagerURL, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("[ERR] failed to post alerts to alertmanager: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Body.Close()
	}))

	if err != nil {
		log.Fatal(err)
	}
}

func printErrs(errs []error) string {
	var errStr string
	for _, err := range errs {
		errStr += err.Error()
	}
	return errStr
}
