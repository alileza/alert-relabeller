# 🚀 Alert Relabeller

Alert Relabeller is a lightweight Go application designed to help you modify and forward Prometheus alerts to various Alertmanager endpoints based on custom-defined rules. This application is especially useful in scenarios where you need to perform alert relabelling and redistribution to different Alertmanager instances.

```yaml
rules:
  - if: "severity == critical"
    then:
      priority: high
  - if: "app == database"
    then:
      team: dba
```

# How it works as Alertmanager sidecar

<img width="1021" alt="Screenshot 2023-09-20 at 03 27 39" src="https://github.com/distrobeat/infrastructure/assets/1962129/ead035e0-d03b-4336-9bc1-9618fb04c741">


# Usage

You can start Alertmanager Relabeller by running the following command, specifying the path to your configuration file (if different from the default config.yml):

```yaml
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


**alertmanager_urls:** List of Alertmanager endpoints where modified alerts will be sent.
**rules**: Custom rules for relabelling alerts based on label conditions.

## Example run:

### With binary

```sh
./alert-relabeller -config /path/to/your/config.yml -alertmanager-url localhost:9093 -port 9999
```

### With Docker

```sh
docker run -d \
  -p 9999:9999 \
  -v /path/to/your/config.yml:/app/config.yml \
  ghcr.io/alileza/alert-relabeller:v0.1.1 \
  -config /app/config.yml -alertmanager-url http://alertmanager:9093
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
