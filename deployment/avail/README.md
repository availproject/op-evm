# Avail Devnet Ansible Logic

TBD ...


## TODO

Potential todos that we should consider doing in the future.

- [] Getting inventory setup with terraform


## HOW-TO

### Server Dependencies

First what we need to do is install dependencies. Following command will ensure following programs are available on the machine:

- Docker
- Cargo/Rust

```sh
ansible-playbook -i deployment/avail/inventory/hosts.yml -l bootnodes deployment/avail/playbooks/dependencies.yml
```



## Examples

### Execute ansible command on group of hosts

Just an example of how custom shell command can be executed on group of the hosts

```sh
ansible -i deployment/avail/inventory/hosts.yml bootnodes -m shell -a "ls -la"
```