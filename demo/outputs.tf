output "nomad_addr" {
  value = "http://${digitalocean_droplet.server.ipv4_address}:4646"
}

output "consul_addr" {
  value = "http://${digitalocean_droplet.server.ipv4_address}:8500"
}

output "vault_addr" {
  value = "http://${digitalocean_droplet.server.ipv4_address}:8200"
}

output "gateway" {
  value = "http://${digitalocean_loadbalancer.public.ip}"
}

output "get_vault_token" {
  value = "ssh root@${digitalocean_droplet.server.ipv4_address} \"grep Initial /etc/vault.d/vault-keys.log | cut -c21-\""
}

output "getting_started" {
  value = <<EOF
export CONSUL_HTTP_ADDR=http://${digitalocean_droplet.server.ipv4_address}:8500
export NOMAD_ADDR=http://${digitalocean_droplet.server.ipv4_address}:4646
export VAULT_ADDR=http://${digitalocean_droplet.server.ipv4_address}:8200
export VAULT_TOKEN=$(ssh root@${digitalocean_droplet.server.ipv4_address} "grep Initial /etc/vault.d/vault-keys.log | cut -c21-")
export OPENFAAS_URL=http://${digitalocean_loadbalancer.public.ip}
vault kv get -field=value openfaas/basic-auth-password | faas-cli login -u admin --password-stdin
EOF
}