<template>
  <div class="profile-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <el-icon class="page-icon"><UserFilled /></el-icon>
          个人资料
        </h1>
        <p class="page-description">管理您的个人信息和偏好设置</p>
      </div>
    </div>

    <!-- 个人资料内容 -->
    <div class="profile-content">
      <el-row :gutter="20">
        <!-- 左侧 - 基本信息 -->
        <el-col :span="16">
          <el-card class="profile-card">
            <template #header>
              <div class="card-header">
                <span class="card-title">
                  <el-icon><Edit /></el-icon>
                  基本信息
                </span>
                <el-button 
                  v-if="!isEditing" 
                  type="primary" 
                  :icon="Edit"
                  @click="startEdit"
                >
                  编辑
                </el-button>
              </div>
            </template>

            <el-form 
              ref="profileFormRef"
              :model="profileForm" 
              :rules="profileFormRules"
              label-width="120px"
              :disabled="!isEditing"
            >
              <el-form-item label="姓名" prop="name">
                <el-input 
                  v-model="profileForm.name" 
                  placeholder="请输入您的姓名"
                  :readonly="!isEditing"
                />
              </el-form-item>
              
              <el-form-item label="用户名" prop="username">
                <el-input 
                  v-model="profileForm.username" 
                  placeholder="请输入用户名"
                  :readonly="!isEditing"
                />
              </el-form-item>
              
              <el-form-item label="邮箱" prop="email">
                <el-input 
                  v-model="profileForm.email" 
                  placeholder="请输入邮箱"
                  :readonly="!isEditing"
                />
              </el-form-item>
              
              <el-form-item label="GitLab ID">
                <el-input 
                  v-model="profileForm.gitlab_id" 
                  placeholder="GitLab用户ID"
                  readonly
                />
              </el-form-item>
              
              <el-form-item label="角色">
                <el-tag :type="getRoleTagType(profileForm.role)">
                  {{ getRoleText(profileForm.role) }}
                </el-tag>
              </el-form-item>
              
              <el-form-item label="最后同步">
                <span class="sync-time">{{ formatDate(profileForm.last_sync_at) }}</span>
              </el-form-item>
              
              <el-form-item v-if="isEditing">
                <el-button 
                  type="primary" 
                  @click="saveProfile"
                  :loading="isSaving"
                >
                  保存
                </el-button>
                <el-button @click="cancelEdit">
                  取消
                </el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </el-col>

        <!-- 右侧 - 头像和统计 -->
        <el-col :span="8">
          <!-- 头像卡片 -->
          <el-card class="avatar-card">
            <template #header>
              <div class="card-header">
                <span class="card-title">
                  <el-icon><Camera /></el-icon>
                  头像
                </span>
              </div>
            </template>

            <div class="avatar-section">
              <div class="avatar-container">
                <el-avatar 
                  :src="profileForm.avatar" 
                  :size="120"
                  :alt="profileForm.name"
                >
                  <el-icon :size="40"><UserFilled /></el-icon>
                </el-avatar>
                
                <el-upload
                  v-if="isEditing"
                  class="avatar-upload"
                  action=""
                  :show-file-list="false"
                  :auto-upload="false"
                  :on-change="handleAvatarChange"
                  accept="image/*"
                >
                  <el-button type="text" class="upload-btn">
                    <el-icon><Plus /></el-icon>
                    更换头像
                  </el-button>
                </el-upload>
              </div>
              
              <div class="user-basic-info">
                <h3 class="user-name">{{ profileForm.name }}</h3>
                <p class="user-email">{{ profileForm.email }}</p>
                <el-tag :type="getRoleTagType(profileForm.role)" size="small">
                  {{ getRoleText(profileForm.role) }}
                </el-tag>
              </div>
            </div>
          </el-card>

          <!-- 统计卡片 -->
          <el-card class="stats-card">
            <template #header>
              <div class="card-header">
                <span class="card-title">
                  <el-icon><DataAnalysis /></el-icon>
                  统计信息
                </span>
              </div>
            </template>

            <div class="stats-grid">
              <div class="stat-item">
                <div class="stat-value">{{ stats.documents_count }}</div>
                <div class="stat-label">文档数量</div>
              </div>
              <div class="stat-item">
                <div class="stat-value">{{ stats.recent_activities.length }}</div>
                <div class="stat-label">最近活动</div>
              </div>
              <div class="stat-item">
                <div class="stat-value">{{ stats.project_memberships.length }}</div>
                <div class="stat-label">项目成员</div>
              </div>
              <div class="stat-item">
                <div class="stat-value">{{ getDaysActive() }}</div>
                <div class="stat-label">活跃天数</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 活动记录 -->
      <el-card class="activity-card">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><Clock /></el-icon>
              最近活动
            </span>
            <el-button 
              type="text" 
              :icon="Refresh"
              @click="loadUserData"
            >
              刷新
            </el-button>
          </div>
        </template>

        <el-timeline>
          <el-timeline-item
            v-for="activity in stats.recent_activities"
            :key="activity.document_id"
            :timestamp="formatDate(activity.updated_at)"
            placement="top"
          >
            <el-card class="activity-item">
              <div class="activity-content">
                <div class="activity-info">
                  <el-icon class="activity-icon"><Document /></el-icon>
                  <div class="activity-text">
                    <span class="activity-action">{{ getActivityAction(activity.type) }}</span>
                    <span class="activity-target">{{ activity.filename }}</span>
                  </div>
                </div>
                <el-tag size="small" :type="getActivityTagType(activity.type)">
                  {{ activity.type }}
                </el-tag>
              </div>
            </el-card>
          </el-timeline-item>
        </el-timeline>

        <div v-if="!stats.recent_activities.length" class="no-activity">
          <el-empty description="暂无活动记录" />
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  UserFilled,
  Edit,
  Camera,
  Plus,
  DataAnalysis,
  Clock,
  Refresh,
  Document
} from '@element-plus/icons-vue'
import { ApiService } from '../services/api'
import type { User, UserDashboard } from '../services/api'

