#! /bin/bash
set -euo pipefail

function net_getDockerAddress() {
  ip -4 address show "docker0" | awk '/inet / { print $2 }' | cut -d/ -f1
}

# Disable interactive apt prompts
export DEBIAN_FRONTEND=noninteractive
echo 'debconf debconf/frontend select Noninteractive' | sudo debconf-set-selections

apt-get update
apt-get install -y dnsmasq curl unzip docker.io

mkdir -p /opt/cni/bin
curl -sSL https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz | tar -xvz -C /opt/cni/bin

curl -sL get.hashi-up.dev | sh

DOCKER_BRIDGE_IP_ADDRESS=$(net_getDockerAddress)

# dnsmasq config
echo "DNSStubListener=no" | sudo tee -a /etc/systemd/resolved.conf
echo "server=/consul/127.0.0.1#8600" > /etc/dnsmasq.d/10-consul
echo "server=8.8.8.8" > /etc/dnsmasq.d/99-default
sudo mv /etc/resolv.conf /etc/resolv.conf.orig
grep -v "nameserver" /etc/resolv.conf.orig | grep -v -e"^#" | grep -v -e '^$' | sudo tee /etc/resolv.conf
echo "nameserver 127.0.0.1" | sudo tee -a /etc/resolv.conf
sudo systemctl restart systemd-resolved
sudo systemctl restart dnsmasq
sudo systemctl restart docker

# Add Docker bridge network IP to /etc/resolv.conf (at the top)
echo "nameserver $DOCKER_BRIDGE_IP_ADDRESS" | sudo tee /etc/resolv.conf.new
cat /etc/resolv.conf | sudo tee --append /etc/resolv.conf.new
sudo mv /etc/resolv.conf.new /etc/resolv.conf

hashi-up consul install --local \
  --connect \
  --client-addr 0.0.0.0 \
  --advertise-addr "{{ GetInterfaceIP \"eth1\" }}" \
  --retry-join ${server_ip}

hashi-up nomad install --local \
  --client \
  --advertise "{{ GetInterfaceIP \"eth1\" }}" \
  --skip-start

sudo tee /etc/nomad.d/config/vault.hcl >/dev/null <<EOF
vault {
  enabled = true
  address = "http://${server_ip}:8200"
}
EOF

systemctl start nomad

echo 'debconf debconf/frontend select Dialog' | sudo debconf-set-selections