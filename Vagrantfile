Vagrant.configure("2") do |config|
  ENV["VAGRANT_NO_PARALLEL"] = "true"
  config.vm.box = "ubuntu/focal64"

  # Master node
  config.vm.define "kube-master" do |master|
    master.vm.hostname = "kube-master"
    master.vm.network "private_network", ip: "192.168.56.10"
    master.vm.provider "virtualbox" do |vb|
      vb.memory = 4096 # You can change master node memory here. Obs: control-plane needs more than 2GB
      vb.cpus = 2 # vCPU
    end

    master.vm.provision "ansible" do |ansible|
      ansible.playbook = "ansible/playbooks/kube-master.yml"
      ansible.inventory_path = "ansible/inventories/hosts.ini"
      ansible.limit = "kube-master"
      ansible.become = true
    end
  end

  # Worker nodes
  (1..2).each do |i|
    config.vm.define "kube-node-#{i}" do |node|
      node.vm.hostname = "kube-node-#{i}"
      node.vm.network "private_network", ip: "192.168.56.#{10 + i}"
      node.vm.provider "virtualbox" do |vb|
        vb.memory = 2048 # You can change worker nodes memory here
        vb.cpus = 2 # vCPU
      end

      node.ssh.extra_args = ['-o', 'PubkeyAcceptedKeyTypes=+ssh-rsa', '-o', 'HostKeyAlgorithms=+ssh-rsa']

      node.vm.provision "ansible" do |ansible|
        ansible.playbook = "ansible/playbooks/kube-node.yml"
        ansible.inventory_path = "ansible/inventories/hosts.ini"
        ansible.limit = "kube-node-#{i}"
        ansible.become = true
      end
    end
  end
end
