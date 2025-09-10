import axios from 'axios'
import type { AxiosInstance, AxiosResponse, AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import type { 
  User, 
  Scene, 
  Topic, 
  Document, 
  Homework, 
  HomeworkSubmission, 
  Notification,
  ApiResponse,
  PaginatedResponse
} from '@/types'

// 创建 axios 实例
const api: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse<any>) => {
    return response.data
  },
  (error: AxiosError<ApiResponse<any>>) => {
    const message = error.response?.data?.message || '请求失败'
    
    if (error.response?.status === 401) {
      const userStore = useUserStore()
      userStore.logout()
      ElMessage.error('登录已过期，请重新登录')
    } else if (error.response?.status === 403) {
      ElMessage.error('权限不足')
    } else if (error.response?.status === 404) {
      // 404错误通常表示资源不存在，不需要显示错误提示
      console.warn('Resource not found:', error.config?.url)
    } else if (error.response && error.response.status >= 500) {
      // 只有服务器错误才显示错误提示
      ElMessage.error(message)
    } else {
      // 其他客户端错误，记录日志但不显示提示
      console.warn('API request failed:', error.response?.status, message)
    }
    
    return Promise.reject(error)
  }
)

// 认证相关 API
export const authService = {
  // GitLab OAuth 认证
  gitlabAuth: () =>
    api.get('/auth/gitlab'),
  
  gitlabCallback: (code: string, state: string) =>
    api.get('/auth/gitlab/callback', { params: { code, state } }),
  
  logout: () =>
    api.post('/auth/logout'),
  
  getCurrentUser: () =>
    api.get('/users/me'),
  
  updateProfile: (data: Partial<User>) =>
    api.put('/users/me', data),
  
  refreshToken: () =>
    api.post('/auth/refresh')
}

// 研究课题相关 API
export const researchService = {
  getProjects: (params?: {
    page?: number
    pageSize?: number
    search?: string
    visibility?: string
    ownerId?: string
  }) =>
    api.get('/research-projects', { params }),
  
  getProject: (id: string) =>
    api.get(`/research-projects/${id}`),
  
  createProject: (data: Partial<Scene>) =>
    api.post('/research-projects', data),
  
  updateProject: (id: string, data: Partial<Scene>) =>
    api.put(`/research-projects/${id}`, data),
  
  deleteProject: (id: string) =>
    api.delete(`/research-projects/${id}`),
  
  getMembers: (id: string) =>
    api.get(`/research-projects/${id}/members`),
  
  addMember: (id: string, userId: string) =>
    api.post(`/research-projects/${id}/members`, { userId }),
  
  removeMember: (id: string, userId: string) =>
    api.delete(`/research-projects/${id}/members/${userId}`),
  
  // Issues (话题) 管理
  getIssues: (id: string) =>
    api.get(`/research-projects/${id}/issues`),
  
  createIssue: (id: string, data: Partial<Topic>) =>
    api.post(`/research-projects/${id}/issues`, data),
  
  getIssue: (projectId: string, issueId: string) =>
    api.get(`/research-projects/${projectId}/issues/${issueId}`),
  
  getDiscussions: (projectId: string, issueId: string) =>
    api.get(`/research-projects/${projectId}/issues/${issueId}/discussions`),
  
  createDiscussion: (projectId: string, issueId: string, body: string) =>
    api.post(`/research-projects/${projectId}/issues/${issueId}/discussions`, { body })
}

// 话题相关 API  
export const topicService = {
  getTopics: (params?: {
    page?: number
    pageSize?: number
    projectId?: string
    authorId?: string
    search?: string
    labels?: string[]
  }) =>
    api.get('/topics', { params }),
  
  getTopic: (id: string) =>
    api.get(`/topics/${id}`),
  
  createTopic: (data: Partial<Topic>) =>
    api.post('/topics', data),
  
  updateTopic: (id: string, data: Partial<Topic>) =>
    api.put(`/topics/${id}`, data),
  
  deleteTopic: (id: string) =>
    api.delete(`/topics/${id}`),
  
  likeTopic: (id: string) =>
    api.post(`/topics/${id}/like`),
  
  unlikeTopic: (id: string) =>
    api.delete(`/topics/${id}/like`),
  
  createComment: (id: string, content: string, parentId?: string) =>
    api.post(`/topics/${id}/comments`, { content, parentId })
}



