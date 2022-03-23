$msg = <<MSG
------------------------------------------------------
Local OpenFaas on Nomad is ready

URLS:
 * Consul     - http://192.168.56.2:8500
 * Nomad      - http://192.168.56.2:4646
 * Vault      - http://192.168.56.2:8200
 * OpenFaaS   - http://192.168.56.2:8080
 * Prometheus - http://192.168.56.2:9090

Start your local faas-provider:

  go run main.go -config contrib/local.env

------------------------------------------------------
MSG

Vagrant.configure("2") do |config|

  config.vm.box = "ubuntu/focal64"
  config.vm.network "private_network", ip: "192.168.56.2"
  config.vm.post_up_message = $msg

  config.vm.provider "virtualbox" do |vb, override|
    vb.gui = false
    vb.memory = "2048"
    vb.cpus = 2
  end

  config.vm.provision "hashi-up", type: "shell", inline: <<-SHELL
    curl -sL get.hashi-up.dev | sh
    hashi-up vault get -v 1.8.0 -d /usr/local/bin
  SHELL

  config.vm.provision :docker do |d|
    d.run 'dev-vault', image: 'vault:1.8.0', args: '-p 8200:8200 -e "VAULT_DEV_ROOT_TOKEN_ID=vagrant"'
  end

  config.vm.provision "populate-vault",     type: "shell", path: "contrib/scripts/populate-vault.sh"
  config.vm.provision "install-hashistack", type: "shell", path: "contrib/scripts/install.sh"

  config.vm.provision "jobs", type: "shell", inline: <<-SHELL
    nomad run -detach /vagrant/contrib/jobs/faas.hcl
  SHELL

end
