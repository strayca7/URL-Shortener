# URL Shortener

A cloud-native high-availability short link generation service system that provides short link generation, redirection, and access statistics functions, integrated with automated CI/CD pipeline.

中文文档 [README_ZH](https://github.com/strayca7/URL-Shortener/blob/main/doc/README_ZH.md)

## Introduction    
This service provides the following core functions:
- **Short url generation and redirection**: Convert long URLs into short links and record access counts. Short links redirect to original URLs automatically.
- **Docker containerization**: Provide optimized Dockerfile for production image.
- **Kuberntes deployment**: Support high availability cluster deployment. 
- **Autoscaling**: Based on CPU/request load, automatically scale Pod instances.
- **Monitoring and Alerting**: Integrate Prometheus to collect metrics, Grafana to visualize monitoring.
- **Scenarios**: Social media sharing, ad tracking, internal links, etc.

---


## Doc
 [Deployment guide](https://github.com/strayca7/URL-Shortener/wiki/Deploy)

---



## Getting Started

User need to manually configure MySQL database, database configuration details [config.yaml](https://github.com/strayca7/URL-Shortener/blob/main/config.yaml).

You can also use [initialization script](https://github.com/strayca7/URL-Shortener/blob/main/script/initmysqldb.sql) .

```bash
docker build -f Dockerfile -t url-shortener:v0.0.1 .
```

```bash
docker run --rm -p 8080:8080 -v ./config.yaml:/app/config.yaml url-shorten:v0.0.1
```
Or you can use the images already built on [dockerhub](https://hub.docker.com/repository/docker/strayca7/url-shortener/general).
