# 🚀 Alert Relabeller

Alert Relabeller is a lightweight Go application designed to help you modify and forward Prometheus alerts to various Alertmanager endpoints based on custom-defined rules. This application is especially useful in scenarios where you need to perform alert relabelling and redistribution to different Alertmanager instances.

It works best as a sidecar to your Alertmanager, would be even nicer if it could be attached as a Prometheus Alertmanager feature 🤞

# Problem statement

In many cases, organizations share common Prometheus alert rules across their teams. However, when multiple teams are involved, routing these shared alerts to the appropriate receivers can be challenging. Ideally, you would configure proper labels in your alerts to ensure they are routed correctly. The issue arises because common alerts cannot have predefined labels since the labels may vary depending on the team responsible for the alert. This is where an alert relabeler becomes invaluable. It allows you to intercept alerts and modify their labels based on the desired configuration, which could be derived from your organization's data or even ArgoCD labels.

# How it works as Alertmanager sidecar

<img width="1021" alt="alert-relabeller" src="https://github.com/alileza/alert-relabeller/assets/1962129/45cd08ec-abff-4c2b-81c5-8cb04dd8ba3b">



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
  ghcr.io/alileza/alert-relabeller:v0.2.1 \
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

## /config (GET)
- **Description:** Retrieve the current configuration in in JSON or YAML format, depends on Content-Type header. (application/json OR text/yaml)
- **HTTP Method:** GET

## /config (POST)
- **Description:** Update the configuration with a new one in JSON or YAML format, depends on Content-Type header. (application/json OR text/yaml)
- **HTTP Method:** POST

## /api/v2/alerts (POST)
- **Description:** Receive incoming Prometheus alerts, apply relabeling rules, and forward them to specified Alertmanager endpoint.
- **HTTP Method:** POST

## /-/ready and /-/healthy
- **Description:** These endpoints are used for readiness and health checks.
- **HTTP Methods:** GET
- **Response:** It will be forwarded directly to Alertmanager endpoint.
