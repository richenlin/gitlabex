# GitLabEx Kubernetes 部署指南

本文档为 Kubernetes 初学者提供详细的分步部署说明，帮助你在 Kubernetes 集群上成功部署 GitLabEx 教育协作平台。

## 📚 学习目标

通过本教程,你将学会:
- 理解 Kubernetes 的基本概念 (Pod, Service, Deployment, PVC 等)
- 如何使用 kubectl 命令行工具操作集群
- 如何部署一个完整的多层应用到 Kubernetes
- 如何管理和排查 Kubernetes 应用

## 📁 目录结构

```
k8s/
├── README.md                    # 本文档 - 详细的部署指南
├── namespace.yaml               # 命名空间定义 - 资源隔离
├── secrets.yaml.example         # Secret 配置示例 - 敏感信息存储
├── postgres.yaml                # PostgreSQL 数据库部署
├── redis.yaml                   # Redis 缓存部署
├── minio.yaml                   # MinIO 对象存储部署
├── gitlab.yaml                  # GitLab 服务部署
├── backend.yaml                 # 后端服务部署
├── frontend.yaml                # 前端服务部署
├── ingress.yaml                 # Ingress 配置（可选）- 统一入口
├── deploy.sh                    # 一键部署脚本（高级用户）
├── undeploy.sh                  # 一键卸载脚本（高级用户）
└── configure-oauth-k8s.sh       # OAuth 配置脚本（高级用户）
```

## 🎓 Kubernetes 基础概念

在开始部署前,让我们先了解几个核心概念:

### 1. Namespace (命名空间)
- **作用**: 在同一个集群中隔离不同的应用和资源
- **类比**: 就像文件系统中的文件夹,将不同项目的资源分开管理
- **本项目**: 我们将创建一个名为 `gitlabex` 的命名空间

### 2. Pod (容器组)
- **作用**: Kubernetes 中最小的部署单元,包含一个或多个容器
- **类比**: 类似于 Docker Compose 中的一个服务实例
- **本项目**: 每个服务(如 PostgreSQL, Redis)都会运行在独立的 Pod 中

### 3. Deployment (部署)
- **作用**: 管理 Pod 的创建、更新和副本数量
- **类比**: 自动化的应用管理器,确保应用始终运行指定数量的副本
- **本项目**: 所有服务都通过 Deployment 管理

### 4. Service (服务)
- **作用**: 为 Pod 提供稳定的网络访问入口
- **类比**: 类似于负载均衡器,将请求分发到后端的 Pod
- **类型**: 
  - ClusterIP: 仅集群内部访问
  - NodePort: 通过节点 IP + 端口对外暴露
  - LoadBalancer: 使用云提供商的负载均衡器

### 5. PersistentVolumeClaim (持久卷声明) - PVC
- **作用**: 请求持久化存储空间
- **类比**: 向管理员申请硬盘空间
- **本项目**: 用于存储数据库数据、文件上传等需要持久化的内容

### 6. Secret (密钥)
- **作用**: 存储敏感信息(密码、密钥等)
- **类比**: 保险箱,安全地存储不应明文暴露的信息
- **本项目**: 存储所有数据库密码、API 密钥等

### 7. ConfigMap (配置映射)
- **作用**: 存储非敏感的配置信息
- **类比**: 配置文件管理器
- **本项目**: 存储应用配置、初始化脚本等

## ✅ 部署前准备

### 第 1 步: 验证 Kubernetes 集群

#### 1.1 检查集群状态

首先,确保你已经有一个可用的 Kubernetes 集群。如果没有,可以使用以下工具创建本地测试集群:
- **Minikube**: 单节点集群,适合学习和测试
- **Kind**: 使用 Docker 容器运行的 Kubernetes
- **K3s**: 轻量级 Kubernetes,适合边缘计算和开发

```bash
# 检查 kubectl 是否已安装
kubectl version --client

# 查看集群信息
kubectl cluster-info

# 查看集群节点
kubectl get nodes

# 期望输出示例:
# NAME          STATUS   ROLES           AGE   VERSION
# k8s-node-1    Ready    control-plane   10d   v1.28.0
```

**解释**:
- `kubectl version`: 显示 kubectl 客户端版本
- `kubectl cluster-info`: 显示集群的主节点地址
- `kubectl get nodes`: 列出所有节点及其状态,STATUS 应该是 Ready

#### 1.2 检查节点资源

```bash
# 查看节点详细信息（包括 CPU 和内存）
kubectl describe nodes

# 查看节点资源使用情况
kubectl top nodes

# 如果 top 命令不可用,需要安装 metrics-server:
# kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

**最低要求**:
- Kubernetes 版本: 1.20+
- 节点数量: 至少 1 个节点（测试环境），推荐 3 个节点（生产环境）
- 每个节点: 至少 4 核 CPU + 8GB 内存

### 第 2 步: 了解资源需求

#### 测试环境配置（最低要求）:

| 组件 | CPU 请求 | 内存请求 | 存储 | 说明 |
|------|---------|---------|------|------|
| PostgreSQL | 0.5 核 | 512 MB | 20 GB | 关系数据库 |
| Redis | 0.25 核 | 256 MB | 5 GB | 缓存服务 |
| MinIO | 0.5 核 | 512 MB | 50 GB | 对象存储 |
| GitLab | 2 核 | 4 GB | 111 GB | Git 仓库管理 |
| Backend (x2) | 1 核 | 1 GB | 25 GB | 后端 API 服务 |
| Frontend (x2) | 0.2 核 | 256 MB | - | 前端 Web 界面 |
| **总计** | **约 4.5 核** | **约 6.5 GB** | **约 211 GB** | |

#### 生产环境配置（推荐）:

| 组件 | CPU 请求 | 内存请求 | 存储 | 说明 |
|------|---------|---------|------|------|
| PostgreSQL | 2 核 | 2 GB | 100 GB | 更高性能 |
| Redis | 1 核 | 1 GB | 10 GB | 更大缓存 |
| MinIO | 2 核 | 2 GB | 500 GB | 更大容量 |
| GitLab | 4 核 | 8 GB | 500 GB | 支持更多用户 |
| Backend (x3) | 6 核 | 6 GB | 100 GB | 更好的并发性能 |
| Frontend (x3) | 1.5 核 | 768 MB | - | 更好的用户体验 |
| **总计** | **约 16.5 核** | **约 19.8 GB** | **约 1.2 TB** | |

**说明**:
- CPU 以"核"为单位,1 核 = 1000m (millicores)
- 内存以 MB/GB 为单位
- 存储需要集群支持动态卷供应（Dynamic Volume Provisioning）

### 第 3 步: 检查存储类（StorageClass）

Kubernetes 需要 StorageClass 来动态创建持久卷 (PersistentVolume)。

```bash
# 查看集群中可用的 StorageClass
kubectl get storageclass

# 期望输出示例:
# NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      AGE
# standard (default)   kubernetes.io/gce-pd    Delete          Immediate              30d
# fast                 kubernetes.io/gce-pd    Delete          WaitForFirstConsumer   30d
```

**解释**:
- **NAME**: 存储类的名称
- **PROVISIONER**: 存储提供商(如 AWS EBS, GCE PD, 本地存储等)
- **RECLAIMPOLICY**: 删除策略 (Delete 或 Retain)
- **(default)**: 标记为默认存储类

**如果没有 StorageClass**:
- **Minikube**: 自动提供 `standard` 存储类
- **Kind**: 使用 `local-path-provisioner`
- **云提供商**: 通常提供默认存储类
- **自建集群**: 需要手动安装存储插件 (如 NFS, Ceph, local-path-provisioner)

```bash
# 查看某个 StorageClass 的详细信息
kubectl describe storageclass standard
```

### 第 4 步: 准备 Docker 镜像

GitLabEx 需要以下镜像:

#### 4.1 拉取公共镜像

这些镜像可以直接从 Docker Hub 拉取:

```bash
# 拉取数据库镜像
docker pull postgres:15

# 拉取缓存镜像  
docker pull redis:7-alpine

# 拉取对象存储镜像
docker pull minio/minio:RELEASE.2024-03-07T00-43-48Z

# 拉取 GitLab 镜像（较大，约 2.5GB，需要耐心等待）
docker pull gitlab/gitlab-ce:15.11.13-ce.0
```

**解释**:
- `docker pull`: 从镜像仓库下载镜像
- 镜像名格式: `仓库名:标签`
- 下载后的镜像会缓存在本地，Kubernetes 可以直接使用

#### 4.2 构建应用镜像

GitLabEx 的后端和前端需要自己构建:

```bash
# 返回项目根目录
cd /path/to/gitlabex2

