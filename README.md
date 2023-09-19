# 🚀 Alert Relabeller

Alert Relabeller is a lightweight Go application designed to help you modify and forward Prometheus alerts to various Alertmanager endpoints based on custom-defined rules. This application is especially useful in scenarios where you need to perform alert relabelling and redistribution to different Alertmanager instances.

```yaml
alertmanager_urls:
  - http://alertmanager1:9093/api/v1/alerts
  - http://alertmanager2:9093/api/v1/alerts

rules:
  - if: "severity == critical"
    then:
      priority: high
  - if: "app == database"
    then:
      team: dba
```

config_last_updated_at: "2023-09-20T14:30:00Z"
alertmanager_urls: List of Alertmanager endpoints where modified alerts will be sent.
rules: Custom rules for relabelling alerts based on label conditions.
config_last_updated_at: Timestamp indicating when the configuration was last updated.

# Usage

You can start Alertmanager Relabeller by running the following command, specifying the path to your configuration file (if different from the default config.yml):

```yaml
alertmanager_urls:
  - http://localhost:9093/api/v1/alerts
rules:
  - if: name == 'argocd'
    then: 
      team: devops
      department: platform
  - if: job == 'rds-exporter'
    then: 
      team: dba
      component: aws
```

## Example rule:

```sh
./alert-relabeller -config /path/to/your/config.yml
```

The application will start a web server on port 9999 by default and listen for incoming alerts and configuration updates.

Custom Rules
You can define custom rules in the configuration file to control how alerts are relabelled. Rules consist of an if condition and a set of label modifications defined in the then section.


yaml
Copy code
- if: "severity==critical"
  then:
    priority: high
This rule checks if the severity label is equal to "critical" and changes the priority label to "high" if the condition is met.

# API Endpoints

- **/config (GET)**: Retrieve the current configuration in YAML format.
- **/config (POST)**: Update the configuration with a new one in JSON or YAML format.
- **/api/v1/alerts (POST)**: Receive incoming Prometheus alerts, apply relabelling rules, and forward them to specified Alertmanager endpoints.
