<template>
  <div class="notifications">
    <div class="header">
      <h1>通知中心</h1>
      <div class="actions">
        <el-button @click="markAllAsRead" :disabled="unreadCount === 0">
          全部标记为已读
        </el-button>
        <el-button type="danger" @click="clearAll">
          清空所有通知
        </el-button>
      </div>
    </div>

    <div class="notification-stats">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-statistic title="总通知数" :value="total" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="未读通知" :value="unreadCount" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="今日通知" :value="todayCount" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="重要通知" :value="importantCount" />
        </el-col>
      </el-row>
    </div>

    <div class="filters">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-select v-model="typeFilter" placeholder="通知类型" clearable @change="fetchNotifications">
            <el-option label="全部" value="" />
            <el-option label="系统通知" value="system" />
            <el-option label="作业通知" value="homework" />
            <el-option label="课题通知" value="project" />
            <el-option label="话题通知" value="topic" />
            <el-option label="文档通知" value="document" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="statusFilter" placeholder="状态" @change="fetchNotifications">
            <el-option label="全部" value="" />
            <el-option label="未读" value="unread" />
            <el-option label="已读" value="read" />
          </el-select>
        </el-col>
      </el-row>
    </div>

    <div class="notification-list" v-loading="loading">
      <div
        v-for="notification in notifications"
        :key="notification.id"
        class="notification-item"
        :class="{ 'unread': !notification.is_read, 'important': isImportant(notification) }"
        @click="handleNotificationClick(notification)"
      >
        <div class="notification-icon">
          <el-icon :class="getNotificationIconClass(notification.type)">
            <component :is="getNotificationIcon(notification.type)" />
          </el-icon>
        </div>

        <div class="notification-content">
          <div class="notification-header">
            <h4 class="notification-title">{{ notification.title }}</h4>
            <div class="notification-meta">
              <el-tag :type="getNotificationTypeColor(notification.type)" size="small">
                {{ getNotificationTypeText(notification.type) }}
              </el-tag>
              <span class="notification-time">{{ formatTime(notification.created_at) }}</span>
            </div>
          </div>
          
          <div class="notification-body">
            {{ notification.content }}
          </div>

          <div class="notification-related" v-if="notification.related_type">
            <el-link :href="getRelatedLink(notification)" :underline="false">
              查看相关 {{ getRelatedTypeText(notification.related_type) }}
            </el-link>
          </div>
        </div>

        <div class="notification-actions" @click.stop>
          <el-button 
            v-if="!notification.is_read"
            size="small" 
            @click="markAsRead(notification.id)"
          >
            标记已读
          </el-button>
          <el-button 
            size="small" 
            type="danger" 
            @click="deleteNotification(notification.id)"
          >
            删除
          </el-button>
        </div>
      </div>

      <div v-if="notifications.length === 0" class="empty-state">
        <el-empty description="暂无通知" />
      </div>
    </div>

    <div class="pagination" v-if="total > pageSize">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="fetchNotifications"
        @size-change="fetchNotifications"
      />
    </div>

    <!-- 公告区域 -->
    <div class="announcements" v-if="announcements.length > 0">
      <h2>系统公告</h2>
      <div
        v-for="announcement in announcements"
        :key="announcement.id"
        class="announcement-item"
      >
        <div class="announcement-header">
          <h3>{{ announcement.title }}</h3>
          <span class="announcement-time">{{ formatDate(announcement.created_at) }}</span>
        </div>
        <div class="announcement-content">
          {{ announcement.content }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { notificationService } from '@/services/api'
import type { Notification } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Bell, Warning, Document, Folder, ChatDotRound, 
  HomeFilled, User, Setting
} from '@element-plus/icons-vue'

const router = useRouter()

// 响应式数据
const notifications = ref<Notification[]>([])
const announcements = ref<any[]>([])
const loading = ref(false)
const typeFilter = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 计算属性
const unreadCount = computed(() => {
  return notifications.value.filter(n => !n.is_read).length
})

const todayCount = computed(() => {
  const today = new Date().toDateString()
  return notifications.value.filter(n => 
    new Date(n.created_at).toDateString() === today
  ).length
})

const importantCount = computed(() => {
  return notifications.value.filter(n => isImportant(n)).length
})

// 方法
const fetchNotifications = async () => {
  loading.value = true
  try {
    const response = await notificationService.getNotifications({
      page: currentPage.value,
      pageSize: pageSize.value,
      isRead: statusFilter.value === 'read' ? true : statusFilter.value === 'unread' ? false : undefined
    })
    notifications.value = response.notifications || []
    total.value = response.pagination?.total || 0
  } catch (error) {
    console.error('获取通知失败:', error)
    ElMessage.error('获取通知失败')
  } finally {
    loading.value = false
  }
}

const fetchAnnouncements = async () => {
  try {
    const response = await notificationService.getAnnouncements()
    announcements.value = (response || []).slice(0, 5)
  } catch (error) {
    console.error('获取公告失败:', error)
  }
}

const markAsRead = async (id: string) => {
  try {
    await notificationService.markAsRead(id)
    const notification = notifications.value.find(n => n.id === id)
    if (notification) {
      notification.is_read = true
    }
    ElMessage.success('已标记为已读')
  } catch (error) {
    console.error('标记已读失败:', error)
    ElMessage.error('标记已读失败')
  }
}

