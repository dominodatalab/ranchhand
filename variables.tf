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
variable "admin_password" {
  description = "Password override for the initial admin user"
  default     = ""
}

variable "ansible_ssh_retries" {
  default = 5
}

variable "ansible_ssh_timeout" {
  default = 30
}


variable "cert_dnsnames" {
  description = "Hostnames for the rancher and rke ssl certs (comma-delimited)"
  default     = ["domino.rancher"]
}

variable "cert_ipaddresses" {
  description = "IP addresses for the rancher and rke ssl certs (comma-delimited)"
  default     = []
}

variable "cert_manager_version" {
  description = "cert-manager helm chart version. With the [v]"
  default     = "v1.13.1"
}

variable "helm_v3_registry_host" {
  description = "default helm registry"
  default     = "gcr.io"
}

variable "helm_v3_namespace" {
  description = "namespace for helm"
  default     = "domino-eng-service-artifacts"
}

variable "helm_v3_registry_password" {
  description = "helm registry password"
  default     = ""
}
variable "helm_v3_registry_user" {
  description = "helm registry username"
  default     = "_json_key"
}

variable "helm_version" {
  description = "version of Helm to install. With the [v]"
  default     = "v3.14.0"
}

# make sure to check that the Rancher chart
# https://github.com/rancher/rancher/blob/v2.7.5/chart/values.yaml
# remember to replace v2.7.5 in the url above with the correct version
# is compatible with the kubectl version
variable "kubectl_version" {
  description = "version of Kubectl to install. With the [v]"
  default     = "v1.26.7"
}

variable "newrelic_license_key" {
  description = "License key for New Relic"
  default     = ""
}

variable "newrelic_namespace" {
  description = "Namespace to install New Relic into"
  default     = "monitoring"
}

variable "newrelic_service_name" {
  description = "Name of the New Relic Service"
  default     = "nri-bundle"
}

variable "newrelic_service_version" {
  description = "The New Relic service version. Without [v]"
  default     = "1.11.3-0.1.0"
}

# For RKE1, Rancher and Kubernetes support matrix please reference
# https://www.suse.com/suse-rancher/support-matrix/all-supported-versions/
# before changing these values

variable "rancher_image_tag" {
  description = "Rancher image for the rancherImageTag helm value. With the [v]"
  default     = "v2.7.5"
}
variable "rancher_version" {
  description = "Rancher version for Helm install. Without [v]"
  default     = "2.7.5"
}

variable "rke_version" {
  description = "RKE version to use to create the underlying Kubernetes cluster. With the [v]"
  default     = "v1.4.8"
}

variable "ssh_key_path" {
  description = "Path to the SSH private key that will be used to connect to the VMs"
  default     = "~/.ssh/id_rsa"
}

variable "ssh_proxy_host" {
  description = "Bastion host used to proxy SSH connections"
  default     = ""
}

variable "ssh_proxy_user" {
  description = "Bastion host SSH username"
  default     = ""
}

variable "ssh_username" {
  description = "SSH username on the nodes"
  default     = "admin"
}


variable "working_dir" {
  description = "Directory where ranchhand should be executed. Defaults to the current working directory."
  default     = ""
}

