#!/bin/bash

kubectl create namespace devops
helm repo add nfs-subdir-external-provisioner https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner/
helm repo update

helm upgrade --install nfs-provisioner nfs-subdir-external-provisioner/nfs-subdir-external-provisioner \
--namespace nfs-provisioner \
--create-namespace \
--set nfs.server=10.3.202.102 \
--set nfs.path=/data/nfs_share \
--set image.repository=swr.cn-north-4.myhuaweicloud.com/ddn-k8s/k8s.gcr.io/sig-storage/nfs-subdir-external-provisioner \
--set image.tag=v4.0.2 \
--set storageClass.name=nfs-storage \
--set storageClass.onDelete="retain" \
--set extraArgs.enableFixPath=true \
--set storageClass.defaultClass=true