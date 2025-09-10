<template>
  <div class="notification-panel">
    <!-- 通知触发按钮 -->
    <el-popover
      placement="bottom-end"
      width="400"
      trigger="click"
      :visible="visible"
      @update:visible="handleVisibleChange"
    >
      <template #reference>
        <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99">
          <el-button 
            circle 
            size="large"
            :class="{ 'notification-btn': true, 'has-unread': unreadCount > 0 }"
          >
            <el-icon><Bell /></el-icon>
          </el-button>
        </el-badge>
      </template>

      <!-- 通知面板内容 -->
      <div class="notification-content">
        <!-- 头部 -->
        <div class="notification-header">
          <div class="header-title">
            <h3>通知中心</h3>
            <span class="connection-status" :class="{ connected: isConnected }">
              {{ isConnected ? '已连接' : '未连接' }}
            </span>
          </div>
          <div class="header-actions">
            <el-button 
              link 
              size="small" 
              @click="markAllAsRead"
              :disabled="unreadCount === 0"
            >
              全部已读
            </el-button>
            <el-button 
              link 
              size="small" 
              @click="clearAllNotifications"
              :disabled="notifications.length === 0"
            >
              清空
            </el-button>
          </div>
        </div>

        <!-- 通知列表 -->
        <div class="notification-list">
          <div v-if="notifications.length === 0" class="empty-state">
            <el-icon size="48" color="#c0c4cc"><Bell /></el-icon>
            <p>暂无通知</p>
          </div>

          <div
            v-for="notification in recentNotifications"
            :key="notification.id"
            class="notification-item"
            :class="{ 
              'unread': !notification.read,
              [`type-${notification.type}`]: true 
            }"
            @click="handleNotificationClick(notification)"
          >
            <div class="notification-icon">
              <el-icon>
                <component :is="getNotificationIcon(notification.type)" />
              </el-icon>
            </div>
            
            <div class="notification-body">
              <div class="notification-title">{{ notification.title }}</div>
              <div class="notification-content-text">{{ notification.content }}</div>
              <div class="notification-time">
                {{ formatTime(notification.timestamp) }}
              </div>
            </div>

            <div class="notification-actions">
              <el-button
                link
                size="small"
                @click.stop="markAsRead(notification.id)"
                v-if="!notification.read"
              >
                标记已读
              </el-button>
              <el-button
                link
                size="small"
                @click.stop="removeNotification(notification.id)"
              >
                删除
              </el-button>
            </div>
          </div>
        </div>

        <!-- 底部 -->
        <div class="notification-footer" v-if="notifications.length > 10">
          <el-button link @click="$router.push('/notifications')">
            查看全部通知
          </el-button>
        </div>
      </div>
    </el-popover>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useNotifications } from '@/composables/useNotifications'
import { 
  Bell, 
  Message, 
  Document, 
  User, 
  Warning, 
  InfoFilled,
  SuccessFilled,
  CircleCheck 
} from '@element-plus/icons-vue'

const router = useRouter()
const {
  notifications,
  isConnected,
  unreadCount,
  recentNotifications,
  markAsRead,
  markAllAsRead,
  removeNotification,
  clearAllNotifications
} = useNotifications()

const visible = ref(false)

// 处理面板显示状态变化
const handleVisibleChange = (val: boolean) => {
  visible.value = val
}

// 获取通知图标
const getNotificationIcon = (type: string) => {
  const iconMap: { [key: string]: any } = {
    system: InfoFilled,
    project: Document,
    homework: Document,
    user: User,
    success: SuccessFilled,
    warning: Warning,
    error: Warning,
    message: Message
  }
  return iconMap[type] || Bell
}

// 处理通知点击
const handleNotificationClick = (notification: any) => {
  markAsRead(notification.id)
  
  // 根据通知类型进行路由跳转
  if (notification.data) {
    switch (notification.type) {
      case 'project':
        if (notification.data.project_id) {
          router.push(`/projects/${notification.data.project_id}`)
          visible.value = false
        }
        break
      case 'homework':
        if (notification.data.homework_id) {
          router.push(`/homework/${notification.data.homework_id}`)
          visible.value = false
        }
        break
      default:
        break
    }
  }
}

// 格式化时间
const formatTime = (timestamp: number) => {
  const now = Date.now()
  const diff = now - timestamp
  
  const minute = 60 * 1000
  const hour = 60 * minute
  const day = 24 * hour
  
  if (diff < minute) {
    return '刚刚'
  } else if (diff < hour) {
    return `${Math.floor(diff / minute)}分钟前`
  } else if (diff < day) {
    return `${Math.floor(diff / hour)}小时前`
  } else {
    return new Date(timestamp).toLocaleDateString()
  }
}
</script>

<style scoped>
.notification-panel {
  position: relative;
}

.notification-btn {
  --el-button-hover-bg-color: var(--el-color-primary-light-9);
  --el-button-hover-border-color: var(--el-color-primary-light-7);
  transition: all 0.3s ease;
}

.notification-btn.has-unread {
  --el-button-bg-color: var(--el-color-primary-light-9);
  --el-button-border-color: var(--el-color-primary-light-7);
}

.notification-content {
  max-height: 500px;
  display: flex;
  flex-direction: column;
}

.notification-header {
  padding: 16px 16px 12px;
  border-bottom: 1px solid var(--el-border-color-light);
}

.header-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.header-title h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.connection-status {
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 4px;
  background-color: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}

.connection-status.connected {
  background-color: var(--el-color-success-light-9);
  color: var(--el-color-success);
}

.header-actions {
  display: flex;
  gap: 8px;
}

.notification-list {
  flex: 1;
  overflow-y: auto;
  max-height: 400px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: var(--el-text-color-placeholder);
}

.empty-state p {
  margin: 8px 0 0;
  font-size: 14px;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.notification-item:hover {
  background-color: var(--el-fill-color-light);
}

.notification-item.unread {
  background-color: var(--el-color-primary-light-9);
  border-left: 3px solid var(--el-color-primary);
}

.notification-icon {
  margin-right: 12px;
  margin-top: 2px;
  flex-shrink: 0;
}

.type-system .notification-icon {
  color: var(--el-color-info);
}

.type-project .notification-icon {
  color: var(--el-color-success);
}

.type-homework .notification-icon {
  color: var(--el-color-warning);
}

.type-error .notification-icon {
  color: var(--el-color-danger);
}

.notification-body {
  flex: 1;
  min-width: 0;
}

.notification-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--el-text-color-primary);
  margin-bottom: 4px;
  line-height: 1.4;
}

.notification-content-text {
  font-size: 13px;
  color: var(--el-text-color-regular);
  line-height: 1.4;
  margin-bottom: 6px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.notification-time {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

.notification-actions {
  margin-left: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.notification-item:hover .notification-actions {
  opacity: 1;
}

.notification-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--el-border-color-light);
  text-align: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .notification-content {
    width: 320px;
  }
  
  .notification-item {
    padding: 10px 12px;
  }
  
  .notification-actions {
    opacity: 1;
  }
}
</style>
