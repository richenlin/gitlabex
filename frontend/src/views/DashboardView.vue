<template>
  <div class="dashboard-view">
    <div class="dashboard-header">
      <h1>仪表板</h1>
      <p>欢迎使用 GitLabEx 教育协作平台</p>
    </div>

    <!-- 用户信息卡片 -->
    <el-row :gutter="24" class="user-info-section">
      <el-col :span="24">
        <el-card class="user-card" v-loading="userLoading">
          <template #header>
            <div class="card-header">
              <span>👤 个人信息</span>
              <el-button text @click="refreshUserInfo">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          <div v-if="currentUser" class="user-info">
            <el-avatar :size="64" :src="currentUser.avatar">
              <el-icon><UserIcon /></el-icon>
            </el-avatar>
            <div class="user-details">
              <h3>{{ currentUser.name }}</h3>
              <p>@{{ currentUser.username }}</p>
              <p>{{ currentUser.email }}</p>
              <el-tag :type="currentUser.is_active ? 'success' : 'danger'">
                {{ currentUser.is_active ? '活跃' : '非活跃' }}
              </el-tag>
            </div>
          </div>
          <div v-else class="no-user">
            <el-empty description="未获取到用户信息" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快速操作 - 教育功能 -->
    <el-row :gutter="24" class="quick-actions-section">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="action-card" shadow="hover" @click="$router.push('/permissions')" v-if="isAdmin">
          <div class="action-content">
            <el-icon class="action-icon" size="32" color="#409EFF"><School /></el-icon>
            <h4>权限管理</h4>
            <p>管理用户权限和角色</p>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="action-card" shadow="hover" @click="$router.push('/projects')">
          <div class="action-content">
            <el-icon class="action-icon" size="32" color="#67C23A"><FolderOpened /></el-icon>
            <h4>课题管理</h4>
            <p>创建和跟踪课题</p>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="action-card" shadow="hover" @click="$router.push('/assignments')">
          <div class="action-content">
            <el-icon class="action-icon" size="32" color="#E6A23C"><Notebook /></el-icon>
            <h4>作业管理</h4>
            <p>布置和批改作业</p>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="action-card" shadow="hover" @click="$router.push('/documents')">
          <div class="action-content">
            <el-icon class="action-icon" size="32" color="#F56C6C"><Document /></el-icon>
            <h4>文档协作</h4>
            <p>在线编辑文档</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 教育统计数据 -->
    <el-row :gutter="24" class="stats-section">
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card>
          <el-statistic title="我的课题" :value="educationStats.activeProjectsCount" />
          <template #suffix>
            <el-icon><School /></el-icon>
          </template>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card>
          <el-statistic title="进行中课题" :value="educationStats.activeProjectsCount" />
          <template #suffix>
            <el-icon><FolderOpened /></el-icon>
          </template>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card>
          <el-statistic title="待批改作业" :value="educationStats.pendingAssignmentsCount" />
          <template #suffix>
            <el-icon><Notebook /></el-icon>
          </template>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card>
          <el-statistic title="协作文档" :value="educationStats.documentsCount" />
          <template #suffix>
            <el-icon><Document /></el-icon>
          </template>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近活动 -->
    <el-row class="recent-activities-section">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>📝 最近活动</span>
            </div>
          </template>
          <el-timeline>
            <el-timeline-item
              v-for="(activity, index) in recentActivities"
              :key="index"
              :timestamp="activity.timestamp"
              placement="top"
            >
              <el-card>
                <h4>{{ activity.title }}</h4>
                <p>{{ activity.description }}</p>
              </el-card>
            </el-timeline-item>
          </el-timeline>
          <div v-if="recentActivities.length === 0" class="no-activities">
            <el-empty description="暂无最近活动" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { ApiService, type User } from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { 
  Refresh, 
  User as UserIcon, 
  Plus, 
  Document, 
  Setting, 
  Folder,
  School,
  FolderOpened,
  Notebook
} from '@element-plus/icons-vue'

// 响应式数据
const authStore = useAuthStore()
const userLoading = ref(false)
const currentUser = ref<User | null>(null)

// 计算属性
const isAdmin = computed(() => {
  const userRole = authStore.userRole
  return userRole === 1 // 1: 管理员
})
const stats = ref({
  documentsCount: 0,
  activeUsers: 0,
  projectsCount: 0
})

