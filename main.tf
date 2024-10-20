locals {
  ip_addresses      = join(",", var.node_ips)
  ansible_ssh_proxy = var.ssh_proxy_host == "" ? "" : format("-o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ProxyCommand=\"ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -W %%h:%%p -q %s@%s\"", var.ssh_key_path, var.ssh_proxy_user, var.ssh_proxy_host)
  cert_dnsnames     = format("DNS:%s", join(",DNS:", var.cert_dnsnames))
  cert_ipaddresses  = length(var.cert_ipaddresses) == 0 ? "" : format(",IP:%s", join(",IP:", var.cert_ipaddresses))
  cert_names        = format("%s%s", local.cert_dnsnames, local.cert_ipaddresses)
}

resource "random_password" "password" {
  count  = var.admin_password == "" ? 1 : 0
  length = 20

  # The default EXCEPT "-" and "'"because it can trigger CLI arguments / mangle quotes
  override_special = "!@#$&*_+?"

  lifecycle {
    ignore_changes = [override_special]
  }
}

resource "null_resource" "ansible_playbook" {
  provisioner "local-exec" {
    command = <<-EOF
      ansible-galaxy install -r ansible/requirements.yml && \
      ansible-playbook \
        -i '${local.ip_addresses},' \
        --private-key=${var.ssh_key_path} \
        --user=${var.ssh_username} \
        --ssh-common-args='-o IdentitiesOnly=yes -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null ${local.ansible_ssh_proxy}' \
        -e 'cert_manager_version=${var.cert_manager_version}' \
        -e 'cert_names=${local.cert_names}' \
        -e '{"helm": { "host": "${var.helm_v3_registry_host}","user": "${var.helm_v3_registry_user}", "namespace": "${var.helm_v3_namespace}","password": "${var.helm_v3_registry_password}" }}' \
        -e 'helm_version=${var.helm_version}' \
        -e 'kubectl_version=${var.kubectl_version}' \
        -e '{ "newrelic": { "license_key": "${var.newrelic_license_key}","namespace": "${var.newrelic_namespace}", "service_name": "${var.newrelic_service_name}","service_version": "${var.newrelic_service_version}" }}' \
        -e 'node_count=${length(var.node_ips)}' \
        -e 'rancher_image_tag=${var.rancher_image_tag}' \
        -e 'rancher_version=${var.rancher_version}' \
        -e 'rke_version=${var.rke_version}' \
        -e 'local_output_dir=${var.working_dir}/ansible.${self.id}' \
        ansible/prod.yml --diff
    EOF

    working_dir = path.module
    environment = {
      ANSIBLE_SSH_RETRIES = var.ansible_ssh_retries
      ANSIBLE_TIMEOUT     = var.ansible_ssh_timeout
      RANCHER_PASSWORD    = nonsensitive(var.admin_password == "" ? join("", random_password.password.*.result) : var.admin_password)
    }
  }

  provisioner "local-exec" {
    command     = "cp ansible.${self.id}/kube_config_rancher-cluster.yml kube_config_rancher-cluster.yml"
    working_dir = var.working_dir
  }
}
