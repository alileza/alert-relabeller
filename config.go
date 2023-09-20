package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Rules               []Rule `yaml:"rules" json:"rules"`
	ConfigLastUpdatedAt string `yaml:"config_last_updated_at" json:"config_last_updated_at"`
}

type Rule struct {
	If   string            `yaml:"if" json:"if"`
	Then map[string]string `yaml:"then" json:"then"`
}

func (c *Config) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(c); err != nil {
		return err
	}
	return nil
}

func (c *Config) Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		c.GetHandler(w, r)
	case http.MethodPost:
		c.PostHandler(w, r)
	default:
		http.Error(w, "invalid method", http.StatusBadRequest)
	}
}

func (c *Config) GetHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(c); err != nil {
			log.Printf("[ERR] failed to encode config (json): %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		w.Header().Set("Content-Type", "application/yaml")
		if err := yaml.NewEncoder(w).Encode(c); err != nil {
			log.Printf("[ERR] failed to encode config (yaml): %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (c *Config) PostHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			log.Printf("[ERR] failed to reload config: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c.ConfigLastUpdatedAt = time.Now().Format(time.RFC3339)
		w.WriteHeader(http.StatusOK)
	case "application/yaml":
		if err := yaml.NewDecoder(r.Body).Decode(&c); err != nil {
			log.Printf("[ERR] failed to reload config: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c.ConfigLastUpdatedAt = time.Now().Format(time.RFC3339)
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}
	log.Printf("Successfully reload config")
}

func sanitize(input string) string {
	// Define a regular expression to match non-alphanumeric characters
	regex := regexp.MustCompile("[^a-zA-Z0-9]+")

	// Replace non-alphanumeric characters with an empty string
	sanitizedString := regex.ReplaceAllString(input, "")

	return strings.ToLower(sanitizedString)
}

func parseCondition(condition string) ([]model.LabelValue, error) {
	ss := strings.Split(condition, "==")

	if len(ss) != 2 {
		return []model.LabelValue{}, fmt.Errorf("invalid if string: %s", condition)
	}

	return model.LabelValues{
		model.LabelValue(sanitize(ss[0])),
		model.LabelValue(sanitize(ss[1])),
	}, nil
}

func relabelling(config *Config, alert *model.Alert) {
	for _, rule := range config.Rules {
		keyAndValue, err := parseCondition(rule.If)
		if err != nil {
			log.Printf("[ERR] failed to parse condition: %s", err)
			continue
		}

		if val, ok := alert.Labels[model.LabelName(keyAndValue[0])]; ok && val == model.LabelValue(keyAndValue[1]) {
			for key, value := range rule.Then {
				alert.Labels[model.LabelName(key)] = model.LabelValue(value)
			}
		}
	}
}
