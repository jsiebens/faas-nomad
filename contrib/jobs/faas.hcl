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
        function_namespace="default"
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

    task "prometheus" {
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
    static_configs:
    - targets: ['localhost:8082']
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
        ]
      }
    }
  }
}