job "faas" {
  datacenters = ["dc1"]
  type        = "service"

  group "faas" {

    restart {
      attempts = 100
      delay    = "5s"
      interval = "10m"
      mode     = "delay"
    }

    network {
      mode = "bridge"
      port "gateway" {
        static = 8080
        to = 8080
      }
      port "metrics" {
        static = 8082
        to = 8082
      }
      port "prometheus" {
        static = 9090
        to = 9090
      }
    }

    service {
      name = "faas-gateway"
      port = "gateway"

      check {
        type = "tcp"
        port = "gateway"
        interval = "5s"
        timeout = "2s"
      }
    }

    task "nats" {
      driver = "docker"
      config {
        image = "docker.io/library/nats-streaming:0.11.2"
        entrypoint = ["/nats-streaming-server"]
        args = [
          "-m",
          "8222",
          "--store=memory",
          "--cluster_id=faas-cluster"
        ]
        ports = ["nats"]
      }
      resources {
        cpu    = 50
        memory = 50
      }
    }

    task "gateway" {
      driver = "docker"

      config {
        image = "ghcr.io/openfaas/gateway:0.20.11"
        ports = ["gateway"]
      }

      env {
        basic_auth="false"
        functions_provider_url="http://192.168.50.1:8080/"
        direct_functions="false"
        function_namespace="openfaas-fn"
        read_timeout="60s"
        write_timeout="60s"
        upstream_timeout="65s"
        faas_prometheus_host="localhost"
        faas_prometheus_port="9090"
        faas_nats_address="localhost"
        faas_nats_port="4222"
        secret_mount_path="/secrets"
        scale_from_zero="true"
      }
      resources {
        cpu    = 50
        memory = 50
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
    - url: http://localhost:8080/system/alert
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

  - job_name: 'gateway'
    scrape_interval: 5s
    static_configs:
    - targets: ['localhost:8082']

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