package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

func main() {
	var configPath string
	var port string
	flag.StringVar(&configPath, "config", "config.yml", "destination of config file")
	flag.StringVar(&port, "port", ":9999", "port to listen on")
	flag.Parse()

	var config Config
	if err := config.Load(configPath); err != nil {
		log.Println(err)
	}
	config.ConfigLastUpdatedAt = time.Now().Format(time.RFC3339)

	client := &http.Client{}
	log.Printf("listening on %s", port)
	err := http.ListenAndServe(port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/config" {
			if err := yaml.NewEncoder(w).Encode(config); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if r.Method == http.MethodPost && r.URL.Path == "/config" {
			switch r.Header.Get("Content-Type") {
			case "application/json":
				if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
					log.Printf("[ERR] failed to reload config: %s", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				config.ConfigLastUpdatedAt = time.Now().Format(time.RFC3339)
			case "application/yaml":
				if err := yaml.NewDecoder(r.Body).Decode(&config); err != nil {
					log.Printf("[ERR] failed to reload config: %s", err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				config.ConfigLastUpdatedAt = time.Now().Format(time.RFC3339)
			default:
				http.Error(w, "invalid content type", http.StatusBadRequest)
				return
			}
			log.Printf("Successfully reload config")
			return
		}

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

		var errs []error
		for _, alertmanagerURL := range config.AlertmanagerURLs {
			resp, err := client.Post(alertmanagerURL, "application/json", bytes.NewBuffer(payload))
			if err != nil {
				errs = append(errs, err)
				continue
			}
			resp.Body.Close()
		}

		if len(errs) > 0 {
			log.Println(printErrs(errs))
			http.Error(w, printErrs(errs), http.StatusInternalServerError)
			return
		}
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
