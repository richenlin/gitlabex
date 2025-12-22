# Kubernetes 部署快速参考

快速查询常用命令和配置信息。

## 🚀 快速部署

```bash
# 1. 准备 Secret
cp secrets.yaml.example secrets.yaml
vim secrets.yaml  # 修改所有密码

# 2. 部署基础设施
kubectl apply -f namespace.yaml
kubectl apply -f secrets.yaml

# 3. 部署数据层
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f minio.yaml

# 4. 部署 GitLab（等待 5-10 分钟）
kubectl apply -f gitlab.yaml

# 5. 配置 OAuth 后部署应用
kubectl apply -f backend.yaml
kubectl apply -f frontend.yaml

# 6. 重启 backend 使 OAuth 生效
kubectl rollout restart deployment/backend -n gitlabex
```

## 📋 常用命令

### 查看状态
```bash
# 查看所有 Pod
kubectl get pods -n gitlabex

# 查看所有服务
kubectl get svc -n gitlabex

# 查看 PVC 状态
kubectl get pvc -n gitlabex

# 实时监控 Pod 状态
kubectl get pods -n gitlabex -w
```

### 查看日志
```bash
# 查看特定 Pod 日志
kubectl logs -f <pod-name> -n gitlabex

# 查看应用日志
kubectl logs -f -n gitlabex -l app=backend
kubectl logs -f -n gitlabex -l app=frontend
kubectl logs -f -n gitlabex -l app=gitlab
```

### 进入容器
```bash
# 进入 PostgreSQL
kubectl exec -it $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -n gitlabex -- psql -U gitlab

# 进入 Redis
kubectl exec -it $(kubectl get pod -n gitlabex -l app=redis -o jsonpath='{.items[0].metadata.name}') -n gitlabex -- redis-cli -a <password>

# 进入 Backend
kubectl exec -it $(kubectl get pod -n gitlabex -l app=backend -o jsonpath='{.items[0].metadata.name}') -n gitlabex -- sh
```

### 重启服务
```bash
# 滚动重启（无中断）
kubectl rollout restart deployment/backend -n gitlabex
kubectl rollout restart deployment/frontend -n gitlabex
kubectl rollout restart deployment/gitlab -n gitlabex

# 查看重启状态
kubectl rollout status deployment/backend -n gitlabex
```

### 更新配置
```bash
# 更新 Secret
vim secrets.yaml
kubectl apply -f secrets.yaml
kubectl rollout restart deployment/backend -n gitlabex

# 更新 ConfigMap
kubectl edit configmap backend-config -n gitlabex
kubectl rollout restart deployment/backend -n gitlabex
```

### 扩缩容
```bash
# 手动扩容
kubectl scale deployment backend -n gitlabex --replicas=5

# 查看副本数
kubectl get deployment backend -n gitlabex
```

## 🔍 故障排查

### Pod 处于 Pending
```bash
# 查看详细信息
kubectl describe pod <pod-name> -n gitlabex

# 检查资源
kubectl top nodes
kubectl get pvc -n gitlabex
```

### Pod 处于 CrashLoopBackOff
```bash
# 查看日志
kubectl logs <pod-name> -n gitlabex
kubectl logs <pod-name> -n gitlabex --previous

# 查看事件
kubectl get events -n gitlabex --sort-by='.lastTimestamp'
```

### 服务无法访问
```bash
# 检查 Service
kubectl get svc -n gitlabex
kubectl describe svc <service-name> -n gitlabex

# 检查 Pod
kubectl get pods -n gitlabex -l app=<app-name>

# 端口转发测试
kubectl port-forward -n gitlabex <pod-name> 8080:8080
```

## 🗑️ 清理

### 完全删除（包括数据）
```bash
# 删除所有资源
kubectl delete namespace gitlabex

# 或逐个删除
kubectl delete -f frontend.yaml
kubectl delete -f backend.yaml
kubectl delete -f gitlab.yaml
kubectl delete -f minio.yaml
kubectl delete -f redis.yaml
kubectl delete -f postgres.yaml
kubectl delete -f secrets.yaml
kubectl delete -f namespace.yaml
```

