<template>
  <div class="notification-bell">
    <el-badge :value="unreadCount" :hidden="unreadCount === 0" class="notification-badge">
      <el-button 
        type="text" 
        @click="showNotifications"
        class="notification-button"
        :class="{ 'has-unread': unreadCount > 0 }"
      >
        <el-icon size="20">
          <Bell />
        </el-icon>
        <span class="notification-text">公告</span>
      </el-button>
    </el-badge>
    
    <!-- 公告弹出框 -->
    <el-drawer
      v-model="notificationDrawer"
      title="公告通知"
      :direction="'rtl'"
      size="400px"
      :show-close="true"
    >
      <div class="notification-drawer">
        <!-- 操作按钮 -->
        <div class="notification-actions">
          <el-button size="small" @click="markAllAsRead" :disabled="unreadCount === 0">
            全部已读
          </el-button>
          <el-button size="small" @click="refreshNotifications">
            刷新
          </el-button>
        </div>
        
        <!-- 公告列表 -->
        <div class="notification-list">
          <div 
            v-for="notification in notifications" 
            :key="notification.id"
            class="notification-item"
            :class="{ 'unread': !notification.is_read }"
            @click="viewNotification(notification)"
          >
            <div class="notification-header">
              <span class="notification-title">{{ notification.title }}</span>
              <span class="notification-time">{{ formatTime(notification.created_at) }}</span>
            </div>
            <div class="notification-content">
              {{ notification.content }}
            </div>
            <div class="notification-status">
              <el-tag v-if="!notification.is_read" type="danger" size="small">未读</el-tag>
              <el-tag v-else type="success" size="small">已读</el-tag>
            </div>
          </div>
          
          <!-- 空状态 -->
          <div v-if="notifications.length === 0" class="notification-empty">
            <el-empty description="暂无公告" :image-size="80" />
          </div>
        </div>
      </div>
    </el-drawer>
    
    <!-- 公告详情对话框 -->
    <el-dialog
      v-model="notificationDetailDialog"
      :title="selectedNotification?.title || '公告详情'"
      width="600px"
      :close-on-click-modal="false"
    >
      <div class="notification-detail">
        <div class="notification-meta">
          <el-tag type="info" size="small">{{ selectedNotification?.type }}</el-tag>
          <span class="notification-date">{{ formatFullTime(selectedNotification?.created_at || '') }}</span>
        </div>
        <div class="notification-body">
          {{ selectedNotification?.content }}
        </div>
      </div>
      <template #footer>
        <el-button @click="notificationDetailDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Bell } from '@element-plus/icons-vue'
import { ApiService } from '../services/api'

interface Notification {
  id: number
  title: string
  content: string
  type: string
  is_read: boolean
  created_at: string
}

// 响应式数据
const notifications = ref<Notification[]>([])
const notificationDrawer = ref(false)
const notificationDetailDialog = ref(false)
const selectedNotification = ref<Notification | null>(null)
const loading = ref(false)

// 计算属性
const unreadCount = computed(() => {
  return notifications.value.filter(n => !n.is_read).length
})

// 生命周期
onMounted(() => {
  loadNotifications()
  // 定时刷新通知 (每30秒)
  setInterval(loadNotifications, 30000)
})

// 方法
const loadNotifications = async () => {
  try {
    loading.value = true
    const response = await ApiService.getNotifications()
    notifications.value = response.data || []
  } catch (error) {
    console.error('加载通知失败:', error)
  } finally {
    loading.value = false
  }
}

const showNotifications = () => {
  notificationDrawer.value = true
}

const viewNotification = async (notification: Notification) => {
  selectedNotification.value = notification
  notificationDetailDialog.value = true
  
  // 如果是未读通知，标记为已读
  if (!notification.is_read) {
    await markAsRead(notification.id)
  }
}

const markAsRead = async (notificationId: number) => {
  try {
    await ApiService.markNotificationAsRead(notificationId)
    // 更新本地状态
    const notification = notifications.value.find(n => n.id === notificationId)
    if (notification) {
      notification.is_read = true
    }
  } catch (error) {
    console.error('标记通知已读失败:', error)
  }
}

const markAllAsRead = async () => {
  try {
    await ApiService.markAllNotificationsAsRead()
    // 更新本地状态
    notifications.value.forEach(n => {
      n.is_read = true
    })
    ElMessage.success('已标记所有通知为已读')
  } catch (error) {
    console.error('标记全部已读失败:', error)
    ElMessage.error('标记全部已读失败')
  }
}

const refreshNotifications = () => {
  loadNotifications()
}

const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMinutes = Math.floor(diffMs / (1000 * 60))
  const diffHours = Math.floor(diffMinutes / 60)
  const diffDays = Math.floor(diffHours / 24)
  
  if (diffMinutes < 60) {
    return `${diffMinutes}分钟前`
  } else if (diffHours < 24) {
    return `${diffHours}小时前`
  } else if (diffDays < 7) {
    return `${diffDays}天前`
  } else {
    return date.toLocaleDateString('zh-CN')
  }
}

const formatFullTime = (timeStr: string) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN')
}
</script>

<style scoped>
.notification-bell {
  position: relative;
}

.notification-button {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  border-radius: 6px;
  transition: all 0.3s;
}

.notification-button:hover {
  background-color: #f0f0f0;
}

.notification-button.has-unread {
  color: #409EFF;
}

.notification-text {
  font-size: 14px;
  color: #333;
}

.notification-badge :deep(.el-badge__content) {
  background-color: #ff4d4f;
  border-color: #ff4d4f;
}

.notification-drawer {
  padding: 16px;
}

.notification-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #eee;
}

.notification-list {
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

.notification-item {
  padding: 12px;
  border: 1px solid #eee;
  border-radius: 8px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.notification-item:hover {
  background-color: #f8f9fa;
  border-color: #409EFF;
}

.notification-item.unread {
  background-color: #f0f9ff;
  border-color: #409EFF;
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.notification-title {
  font-weight: 600;
  color: #333;
  font-size: 14px;
}

.notification-time {
  font-size: 12px;
  color: #999;
}

.notification-content {
  font-size: 13px;
  color: #666;
  margin-bottom: 8px;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.notification-status {
  display: flex;
  justify-content: flex-end;
}

.notification-empty {
  text-align: center;
  padding: 40px 20px;
}

.notification-detail {
  max-height: 400px;
  overflow-y: auto;
}

.notification-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #eee;
}

.notification-date {
  font-size: 14px;
  color: #999;
}

.notification-body {
  font-size: 14px;
  line-height: 1.6;
  color: #333;
}
</style> 