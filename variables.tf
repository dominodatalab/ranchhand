#------------------------------------------------------------------------------
# REQUIRED
#------------------------------------------------------------------------------
variable "node_ips" {
  description = ""
  type        = list(string)
}

#------------------------------------------------------------------------------
# OPTIONAL
#------------------------------------------------------------------------------
variable "ansible_ssh_retries" {
  default = 5
}

variable "ansible_ssh_timeout" {
  default = 30
}

variable "working_dir" {
  description = "Directory where ranchhand should be executed. Defaults to the current working directory."
  default     = ""
}

variable "cert_dnsnames" {
  description = "Hostnames for the rancher and rke ssl certs (comma-delimited)"
  default     = ["domino.rancher"]
}

variable "cert_ipaddresses" {
  description = "IP addresses for the rancher and rke ssl certs (comma-delimited)"
  default     = []
}

variable "rancher_version" {
  description = "Rancher version to install into the cluster."
  default     = ""
}

variable "rke_version" {
  description = "RKE version to use to create the underlying Kubernetes cluster."
  default     = ""
}

variable "ssh_username" {
  description = "SSH username on the nodes"
  default     = "admin"
}

variable "ssh_key_path" {
  description = "Path to the SSH private key that will be used to connect to the VMs"
  default     = "~/.ssh/id_rsa"
}

variable "ssh_proxy_user" {
  description = "Bastion host SSH username"
  default     = ""
}

variable "ssh_proxy_host" {
  description = "Bastion host used to proxy SSH connections"
  default     = ""
}

variable "admin_password" {
  description = "Password override for the initial admin user"
  default     = ""
}

variable "helm_v3_registry_host" {
  default = "gcr.io"
}

variable "helm_v3_registry_user" {
  default = ""
}

variable "helm_v3_registry_password" {
  default = ""
}
