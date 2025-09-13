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
    console.log('API请求拦截器 - URL:', config.url)
    console.log('API请求拦截器 - Token:', userStore.token)
    console.log('API请求拦截器 - User:', userStore.user)
    
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
      console.log('API请求拦截器 - 已添加Authorization头')
    } else {
      console.warn('API请求拦截器 - 没有token，无法添加Authorization头')
    }
    
    console.log('API请求拦截器 - 最终headers:', config.headers)
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// 获取错误消息的辅助函数
const getErrorMessage = (error: AxiosError<any>): string => {
  // 优先使用后端返回的错误消息
  if (error.response?.data) {
    // 尝试不同的错误消息字段
    return error.response.data.message || 
           error.response.data.error || 
           error.response.data.msg ||
           error.response.data.detail ||
           `请求失败 (${error.response.status})`
  }
  
  // 网络错误或其他错误
  if (error.message) {
    return error.message
  }
  
  return '请求失败'
}

// 响应拦截器
api.interceptors.response.use(
  (response: AxiosResponse<any>) => {
    // 对于blob响应（下载请求），返回完整的response对象
    if (response.config.responseType === 'blob') {
      return response
    }
    return response.data
  },
  (error: AxiosError<ApiResponse<any>>) => {
    const message = getErrorMessage(error)
    
    if (error.response?.status === 401) {
      const userStore = useUserStore()
      const url = error.config?.url || ''
      
      // 只有在用户已登录的情况下才自动退出登录
      if (userStore.isLoggedIn) {
        // 如果是GitLab相关的API调用失败，可能是GitLab token问题，不要自动退出登录
        if (url.includes('/gitlab/')) {
          console.warn('GitLab API authentication failed, but keeping user logged in:', url)
          // 不退出登录，只记录警告，但显示具体错误信息
          ElMessage.error(message)
        } else {
          // 其他API的401错误，说明JWT token真的过期了，需要退出登录
          userStore.logout()
          ElMessage.error('登录已过期，请重新登录')
        }
      } else {
        // 用户未登录时，显示具体的认证错误信息
        ElMessage.error(message)
      }
    } else if (error.response?.status === 403) {
      // 显示后端返回的具体权限错误信息
      ElMessage.error(message)
    } else if (error.response?.status === 404) {
      // 404错误记录日志，但不自动显示提示（由组件决定是否显示）
      console.warn('Resource not found:', error.config?.url, message)
    } else if (error.response?.status === 400) {
      // 400错误通常是参数错误，显示具体错误信息
      ElMessage.error(message)
    } else if (error.response && error.response.status >= 500) {
      // 服务器错误显示具体错误信息
      ElMessage.error(message)
    } else if (error.response) {
      // 其他HTTP错误，显示具体错误信息
      ElMessage.error(message)
    } else {
      // 网络错误等
      ElMessage.error(message)
    }
    
    return Promise.reject(error)
  }
)

// 认证相关 API
export const authService = {
  // 获取GitLab OAuth授权URL
  getGitLabAuthUrl: () =>
    api.get('/auth/gitlab'),
  
  // GitLab OAuth回调处理
  gitLabCallback: (code: string, state: string) =>
    api.get('/auth/gitlab/callback', { params: { code, state } }),
  
  logout: () =>
    api.post('/auth/logout'),
  
  getCurrentUser: () =>
    api.get('/users/me'),
  
  updateProfile: (data: Partial<User>) =>
    api.put('/users/me', data),
  
  getUserStats: () =>
    api.get('/users/me/stats'),
  
  // SSH密钥管理
  getSSHKeys: () =>
    api.get('/users/me/ssh-keys'),
  
  addSSHKey: (data: { title: string; key: string }) =>
    api.post('/users/me/ssh-keys', data),
  
  deleteSSHKey: (id: number) =>
    api.delete(`/users/me/ssh-keys/${id}`),
  
  // 密码管理
  changePassword: (data: { currentPassword: string; newPassword: string }) =>
    api.put('/users/me/password', data),
  
  // 通知管理
  getNotifications: (params?: { page?: number; per_page?: number }) =>
    api.get('/users/me/notifications', { params }),
  
  markNotificationAsRead: (id: string) =>
    api.post(`/users/me/notifications/${id}/read`),
  
  markAllNotificationsAsRead: () =>
    api.post('/users/me/notifications/read-all'),
  
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
  
  getHotProjects: (limit?: number) =>
    api.get('/research-projects/hot', { params: { limit } }),
  
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
    limit?: number
    projectId?: string
    project_id?: string
    authorId?: string
    search?: string
    labels?: string[]
  }) =>
    api.get('/topics', { params }),
  
  getHotTopics: (limit?: number) =>
    api.get('/topics/hot', { params: { limit } }),
  
  getTopic: (id: string, projectId: string) =>
    api.get(`/topics/${id}?project_id=${projectId}`),
  
  createTopic: (data: Partial<Topic>) =>
    api.post('/topics', data),
  
  updateTopic: (id: string, data: Partial<Topic>) =>
    api.put(`/topics/${id}`, data),
  
  deleteTopic: (id: string) =>
    api.delete(`/topics/${id}`),
  
  likeTopic: (id: string, projectId: string) =>
    api.post(`/topics/${id}/like?project_id=${projectId}`),
  
  unlikeTopic: (id: string, projectId: string) =>
    api.delete(`/topics/${id}/like?project_id=${projectId}`),
  
  dislikeTopic: (id: string, projectId: string) =>
    api.post(`/topics/${id}/dislike?project_id=${projectId}`),
  
  undislikeTopic: (id: string, projectId: string) =>
    api.delete(`/topics/${id}/dislike?project_id=${projectId}`),
  
  createComment: (id: string, content: string, projectId: string, parentId?: string) =>
    api.post(`/topics/${id}/comments?project_id=${projectId}`, { content, parentId })
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
    status?: string
    uploaderId?: string
  }) =>
    api.get('/documents', { params }),
  
  getDocument: (id: string) =>
    api.get(`/documents/${id}`),
  
  createDocument: (formData: FormData) =>
    api.post('/documents', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    }),
  
  createStandaloneDocument: (formData: FormData) =>
    api.post('/documents/standalone', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    }),
  
  updateDocument: (id: string, data: Partial<Document>) =>
    api.put(`/documents/${id}/with-permission-check`, data),
  
  deleteDocument: (id: string) =>
    api.delete(`/documents/${id}`),
  
  downloadDocument: (id: string) =>
    api.get(`/documents/${id}/download`, { responseType: 'blob' }),
  
  getDocumentEditHistory: (id: string) =>
    api.get(`/documents/${id}/edit-history`),
  
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
  
  gradeHomework: (submissionId: string, data: { grade: number; feedback: string }) =>
    api.post(`/homework/submissions/${submissionId}/grade`, data),
  
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
  
  getSubmissionViewURL: (submissionId: string) =>
    api.get(`/homework/submissions/${submissionId}/view-url`),
  
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
  getConfig: () =>
    api.get('/gitlab/config'),
  
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
  
  getRepositoryTree: (projectId: string, path?: string) =>
    api.get(`/gitlab/projects/${projectId}/tree`, { params: { path } }),
  
  // Issues 管理 (直接GitLab API调用)
  createIssue: (projectId: string, data: {
    title: string
    description: string
    labels?: string[]
    assigneeId?: number
  }) =>
    api.post(`/gitlab/projects/${projectId}/issues`, data)
}

