#!/bin/bash
# Install NFS server
sudo yum install -y nfs-utils
mkdir /data/nfs_share
echo "/data/nfs_share *(rw,sync,no_root_squash,no_subtree_check)" > /etc/exports
systemctl enable --now nfs-server
showmount -e localhost