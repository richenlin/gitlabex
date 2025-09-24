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
          </div>
          <div class="header-actions">
            <!-- 暂时屏蔽全部已读功能，因为GitLab没有通知已读机制 -->
            <!-- <el-button 
              link 
              size="small" 
              @click="markAllAsRead"
              :disabled="unreadCount === 0"
              :loading="markAllLoading"
            >
              全部已读
            </el-button> -->
            <el-button 
              link 
              size="small" 
              @click="refreshNotifications"
              :loading="loading"
            >
              刷新
            </el-button>
          </div>
        </div>

        <!-- 通知列表 -->
        <div class="notification-list" v-loading="loading">
          <div v-if="notifications.length === 0 && !loading" class="empty-state">
            <el-icon size="48" color="#c0c4cc"><Bell /></el-icon>
            <p>暂无通知</p>
          </div>

          <div
            v-for="notification in notifications"
            :key="notification.id"
            class="notification-item"
            :class="{ 
              // 暂时屏蔽未读状态样式，因为GitLab没有通知已读机制
              // 'unread': !notification.read,
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
              <div class="notification-title">{{ notification.title || 'GitLab 通知' }}</div>
              <div class="notification-content-text">{{ getNotificationContent(notification) }}</div>
              <div class="notification-time">
                {{ formatTime(notification.created_at || notification.timestamp) }}
              </div>
            </div>

            <div class="notification-actions">
              <!-- 暂时屏蔽标记已读功能，因为GitLab没有通知已读机制 -->
              <!-- <el-button
                link
                size="small"
                @click.stop="markAsRead(notification.id)"
                v-if="!notification.read"
                :loading="markingRead === notification.id"
              >
                标记已读
              </el-button> -->
            </div>
          </div>
        </div>

        <!-- 分页 -->
        <div class="notification-footer" v-if="notifications.length > 0">
          <el-pagination
            v-model:current-page="currentPage"
            :page-size="pageSize"
            :total="total"
            layout="prev, pager, next"
            small
            @current-change="handlePageChange"
          />
        </div>
      </div>
    </el-popover>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { authService } from '@/services/api'
import { 
  Bell, 
  Message, 
  Document, 
  User, 
  Warning, 
  InfoFilled,
  SuccessFilled,
  CircleCheck,
  Link,
  ArrowRight,
  Switch
} from '@element-plus/icons-vue'

const router = useRouter()

// 响应式数据
const visible = ref(false)
const notifications = ref<any[]>([])
const loading = ref(false)
const markAllLoading = ref(false)
const markingRead = ref<string | null>(null)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 计算属性
const unreadCount = computed(() => {
  // 暂时屏蔽未读计数，因为GitLab没有通知已读机制
  // return notifications.value.filter(n => !n.read).length
  return 0
})

// 获取通知列表
const fetchNotifications = async () => {
  try {
    loading.value = true
    console.log('获取通知列表...', { page: currentPage.value, per_page: pageSize.value })
    
    const response = await authService.getNotifications({
      page: currentPage.value,
      per_page: pageSize.value
    })
    
    console.log('通知API响应:', response)
    
    const data = response?.data || response
    notifications.value = data?.notifications || []
    total.value = data?.total || notifications.value.length
    
    console.log('通知列表:', notifications.value)
  } catch (error) {
    console.error('获取通知列表失败:', error)
    ElMessage.error('获取通知列表失败')
    notifications.value = []
  } finally {
    loading.value = false
  }
}

// 标记单个通知为已读
const markAsRead = async (notificationId: string) => {
  try {
    markingRead.value = notificationId
    console.log('标记通知为已读:', notificationId)
    
    await authService.markNotificationAsRead(notificationId)
    
    // 更新本地状态
    const notification = notifications.value.find(n => n.id === notificationId)
    if (notification) {
      notification.read = true
    }
    
    ElMessage.success('通知已标记为已读')
  } catch (error) {
    console.error('标记通知为已读失败:', error)
    ElMessage.error('标记通知为已读失败')
  } finally {
    markingRead.value = null
  }
}

