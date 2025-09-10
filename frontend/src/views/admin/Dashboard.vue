<template>
  <div class="admin-dashboard">
    <div class="header">
      <h1>管理员后台</h1>
      <div class="header-actions">
        <el-button @click="refreshStats" :loading="refreshing">
          <el-icon><Refresh /></el-icon>
          刷新数据
        </el-button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon user-icon">
                <el-icon><User /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-number">{{ stats.totalUsers }}</div>
                <div class="stat-label">总用户数</div>
                <div class="stat-change positive">+{{ stats.newUsersToday }} 今日新增</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon project-icon">
                <el-icon><Folder /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-number">{{ stats.totalProjects }}</div>
                <div class="stat-label">研究课题</div>
                <div class="stat-change positive">+{{ stats.newProjectsToday }} 今日新增</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon topic-icon">
                <el-icon><ChatDotRound /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-number">{{ stats.totalTopics }}</div>
                <div class="stat-label">话题讨论</div>
                <div class="stat-change positive">+{{ stats.newTopicsToday }} 今日新增</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-icon document-icon">
                <el-icon><Document /></el-icon>
              </div>
              <div class="stat-info">
                <div class="stat-number">{{ stats.totalDocuments }}</div>
                <div class="stat-label">文档资料</div>
                <div class="stat-change">{{ formatFileSize(stats.totalDocumentSize) }}</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :span="12">
        <el-card title="用户增长趋势">
          <div class="chart-container">
            <!-- TODO: 集成图表库 -->
            <div class="chart-placeholder">
              <p>用户增长趋势图表</p>
              <p>（需要集成 ECharts 或其他图表库）</p>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card title="活跃度统计">
          <div class="chart-container">
            <div class="chart-placeholder">
              <p>系统活跃度图表</p>
              <p>（需要集成 ECharts 或其他图表库）</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快速操作 -->
    <el-card title="快速操作" class="quick-actions">
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="action-item" @click="manageUsers">
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="action-item" @click="manageProjects">
            <el-icon><Folder /></el-icon>
            <span>课题管理</span>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="action-item" @click="manageContent">
            <el-icon><Document /></el-icon>
            <span>内容管理</span>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="action-item" @click="systemSettings">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- 最近活动 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card title="最近用户活动">
          <div class="activity-list">
            <div
              v-for="activity in recentActivities"
              :key="activity.id"
              class="activity-item"
            >
              <div class="activity-avatar">
                <el-avatar :size="32" :src="activity.user.avatar_url">
                  {{ activity.user.name.charAt(0) }}
                </el-avatar>
              </div>
              <div class="activity-content">
                <div class="activity-text">
                  <strong>{{ activity.user.name }}</strong> {{ activity.action }}
                </div>
                <div class="activity-time">{{ formatTime(activity.created_at) }}</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card title="系统状态">
          <div class="system-status">
            <div class="status-item">
              <div class="status-label">服务器状态</div>
              <el-tag type="success">正常运行</el-tag>
            </div>
            <div class="status-item">
              <div class="status-label">数据库状态</div>
              <el-tag type="success">连接正常</el-tag>
            </div>
            <div class="status-item">
              <div class="status-label">GitLab 集成</div>
              <el-tag type="success">连接正常</el-tag>
            </div>
            <div class="status-item">
              <div class="status-label">存储空间</div>
              <el-progress :percentage="storageUsage" :show-text="false" />
              <span class="storage-text">{{ storageUsage }}% 已使用</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 通知发布 -->
    <el-card title="发布系统通知" class="notification-panel">
      <el-form
        ref="notificationFormRef"
        :model="notificationForm"
        :rules="notificationRules"
        label-width="100px"
      >
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="通知标题" prop="title">
              <el-input v-model="notificationForm.title" placeholder="请输入通知标题" />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="通知类型" prop="type">
              <el-select v-model="notificationForm.type" placeholder="选择类型">
                <el-option label="系统通知" value="system" />
                <el-option label="重要公告" value="announcement" />
                <el-option label="维护通知" value="maintenance" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="目标用户" prop="target">
              <el-select v-model="notificationForm.target" placeholder="选择用户">
                <el-option label="全部用户" value="all" />
                <el-option label="教师" value="teachers" />
                <el-option label="学生" value="students" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="4">
            <el-form-item>
              <el-button type="primary" @click="sendNotification" :loading="sending">
                发送通知
              </el-button>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="通知内容" prop="content">
          <el-input
            v-model="notificationForm.content"
            type="textarea"
            :rows="3"
            placeholder="请输入通知内容"
          />
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import {
  Refresh, User, Folder, ChatDotRound, Document, Setting
} from '@element-plus/icons-vue'

const router = useRouter()

// 响应式数据
const refreshing = ref(false)
const sending = ref(false)
const notificationFormRef = ref<FormInstance>()

