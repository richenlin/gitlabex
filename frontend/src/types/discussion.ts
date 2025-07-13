// Discussion相关的类型定义

export interface User {
  id: number
  name: string
  email: string
  avatar?: string
}

export interface Project {
  id: number
  name: string
  description?: string
  path?: string
}

export interface Discussion {
  id: number
  title: string
  content: string
  project_id: number
  author_id: number
  author: User
  category: string
  tags: string
  status: 'open' | 'closed'
  is_public: boolean
  is_pinned: boolean
  is_resolved: boolean
  view_count: number
  reply_count: number
  like_count: number
  gitlab_issue_id?: number
  gitlab_issue_url?: string
  created_at: string
  updated_at: string
}

export interface DiscussionReply {
  id: number
  discussion_id: number
  author_id: number
  author: User
  content: string
  parent_reply_id?: number
  gitlab_note_id?: number
  gitlab_note_url?: string
  is_resolved: boolean
  created_at: string
  updated_at: string
}

export interface DiscussionListResponse {
  discussions: Discussion[]
  total: number
  page: number
  page_size: number
}

export interface DiscussionDetailResponse {
  discussion: Discussion
  replies: DiscussionReply[]
  total_replies: number
}

export interface DiscussionFilters {
  projectId: number | string
  category: string
  status: string
}

export interface CreateDiscussionRequest {
  project_id: number
  title: string
  content: string
  category?: string
  tags?: string
  is_public?: boolean
}

export interface CreateReplyRequest {
  content: string
  parent_reply_id?: number
}

export interface UpdateDiscussionRequest {
  title?: string
  content?: string
  category?: string
  tags?: string
  is_public?: boolean
}

export interface CategoryResponse {
  categories: string[]
}

export type DiscussionCategory = 'general' | 'question' | 'announcement' | 'help' | 'feedback'
export type DiscussionStatus = 'open' | 'closed' 