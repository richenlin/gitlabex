<template>
  <div class="document-detail" v-loading="loading">
    <div v-if="document" class="document-container">
      <!-- 面包屑导航 -->
      <el-breadcrumb class="breadcrumb" separator=">">
        <el-breadcrumb-item :to="{ path: '/documents' }">文档</el-breadcrumb-item>
        <el-breadcrumb-item>{{ document.title }}</el-breadcrumb-item>
      </el-breadcrumb>

      <!-- 文档头部信息 -->
      <div class="document-header card">
        <div class="header-content">
          <div class="title-section">
            <h1 class="document-title">{{ document.title }}</h1>
            <div class="document-meta">
              <span>上传时间: {{ formatDate(document.created_at) }}</span>
              <span>上传者: {{ document.upload_user?.name }}</span>
              <span>文件大小: {{ formatFileSize(document.file_size) }}</span>
              <span>下载次数: {{ document.download_count || 0 }}</span>
              <el-tag :type="getStatusTagType(document.status)" size="small">
                {{ getDocumentStatusText(document.status) }}
              </el-tag>
            </div>
          </div>
          <div class="header-actions">
            <el-button @click="downloadDocument" type="primary">
              <el-icon><Download /></el-icon>
              下载文档
            </el-button>
            <el-button @click="editDocument" v-if="canEdit">
              <el-icon><Edit /></el-icon>
              编辑信息
            </el-button>
            <el-button @click="requestEdit" v-if="canRequestEdit">
              <el-icon><EditPen /></el-icon>
              申请修改
            </el-button>
          </div>
        </div>
        
        <div class="document-description" v-if="document.description">
          <h3>文档描述</h3>
          <p>{{ document.description }}</p>
        </div>

        <div class="document-info">
          <div class="info-item">
            <span class="label">文件类型:</span>
            <span class="value">{{ getFileTypeText(document.file_type) }}</span>
          </div>
          <div class="info-item">
            <span class="label">所属分类:</span>
            <span class="value">{{ document.category || '未分类' }}</span>
          </div>
          <div class="info-item" v-if="document.project">
            <span class="label">所属课题:</span>
            <router-link :to="`/scenes/${document.project.id}`" class="project-link">
              {{ document.project.name }}
            </router-link>
          </div>
          <div class="info-item" v-if="document.auto_indexed">
            <span class="label">索引方式:</span>
            <span class="value">自动索引</span>
          </div>
        </div>

        <div class="document-tags" v-if="document.tags?.length">
          <h4>标签</h4>
          <div class="tags-list">
            <el-tag 
              v-for="tag in document.tags" 
              :key="tag" 
              size="small"
              class="document-tag"
            >
              {{ tag }}
            </el-tag>
          </div>
        </div>
      </div>


      <!-- 待审核的编辑请求 -->
      <div class="pending-requests card" v-if="canViewHistory && pendingRequests.length > 0">
        <div class="requests-header">
          <h3>待审核的编辑请求</h3>
        </div>
        
        <div class="requests-list" v-loading="requestsLoading">
          <div 
            v-for="request in pendingRequests" 
            :key="request.id"
            class="request-item"
          >
            <div class="request-info">
              <div class="request-user">
                <span>申请人: {{ request.requester_id }}</span>
              </div>
              <div class="request-time">{{ formatDate(request.created_at) }}</div>
            </div>
            <div class="request-content">
              <div class="request-reason" v-if="request.reason">
                <strong>修改原因:</strong> {{ request.reason }}
              </div>
              <div class="request-changes" v-if="request.title || request.description">
                <div v-if="request.title">
                  <strong>标题:</strong> {{ request.title }}
                </div>
                <div v-if="request.description">
                  <strong>描述:</strong> {{ request.description }}
                </div>
                <div v-if="request.category">
                  <strong>分类:</strong> {{ request.category }}
                </div>
              </div>
            </div>
            <div class="request-actions">
              <el-button @click="approveRequest(request.id)" type="success" size="small">
                <el-icon><Check /></el-icon>
                通过
              </el-button>
              <el-button @click="rejectRequest(request.id)" type="danger" size="small">
                <el-icon><Close /></el-icon>
                拒绝
              </el-button>
            </div>
          </div>
        </div>
      </div>

      <!-- 编辑历史 -->
      <div class="edit-history card" v-if="canViewHistory">
        <div class="history-header">
          <h3>编辑历史</h3>
        </div>
        
        <div class="history-list" v-loading="historyLoading">
          <div 
            v-for="history in editHistory" 
            :key="history.id"
            class="history-item"
          >
            <div class="history-info">
              <div class="history-user">
                <img :src="history.requester?.avatar_url || '/default-avatar.png'" alt="用户头像">
                <span>{{ history.requester?.name }}</span>
              </div>
              <div class="history-time">{{ formatDate(history.created_at) }}</div>
            </div>
            <div class="history-content">
              <div class="history-action">
                <el-tag :type="getEditStatusTagType(history.status)">
                  {{ getEditStatusText(history.status) }}
                </el-tag>
              </div>
              <div class="history-reason" v-if="history.reason">
                <strong>修改原因:</strong> {{ history.reason }}
              </div>
              <div class="history-changes" v-if="history.title || history.description">
                <div v-if="history.title">
                  <strong>标题:</strong> {{ history.title }}
                </div>
                <div v-if="history.description">
                  <strong>描述:</strong> {{ history.description }}
                </div>
              </div>
              <div class="history-review" v-if="history.review_comments">
                <strong>审核意见:</strong> {{ history.review_comments }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 相关文档推荐 -->
      <div class="related-documents card" v-if="relatedDocuments.length">
        <h3>相关文档</h3>
        <div class="related-list">
          <div 
            v-for="doc in relatedDocuments" 
            :key="doc.id"
            class="related-item"
            @click="viewDocument(doc.id)"
          >
            <div class="related-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="related-info">
              <h4>{{ doc.title }}</h4>
              <p>{{ doc.description?.substring(0, 100) }}...</p>
              <div class="related-meta">
                <span>{{ formatFileSize(doc.file_size) }}</span>
                <span>{{ doc.download_count || 0 }} 次下载</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 编辑文档信息对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑文档信息" width="600px">
      <el-form :model="editForm" :rules="editRules" ref="editFormRef" label-width="80px">
        <el-form-item label="文档标题" prop="title">
          <el-input v-model="editForm.title" placeholder="请输入文档标题" />
        </el-form-item>
        <el-form-item label="文档描述" prop="description">
          <el-input 
            v-model="editForm.description" 
            type="textarea" 
            :rows="4"
            placeholder="请输入文档描述"
          />
        </el-form-item>
        <el-form-item label="文档分类" prop="category">
          <el-select v-model="editForm.category" placeholder="请选择分类" style="width: 100%">
            <el-option 
              v-for="category in categories" 
              :key="category" 
              :label="category" 
              :value="category"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-input
            v-model="tagInput"
            placeholder="输入标签后回车添加"
            @keyup.enter="addTag"
          />
          <div class="tags-container" v-if="editForm.tags.length">
            <el-tag
              v-for="tag in editForm.tags"
              :key="tag"
              closable
              @close="removeTag(tag)"
              style="margin: 5px 5px 0 0"
            >
              {{ tag }}
            </el-tag>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveEdit" :loading="saving">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 申请修改对话框 -->
    <el-dialog v-model="requestEditDialogVisible" title="申请修改文档信息" width="600px">
      <el-form :model="requestForm" label-width="80px">
        <el-form-item label="修改标题">
          <el-input v-model="requestForm.title" placeholder="新的文档标题" />
        </el-form-item>
        <el-form-item label="修改描述">
          <el-input 
            v-model="requestForm.description" 
            type="textarea" 
            :rows="4"
            placeholder="新的文档描述"
          />
        </el-form-item>
        <el-form-item label="修改分类">
          <el-select v-model="requestForm.category" placeholder="请选择分类" style="width: 100%">
            <el-option 
              v-for="category in categories" 
              :key="category" 
              :label="category" 
              :value="category"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="修改原因" required>
          <el-input 
            v-model="requestForm.reason" 
            type="textarea" 
            :rows="3"
            placeholder="请说明修改原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="requestEditDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEditRequest" :loading="requesting">
          提交申请
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { documentService } from '@/services/api'
import type { Document } from '@/types'
import { ElMessage, type FormInstance } from 'element-plus'
import { 
  Download, 
  Edit, 
  EditPen, 
  FullScreen, 
  Document as DocumentIcon,
  Check,
  Close
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const document = ref<Document | null>(null)
const editHistory = ref<any[]>([])
const pendingRequests = ref<any[]>([])
const relatedDocuments = ref<Document[]>([])
const categories = ref<string[]>([])
const textContent = ref('')

const loading = ref(false)
const historyLoading = ref(false)
const requestsLoading = ref(false)
const saving = ref(false)
const requesting = ref(false)
const isFullscreen = ref(false)

// 对话框状态
const editDialogVisible = ref(false)
const requestEditDialogVisible = ref(false)
const editFormRef = ref<FormInstance>()
const tagInput = ref('')

// 表单数据
const editForm = ref({
  title: '',
  description: '',
  category: '',
  tags: [] as string[]
})

const requestForm = ref({
  title: '',
  description: '',
  category: '',
  reason: ''
})

const editRules = {
  title: [{ required: true, message: '请输入文档标题', trigger: 'blur' }]
}

// 计算属性
const documentId = computed(() => route.params.id as string)

const canEdit = computed(() => {
  // 简化检查：管理员、教师或上传者，具体权限由后端验证
  return userStore.hasRole('admin') || 
         userStore.hasRole('teacher') || 
         document.value?.uploader_id === userStore.user?.id
})

const canRequestEdit = computed(() => {
  return userStore.isLoggedIn && !canEdit.value
})


const canViewHistory = computed(() => {
  // 简化检查：管理员或教师，具体权限由后端验证
  return userStore.hasRole('admin') || userStore.hasRole('teacher')
})

// 方法
const fetchDocument = async () => {
  loading.value = true
  try {
    const response: any = await documentService.getDocument(documentId.value)
    document.value = response
    
    
    // 加载相关数据
    fetchRelatedDocuments()
    if (canViewHistory.value) {
      fetchEditHistory()
    }
  } catch (error) {
    console.error('获取文档详情失败:', error)
    ElMessage.error('获取文档详情失败')
  } finally {
    loading.value = false
  }
}

const fetchEditHistory = async () => {
  historyLoading.value = true
  try {
    const response: any = await documentService.getDocumentEditHistory(documentId.value)
    editHistory.value = response.edit_history || []
  } catch (error) {
    console.error('获取编辑历史失败:', error)
  } finally {
    historyLoading.value = false
  }
}

const fetchPendingRequests = async () => {
  requestsLoading.value = true
  try {
    const response: any = await documentService.getEditRequests()
    // 过滤出当前文档的待审核请求
    pendingRequests.value = (response.edit_requests || []).filter((req: any) => 
      req.document_id === documentId.value && req.status === 'pending'
    )
  } catch (error) {
    console.error('获取待审核请求失败:', error)
  } finally {
    requestsLoading.value = false
  }
}

const fetchRelatedDocuments = async () => {
  if (!document.value?.project?.id) return
  
  try {
    const response = await documentService.getDocuments({
      projectId: document.value.project.id,
      pageSize: 5
    })
    const docs = (response as any).documents || []
    relatedDocuments.value = docs.filter((doc: Document) => doc.id !== documentId.value)
  } catch (error) {
    console.error('获取相关文档失败:', error)
  }
}

const fetchCategories = async () => {
  try {
    const response: any = await documentService.getCategories()
    categories.value = response || []
  } catch (error) {
    console.error('获取分类列表失败:', error)
  }
}


const downloadDocument = async () => {
  if (!document.value) return
  
  try {
    const response = await documentService.downloadDocument(document.value.id)
    
    // 创建下载链接
    const blob = new Blob([response.data], { type: 'application/octet-stream' })
    const url = window.URL.createObjectURL(blob)
    const link = window.document.createElement('a')
    link.href = url
    
    // 从响应头获取文件名，如果没有则使用文档标题
    const contentDisposition = response.headers?.['content-disposition']
    let filename = document.value.title
    if (contentDisposition) {
      // 改进的文件名解析逻辑
      const filenameMatch = contentDisposition.match(/filename[^;=\n]*=([^;=\n]+)/)
      if (filenameMatch && filenameMatch[1]) {
        filename = filenameMatch[1].trim().replace(/['"]/g, '')
      }
    }
    
    // 确保文件名有扩展名
    if (document.value.file_type && !filename.includes('.')) {
      filename = `${filename}.${document.value.file_type}`
    }
    
    link.download = filename
    window.document.body.appendChild(link)
    link.click()
    window.document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('文档下载成功')
  } catch (error) {
    console.error('下载文档失败:', error)
    ElMessage.error('下载文档失败')
  }
}

const editDocument = () => {
  if (!document.value) return
  
  editForm.value = {
    title: document.value.title,
    description: document.value.description || '',
    category: document.value.category || '',
    tags: [...(document.value.tags || [])]
  }
  editDialogVisible.value = true
  fetchCategories()
}

const requestEdit = () => {
  if (!document.value) return
  
  requestForm.value = {
    title: document.value.title,
    description: document.value.description || '',
    category: document.value.category || '',
    reason: ''
  }
  requestEditDialogVisible.value = true
  fetchCategories()
}

const saveEdit = async () => {
  if (!editFormRef.value || !document.value) return
  
  try {
    await editFormRef.value.validate()
    saving.value = true
    
    await documentService.updateDocument(document.value.id, {
      title: editForm.value.title,
      description: editForm.value.description,
      category: editForm.value.category,
      tags: editForm.value.tags
    })
    
    ElMessage.success('文档信息更新成功')
    editDialogVisible.value = false
    fetchDocument()
  } catch (error) {
    console.error('更新文档信息失败:', error)
    ElMessage.error('更新文档信息失败')
  } finally {
    saving.value = false
  }
}

const submitEditRequest = async () => {
  if (!requestForm.value.reason.trim()) {
    ElMessage.warning('请填写修改原因')
    return
  }
  
  requesting.value = true
  try {
    await documentService.submitEditRequest(documentId.value, {
      proposed_changes: {
        title: requestForm.value.title,
        description: requestForm.value.description,
        category: requestForm.value.category
      },
      reason: requestForm.value.reason
    })
    
    ElMessage.success('修改申请提交成功，等待审核')
    requestEditDialogVisible.value = false
    // 刷新待审核请求列表
    fetchPendingRequests()
  } catch (error) {
    console.error('提交修改申请失败:', error)
    ElMessage.error('提交修改申请失败')
  } finally {
    requesting.value = false
  }
}

const approveRequest = async (requestId: string) => {
  try {
    await documentService.reviewEditRequest(requestId, {
      approved: true,
      comments: '审核通过'
    })
    
    ElMessage.success('编辑请求已通过')
    // 刷新数据
    fetchPendingRequests()
    fetchEditHistory()
    fetchDocument()
  } catch (error) {
    console.error('审核失败:', error)
    ElMessage.error('审核失败')
  }
}

const rejectRequest = async (requestId: string) => {
  try {
    await documentService.reviewEditRequest(requestId, {
      approved: false,
      comments: '审核拒绝'
    })
    
    ElMessage.success('编辑请求已拒绝')
    // 刷新数据
    fetchPendingRequests()
    fetchEditHistory()
  } catch (error) {
    console.error('审核失败:', error)
    ElMessage.error('审核失败')
  }
}

const addTag = () => {
  if (tagInput.value.trim() && !editForm.value.tags.includes(tagInput.value.trim())) {
    editForm.value.tags.push(tagInput.value.trim())
    tagInput.value = ''
  }
}

const removeTag = (tag: string) => {
  const index = editForm.value.tags.indexOf(tag)
  if (index > -1) {
    editForm.value.tags.splice(index, 1)
  }
}

const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value
}

const viewDocument = (id: string) => {
  router.push(`/documents/${id}`)
}


const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatFileSize = (size: number) => {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}

const getFileTypeText = (fileType: string) => {
  const typeMap: Record<string, string> = {
    pdf: 'PDF文档',
    doc: 'Word文档',
    docx: 'Word文档',
    xls: 'Excel表格',
    xlsx: 'Excel表格',
    ppt: 'PowerPoint演示文稿',
    pptx: 'PowerPoint演示文稿',
    txt: '文本文件',
    md: 'Markdown文档',
    jpg: 'JPEG图片',
    jpeg: 'JPEG图片',
    png: 'PNG图片',
    gif: 'GIF图片'
  }
  return typeMap[fileType.toLowerCase()] || fileType.toUpperCase()
}

const getDocumentStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝'
  }
  return statusMap[status] || status
}

const getStatusTagType = (status: string) => {
  const typeMap: Record<string, string> = {
    pending: 'warning',
    approved: 'success',
    rejected: 'danger'
  }
  return typeMap[status] || 'info'
}

const getEditStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝'
  }
  return statusMap[status] || status
}