const stats = ref({
  totalUsers: 1248,
  newUsersToday: 12,
  totalProjects: 156,
  newProjectsToday: 3,
  totalTopics: 2341,
  newTopicsToday: 45,
  totalDocuments: 5628,
  totalDocumentSize: 2.8 * 1024 * 1024 * 1024 // 2.8GB
})

const recentActivities = ref([
  {
    id: '1',
    user: { name: '张三', avatar_url: '' },
    action: '创建了新的研究课题 "深度学习研究"',
    created_at: new Date(Date.now() - 30000).toISOString()
  },
  {
    id: '2',
    user: { name: '李四', avatar_url: '' },
    action: '在话题 "AI算法讨论" 中发表了评论',
    created_at: new Date(Date.now() - 300000).toISOString()
  },
  {
    id: '3',
    user: { name: '王五', avatar_url: '' },
    action: '上传了文档 "机器学习基础.pdf"',
    created_at: new Date(Date.now() - 600000).toISOString()
  },
  {
    id: '4',
    user: { name: '赵六', avatar_url: '' },
    action: '提交了作业 "数据结构实验报告"',
    created_at: new Date(Date.now() - 900000).toISOString()
  }
])

const storageUsage = ref(68)

const notificationForm = ref({
  title: '',
  content: '',
  type: 'system',
  target: 'all'
})

const notificationRules = {
  title: [{ required: true, message: '请输入通知标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入通知内容', trigger: 'blur' }],
  type: [{ required: true, message: '请选择通知类型', trigger: 'change' }],
  target: [{ required: true, message: '请选择目标用户', trigger: 'change' }]
}

// 方法
const refreshStats = async () => {
  refreshing.value = true
  try {
    // TODO: 调用 API 获取最新统计数据
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('数据刷新成功')
  } catch (error) {
    console.error('刷新数据失败:', error)
    ElMessage.error('刷新数据失败')
  } finally {
    refreshing.value = false
  }
}

const manageUsers = () => {
  ElMessage.info('用户管理功能待实现')
}

const manageProjects = () => {
  router.push('/scenes')
}

const manageContent = () => {
  router.push('/documents')
}

const systemSettings = () => {
  ElMessage.info('系统设置功能待实现')
}

const sendNotification = async () => {
  if (!notificationFormRef.value) return
  
  try {
    await notificationFormRef.value.validate()
    sending.value = true
    
    // TODO: 调用 API 发送通知
    await new Promise(resolve => setTimeout(resolve, 1500))
    
    ElMessage.success('通知发送成功')
    resetNotificationForm()
  } catch (error) {
    console.error('发送通知失败:', error)
    ElMessage.error('发送通知失败')
  } finally {
    sending.value = false
  }
}

const resetNotificationForm = () => {
  notificationForm.value = {
    title: '',
    content: '',
    type: 'system',
    target: 'all'
  }
}

const formatFileSize = (size: number) => {
  if (size < 1024) return `${size}B`
  if (size < 1024 * 1024) return `${Math.round(size / 1024)}KB`
  if (size < 1024 * 1024 * 1024) return `${Math.round(size / 1024 / 1024)}MB`
  return `${Math.round(size / 1024 / 1024 / 1024)}GB`
}

const formatTime = (time: string) => {
  const now = new Date()
  const activityTime = new Date(time)
  const diff = now.getTime() - activityTime.getTime()
  
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return activityTime.toLocaleDateString('zh-CN')
}

// 生命周期
onMounted(() => {
  // 初始化数据
})
</script>

<style scoped>
.admin-dashboard {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.stats-grid {
  margin-bottom: 30px;
}

.stat-card {
  height: 120px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 15px;
  height: 100%;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.user-icon { background: #e1f3d8; color: #67c23a; }
.project-icon { background: #ecf5ff; color: #409eff; }
.topic-icon { background: #fdf6ec; color: #e6a23c; }
.document-icon { background: #f4f4f5; color: #909399; }

.stat-info {
  flex: 1;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 5px;
}

.stat-change {
  font-size: 12px;
  color: #909399;
}

.stat-change.positive {
  color: #67c23a;
}

.charts-row {
  margin-bottom: 30px;
}

.chart-container {
  height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-placeholder {
  text-align: center;
  color: #909399;
  background: #f5f7fa;
  border-radius: 8px;
  padding: 40px;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.quick-actions {
  margin-bottom: 30px;
}

.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 20px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.action-item:hover {
  border-color: #409eff;
  background: #ecf5ff;
  color: #409eff;
}

.action-item .el-icon {
  font-size: 32px;
}

.action-item span {
  font-size: 14px;
  font-weight: 500;
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
  max-height: 300px;
  overflow-y: auto;
}

.activity-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.activity-content {
  flex: 1;
}

.activity-text {
  color: #303133;
  line-height: 1.5;
  margin-bottom: 5px;
}

.activity-time {
  font-size: 12px;
  color: #909399;
}

.system-status {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-label {
  font-weight: 500;
  color: #303133;
}

.storage-text {
  margin-left: 10px;
  font-size: 12px;
  color: #909399;
}

.notification-panel {
  margin-top: 30px;
}
</style>
