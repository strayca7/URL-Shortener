> Rocky LInux，Kubenetes 1.29.2
>
> |   节点   |      IP      |
>| :------: | :----------: |
> | Master01 | 10.3.202.100 |
> |  Node01  | 10.3.202.101 |
> |  Node02  | 10.3.202.102 |
> |  Node03  | 10.3.202.103 |





# 前置

```bash
./deploy/pre-work.sh
```



# NFS

## NFS 服务搭建

所有节点执行：

```bash
yum install -y nfs-utils
```

以 Node02 作为 NFS 服务器：

```bash
mkdir /data/nfs_share
echo "/data/nfs_share *(rw,sync,no_root_squash,no_subtree_check)" > /etc/exports
systemctl enable --now nfs-server
# /data/nfs_share *
```

验证共享目录：

```bash
showmount -e localhost		# Node02
showmount -e 10.3.202.102	# other
```

## 创建 StorageClass

helm 创建 nfs-provisioner：

```bash
helm repo add nfs-subdir-external-provisioner https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner/
helm repo update
```

```bash
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
```

验证 nfs-provisioner：

```bash
kubectl get pod -n nfs-provisioner -w
# NAME                                                              READY   STATUS    RESTARTS   AGE
# nfs-provisioner-nfs-subdir-external-provisioner-7d88f5d58-xxxxx   1/1     Running   0          30s

kubectl get storageclass
```

若上述方法出现镜像下载错误，可移除 `image` 的两个选项。



# Jenkins

```bash
kubectl apply -f deploy/jenkins
```



# GitLab

确保调度节点内存大于 4GB。

```bash
kubectl apply -f deploy/gitlab
```

























