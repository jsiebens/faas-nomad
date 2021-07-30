job "faas-nats" {
  datacenters = ["dc1"]
  type        = "service"

  group "nats" {

    restart {
      attempts = 100
      delay    = "5s"
      interval = "10m"
      mode     = "delay"
    }

    network {
      port "client" {
        static = 4222
        to     = 4222
      }
      port "monitoring" {
        static = 8222
        to     = 8222
      }
    }

    task "nats" {
      driver = "docker"
      config {
        image = "docker.io/library/nats-streaming:0.22.0"
        entrypoint = [
        "/nats-streaming-server"]
        args = [
          "-m",
          "8222",
          "--store=memory",
          "--cluster_id=faas-cluster"
        ]
        ports = ["client", "monitoring"]
      }
      resources {
        cpu    = 50
        memory = 50
      }

      service {
        port = "client"
        name = "nats"
        tags = ["faas"]

        check {
          type     = "http"
          port     = "monitoring"
          path     = "/connz"
          interval = "5s"
          timeout  = "2s"
        }
      }
    }
  }
}