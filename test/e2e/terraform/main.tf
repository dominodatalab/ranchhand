provider "aws" {
  region = local.region

  default_tags {
    tags = var.tags
  }
}

locals {
  region = "us-east-1"
  name   = "ranchhand-e2e-test"

  amis = {
    "ubuntu_xenial" = { owner = "099720109477", name = "ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*" }
    "ubuntu_focal"  = { owner = "099720109477", name = "ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*" }
    "centos7"        = { owner = "679593333241", name = "CentOS Linux 7 x86_64 HVM EBS*" }
  }
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 3.0"

  name = var.name
  cidr = "10.0.0.0/18"

  azs            = ["${local.region}a", "${local.region}b", "${local.region}c"]
  public_subnets = ["10.0.0.0/24", "10.0.1.0/24", "10.0.2.0/24"]
}

module "security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 4.0"

  name        = var.name
  description = "Security group for example usage with EC2 instance"
  vpc_id      = module.vpc.vpc_id

  ingress_cidr_blocks = ["0.0.0.0/0"]
  ingress_rules       = ["ssh-tcp", "all-icmp"]
  egress_rules        = ["all-all"]
}

data "aws_ami" "this" {
  most_recent = true
  owners      = [local.amis[var.ami_name].owner]

  filter {
    name   = "name"
    values = [local.amis[var.ami_name].name]
  }
}

resource "aws_key_pair" "this" {
  key_name   = var.name
  public_key = file(var.public_key)
}

module "ec2_instance" {
  source = "terraform-aws-modules/ec2-instance/aws"

  name = var.name

  ami                         = data.aws_ami.this.id
  instance_type               = "t3.micro"
  key_name                    = aws_key_pair.this.key_name
  cpu_credits                 = "unlimited"
  subnet_id                   = element(module.vpc.public_subnets, 0)
  vpc_security_group_ids      = [module.security_group.security_group_id]
  associate_public_ip_address = true
}
