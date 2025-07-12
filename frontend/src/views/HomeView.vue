<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { ApiService, type User } from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { 
  Refresh, 
  User as UserIcon, 
  Document, 
  School,
  FolderOpened,
  Notebook
} from '@element-plus/icons-vue'

const authStore = useAuthStore()

// 响应式数据
const userLoading = ref(false)
const currentUser = ref<User | null>(null)

const educationStats = ref({
  classesCount: 0,
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

// 计算属性
const isTeacherOrAdmin = computed(() => {
  const userRole = authStore.userRole
  return userRole === 1 || userRole === 2 // 1: 管理员, 2: 教师
})

// 组件挂载时加载数据
onMounted(async () => {
  await loadUserInfo()
  await loadEducationStats()
  await loadRecentActivities()
})

// 加载用户信息
const loadUserInfo = async () => {
  userLoading.value = true
  try {
    currentUser.value = await ApiService.getCurrentUser()
  } catch (error) {
    console.error('加载用户信息失败:', error)
    ElMessage.error('加载用户信息失败')
  } finally {
    userLoading.value = false
  }
}

// 刷新用户信息
const refreshUserInfo = async () => {
  await loadUserInfo()
  ElMessage.success('用户信息已刷新')
}

// 加载教育统计数据
const loadEducationStats = async () => {
  try {
    // 使用可用的API或提供模拟数据
    const userRole = authStore.userRole
    educationStats.value = {
      classesCount: userRole === 2 ? 3 : 1, // 教师显示3个班级，学生显示1个
      activeProjectsCount: userRole === 2 ? 8 : 4, // 教师显示8个项目，学生显示4个
      pendingAssignmentsCount: userRole === 2 ? 12 : 2, // 教师显示12个待批改，学生显示2个待完成
      documentsCount: 15
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

// 加载最近活动
const loadRecentActivities = async () => {
  try {
    // 提供示例活动数据
    const activities = [
      {
        title: '欢迎使用 GitLabEx',
        description: '您已成功登录GitLabEx教育协作平台',
        timestamp: new Date().toLocaleString()
      },
      {
        title: '系统更新',
        description: '平台已更新至最新版本，新增了更多教育功能',
        timestamp: new Date(Date.now() - 3600000).toLocaleString()
      }
    ]
    recentActivities.value = activities
  } catch (error) {
    console.error('加载最近活动失败:', error)
  }
}

// 获取用户角色类型
const getUserRoleType = (role: number) => {
  switch (role) {
    case 1: return 'danger'  // 管理员
    case 2: return 'warning' // 教师
    case 3: return 'success' // 学生
    default: return 'info'   // 访客
  }
}

// 获取用户角色文本
const getUserRoleText = (role: number) => {
  switch (role) {
    case 1: return '管理员'
    case 2: return '教师'
    case 3: return '学生'
    default: return '访客'
  }
}
</script>

<template>
  <div class="dashboard-view">
    <div class="dashboard-header">
      <h1>首页（仪表盘）</h1>
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
              <el-tag :type="getUserRoleType(currentUser.role)">
                {{ getUserRoleText(currentUser.role) }}
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
      <el-col :xs="24" :sm="12" :md="6" v-if="isTeacherOrAdmin">
        <el-card class="action-card" shadow="hover" @click="$router.push('/classes')">
          <div class="action-content">
            <el-icon class="action-icon" size="32" color="#409EFF"><School /></el-icon>
            <h4>班级管理</h4>
            <p>创建和管理班级</p>
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
            <h4>文档管理</h4>
            <p>在线编辑文档</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 教育统计数据 -->
    <el-row :gutter="24" class="stats-section">
      <el-col :xs="24" :sm="12" :lg="6">
        <el-card>
          <el-statistic title="我的班级" :value="educationStats.classesCount" />
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
              <el-button text @click="loadRecentActivities">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
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

<style scoped>
.dashboard-view {
  padding: 20px;
}

.dashboard-header {
  margin-bottom: 20px;
  text-align: center;
}

.dashboard-header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.dashboard-header p {
  margin: 0;
  color: #606266;
  font-size: 16px;
}

.user-info-section {
  margin-bottom: 20px;
}

.user-card {
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-details h3 {
  margin: 0 0 8px 0;
  font-size: 20px;
  color: #303133;
}

.user-details p {
  margin: 4px 0;
  color: #606266;
  font-size: 14px;
}

.quick-actions-section {
  margin-bottom: 20px;
}

.action-card {
  cursor: pointer;
  transition: all 0.3s;
  border-radius: 8px;
  height: 120px;
}

.action-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
}

.action-content {
  text-align: center;
  padding: 20px;
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
  color: #606266;
  font-size: 14px;
}

.stats-section {
  margin-bottom: 20px;
}

.stats-section .el-card {
  border-radius: 8px;
  text-align: center;
}

.recent-activities-section {
  margin-bottom: 20px;
}

.no-activities {
  text-align: center;
  padding: 40px;
}

.no-user {
  text-align: center;
  padding: 40px;
}
</style>