// 权限相关 API
export const permissionService = {
  checkPermission: (data: {
    action: string
    resource: string
    resource_id?: string
  }) =>
    api.post('/permissions/check', data),
  
  checkProjectPermission: (projectId: string, action?: string) =>
    api.get(`/permissions/projects/${projectId}`, { 
      params: action ? { action } : undefined 
    }),
  
  checkProjectPermissionDetailed: (projectId: string) =>
    api.get(`/permissions/projects/${projectId}`, { 
      params: { detailed: 'true' }
    }),
  
  getUserPermissions: (projectId?: string) =>
    api.get('/permissions/user', { 
      params: projectId ? { project_id: projectId } : undefined 
    })
}

// 用户管理相关 API (管理员专用)
export const userManagementService = {
  // 获取用户列表
  getUsers: (params?: {
    page?: number
    pageSize?: number
    search?: string
    role?: string
  }) =>
    api.get('/admin/users', { params }),
  
  // 创建用户
  createUser: (data: {
    username: string
    name: string
    email: string
    password: string
    is_admin?: boolean
    default_role?: string
  }) =>
    api.post('/admin/users', data),
  
  // 更新用户信息
  updateUser: (userId: string, data: Partial<User>) =>
    api.put(`/admin/users/${userId}`, data),
  
  // 删除用户
  deleteUser: (userId: string) =>
    api.delete(`/admin/users/${userId}`),
  
  // 获取用户详情
  getUserDetails: (userId: string) =>
    api.get(`/admin/users/${userId}`),
  
  // 更新用户角色
  updateUserRoles: (userId: string, data: {
    is_admin?: boolean
    project_roles?: Array<{
      project_id: string
      role: string
    }>
  }) =>
    api.put(`/admin/users/${userId}/roles`, data),
  
  // 获取用户项目角色
  getUserProjectRoles: (userId: string) =>
    api.get(`/admin/users/${userId}/project-roles`),
  
  // 添加用户到项目
  addUserToProject: (userId: string, projectId: string, role: string) =>
    api.post(`/admin/users/${userId}/projects`, { project_id: projectId, role }),
  
  // 从项目移除用户
  removeUserFromProject: (userId: string, projectId: string) =>
    api.delete(`/admin/users/${userId}/projects/${projectId}`),
  
  // 批量操作用户
  batchUpdateUsers: (data: {
    user_ids: string[]
    action: 'enable' | 'disable' | 'delete'
    roles?: any
  }) =>
    api.post('/admin/users/batch', data),
  
  // 获取用户统计信息
  getUserStats: () =>
    api.get('/admin/users/stats')
}

// 活动相关 API
export const activityService = {
  getRecentActivities: (limit?: number) =>
    api.get('/activities/recent', { params: { limit } }),
  
  getUserActivities: (userId: string, limit?: number) =>
    api.get(`/activities/users/${userId}`, { params: { limit } }),
  
  getMyActivities: (limit?: number) =>
    api.get('/activities/users/me', { params: { limit } })
}

export default api
