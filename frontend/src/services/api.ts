import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('authToken')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    if (error.response) {
      switch (error.response.status) {
        case 401:
          // 未授权，清除token并跳转到登录页
          localStorage.removeItem('authToken')
          window.location.href = '/login'
          break
        case 403:
          // 权限不足
          console.error('权限不足')
          break
        case 404:
          // 资源不存在
          console.error('请求的资源不存在')
          break
        case 500:
          // 服务器错误
          console.error('服务器错误')
          break
        default:
          console.error('请求失败:', error.response.data)
      }
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
    const response = await api.get('/api/health')
    return response
  }

  // 用户相关API
  static async getCurrentUser(): Promise<User> {
    const response = await api.get('/api/users/current')
    // 处理后端返回的数据结构 {"data": {"user": {...}}}
    return response.data.user
  }

  static async getActiveUsers(): Promise<{ data: User[], total: number }> {
    const response = await api.get('/api/users/active')
    return response.data
  }

  static async getUserDashboard(): Promise<UserDashboard> {
    const response = await api.get('/api/users/current')
    return response.data
  }

  static async getUserById(id: number): Promise<User> {
    const response = await api.get(`/api/users/${id}`)
    return response.data.data
  }

  static async updateUserProfile(userData: Partial<User>): Promise<User> {
    const response = await api.put('/api/users/me/profile', userData)
    return response.data
  }

  // 获取保存的测试用户资料


  // 认证相关API
  static async getGitLabOAuthUrl(): Promise<{ url: string }> {
    return await api.get('/api/auth/gitlab')
  }

  static async handleOAuthCallback(code: string, state?: string): Promise<{ token: string, user: User }> {
    const response = await api.post('/api/auth/gitlab/callback', { code, state })
    return response.data
  }

  static async logout(): Promise<void> {
    await api.post('/api/auth/logout')
  }

  // 文档相关API
  static async createTestDocument(): Promise<Document> {
    const response = await api.get('/api/documents/test')
    return response.data
  }

  static async getDocumentConfig(id: number): Promise<DocumentConfig> {
    const response = await api.get(`/api/documents/${id}/config`)
    return response.data
  }

  static async uploadDocument(file: File, mode: string = 'edit'): Promise<Document> {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('mode', mode)
    
    const response = await api.post('/api/documents/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  }

  // 获取文档编辑器URL
  static getDocumentEditorUrl(id: number): string {
    return `/api/documents/${id}/editor`
  }

  // 获取文档内容URL
  static getDocumentContentUrl(id: number): string {
    return `/api/documents/${id}/content`
  }

  // 学习进度跟踪相关API
  static async getLearningProgressUsers(): Promise<any> {
    const response = await api.get('/api/learning-progress/users')
    return response.data.data
  }

  static async getLearningProgress(userId: number): Promise<any> {
    const response = await api.get(`/api/learning-progress/user/${userId}`)
    return response.data.data
  }

  // 通知系统相关API
  static async getNotifications(params?: { type?: string; read?: string }): Promise<any> {
    const response = await api.get('/api/notifications', { params })
    return response.data
  }

  static async markNotificationAsRead(notificationId: number): Promise<any> {
    const response = await api.put(`/api/notifications/${notificationId}/read`)
    return response.data
  }

  static async markAllNotificationsAsRead(): Promise<any> {
    const response = await api.put('/api/notifications/read-all')
    return response.data
  }

  static async deleteNotification(notificationId: number): Promise<any> {
    const response = await api.delete(`/api/notifications/${notificationId}`)
    return response.data
  }

  static async deleteNotifications(ids: number[]): Promise<any> {
    const response = await api.delete('/api/notifications', { data: { ids } })
    return response.data
  }

  // 教育报表相关API
  static async getEducationReports(params?: { time_range?: string; class?: string }): Promise<any> {
    const response = await api.get('/api/education-reports', { params })
    return response.data.data
  }

  static async getEducationReportClasses(): Promise<any> {
    const response = await api.get('/api/education-reports/classes')
    return response.data.data
  }

  static async exportEducationReport(params?: { format?: string; time_range?: string; class?: string }): Promise<any> {
    const response = await api.post('/api/education-reports/export', null, { params })
    return response.data
  }

  // 分析统计API
  static async getAnalyticsOverview(): Promise<any> {
    const response = await api.get('/api/analytics/overview')
    return response.data
  }

  static async getAnalyticsProjectStats(): Promise<any> {
    const response = await api.get('/api/analytics/project-stats')
    return response.data
  }

  static async getStudentStats(): Promise<any> {
    const response = await api.get('/api/analytics/student-stats')
    return response.data
  }

  static async getAssignmentStats(): Promise<any> {
    const response = await api.get('/api/analytics/assignment-stats')
    return response.data
  }

  static async getSubmissionTrend(params?: { start_date?: Date; end_date?: Date }): Promise<any> {
    const response = await api.get('/api/analytics/submission-trend', { params })
    return response.data
  }

  static async getProjectDistribution(): Promise<any> {
    const response = await api.get('/api/analytics/project-distribution')
    return response.data
  }

  static async getGradeDistribution(): Promise<any> {
    const response = await api.get('/api/analytics/grade-distribution')
    return response.data
  }

  static async getActivityStats(): Promise<any> {
    const response = await api.get('/api/analytics/activity-stats')
    return response.data
  }

  static async getDashboardStats(): Promise<any> {
    const response = await api.get('/api/analytics/dashboard-stats')
    return response.data
  }

  static async getRecentActivities(params?: { limit?: number }): Promise<any> {
    const response = await api.get('/api/analytics/recent-activities', { params })
    return response.data
  }

  // 话题讨论相关API
  static async getDiscussions(params?: { 
    project_id?: number; 
    page?: number; 
    page_size?: number; 
    category?: string; 
    status?: string 
  }): Promise<any> {
    const response = await api.get('/api/discussions', { params })
    return response.data
  }

  static async getDiscussionDetail(id: number): Promise<any> {
    const response = await api.get(`/api/discussions/${id}`)
    return response.data
  }

  static async createDiscussion(data: {
    title: string;
    content: string;
    project_id: number;
    category?: string;
    tags?: string;
    is_public?: boolean;
  }): Promise<any> {
    const response = await api.post('/api/discussions', data)
    return response.data
  }

  static async updateDiscussion(id: number, data: {
    title?: string;
    content?: string;
    category?: string;
    tags?: string;
    is_public?: boolean;
  }): Promise<any> {
    const response = await api.put(`/api/discussions/${id}`, data)
    return response.data
  }

  static async deleteDiscussion(id: number): Promise<any> {
    const response = await api.delete(`/api/discussions/${id}`)
    return response.data
  }

  static async createReply(discussionId: number, data: {
    content: string;
    parent_reply_id?: number;
  }): Promise<any> {
    const response = await api.post(`/api/discussions/${discussionId}/replies`, data)
    return response.data
  }

  static async likeDiscussion(id: number): Promise<any> {
    const response = await api.post(`/api/discussions/${id}/like`)
    return response.data
  }

  static async unlikeDiscussion(id: number): Promise<any> {
    const response = await api.delete(`/api/discussions/${id}/like`)
    return response.data
  }

  static async pinDiscussion(id: number): Promise<any> {
    const response = await api.post(`/api/discussions/${id}/pin`)
    return response.data
  }

  static async getDiscussionCategories(): Promise<any> {
    const response = await api.get('/api/discussions/categories')
    return response.data
  }

  static async syncDiscussionsFromGitLab(projectId: number): Promise<any> {
    const response = await api.post(`/api/discussions/sync/${projectId}`)
    return response.data
  }

  static async getProjects(params?: { 
    class_id?: number; 
    page?: number; 
    page_size?: number; 
    status?: string; 
    type?: string 
  }): Promise<any> {
    const response = await api.get('/api/projects', { params })
    // 响应拦截器返回response.data，如果后端返回{data: [...], total: number}
    // 那么这里的response就是{data: [...], total: number}
    // 如果后端直接返回数组，那么这里的response就是数组
    console.log('ApiService.getProjects response:', response)
    return response
  }

  static async getProject(id: number): Promise<any> {
    const response = await api.get(`/api/projects/${id}`)
    return response.data
  }

  static async createProject(data: {
    name: string;
    description: string;
    type: string;
    class_id?: number;
    start_date: string;
    end_date: string;
    max_members?: number;
    wiki_enabled?: boolean;
    issues_enabled?: boolean;
    mr_enabled?: boolean;
  }): Promise<any> {
    const response = await api.post('/api/projects', data)
    return response.data
  }

  static async updateProject(id: number, data: {
    name?: string;
    description?: string;
    status?: string;
    start_date?: string;
    end_date?: string;
  }): Promise<any> {
    const response = await api.put(`/api/projects/${id}`, data)
    return response.data
  }

  static async deleteProject(id: number): Promise<any> {
    const response = await api.delete(`/api/projects/${id}`)
    return response.data
  }

  static async joinProject(code: string): Promise<any> {
    const response = await api.post('/api/projects/join', { code })
    return response.data
  }

  static async getProjectMembers(id: number): Promise<any> {
    const response = await api.get(`/api/projects/${id}/members`)
    return response.data
  }

  static async addProjectMember(id: number, userId: number): Promise<any> {
    const response = await api.post(`/api/projects/${id}/members`, { user_id: userId })
    return response.data
  }

  static async removeProjectMember(id: number, userId: number): Promise<any> {
    const response = await api.delete(`/api/projects/${id}/members/${userId}`)
    return response.data
  }

  static async getProjectStats(id: number): Promise<any> {
    const response = await api.get(`/api/projects/${id}/stats`)
    return response.data
  }

  static async getProjectAssignments(id: number): Promise<any> {
    const response = await api.get(`/api/projects/${id}/assignments`)
    return response.data
  }
}

export default api 