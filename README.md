# URL Shortener


一个云原生高可用短链接生成服务系统，提供短链生成、跳转、访问统计等功能，集成自动化CI/CD流水线。

## **📖 项目简介**  
本服务提供以下核心功能：  
- **短链生成与重定向**：将长 URL 转换为短链并记录访问次数，短链访问时自动跳转至原始 URL。
- **Docker 容器化**：提供优化的 Dockerfile 构建生产镜像。
- **Kubernetes 部署**：支持高可用集群部署。
- **弹性伸缩**：基于 CPU/请求量自动扩缩 Pod 实例。 
- **适用场景**：社交媒体分享、广告跟踪、内部链接管理等。
- **监控告警**：集成 Prometheus 采集指标，Grafana 可视化监控。

---


### **🔗 相关资源**   
 [部署指南](https://github.com/strayca7/URL-Shortener/wiki/Deploy)（支持 Docker 和 Kubernetes）  

---



## 快速开始

需要在外部手动配置 MySQL 数据库。（数据库配置详请 [config.yaml](https://github.com/strayca7/URL-Shortener/blob/main/config.yaml)）

也可使用 [初始化脚本](https://github.com/strayca7/URL-Shortener/blob/main/script/initmysqldb.sql) 。

```bash
docker build docker build -f Dockerfile.arm64 -t url-shorten:arm64/0.0.1 .
```

```bash
docker run --rm -d -p 8080:8080 -v ./config.yaml:/app/config.yaml url-shorten:arm64/0.0.1
```

