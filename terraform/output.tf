output "cluster_provisioned" {
  value = "${null_resource.provisioner.id}"
}

output "admin_password" {
  description = "Generated Rancher admin user password"
  value       = "${var.admin_password == "" ? join("", random_string.password.*.result) : var.admin_password}"
}
