output "cluster_provisioned" {
  value = null_resource.ansible_playbook.id
}

output "admin_password" {
  description = "Generated Rancher admin user password"
  value       = var.admin_password == "" ? join("", random_password.password.*.result) : var.admin_password
}