# GitLabEx Kubernetes 快速参考

## 常用命令速查

### 部署和管理

```bash
# 一键部署
cd k8s && ./deploy.sh

# 一键卸载
cd k8s && ./undeploy.sh

# 配置 OAuth
cd k8s && ./configure-oauth-k8s.sh

# 使用 kustomize 部署
kubectl apply -k k8s/

# 查看所有资源
kubectl get all -n gitlabex
```

### Pod 管理

```bash
# 查看所有 Pod
kubectl get pods -n gitlabex

# 查看特定应用的 Pod
kubectl get pods -n gitlabex -l app=backend
kubectl get pods -n gitlabex -l app=frontend
kubectl get pods -n gitlabex -l app=gitlab

# 查看 Pod 详细信息
kubectl describe pod <pod-name> -n gitlabex

# 查看 Pod 日志
kubectl logs -f <pod-name> -n gitlabex

# 查看前 100 行日志
kubectl logs --tail=100 <pod-name> -n gitlabex

# 进入 Pod
kubectl exec -it <pod-name> -n gitlabex -- /bin/bash

# 删除 Pod（会自动重建）
kubectl delete pod <pod-name> -n gitlabex
```

### 服务管理

```bash
# 查看所有服务
kubectl get svc -n gitlabex

# 查看服务详情
kubectl describe svc <service-name> -n gitlabex

# 获取服务访问地址
kubectl get svc -n gitlabex -o wide
```

### 部署管理

```bash
# 查看所有部署
kubectl get deployments -n gitlabex

# 扩缩容
kubectl scale deployment backend -n gitlabex --replicas=3

# 更新镜像
kubectl set image deployment/backend backend=gitlabex-backend:v2 -n gitlabex

# 查看滚动更新状态
kubectl rollout status deployment/backend -n gitlabex

# 回滚
kubectl rollout undo deployment/backend -n gitlabex

# 查看历史版本
kubectl rollout history deployment/backend -n gitlabex

# 重启部署
kubectl rollout restart deployment/backend -n gitlabex
```

### 配置管理

```bash
# 查看 ConfigMap
kubectl get configmap -n gitlabex
kubectl describe configmap backend-config -n gitlabex

# 查看 Secret
kubectl get secrets -n gitlabex
kubectl describe secret gitlabex-secrets -n gitlabex

# 编辑 ConfigMap
kubectl edit configmap backend-config -n gitlabex

# 编辑 Secret
kubectl edit secret gitlabex-secrets -n gitlabex

# 从文件更新 ConfigMap
kubectl create configmap backend-config --from-file=config.yml -n gitlabex --dry-run=client -o yaml | kubectl apply -f -
```

### 存储管理

```bash
# 查看 PVC
kubectl get pvc -n gitlabex

# 查看 PVC 详情
kubectl describe pvc <pvc-name> -n gitlabex

# 查看 PV
kubectl get pv
```

### 日志和监控

```bash
# 查看资源使用情况
kubectl top nodes
kubectl top pods -n gitlabex

# 查看事件
kubectl get events -n gitlabex --sort-by='.lastTimestamp'

# 查看最近的事件
kubectl get events -n gitlabex --sort-by='.lastTimestamp' | tail -20

# 实时监控 Pod
watch kubectl get pods -n gitlabex
```

### 故障排查

```bash
# 查看 Pod 状态
kubectl get pods -n gitlabex -o wide

# 查看 Pod 详细信息（包含错误信息）
kubectl describe pod <pod-name> -n gitlabex

# 查看容器日志
kubectl logs <pod-name> -n gitlabex
kubectl logs <pod-name> -c <container-name> -n gitlabex

# 查看上一个容器的日志（崩溃重启后）
kubectl logs <pod-name> -n gitlabex --previous

# 进入容器调试
kubectl exec -it <pod-name> -n gitlabex -- /bin/sh

# 运行临时调试 Pod
kubectl run -it --rm debug --image=busybox --restart=Never -n gitlabex -- sh

# 端口转发（本地调试）
kubectl port-forward <pod-name> 8080:8080 -n gitlabex

# 复制文件
kubectl cp <pod-name>:/path/to/file /local/path -n gitlabex
kubectl cp /local/path <pod-name>:/path/to/file -n gitlabex
```

### 网络测试

```bash
# 测试 DNS 解析
kubectl run -it --rm debug --image=busybox --restart=Never -n gitlabex -- nslookup gitlabex-postgres

# 测试服务连接
kubectl run -it --rm debug --image=busybox --restart=Never -n gitlabex -- wget -O- http://gitlabex-backend:8080/health

# 测试数据库连接
kubectl run -it --rm debug --image=postgres:15 --restart=Never -n gitlabex -- \
  psql -h gitlabex-postgres -U gitlab -d gitlab -c "SELECT version();"
```