// 文档相关 API
export const documentService = {
  getDocuments: (params?: {
    page?: number
    pageSize?: number
    projectId?: string
    homeworkId?: string
    category?: string
    search?: string
  }) =>
    api.get('/documents', { params }),
  
  getDocument: (id: string) =>
    api.get(`/documents/${id}`),
  
  createDocument: (data: Partial<Document>) =>
    api.post('/documents', data),
  
  updateDocument: (id: string, data: Partial<Document>) =>
    api.put(`/documents/${id}`, data),
  
  deleteDocument: (id: string) =>
    api.delete(`/documents/${id}`),
  
  getCategories: () =>
    api.get('/documents/categories'),
  
  searchDocuments: (query: string) =>
    api.get('/documents/search', { params: { q: query } }),
  
  getStats: () =>
    api.get('/documents/stats'),
  
  syncDocuments: (projectId: string) =>
    api.post(`/documents/sync/${projectId}`),
  
  scanProjectDocuments: (projectId: string) =>
    api.post(`/documents/scan/${projectId}`),
  
  submitEditRequest: (documentId: string, data: {
    proposed_changes: any
    reason: string
  }) =>
    api.post(`/documents/${documentId}/edit-request`, data),
  
  getEditRequests: (params?: {
    document_id?: string
    status?: string
  }) =>
    api.get('/documents/edit-requests', { params }),
  
  reviewEditRequest: (requestId: string, data: {
    approved: boolean
    comments: string
  }) =>
    api.put(`/documents/edit-requests/${requestId}/review`, data),
  
  getMyEditRequests: (documentId: string) =>
    api.get(`/documents/${documentId}/my-edit-requests`)
}

// 作业相关 API
export const homeworkService = {
  getHomeworks: (params?: {
    page?: number
    pageSize?: number
    projectId?: string
    authorId?: string
    status?: string
  }) =>
    api.get('/homework', { params }),
  
  getHomework: (id: string) =>
    api.get(`/homework/${id}`),
  
  getHomeworkById: (id: string) =>
    api.get(`/homework/${id}`),
  
  createHomework: (data: Partial<Homework>) =>
    api.post('/homework', data),
  
  updateHomework: (id: string, data: Partial<Homework>) =>
    api.put(`/homework/${id}`, data),
  
  deleteHomework: (id: string) =>
    api.delete(`/homework/${id}`),
  
  getSubmissions: (homeworkId: string) =>
    api.get(`/homework/${homeworkId}/submissions`),
  
  submitHomework: (data: {
    homework_id: string
    content?: string
    notes?: string
    files?: string[]
    branch_name?: string
  }) =>
    api.post(`/homework/${data.homework_id}/submit`, data),
  
  getMySubmission: (homeworkId: string) =>
    api.get(`/homework/${homeworkId}/my-submission`),
  
  createStudentBranch: (homeworkId: string) =>
    api.post(`/homework/${homeworkId}/create-branch`),
  
  getStudentBranchInfo: (homeworkId: string) =>
    api.get(`/homework/${homeworkId}/branch-info`),
  
  gradeSubmission: (submissionId: string, grade: number, feedback?: string) =>
    api.put(`/submissions/${submissionId}/grade`, { grade, feedback }),
  
  getSubmissionsByProject: (projectId: string) =>
    api.get(`/research-projects/${projectId}/homework`)
}

// 通知相关 API
export const notificationService = {
  getNotifications: (params?: {
    page?: number
    pageSize?: number
    isRead?: boolean
  }) =>
    api.get('/notifications', { params }),
  
  getNotification: (id: string) =>
    api.get(`/notifications/${id}`),
  
  markAsRead: (id: string) =>
    api.put(`/notifications/${id}/read`),
  
  markAllAsRead: () =>
    api.put('/notifications/read-all'),
  
  deleteNotification: (id: string) =>
    api.delete(`/notifications/${id}`),
  
  clearAll: () =>
    api.delete('/notifications'),
  
  // 公告相关
  getAnnouncements: () =>
    api.get('/announcements'),
  
  createAnnouncement: (data: any) =>
    api.post('/announcements', data)
}

// GitLab 相关 API
export const gitlabService = {
  getProjects: () =>
    api.get('/gitlab/projects'),
  
  getProject: (id: string) =>
    api.get(`/gitlab/projects/${id}`),
  
  getBranches: (projectId: string) =>
    api.get(`/gitlab/projects/${projectId}/branches`),
  
  getCommits: (projectId: string, branch?: string) =>
    api.get(`/gitlab/projects/${projectId}/commits`, { params: { branch } }),
  
  getFileContent: (projectId: string, path: string, branch?: string) =>
    api.get(`/gitlab/projects/${projectId}/files/${encodeURIComponent(path)}`, {
      params: { branch }
    }),
  
  updateFile: (projectId: string, path: string, content: string, branch?: string) =>
    api.put(`/gitlab/projects/${projectId}/files/${encodeURIComponent(path)}`, {
      content,
      branch
    }),
  
  createMergeRequest: (projectId: string, data: {
    title: string
    description: string
    sourceBranch: string
    targetBranch: string
  }) =>
    api.post(`/gitlab/projects/${projectId}/merge-requests`, data),
  
  // Issues 管理 (直接GitLab API调用)
  createIssue: (projectId: string, data: {
    title: string
    description: string
    labels?: string[]
    assigneeId?: number
  }) =>
    api.post(`/gitlab/projects/${projectId}/issues`, data)
}

export default api