// 标记所有通知为已读
const markAllAsRead = async () => {
  try {
    markAllLoading.value = true
    console.log('标记所有通知为已读')
    
    await authService.markAllNotificationsAsRead()
    
    // 更新本地状态
    notifications.value.forEach(n => {
      n.read = true
    })
    
    ElMessage.success('所有通知已标记为已读')
  } catch (error) {
    console.error('标记所有通知为已读失败:', error)
    ElMessage.error('标记所有通知为已读失败')
  } finally {
    markAllLoading.value = false
  }
}

// 刷新通知
const refreshNotifications = async () => {
  await fetchNotifications()
}

// 处理面板显示状态变化
const handleVisibleChange = (val: boolean) => {
  visible.value = val
  if (val) {
    fetchNotifications()
  }
}

// 处理分页变化
const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchNotifications()
}

// 获取通知图标
const getNotificationIcon = (type: string) => {
  const iconMap: { [key: string]: any } = {
    issue: Warning,
    merge_request: ArrowRight,
    commit: Link,
    push: Switch,
    project: Document,
    user: User,
    success: SuccessFilled,
    warning: Warning,
    error: Warning,
    message: Message,
    system: InfoFilled
  }
  return iconMap[type] || Bell
}

// 获取通知内容
const getNotificationContent = (notification: any) => {
  // GitLab通知的内容通常在body或message字段中
  if (notification.body) {
    return notification.body
  }
  if (notification.message) {
    return notification.message
  }
  if (notification.subject) {
    return notification.subject.title || notification.subject.name || 'GitLab 通知'
  }
  return 'GitLab 通知'
}

// 处理通知点击
const handleNotificationClick = (notification: any) => {
  // 暂时屏蔽自动标记已读功能，因为GitLab没有通知已读机制
  // markAsRead(notification.id)
  
  // 根据GitLab通知类型进行路由跳转
  if (notification.subject) {
    switch (notification.subject.type) {
      case 'MergeRequest':
        if (notification.subject.id) {
          // 跳转到合并请求页面
          window.open(`${notification.project?.web_url}/-/merge_requests/${notification.subject.id}`, '_blank')
          visible.value = false
        }
        break
      case 'Issue':
        if (notification.subject.id) {
          // 跳转到问题页面
          window.open(`${notification.project?.web_url}/-/issues/${notification.subject.id}`, '_blank')
          visible.value = false
        }
        break
      case 'Commit':
        if (notification.subject.id) {
          // 跳转到提交页面
          window.open(`${notification.project?.web_url}/-/commit/${notification.subject.id}`, '_blank')
          visible.value = false
        }
        break
      case 'Project':
        if (notification.project?.web_url) {
          // 跳转到项目页面
          window.open(notification.project.web_url, '_blank')
          visible.value = false
        }
        break
      default:
        break
    }
  }
}

// 格式化时间
const formatTime = (timestamp: string | number) => {
  if (!timestamp) return '未知时间'
  
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  const minute = 60 * 1000
  const hour = 60 * minute
  const day = 24 * hour
  
  if (diff < minute) {
    return '刚刚'
  } else if (diff < hour) {
    return `${Math.floor(diff / minute)}分钟前`
  } else if (diff < day) {
    return `${Math.floor(diff / hour)}小时前`
  } else if (diff < 7 * day) {
    return `${Math.floor(diff / day)}天前`
  } else {
    return date.toLocaleDateString()
  }
}

// 监听面板显示状态，自动刷新通知
watch(visible, (newVal) => {
  if (newVal) {
    fetchNotifications()
  }
})

// 组件挂载时获取通知
onMounted(() => {
  // 延迟获取，避免在用户未登录时调用
  setTimeout(() => {
    fetchNotifications()
  }, 1000)
})
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

.type-issue .notification-icon {
  color: var(--el-color-warning);
}

.type-merge_request .notification-icon {
  color: var(--el-color-success);
}

.type-commit .notification-icon {
  color: var(--el-color-info);
}

.type-push .notification-icon {
  color: var(--el-color-primary);
}

.type-project .notification-icon {
  color: var(--el-color-success);
}

.type-system .notification-icon {
  color: var(--el-color-info);
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