# 构建镜像（会同时构建 backend 和 frontend）
./scripts/build-images.sh

# 验证镜像已创建
docker images | grep gitlabex

# 期望输出:
# gitlabex-backend   latest   abc123def456   2 minutes ago   50MB
# gitlabex-frontend  latest   789ghi012jkl   1 minute ago    25MB
```

**解释**:
- `build-images.sh`: 项目提供的镜像构建脚本
- 脚本会读取 Dockerfile 并构建镜像
- 构建完成后，镜像保存在本地 Docker 中

#### 4.3 处理镜像仓库（可选）

如果你的 Kubernetes 集群和 Docker 在不同的机器上,需要:

**方案A: 使用私有镜像仓库**
```bash
# 标记镜像
docker tag gitlabex-backend:latest your-registry.com/gitlabex-backend:latest
docker tag gitlabex-frontend:latest your-registry.com/gitlabex-frontend:latest

# 推送到私有仓库
docker push your-registry.com/gitlabex-backend:latest
docker push your-registry.com/gitlabex-frontend:latest

# 然后修改 YAML 文件中的 image 字段为: your-registry.com/gitlabex-backend:latest
```

**方案B: 导出/导入镜像**
```bash
# 在构建机器上导出镜像
docker save gitlabex-backend:latest -o backend.tar
docker save gitlabex-frontend:latest -o frontend.tar

# 复制到 Kubernetes 节点
scp backend.tar frontend.tar node-ip:/tmp/

# 在 Kubernetes 节点上导入
docker load -i /tmp/backend.tar
docker load -i /tmp/frontend.tar
```

**方案C: 本地开发（Minikube/Kind）**
```bash
# 如果使用 Minikube，直接使用 Minikube 的 Docker 守护进程
eval $(minikube docker-env)
./scripts/build-images.sh

# 如果使用 Kind，加载镜像到 Kind 集群
kind load docker-image gitlabex-backend:latest
kind load docker-image gitlabex-frontend:latest
```

---

## 🚀 开始部署

现在我们开始实际的部署过程。我们将按照服务的依赖顺序,逐步部署每个组件。

### 部署步骤 1: 创建命名空间

命名空间用于隔离 GitLabEx 的所有资源。

#### 1.1 查看命名空间配置文件

```bash
# 进入 k8s 目录
cd k8s

# 查看 namespace.yaml 文件内容
cat namespace.yaml
```

你会看到类似的内容:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: gitlabex
  labels:
    name: gitlabex
    environment: production
```

**配置说明**:
- `apiVersion: v1`: 使用 Kubernetes API 的 v1 版本
- `kind: Namespace`: 资源类型是命名空间
- `metadata.name: gitlabex`: 命名空间的名称
- `labels`: 附加的标签,用于组织和筛选资源

#### 1.2 创建命名空间

```bash
# 应用配置文件
kubectl apply -f namespace.yaml

# 期望输出:
# namespace/gitlabex created
```

**解释**:
- `kubectl apply`: 创建或更新资源
- `-f namespace.yaml`: 指定配置文件

#### 1.3 验证命名空间

```bash
# 查看所有命名空间
kubectl get namespaces

# 或者简写
kubectl get ns

# 查看 gitlabex 命名空间的详细信息
kubectl describe namespace gitlabex
```

**技巧**: 后续所有命令都需要加上 `-n gitlabex` 来指定命名空间,你也可以设置默认命名空间:
```bash
# 设置当前上下文的默认命名空间
kubectl config set-context --current --namespace=gitlabex

# 验证
kubectl config view --minify | grep namespace:
```

### 部署步骤 2: 创建 Secret（敏感信息）

Secret 用于存储所有密码、密钥等敏感信息。

#### 2.1 准备 Secret 配置文件

```bash
# 复制示例文件
cp secrets.yaml.example secrets.yaml

# 查看示例文件
cat secrets.yaml.example
```

**重要**: `secrets.yaml` 包含所有敏感信息,已被 `.gitignore` 忽略,不会提交到版本控制!

#### 2.2 理解 Secret 的结构

Secret 文件包含多个 key-value 对,所有值都需要 **base64 编码**:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: gitlabex-secrets
  namespace: gitlabex
type: Opaque
data:
  # 所有的值都是 base64 编码的
  postgres-password: cGFzc3dvcmQxMjM=  # 原始值: password123
  redis-password: cGFzc3dvcmQxMjM=     # 原始值: password123
  # ... 更多密钥
```

**为什么需要 base64 编码?**
- Kubernetes Secret 要求值必须是 base64 编码
- base64 编码不是加密,只是编码方式
- Kubernetes 会自动解码后注入到容器中

#### 2.3 生成新的密码

**强烈建议**: 为生产环境生成新的强密码,不要使用示例中的默认值!

```bash
# 方法1: 生成随机密码（推荐）
openssl rand -base64 32

# 输出示例: xK7mN9pQ2vR8sT4uV6wX1yZ3aB5cD7eF

# 方法2: 生成简单随机字符串
cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1

# 输出示例: aB3dE5fG7hI9jK1lM3nO5pQ7rS9tU
```

#### 2.4 将密码转换为 base64

```bash
# 将字符串转换为 base64（务必使用 -n 参数，避免换行符）
echo -n "your-new-password" | base64

# 输出示例: eW91ci1uZXctcGFzc3dvcmQ=

# 验证解码（确保编码正确）
echo "eW91ci1uZXctcGFzc3dvcmQ=" | base64 -d

# 输出应该是: your-new-password
```

**关键步骤**:
1. 生成新密码
2. 使用 `echo -n "密码" | base64` 转换为 base64
3. 将 base64 值复制到 `secrets.yaml` 对应的字段
4. 验证解码是否正确

#### 2.5 编辑 secrets.yaml

使用文本编辑器打开 `secrets.yaml`,替换所有需要修改的值:

```bash
# 使用 vi/vim 编辑
vim secrets.yaml

# 或使用 nano（更简单）
nano secrets.yaml

# 或使用 VS Code
code secrets.yaml
```

**需要修改的关键字段**:

| 字段名 | 说明 | 如何生成 |
|--------|------|----------|
| `postgres-password` | PostgreSQL 密码 | `echo -n "新密码" \| base64` |
| `redis-password` | Redis 密码 | `echo -n "新密码" \| base64` |
| `gitlab-root-password` | GitLab root 用户密码 | `echo -n "新密码" \| base64` |
| `minio-root-password` | MinIO 管理密码 | `echo -n "新密码" \| base64` |
| `jwt-secret` | JWT 签名密钥 | `openssl rand -base64 32` (已是base64) |
| `gitlab-client-id` | OAuth 客户端 ID | 稍后配置,暂时保持示例值 |
| `gitlab-client-secret` | OAuth 客户端密钥 | 稍后配置,暂时保持示例值 |
| `gitlab-system-token` | GitLab 系统令牌 | 稍后配置,暂时保持示例值 |
| `third-party-api-key` | 第三方 API 密钥 | 根据需求设置,或使用随机值 |

**示例**: 修改 PostgreSQL 密码
```bash
# 1. 生成新密码
NEW_PASS=$(openssl rand -base64 20)
echo "生成的新密码: $NEW_PASS"

# 2. 转换为 base64
NEW_PASS_B64=$(echo -n "$NEW_PASS" | base64)
echo "Base64 编码: $NEW_PASS_B64"

# 3. 在 secrets.yaml 中找到 postgres-password 字段，替换为 $NEW_PASS_B64
```

#### 2.6 创建 Secret

```bash
# 应用 Secret 配置
kubectl apply -f secrets.yaml

# 期望输出:
# secret/gitlabex-secrets created
```

#### 2.7 验证 Secret

```bash
# 查看 Secret（不会显示具体值）
kubectl get secret gitlabex-secrets -n gitlabex

# 查看 Secret 包含的 key
kubectl describe secret gitlabex-secrets -n gitlabex

# 如果需要查看某个值（会以 base64 显示）
kubectl get secret gitlabex-secrets -n gitlabex -o jsonpath='{.data.postgres-password}'

