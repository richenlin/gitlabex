<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { ApiService } from '../services/api'
import {
  Star,
  DataBoard,
  Document,
  Plus,
  Edit,
  Folder,
  DocumentChecked,
  User,
  ChatDotRound
} from '@element-plus/icons-vue'

const router = useRouter()

// 响应式数据
const loading = ref(false)
const lastDocumentId = ref<number | null>(null)

// 功能特性数据
const features = ref([
  {
    title: '在线协作编辑',
    description: '基于 OnlyOffice 的实时文档协作，支持多人同时编辑，实时保存',
    icon: DocumentChecked,
    color: '#409EFF'
  },
  {
    title: '用户权限管理',
    description: '完整的用户权限体系，支持角色分配和权限控制',
    icon: User,
    color: '#67C23A'
  },
  {
    title: '教育场景优化',
    description: '针对教育场景的界面优化，提供更好的学习体验',
    icon: ChatDotRound,
    color: '#E6A23C'
  }
])

// 系统状态数据
const systemStatus = ref([
  {
    name: 'GitLabEx Backend',
    status: 'running',
    description: 'Go 后端服务'
  },
  {
    name: 'GitLab',
    status: 'running',
    description: 'GitLab CE 服务'
  },
  {
    name: 'OnlyOffice',
    status: 'running',
    description: '文档服务'
  },
  {
    name: 'PostgreSQL',
    status: 'running',
    description: '数据库服务'
  }
])

// 技术栈数据
const techStack = ref([
  'Vue 3',
  'TypeScript',
  'Element Plus',
  'Go',
  'Gin',
  'PostgreSQL',
  'Redis',
  'Docker',
  'GitLab',
  'OnlyOffice'
])

// 生命周期
onMounted(() => {
  checkSystemStatus()
})

// 方法
const createTestDocument = async () => {
  loading.value = true
  try {
    // 调用真实的API创建测试文档
    const response = await ApiService.createTestDocument()
    lastDocumentId.value = response.document_id
    
    ElMessage.success('测试文档创建成功！')
    ElNotification({
      title: '文档创建成功',
      message: `文档 ID: ${response.document_id}`,
      type: 'success'
    })
  } catch (error) {
    console.error('创建文档失败:', error)
    ElMessage.error('创建测试文档失败')
  } finally {
    loading.value = false
  }
}

const openEditor = () => {
  if (lastDocumentId.value) {
    router.push(`/documents/${lastDocumentId.value}/editor`)
  }
}

const viewDocuments = () => {
  router.push('/documents')
}

const checkSystemStatus = async () => {
  try {
    // 检查后端状态
    const healthResponse = await ApiService.healthCheck()
    if (healthResponse.status === 'ok') {
      systemStatus.value[0].status = 'running'
    } else {
      systemStatus.value[0].status = 'error'
    }
  } catch (error) {
    console.error('检查系统状态失败:', error)
    systemStatus.value[0].status = 'error'
  }
}
</script>

<template>
  <div class="home-view">
    <!-- 头部横幅 -->
    <el-row class="hero-section">
      <el-col :span="24">
        <div class="hero-content">
          <h1 class="hero-title">
            <el-icon class="hero-icon"><Star /></el-icon>
            GitLabEx
          </h1>
          <p class="hero-subtitle">基于 GitLab + OnlyOffice 的现代化教育协作平台</p>
          <div class="hero-actions">
            <el-button type="primary" size="large" @click="$router.push('/dashboard')">
              <el-icon><DataBoard /></el-icon>
              进入仪表板
            </el-button>
            <el-button size="large" @click="$router.push('/documents')">
              <el-icon><Document /></el-icon>
              文档管理
            </el-button>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 功能特性 -->
    <el-row :gutter="24" class="features-section">
      <el-col :xs="24" :sm="8" v-for="feature in features" :key="feature.title">
        <el-card class="feature-card" shadow="hover">
          <div class="feature-icon">
            <el-icon size="48" :color="feature.color">
              <component :is="feature.icon" />
            </el-icon>
          </div>
          <h3>{{ feature.title }}</h3>
          <p>{{ feature.description }}</p>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快速操作 -->
    <el-row class="quick-actions-section">
      <el-col :span="24">
        <h2 class="section-title">🎯 快速操作</h2>
        <div class="action-buttons">
          <el-button-group>
            <el-button type="primary" @click="createTestDocument" :loading="loading">
              <el-icon><Plus /></el-icon>
              创建测试文档
            </el-button>
            <el-button @click="openEditor" :disabled="!lastDocumentId">
              <el-icon><Edit /></el-icon>
              打开编辑器
            </el-button>
            <el-button @click="viewDocuments">
              <el-icon><Folder /></el-icon>
              查看所有文档
            </el-button>
          </el-button-group>
        </div>
      </el-col>
    </el-row>

    <!-- 系统状态 -->
    <el-row class="status-section">
      <el-col :span="24">
        <h2 class="section-title">🚦 系统状态</h2>
        <el-card>
          <el-row :gutter="16">
            <el-col :xs="24" :sm="12" :md="6" v-for="service in systemStatus" :key="service.name">
              <div class="status-item">
                <div class="status-info">
                  <span class="service-name">{{ service.name }}</span>
                  <el-tag :type="service.status === 'running' ? 'success' : 'danger'" size="small">
                    {{ service.status === 'running' ? '✅ 运行中' : '❌ 异常' }}
                  </el-tag>
                </div>
                <div class="status-details">{{ service.description }}</div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>

    <!-- 技术栈 -->
    <el-row class="tech-stack-section">
      <el-col :span="24">
        <h2 class="section-title">🔧 技术栈</h2>
        <div class="tech-items">
          <el-tag 
            v-for="tech in techStack" 
            :key="tech" 
            class="tech-item" 
            size="large"
            effect="plain"
          >
            {{ tech }}
          </el-tag>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.home-view {
  min-height: calc(100vh - 60px);
}

.hero-section {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 80px 20px;
  text-align: center;
  margin-bottom: 40px;
}

.hero-content {
  max-width: 800px;
  margin: 0 auto;
}

.hero-title {
  font-size: 3.5rem;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.hero-icon {
  font-size: 3.5rem;
}

.hero-subtitle {
  font-size: 1.3rem;
  margin-bottom: 40px;
  opacity: 0.9;
}

.hero-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
  flex-wrap: wrap;
}

.features-section {
  margin-bottom: 60px;
  padding: 0 20px;
}

.feature-card {
  text-align: center;
  height: 100%;
  transition: transform 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-5px);
}

.feature-icon {
  margin-bottom: 16px;
}

.feature-card h3 {
  font-size: 1.3rem;
  margin-bottom: 12px;
  color: #409EFF;
}

.feature-card p {
  color: #666;
  line-height: 1.6;
}

.quick-actions-section,
.status-section,
.tech-stack-section {
  margin-bottom: 40px;
  padding: 0 20px;
}

.section-title {
  text-align: center;
  margin-bottom: 30px;
  font-size: 1.8rem;
  color: #409EFF;
}

.action-buttons {
  text-align: center;
}

.status-item {
  padding: 16px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  margin-bottom: 16px;
}

.status-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.service-name {
  font-weight: 500;
  color: #303133;
}

.status-details {
  font-size: 12px;
  color: #909399;
}

.tech-items {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: center;
}

.tech-item {
  font-size: 14px;
  padding: 8px 16px;
}

@media (max-width: 768px) {
  .hero-title {
    font-size: 2.5rem;
  }
  
  .hero-actions {
    flex-direction: column;
    align-items: center;
  }
  
  .features-section {
    margin-bottom: 40px;
  }
}
</style>
