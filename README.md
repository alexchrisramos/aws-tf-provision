# Set your aws cli credentials

# Set the TF variables
```
export TF_VAR_tags='{"owner":'\"$(whoami)\"',"expiration":"4h"}'
export TF_VAR_controlplane_count=3
export TF_VAR_worker_count=4
export TF_VAR_bastion_host_count=1
export TF_VAR_registry_host_count=1
export TF_VAR_node_ami=ami-0e87ed6656c4945c7
export TF_VAR_bastion_ami=ami-0e87ed6656c4945c7
export TF_VAR_registry_ami=ami-0e87ed6656c4945c7
export TF_VAR_ssh_key_name=alex-keypair
export TF_VAR_ssh_username=ubuntu
export TF_VAR_worker_instance_type=m5.2xlarge
export TF_VAR_root_volume_size=80
export mykeypair=alex-keypair.pem
export TF_VAR_control_plane_instance_type=m5.2xlarge
export TF_VAR_is_airgap=true
```

# set ssh agent
```
eval `ssh-agent`
ssh-add ${mykeypair}
ssh-add -l
```

# Applyop
```
terraform -chdir=infra/ init
terraform -chdir=infra/ apply -auto-approve
```

# Update manifests/inventory.yaml
# execute ansible playbook to provision cluster hosts with mount volumes
```
ansible-playbook ansible/provision-cluster-hosts.yaml -i ansible/inventory.ini -u ${TF_VAR_ssh_username} --private-key ${mykeypair} --ssh-common-args='-o StrictHostKeyChecking=no'
```

# execute ansible playbook to provision bastion
```
ansible-playbook ansible/provision-bastion-host.yaml -i ansible/inventory.ini -u ${TF_VAR_ssh_username} --private-key ${mykeypair} --ssh-common-args='-o StrictHostKeyChecking=no'
```

# execute ansible playbook to provision registry
```
ansible-playbook ansible/provision-registry-host.yaml -i ansible/inventory.ini -u ${TF_VAR_ssh_username} --private-key ${mykeypair} --ssh-common-args='-o StrictHostKeyChecking=no'
```