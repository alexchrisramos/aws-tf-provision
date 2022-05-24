#tf to provision cluster in aws
#from https://github.com/mesosphere/cluster-api-provider-preprovisioned

#aws cli credentials
eval "$(maws login 337834004759_Mesosphere-PowerUser)"
AWS_PROFILE=337834004759_Mesosphere-PowerUser

#instruction below is same on instruction in the github repo
export TF_VAR_tags='{"owner":'\"$(whoami)\"',"expiration":"10h"}'
export TF_VAR_node_ami=ami-0fe1f6c5052e9fc01
export TF_VAR_ssh_key_name=alex-keypair

#deploy
terraform -chdir=infra/ init
terraform -chdir=infra/ apply -auto-approve