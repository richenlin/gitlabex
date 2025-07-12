<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ApiService } from '../services/api'
import {
  Bell,
  Message,
  Warning,
  Check,
  Delete,
  Filter,
  Refresh,
  Setting,
  Document,
  FolderOpened,
  User,
  Calendar,
  CircleCheck,
  CircleClose,
  InfoFilled
} from '@element-plus/icons-vue'

interface Notification {
  id: number
  type: 'assignment' | 'project' | 'system' | 'reminder' | 'warning'
  title: string
  content: string
  created_at: string
  read: boolean
  priority: 'low' | 'medium' | 'high'
  category: string
  actions?: {
    label: string
    action: string
    type?: 'primary' | 'success' | 'warning' | 'danger'
  }[]
}

// 响应式数据
const loading = ref(false)
const notifications = ref<Notification[]>([])
const filterType = ref<string>('all')
const filterRead = ref<string>('all')
const selectedNotifications = ref<number[]>([])

// 通知类型配置
const notificationTypes = [
  { value: 'all', label: '全部', icon: Bell },
  { value: 'assignment', label: '作业', icon: Document },
  { value: 'project', label: '项目', icon: FolderOpened },
  { value: 'system', label: '系统', icon: Setting },
  { value: 'reminder', label: '提醒', icon: Calendar },
  { value: 'warning', label: '警告', icon: Warning }
]

// 读取状态筛选
const readStatusOptions = [
  { value: 'all', label: '全部' },
  { value: 'unread', label: '未读' },
  { value: 'read', label: '已读' }
]

// 模拟通知数据
const mockNotifications: Notification[] = [
  {
    id: 1,
    type: 'assignment',
    title: '作业提醒：数据结构实验',
    content: '数据结构实验作业将于明天 23:59 截止，请及时提交。',
    created_at: '2024-03-15 14:30:00',
    read: false,
    priority: 'high',
    category: '作业',
    actions: [
      { label: '查看作业', action: 'view-assignment', type: 'primary' },
      { label: '提交作业', action: 'submit-assignment', type: 'success' }
    ]
  },
  {
    id: 2,
    type: 'project',
    title: '项目更新：Web开发项目',
    content: '项目 "Web开发项目" 有新的提交，请查看最新进展。',
    created_at: '2024-03-15 10:15:00',
    read: true,
    priority: 'medium',
    category: '项目',
    actions: [
      { label: '查看项目', action: 'view-project', type: 'primary' }
    ]
  },
  {
    id: 3,
    type: 'system',
    title: '系统维护通知',
    content: '系统将于今晚 22:00-24:00 进行维护，期间可能影响服务使用。',
    created_at: '2024-03-14 16:45:00',
    read: false,
    priority: 'medium',
    category: '系统'
  },
  {
    id: 4,
    type: 'reminder',
    title: '课程提醒',
    content: '明天上午 9:00 有 "算法分析" 课程，请准时参加。',
    created_at: '2024-03-14 12:00:00',
    read: true,
    priority: 'low',
    category: '提醒'
  },
  {
    id: 5,
    type: 'warning',
    title: '作业逾期警告',
    content: '您有 2 个作业已逾期，请尽快联系教师处理。',
    created_at: '2024-03-13 09:30:00',
    read: false,
    priority: 'high',
    category: '警告',
    actions: [
      { label: '查看逾期作业', action: 'view-overdue', type: 'warning' }
    ]
  }
]

// 计算属性
const filteredNotifications = computed(() => {
  let filtered = notifications.value

  // 按类型筛选
  if (filterType.value !== 'all') {
    filtered = filtered.filter(n => n.type === filterType.value)
  }

  // 按读取状态筛选
  if (filterRead.value !== 'all') {
    filtered = filtered.filter(n => {
      return filterRead.value === 'read' ? n.read : !n.read
    })
  }

  // 按优先级排序
  return filtered.sort((a, b) => {
    const priorityOrder = { 'high': 3, 'medium': 2, 'low': 1 }
    return priorityOrder[b.priority] - priorityOrder[a.priority]
  })
})

const unreadCount = computed(() => {
  return notifications.value.filter(n => !n.read).length
})

const selectedCount = computed(() => {
  return selectedNotifications.value.length
})

// 生命周期
onMounted(() => {
  loadNotifications()
})

// 方法
const loadNotifications = async () => {
  loading.value = true
  try {
    const params = {
      type: filterType.value !== 'all' ? filterType.value : undefined,
      read: filterRead.value !== 'all' ? filterRead.value : undefined
    }
    
    const response = await ApiService.getNotifications(params)
    notifications.value = response
    ElMessage.success('通知加载成功')
  } catch (error) {
    console.error('加载通知失败:', error)
    ElMessage.error('加载通知失败')
    // 使用模拟数据作为后备
    notifications.value = mockNotifications
  } finally {
    loading.value = false
  }
}