# 解码查看实际值（调试用，生产环境不要这样做！）
kubectl get secret gitlabex-secrets -n gitlabex -o jsonpath='{.data.postgres-password}' | base64 -d
```

**安全提示**:
- ⚠️ 不要将 `secrets.yaml` 提交到 Git
- ⚠️ 定期更换密码（尤其是生产环境）
- ⚠️ 限制对 Secret 的访问权限

### 部署步骤 3: 部署 PostgreSQL 数据库

PostgreSQL 是 GitLabEx 和 GitLab 的主数据库。

#### 3.1 理解 postgres.yaml 的结构

```bash
# 查看配置文件
cat postgres.yaml
```

这个文件包含 4 个 Kubernetes 资源:
1. **PersistentVolumeClaim (PVC)**: 申请 20GB 存储空间用于数据持久化
2. **ConfigMap**: 存储数据库初始化脚本
3. **Deployment**: 定义如何运行 PostgreSQL Pod
4. **Service**: 提供稳定的网络访问入口

#### 3.2 修改存储类（如果需要）

```bash
# 检查集群的默认存储类
kubectl get storageclass

# 如果没有标记为 (default) 的存储类，需要在 postgres.yaml 中指定
# 编辑 postgres.yaml，在 PVC 部分取消注释并设置 storageClassName
```

找到这一部分:
```yaml
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
  # storageClassName: local-path  # 取消注释并修改为你的存储类名称
```

**如果有默认存储类**: 不需要修改
**如果没有默认存储类**: 取消注释并设置为实际的存储类名称

#### 3.3 部署 PostgreSQL

```bash
# 部署 PostgreSQL
kubectl apply -f postgres.yaml

# 期望输出:
# persistentvolumeclaim/postgres-pvc created
# configmap/postgres-init-script created
# deployment.apps/postgres created
# service/gitlabex-postgres created
```

**解释**:
- 创建了 4 个资源
- Kubernetes 会自动创建 PersistentVolume (PV) 来满足 PVC 的需求
- Deployment 会创建一个 Pod 运行 PostgreSQL

#### 3.4 查看 PostgreSQL 部署状态

```bash
# 查看 Pod 状态
kubectl get pods -n gitlabex -l app=postgres

# 期望输出示例:
# NAME                        READY   STATUS    RESTARTS   AGE
# postgres-5f7b8c9d6f-x7k2m   1/1     Running   0          2m

# 如果 STATUS 是 Pending 或 ContainerCreating，等待几分钟
# 如果是 Error 或 CrashLoopBackOff，查看详细信息:
kubectl describe pod -n gitlabex -l app=postgres
```

**Pod 状态说明**:
- `Pending`: 等待调度到节点
- `ContainerCreating`: 正在拉取镜像和创建容器
- `Running`: 正常运行
- `Error` / `CrashLoopBackOff`: 出现错误，需要查看日志

#### 3.5 查看 PostgreSQL 日志

```bash
# 查看实时日志
kubectl logs -f -n gitlabex -l app=postgres

# 应该看到类似输出:
# PostgreSQL init process complete; ready for start up.
# LOG:  database system is ready to accept connections

# 按 Ctrl+C 退出日志查看
```

#### 3.6 验证 PVC 状态

```bash
# 查看 PVC
kubectl get pvc -n gitlabex

# 期望输出:
# NAME           STATUS   VOLUME                  CAPACITY   ACCESS MODES   STORAGECLASS   AGE
# postgres-pvc   Bound    pvc-abc123...           20Gi       RWO            standard       5m

# STATUS 应该是 Bound（已绑定）
```

**如果 PVC 状态是 Pending**:
```bash
# 查看详细信息
kubectl describe pvc postgres-pvc -n gitlabex

# 常见问题:
# 1. 没有可用的 StorageClass
# 2. 存储空间不足
# 3. 存储类配置错误
```

#### 3.7 测试数据库连接

```bash
# 方法1: 进入 Pod 直接测试
kubectl exec -it -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- psql -U gitlab -d gitlab -c "SELECT version();"

# 应该看到 PostgreSQL 版本信息

# 方法2: 运行临时 Pod 测试连接
kubectl run -it --rm pg-test --image=postgres:15 --restart=Never -n gitlabex -- psql -h gitlabex-postgres -U gitlab -d gitlab -c "SELECT 1;"

# 需要输入密码（在 secrets.yaml 中设置的密码）
```

### 部署步骤 4: 部署 Redis 缓存

Redis 用于缓存和会话存储。

#### 4.1 部署 Redis

```bash
# 部署 Redis
kubectl apply -f redis.yaml

# 期望输出:
# persistentvolumeclaim/redis-pvc created
# deployment.apps/redis created
# service/gitlabex-redis created
```

#### 4.2 查看 Redis 状态

```bash
# 查看 Pod 状态
kubectl get pods -n gitlabex -l app=redis

# 查看日志
kubectl logs -f -n gitlabex -l app=redis

# 应该看到:
# Ready to accept connections
```

#### 4.3 测试 Redis 连接

```bash
# 进入 Redis Pod 测试
kubectl exec -it -n gitlabex $(kubectl get pod -n gitlabex -l app=redis -o jsonpath='{.items[0].metadata.name}') -- redis-cli -a $(kubectl get secret gitlabex-secrets -n gitlabex -o jsonpath='{.data.redis-password}' | base64 -d) ping

# 应该返回: PONG
```

### 部署步骤 5: 部署 MinIO 对象存储

MinIO 用于存储文件上传（文档、头像等）。

#### 5.1 部署 MinIO

```bash
# 部署 MinIO
kubectl apply -f minio.yaml

# 期望输出:
# persistentvolumeclaim/minio-pvc created
# deployment.apps/minio created
# service/gitlabex-minio created
# service/gitlabex-minio-console created
```

#### 5.2 查看 MinIO 状态

```bash
# 查看 Pod
kubectl get pods -n gitlabex -l app=minio

# 查看日志
kubectl logs -f -n gitlabex -l app=minio
```

#### 5.3 访问 MinIO 控制台

```bash
# 获取 MinIO 控制台访问地址
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
MINIO_PORT=$(kubectl get svc gitlabex-minio-console -n gitlabex -o jsonpath='{.spec.ports[0].nodePort}')

echo "MinIO Console: http://${NODE_IP}:${MINIO_PORT}"

# 用户名: admin
# 密码: 在 secrets.yaml 中设置的 minio-root-password（base64 解码后）
```

在浏览器打开地址，使用 `admin` 和你设置的密码登录。

**提示**: 如果无法访问，检查:
1. 防火墙规则
2. 节点 IP 是否正确
3. Pod 是否正常运行

### 部署步骤 6: 部署 GitLab

GitLab 是整个系统的核心，需要 5-10 分钟才能完全启动。

#### 6.1 修改 GitLab 配置（重要！）

在部署前，需要修改 `gitlab.yaml` 中的访问地址:

```bash
# 编辑 gitlab.yaml
vim gitlab.yaml

# 或
nano gitlab.yaml
```

找到 ConfigMap 部分，修改以下内容:
```yaml
data:
  GITLAB_OMNIBUS_CONFIG: |
    external_url 'http://YOUR_NODE_IP:30081'  # 修改为实际的节点 IP
    # ...
    gitlab_rails['gitlab_host'] = 'YOUR_NODE_IP'  # 修改为实际的节点 IP
```

**如何获取节点 IP**:
```bash
# 获取第一个节点的 IP
kubectl get nodes -o wide

# 或使用命令直接获取
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
echo "节点 IP: $NODE_IP"
```

将配置文件中的 `YOUR_NODE_IP` 替换为实际的节点 IP。

**示例**:
```yaml
external_url 'http://192.168.1.100:30081'
gitlab_rails['gitlab_host'] = '192.168.1.100'
```

#### 6.2 部署 GitLab

```bash
# 部署 GitLab
kubectl apply -f gitlab.yaml

# 期望输出:
# persistentvolumeclaim/gitlab-config-pvc created
# persistentvolumeclaim/gitlab-logs-pvc created
# persistentvolumeclaim/gitlab-data-pvc created
# persistentvolumeclaim/gitlab-oauth-pvc created
# configmap/gitlab-config created
# deployment.apps/gitlab created
# service/gitlabex-gitlab created
# service/gitlabex-gitlab-ssh created
```

**解释**:
- 创建了 4 个 PVC (总共 111 GB)
- 创建了 1 个 ConfigMap
- 创建了 1 个 Deployment
- 创建了 2 个 Service (HTTP 和 SSH)

#### 6.3 监控 GitLab 启动过程

GitLab 启动需要 5-10 分钟，要有耐心！

```bash
# 查看 Pod 状态（会持续更新）
kubectl get pods -n gitlabex -l app=gitlab -w

