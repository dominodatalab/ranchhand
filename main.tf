locals {
  ip_addresses = join(",", var.node_ips)
  ansbile_ssh_proxy = var.ssh_proxy_host == "" ? "" : format("-o StrictHostKeyChecking=no -o ProxyCommand=\"ssh -i %s -o StrictHostKeyChecking=no -W %%h:%%p -q %s@%s\"", var.ssh_key_path, var.ssh_proxy_user, var.ssh_proxy_host)
  cert_dnsnames = format("DNS:%s", join(",DNS:", var.cert_dnsnames))
  cert_ipaddresses = length(var.cert_ipaddresses) == 0 ? "" : format(",IP:%s", join(",IP:", var.cert_ipaddresses))
  cert_names = format("%s%s", local.cert_dnsnames, local.cert_ipaddresses)
  working_dir = format("%s/ranchhand-ansible.%s", var.working_dir, formatdate("YYYYMMDDhhmmss", timestamp()))
}

resource "random_password" "password" {
  count  = var.admin_password == "" ? 1 : 0
  length = 20

  # The default EXCEPT "-" and "'"because it can trigger CLI arguments / mangle quotes
  override_special = "!@#$%&*()_=+[]{}<>:?"
}

resource "local_file" "current_work_directory" {
    filename = "${local.working_dir}/.ansible"
}

resource "local_file" "copy_kubeconfig" {
    content = file("${local.working_dir}/kube_config_rancher-cluster.yml")
    filename = "${var.working_dir}/kube_config_rancher-cluster.yml"
    depends_on = ["null_resource.ansible_playbook"]
}

resource "null_resource" "ansible_playbook" {
  provisioner "local-exec" {
    command = join(" ", [
      "ansible-playbook",
      "-i '${local.ip_addresses},'",
      "--private-key=${var.ssh_key_path}",
      "--user=${var.ssh_username}",
      "--ssh-common-args='-o StrictHostKeyChecking=no ${local.ansbile_ssh_proxy}'",
      "-e 'cert_names=${local.cert_names}'",
      "-e 'local_output_dir=${local.working_dir}'",
      "ansible/prod.yml --diff"
    ])
    
    working_dir = "${path.module}"
    environment = {
      RANCHER_PASSWORD = var.admin_password == "" ? join("", random_password.password.*.result) : var.admin_password
    }
  }
  depends_on = ["local_file.current_work_directory"]
}