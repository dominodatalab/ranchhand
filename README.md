# WIP: RanchHand

[![Release](https://img.shields.io/github/release/dominodatalab/ranchhand.svg?style=flat-square)](https://github.com/dominodatalab/ranchhand/releases/latest)
[![CircleCI](https://circleci.com/gh/dominodatalab/ranchhand.svg?style=svg)](https://circleci.com/gh/dominodatalab/ranchhand)

This tool deploys Rancher in HA mode onto existing hardware.

## Design

**ranchhand -> rke -> k8s -> rancher -> rke -> k8s**

Simple, right?