# 按 Ctrl+C 停止监控
```

你会看到类似的状态变化:
```
NAME                      READY   STATUS              RESTARTS   AGE
gitlab-7b8c9d5f4-x9k2m    0/1     ContainerCreating   0          30s
gitlab-7b8c9d5f4-x9k2m    0/1     Running             0          2m
gitlab-7b8c9d5f4-x9k2m    1/1     Running             0          8m
```

**状态说明**:
- `ContainerCreating`: 拉取镜像（GitLab 镜像很大，约 2.5GB）
- `Running` 但 `READY 0/1`: 容器启动中，正在初始化
- `Running` 且 `READY 1/1`: 启动完成，健康检查通过 ✅

#### 6.4 查看 GitLab 启动日志

```bash
# 查看实时日志
kubectl logs -f -n gitlabex -l app=gitlab

# 你会看到大量的初始化输出
# 当看到类似 "gitlab Reconfigured!" 时表示配置完成
```

#### 6.5 验证 GitLab 服务

```bash
# 查看 Service
kubectl get svc -n gitlabex | grep gitlab

# 期望输出:
# gitlabex-gitlab      NodePort   10.96.123.45   <none>  80:30081/TCP   10m
# gitlabex-gitlab-ssh  NodePort   10.96.123.46   <none>  22:30222/TCP   10m
```

#### 6.6 访问 GitLab Web 界面

```bash
# 获取访问地址
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
echo "GitLab URL: http://${NODE_IP}:30081"
```

在浏览器中打开这个地址：

1. **首次访问**: 可能需要等待 1-2 分钟才能看到页面（GitLab 还在初始化）
2. **登录页面**: 看到登录页面表示 GitLab 已启动成功 🎉
3. **root 用户**: 
   - 用户名: `root`
   - 密码: 在 `secrets.yaml` 中设置的 `gitlab-root-password`（需要 base64 解码）

```bash
# 查看 root 密码
kubectl get secret gitlabex-secrets -n gitlabex -o jsonpath='{.data.gitlab-root-password}' | base64 -d
echo ""
```

### 部署步骤 7: 配置 GitLab OAuth

在部署后端服务之前，我们需要在 GitLab 中创建 OAuth 应用。

#### 7.1 创建 OAuth 应用

1. **登录 GitLab**: 使用 root 用户登录
2. **进入 Admin Area**: 点击左上角的扳手图标（Admin Area）
3. **进入 Applications**: 左侧菜单选择 `Applications`
4. **创建新应用**: 点击 `New application` 按钮

填写以下信息:
```
Name: GitLabEx
Redirect URI: http://NODE_IP:30000/auth/callback  # 替换 NODE_IP 为实际节点 IP
Confidential: ✓ (勾选)
Scopes: 
  ✓ api
  ✓ read_user
  ✓ read_repository
```

**示例 Redirect URI**: `http://192.168.1.100:30000/auth/callback`

5. **保存**: 点击 `Save application`

#### 7.2 获取 OAuth 凭据

保存后会显示:
- **Application ID**: 类似 `abc123def456...`
- **Secret**: 类似 `xyz789uvw012...`

**重要**: 复制这两个值，稍后需要使用！

#### 7.3 创建 System Token

System Token 用于后端服务调用 GitLab API。

**方法 1: 通过 GitLab Web 界面**
1. 登录 GitLab (root 用户)
2. 点击右上角头像 → `Edit profile`
3. 左侧菜单选择 `Access Tokens`
4. 填写信息:
   ```
   Token name: GitLabEx System Token
   Expiration date: 1 year from now
   Scopes: 
     ✓ api
     ✓ read_user
     ✓ read_repository
   ```
5. 点击 `Create personal access token`
6. **复制生成的 Token**: 类似 `glpat-xxxxxxxxxxxxxxxxxxxx`

**方法 2: 通过命令行（高级）**
```bash
# 进入 GitLab Pod
kubectl exec -it -n gitlabex $(kubectl get pod -n gitlabex -l app=gitlab -o jsonpath='{.items[0].metadata.name}') -- bash

# 在 Pod 中运行 Rails console
gitlab-rails console

# 在 console 中执行（复制整段代码）
token = User.find_by_username('root').personal_access_tokens.create(
  scopes: [:api, :read_user, :read_repository], 
  name: 'GitLabEx System Token',
  expires_at: 365.days.from_now
)
token.set_token('glpat-' + SecureRandom.hex(20))
token.save!
puts token.token

# 复制输出的 token，然后退出
exit
exit
```

#### 7.4 更新 Secret

现在我们有了 3 个值:
1. Application ID (OAuth Client ID)
2. Application Secret (OAuth Client Secret)
3. Personal Access Token (System Token)

将它们更新到 Secret 中:

```bash
# 1. 将值转换为 base64
CLIENT_ID="粘贴你的 Application ID"
CLIENT_SECRET="粘贴你的 Application Secret"
SYSTEM_TOKEN="粘贴你的 System Token"

CLIENT_ID_B64=$(echo -n "$CLIENT_ID" | base64)
CLIENT_SECRET_B64=$(echo -n "$CLIENT_SECRET" | base64)
SYSTEM_TOKEN_B64=$(echo -n "$SYSTEM_TOKEN" | base64)

# 2. 打印 base64 值（复制这些值）
echo "Client ID (base64): $CLIENT_ID_B64"
echo "Client Secret (base64): $CLIENT_SECRET_B64"
echo "System Token (base64): $SYSTEM_TOKEN_B64"

# 3. 编辑 secrets.yaml，替换对应的值
vim secrets.yaml
```

在 `secrets.yaml` 中找到并替换:
```yaml
data:
  gitlab-client-id: <粘贴 CLIENT_ID_B64>
  gitlab-client-secret: <粘贴 CLIENT_SECRET_B64>
  gitlab-system-token: <粘贴 SYSTEM_TOKEN_B64>
```

#### 7.5 重新应用 Secret

```bash
# 更新 Secret
kubectl apply -f secrets.yaml

# 输出: secret/gitlabex-secrets configured
```

### 部署步骤 8: 部署后端服务

现在我们可以部署后端 API 服务了。

#### 8.1 部署后端

```bash
# 部署后端
kubectl apply -f backend.yaml

# 期望输出:
# persistentvolumeclaim/backend-uploads-pvc created
# persistentvolumeclaim/backend-logs-pvc created
# configmap/backend-config created
# deployment.apps/backend created
# service/gitlabex-backend created
```

#### 8.2 查看后端状态

```bash
# 查看 Pod（应该有 2 个副本）
kubectl get pods -n gitlabex -l app=backend

# 期望输出:
# NAME                       READY   STATUS    RESTARTS   AGE
# backend-5f7b8c9d6f-abc12   1/1     Running   0          2m
# backend-5f7b8c9d6f-def34   1/1     Running   0          2m

# 查看日志
kubectl logs -f -n gitlabex -l app=backend --tail=50
```

#### 8.3 测试后端 API

```bash
# 获取后端 API 地址
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
echo "Backend API: http://${NODE_IP}:30080"

# 测试健康检查接口
curl http://${NODE_IP}:30080/health

# 应该返回: {"status":"ok"} 或类似的成功响应
```

如果返回错误，查看 Pod 日志排查问题。

### 部署步骤 9: 部署前端服务

最后部署前端 Web 应用。

#### 9.1 修改前端配置

编辑 `frontend.yaml`，修改环境变量:

```bash
vim frontend.yaml
```

找到 `env` 部分:
```yaml
env:
- name: BACKEND_URL
  value: "http://YOUR_NODE_IP:30080"  # 修改为实际节点 IP
- name: VITE_GITLAB_URL
  value: "http://YOUR_NODE_IP:30081"  # 修改为实际节点 IP
```

替换为实际的节点 IP。

#### 9.2 部署前端

```bash
# 部署前端
kubectl apply -f frontend.yaml

# 期望输出:
# deployment.apps/frontend created
# service/gitlabex-frontend created
```

#### 9.3 查看前端状态

```bash
# 查看 Pod（应该有 2 个副本）
kubectl get pods -n gitlabex -l app=frontend

# 查看日志
kubectl logs -f -n gitlabex -l app=frontend
```

#### 9.4 访问前端应用

```bash
# 获取前端地址
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
echo "Frontend URL: http://${NODE_IP}:30000"
```

在浏览器中打开这个地址，你应该看到 GitLabEx 的登录页面！🎉

### 部署步骤 10: 验证完整系统

#### 10.1 查看所有资源

```bash
# 查看所有 Pod
kubectl get pods -n gitlabex

# 期望输出: 所有 Pod 的 STATUS 都是 Running，READY 显示为 1/1 或 2/2
# NAME                        READY   STATUS    RESTARTS   AGE
# postgres-xxx                1/1     Running   0          20m
# redis-xxx                   1/1     Running   0          18m
# minio-xxx                   1/1     Running   0          16m
# gitlab-xxx                  1/1     Running   0          14m
# backend-xxx                 1/1     Running   0          5m
# backend-yyy                 1/1     Running   0          5m
# frontend-xxx                1/1     Running   0          3m
# frontend-yyy                1/1     Running   0          3m

# 查看所有 Service
kubectl get svc -n gitlabex

# 查看所有 PVC（所有 STATUS 应该是 Bound）
kubectl get pvc -n gitlabex
```