// 响应式数据
const isEditing = ref(false)
const isSaving = ref(false)
const isLoading = ref(false)

const profileForm = ref<User>({
  id: 0,
  gitlab_id: 0,
  username: '',
  email: '',
  name: '',
  avatar: '',
  role: 2,
  last_sync_at: '',
  is_active: true
})

const originalProfile = ref<User>({
  id: 0,
  gitlab_id: 0,
  username: '',
  email: '',
  name: '',
  avatar: '',
  role: 2,
  last_sync_at: '',
  is_active: true
})

const stats = ref({
  documents_count: 0,
  recent_activities: [] as Array<{
    type: string
    document_id: number
    filename: string
    updated_at: string
  }>,
  project_memberships: [] as Array<{
    project_id: number
    project_name: string
    role: string
    web_url: string
  }>
})

// 表单验证规则
const profileFormRules = {
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' }
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ]
}

// 生命周期
onMounted(() => {
  loadUserData()
})

// 方法
const loadUserData = async () => {
  isLoading.value = true
  try {
    const dashboard: any = await ApiService.getUserDashboard()
    
    profileForm.value = { ...dashboard.user }
    originalProfile.value = { ...dashboard.user }
    
    // 适配后端返回的数据结构到前端期望的格式
    stats.value = {
      documents_count: 0, // 暂时使用默认值
      recent_activities: [], // 暂时使用空数组
      project_memberships: dashboard.projects ? dashboard.projects.map((project: any) => ({
        project_id: project.id,
        project_name: project.name,
        role: 'member',
        web_url: project.web_url
      })) : []
    }
    
    ElMessage.success('个人资料加载成功')
  } catch (error) {
    console.error('加载个人资料失败:', error)
    ElMessage.error('加载个人资料失败')
  } finally {
    isLoading.value = false
  }
}

