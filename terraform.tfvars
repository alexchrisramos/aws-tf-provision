controlplane_count = 3
worker_count = 4
bastion_host_count = 1
registry_host_count = 1
node_ami = "ami-0e87ed6656c4945c7"
bastion_ami = "ami-0e87ed6656c4945c7"
registry_ami = "ami-0e87ed6656c4945c7"
ssh_key_name = "alex-keypair"
ssh_username = "ubuntu"
worker_instance_type = "m5.2xlarge"
root_volume_size = 80
mykeypair = alex-keypair.pem
control_plane_instance_type = "m5.2xlarge"
is_airgap = false