# WIP: RanchHand

[![Release](https://img.shields.io/github/release/dominodatalab/ranchhand.svg?style=flat-square)](https://github.com/dominodatalab/ranchhand/releases/latest)
[![CircleCI](https://circleci.com/gh/dominodatalab/ranchhand.svg?style=svg)](https://circleci.com/gh/dominodatalab/ranchhand)

This tool deploys Rancher in HA mode onto existing hardware.

## Design

**ranchhand -> rke -> k8s -> rancher -> rke -> k8s**

Simple, right?

## Development

You can test your changes locally using Vagrant and VirtualBox

1. Make sure you have Vagrant and VirtualBox installed.

    `brew cask install vagrant virtualbox`

1. Create one or more VMs.

    ```
    cd test/
    NODE_COUNT=N NODE_DISTRO="ubuntu_xenial|ubuntu_bionic|centos|rhel" vagrant up
    ```

1. Use go to launch the entrypoint to the application.

    `go run main.go -n 192.168.50.10 ...`
