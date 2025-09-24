import { ref, computed } from 'vue'
import { ElNotification } from 'element-plus'
import { authService } from '@/services/api'

interface GitLabNotification {
  id: string
  type: string
  title?: string
  body?: string
  message?: string
  subject?: {
    type: string
    id: number
    title?: string
    name?: string
  }
  project?: {
    id: number
    name: string
    web_url: string
  }
  created_at: string
  read: boolean
  timestamp?: number
}

// 全局通知状态
const notifications = ref<GitLabNotification[]>([])
const loading = ref(false)
const isConnected = ref(true) // GitLab API始终可用

export function useNotifications() {
  // 计算属性
  const unreadCount = computed(() => 
    notifications.value.filter(n => !n.read).length
  )

  const recentNotifications = computed(() =>
    notifications.value
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
      .slice(0, 10)
  )

  // 获取通知列表
  const fetchNotifications = async (page = 1, perPage = 20) => {
    try {
      loading.value = true
      console.log('获取GitLab通知列表...', { page, perPage })
      
      const response = await authService.getNotifications({
        page,
        per_page: perPage
      })
      
      console.log('GitLab通知API响应:', response)
      
      const data = response?.data || response
      notifications.value = data?.notifications || []
      
      console.log('GitLab通知列表:', notifications.value)
      return notifications.value
    } catch (error) {
      console.error('获取GitLab通知列表失败:', error)
      notifications.value = []
      return []
    } finally {
      loading.value = false
    }
  }

  // 标记单个通知为已读
  const markAsRead = async (notificationId: string) => {
    try {
      console.log('标记GitLab通知为已读:', notificationId)
      
      await authService.markNotificationAsRead(notificationId)
      
      // 更新本地状态
      const notification = notifications.value.find(n => n.id === notificationId)
      if (notification) {
        notification.read = true
      }
      
      return true
    } catch (error) {
      console.error('标记GitLab通知为已读失败:', error)
      return false
    }
  }

  // 标记所有通知为已读
  const markAllAsRead = async () => {
    try {
      console.log('标记所有GitLab通知为已读')
      
      await authService.markAllNotificationsAsRead()
      
      // 更新本地状态
      notifications.value.forEach(n => {
        n.read = true
      })
      
      return true
    } catch (error) {
      console.error('标记所有GitLab通知为已读失败:', error)
      return false
    }
  }

  // 清除通知（本地操作）
  const removeNotification = (notificationId: string) => {
    const index = notifications.value.findIndex(n => n.id === notificationId)
    if (index !== -1) {
      notifications.value.splice(index, 1)
    }
  }

  // 清除所有通知（本地操作）
  const clearAllNotifications = () => {
    notifications.value = []
  }

  // 处理GitLab通知点击
  const handleNotificationClick = (notification: GitLabNotification) => {
    // 根据GitLab通知类型进行路由跳转
    if (notification.subject && notification.project) {
      switch (notification.subject.type) {
        case 'MergeRequest':
          if (notification.subject.id) {
            // 跳转到合并请求页面
            window.open(`${notification.project.web_url}/-/merge_requests/${notification.subject.id}`, '_blank')
          }
          break
        case 'Issue':
          if (notification.subject.id) {
            // 跳转到问题页面
            window.open(`${notification.project.web_url}/-/issues/${notification.subject.id}`, '_blank')
          }
          break
        case 'Commit':
          if (notification.subject.id) {
            // 跳转到提交页面
            window.open(`${notification.project.web_url}/-/commit/${notification.subject.id}`, '_blank')
          }
          break
        case 'Project':
          if (notification.project.web_url) {
            // 跳转到项目页面
            window.open(notification.project.web_url, '_blank')
          }
          break
        default:
          break
      }
    }
  }

  // 显示桌面通知
  const showDesktopNotification = (notification: GitLabNotification) => {
    // 根据通知类型设置不同的样式
    let notificationType: 'success' | 'warning' | 'info' | 'error' = 'info'
    
    switch (notification.subject?.type) {
      case 'MergeRequest':
        notificationType = 'success'
        break
      case 'Issue':
        notificationType = 'warning'
        break
      case 'Commit':
        notificationType = 'info'
        break
      case 'Project':
        notificationType = 'info'
        break
      default:
        notificationType = 'info'
    }

    const title = notification.title || notification.subject?.title || 'GitLab 通知'
    const content = notification.body || notification.message || notification.subject?.name || 'GitLab 通知'

    ElNotification({
      title,
      message: content,
      type: notificationType,
      duration: 5000,
      position: 'top-right',
      onClick: () => {
        markAsRead(notification.id)
        handleNotificationClick(notification)
      }
    })
  }

  // 获取通知图标类型
  const getNotificationIconType = (notification: GitLabNotification) => {
    return notification.subject?.type || 'system'
  }

  // 获取通知内容
  const getNotificationContent = (notification: GitLabNotification) => {
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

  // 连接状态（GitLab API始终可用）
  const connect = () => {
    isConnected.value = true
    console.log('GitLab通知API连接状态：已连接')
  }

  // 断开连接（GitLab API始终可用）
  const disconnect = () => {
    isConnected.value = false
    console.log('GitLab通知API连接状态：已断开')
  }

  // 发送消息（GitLab通知由GitLab系统自动生成）
  const sendMessage = (message: any) => {
    console.log('GitLab通知由系统自动生成，不支持手动发送:', message)
  }

  // 请求桌面通知权限
  const requestNotificationPermission = async () => {
    if ('Notification' in window && Notification.permission === 'default') {
      await Notification.requestPermission()
    }
  }

  return {
    // 状态
    notifications,
    loading,
    isConnected,
    unreadCount,
    recentNotifications,

    // 方法
    fetchNotifications,
    markAsRead,
    markAllAsRead,
    removeNotification,
    clearAllNotifications,
    handleNotificationClick,
    showDesktopNotification,
    getNotificationIconType,
    getNotificationContent,
    formatTime,
    connect,
    disconnect,
    sendMessage,
    requestNotificationPermission
  }
}

// 在路由守卫中使用
export function setupNotifications() {
  const { requestNotificationPermission, fetchNotifications } = useNotifications()

  // 用户登录后请求通知权限并获取通知
  requestNotificationPermission()
  
  // 初始获取通知
  fetchNotifications()
}