const markAsRead = async (notificationId: number) => {
  try {
    await ApiService.markNotificationAsRead(notificationId)
    
    // 更新本地状态
    const notification = notifications.value.find(n => n.id === notificationId)
    if (notification) {
      notification.read = true
    }
    
    ElMessage.success('已标记为已读')
  } catch (error) {
    console.error('标记已读失败:', error)
    ElMessage.error('标记已读失败')
  }
}

const markAllAsRead = async () => {
  try {
    await ApiService.markAllNotificationsAsRead()
    
    // 更新本地状态
    notifications.value.forEach(n => {
      n.read = true
    })
    
    ElMessage.success('所有通知已标记为已读')
  } catch (error) {
    console.error('标记所有已读失败:', error)
    ElMessage.error('标记所有已读失败')
  }
}

const deleteNotification = async (notificationId: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这条通知吗？', '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await ApiService.deleteNotification(notificationId)
    
    // 更新本地状态
    const index = notifications.value.findIndex(n => n.id === notificationId)
    if (index > -1) {
      notifications.value.splice(index, 1)
    }
    
    ElMessage.success('通知已删除')
  } catch (error) {
    // 用户取消删除
    if (error !== 'cancel') {
      console.error('删除通知失败:', error)
      ElMessage.error('删除通知失败')
    }
  }
}