const getEditStatusTagType = (status: string) => {
  const typeMap: Record<string, string> = {
    pending: 'warning',
    approved: 'success',
    rejected: 'danger'
  }
  return typeMap[status] || 'info'
}

// 生命周期
onMounted(() => {
  fetchDocument()
  if (canViewHistory.value) {
    fetchPendingRequests()
  }
})
</script>

<style scoped>
.document-detail {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.breadcrumb {
  margin-bottom: 20px;
}

.document-header {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.title-section h1 {
  margin: 0 0 8px 0;
  color: var(--primary-color);
}

.document-meta {
  display: flex;
  gap: 16px;
  align-items: center;
  font-size: 14px;
  color: var(--light-text);
  flex-wrap: wrap;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.document-description {
  margin-bottom: 20px;
}

.document-description h3 {
  margin: 0 0 8px 0;
  color: var(--text-color);
}

.document-description p {
  margin: 0;
  color: var(--light-text);
  line-height: 1.6;
}

.document-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.info-item {
  display: flex;
  gap: 8px;
}

.info-item .label {
  font-weight: 500;
  color: var(--text-color);
  min-width: 80px;
}

.info-item .value {
  color: var(--light-text);
}

.project-link {
  color: var(--primary-color);
  text-decoration: none;
}

.project-link:hover {
  text-decoration: underline;
}

.document-tags h4 {
  margin: 0 0 8px 0;
  color: var(--text-color);
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.document-tag {
  margin: 0;
}


/* 编辑历史样式 */
.edit-history {
  margin-bottom: 20px;
}

.history-header h3 {
  margin: 0 0 16px 0;
  color: var(--text-color);
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.history-item {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
}

.history-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.history-user {
  display: flex;
  align-items: center;
  gap: 8px;
}

.history-user img {
  width: 24px;
  height: 24px;
  border-radius: 50%;
}

.history-time {
  font-size: 12px;
  color: var(--light-text);
}

.history-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.history-reason,
.history-changes,
.history-review {
  font-size: 14px;
  color: var(--text-color);
}

/* 相关文档样式 */
.related-documents h3 {
  margin: 0 0 16px 0;
  color: var(--text-color);
}

.related-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.related-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.related-item:hover {
  background-color: var(--background-light);
}

.related-icon {
  font-size: 24px;
  color: var(--primary-color);
}

.related-info {
  flex: 1;
}

.related-info h4 {
  margin: 0 0 4px 0;
  font-size: 14px;
  color: var(--text-color);
}

.related-info p {
  margin: 0 0 8px 0;
  font-size: 12px;
  color: var(--light-text);
  line-height: 1.4;
}

.related-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--lighter-text);
}

/* 待审核请求样式 */
.pending-requests {
  margin-bottom: 20px;
}

.requests-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}

.requests-header h3 {
  margin: 0;
  color: var(--primary-color);
  font-size: 18px;
}

.requests-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.request-item {
  padding: 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-color);
  transition: all 0.3s ease;
}

.request-item:hover {
  border-color: var(--primary-color);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.request-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.request-user {
  font-weight: 500;
  color: var(--text-color);
}

.request-time {
  font-size: 12px;
  color: var(--light-text);
}

.request-content {
  margin-bottom: 16px;
}

.request-reason {
  margin-bottom: 12px;
  padding: 8px 12px;
  background: var(--light-bg);
  border-radius: 6px;
  font-size: 14px;
}

.request-changes {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 14px;
}

.request-changes div {
  padding: 6px 10px;
  background: var(--light-bg);
  border-radius: 4px;
}

.request-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

/* 表单样式 */
.tags-container {
  margin-top: 10px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .document-detail {
    padding: 10px;
  }
  
  .header-content {
    flex-direction: column;
    gap: 16px;
  }
  
  .document-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .document-info {
    grid-template-columns: 1fr;
  }
  
  
  .history-info {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .related-item {
    flex-direction: column;
    text-align: center;
  }
}
</style>
