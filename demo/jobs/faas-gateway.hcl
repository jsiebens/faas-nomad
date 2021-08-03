job "faas-gateway" {
  datacenters = ["dc1"]
  type        = "service"

  group "queue-worker" {
    task "queue-worker" {
      driver = "docker"

      config {
        image = "ghcr.io/openfaas/queue-worker:0.12.4"
        volumes = [
          "secrets/basic-auth-user:/run/secrets/basic-auth-user",
          "secrets/basic-auth-password:/run/secrets/basic-auth-password"
        ]
      }

      vault {
        policies = ["openfaas"]
      }

      template {
        data        = "{{with secret \"openfaas/basic-auth-password\"}}{{.Data.value}}{{end}}"
        destination = "secrets/basic-auth-password"
      }

      template {
        data        = "{{with secret \"openfaas/basic-auth-user\"}}{{.Data.value}}{{end}}"
        destination = "secrets/basic-auth-user"
      }

      env {
        basic_auth           = "true"
        secret_mount_path    = "/run/secrets"
        gateway_invoke       = "true"
        faas_gateway_address = "gateway.service.consul"
        faas_gateway_port    = "8080"
        faas_nats_address    = "nats.service.consul"
        faas_nats_port       = "4222"
      }

      resources {
        cpu    = 50
        memory = 128
      }
    }
  }

  group "gateway" {
    count = 2

    network {
      mode = "bridge"
      port "gateway" {
        static = 8080
        to     = 8080
      }
      port "metrics" {
        static = 8082
        to     = 8082
      }
    }

    service {
      name = "faas-gateway"
      port = "gateway"
      tags = ["faas", "gateway"]

      check {
        type     = "tcp"
        port     = "gateway"
        interval = "5s"
        timeout  = "2s"
      }
    }

    service {
      name = "faas-gateway"
      port = "metrics"
      tags = ["faas", "metrics"]

      check {
        type     = "tcp"
        port     = "metrics"
        interval = "5s"
        timeout  = "2s"
      }
    }

    task "gateway" {
      driver = "docker"

      config {
        image = "ghcr.io/openfaas/gateway:0.21.1"
        ports = ["gateway"]
        volumes = [
          "secrets/basic-auth-user:/run/secrets/basic-auth-user",
          "secrets/basic-auth-password:/run/secrets/basic-auth-password"
        ]
      }

      vault {
        policies = ["openfaas"]
      }

      template {
        data        = "{{with secret \"openfaas/basic-auth-password\"}}{{.Data.value}}{{end}}"
        destination = "secrets/basic-auth-password"
      }

      template {
        data        = "{{with secret \"openfaas/basic-auth-user\"}}{{.Data.value}}{{end}}"
        destination = "secrets/basic-auth-user"
      }

      env {
        basic_auth             = "true"
        auth_proxy_url         = "http://localhost:7000/validate"
        auth_pass_body         = "false"
        secret_mount_path      = "/run/secrets"
        functions_provider_url = "http://127.0.0.1:8081/"
        direct_functions       = "false"
        function_namespace     = "default"
        read_timeout           = "60s"
        write_timeout          = "60s"
        upstream_timeout       = "65s"
        faas_prometheus_host   = "prometheus.service.consul"
        faas_prometheus_port   = "9090"
        faas_nats_address      = "nats.service.consul"
        faas_nats_port         = "4222"
        scale_from_zero        = "true"
      }

      resources {
        cpu    = 50
        memory = 128
      }
    }

    task "provider" {
      driver = "docker"

      config {
        image = "ghcr.io/jsiebens/faas-nomad:latest"
        ports = ["provider"]
      }

      vault {
        policies      = ["openfaas"]
        change_mode   = "signal"
        change_signal = "SIGUSR1"
      }

      env {
        port        = "8081"
        vault_addr  = "http://vault.service.consul:8200"
        nomad_addr  = "http://nomad.service.consul:4646"
        consul_addr = "http://consul.service.consul:8500"
        job_purge   = "true"
      }
    }

    task "basic-auth" {
      driver = "docker"

      config {
        image = "ghcr.io/openfaas/basic-auth:0.21.1"
        volumes = [
          "secrets/basic-auth-user:/run/secrets/basic-auth-user",
          "secrets/basic-auth-password:/run/secrets/basic-auth-password"
        ]
      }

      vault {
        policies = ["openfaas"]
      }

      template {
        data        = "{{with secret \"openfaas/basic-auth-password\"}}{{.Data.value}}{{end}}"
        destination = "secrets/basic-auth-password"
      }

      template {
        data        = "{{with secret \"openfaas/basic-auth-user\"}}{{.Data.value}}{{end}}"
        destination = "secrets/basic-auth-user"
      }

      env {
        port              = "7000"
        secret_mount_path = "/run/secrets"
        user_filename     = "basic-auth-user"
        pass_filename     = "basic-auth-password"
      }
    }
  }

}