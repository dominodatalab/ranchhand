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

  # TODO: Upgrade to Terraform 1.12 & use "${format("YYYYMMDDhhmmss", timestamp())}"
  working_dir = "${format("%s/ansible.%s", var.working_dir, timestamp())}"
}

resource "random_password" "password" {
  count  = "${var.admin_password == "" ? 1 : 0}"
  length = 20

  # The default EXCEPT "-" and "'"because it can trigger CLI arguments / mangle quotes
  override_special = "!@#$%&*()_=+[]{}<>:?"
}

resource "local_file" "create_directory" {
    content  = ""
    filename = "${local.working_dir}/.ansible"
}

resource "null_resource" "ansible-playbook" {
  provisioner "local-exec" {
    command = "ansible-playbook -i '${local.ip_addresses},' --private-key=${var.ssh_key_path} --user=${var.ssh_username} --ssh-common-args='-o StrictHostKeyChecking=no ${local.ansbile_ssh_proxy}' -e 'cert_names=${local.cert_names}' -e 'local_output_dir=${local.working_dir}' ansible/prod.yml --diff"
    
    working_dir = "${path.module}"
    environment = {
      RANCHER_PASSWORD = "${var.admin_password == "" ? join("", random_password.password.*.result) : var.admin_password}"
    }
  }
  depends_on = ["local_file.create_directory"]
}