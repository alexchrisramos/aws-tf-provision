terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = "us-west-2"
}

variable "aws_availability_zones" {
  type        = list
  description = "Availability zones to be used"
  default     = ["us-west-2c"]
}

variable "vpc_cidr" {
  description = "The CIDR used for vpc"
  default     = "10.0.0.0/16"
}

variable "ssh_key_name" {
  description = "The SSH key pair to use for the public instance"
}

variable "is_airgap" {
  description = "Will the cluster be airgapped or not - True or False"
  type        = bool
  default     = false
}


variable "tags" {
  description = "Map of tags to add to all resources"
  type        = map
  default     = {}
}

variable "egress_cidr" {
  description = "The CIDR used in the egress security group"
  default     = "0.0.0.0/0"
}

variable "node_ami" {
  description = "The AMI for the cluster nodes"
}

variable "bastion_ami" {
  description = "The AMI for the cluster nodes"
}

variable "registry_ami" {
  description = "The AMI for the cluster nodes"
}

variable "control_plane_instance_type" {
  description = "The instance type for the control-plane nodes"
  default = "m5.2xlarge"
}

variable "worker_instance_type" {
  description = "The instance type for the worker nodes"
  default = "m5.2xlarge"
}

variable "bastion_instance_type" {
  description = "The instance type for the worker nodes"
  default = "t2.medium"
}

variable "registry_instance_type" {
  description = "The instance type for the registry host"
  default = "t2.medium"
}


variable "controlplane_count" {
  description = "The number of worker nodes"
  default = 3
}


variable "worker_count" {
  description = "The number of worker nodes"
  default = 1
}

variable "bastion_host_count" {
  description = "The number of bastion hosts"
  default = 0
}

variable "registry_host_count" {
  description = "The number of bastion hosts"
  default = 0
}

variable "root_volume_size" {
  description = "Instance root volume size"
  default = 100
}

variable "ssh_username" {
  description = "The user for connecting to the instance over ssh"
  default = "centos"
}

locals {
  public_subnet_range         = var.vpc_cidr
  public_ip                   = var.is_airgap ? "false" : "true"
  controlplane_count          = var.is_airgap ? 0 : var.controlplane_count
  worker_count                = var.is_airgap ? 0 : var.worker_count
  controlplane_airgap_count   = var.is_airgap ? var.controlplane_count : 0
  worker_airgap_count         = var.is_airgap ? var.worker_count : 0
}


resource "aws_vpc" "konvoy_vpc" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = var.tags
}

resource "aws_internet_gateway" "konvoy_gateway" {
  vpc_id = aws_vpc.konvoy_vpc.id

  tags = var.tags
}

resource "aws_subnet" "konvoy_public" {
  vpc_id                  = aws_vpc.konvoy_vpc.id
  cidr_block              = local.public_subnet_range
  map_public_ip_on_launch = true
  availability_zone       = var.aws_availability_zones[0]

  tags = var.tags
}

  resource "aws_route_table" "konvoy_public_rt" {
    vpc_id = aws_vpc.konvoy_vpc.id

    route {
      cidr_block = "0.0.0.0/0"
      gateway_id = aws_internet_gateway.konvoy_gateway.id
    }

    tags = var.tags
  }

  resource "aws_route_table_association" "konvoy_public_rta" {
    subnet_id      = aws_subnet.konvoy_public.id
    route_table_id = aws_route_table.konvoy_public_rt.id
  }

resource "aws_security_group" "konvoy_ssh" {
  description = "Allow inbound SSH for Konvoy."
  vpc_id      = aws_vpc.konvoy_vpc.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

resource "aws_security_group" "konvoy_private" {
  description = "Allow all communication between instances"
  vpc_id      = aws_vpc.konvoy_vpc.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  tags = var.tags
}

resource "aws_security_group" "konvoy_egress" {
  description = "Allow all egress communication."
  vpc_id      = aws_vpc.konvoy_vpc.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [var.egress_cidr]
  }

  tags = var.tags
}

resource "aws_security_group" "public_facing_instances" {
  description = "Allow inbound http and https for public facing instances."
  vpc_id      = aws_vpc.konvoy_vpc.id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}


