#!/bin/bash

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

TOKEN=vagrant
VAULT_URL=http://127.0.0.1:8200

export VAULT_ADDR=${VAULT_URL}
export VAULT_TOKEN=${TOKEN}

vault secrets disable secret/
vault secrets enable -path=openfaas-fn -version=1 kv

curl https://nomadproject.io/data/vault/nomad-server-policy.hcl -O -s -L
curl https://nomadproject.io/data/vault/nomad-cluster-role.json -O -s -L

tee openfaas-fn.hcl >/dev/null <<EOF
path "openfaas-fn/*" {
  capabilities = ["read"]
}
EOF

vault policy write openfaas-fn openfaas-fn.hcl
vault policy write nomad-server nomad-server-policy.hcl
vault write /auth/token/roles/nomad-cluster @nomad-cluster-role.json