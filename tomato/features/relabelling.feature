Feature: test alert relabelling behaviour

    Scenario: test update config API
        Given "httpcli" set request header key "Content-Type" with value "application/json"
        Given "httpcli" send request to "GET /config"
        Then "httpcli" response code should be 200
        And "httpcli" response header "Content-Type" should be "application/json"
        And "httpcli" response body should equal
        """
        {
            "rules": [
                {
                    "if": "name == 'argocd'",
                    "then": {
                        "department": "platform",
                        "team": "devops"
                    }
                },
                {
                    "if": "job == 'rds-exporter'",
                    "then": {
                        "component": "aws",
                        "team": "dba"
                    }
                }
            ],
            "config_last_updated_at": "*"
        }
        """

    Scenario: test alert relabeling
        Given "httpcli" send request to "POST /api/v1/alerts" with body
        """
        [
            {
                "annotations": {
                    "summary": "High request latency"
                },
                "endsAt": "2023-09-19T21:40:24.682Z",
                "startsAt": "2023-09-19T21:36:24.682Z",
                "generatorURL": "http://20010cdf93ad:9090/graph?g0.expr=up+%3D%3D+1\u0026g0.tab=1",
                "labels": {
                    "alertname": "HighRequestLatency",
                    "instance": "localhost:9090",
                    "job": "prometheus",
                    "severity": "critical",
                    "name": "argocd"
                }
            }
        ]
        """
        Then "httpcli" response code should be 200
        Then "alertmanager" with path "POST /api/v1/alerts" request count should be 1
        
