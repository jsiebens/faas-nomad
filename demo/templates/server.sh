#! /bin/bash
set -euo pipefail

# Disable interactive apt prompts
export DEBIAN_FRONTEND=noninteractive
echo 'debconf debconf/frontend select Noninteractive' | sudo debconf-set-selections

apt-get update
apt-get install -y dnsmasq curl unzip

curl -sL get.hashi-up.dev | sh

# dnsmasq config
echo "DNSStubListener=no" | sudo tee -a /etc/systemd/resolved.conf
echo "server=/consul/127.0.0.1#8600" > /etc/dnsmasq.d/10-consul
echo "server=8.8.8.8" > /etc/dnsmasq.d/99-default
sudo mv /etc/resolv.conf /etc/resolv.conf.orig
grep -v "nameserver" /etc/resolv.conf.orig | grep -v -e"^#" | grep -v -e '^$' | sudo tee /etc/resolv.conf
echo "nameserver 127.0.0.1" | sudo tee -a /etc/resolv.conf
sudo systemctl restart systemd-resolved
sudo systemctl restart dnsmasq

# install and start Consul
hashi-up consul install --local \
  --server \
  --connect \
  --client-addr 0.0.0.0 \
  --advertise-addr "{{ GetInterfaceIP \"eth1\" }}"

# install and start Vault
hashi-up vault install \
  --local \
  --storage consul

sleep 10

export VAULT_ADDR=http://localhost:8200

vault operator init -key-shares=1 -key-threshold=1 > /etc/vault.d/vault-keys.log

UNSEAL_KEY=$(grep Unseal /etc/vault.d/vault-keys.log | cut -c15-)
export VAULT_TOKEN=$(grep Initial /etc/vault.d/vault-keys.log | cut -c21-)

vault operator unseal $${UNSEAL_KEY}
vault secrets disable secret/
vault secrets enable -path=openfaas -version=1 kv
vault secrets enable -path=openfaas-fn -version=1 kv

vault kv put openfaas/basic-auth-password value=${token}
vault kv put openfaas/basic-auth-user value=admin

curl https://nomadproject.io/data/vault/nomad-server-policy.hcl -O -s -L
curl https://nomadproject.io/data/vault/nomad-cluster-role.json -O -s -L

tee openfaas.hcl >/dev/null <<EOF
path "openfaas/*" {
  capabilities = ["read", "list"]
}
path "openfaas-fn/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF

tee openfaas-fn.hcl >/dev/null <<EOF
path "openfaas-fn/*" {
  capabilities = ["read"]
}
EOF

vault policy write openfaas openfaas.hcl
vault policy write openfaas-fn openfaas-fn.hcl
vault policy write nomad-server nomad-server-policy.hcl
vault write /auth/token/roles/nomad-cluster @nomad-cluster-role.json

# install and start Nomad
hashi-up nomad install --local \
  --server \
  --advertise "{{ GetInterfaceIP \"eth1\" }}" \
  --skip-start

sudo tee /etc/nomad.d/config/vault.hcl >/dev/null <<EOF
vault {
  enabled = true
  address = "http://127.0.0.1:8200"
  token = "$${VAULT_TOKEN}"
  create_from_role = "nomad-cluster"
}
EOF

systemctl start nomad

echo 'debconf debconf/frontend select Dialog' | sudo debconf-set-selections