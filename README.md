# Set your aws cli credentials

# Set the TF variables
Environment variable
```
export TF_VAR_tags='{"owner":'\"$(whoami)\"',"expiration":"4h"}'
```
Also edit the terraform.tfvars for the other intended values.

# Apply
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