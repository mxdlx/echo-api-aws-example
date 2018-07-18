# REST API example
This is a REST API implemented with Golang [Echo](https://echo.labstack.com/) using [ReJSON](https://oss.redislabs.com/rejson/) as storage, [Traefik](https://traefik.io/) as web server.
Everything is deployed in a [Docker Swarm](https://docs.docker.com/engine/swarm/) with no failover (manager-manager) on [AWS](https://aws.amazon.com/) using [CloudFormation](https://aws.amazon.com/cloudformation/) and [Ansible](https://www.ansible.com/).

## Why
This is an example I did for an interview, so it's just an _example_. Docker Swarm with only two hosts makes no sense but I wanted to test it with overlay networking. Both EC2 instances share Redis append-only-file which may be dangerous but I wanted to test AWS EFS, also Redis has no multi-master mode but it's easier to use than other document database systems. I had no experience with AWS so CloudFormation template might not be as nice as other examples out there but it works.

## Requirements
* Ansible 2.2+, EC2 instances must be defined in hosts file after stack creation, example:

```bash
# Ansible hosts file in yaml format
aws:
  hosts:
    inst0:
      ansible_ssh_host: <ec2-ip-address>.compute.amazonaws.com
      ansible_connection: ssh
      ansible_port: 22
      ansible_user: ec2-user
      ansible_ssh_common_args: "-o StrictHostKeyChecking=no"
    inst1:
      ansible_ssh_host: <ec2-ip-address>.compute.amazonaws.com
      ansible_connection: ssh
      ansible_port: 22
      ansible_user: ec2-user
      ansible_ssh_common_args: "-o StrictHostKeyChecking=no"
```

* AWS account with KeyPair for your region.
* SSH config: it's useful to define AWS hosts in your SSH config file, example:

```bash
# AWS
Host *.compute.amazonaws.com
IdentityFile ~/.ssh/your-aws-keypair
```

## HowTo
1. Create cloudformation stack with provided template.
2. Setup your ansible hosts file, this playbook needs "inst0" and "inst1" hosts defined as above example.
3. Run `$ ansible-playbook playbook.yml`
