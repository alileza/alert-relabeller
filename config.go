package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

type Config struct {
	AlertmanagerURLs    []string `yaml:"alertmanager_urls" json:"alertmanager_urls"`
	Rules               []Rule   `yaml:"rules" json:"rules"`
	ConfigLastUpdatedAt string   `yaml:"config_last_updated_at" json:"config_last_updated_at"`
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
