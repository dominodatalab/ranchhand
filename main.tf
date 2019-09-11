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
  ansbile_ssh_proxy = "${var.ssh_proxy_host == "" ? "" : format("-o StrictHostKeyChecking=no -o ProxyCommand=\"ssh -o StrictHostKeyChecking=no -W %%h:%%p -q %s@%s\"", var.ssh_proxy_user, var.ssh_proxy_host)}"
  cert_dnsnames = "${format("DNS:%s", join(",DNS:", var.cert_dnsnames))}"
  cert_ipaddresses = "${length(var.cert_ipaddresses) == 0 ? "" : format(",IP:%s", join(",IP:", var.cert_ipaddresses))}"
  cert_names = "${format("%s%s", local.cert_dnsnames, local.cert_ipaddresses)}"
}

resource "random_password" "password" {
  count  = "${var.admin_password == "" ? 1 : 0}"
  length = 20

  # The default EXCEPT "-" and "'"because it can trigger CLI arguments / mangle quotes
  override_special = "!@#$%&*()_=+[]{}<>:?"
}

resource "null_resource" "ansible-playbook" {
  provisioner "local-exec" {
    command = "ansible-playbook -i '${local.ip_addresses},' --private-key=${var.ssh_key_path} --user=${var.ssh_username} --ssh-common-args='-o StrictHostKeyChecking=no ${local.ansbile_ssh_proxy}' -e 'cert_names=${local.cert_names}' prod.yml --diff"
    
    working_dir = "${path.module}/ansible"
    environment = {
      RANCHER_PASSWORD = "${var.admin_password == "" ? join("", random_password.password.*.result) : var.admin_password}"
    }
  }
}



#------------------------------------------------------------------------------
# DEPRECATE BELOW
#------------------------------------------------------------------------------

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
