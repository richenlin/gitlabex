// 用户相关类型
export interface User {
  id: string
  username: string
  email: string
  name: string
  avatar_url?: string
  role: UserRole
  edu_role: EducationRole
  gitlab_id: number
  is_active: boolean
  last_login_at?: string
  created_at: string
  updated_at: string
}

export enum UserRole {
  GUEST = 'guest',
  STUDENT = 'student', 
  ASSISTANT = 'assistant',
  TEACHER = 'teacher',
  ADMIN = 'admin'
}

export enum EducationRole {
  GUEST = 10,
  STUDENT = 20,
  ASSISTANT = 30,
  TEACHER = 40,
  ADMIN = 50
}

// 研究课题相关类型
export interface ResearchProject {
  id: string
  name: string
  description: string
  is_public: boolean
  gitlab_project_id?: number
  gitlab_url?: string
  gitlab_ssh_url?: string
  gitlab_namespace?: string
  creator_id: string
  creator: User
  status: 'active' | 'archived' | 'suspended'
  auto_index_enabled: boolean
  last_sync_time?: string
  created_at: string
  updated_at: string
  members?: ProjectMember[]
  tags?: string[]
  view_count?: number
}

export interface ProjectMember {
  id: string
  project_id: string
  user_id: string
  user: User
  role: ProjectRole
  joined_at: string
}

export enum ProjectRole {
  OWNER = 'owner',
  MAINTAINER = 'maintainer', 
  DEVELOPER = 'developer',
  REPORTER = 'reporter',
  GUEST = 'guest'
}

// 为了向后兼容，保留Scene类型别名
export type Scene = ResearchProject

// 话题相关类型
export interface Topic {
  id: string
  title: string
  content: string
  project_id?: string
  project?: ResearchProject
  author_id: string
  author: User
  gitlab_issue_id?: number
  status: 'open' | 'closed'
  labels: string[]
  likes_count: number
  comments_count: number
  is_pinned: boolean
  priority: 'low' | 'medium' | 'high' | 'urgent'
  created_at: string
  updated_at: string
  comments?: Comment[]
}

// 评论类型
export interface Comment {
  id: string
  content: string
  topic_id: string
  author_id: string
  author: User
  parent_id?: string
  replies?: Comment[]
  likes_count: number
  created_at: string
  updated_at: string
}

// 文档相关类型
export interface Document {
  id: string
  title: string
  description: string
  file_path: string
  file_type: string
  file_size: number
  category: string
  status: string
  project_id?: string
  project?: ResearchProject
  homework_id?: string
  homework?: Homework
  uploader_id: string
  upload_user: User
  gitlab_file_path?: string
  gitlab_id?: string
  auto_indexed: boolean
  last_sync_time?: string
  download_count?: number
  tags?: string[]
  created_at: string
  updated_at: string
}

// 作业相关类型
export interface Homework {
  id: string
  title: string
  description: string
  project_id: string
  project: ResearchProject
  creator_id: string
  author: User
  deadline: string
  max_grade: number
  status: 'draft' | 'published' | 'closed'
  gitlab_branch?: string
  gitlab_path?: string
  submissions?: HomeworkSubmission[]
  created_at: string
  updated_at: string
}

export interface HomeworkSubmission {
  id: string
  homework_id: string
  student_id: string
  student: User
  content: string
  files?: Document[]
  grade?: number
  feedback?: string
  status: 'pending' | 'submitted' | 'graded'
  gitlab_commit_id?: string
  gitlab_branch?: string
  submitted_at?: string
  created_at: string
  updated_at: string
}

// 通知相关类型
export interface Notification {
  id: string
  title: string
  content: string
  type: 'info' | 'warning' | 'error' | 'success' | 'system' | 'homework' | 'project' | 'topic' | 'document'
  user_id: string
  is_read: boolean
  related_type?: string
  related_id?: string
  action_url?: string
  created_at: string
}

// API响应类型
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

// 路由元信息类型
export interface RouteMeta {
  title?: string
  requiresAuth?: boolean
  roles?: UserRole[]
  layout?: string
}
