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