const markAllAsRead = async () => {
  try {
    await notificationService.markAllAsRead()
    notifications.value.forEach(n => n.is_read = true)
    ElMessage.success('全部通知已标记为已读')
  } catch (error) {
    console.error('批量标记已读失败:', error)
    ElMessage.error('批量标记已读失败')
  }
}

const deleteNotification = async (id: string) => {
  try {
    await notificationService.deleteNotification(id)
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
      notifications.value.splice(index, 1)
      total.value--
    }
    ElMessage.success('通知删除成功')
  } catch (error) {
    console.error('删除通知失败:', error)
    ElMessage.error('删除通知失败')
  }
}

const clearAll = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有通知吗？此操作不可恢复。', '确认清空', {
      type: 'warning'
    })
    
    await notificationService.clearAll()
    notifications.value = []
    total.value = 0
    ElMessage.success('所有通知已清空')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('清空通知失败:', error)
      ElMessage.error('清空通知失败')
    }
  }
}

const handleNotificationClick = async (notification: Notification) => {
  // 标记为已读
  if (!notification.is_read) {
    await markAsRead(notification.id)
  }
  
  // 跳转到相关页面
  if (notification.related_type && notification.related_id) {
    const link = getRelatedLink(notification)
    if (link) {
      router.push(link)
    }
  }
}

const getNotificationIcon = (type: string) => {
  const iconMap: Record<string, any> = {
    system: Setting,
    homework: Document,
    project: Folder,
    topic: ChatDotRound,
    document: Document,
    user: User,
    warning: Warning,
    info: Bell
  }
  return iconMap[type] || Bell
}

const getNotificationIconClass = (type: string) => {
  const classMap: Record<string, string> = {
    system: 'system-icon',
    homework: 'homework-icon',
    project: 'project-icon',
    topic: 'topic-icon',
    document: 'document-icon',
    user: 'user-icon',
    warning: 'warning-icon',
    info: 'info-icon'
  }
  return classMap[type] || 'default-icon'
}

const getNotificationTypeColor = (type: string) => {
  const colorMap: Record<string, string> = {
    system: 'info',
    homework: 'warning',
    project: 'success',
    topic: 'primary',
    document: 'info',
    user: 'success',
    warning: 'danger',
    info: 'info'
  }
  return colorMap[type] || 'info'
}

const getNotificationTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    system: '系统',
    homework: '作业',
    project: '课题',
    topic: '话题',
    document: '文档',
    user: '用户',
    warning: '警告',
    info: '信息'
  }
  return textMap[type] || type
}

const getRelatedTypeText = (relatedType: string) => {
  const textMap: Record<string, string> = {
    homework: '作业',
    project: '课题',
    topic: '话题',
    document: '文档'
  }
  return textMap[relatedType] || relatedType
}

const getRelatedLink = (notification: Notification) => {
  if (!notification.related_type || !notification.related_id) return ''
  
  const linkMap: Record<string, string> = {
    homework: `/homeworks/${notification.related_id}`,
    project: `/scenes/${notification.related_id}`,
    topic: `/topics/${notification.related_id}`,
    document: `/documents/${notification.related_id}`
  }
  return linkMap[notification.related_type] || ''
}

const isImportant = (notification: Notification) => {
  return notification.type === 'warning' || notification.type === 'system'
}

const formatTime = (time: string) => {
  const now = new Date()
  const notificationTime = new Date(time)
  const diff = now.getTime() - notificationTime.getTime()
  
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`
  return notificationTime.toLocaleDateString('zh-CN')
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('zh-CN')
}

// 生命周期
onMounted(() => {
  fetchNotifications()
  fetchAnnouncements()
})
</script>

<style scoped>
.notifications {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.actions {
  display: flex;
  gap: 10px;
}

.notification-stats {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.filters {
  margin-bottom: 20px;
}

.notification-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 30px;
}

.notification-item {
  display: flex;
  align-items: flex-start;
  gap: 15px;
  padding: 15px;
  background: white;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.notification-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.notification-item.unread {
  border-left: 4px solid #409eff;
  background: #f8faff;
}

.notification-item.important {
  border-left: 4px solid #f56c6c;
}

.notification-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.system-icon { background: #e1f3d8; color: #67c23a; }
.homework-icon { background: #fdf6ec; color: #e6a23c; }
.project-icon { background: #e1f3d8; color: #67c23a; }
.topic-icon { background: #ecf5ff; color: #409eff; }
.document-icon { background: #f4f4f5; color: #909399; }
.user-icon { background: #e1f3d8; color: #67c23a; }
.warning-icon { background: #fef0f0; color: #f56c6c; }
.info-icon { background: #ecf5ff; color: #409eff; }

.notification-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.notification-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.notification-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.notification-meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.notification-time {
  font-size: 12px;
  color: #909399;
}

.notification-body {
  color: #606266;
  line-height: 1.5;
}

.notification-related {
  margin-top: 5px;
}

.notification-actions {
  display: flex;
  gap: 10px;
}

.empty-state {
  text-align: center;
  padding: 40px;
}

.pagination {
  display: flex;
  justify-content: center;
}

.announcements {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-top: 30px;
}

.announcements h2 {
  margin: 0 0 20px 0;
  color: #303133;
}

.announcement-item {
  padding: 15px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  margin-bottom: 15px;
}

.announcement-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.announcement-header h3 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.announcement-time {
  font-size: 12px;
  color: #909399;
}

.announcement-content {
  color: #606266;
  line-height: 1.6;
}
</style>