### 保留数据删除
```bash
# 只删除应用，保留 PVC
kubectl delete deployment --all -n gitlabex
kubectl delete service --all -n gitlabex
kubectl delete configmap --all -n gitlabex
kubectl delete secret --all -n gitlabex

# PVC 仍然存在
kubectl get pvc -n gitlabex
```

## 🌐 访问地址

```bash
# 获取节点 IP
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')

# 服务地址
echo "前端:        http://${NODE_IP}:30000"
echo "后端 API:    http://${NODE_IP}:30080"
echo "后端 Metrics: http://${NODE_IP}:30090"
echo "GitLab:      http://${NODE_IP}:30081"
echo "GitLab SSH:  ssh://git@${NODE_IP}:30222"
echo "MinIO:       http://${NODE_IP}:30901"
```

## 💾 备份

### 备份数据库
```bash
# 备份 gitlab 数据库
kubectl exec -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  pg_dump -U gitlab gitlab | gzip > backup-gitlab-$(date +%Y%m%d).sql.gz

# 备份 gitlabex 数据库
kubectl exec -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  pg_dump -U gitlab gitlabex | gzip > backup-gitlabex-$(date +%Y%m%d).sql.gz
```

### 备份配置
```bash
# 备份所有资源配置
kubectl get all,pvc,secrets,configmaps -n gitlabex -o yaml | gzip > backup-k8s-$(date +%Y%m%d).yaml.gz
```

### 恢复数据库
```bash
# 恢复数据库
gunzip -c backup-gitlab-20241222.sql.gz | \
  kubectl exec -i -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  psql -U gitlab gitlab
```

## 📊 监控

### 资源使用
```bash
# 节点资源
kubectl top nodes

# Pod 资源
kubectl top pods -n gitlabex

# 按资源排序
kubectl top pods -n gitlabex --sort-by=memory
kubectl top pods -n gitlabex --sort-by=cpu
```

### 健康检查
```bash
# 后端健康检查
curl http://${NODE_IP}:30080/health

# 后端 Metrics
curl http://${NODE_IP}:30090/metrics
```

## 🔐 Secret 管理

### 查看 Secret
```bash
# 列出所有 key
kubectl describe secret gitlabex-secrets -n gitlabex

# 查看特定值（base64 编码）
kubectl get secret gitlabex-secrets -n gitlabex -o jsonpath='{.data.postgres-password}'

# 解码查看（调试用）
kubectl get secret gitlabex-secrets -n gitlabex -o jsonpath='{.data.postgres-password}' | base64 -d
```

### 生成新密码
```bash
# 生成随机密码
openssl rand -base64 32

# 转换为 base64
echo -n "your-password" | base64

# 验证解码
echo "base64-string" | base64 -d
```

## ⚠️ 重要提示

### Secret 更新后必须重启
```bash
kubectl apply -f secrets.yaml
kubectl rollout restart deployment/backend -n gitlabex
```

### PostgreSQL 初始化只执行一次
```bash
# 需要重新初始化时
kubectl delete -f postgres.yaml
kubectl delete pvc postgres-pvc -n gitlabex
kubectl apply -f postgres.yaml
```

### 验证数据库权限
```bash
kubectl exec -it -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  psql -U gitlab gitlabex -c "CREATE TABLE test(id INT); DROP TABLE test;"
```

## 📚 相关文档

- [README.md](README.md) - 完整部署指南
- [DEPLOYMENT_COMPARISON.md](DEPLOYMENT_COMPARISON.md) - Docker Compose vs K8s 对比
- [FIXES_SUMMARY.md](FIXES_SUMMARY.md) - 配置修复总结

---

**提示**：将 `<pod-name>`, `<service-name>`, `<app-name>` 等替换为实际值。
