#!/bin/bash

mkdir -p /opt/cni/bin
curl -sSL https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz | tar -xvz -C /opt/cni/bin

hashi-up consul install --local \
  --version 1.10.1 \
  --server \
  --connect \
  --client-addr 0.0.0.0 \
  --advertise-addr "{{ GetInterfaceIP \"eth1\" }}" \
  --skip-start

hashi-up nomad install --local \
  --version 1.1.3 \
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

sleep 10