#tf to provision cluster in aws
#from https://github.com/mesosphere/cluster-api-provider-preprovisioned

#aws cli credentials

#instruction below is same on instruction in the github repo
export TF_VAR_tags='{"owner":'\"$(whoami)\"',"expiration":"10h"}'
export TF_VAR_node_ami=ami-000d6375f955d3d80
export TF_VAR_ssh_key_name=alex-keypair

#deploy
terraform -chdir=deploy/infra init
terraform -chdir=deploy/infra apply -auto-approve