#### 10.2 测试登录流程

1. 打开前端地址: `http://NODE_IP:30000`
2. 点击 "使用 GitLab 登录" 按钮
3. 跳转到 GitLab OAuth 授权页面
4. 使用 root 用户登录
5. 授权应用访问
6. 成功登录到 GitLabEx 系统 ✅

#### 10.3 服务访问地址汇总

```bash
# 一键获取所有访问地址
NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')

echo "=========================================="
echo "GitLabEx 服务访问地址"
echo "=========================================="
echo "前端应用:      http://${NODE_IP}:30000"
echo "后端 API:      http://${NODE_IP}:30080"
echo "后端 Metrics:  http://${NODE_IP}:30090"
echo "GitLab Web:    http://${NODE_IP}:30081"
echo "GitLab SSH:    ssh://git@${NODE_IP}:30222"
echo "MinIO 控制台:  http://${NODE_IP}:30901"
echo "=========================================="
```

恭喜！🎉 你已经成功在 Kubernetes 上部署了 GitLabEx！

---

## 📖 日常管理操作

部署完成后，这里是一些常用的管理操作。

### 查看系统状态

#### 快速查看所有资源

```bash
# 查看所有 Pod
kubectl get pods -n gitlabex -o wide

# 输出说明:
# NAME: Pod 名称
# READY: 就绪的容器数/总容器数
# STATUS: 运行状态 (Running, Pending, Error 等)
# RESTARTS: 重启次数
# AGE: 运行时长
# IP: Pod 的 IP 地址
# NODE: 运行在哪个节点上

# 查看所有 Service
kubectl get svc -n gitlabex

# 查看所有 PVC
kubectl get pvc -n gitlabex

# 一次查看所有资源
kubectl get all -n gitlabex
```

#### 查看资源使用情况

```bash
# 查看节点资源使用
kubectl top nodes

# 查看 Pod 资源使用
kubectl top pods -n gitlabex

# 如果提示 Metrics API not available，需要安装 metrics-server
```

#### 查看详细信息

```bash
# 查看 Pod 详细信息（包括事件、配置等）
kubectl describe pod <pod-name> -n gitlabex

# 示例: 查看 backend Pod
kubectl describe pod $(kubectl get pod -n gitlabex -l app=backend -o jsonpath='{.items[0].metadata.name}') -n gitlabex

# 查看 Service 详细信息
kubectl describe svc gitlabex-backend -n gitlabex
```

### 查看日志

日志是排查问题的关键工具。

#### 基本日志查看

```bash
# 查看 Pod 日志（实时滚动）
kubectl logs -f <pod-name> -n gitlabex

# 示例: 查看后端日志
kubectl logs -f -n gitlabex -l app=backend

# 查看最近 100 行日志
kubectl logs --tail=100 <pod-name> -n gitlabex

# 查看过去 1 小时的日志
kubectl logs --since=1h <pod-name> -n gitlabex

# 查看上一个崩溃容器的日志
kubectl logs --previous <pod-name> -n gitlabex
```

#### 多 Pod 日志查看

当一个应用有多个 Pod 副本时:

```bash
# 查看所有 backend Pod 的日志
kubectl logs -f -n gitlabex -l app=backend --all-containers=true

# 查看特定 Pod 的日志
kubectl logs -f -n gitlabex backend-5f7b8c9d6f-abc12
```

#### 保存日志到文件

```bash
# 保存日志到文件
kubectl logs <pod-name> -n gitlabex > pod-log.txt

# 保存所有 Pod 日志
for pod in $(kubectl get pods -n gitlabex -o name); do
  echo "=== $pod ===" >> all-logs.txt
  kubectl logs -n gitlabex $pod >> all-logs.txt 2>&1
done
```

### 扩缩容操作

根据负载动态调整服务实例数量。

#### 手动扩缩容

```bash
# 扩展后端服务到 3 个副本
kubectl scale deployment backend -n gitlabex --replicas=3

# 查看扩容结果
kubectl get pods -n gitlabex -l app=backend

# 缩减到 1 个副本
kubectl scale deployment backend -n gitlabex --replicas=1

# 扩展前端服务
kubectl scale deployment frontend -n gitlabex --replicas=4
```

#### 自动扩缩容（HPA）

```bash
# 创建 HPA，基于 CPU 使用率自动扩缩容
kubectl autoscale deployment backend -n gitlabex --cpu-percent=70 --min=2 --max=10

# 查看 HPA 状态
kubectl get hpa -n gitlabex

# 查看 HPA 详细信息
kubectl describe hpa backend -n gitlabex

# 删除 HPA
kubectl delete hpa backend -n gitlabex
```

**HPA 说明**:
- `--cpu-percent=70`: 当 CPU 使用率超过 70% 时扩容
- `--min=2`: 最少保持 2 个副本
- `--max=10`: 最多扩展到 10 个副本

### 更新应用

#### 更新镜像版本

```bash
# 构建新版本镜像
cd /path/to/gitlabex2
./scripts/build-images.sh
# 给新镜像打标签
docker tag gitlabex-backend:latest gitlabex-backend:v2.0

# 更新 Deployment 使用新镜像
kubectl set image deployment/backend backend=gitlabex-backend:v2.0 -n gitlabex

# 查看滚动更新状态
kubectl rollout status deployment/backend -n gitlabex

# 输出: deployment "backend" successfully rolled out
```

#### 查看更新历史

```bash
# 查看部署历史
kubectl rollout history deployment/backend -n gitlabex

# 输出示例:
# REVISION  CHANGE-CAUSE
# 1         <none>
# 2         kubectl set image deployment/backend backend=gitlabex-backend:v2.0
```

#### 回滚到上一个版本

```bash
# 回滚到上一个版本
kubectl rollout undo deployment/backend -n gitlabex

# 回滚到指定版本
kubectl rollout undo deployment/backend -n gitlabex --to-revision=1

# 查看回滚状态
kubectl rollout status deployment/backend -n gitlabex
```

### 重启服务

有时需要重启服务以应用新配置或解决问题。

```bash
# 滚动重启后端服务（不会中断服务）
kubectl rollout restart deployment/backend -n gitlabex

# 滚动重启前端服务
kubectl rollout restart deployment/frontend -n gitlabex

# 重启 GitLab（会有短暂中断）
kubectl rollout restart deployment/gitlab -n gitlabex

# 强制删除并重建 Pod（不推荐，会中断服务）
kubectl delete pod <pod-name> -n gitlabex
```

**滚动重启过程**:
1. 创建新的 Pod
2. 等待新 Pod 就绪
3. 删除旧的 Pod
4. 重复直到所有 Pod 都更新

### 进入容器调试

#### 进入 Pod 执行命令

```bash
# 进入 Pod 的 shell（如果容器有 bash）
kubectl exec -it <pod-name> -n gitlabex -- /bin/bash

# 如果没有 bash，使用 sh
kubectl exec -it <pod-name> -n gitlabex -- /bin/sh

# 在 Pod 中执行单个命令
kubectl exec <pod-name> -n gitlabex -- ls -la /app

# 示例: 进入 PostgreSQL Pod
kubectl exec -it $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -n gitlabex -- bash

# 在 PostgreSQL Pod 中连接数据库
psql -U gitlab -d gitlab
```

#### 在 Pod 中执行数据库操作

```bash
# PostgreSQL 操作示例
kubectl exec -it $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -n gitlabex -- psql -U gitlab -d gitlabex -c "SELECT COUNT(*) FROM users;"

# Redis 操作示例
kubectl exec -it $(kubectl get pod -n gitlabex -l app=redis -o jsonpath='{.items[0].metadata.name}') -n gitlabex -- redis-cli -a $(kubectl get secret gitlabex-secrets -n gitlabex -o jsonpath='{.data.redis-password}' | base64 -d) KEYS '*'
```

### 更新配置

#### 更新 Secret

```bash
# 1. 修改 secrets.yaml 文件
vim secrets.yaml

# 2. 应用更新
kubectl apply -f secrets.yaml

# 输出: secret/gitlabex-secrets configured

# 3. 重启依赖这个 Secret 的服务
kubectl rollout restart deployment/backend -n gitlabex
kubectl rollout restart deployment/gitlab -n gitlabex
```

#### 更新 ConfigMap

