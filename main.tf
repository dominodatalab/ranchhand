locals {
  ip_addresses      = join(",", var.node_ips)
  ansible_ssh_proxy = var.ssh_proxy_host == "" ? "" : format("-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ProxyCommand=\"ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -W %%h:%%p -q %s@%s\"", var.ssh_key_path, var.ssh_proxy_user, var.ssh_proxy_host)
  cert_dnsnames     = format("DNS:%s", join(",DNS:", var.cert_dnsnames))
  cert_ipaddresses  = length(var.cert_ipaddresses) == 0 ? "" : format(",IP:%s", join(",IP:", var.cert_ipaddresses))
  cert_names        = format("%s%s", local.cert_dnsnames, local.cert_ipaddresses)
}

resource "random_password" "password" {
  count  = var.admin_password == "" ? 1 : 0
  length = 20

  # The default EXCEPT "-" and "'"because it can trigger CLI arguments / mangle quotes
  override_special = "!@#$&*_+?"
}

resource "null_resource" "ansible_playbook" {
  provisioner "local-exec" {
    command = <<-EOF
      ansible-galaxy install -r ansible/requirements.yml && \
      ansible-playbook \
        -i '${local.ip_addresses},' \
        --private-key=${var.ssh_key_path} \
        --user=${var.ssh_username} \
        --ssh-common-args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ${local.ansible_ssh_proxy}' \
        -e 'cert_names=${local.cert_names}' \
        -e 'node_count=${length(var.node_ips)}' \
        -e 'rancher_version=${var.rancher_version}' \
        -e 'rke_version=${var.rke_version}' \
        -e 'local_output_dir=${var.working_dir}/ansible.${self.id}' \
        ansible/prod.yml --diff
    EOF

    working_dir = path.module
    environment = {
      ANSIBLE_SSH_RETRIES       = var.ansible_ssh_retries
      ANSIBLE_TIMEOUT           = var.ansible_ssh_timeout
      RANCHER_PASSWORD          = nonsensitive(var.admin_password == "" ? join("", random_password.password.*.result) : var.admin_password)
      RANCHER_IMAGE_TAG         = var.rancher_image_tag
      HELM_V3_REGISTRY_HOST     = var.helm_v3_registry_host
      HELM_V3_REGISTRY_USER     = var.helm_v3_registry_user
      HELM_V3_REGISTRY_PASSWORD = var.helm_v3_registry_password
      NEWRELIC_LICENSEKEY       = var.newrelic_licensekey
    }
  }

  provisioner "local-exec" {
    command     = "cp ansible.${self.id}/kube_config_rancher-cluster.yml kube_config_rancher-cluster.yml"
    working_dir = var.working_dir
  }
}
