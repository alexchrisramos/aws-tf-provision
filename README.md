tf to provision cluster in aws
from https://github.com/mesosphere/cluster-api-provider-preprovisioned
ansible c/o Pogz

# Set your aws cli credentials

# Set the TF variables
export TF_VAR_tags='{"owner":'\"$(whoami)\"',"expiration":"16h"}'
export TF_VAR_worker_count=1
export TF_VAR_extra_worker_count=0
export TF_VAR_node_ami=ami-0b52d4ef0c06bb1ac
export TF_VAR_ssh_key_name=<yourKeyPair>
export TF_VAR_ssh_username=centos
export TF_VAR_worker_instance_type=m5.2xlarge
export TF_VAR_root_volume_size=200
export mykeypair=<yourKeyPair>.pem

# set ssh agent
eval `ssh-agent`
ssh-add ${mykeypair}
ssh-add -l

# Apply
terraform -chdir=infra/ init
terraform -chdir=infra/ plan
terraform -chdir=infra/ apply -auto-approve