```bash
# 方法1: 修改 YAML 文件后重新应用
vim backend.yaml
kubectl apply -f backend.yaml
kubectl rollout restart deployment/backend -n gitlabex

# 方法2: 直接编辑 ConfigMap
kubectl edit configmap backend-config -n gitlabex
kubectl rollout restart deployment/backend -n gitlabex
```

### 复制文件到/从 Pod

```bash
# 从 Pod 复制文件到本地
kubectl cp gitlabex/<pod-name>:/path/to/file ./local-file

# 从本地复制文件到 Pod
kubectl cp ./local-file gitlabex/<pod-name>:/path/to/file

# 示例: 复制后端日志到本地
kubectl cp gitlabex/$(kubectl get pod -n gitlabex -l app=backend -o jsonpath='{.items[0].metadata.name}'):/app/logs/app.log ./backend-app.log
```

---

## 🗑️ 卸载应用

如果需要删除 GitLabEx，请按照以下步骤操作。

### 方法 1: 完全删除（包括数据）

⚠️ **警告**: 这将删除所有数据，包括数据库、文件等，且**不可恢复**！

```bash
# 按相反顺序删除应用
kubectl delete -f frontend.yaml
kubectl delete -f backend.yaml
kubectl delete -f gitlab.yaml
kubectl delete -f minio.yaml
kubectl delete -f redis.yaml
kubectl delete -f postgres.yaml

# 删除 Secret
kubectl delete -f secrets.yaml

# 删除命名空间（会删除命名空间中的所有资源，包括 PVC）
kubectl delete -f namespace.yaml
```

**或者使用一条命令**:
```bash
# 删除整个命名空间（所有资源都会被删除）
kubectl delete namespace gitlabex

# 等待删除完成（可能需要几分钟）
```

### 方法 2: 保留数据的删除

如果将来还要恢复，可以只删除应用，保留 PVC（数据）。

```bash
# 只删除 Deployment 和 Service，保留 PVC
kubectl delete deployment --all -n gitlabex
kubectl delete service --all -n gitlabex
kubectl delete configmap --all -n gitlabex
kubectl delete secret --all -n gitlabex

# 查看保留的 PVC
kubectl get pvc -n gitlabex

# 数据仍然存在，可以重新部署应用来恢复
```

### 验证卸载

```bash
# 检查命名空间是否删除
kubectl get namespace gitlabex

# 如果命名空间还在删除中（Terminating 状态），等待完成
# 检查是否还有残留资源
kubectl api-resources --verbs=list --namespaced -o name | xargs -n 1 kubectl get --show-kind --ignore-not-found -n gitlabex
```

---

## 💾 备份和恢复

定期备份数据是非常重要的。

### 备份策略

建议定期备份:
- **数据库**: 每天备份
- **配置文件**: 每次修改后备份
- **文件存储**: 每周备份

### 备份数据库

#### 备份 PostgreSQL

```bash
# 创建备份目录
mkdir -p backups

# 备份 gitlab 数据库
kubectl exec -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  pg_dump -U gitlab gitlab > backups/gitlab-$(date +%Y%m%d-%H%M%S).sql

# 备份 gitlabex 数据库
kubectl exec -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  pg_dump -U gitlab gitlabex > backups/gitlabex-$(date +%Y%m%d-%H%M%S).sql

# 查看备份文件
ls -lh backups/
```

#### 压缩备份文件

```bash
# 压缩备份文件以节省空间
gzip backups/gitlab-*.sql
gzip backups/gitlabex-*.sql

# 或者直接生成压缩的备份
kubectl exec -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  pg_dump -U gitlab gitlab | gzip > backups/gitlab-$(date +%Y%m%d).sql.gz
```

### 备份配置文件

```bash
# 备份所有 Kubernetes 配置
kubectl get all,pvc,secrets,configmaps -n gitlabex -o yaml > backups/k8s-config-$(date +%Y%m%d).yaml

# 备份 Secret（加密存储）
kubectl get secret gitlabex-secrets -n gitlabex -o yaml > backups/secrets-$(date +%Y%m%d).yaml

# 重要: 妥善保管 secrets 备份文件！
```

### 备份持久卷数据

#### 方法 1: 使用卷快照（推荐）

如果你的存储类支持快照功能:

```bash
# 创建 VolumeSnapshot（需要 VolumeSnapshot CRD）
cat <<EOF | kubectl apply -f -
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: postgres-snapshot-$(date +%Y%m%d)
  namespace: gitlabex
spec:
  source:
    persistentVolumeClaimName: postgres-pvc
EOF

# 查看快照
kubectl get volumesnapshot -n gitlabex
```

#### 方法 2: 手动复制数据

```bash
# 创建临时 Pod 挂载 PVC
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: backup-pod
  namespace: gitlabex
spec:
  containers:
  - name: backup
    image: busybox
    command: ["sleep", "3600"]
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: postgres-pvc
EOF

# 等待 Pod 启动
kubectl wait --for=condition=ready pod/backup-pod -n gitlabex --timeout=120s

# 打包数据
kubectl exec -n gitlabex backup-pod -- tar czf /tmp/postgres-data.tar.gz -C /data .

# 复制到本地
kubectl cp gitlabex/backup-pod:/tmp/postgres-data.tar.gz backups/postgres-data-$(date +%Y%m%d).tar.gz

# 清理临时 Pod
kubectl delete pod backup-pod -n gitlabex
```

### 恢复数据

#### 恢复数据库

```bash
# 恢复 gitlab 数据库
kubectl exec -i -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  psql -U gitlab gitlab < backups/gitlab-20241217.sql

# 恢复 gitlabex 数据库
kubectl exec -i -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  psql -U gitlab gitlabex < backups/gitlabex-20241217.sql

# 如果是压缩文件
gunzip -c backups/gitlab-20241217.sql.gz | \
  kubectl exec -i -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  psql -U gitlab gitlab
```

#### 恢复 PVC 数据

```bash
# 1. 创建临时 Pod
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: restore-pod
  namespace: gitlabex
spec:
  containers:
  - name: restore
    image: busybox
    command: ["sleep", "3600"]
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: postgres-pvc
EOF

# 2. 复制备份文件到 Pod
kubectl cp backups/postgres-data-20241217.tar.gz gitlabex/restore-pod:/tmp/postgres-data.tar.gz

# 3. 解压恢复
kubectl exec -n gitlabex restore-pod -- sh -c "cd /data && rm -rf * && tar xzf /tmp/postgres-data.tar.gz"

# 4. 清理
kubectl delete pod restore-pod -n gitlabex
```

### 自动化备份脚本

创建定时备份任务:

```bash
# 创建备份脚本
cat > backup-gitlabex.sh <<'EOF'
#!/bin/bash
BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p $BACKUP_DIR

# 备份数据库
kubectl exec -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  pg_dump -U gitlab gitlab | gzip > $BACKUP_DIR/gitlab-$DATE.sql.gz

kubectl exec -n gitlabex $(kubectl get pod -n gitlabex -l app=postgres -o jsonpath='{.items[0].metadata.name}') -- \
  pg_dump -U gitlab gitlabex | gzip > $BACKUP_DIR/gitlabex-$DATE.sql.gz

# 备份配置
kubectl get all,pvc,secrets,configmaps -n gitlabex -o yaml | gzip > $BACKUP_DIR/k8s-config-$DATE.yaml.gz

# 删除 30 天前的备份
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete

echo "Backup completed: $DATE"
EOF

chmod +x backup-gitlabex.sh

# 测试备份脚本
./backup-gitlabex.sh
```

使用 crontab 定时执行:

```bash
# 编辑 crontab
crontab -e

# 添加每天凌晨 2 点执行备份
0 2 * * * /path/to/backup-gitlabex.sh >> /var/log/gitlabex-backup.log 2>&1
```

---

## 🔧 故障排查指南

遇到问题时，按照这个系统化的流程排查。

### 排查步骤总览

1. **查看 Pod 状态** - 确定哪个组件有问题
2. **查看 Pod 日志** - 了解错误信息
3. **查看 Pod 详情** - 查看事件和配置
4. **检查依赖服务** - 确认依赖的服务是否正常
5. **测试网络连接** - 验证服务间的网络通信

### 常见问题 1: Pod 处于 Pending 状态

**现象**: Pod 一直处于 Pending 状态，无法启动

```bash
# 查看 Pod 状态
kubectl get pods -n gitlabex

# NAME                        READY   STATUS    RESTARTS   AGE
# postgres-5f7b8c9d6f-x7k2m   0/1     Pending   0          5m
```

**排查步骤**:

