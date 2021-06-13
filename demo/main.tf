terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "2.8.0"
    }
  }
}

locals {
  name = random_pet.name.id
}

resource "random_pet" "name" {
}

resource "random_password" "token" {
  length  = 16
  special = false
}

resource "digitalocean_tag" "main" {
  name = local.name
}

resource "digitalocean_tag" "server" {
  name = format("%s-server", local.name)
}

resource "digitalocean_tag" "client" {
  name = format("%s-client", local.name)
}

resource "digitalocean_vpc" "this" {
  name     = local.name
  region   = var.region
  ip_range = var.ip_range
}

resource "digitalocean_droplet" "server" {
  image     = "ubuntu-20-04-x64"
  name      = format("%s-server", local.name)
  region    = var.region
  size      = "s-1vcpu-1gb"
  tags      = [digitalocean_tag.main.id, digitalocean_tag.server.id]
  user_data = templatefile("${path.module}/templates/server.sh", { token = random_password.token.result })
  vpc_uuid  = digitalocean_vpc.this.id
  ssh_keys  = [var.ssh_key]
}

resource "digitalocean_droplet" "client" {
  count     = 3
  image     = "ubuntu-20-04-x64"
  name      = format("%s-client-0%s", local.name, count.index + 1)
  region    = var.region
  size      = "s-1vcpu-2gb"
  tags      = [digitalocean_tag.main.id, digitalocean_tag.client.id]
  user_data = templatefile("${path.module}/templates/client.sh", { server_ip = digitalocean_droplet.server.ipv4_address_private })
  vpc_uuid  = digitalocean_vpc.this.id
  ssh_keys  = [var.ssh_key]
}

resource "digitalocean_loadbalancer" "public" {
  name     = local.name
  region   = var.region
  vpc_uuid = digitalocean_vpc.this.id

  forwarding_rule {
    entry_port     = 80
    entry_protocol = "http"

    target_port     = 8080
    target_protocol = "http"
  }

  healthcheck {
    port     = 8080
    protocol = "tcp"
  }

  droplet_tag = digitalocean_tag.client.id
}

module "my_ip_address" {
  source  = "matti/resource/shell"
  command = "curl https://ipinfo.io/ip"
}

resource "digitalocean_firewall" "gateway" {
  name = format("%s-gateway", local.name)

  tags = [digitalocean_tag.client.id]

  inbound_rule {
    protocol         = "tcp"
    port_range       = "8080"
    source_addresses = ["0.0.0.0/0"]
  }
}

resource "digitalocean_firewall" "internal" {
  name = local.name

  tags = [digitalocean_tag.main.id]

  inbound_rule {
    protocol         = "tcp"
    port_range       = "1-65535"
    source_addresses = ["${module.my_ip_address.stdout}/32"]
  }

  inbound_rule {
    protocol    = "tcp"
    port_range  = "1-65535"
    source_tags = [digitalocean_tag.main.id]
  }

  inbound_rule {
    protocol    = "udp"
    port_range  = "1-65535"
    source_tags = [digitalocean_tag.main.id]
  }

  inbound_rule {
    protocol    = "icmp"
    source_tags = [digitalocean_tag.main.id]
  }

  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}