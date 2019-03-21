# WIP: RanchHand

[![Release](https://img.shields.io/github/release/dominodatalab/ranchhand.svg)](https://github.com/dominodatalab/ranchhand/releases/latest)
[![CircleCI](https://img.shields.io/circleci/project/github/dominodatalab/ranchhand/master.svg)](https://img.shields.io/circleci/project/github/dominodatalab/ranchhand/master.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/dominodatalab/ranchhand)](https://goreportcard.com/report/github.com/dominodatalab/ranchhand)
[![GoDoc](https://godoc.org/github.com/dominodatalab/ranchhand?status.svg)](https://godoc.org/github.com/dominodatalab/ranchhand)

This tool deploys Rancher in HA mode onto existing hardware.

## Design

**ranchhand -> rke -> k8s -> rancher -> rke -> k8s**

Simple, right?

## Development

You can test your changes locally using Vagrant and VirtualBox.

1. Make sure you have Vagrant and VirtualBox installed.

    `brew cask install vagrant virtualbox`

1. Create one or more VMs.

    ```
    cd test/
    NODE_COUNT=N NODE_DISTRO="ubuntu_xenial|ubuntu_bionic|centos|rhel" vagrant up
    ```

1. Use go to launch a ranchhand run.

    `go run main.go run -u vagrant -i ~/.ssh/id_rsa -n 192.168.50.10,...`