```bash
# 1. 查看详细信息
kubectl describe pod <pod-name> -n gitlabex

# 重点关注 Events 部分的错误信息
```

**常见原因和解决方法**:

| 错误信息 | 原因 | 解决方法 |
|---------|------|---------|
| `Insufficient cpu/memory` | 节点资源不足 | 清理资源或增加节点 |
| `PersistentVolumeClaim is not bound` | PVC 未绑定 | 检查 StorageClass 配置 |
| `0/3 nodes are available: pod has unbound immediate PersistentVolumeClaims` | 存储类问题 | 查看下方 PVC 排查 |
| `No nodes available` | 无可用节点 | 检查节点状态 |

```bash
# 检查节点资源
kubectl top nodes
kubectl describe nodes

# 检查 PVC
kubectl get pvc -n gitlabex
kubectl describe pvc <pvc-name> -n gitlabex
```

### 常见问题 2: Pod 处于 CrashLoopBackOff 状态

**现象**: Pod 不断重启

```bash
# NAME                        READY   STATUS             RESTARTS   AGE
# backend-5f7b8c9d6f-abc12    0/1     CrashLoopBackOff   5          10m
```

**排查步骤**:

```bash
# 1. 查看当前日志
kubectl logs <pod-name> -n gitlabex

# 2. 查看上一次崩溃的日志（更重要！）
kubectl logs <pod-name> -n gitlabex --previous

# 3. 查看 Pod 详情
kubectl describe pod <pod-name> -n gitlabex
```

**常见原因**:

1. **配置错误** - 数据库连接信息错误
2. **依赖服务未就绪** - 数据库还未启动
3. **镜像问题** - 镜像不存在或启动命令错误
4. **资源限制** - 内存不足被 OOM Kill

**解决示例**:

```bash
# 如果是数据库连接问题，检查 Secret
kubectl get secret gitlabex-secrets -n gitlabex -o jsonpath='{.data.postgres-password}' | base64 -d

# 如果是依赖服务问题，确认数据库已就绪
kubectl get pods -n gitlabex -l app=postgres

# 如果是资源问题，检查资源使用
kubectl top pod <pod-name> -n gitlabex
```

### 常见问题 3: PVC 处于 Pending 状态

**现象**: PersistentVolumeClaim 无法绑定

```bash
kubectl get pvc -n gitlabex

# NAME           STATUS    VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
# postgres-pvc   Pending                                      standard       5m
```

**排查步骤**:

```bash
# 查看 PVC 详情
kubectl describe pvc postgres-pvc -n gitlabex

# 重点查看 Events 部分
```

**常见原因和解决方法**:

**原因 1**: 没有 StorageClass
```bash
# 检查 StorageClass
kubectl get storageclass

# 如果没有输出，需要安装存储插件
# Minikube 用户:
minikube addons enable storage-provisioner

# Kind 用户: Kind 自带 local-path-provisioner

# 其他环境: 安装 local-path-provisioner
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
```

**原因 2**: StorageClass 名称不匹配
```bash
# 查看默认 StorageClass
kubectl get storageclass

# 修改 PVC 配置使用正确的 StorageClass
# 或设置默认 StorageClass:
kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

**原因 3**: 存储空间不足
```bash
# 检查节点存储空间
kubectl get nodes -o custom-columns=NAME:.metadata.name,DISK:.status.allocatable.ephemeral-storage

# 如果空间不足，清理或扩展存储
```

### 常见问题 4: GitLab 启动失败或很慢

**现象**: GitLab Pod 启动超过 10 分钟或失败

**排查步骤**:

```bash
# 1. 查看 GitLab Pod 状态
kubectl get pods -n gitlabex -l app=gitlab

# 2. 查看实时日志
kubectl logs -f -n gitlabex -l app=gitlab

# 3. 查看资源使用
kubectl top pod -n gitlabex -l app=gitlab

# 4. 查看 Pod 详情
kubectl describe pod -n gitlabex -l app=gitlab
```

**常见问题**:

**问题 1**: 内存不足
```bash
# 查看节点内存
kubectl top nodes

# 如果内存不足，需要:
# - 增加节点内存
# - 或减少 GitLab 的 resource requests

# 临时解决: 编辑 Deployment 减少资源请求
kubectl edit deployment gitlab -n gitlabex
# 修改 resources.requests.memory 为更小的值 (如 2Gi)
```

**问题 2**: 数据库连接失败
```bash
# 进入 GitLab Pod 检查
kubectl exec -it $(kubectl get pod -n gitlabex -l app=gitlab -o jsonpath='{.items[0].metadata.name}') -n gitlabex -- bash

# 在 Pod 中测试数据库连接
nc -zv gitlabex-postgres 5432

# 如果无法连接，检查 PostgreSQL Service
kubectl get svc gitlabex-postgres -n gitlabex

# 检查 PostgreSQL Pod
kubectl get pods -n gitlabex -l app=postgres
```

**问题 3**: 配置错误
```bash
# 查看 GitLab ConfigMap
kubectl get configmap gitlab-config -n gitlabex -o yaml

# 检查 external_url 和数据库配置是否正确
```

### 常见问题 5: 无法访问服务（NodePort）

**现象**: 在浏览器中无法打开 `http://NODE_IP:30000`

**排查步骤**:

```bash
# 1. 确认 Service 存在
kubectl get svc -n gitlabex

# 2. 确认 NodePort 端口
kubectl get svc gitlabex-frontend -n gitlabex -o jsonpath='{.spec.ports[0].nodePort}'

# 3. 确认 Pod 正常运行
kubectl get pods -n gitlabex -l app=frontend

# 4. 测试 Service 内部访问
kubectl run -it --rm debug --image=busybox --restart=Never -n gitlabex -- \
  wget -O- http://gitlabex-frontend

# 5. 检查节点 IP
kubectl get nodes -o wide
```

**常见原因**:

**原因 1**: 防火墙阻止
```bash
# 检查防火墙规则
# Ubuntu/Debian:
sudo ufw status
sudo ufw allow 30000:32767/tcp

# CentOS/RHEL:
sudo firewall-cmd --list-all
sudo firewall-cmd --add-port=30000-32767/tcp --permanent
sudo firewall-cmd --reload
```

**原因 2**: 使用了错误的节点 IP
```bash
# 获取正确的节点 IP
kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}'
kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}'

# 如果是云环境，可能需要使用 ExternalIP
# 如果是内网环境，使用 InternalIP
```

**原因 3**: Pod 未就绪
```bash
# 检查 Pod 的就绪状态
kubectl get pods -n gitlabex -l app=frontend

# 如果 READY 是 0/1，查看日志
kubectl logs -n gitlabex -l app=frontend
```

### 常见问题 6: 后端无法连接数据库

**现象**: 后端日志显示数据库连接错误

**排查步骤**:

```bash
# 1. 查看后端日志
kubectl logs -n gitlabex -l app=backend | grep -i "database\|postgres\|connection"

# 2. 测试从后端 Pod 连接数据库
kubectl exec -it $(kubectl get pod -n gitlabex -l app=backend -o jsonpath='{.items[0].metadata.name}') -n gitlabex -- sh

# 在 Pod 中:
# 测试 DNS 解析
nslookup gitlabex-postgres

# 测试端口连通性
nc -zv gitlabex-postgres 5432

# 3. 检查 Secret 中的密码
kubectl get secret gitlabex-secrets -n gitlabex -o jsonpath='{.data.postgres-password}' | base64 -d
```

**解决方法**:

```bash
# 如果密码错误，更新 Secret
vim secrets.yaml
kubectl apply -f secrets.yaml
kubectl rollout restart deployment/backend -n gitlabex

# 如果 DNS 解析失败，检查 Service
kubectl get svc gitlabex-postgres -n gitlabex

# 如果 PostgreSQL 未启动，检查 PostgreSQL Pod
kubectl get pods -n gitlabex -l app=postgres
kubectl logs -n gitlabex -l app=postgres
```

### 通用调试技巧

#### 技巧 1: 使用临时 Debug Pod

```bash
# 创建一个 debug Pod 用于网络测试
kubectl run -it --rm debug --image=nicolaka/netshoot --restart=Never -n gitlabex -- bash

# 在 debug Pod 中可以使用各种网络工具:
# - ping, curl, wget
# - nslookup, dig
# - telnet, nc (netcat)
# - traceroute, tcpdump

# 示例: 测试服务连接
curl http://gitlabex-backend:8080/health
nslookup gitlabex-postgres
nc -zv gitlabex-redis 6379
```

#### 技巧 2: 查看系统事件