resource "aws_security_group" "konvoy_control_plane" {
  description = "Allow inbound 6443 for CP."
  vpc_id      = aws_vpc.konvoy_vpc.id

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

resource "aws_instance" "control_plane" {
  count                       = local.controlplane_count
  vpc_security_group_ids      = [aws_security_group.konvoy_ssh.id, aws_security_group.konvoy_private.id, aws_security_group.konvoy_egress.id]
  subnet_id                   = aws_subnet.konvoy_public.id
  key_name                    = var.ssh_key_name
  ami                         = var.node_ami
  instance_type               = var.control_plane_instance_type
  availability_zone           = var.aws_availability_zones[0]
  source_dest_check           = "false"
  associate_public_ip_address = local.public_ip

  tags = var.tags

  root_block_device {
    volume_size = var.root_volume_size
  }

  provisioner "remote-exec" {
    inline = [
      "echo ok"
    ]

    connection {
      type = "ssh"
      user = var.ssh_username
      agent = true
      host = self.public_dns
      timeout = "15m"
    }
  }
}

resource "aws_instance" "control_plane_airgap" {
  count                       = local.controlplane_airgap_count
  vpc_security_group_ids      = [aws_security_group.konvoy_ssh.id, aws_security_group.konvoy_private.id, aws_security_group.konvoy_egress.id]
  subnet_id                   = aws_subnet.konvoy_public.id
  key_name                    = var.ssh_key_name
  ami                         = var.node_ami
  instance_type               = var.control_plane_instance_type
  availability_zone           = var.aws_availability_zones[0]
  source_dest_check           = "false"
  associate_public_ip_address = local.public_ip

  tags = var.tags

  root_block_device {
    volume_size = var.root_volume_size
  }

  /*provisioner "remote-exec" {
    inline = [
      "Skipping ssh test since airgap is enabled"
    ]

    connection {
      type = "ssh"
      user = var.ssh_username
      agent = true
      host = self.public_dns
      timeout = "15m"
    }
  }*/
}

resource "aws_instance" "worker" {
  count                       = local.worker_count
  vpc_security_group_ids      = [aws_security_group.konvoy_ssh.id, aws_security_group.konvoy_private.id, aws_security_group.konvoy_egress.id, aws_security_group.public_facing_instances.id]
  subnet_id                   = aws_subnet.konvoy_public.id
  key_name                    = var.ssh_key_name
  ami                         = var.node_ami
  instance_type               = var.worker_instance_type
  availability_zone           = var.aws_availability_zones[0]
  source_dest_check           = "false"
  associate_public_ip_address = local.public_ip

  tags = var.tags

  root_block_device {
    volume_size = var.root_volume_size
  }

  provisioner "remote-exec" {
    inline = [
      "echo ok"
    ]

    connection {
      type = "ssh"
      user = var.ssh_username
      agent = true
      host = self.public_dns
      timeout = "15m"
    }
  }
}

resource "aws_instance" "worker_airgap" {
  count                       = local.worker_airgap_count
  vpc_security_group_ids      = [aws_security_group.konvoy_ssh.id, aws_security_group.konvoy_private.id, aws_security_group.konvoy_egress.id, aws_security_group.public_facing_instances.id]
  subnet_id                   = aws_subnet.konvoy_public.id
  key_name                    = var.ssh_key_name
  ami                         = var.node_ami
  instance_type               = var.worker_instance_type
  availability_zone           = var.aws_availability_zones[0]
  source_dest_check           = "false"
  associate_public_ip_address = local.public_ip

  tags = var.tags

  root_block_device {
    volume_size = var.root_volume_size
  }

  /*provisioner "remote-exec" {
    inline = [
      "Skipping ssh test since airgap is enabled"
    ]

    connection {
      type = "ssh"
      user = var.ssh_username
      agent = true
      host = self.public_dns
      timeout = "15m"
    }
  }*/
}

resource "aws_instance" "bastion_host" {
  count                       = var.bastion_host_count
  vpc_security_group_ids      = [aws_security_group.konvoy_ssh.id, aws_security_group.konvoy_private.id, aws_security_group.konvoy_egress.id]
  subnet_id                   = aws_subnet.konvoy_public.id
  key_name                    = var.ssh_key_name
  ami                         = var.bastion_ami
  instance_type               = var.bastion_instance_type
  availability_zone           = var.aws_availability_zones[0]
  source_dest_check           = "false"
  associate_public_ip_address = "true"

  tags = var.tags

  root_block_device {
    volume_size = var.root_volume_size
  }

  provisioner "remote-exec" {
    inline = [
      "echo ok"
    ]

    connection {
      type = "ssh"
      user = var.ssh_username
      agent = true
      host = self.public_dns
      timeout = "15m"
    }
  }
}

resource "aws_instance" "registry_host" {
  count                       = var.registry_host_count
  vpc_security_group_ids      = [aws_security_group.konvoy_ssh.id, aws_security_group.konvoy_private.id, aws_security_group.konvoy_egress.id, aws_security_group.public_facing_instances.id]
  subnet_id                   = aws_subnet.konvoy_public.id
  key_name                    = var.ssh_key_name
  ami                         = var.registry_ami
  instance_type               = var.registry_instance_type
  availability_zone           = var.aws_availability_zones[0]
  source_dest_check           = "false"
  associate_public_ip_address = "true"

  tags = var.tags

  root_block_device {
    volume_size = var.root_volume_size
  }

  provisioner "remote-exec" {
    inline = [
      "echo ok"
    ]

    connection {
      type = "ssh"
      user = var.ssh_username
      agent = true
      host = self.public_dns
      timeout = "15m"
    }
  }
}


resource "aws_elb" "konvoy_control_plane" {
  internal                  = false
  security_groups           = [aws_security_group.konvoy_private.id, aws_security_group.konvoy_control_plane.id]
  subnets                   = [aws_subnet.konvoy_public.id]
  connection_draining       = true
  cross_zone_load_balancing = true


  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    target              = "HTTPS:6443/healthz"
    interval            = 10
  }

  listener {
    instance_port     = 6443
    instance_protocol = "tcp"
    lb_port           = 6443
    lb_protocol       = "tcp"
  }

  instances = aws_instance.control_plane.*.id

  tags = var.tags
}


  resource "local_file" "ansible_inventory" {
    filename = "../ansible/inventory.ini"
    content = <<EOT
[bastion]
%{ for bastion_ip in aws_instance.bastion_host.*.public_ip ~}
${bastion_ip}
%{ endfor ~}
[registry]
%{ for registry_ip in aws_instance.registry_host.*.public_ip ~}
${registry_ip}
%{ endfor ~}
[cluster]
%{ for cp_ip in aws_instance.control_plane.*.public_ip ~}
${cp_ip}
%{ endfor ~}
%{ for cpag_ip in aws_instance.control_plane_airgap.*.private_ip ~}
${cpag_ip}
%{ endfor ~}
%{ for worker_ip in aws_instance.worker.*.public_ip ~}
${worker_ip}
%{ endfor ~}
%{ for workerag_ip in aws_instance.worker_airgap.*.private_ip}
${workerag_ip}
%{ endfor ~}
EOT
  }

output "vpc_cidr" {
  value = aws_vpc.konvoy_vpc.cidr_block
}

output "kube_apiserver_address" {
  value = aws_elb.konvoy_control_plane.dns_name
}

output "control_plane_public_ips" {
  value = aws_instance.control_plane.*.public_ip
}
output "control_plane_airgap_private_ips"{
  value = aws_instance.control_plane_airgap.*.private_ip
}

output "worker_public_ips" {
  value = aws_instance.worker.*.public_ip
}

output "worker_airgap_private_ips"{
  value = aws_instance.worker_airgap.*.private_ip
}

output "bastion_public_ips" {
  value = aws_instance.bastion_host.*.public_ip
}

output "bastion_private_ips" {
  value = aws_instance.bastion_host.*.private_ip
}

output "registry_public_ips" {
  value = aws_instance.registry_host.*.public_ip
}

output "registry_private_ips" {
  value = aws_instance.registry_host.*.private_ip
}