## 服务端口映射

| 服务 | 内部端口 | NodePort | 说明 |
|------|----------|----------|------|
| Frontend | 80 | 30000 | 前端 Web 界面 |
| Backend | 8080 | 30080 | 后端 API |
| Backend Metrics | 9090 | 30090 | Prometheus 指标 |
| GitLab HTTP | 80 | 30081 | GitLab Web |
| GitLab SSH | 22 | 30222 | Git SSH 访问 |
| MinIO Console | 9001 | 30901 | MinIO 管理界面 |

## 访问地址

```bash
# 获取节点 IP
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')

# 前端
echo "Frontend: http://${NODE_IP}:30000"

# 后端 API
echo "Backend API: http://${NODE_IP}:30080"

# GitLab
echo "GitLab: http://${NODE_IP}:30081"

# MinIO Console
echo "MinIO Console: http://${NODE_IP}:30901"
```

## 默认凭据

### GitLab
- 用户名: `root`
- 密码: 在 `secrets.yaml` 中的 `gitlab-root-password`

### MinIO
- 用户名: `admin`
- 密码: 在 `secrets.yaml` 中的 `minio-root-password`

### PostgreSQL
- 用户名: `gitlab`
- 密码: 在 `secrets.yaml` 中的 `postgres-password`

### Redis
- 密码: 在 `secrets.yaml` 中的 `redis-password`

## 常见操作场景

### 场景 1: 更新后端代码

```bash
# 1. 构建新镜像
cd backend && docker build -t gitlabex-backend:v2 .

# 2. 推送镜像（如使用私有仓库）
docker push your-registry.com/gitlabex-backend:v2

# 3. 更新部署
kubectl set image deployment/backend backend=gitlabex-backend:v2 -n gitlabex

# 4. 查看滚动更新状态
kubectl rollout status deployment/backend -n gitlabex

# 5. 验证更新
kubectl get pods -n gitlabex -l app=backend
```

### 场景 2: 扩展后端服务

```bash
# 扩展到 3 个副本
kubectl scale deployment backend -n gitlabex --replicas=3

# 验证
kubectl get pods -n gitlabex -l app=backend

# 配置自动扩缩容
kubectl autoscale deployment backend -n gitlabex --cpu-percent=70 --min=2 --max=10
```

### 场景 3: 查看服务日志

```bash
# 查看后端日志
kubectl logs -f deployment/backend -n gitlabex

# 查看 GitLab 日志
kubectl logs -f deployment/gitlab -n gitlabex

# 查看所有后端 Pod 的日志
kubectl logs -l app=backend -n gitlabex --tail=100
```

### 场景 4: 重启服务

```bash
# 重启后端（滚动重启，不影响服务）
kubectl rollout restart deployment/backend -n gitlabex

# 重启前端
kubectl rollout restart deployment/frontend -n gitlabex

# 重启 GitLab（会有短暂中断）
kubectl rollout restart deployment/gitlab -n gitlabex
```

### 场景 5: 备份数据库

```bash
# 备份 PostgreSQL
kubectl exec -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  pg_dump -U gitlab gitlab > gitlab-backup-$(date +%Y%m%d).sql

kubectl exec -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  pg_dump -U gitlab gitlabex > gitlabex-backup-$(date +%Y%m%d).sql

# 恢复数据库
kubectl exec -i -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  psql -U gitlab gitlab < gitlab-backup-20240101.sql
```

### 场景 6: 更新配置

```bash
# 1. 修改 ConfigMap 或 Secret
kubectl edit configmap backend-config -n gitlabex
# 或
kubectl edit secret gitlabex-secrets -n gitlabex

# 2. 重启服务以应用新配置
kubectl rollout restart deployment/backend -n gitlabex

# 3. 验证
kubectl rollout status deployment/backend -n gitlabex
```

### 场景 7: 故障排查

```bash
# 1. 查看 Pod 状态
kubectl get pods -n gitlabex

# 2. 查看异常 Pod 详情
kubectl describe pod <failing-pod> -n gitlabex

# 3. 查看日志
kubectl logs <failing-pod> -n gitlabex

# 4. 查看事件
kubectl get events -n gitlabex --sort-by='.lastTimestamp' | tail -20

# 5. 进入容器调试
kubectl exec -it <pod-name> -n gitlabex -- /bin/bash

# 6. 检查资源使用
kubectl top pod <pod-name> -n gitlabex
```

