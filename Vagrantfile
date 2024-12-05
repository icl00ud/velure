Vagrant.configure("2") do |config|
    # Configurações globais
    config.vm.box = "ubuntu/focal64"
  
    # Nó mestre
    config.vm.define "kube-master" do |master|
      master.vm.hostname = "kube-master"
      master.vm.network "private_network", ip: "192.168.56.10"
  
      master.vm.provision "ansible" do |ansible|
        ansible.playbook = "ansible/playbooks/kube-master.yml"
        ansible.inventory_path = "ansible/inventories/hosts.ini"
        ansible.roles_path = "ansible/roles"
        ansible.limit = "kube-master"
        ansible.become = true
      end
    end
  
    # Nós workers
    (1..2).each do |i|
      config.vm.define "kube-node-#{i}" do |node|
        node.vm.hostname = "kube-node-#{i}"
        node.vm.network "private_network", ip: "192.168.56.1#{i}"
  
        node.vm.provision "ansible" do |ansible|
          ansible.playbook = "ansible/playbooks/kube-node.yml"
          ansible.inventory_path = "ansible/inventories/hosts.ini"
          ansible.roles_path = "ansible/roles"
          ansible.limit = "kube-node-#{i}"
          ansible.become = true
        end
      end
    end
  end
  