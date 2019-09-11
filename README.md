# RanchHand

[![Release](https://img.shields.io/github/release/dominodatalab/ranchhand.svg)](https://github.com/dominodatalab/ranchhand/releases/latest)
[![CircleCI](https://img.shields.io/circleci/project/github/dominodatalab/ranchhand/master.svg)](https://circleci.com/gh/dominodatalab/ranchhand)
[![Go Report Card](https://goreportcard.com/badge/github.com/dominodatalab/ranchhand)](https://goreportcard.com/report/github.com/dominodatalab/ranchhand)
[![GoDoc](https://godoc.org/github.com/dominodatalab/ranchhand?status.svg)](https://godoc.org/github.com/dominodatalab/ranchhand)

Deploy Rancher in HA mode onto existing hardware.

## Design

This tool aims to automate the steps listed in Rancher's official [HA Install][] documentation in a reproducable manner. It also enforces many of the recommendations given inside Rancher's [hardening guide][].

## Usage

1. Download the [latest release][] from GitHub.
1. [Install Ansible](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html) (version >=2.8) locally
1.  (optional) To update the Rancher default password, set the `RANCHER_PASSWORD` environment variable: 
     `export RANCHER_PASSWORD=<new password>`
1. Execute `ansible-playbook -i '1.2.4.5,...,10.20.30.40,' --private-key=~/.ssh/id_rsa --user=ubuntu ansible/prod.yml --diff --check` to perform a dry run of all the changes.

### Example
This example shows a manual run of the production playbook (prod.yml) from a local machine imaging a cluster behind a bastion/proxy server.

```
ansible-playbook -i '10.0.1.6,10.0.1.51,10.0.1.94,' --private-key=/Users/myhost/.ssh/id_rsa --user=ubuntu --ssh-common-args='-o StrictHostKeyChecking=no -o StrictHostKeyChecking=no -o ProxyCommand="ssh -o StrictHostKeyChecking=no -W %h:%p -q ubuntu@54.190.1.95"' ansible/prod.yml --diff
```

In the example above, only the bastion server, 54.190.1.95, is publicly accessible. However, including the Terraform module should be sufficient for most users.

## Terraform

Using the Terraform module, you can leverage Ranchhand to create a Rancher cluster on a specific set of nodes.

```hcl
module "ranchhand" {
  source = "github.com/dominodatalab/ranchhand"

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

    `ansible-playbook -i '192.168.50.10,...,' --private-key=~/.ssh/id_rsa --user=ubuntu ansible/prod.yml --diff --check `
    
    _Note the trailing comma (",") in the host/ip list._
    
### Ansible References
Here are some helpful Ansible references for getting started with Ansible.

1. [Ansible Overview](https://docs.ansible.com/ansible/latest/index.html)
1. [Installation Guide](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)
1. [Project Directory Layout](https://docs.ansible.com/ansible/latest/user_guide/playbooks_best_practices.html#content-organization)
1. [Roles](https://docs.ansible.com/ansible/latest/user_guide/playbooks_reuse_roles.html)
1. [Best Practices](https://docs.ansible.com/ansible/latest/user_guide/playbooks_best_practices.html#best-practices)


## Contribute

Contributions are always welcome! Please submit any questions, bugs or changes via an issue or PR. Thank you.

[vagrantfile]: test/Vagrantfile

[rke]: https://github.com/rancher/rke
[ha install]: https://rancher.com/docs/rancher/v2.x/en/installation/ha/
[hardening guide]: https://releases.rancher.com/documents/security/latest/Rancher_Hardening_Guide.pdf
[latest release]: https://github.com/dominodatalab/ranchhand/releases/latest
