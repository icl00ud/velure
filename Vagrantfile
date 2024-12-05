Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/focal64"

  # Configuração global de SSH
  config.vm.provider "virtualbox" do |vb|
    vb.memory = "2048"
    vb.cpus = 2
    vb.name = "kube-master"
    vb.gui = false
  end

  # Definir o nó mestre
  config.vm.define "kube-master" do |master|
    master.vm.hostname = "kube-master"

    # Provisionamento com Ansible
    master.vm.provision "ansible" do |ansible|
      ansible.playbook = "ansible/playbooks/kube-master.yml"
      ansible.inventory_path = "ansible/inventories/hosts.ini"
      ansible.become = true
    end
  end

  # Definir workers
  (1..2).each do |i|
    config.vm.define "kube-node-#{i}" do |node|
      node.vm.hostname = "kube-node-#{i}"

      # Provisionamento com Ansible
      node.vm.provision "ansible" do |ansible|
        ansible.playbook = "ansible/playbooks/kube-node.yml"
        ansible.inventory_path = "ansible/inventories/hosts.ini"
        ansible.become = true
      end
    end
  end
end
