output "summary" {
  value = <<CONFIGURATION

The Nomad UI can be accessed at http://${digitalocean_droplet.server.ipv4_address}:4646/ui
The Consul UI can be accessed at http://${digitalocean_droplet.server.ipv4_address}:8500/ui
The Vault UI can be accessed at http://${digitalocean_droplet.server.ipv4_address}:8200/ui
The OpenFaaS UI can be accessed at http://${digitalocean_loadbalancer.public.ip}/ui

CLI environment variables:

export CONSUL_HTTP_ADDR=http://${digitalocean_droplet.server.ipv4_address}:8500
export NOMAD_ADDR=http://${digitalocean_droplet.server.ipv4_address}:4646
export VAULT_ADDR=http://${digitalocean_droplet.server.ipv4_address}:8200
export OPENFAAS_URL=http://${digitalocean_loadbalancer.public.ip}
export VAULT_TOKEN=$(ssh root@${digitalocean_droplet.server.ipv4_address} "grep Initial /etc/vault.d/vault-keys.log | cut -c21-")

Authenticate with faas-cli:

vault kv get -field=value openfaas/basic-auth-password | faas-cli login -u admin --password-stdin

CONFIGURATION
}

output "env" {
  value = <<CONFIGURATION
export CONSUL_HTTP_ADDR=http://${digitalocean_droplet.server.ipv4_address}:8500
export NOMAD_ADDR=http://${digitalocean_droplet.server.ipv4_address}:4646
export VAULT_ADDR=http://${digitalocean_droplet.server.ipv4_address}:8200
export OPENFAAS_URL=http://${digitalocean_loadbalancer.public.ip}
export VAULT_TOKEN=$(ssh root@${digitalocean_droplet.server.ipv4_address} "grep Initial /etc/vault.d/vault-keys.log | cut -c21-")
CONFIGURATION
}