### 场景 8: 清理和重新部署

```bash
# 1. 删除所有应用（保留 PVC）
kubectl delete -f frontend.yaml
kubectl delete -f backend.yaml
kubectl delete -f gitlab.yaml
kubectl delete -f minio.yaml
kubectl delete -f redis.yaml
kubectl delete -f postgres.yaml

# 2. 重新部署
./deploy.sh

# 完全清理（包括数据）
./undeploy.sh
```

## 性能优化技巧

### 1. 资源限制调优

```bash
# 编辑部署，调整资源限制
kubectl edit deployment backend -n gitlabex

# 或直接 patch
kubectl patch deployment backend -n gitlabex -p '{"spec":{"template":{"spec":{"containers":[{"name":"backend","resources":{"requests":{"memory":"1Gi","cpu":"1000m"},"limits":{"memory":"2Gi","cpu":"2000m"}}}]}}}}'
```

### 2. 启用 HPA

```bash
# 为后端创建 HPA
kubectl autoscale deployment backend -n gitlabex --cpu-percent=70 --min=2 --max=10

# 查看 HPA 状态
kubectl get hpa -n gitlabex
```

### 3. 配置亲和性

```yaml
# 在 deployment.yaml 中添加
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app
            operator: In
            values:
            - backend
        topologyKey: kubernetes.io/hostname
```

## 安全最佳实践

```bash
# 1. 查看 Secret（base64 解码）
kubectl get secret gitlabex-secrets -n gitlabex -o jsonpath='{.data.postgres-password}' | base64 -d

# 2. 更新 Secret
kubectl create secret generic gitlabex-secrets \
  --from-literal=postgres-password=newpassword \
  --dry-run=client -o yaml | kubectl apply -f -

# 3. 轮换密钥后重启服务
kubectl rollout restart deployment/backend -n gitlabex

# 4. 检查安全上下文
kubectl get pod <pod-name> -n gitlabex -o jsonpath='{.spec.securityContext}'
```

## 监控和告警

```bash
# 查看资源使用
kubectl top pods -n gitlabex
kubectl top nodes

# 查看 Metrics Server
kubectl get --raw /apis/metrics.k8s.io/v1beta1/nodes
kubectl get --raw /apis/metrics.k8s.io/v1beta1/namespaces/gitlabex/pods

# 访问 Backend Metrics（Prometheus 格式）
curl http://<NODE_IP>:30090/metrics
```

## 有用的别名

将以下内容添加到 `~/.bashrc` 或 `~/.zshrc`：

```bash
# Kubernetes 别名
alias k='kubectl'
alias kn='kubectl -n gitlabex'
alias kgp='kubectl get pods -n gitlabex'
alias kgs='kubectl get svc -n gitlabex'
alias kgd='kubectl get deployments -n gitlabex'
alias kl='kubectl logs -f -n gitlabex'
alias kex='kubectl exec -it -n gitlabex'
alias kdel='kubectl delete -n gitlabex'
alias kdes='kubectl describe -n gitlabex'

# GitLabEx 特定
alias gitlabex-pods='kubectl get pods -n gitlabex -o wide'
alias gitlabex-logs-backend='kubectl logs -f deployment/backend -n gitlabex'
alias gitlabex-logs-frontend='kubectl logs -f deployment/frontend -n gitlabex'
alias gitlabex-restart-backend='kubectl rollout restart deployment/backend -n gitlabex'
alias gitlabex-restart-frontend='kubectl rollout restart deployment/frontend -n gitlabex'
```

## 常见错误和解决方案

### ImagePullBackOff

```bash
# 检查镜像是否存在
docker images | grep gitlabex

# 检查 ImagePullSecrets
kubectl get pod <pod-name> -n gitlabex -o jsonpath='{.spec.imagePullSecrets}'

# 创建 ImagePullSecret（如需要）
kubectl create secret docker-registry regcred \
  --docker-server=<registry> \
  --docker-username=<username> \
  --docker-password=<password> \
  -n gitlabex
```

### CrashLoopBackOff

```bash
# 查看日志
kubectl logs <pod-name> -n gitlabex
kubectl logs <pod-name> -n gitlabex --previous

# 查看详细信息
kubectl describe pod <pod-name> -n gitlabex
```

### PVC Pending

```bash
# 检查 StorageClass
kubectl get storageclass

# 检查 PVC 详情
kubectl describe pvc <pvc-name> -n gitlabex

# 检查是否有可用的 PV
kubectl get pv
```

## 相关链接

- [完整部署文档](README.md)
- [部署检查清单](DEPLOY_CHECKLIST.md)
- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [kubectl 速查表](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