const educationStats = ref({
  activeProjectsCount: 0,
  pendingAssignmentsCount: 0,
  documentsCount: 0
})

const recentActivities = ref([
  {
    title: '欢迎使用 GitLabEx',
    description: '这是您的第一次访问仪表板',
    timestamp: new Date().toLocaleString()
  }
])

// 组件挂载时加载数据
onMounted(async () => {
  await loadUserInfo()
  await loadStats()
  await loadEducationStats()
})

// 加载用户信息
const loadUserInfo = async () => {
  userLoading.value = true
  try {
    currentUser.value = await ApiService.getCurrentUser()
  } catch (error) {
    console.error('获取用户信息失败:', error)
    ElMessage.error('获取用户信息失败')
  } finally {
    userLoading.value = false
  }
}

// 加载统计数据
const loadStats = async () => {
  try {
    // 获取活跃用户数
    const activeUsersData = await ApiService.getActiveUsers()
    stats.value.activeUsers = activeUsersData.total
    
    // 其他统计数据暂时使用模拟数据
    stats.value.documentsCount = Math.floor(Math.random() * 10) + 1
    stats.value.projectsCount = Math.floor(Math.random() * 5) + 1
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

// 加载教育统计数据
const loadEducationStats = async () => {
  try {
    // 从API获取教育统计数据
    const response = await fetch('/api/education/stats', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      educationStats.value = data.data || {
        activeProjectsCount: 0,
        pendingAssignmentsCount: 0,
        documentsCount: 0
      }
    } else {
      // 使用模拟数据
      educationStats.value = {
        activeProjectsCount: Math.floor(Math.random() * 8) + 2,
        pendingAssignmentsCount: Math.floor(Math.random() * 10) + 1,
        documentsCount: Math.floor(Math.random() * 15) + 3
      }
    }
  } catch (error) {
    console.error('获取教育统计数据失败:', error)
    // 使用模拟数据
    educationStats.value = {
      activeProjectsCount: Math.floor(Math.random() * 8) + 2,
      pendingAssignmentsCount: Math.floor(Math.random() * 10) + 1,
      documentsCount: Math.floor(Math.random() * 15) + 3
    }
  }
}

// 刷新用户信息
const refreshUserInfo = async () => {
  await loadUserInfo()
  ElMessage.success('用户信息已刷新')
}

// 创建文档
const createDocument = async () => {
  try {
    const result = await ApiService.createTestDocument()
    ElMessage.success(`文档创建成功，ID: ${result.document_id}`)
    
    // 添加到最近活动
    recentActivities.value.unshift({
      title: '创建文档',
      description: `成功创建文档，ID: ${result.document_id}`,
      timestamp: new Date().toLocaleString()
    })
    
    // 更新统计数据
    stats.value.documentsCount++
  } catch (error) {
    console.error('创建文档失败:', error)
    ElMessage.error('创建文档失败')
  }
}
</script>

<style scoped>
.dashboard-view {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.dashboard-header {
  margin-bottom: 24px;
}

.dashboard-header h1 {
  font-size: 28px;
  color: #303133;
  margin-bottom: 8px;
}

.dashboard-header p {
  color: #909399;
  font-size: 16px;
}

.user-info-section,
.quick-actions-section,
.stats-section,
.recent-activities-section {
  margin-bottom: 24px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-card .user-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-details h3 {
  margin: 0 0 8px 0;
  font-size: 20px;
  color: #303133;
}

.user-details p {
  margin: 4px 0;
  color: #606266;
}

.no-user {
  text-align: center;
  padding: 20px;
}

.action-card {
  cursor: pointer;
  transition: transform 0.3s ease;
}

.action-card:hover {
  transform: translateY(-4px);
}

.action-content {
  text-align: center;
  padding: 16px 8px;
}

.action-icon {
  margin-bottom: 12px;
}

.action-content h4 {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: #303133;
}

.action-content p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.no-activities {
  text-align: center;
  padding: 20px;
}

@media (max-width: 768px) {
  .dashboard-view {
    padding: 16px;
  }
  
  .dashboard-header h1 {
    font-size: 24px;
  }
  
  .user-info {
    flex-direction: column;
    text-align: center;
  }
}
</style> 