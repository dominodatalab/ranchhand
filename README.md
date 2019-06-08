# RanchHand

[![Release](https://img.shields.io/github/release/dominodatalab/ranchhand.svg)](https://github.com/dominodatalab/ranchhand/releases/latest)
[![CircleCI](https://img.shields.io/circleci/project/github/dominodatalab/ranchhand/master.svg)](https://circleci.com/gh/dominodatalab/ranchhand)
[![Go Report Card](https://goreportcard.com/badge/github.com/dominodatalab/ranchhand)](https://goreportcard.com/report/github.com/dominodatalab/ranchhand)
[![GoDoc](https://godoc.org/github.com/dominodatalab/ranchhand?status.svg)](https://godoc.org/github.com/dominodatalab/ranchhand)

Deploy Rancher in HA mode onto existing hardware.

## Design

This tool aims to automate the steps listed in Rancher's official [HA Install][] documentation in a reproducable manner. It also enforces many of the recommendations given inside Rancher's [hardening guide][].

## Usage

1. Download the latest [latest release][] from GitHub.
2. Execute `ranchhand run -h` to see all of the available options.

## Terraform

Using the Terraform module, you can leverage Ranchhand to create a Rancher cluster on a specific set of nodes.

```hcl
module "ranchhand" {
  source = "github.com/dominodatalab/ranchhand/terraform"

  node_ips         = ["..."]
  distro           = "darwin"
  release          = "latest"
  working_dir      = "..."
  cert_dnsnames    = ["..."]
  cert_ipaddresses = ["..."]

  ssh_username   = "..."
  ssh_key_path   = "..."
  ssh_proxy_user = "..."
  ssh_proxy_host = "..."
}
```

## Development

Please submit any feature enhancements, bug fixes, or ideas via pull requests or issues.  If you need to test local changes e2e, you can do so using Vagrant and Virtualbox. Here are the recommended steps:

1. Make sure you have Vagrant and VirtualBox installed.

    `brew cask install vagrant virtualbox`

1. Create one or more VMs. For convenience, a pre-configured [Vagrantfile][] is available.

    ```
    cd test/
    NODE_COUNT=N NODE_DISTRO="ubuntu_xenial|ubuntu_bionic|centos|rhel" vagrant up
    ```

1. Use `go` to launch a Ranchhand run against your VM(s) and verify your changes.

    `go run main.go run -u vagrant -i ~/.ssh/id_rsa -n 192.168.50.10,...`

[vagrantfile]: test/Vagrantfile

[rke]: https://github.com/rancher/rke
[ha install]: https://rancher.com/docs/rancher/v2.x/en/installation/ha/
[hardening guide]: https://releases.rancher.com/documents/security/latest/Rancher_Hardening_Guide.pdf
[latest release]: https://github.com/dominodatalab/ranchhand/releases/latest
