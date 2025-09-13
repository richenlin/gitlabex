import { ref, computed } from 'vue'
import { ElNotification } from 'element-plus'
import { createWebSocketService, getWebSocketService, disconnectWebSocket } from '@/services/websocket'
import { useUserStore } from '@/stores/user'

interface NotificationItem {
  id: string
  type: string
  title: string
  content: string
  data?: any
  timestamp: number
  read: boolean
}

interface NotificationMessage {
  type: string
  title: string
  content: string
  data?: any
  timestamp: number
}

// 全局通知状态
const notifications = ref<NotificationItem[]>([])
const isConnected = ref(false)

export function useNotifications() {
  const userStore = useUserStore()

  // 计算属性
  const unreadCount = computed(() => 
    notifications.value.filter(n => !n.read).length
  )

  const recentNotifications = computed(() =>
    notifications.value
      .sort((a, b) => b.timestamp - a.timestamp)
      .slice(0, 10)
  )

  // 连接WebSocket
  const connect = () => {
    if (!userStore.token || !userStore.user?.id) {
      console.warn('用户未登录，无法连接WebSocket')
      return
    }

    const wsService = createWebSocketService(userStore.token, {
      onConnect: () => {
        console.log('WebSocket已连接')
        isConnected.value = true
      },
      
      onDisconnect: () => {
        console.log('WebSocket已断开')
        isConnected.value = false
      },
      
      onError: (error) => {
        console.error('WebSocket错误:', error)
        isConnected.value = false
      },
      
      onMessage: (message: NotificationMessage) => {
        handleNotificationMessage(message)
      }
    })

    wsService.connect(userStore.user.id.toString())
  }

  // 断开连接
  const disconnect = () => {
    disconnectWebSocket()
    isConnected.value = false
  }

  // 处理通知消息
  const handleNotificationMessage = (message: NotificationMessage) => {
    const notification: NotificationItem = {
      id: generateNotificationId(),
      type: message.type,
      title: message.title,
      content: message.content,
      data: message.data,
      timestamp: message.timestamp * 1000, // 转换为毫秒
      read: false
    }

    // 添加到通知列表
    notifications.value.unshift(notification)

    // 限制通知数量
    if (notifications.value.length > 100) {
      notifications.value = notifications.value.slice(0, 100)
    }

    // 显示桌面通知
    showDesktopNotification(notification)
  }

  // 显示桌面通知
  const showDesktopNotification = (notification: NotificationItem) => {
    // 根据通知类型设置不同的样式
    let notificationType: 'success' | 'warning' | 'info' | 'error' = 'info'
    
    switch (notification.type) {
      case 'system':
        notificationType = 'info'
        break
      case 'project':
        notificationType = 'success'
        break
      case 'homework':
        notificationType = 'warning'
        break
      case 'error':
        notificationType = 'error'
        break
      default:
        notificationType = 'info'
    }

    ElNotification({
      title: notification.title,
      message: notification.content,
      type: notificationType,
      duration: 5000,
      position: 'top-right',
      onClick: () => {
        markAsRead(notification.id)
        handleNotificationClick(notification)
      }
    })
  }

  // 处理通知点击
  const handleNotificationClick = (notification: NotificationItem) => {
    // 根据通知类型和数据进行路由跳转
    if (notification.data) {
      const router = useRouter()
      
      switch (notification.type) {
        case 'project':
          if (notification.data.project_id) {
            router.push(`/projects/${notification.data.project_id}`)
          }
          break
        case 'homework':
          if (notification.data.homework_id) {
            router.push(`/homework/${notification.data.homework_id}`)
          }
          break
        default:
          // 其他类型的通知处理
          break
      }
    }
  }

  // 标记为已读
  const markAsRead = (notificationId: string) => {
    const notification = notifications.value.find(n => n.id === notificationId)
    if (notification) {
      notification.read = true
    }
  }

  // 标记所有为已读
  const markAllAsRead = () => {
    notifications.value.forEach(n => n.read = true)
  }

  // 清除通知
  const removeNotification = (notificationId: string) => {
    const index = notifications.value.findIndex(n => n.id === notificationId)
    if (index !== -1) {
      notifications.value.splice(index, 1)
    }
  }

  // 清除所有通知
  const clearAllNotifications = () => {
    notifications.value = []
  }

  // 生成通知ID
  const generateNotificationId = () => {
    return Date.now().toString() + Math.random().toString(36).substr(2, 9)
  }

  // 发送消息到WebSocket
  const sendMessage = (message: any) => {
    const wsService = getWebSocketService()
    if (wsService) {
      wsService.send(message)
    }
  }

  // 请求桌面通知权限
  const requestNotificationPermission = async () => {
    if ('Notification' in window && Notification.permission === 'default') {
      await Notification.requestPermission()
    }
  }

  return {
    // 状态
    notifications: notifications,
    isConnected,
    unreadCount,
    recentNotifications,

    // 方法
    connect,
    disconnect,
    markAsRead,
    markAllAsRead,
    removeNotification,
    clearAllNotifications,
    sendMessage,
    requestNotificationPermission
  }
}

// 在路由守卫中使用
export function setupNotifications() {
  const { connect, requestNotificationPermission } = useNotifications()
  const userStore = useUserStore()

  // 用户登录后自动连接
  if (userStore.isLoggedIn) {
    requestNotificationPermission()
    connect()
  }
}

// 导入useRouter（需要在组件内使用）
function useRouter() {
  // 这个函数需要在组件内部调用，这里只是占位符
  // 实际使用时需要从vue-router导入
  return {
    push: (path: string) => {
      console.log('导航到:', path)
    }
  }
}
