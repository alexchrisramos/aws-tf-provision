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

variable "control_plane_instance_type" {
  description = "The instance type for the control-plane nodes"
  default = "m5.2xlarge"
}

variable "worker_instance_type" {
  description = "The instance type for the worker nodes"
  default = "m5.2xlarge"
}

variable "worker_count" {
  description = "The number of worker nodes"
  default = 1
}

variable "extra_worker_count" {
  description = "The number of worker nodes"
  default = 0
}

variable "extra_worker_instance_type" {
  description = "The instance type for the worker nodes"
  default = "m5.2xlarge"
}


variable "ssh_username" {
  description = "The user for connecting to the instance over ssh"
  default = "centos"
}

locals {
  public_subnet_range        = var.vpc_cidr
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

resource "aws_instance" "control_plane" {
  count                       = 3
  vpc_security_group_ids      = [aws_security_group.konvoy_ssh.id, aws_security_group.konvoy_private.id, aws_security_group.konvoy_egress.id]
  subnet_id                   = aws_subnet.konvoy_public.id
  key_name                    = var.ssh_key_name
  ami                         = var.node_ami
  instance_type               = var.control_plane_instance_type
  availability_zone           = var.aws_availability_zones[0]
  source_dest_check           = "false"
  associate_public_ip_address = "true"

  tags = var.tags

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

resource "aws_instance" "worker" {
  count                       = var.worker_count
  vpc_security_group_ids      = [aws_security_group.konvoy_ssh.id, aws_security_group.konvoy_private.id, aws_security_group.konvoy_egress.id]
  subnet_id                   = aws_subnet.konvoy_public.id
  key_name                    = var.ssh_key_name
  ami                         = var.node_ami
  instance_type               = var.worker_instance_type
  availability_zone           = var.aws_availability_zones[0]
  source_dest_check           = "false"
  associate_public_ip_address = "true"

  tags = var.tags

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

resource "aws_instance" "extra_worker" {
  count                       = var.extra_worker_count
  vpc_security_group_ids      = [aws_security_group.konvoy_ssh.id, aws_security_group.konvoy_private.id, aws_security_group.konvoy_egress.id]
  subnet_id                   = aws_subnet.konvoy_public.id
  key_name                    = var.ssh_key_name
  ami                         = var.node_ami
  instance_type               = var.extra_worker_instance_type
  availability_zone           = var.aws_availability_zones[0]
  source_dest_check           = "false"
  associate_public_ip_address = "true"

  tags = var.tags

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

resource "aws_security_group" "konvoy_control_plane" {
  description = "Allow inbound SSH for Konvoy."
  vpc_id      = aws_vpc.konvoy_vpc.id

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
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

output "kube_apiserver_address" {
  value = aws_elb.konvoy_control_plane.dns_name
}

output "control_plane_public_ips" {
  value = aws_instance.control_plane.*.public_ip
}

output "worker_public_ips" {
  value = aws_instance.worker.*.public_ip
}

output "extra_worker_public_ips" {
  value = aws_instance.extra_worker.*.public_ip
}
