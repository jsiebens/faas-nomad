job "faas-monitoring" {
  datacenters = ["dc1"]
  type        = "service"

  group "monitoring" {

    restart {
      attempts = 100
      delay    = "5s"
      interval = "10m"
      mode     = "delay"
    }

    network {
      mode = "bridge"
      port "prometheus" {
        static = 9090
        to = 9090
      }
    }

    service {
      port = "prometheus"
      name = "prometheus"
      tags = ["faas"]

      check {
        type     = "http"
        port     = "prometheus"
        interval = "10s"
        timeout  = "2s"
        path     = "/graph"
      }
    }

    task "alertmanager" {
      template {
        change_mode = "noop"
        destination = "local/alertmanager.yml"

        data = <<EOH
---
route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 5s
  group_interval: 10s
  repeat_interval: 30s
  receiver: scale-up
  routes:
  - match:
      service: gateway
      receiver: scale-up
      severity: major
inhibit_rules:
- source_match:
    severity: 'critical'
  target_match:
    severity: 'warning'
  equal: ['alertname', 'cluster', 'service']
receivers:
- name: 'scale-up'
  webhook_configs:
    - url: http://gateway.service.consul:8080/system/alert
      send_resolved: true
EOH
      }

      driver = "docker"

      config {
        image = "prom/alertmanager:v0.22.1"
        args = [
          "--config.file=/etc/alertmanager/alertmanager.yml"
        ]

        volumes = [
          "local/alertmanager.yml:/etc/alertmanager/alertmanager.yml",
        ]
      }
    }

    task "prometheus" {
      template {
        change_mode = "noop"
        destination = "local/alert.rules.yml"

        data = <<EOH
---
groups:
  - name: openfaas
    rules:
    - alert: service_down
      expr: up == 0
    - alert: APIHighInvocationRate
      expr: sum(rate(gateway_function_invocation_total{code="200"}[10s])) BY (function_name) > 5
      for: 5s
      labels:
        service: gateway
        severity: major
      annotations:
        description: High invocation total on "{{ "{{" }}$labels.function_name{{ "}}" }}"
        summary: High invocation total on "{{ "{{" }}$labels.function_name{{ "}}" }}"
EOH
      }

      template {
        change_mode = "noop"
        destination = "local/prometheus.yml"

        data = <<EOH
---
global:
  scrape_interval:     10s
  evaluation_interval: 10s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
    - targets: ['localhost:9090']

  - job_name: "gateway"
    scrape_interval: 5s
    consul_sd_configs:
      - server: '{{ env "attr.unique.network.ip-address" }}:8500'
        services: ["gateway-metrics"]

rule_files:
  - '/etc/prometheus/alert.rules.yml'

alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - localhost:9093
EOH
      }

      driver = "docker"

      config {
        image = "prom/prometheus:v2.14.0"
        args = [
          "--config.file=/etc/prometheus/prometheus.yml"
        ]

        volumes = [
          "local/prometheus.yml:/etc/prometheus/prometheus.yml",
          "local/alert.rules.yml:/etc/prometheus/alert.rules.yml",
        ]
      }
    }
  }
}