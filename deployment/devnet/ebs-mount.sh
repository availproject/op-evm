#cloud-config

runcmd:
- 'sudo mkdir /home/ubuntu/workspace -p'
- 'disk="/dev/nvme1n1"'
- '[ -z "$(lsblk -no FSTYPE "$disk")" ] && mkfs.ext4 "$disk"' # Format the disk as ext4 if the disk is empty and not formattet
- 'sudo mount "$disk" /home/ubuntu/workspace -t ext4'
- 'sudo chown -R ubuntu:ubuntu /home/ubuntu/workspace'

output : { all : '| tee -a /var/log/cloud-init-output.log' }