const deleteSelected = async () => {
  if (selectedNotifications.value.length === 0) {
    ElMessage.warning('请先选择要删除的通知')
    return
  }
  
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedNotifications.value.length} 条通知吗？`, '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await ApiService.deleteNotifications(selectedNotifications.value)
    
    // 更新本地状态
    notifications.value = notifications.value.filter(n => !selectedNotifications.value.includes(n.id))
    selectedNotifications.value = []
    
    ElMessage.success('选中的通知已删除')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error('批量删除失败')
    }
  }
}

const handleAction = (notification: Notification, action: string) => {
  switch (action) {
    case 'view-assignment':
      ElMessage.info(`跳转到作业：${notification.title}`)
      break
    case 'submit-assignment':
      ElMessage.info(`跳转到提交作业页面`)
      break
    case 'view-project':
      ElMessage.info(`跳转到项目页面`)
      break
    case 'view-overdue':
      ElMessage.info(`跳转到逾期作业页面`)
      break
    default:
      ElMessage.info(`执行操作：${action}`)
  }
}

const getTypeIcon = (type: string) => {
  switch (type) {
    case 'assignment':
      return Document
    case 'project':
      return FolderOpened
    case 'system':
      return Setting
    case 'reminder':
      return Calendar
    case 'warning':
      return Warning
    default:
      return Bell
  }
}

const getTypeColor = (type: string) => {
  switch (type) {
    case 'assignment':
      return '#409EFF'
    case 'project':
      return '#67C23A'
    case 'system':
      return '#909399'
    case 'reminder':
      return '#E6A23C'
    case 'warning':
      return '#F56C6C'
    default:
      return '#909399'
  }
}

const getPriorityColor = (priority: string) => {
  switch (priority) {
    case 'high':
      return '#F56C6C'
    case 'medium':
      return '#E6A23C'
    case 'low':
      return '#67C23A'
    default:
      return '#909399'
  }
}

const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(hours / 24)
  
  if (days > 0) {
    return `${days}天前`
  } else if (hours > 0) {
    return `${hours}小时前`
  } else {
    return '刚刚'
  }
}

const handleSelectAll = (checked: boolean) => {
  if (checked) {
    selectedNotifications.value = filteredNotifications.value.map(n => n.id)
  } else {
    selectedNotifications.value = []
  }
}
</script>

<template>
  <div class="notifications-view">
    <!-- 页面标题 -->
    <el-row class="page-header">
      <el-col :span="24">
        <div class="header-content">
          <h1 class="page-title">
            <el-icon><Bell /></el-icon>
            通知系统
            <el-badge :value="unreadCount" :hidden="unreadCount === 0" class="notification-badge">
              <span></span>
            </el-badge>
          </h1>
          <p class="page-description">管理您的通知消息，及时了解重要信息</p>
        </div>
      </el-col>
    </el-row>

    <!-- 操作栏 -->
    <el-row class="action-bar">
      <el-col :span="24">
        <el-card>
          <div class="action-content">
            <!-- 筛选器 -->
            <div class="filters">
              <el-select v-model="filterType" placeholder="通知类型" size="default">
                <el-option
                  v-for="type in notificationTypes"
                  :key="type.value"
                  :label="type.label"
                  :value="type.value"
                >
                  <div class="filter-option">
                    <el-icon><component :is="type.icon" /></el-icon>
                    <span>{{ type.label }}</span>
                  </div>
                </el-option>
              </el-select>
              
              <el-select v-model="filterRead" placeholder="读取状态" size="default">
                <el-option
                  v-for="status in readStatusOptions"
                  :key="status.value"
                  :label="status.label"
                  :value="status.value"
                />
              </el-select>
            </div>

            <!-- 操作按钮 -->
            <div class="actions">
              <el-button @click="loadNotifications" :loading="loading">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
              <el-button @click="markAllAsRead" :disabled="unreadCount === 0">
                <el-icon><CircleCheck /></el-icon>
                全部已读
              </el-button>
              <el-button @click="deleteSelected" :disabled="selectedCount === 0" type="danger">
                <el-icon><Delete /></el-icon>
                删除选中 ({{ selectedCount }})
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 通知列表 -->
    <el-row class="notifications-section">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="list-header">
              <el-checkbox 
                :model-value="selectedCount > 0 && selectedCount === filteredNotifications.length"
                :indeterminate="selectedCount > 0 && selectedCount < filteredNotifications.length"
                @change="handleSelectAll"
              >
                全选 ({{ filteredNotifications.length }} 条)
              </el-checkbox>
              <div class="list-info">
                <span>未读：{{ unreadCount }} 条</span>
              </div>
            </div>
          </template>
          
          <div class="notifications-list" v-loading="loading">
            <div 
              v-for="notification in filteredNotifications" 
              :key="notification.id"
              class="notification-item"
              :class="{ 'unread': !notification.read }"
            >
              <div class="notification-checkbox">
                <el-checkbox v-model="selectedNotifications" :value="notification.id" />
              </div>
              
              <div class="notification-icon">
                <el-icon :color="getTypeColor(notification.type)" size="20">
                  <component :is="getTypeIcon(notification.type)" />
                </el-icon>
              </div>
              
              <div class="notification-content">
                <div class="notification-header">
                  <div class="notification-title">{{ notification.title }}</div>
                  <div class="notification-meta">
                    <el-tag 
                      :color="getPriorityColor(notification.priority)" 
                      size="small"
                      effect="light"
                    >
                      {{ notification.priority === 'high' ? '高' : notification.priority === 'medium' ? '中' : '低' }}
                    </el-tag>
                    <span class="notification-time">{{ formatTime(notification.created_at) }}</span>
                  </div>
                </div>
                
                <div class="notification-body">{{ notification.content }}</div>
                
                <div class="notification-actions" v-if="notification.actions">
                  <el-button
                    v-for="action in notification.actions"
                    :key="action.action"
                    :type="action.type || 'default'"
                    size="small"
                    @click="handleAction(notification, action.action)"
                  >
                    {{ action.label }}
                  </el-button>
                </div>
              </div>
              
              <div class="notification-controls">
                <el-button 
                  v-if="!notification.read"
                  @click="markAsRead(notification.id)"
                  type="primary"
                  size="small"
                  text
                >
                  <el-icon><Check /></el-icon>
                  标记已读
                </el-button>
                <el-button 
                  @click="deleteNotification(notification.id)"
                  type="danger"
                  size="small"
                  text
                >
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </div>
            </div>
            
            <div v-if="filteredNotifications.length === 0" class="empty-state">
              <el-icon size="48" color="#ddd"><Bell /></el-icon>
              <p>暂无通知</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.notifications-view {
  padding: 20px;
}

.page-header {
  margin-bottom: 24px;
}

.header-content {
  text-align: center;
}

.page-title {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.notification-badge {
  position: relative;
  top: -4px;
}

.page-description {
  color: #606266;
  font-size: 16px;
  margin: 0;
}

.action-bar {
  margin-bottom: 24px;
}

.action-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.filters {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.filter-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.notifications-section {
  margin-bottom: 24px;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.list-info {
  color: #606266;
  font-size: 14px;
}

.notifications-list {
  min-height: 400px;
}

.notification-item {
  display: flex;
  gap: 16px;
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
  transition: all 0.3s;
}

.notification-item:hover {
  background-color: #f8f9fa;
}

.notification-item.unread {
  background-color: #f0f9ff;
  border-left: 4px solid #409EFF;
}

.notification-item:last-child {
  border-bottom: none;
}

.notification-checkbox {
  display: flex;
  align-items: flex-start;
  padding-top: 2px;
}

.notification-icon {
  display: flex;
  align-items: flex-start;
  padding-top: 2px;
}

.notification-content {
  flex: 1;
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
  gap: 12px;
}

.notification-title {
  font-weight: 500;
  color: #303133;
  font-size: 16px;
}

.notification-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.notification-time {
  color: #909399;
  font-size: 12px;
}

.notification-body {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 12px;
}

.notification-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.notification-controls {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: flex-end;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #909399;
}

.empty-state p {
  margin-top: 12px;
  font-size: 16px;
}

@media (max-width: 768px) {
  .notifications-view {
    padding: 16px;
  }
  
  .page-title {
    font-size: 24px;
  }
  
  .action-content {
    flex-direction: column;
    align-items: stretch;
  }
  
  .filters {
    justify-content: center;
  }
  
  .actions {
    justify-content: center;
  }
  
  .notification-item {
    flex-direction: column;
    gap: 12px;
  }
  
  .notification-checkbox {
    order: -1;
  }
  
  .notification-controls {
    flex-direction: row;
    justify-content: flex-end;
  }
  
  .notification-header {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  
  .notification-meta {
    justify-content: flex-end;
  }
}
</style> 