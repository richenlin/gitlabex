export interface Project {
  id: number
  name: string
  path_with_namespace: string
}

export interface WikiPage {
  title: string
  slug: string
  content?: string
  editableAttachments: number
  updated_at: string
}

export interface WikiAttachment {
  id: number
  name: string
  path: string
  size: number
  content_type: string
  created_at: string
  updated_at: string
  can_edit: boolean
} 