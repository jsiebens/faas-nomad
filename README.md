faas-nomad - OpenFaas provider for Nomad
===========

[![Build Status](https://github.com/jsiebens/faas-nomad/workflows/build/badge.svg?branch=main)](https://github.com/jsiebens/faas-nomad/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/jsiebens/faas-nomad)](https://goreportcard.com/report/github.com/jsiebens/faas-nomad)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![OpenFaaS](https://img.shields.io/badge/openfaas-serverless-blue.svg)](https://www.openfaas.com)

This repository contains an OpenFaaS provider for the HashiCorp Nomad scheduler. 
OpenFaaS allows you to run your private functions as a service. 
Functions are packaged in Docker Containers which enables you to work in any language and also interact with any software which can also be installed in the container.

Inspired by [hashicorp/faas-nomad](https://github.com/hashicorp/faas-nomad).

## Using Vagrant for Local Development
Vagrant is a tool for provisioning dev environments. The `Vagrantfile` governs the Vagrant configuration:
1) Install Vagrant via [download links](https://www.vagrantup.com/downloads.html) or package manager
2) Install VirtualBox via [download links](https://www.virtualbox.org/wiki/Downloads)
3) `vagrant up`

The provisioners install Docker, Nomad, Consul, and Vault then launch OpenFaaS components with Nomad. 
If successful, the following services will be available over the private network (192.168.50.2):
- Nomad (v1.1.3) - http://192.168.50.2:4646
- Consul (v1.10.1) - http://192.168.50.2:8500
- Vault (v1.8.0) - http://192.168.50.2:8200
- Prometheus (2.14.0) - http://192.168.50.2:9090
- OpenFaaS Gateway (0.21.1) - http://192.168.50.2:8080

This setup is intended to streamline local development of the faas-nomad provider with a more complete setup of the hashicorp ecosystem. Therefore, it is assumed that the faas-nomad source code is located on your workstation, and or is configured to listen on 0.0.0.0:8080 when debugging/running the Go process.

## Starting a remote Nomad / OpenFaaS environment
If you would like to test OpenFaaS running on a remote cluster, more demos and instructions are (or will be) available here:
[jsiebens/faas-nomad-demos · GitHub](https://github.com/jsiebens/faas-nomad-demos)

Regardless of which method you use interacting with OpenFaaS is the same.

## Resources

- [OpenFaaS Docs](https://docs.openfaas.com/)
- [faas-provider](https://github.com/openfaas/faas-provider)
- [HashiCorp Nomad](https://nomadproject.io)
- [HashiCorp Consul](https://consul.io)
- [HashiCorp Vault](https://vaultproject.io)
