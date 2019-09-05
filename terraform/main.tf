provider "local" {
  version = "~> 1.2"
}

provider "null" {
  version = "~> 2.1"
}

provider "random" {
  version = "~> 2.1"
}

provider "template" {
  version = "~> 2.1"
}

locals {
  script       = "launch_ranchhand.sh"
  ip_addresses = "${join(",", var.node_ips)}"
}

resource "null_resource" {
  provisioner "local-exec" {
    command = "ansible-playbook -i '${local.ip_addresses},' --private-key=${var.ssh_key_path} --remote-user=${var.ssh_username} prod.yml --diff"
  }
}



#------------------------------------------------------------------------------
# DEPRECATE BELOW
#------------------------------------------------------------------------------

# resource "random_password" "password" {
#   count  = "${var.admin_password == "" ? 1 : 0}"
#   length = 20

#   # The default EXCEPT "-" and "'"because it can trigger CLI arguments / mangle quotes
#   override_special = "!@#$%&*()_=+[]{}<>:?"
# }

# data "template_file" "launcher" {
#   template = "${file("${path.module}/templates/${local.script}")}"

#   vars {
#     distro   = "${var.distro}"
#     release  = "${var.release}"
#     node_ips = "${local.ip_addresses}"

#     cert_ips       = "${join(",", var.cert_ipaddresses)}"
#     cert_dns_names = "${join(",", var.cert_dnsnames)}"

#     ssh_user       = "${var.ssh_username}"
#     ssh_key_path   = "${var.ssh_key_path}"
#     ssh_proxy_user = "${var.ssh_proxy_user}"
#     ssh_proxy_host = "${var.ssh_proxy_host}"
#   }
# }

# resource "local_file" "launcher" {
#   content  = "${data.template_file.launcher.rendered}"
#   filename = "${var.working_dir == "" ? local.script : "${var.working_dir}/${local.script}"}"
# }

# resource "null_resource" "provisioner" {
#   triggers {
#     instance_ids = "${local.ip_addresses}"
#   }

#   provisioner "local-exec" {
#     command     = "${data.template_file.launcher.rendered}"
#     interpreter = ["/bin/bash", "-c"]
#     working_dir = "${var.working_dir}"

#     environment = {
#       RANCHER_PASSWORD = "${var.admin_password == "" ? join("", random_password.password.*.result) : var.admin_password}"
#     }
#   }
# }
