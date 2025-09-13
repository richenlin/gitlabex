// GitLab用户相关类型
export interface User {
  id: number          // GitLab用户ID
  username: string
  email: string
  name: string
  avatar_url?: string
  bio?: string        // 个人简介
  location?: string   // 所在地
  website_url?: string // 个人网站
  is_admin: boolean   // GitLab管理员权限
  gitlab_role?: GitLabRole  // 在项目中的角色
  role?: string       // 用户角色
  edu_role?: string   // 教育角色
  last_login_at?: string // 最后登录时间
  token_expiry?: string  // Token过期时间
  is_active?: boolean    // 是否活跃
  created_at: string
  updated_at?: string
}

// GitLab角色枚举（对应GitLab访问级别）
export enum GitLabRole {
  GUEST = 'guest',           // 10 - 访客
  REPORTER = 'reporter',     // 20 - 学生
  DEVELOPER = 'developer',   // 30 - 研究员
  MAINTAINER = 'maintainer', // 40 - 教师
  OWNER = 'owner'           // 50 - 管理员
}

// 为了向后兼容，保留旧的枚举但映射到GitLab角色
export enum UserRole {
  GUEST = 'guest',
  STUDENT = 'reporter',     // 映射到GitLab Reporter
  ASSISTANT = 'developer',  // 映射到GitLab Developer
  TEACHER = 'maintainer',   // 映射到GitLab Maintainer
  ADMIN = 'owner'          // 映射到GitLab Owner
}

export enum EducationRole {
  GUEST = 10,
  STUDENT = 20,
  ASSISTANT = 30,
  TEACHER = 40,
  ADMIN = 50
}

// GitLab项目成员类型
export interface GitLabProjectMember {
  id: number
  username: string
  name: string
  email?: string
  avatar_url: string
  access_level: number
  state: string
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
  creator_id: number  // 修改为number类型，与User.id保持一致
  creator?: User       // 可选，因为有时候可能不包含创建者详细信息
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
  status: 'opened' | 'closed'
  labels: string[]
  likes_count: number
  like_count?: number // 别名，用于兼容
  dislike_count?: number
  comments_count: number
  is_pinned: boolean
  priority: 'low' | 'medium' | 'high' | 'urgent'
  created_at: string
  updated_at: string
  comments?: Comment[]
  view_count?: number
  
  // 用户交互状态 (前端使用，不从后端获取)
  user_liked?: boolean
  user_disliked?: boolean
  liking?: boolean
  disliking?: boolean
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
  deadline?: string
  due_date?: string  // 别名，用于兼容后端
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

// 活动相关类型
export interface ActivityItem {
  id: string
  type: 'document' | 'topic' | 'homework' | 'comment'
  title: string
  description: string
  user_name: string
  user_avatar?: string
  project_name?: string
  created_at: string
  url: string
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
  roles?: UserRole[] | string[]
  layout?: string
}
