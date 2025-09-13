<template>
  <div class="profile-container">
    <!-- 用户信息卡片 -->
    <el-card class="user-info-card" v-loading="loading">
      <div class="user-header">
        <div class="user-avatar">
          <el-avatar :size="120" :src="userInfo?.avatar_url" :alt="userInfo?.name">
            <el-icon><User /></el-icon>
          </el-avatar>
        </div>
        <div class="user-details">
          <h2>{{ userInfo?.name }}</h2>
          <p class="username">@{{ userInfo?.username }}</p>
          <p class="email">{{ userInfo?.email }}</p>
          <div class="user-badges">
            <el-tag :type="getRoleColor(userInfo?.role || userInfo?.gitlab_role)" size="large">
              {{ getRoleText(userInfo?.role || userInfo?.gitlab_role) }}
            </el-tag>
            <el-tag type="info" size="small" v-if="userInfo?.edu_role">
              教育等级: {{ userInfo?.edu_role ? getEduRoleText(Number(userInfo.edu_role)) : '未设置' }}
            </el-tag>
          </div>
          <div class="user-stats">
            <el-statistic title="上次登录" :value="userInfo?.last_login_at ? formatDate(userInfo.last_login_at) : '从未登录'" />
            <el-statistic title="注册时间" :value="formatDate(userInfo?.created_at)" />
            <el-statistic title="活跃状态" :value="userInfo?.is_active ? '活跃' : '非活跃'" />
          </div>
        </div>
        <div class="user-actions">
          <el-button type="primary" @click="showEditProfile = true">
            <el-icon><Edit /></el-icon>
            编辑资料
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 资源统计 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <el-statistic title="创建的课题" :value="stats.projectsCount" />
          <div class="stat-icon project-icon">
            <el-icon><Folder /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <el-statistic title="发布的话题" :value="stats.topicsCount" />
          <div class="stat-icon topic-icon">
            <el-icon><ChatDotRound /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <el-statistic title="上传的文档" :value="stats.documentsCount" />
          <div class="stat-icon document-icon">
            <el-icon><Document /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <el-statistic title="作业提交" :value="stats.submissionsCount" />
          <div class="stat-icon homework-icon">
            <el-icon><EditPen /></el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近活动标签页 -->
    <el-card class="activity-card">
      <template #header>
        <div class="card-header">
          <h3>最近活动</h3>
          <el-button text @click="refreshActivity">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      
      <el-tabs v-model="activeTab" class="activity-tabs" @tab-change="handleTabChange">
        <el-tab-pane label="我的课题" name="projects">
          <div v-loading="projectsLoading">
            <div v-if="myProjects.length === 0" class="empty-state">
              <el-empty description="暂无课题" />
            </div>
            <div v-else class="resource-list">
              <div
                v-for="project in myProjects"
                :key="project.id"
                class="resource-item"
                @click="goToProject(project.id)"
              >
                <div class="resource-icon">
                  <el-icon><Folder /></el-icon>
                </div>
                <div class="resource-content">
                  <h4>{{ project.name }}</h4>
                  <p>{{ project.description }}</p>
                  <div class="resource-meta">
                    <el-tag size="small" :type="project.status === 'active' ? 'success' : 'info'">
                      {{ getProjectStatusText(project.status) }}
                    </el-tag>
                    <span class="meta-item">{{ formatDate(project.updated_at) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="我的话题" name="topics">
          <div v-loading="topicsLoading">
            <div v-if="myTopics.length === 0" class="empty-state">
              <el-empty description="暂无话题" />
            </div>
            <div v-else class="resource-list">
              <div
                v-for="topic in myTopics"
                :key="topic.id"
                class="resource-item"
                @click="goToTopic(topic.id)"
              >
                <div class="resource-icon">
                  <el-icon><ChatDotRound /></el-icon>
                </div>
                <div class="resource-content">
                  <h4>{{ topic.title }}</h4>
                  <p>{{ topic.content.substring(0, 100) }}...</p>
                  <div class="resource-meta">
                    <el-tag size="small" :type="topic.status === 'opened' ? 'success' : 'info'">
                      {{ topic.status === 'opened' ? '开放' : '已关闭' }}
                    </el-tag>
                    <span class="meta-item">{{ topic.likes_count }} 点赞</span>
                    <span class="meta-item">{{ topic.comments_count }} 评论</span>
                    <span class="meta-item">{{ formatDate(topic.updated_at) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="我的文档" name="documents">
          <div v-loading="documentsLoading">
            <div v-if="myDocuments.length === 0" class="empty-state">
              <el-empty description="暂无文档" />
            </div>
            <div v-else class="resource-list">
              <div
                v-for="document in myDocuments"
                :key="document.id"
                class="resource-item"
                @click="goToDocument(document.id)"
              >
                <div class="resource-icon">
                  <el-icon><Document /></el-icon>
                </div>
                <div class="resource-content">
                  <h4>{{ document.title }}</h4>
                  <p>{{ document.description }}</p>
                  <div class="resource-meta">
                    <el-tag size="small" type="info">{{ document.file_type }}</el-tag>
                    <span class="meta-item">{{ formatFileSize(document.file_size) }}</span>
                    <span class="meta-item">{{ document.download_count || 0 }} 下载</span>
                    <span class="meta-item">{{ formatDate(document.updated_at) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 编辑资料对话框 -->
    <el-dialog
      v-model="showEditProfile"
      title="编辑个人资料"
      width="500px"
      :before-close="handleEditClose"
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editFormRules"
        label-width="80px"
      >
        <el-form-item label="头像" prop="avatar_url">
          <div class="avatar-upload">
            <el-avatar :size="80" :src="editForm.avatar_url">
              <el-icon><User /></el-icon>
            </el-avatar>
            <div class="avatar-actions">
              <el-input
                v-model="editForm.avatar_url"
                placeholder="请输入头像URL"
                clearable
              />
              <el-text type="info" size="small">
                提示：请输入有效的图片URL
              </el-text>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="editForm.name" placeholder="请输入姓名" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="handleEditClose">取消</el-button>
          <el-button type="primary" :loading="editLoading" @click="handleSaveProfile">
            保存
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { 
  authService, 
  researchService, 
  topicService, 
  documentService
} from '@/services/api'
import type { User, ResearchProject, Topic, Document } from '@/types'
import { ElMessage } from 'element-plus'
import {
  User as UserIcon,
  Edit,
  Folder,
  ChatDotRound,
  Document as DocumentIcon,
  EditPen,
  Refresh
} from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const loading = ref(false)
const userInfo = ref<User | null>(null)
const showEditProfile = ref(false)
const editLoading = ref(false)
const activeTab = ref('projects')

// 统计数据
const stats = ref({
  projectsCount: 0,
  topicsCount: 0,
  documentsCount: 0,
  submissionsCount: 0
})

// 资源数据
const myProjects = ref<ResearchProject[]>([])
const myTopics = ref<Topic[]>([])
const myDocuments = ref<Document[]>([])

// 加载状态
const projectsLoading = ref(false)
const topicsLoading = ref(false)
const documentsLoading = ref(false)

// 编辑表单
const editFormRef = ref()
const editForm = ref({
  name: '',
  avatar_url: ''
})

const editFormRules = {
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    { min: 2, max: 50, message: '姓名长度应在 2 到 50 个字符', trigger: 'blur' }
  ]
}

// 方法
const fetchUserInfo = async () => {
  loading.value = true
  try {
    console.log('调用 authService.getCurrentUser()...')
    const response = await authService.getCurrentUser()
    console.log('API响应:', response)
    
    // 由于响应拦截器已经返回了data，所以response就是用户数据
    userInfo.value = response.data || response
    
    // 更新编辑表单
    if (userInfo.value) {
      editForm.value = {
        name: userInfo.value.name,
        avatar_url: userInfo.value.avatar_url || ''
      }
    }
    
    console.log('用户信息设置成功:', userInfo.value)
  } catch (error: any) {
    console.error('获取用户信息失败:', error)
    
    // 根据错误类型显示不同的提示
    if (error.response?.status === 401) {
      ElMessage.error('登录已过期，请重新登录')
      userStore.logout()
      router.push('/auth/login')
    } else if (error.response?.status === 403) {
      ElMessage.error('权限不足，无法访问个人资料')
    } else {
      ElMessage.error('获取用户信息失败，请稍后重试')
    }
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    // 并行获取统计数据
    const [projectsRes, topicsRes, documentsRes] = await Promise.allSettled([
      researchService.getProjects({ ownerId: userInfo.value?.id.toString() }),
      topicService.getTopics({ authorId: userInfo.value?.id.toString() }),
      documentService.getDocuments({ })
    ])

    // 由于响应拦截器已经返回了data，所以value就是实际数据
    stats.value.projectsCount = projectsRes.status === 'fulfilled' ? 
      ((projectsRes.value as any)?.data?.total || (projectsRes.value as any)?.total || (projectsRes.value as any)?.length || 0) : 0
    stats.value.topicsCount = topicsRes.status === 'fulfilled' ? 
      ((topicsRes.value as any)?.data?.total || (topicsRes.value as any)?.total || (topicsRes.value as any)?.length || 0) : 0
    stats.value.documentsCount = documentsRes.status === 'fulfilled' ? 
      ((documentsRes.value as any)?.data?.total || (documentsRes.value as any)?.total || (documentsRes.value as any)?.length || 0) : 0
      
    console.log('统计数据:', stats.value)
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

const fetchMyProjects = async () => {
  if (!userInfo.value?.id) return
  
  projectsLoading.value = true
  try {
    console.log('获取我的课题，用户ID:', userInfo.value.id)
    const response = await researchService.getProjects({ 
      ownerId: userInfo.value.id.toString(),
      pageSize: 10 
    })
    console.log('课题API响应:', response)
    
    // 由于响应拦截器已经返回了data，所以response就是实际数据
    const data = response?.data || response
    myProjects.value = data?.projects || data || []
    console.log('我的课题:', myProjects.value)
  } catch (error) {
    console.error('获取我的课题失败:', error)
  } finally {
    projectsLoading.value = false
  }
}

const fetchMyTopics = async () => {
  if (!userInfo.value?.id) return
  
  topicsLoading.value = true
  try {
    console.log('获取我的话题，用户ID:', userInfo.value.id)
    const response = await topicService.getTopics({ 
      authorId: userInfo.value.id.toString(),
      pageSize: 10 
    })
    console.log('话题API响应:', response)
    
    // 由于响应拦截器已经返回了data，所以response就是实际数据
    const data = response?.data || response
    myTopics.value = data?.topics || data || []
    console.log('我的话题:', myTopics.value)
  } catch (error) {
    console.error('获取我的话题失败:', error)
  } finally {
    topicsLoading.value = false
  }
}

const fetchMyDocuments = async () => {
  documentsLoading.value = true
  try {
    console.log('获取我的文档')
    const response = await documentService.getDocuments({ pageSize: 10 })
    console.log('文档API响应:', response)
    
    // 过滤出当前用户上传的文档
    // 由于响应拦截器已经返回了data，所以response就是实际数据
    const data = response?.data || response
    const allDocs = data?.documents || data || []
    myDocuments.value = allDocs.filter((doc: Document) => doc.uploader_id === userInfo.value?.id.toString())
    console.log('我的文档:', myDocuments.value)
  } catch (error) {
    console.error('获取我的文档失败:', error)
  } finally {
    documentsLoading.value = false
  }
}

const refreshActivity = async () => {
  switch (activeTab.value) {
    case 'projects':
      await fetchMyProjects()
      break
    case 'topics':
      await fetchMyTopics()
      break
    case 'documents':
      await fetchMyDocuments()
      break
  }
}

const handleSaveProfile = async () => {
  if (!editFormRef.value) return
  
  try {
    await editFormRef.value.validate()
    editLoading.value = true
    
    await authService.updateProfile(editForm.value)
    
    // 更新本地用户信息
    if (userInfo.value) {
      userInfo.value.name = editForm.value.name
      userInfo.value.avatar_url = editForm.value.avatar_url
    }
    
    // 更新store中的用户信息
    userStore.updateUser(editForm.value)
    
    ElMessage.success('个人资料更新成功')
    showEditProfile.value = false
  } catch (error) {
    console.error('更新个人资料失败:', error)
    ElMessage.error('更新个人资料失败')
  } finally {
    editLoading.value = false
  }
}

const handleEditClose = () => {
  showEditProfile.value = false
  // 重置表单
  if (userInfo.value) {
    editForm.value = {
      name: userInfo.value.name,
      avatar_url: userInfo.value.avatar_url || ''
    }
  }
}

// 导航方法
const goToProject = (id: string) => {
  router.push(`/scenes/${id}`)
}

const goToTopic = (id: string) => {
  router.push(`/topics/${id}`)
}

const goToDocument = (id: string) => {
  router.push(`/documents/${id}`)
}

// 标签页切换处理
const handleTabChange = (tabName: string) => {
  switch (tabName) {
    case 'projects':
      if (myProjects.value.length === 0) fetchMyProjects()
      break
    case 'topics':
      if (myTopics.value.length === 0) fetchMyTopics()
      break
    case 'documents':
      if (myDocuments.value.length === 0) fetchMyDocuments()
      break
  }
}

// 工具方法
const getRoleColor = (role?: string) => {
  const colorMap: Record<string, string> = {
    admin: 'danger',
    teacher: 'warning',
    assistant: 'info',
    student: 'success',
    guest: ''
  }
  return colorMap[role || ''] || ''
}

const getRoleText = (role?: string) => {
  const textMap: Record<string, string> = {
    admin: '管理员',
    teacher: '教师',
    assistant: '助教',
    student: '学生',
    guest: '访客'
  }
  return textMap[role || ''] || role || ''
}

const getEduRoleText = (eduRole?: number) => {
  const textMap: Record<number, string> = {
    50: '管理员',
    40: '教师',
    30: '研究员',
    20: '学生',
    10: '访客'
  }
  return textMap[eduRole || 0] || '未知'
}

const getProjectStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    active: '活跃',
    archived: '已归档',
    suspended: '已暂停'
  }
  return textMap[status] || status
}

const formatDate = (date?: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

const formatFileSize = (size: number) => {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

// 监听标签页变化
watch(() => activeTab.value, handleTabChange)

// 生命周期
onMounted(async () => {
  console.log('Profile页面加载，检查登录状态...')
  console.log('Token:', userStore.token)
  console.log('User:', userStore.user)
  console.log('IsLoggedIn:', userStore.isLoggedIn)

  // 检查是否有token
  if (!userStore.token) {
    console.warn('没有token，重定向到登录页面')
    ElMessage.warning('请先登录以查看个人资料')
    router.push('/auth/login')
    return
  }

  // 如果有token但没有用户信息，先尝试获取用户信息
  if (!userStore.user) {
    console.log('有token但没有用户信息，尝试获取用户信息...')
    const success = await userStore.fetchCurrentUser()
    if (!success) {
      console.warn('获取用户信息失败，重定向到登录页面')
      ElMessage.error('登录状态已过期，请重新登录')
      router.push('/auth/login')
      return
    }
  }

  // 如果用户已登录，从store获取用户信息
  if (userStore.user) {
    userInfo.value = userStore.user
    editForm.value = {
      name: userStore.user.name,
      avatar_url: userStore.user.avatar_url || ''
    }
  }

  console.log('开始获取用户详细信息和统计数据...')
  await fetchUserInfo()
  await fetchStats()
  
  // 默认加载第一个标签页的数据
  await fetchMyProjects()
})
</script>

<style scoped>
.profile-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.user-info-card {
  margin-bottom: 20px;
}

.user-header {
  display: flex;
  gap: 30px;
  align-items: flex-start;
}

.user-avatar {
  flex-shrink: 0;
}

.user-details {
  flex: 1;
}

.user-details h2 {
  margin: 0 0 10px 0;
  color: #303133;
  font-size: 28px;
}

.username {
  color: #909399;
  margin: 0 0 5px 0;
  font-size: 16px;
}

.email {
  color: #606266;
  margin: 0 0 15px 0;
}

.user-badges {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.user-stats {
  display: flex;
  gap: 40px;
}

.user-actions {
  flex-shrink: 0;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  position: relative;
  overflow: hidden;
}

.stat-icon {
  position: absolute;
  right: 20px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 40px;
  opacity: 0.1;
}

.project-icon { color: #67c23a; }
.topic-icon { color: #409eff; }
.document-icon { color: #909399; }
.homework-icon { color: #e6a23c; }

.activity-card {
  min-height: 500px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  color: #303133;
}

.activity-tabs {
  margin-top: 20px;
}

.resource-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.resource-item {
  display: flex;
  align-items: flex-start;
  gap: 15px;
  padding: 15px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.resource-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.resource-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #f0f9ff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #409eff;
  flex-shrink: 0;
}

.resource-content {
  flex: 1;
}

.resource-content h4 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 16px;
}

.resource-content p {
  margin: 0 0 10px 0;
  color: #606266;
  line-height: 1.5;
}

.resource-meta {
  display: flex;
  align-items: center;
  gap: 15px;
  font-size: 12px;
  color: #909399;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.empty-state {
  text-align: center;
  padding: 40px;
}

.avatar-upload {
  display: flex;
  align-items: center;
  gap: 20px;
}

.avatar-actions {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dialog-footer {
  text-align: right;
}
</style>
