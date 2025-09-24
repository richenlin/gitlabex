interface NotificationMessage {
  type: string
  title: string
  content: string
  data?: any
  timestamp: number
}

interface WebSocketOptions {
  onMessage?: (message: NotificationMessage) => void
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: Event) => void
  reconnectInterval?: number
  maxReconnectAttempts?: number
}

export class WebSocketService {
  private ws: WebSocket | null = null
  private url: string
  private token: string
  private options: WebSocketOptions
  private reconnectAttempts = 0
  private reconnectTimer: number | null = null
  private isConnecting = false
  private shouldReconnect = true

  constructor(url: string, token: string, options: WebSocketOptions = {}) {
    this.url = url
    this.token = token
    this.options = {
      reconnectInterval: 5000,
      maxReconnectAttempts: 10,
      ...options
    }
  }

  connect(userId: string): void {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return
    }

    this.isConnecting = true
    const wsUrl = `${this.url}?user_id=${userId}&token=${this.token}`

    try {
      this.ws = new WebSocket(wsUrl)
      this.setupEventListeners()
    } catch (error) {
      console.error('WebSocket连接失败:', error)
      this.isConnecting = false
      this.scheduleReconnect(userId)
    }
  }

  private setupEventListeners(): void {
    if (!this.ws) return

    this.ws.onopen = () => {
      console.log('WebSocket连接已建立')
      this.isConnecting = false
      this.reconnectAttempts = 0
      
      // 发送心跳包
      this.startHeartbeat()
      
      this.options.onConnect?.()
    }

    this.ws.onmessage = (event) => {
      try {
        const message: NotificationMessage = JSON.parse(event.data)
        console.log('收到WebSocket消息:', message)
        
        // 处理心跳包响应
        if (message.type === 'pong') {
          return
        }
        
        this.options.onMessage?.(message)
      } catch (error) {
        console.error('解析WebSocket消息失败:', error)
      }
    }

    this.ws.onclose = (event) => {
      console.log('WebSocket连接已关闭:', event.code, event.reason)
      this.isConnecting = false
      this.stopHeartbeat()
      
      this.options.onDisconnect?.()
      
      if (this.shouldReconnect) {
        this.scheduleReconnect()
      }
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket错误:', error)
      this.options.onError?.(error)
    }
  }

  private scheduleReconnect(userId?: string): void {
    if (this.reconnectAttempts >= (this.options.maxReconnectAttempts || 10)) {
      console.error('达到最大重连次数，停止重连')
      return
    }

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }

    this.reconnectTimer = setTimeout(() => {
      console.log(`尝试第 ${this.reconnectAttempts + 1} 次重连...`)
      this.reconnectAttempts++
      
      if (userId) {
        this.connect(userId)
      }
    }, this.options.reconnectInterval || 5000)
  }

  private heartbeatTimer: number | null = null

  private startHeartbeat(): void {
    this.stopHeartbeat()
    
    this.heartbeatTimer = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping', timestamp: Date.now() }))
      }
    }, 30000) // 每30秒发送一次心跳
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  disconnect(): void {
    this.shouldReconnect = false
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    
    this.stopHeartbeat()
    
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  send(message: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket未连接，无法发送消息')
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

// 单例WebSocket服务
let wsService: WebSocketService | null = null

export function createWebSocketService(token: string, options: WebSocketOptions = {}): WebSocketService {
  const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/api/v1/ws/connect'
  
  if (wsService) {
    wsService.disconnect()
  }
  
  wsService = new WebSocketService(wsUrl, token, options)
  return wsService
}

export function getWebSocketService(): WebSocketService | null {
  return wsService
}

export function disconnectWebSocket(): void {
  if (wsService) {
    wsService.disconnect()
    wsService = null
  }
}