const startEdit = () => {
  isEditing.value = true
  originalProfile.value = { ...profileForm.value }
}

const cancelEdit = () => {
  isEditing.value = false
  profileForm.value = { ...originalProfile.value }
}

const saveProfile = async () => {
  isSaving.value = true
  try {
    // 这里应该调用更新用户资料API
    // await ApiService.updateUserProfile(profileForm.value)
    
    // 模拟保存
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    originalProfile.value = { ...profileForm.value }
    isEditing.value = false
    
    ElMessage.success('个人资料保存成功')
  } catch (error) {
    console.error('保存个人资料失败:', error)
    ElMessage.error('保存个人资料失败')
  } finally {
    isSaving.value = false
  }
}

const handleAvatarChange = (file: any) => {
  // 模拟头像上传
  const reader = new FileReader()
  reader.onload = (e: any) => {
    profileForm.value.avatar = e.target.result
  }
  reader.readAsDataURL(file.raw)
}

// 工具方法
const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN')
}

const getRoleText = (role: number) => {
  const roleMap: { [key: number]: string } = {
    1: '访客',
    2: '学生',
    3: '助教',
    4: '教师',
    5: '管理员'
  }
  return roleMap[role] || '未知'
}

const getRoleTagType = (role: number) => {
  const typeMap: { [key: number]: string } = {
    1: 'info',
    2: 'primary',
    3: 'success',
    4: 'warning',
    5: 'danger'
  }
  return typeMap[role] || 'info'
}

const getActivityAction = (type: string) => {
  const actionMap: { [key: string]: string } = {
    'create': '创建了',
    'update': '更新了',
    'delete': '删除了',
    'view': '查看了'
  }
  return actionMap[type] || '操作了'
}

const getActivityTagType = (type: string) => {
  const typeMap: { [key: string]: string } = {
    'create': 'success',
    'update': 'primary',
    'delete': 'danger',
    'view': 'info'
  }
  return typeMap[type] || 'info'
}

const getDaysActive = () => {
  // 模拟计算活跃天数
  const now = new Date()
  const lastSync = new Date(profileForm.value.last_sync_at)
  const diffTime = Math.abs(now.getTime() - lastSync.getTime())
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
  return Math.max(1, Math.min(diffDays, 30)) // 限制在1-30天之间
}
</script>

<style scoped>
.profile-container {
  padding: 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 100vh;
}

.page-header {
  margin-bottom: 20px;
  background: white;
  padding: 24px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.header-content h1 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-icon {
  font-size: 28px;
  color: #409eff;
}

.page-description {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
}

.profile-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.profile-card,
.avatar-card,
.stats-card,
.activity-card {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  border-radius: 12px;
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #2c3e50;
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.avatar-container {
  position: relative;
  margin-bottom: 20px;
}

.avatar-upload {
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
}

.upload-btn {
  padding: 4px 8px;
  font-size: 12px;
  color: #409eff;
}

.user-basic-info {
  width: 100%;
}

.user-name {
  margin: 0 0 8px 0;
  color: #2c3e50;
  font-size: 18px;
}

.user-email {
  margin: 0 0 12px 0;
  color: #7f8c8d;
  font-size: 14px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  text-align: center;
}

.stat-item {
  padding: 12px;
  border-radius: 8px;
  background: #f8f9ff;
  border: 1px solid #e1e6ff;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #409eff;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #7f8c8d;
}

.activity-item {
  margin-bottom: 0;
  border: 1px solid #e4e7ed;
  box-shadow: none;
}

.activity-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.activity-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.activity-icon {
  color: #409eff;
  font-size: 16px;
}

.activity-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.activity-action {
  font-size: 14px;
  color: #2c3e50;
}

.activity-target {
  font-size: 12px;
  color: #7f8c8d;
}

.sync-time {
  color: #7f8c8d;
  font-size: 14px;
}

.no-activity {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .profile-container {
    padding: 10px;
  }
  
  .el-col {
    width: 100%;
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }
  
  .activity-content {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }
}
</style> 