```bash
# 查看命名空间的所有事件
kubectl get events -n gitlabex --sort-by='.lastTimestamp'

# 只查看 Warning 和 Error 事件
kubectl get events -n gitlabex --field-selector type!=Normal

# 实时监控事件
kubectl get events -n gitlabex --watch
```

#### 技巧 3: 端口转发调试

```bash
# 将远程 Pod 端口转发到本地
kubectl port-forward -n gitlabex <pod-name> 8080:8080

# 现在可以在本地访问: http://localhost:8080

# 示例: 调试后端服务
kubectl port-forward -n gitlabex $(kubectl get pod -n gitlabex -l app=backend -o jsonpath='{.items[0].metadata.name}') 8080:8080

# 在另一个终端测试
curl http://localhost:8080/health
```

#### 技巧 4: 查看资源使用情况

```bash
# 查看 Pod 资源使用
kubectl top pods -n gitlabex

# 查看特定 Pod 的资源限制
kubectl describe pod <pod-name> -n gitlabex | grep -A 5 "Limits\|Requests"

# 如果资源使用接近限制，考虑调整 resources 配置
```

### 获取帮助

如果以上方法都无法解决问题:

1. **收集诊断信息**:
```bash
# 导出所有资源状态
kubectl get all -n gitlabex -o yaml > debug-all-resources.yaml

# 导出所有事件
kubectl get events -n gitlabex --sort-by='.lastTimestamp' > debug-events.txt

# 导出所有 Pod 日志
for pod in $(kubectl get pods -n gitlabex -o name); do
  echo "=== $pod ===" >> debug-logs.txt
  kubectl logs -n gitlabex $pod >> debug-logs.txt 2>&1
done
```

2. **查看项目文档**: [项目 README](../README.md)
3. **搜索类似问题**: GitHub Issues
4. **提交问题**: 提供诊断信息和详细的错误描述

---

## 📝 常见问题 FAQ

### Q1: 如何修改 NodePort 端口？

**A**: 编辑对应 Service 的 YAML 文件，修改 `nodePort` 字段:

```bash
# 编辑 Service
vim frontend.yaml

# 找到 nodePort 字段并修改
spec:
  ports:
  - port: 80
    targetPort: 80
    nodePort: 30100  # 修改为你想要的端口 (30000-32767)

# 重新应用
kubectl apply -f frontend.yaml
```

### Q2: 如何使用域名访问而不是 IP:端口？

**A**: 使用 Ingress 配置域名路由:

```bash
# 1. 确保集群有 Ingress Controller
kubectl get pods -n ingress-nginx

# 2. 配置 DNS 将域名解析到集群节点 IP
# gitlabex.example.com -> NODE_IP

# 3. 修改 ingress.yaml
vim ingress.yaml
# 将域名修改为你的实际域名

# 4. 应用 Ingress 配置
kubectl apply -f ingress.yaml

# 5. 现在可以通过域名访问
# http://gitlabex.example.com
```

### Q3: 如何增加后端/前端的副本数？

**A**: 使用 `kubectl scale` 命令:

```bash
# 扩展后端到 5 个副本
kubectl scale deployment backend -n gitlabex --replicas=5

# 扩展前端到 3 个副本
kubectl scale deployment frontend -n gitlabex --replicas=3

# 查看结果
kubectl get pods -n gitlabex
```

### Q4: 如何升级应用到新版本？

**A**: 构建新镜像并滚动更新:

```bash
# 1. 构建新版本
cd /path/to/gitlabex2
./scripts/build-images.sh
docker tag gitlabex-backend:latest gitlabex-backend:v2.0

# 2. 更新 Deployment
kubectl set image deployment/backend backend=gitlabex-backend:v2.0 -n gitlabex

# 3. 监控更新过程
kubectl rollout status deployment/backend -n gitlabex

# 4. 如果有问题，立即回滚
kubectl rollout undo deployment/backend -n gitlabex
```

### Q5: 如何迁移到其他 Kubernetes 集群？

**A**: 按照以下步骤:

```bash
# 在旧集群:
# 1. 备份数据
./backups/backup-all.sh

# 2. 导出配置
kubectl get all,pvc,secrets,configmaps -n gitlabex -o yaml > cluster-export.yaml

# 在新集群:
# 3. 创建命名空间和 Secret
kubectl apply -f namespace.yaml
kubectl apply -f secrets.yaml

# 4. 部署应用
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
# ... 等等

# 5. 恢复数据
# 参考"备份和恢复"部分
```

### Q6: 如何监控资源使用情况？

**A**: 使用 `kubectl top` 和 Prometheus:

```bash
# 查看节点资源
kubectl top nodes

# 查看 Pod 资源
kubectl top pods -n gitlabex

# 查看 Metrics 端点（后端）
curl http://NODE_IP:30090/metrics
```

### Q7: 如何清理不再使用的镜像？

**A**: 在每个节点上执行:

```bash
# 查看所有镜像
docker images

# 清理未使用的镜像
docker image prune -a

# 清理所有未使用的资源（镜像、容器、网络、卷）
docker system prune -a --volumes
```

---

## 🎯 最佳实践建议

### 1. 资源管理

```yaml
# 为所有 Pod 设置合理的资源请求和限制
resources:
  requests:  # 最小保证资源
    memory: "512Mi"
    cpu: "500m"
  limits:    # 最大使用资源
    memory: "1Gi"
    cpu: "1000m"
```

### 2. 高可用配置

- **多副本**: 关键服务至少 2 个副本
- **Pod 反亲和性**: 确保副本分布在不同节点
- **就绪探针**: 正确配置健康检查
- **资源预留**: 避免资源耗尽

### 3. 安全加固

- ✅ 使用强密码（至少 16 位随机字符）
- ✅ 定期轮换 Secret 中的密钥
- ✅ 启用 RBAC 权限控制
- ✅ 使用 NetworkPolicy 限制网络访问
- ✅ 启用 TLS/SSL 加密传输
- ✅ 定期更新镜像修复漏洞

### 4. 备份策略

- **数据库**: 每天全量备份 + 事务日志
- **配置**: 版本控制 + 定期备份
- **文件**: 每周全量 + 每天增量
- **测试恢复**: 每月测试恢复流程

### 5. 监控告警

建议监控的关键指标:
- Pod 状态和重启次数
- CPU 和内存使用率
- 磁盘空间使用率
- 数据库连接数和慢查询
- HTTP 请求响应时间
- 错误率和日志异常

### 6. 日志管理

```bash
# 集中化日志管理建议:
# 1. 部署 EFK/ELK Stack
# 2. 或使用 Loki + Grafana
# 3. 或使用云服务（如 AWS CloudWatch）

# 日志保留策略:
# - 应用日志: 保留 30 天
# - 审计日志: 保留 90 天
# - 归档到对象存储: 1 年
```

---

## 📚 学习资源

### Kubernetes 基础

- [Kubernetes 官方文档](https://kubernetes.io/docs/)
- [Kubernetes 中文文档](https://kubernetes.io/zh-cn/docs/)
- [kubectl 速查表](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)

### 实践教程

- [Kubernetes By Example](http://kubernetesbyexample.com/)
- [Katacoda Kubernetes 交互式教程](https://www.katacoda.com/courses/kubernetes)

### 相关项目文档

- [GitLabEx 项目 README](../README.md)
- [Docker Compose 部署文档](../docker-compose.prod.yml)
- [第三方系统集成文档](../docs/SYNC_USER.md)
- [快速参考指南](QUICK_REFERENCE.md)

---

## 🎉 总结

恭喜你完成了 GitLabEx 的 Kubernetes 部署！

通过本教程,你已经学会了:
- ✅ Kubernetes 的核心概念（Pod, Service, Deployment, PVC 等）
- ✅ 如何使用 kubectl 管理集群
- ✅ 如何部署一个完整的多层应用
- ✅ 如何配置持久化存储
- ✅ 如何管理 Secret 和 ConfigMap
- ✅ 如何排查常见问题
- ✅ 如何备份和恢复数据

### 下一步

- 🔒 **加固安全**: 启用 HTTPS，配置网络策略
- 📊 **配置监控**: 部署 Prometheus + Grafana
- 🚀 **性能优化**: 根据实际负载调整资源配置
- 🔄 **自动化**: 使用 CI/CD 自动部署更新
- 📖 **深入学习**: 了解更多 Kubernetes 高级特性

### 需要帮助？

- 📝 查看[故障排查指南](#故障排查指南)
- 💬 提交 [GitHub Issues](https://github.com/your-org/gitlabex/issues)
- 📧 联系技术支持

---

**最后更新**: 2024-12-17  
**维护者**: GitLabEx Team  
**版本**: v1.0

**祝使用愉快！** 🎉
