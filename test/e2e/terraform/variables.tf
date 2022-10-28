variable "name" {
  type = string
}

variable "tags" {
  type    = map(any)
  default = {}
}

variable "ami_name" {
  type    = string
  default = "ubuntu_focal"
}

variable "public_key" {
  type = string
}
