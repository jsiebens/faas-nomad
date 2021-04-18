Vagrant.configure("2") do |config|

  config.vm.box = "hashicorp/bionic64"
  config.vm.network "private_network", ip: "192.168.50.2"

  config.vm.provider "virtualbox" do |vb, override|
    vb.gui = false
    vb.memory = "2048"
    vb.cpus = 2
  end

  config.vm.provision :docker do |d|
    d.run 'dev-vault', image: 'vault:1.6.2',
      args: '-p 8200:8200 -e "VAULT_DEV_ROOT_TOKEN_ID=vagrant"'
  end

  config.vm.provision "shell", path: "contrib/scripts/install.sh"

end
