import axios, { type AxiosInstance, type AxiosResponse } from 'axios'

// API基础配置
const API_BASE_URL = 'http://localhost:8080/api'

// 创建axios实例
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 这里可以添加认证token
    // const token = localStorage.getItem('token')
    // if (token) {
    //   config.headers.Authorization = `Bearer ${token}`
    // }
    console.log('API Request:', config.method?.toUpperCase(), config.url)
    return config
  },
  (error) => {
    console.error('Request Error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse) => {
    console.log('API Response:', response.status, response.config.url)
    return response
  },
  (error) => {
    console.error('Response Error:', error.response?.status, error.response?.data)
    
    // 处理常见错误
    if (error.response?.status === 401) {
      // 未授权，可以重定向到登录页
      console.error('Unauthorized access')
    } else if (error.response?.status === 500) {
      console.error('Server error')
    }
    
    return Promise.reject(error)
  }
)

// API接口定义
export interface User {
  id: number
  gitlab_id: number
  username: string
  email: string
  name: string
  avatar: string
  role: number
  last_sync_at: string
  is_active: boolean
}

export interface UserDashboard {
  user: User
  stats: {
    documents_count: number
    recent_activities: Array<{
      type: string
      document_id: number
      filename: string
      updated_at: string
    }>
    project_memberships: Array<{
      project_id: number
      project_name: string
      role: string
      web_url: string
    }>
  }
}

export interface Document {
  document_id: number
  editor_url: string
  message: string
}

export interface DocumentConfig {
  document: {
    fileType: string
    key: string
    title: string
    url: string
    permissions: {
      comment: boolean
      download: boolean
      edit: boolean
      fillForms: boolean
      modifyFilter: boolean
      print: boolean
      review: boolean
    }
  }
  editor: {
    callbackUrl: string
    lang: string
    mode: string
    user: {
      id: string
      name: string
    }
  }
  callbackUrl: string
  token: string
  type: string
  width: string
  height: string
  embedded: {
    saveUrl: string
    shareUrl: string
    toolbarDocked: string
  }
}

// API服务类
export class ApiService {
  
  // 健康检查
  static async healthCheck(): Promise<any> {
    const response = await api.get('/health')
    return response.data
  }

  // 用户相关API
  static async getCurrentUser(): Promise<User> {
    const response = await api.get('/users/me')
    return response.data.data
  }

  static async getActiveUsers(): Promise<{ data: User[], total: number }> {
    const response = await api.get('/users/active')
    return response.data
  }

  static async getUserDashboard(): Promise<UserDashboard> {
    const response = await api.get('/users/me/dashboard')
    return response.data.data
  }

  static async getUserById(id: number): Promise<User> {
    const response = await api.get(`/users/${id}`)
    return response.data.data
  }

  // 文档相关API
  static async createTestDocument(): Promise<Document> {
    const response = await api.get('/documents/test')
    return response.data
  }

  static async getDocumentConfig(id: number): Promise<DocumentConfig> {
    const response = await api.get(`/documents/${id}/config`)
    return response.data
  }

  static async uploadDocument(file: File, mode: string = 'edit'): Promise<Document> {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('mode', mode)
    
    const response = await api.post('/documents/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  }

  // 获取文档编辑器URL
  static getDocumentEditorUrl(id: number): string {
    return `${API_BASE_URL.replace('/api', '')}/api/documents/${id}/editor`
  }

  // 获取文档内容URL
  static getDocumentContentUrl(id: number): string {
    return `${API_BASE_URL.replace('/api', '')}/api/documents/${id}/content`
  }
}

export default api 