#!/bin/bash

mkdir -p /opt/cni/bin
curl -sSL https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz | tar -xvz -C /opt/cni/bin

curl -sL get.hashi-up.dev | sh

hashi-up vault get -d /usr/local/bin

echo 'Waiting for vault...'
while true
do
  START=`docker logs dev-vault 2>&1 | grep "post-unseal setup complete"`
  if [ -n "$START" ]; then
    break
  else
    sleep 2
  fi
done

POLICY_NAME=openfaas
TOKEN=vagrant
VAULT_URL=http://127.0.0.1:8200

export VAULT_ADDR=${VAULT_URL}
export VAULT_TOKEN=${TOKEN}

vault secrets disable secret/
vault secrets enable -path=openfaas -version=1 kv

curl https://nomadproject.io/data/vault/nomad-server-policy.hcl -O -s -L
curl https://nomadproject.io/data/vault/nomad-cluster-role.json -O -s -L

tee openfaas.hcl >/dev/null <<EOF
path "openfaas/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF

vault policy write openfaas openfaas.hcl
vault policy write nomad-server nomad-server-policy.hcl
vault write /auth/token/roles/nomad-cluster @nomad-cluster-role.json

hashi-up consul install --local \
  --server \
  --connect \
  --client-addr 0.0.0.0 \
  --advertise-addr "{{ GetInterfaceIP \"eth1\" }}" \
  --skip-start

hashi-up nomad install --local \
  --server \
  --client \
  --advertise "{{ GetInterfaceIP \"eth1\" }}" \
  --skip-start

sudo tee /etc/nomad.d/config/vault.hcl >/dev/null <<EOF
vault {
  enabled = true
  address = "http://127.0.0.1:8200"
  token = "vagrant"
  create_from_role = "nomad-cluster"
}
EOF

tee /etc/nomad.d/config/client.hcl >/dev/null <<EOF
client {
  network_interface = "eth1"
}
EOF

systemctl start consul
systemctl start nomad