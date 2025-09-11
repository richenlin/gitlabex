<template>
  <div class="document-list">
    <div class="header">
      <h1>文档管理</h1>
      <div class="actions" v-if="userStore.isLoggedIn">
        <el-button @click="syncDocuments" :loading="syncing">
          <el-icon><Refresh /></el-icon>
          同步文档
        </el-button>
        <el-button type="primary" @click="showUploadDialog">
          <el-icon><Upload /></el-icon>
          上传文档
        </el-button>
      </div>
    </div>

    <div class="filters">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-input
            v-model="searchQuery"
            placeholder="搜索文档..."
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select v-model="projectFilter" placeholder="选择课题" clearable @change="fetchDocuments">
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="categoryFilter" placeholder="文档分类" clearable @change="fetchDocuments">
            <el-option
              v-for="category in categories"
              :key="category"
              :label="category"
              :value="category"
            />
          </el-select>
        </el-col>
        <el-col :span="3">
          <el-select v-model="typeFilter" placeholder="文件类型" clearable @change="fetchDocuments">
            <el-option label="PDF" value="pdf" />
            <el-option label="Word" value="doc" />
            <el-option label="Excel" value="excel" />
            <el-option label="PPT" value="ppt" />
            <el-option label="图片" value="image" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-col>
      </el-row>
    </div>

    <div class="document-stats">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-statistic title="总文档数" :value="total" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="今日上传" :value="todayUploads" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="总大小" :value="totalSize" suffix="MB" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="文档分类" :value="categories.length" />
        </el-col>
      </el-row>
    </div>

    <div class="view-switch">
      <el-radio-group v-model="viewMode" @change="fetchDocuments">
        <el-radio-button label="list">列表视图</el-radio-button>
        <el-radio-button label="grid">网格视图</el-radio-button>
      </el-radio-group>
    </div>

    <!-- 列表视图 -->
    <el-table
      v-if="viewMode === 'list'"
      :data="documents"
      v-loading="loading"
      style="width: 100%"
    >
      <el-table-column label="文档名称" min-width="200">
        <template #default="{ row }">
          <div class="document-name">
            <el-icon class="file-icon">
              <component :is="getFileIcon(row.file_type)" />
            </el-icon>
            <el-link @click="viewDocument(row.id)" :underline="false">
              {{ row.title }}
            </el-link>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="category" label="分类" width="100" />
      <el-table-column prop="file_type" label="类型" width="80">
        <template #default="{ row }">
          <el-tag size="small">{{ getFileTypeText(row.file_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="file_size" label="大小" width="100">
        <template #default="{ row }">
          {{ formatFileSize(row.file_size) }}
        </template>
      </el-table-column>
      <el-table-column label="上传者" width="120">
        <template #default="{ row }">
          {{ row.upload_user?.name }}
        </template>
      </el-table-column>
      <el-table-column label="关联课题" width="150">
        <template #default="{ row }">
          <el-link v-if="row.project" :href="`/scenes/${row.project.id}`" :underline="false">
            {{ row.project.name }}
          </el-link>
          <span v-else>-</span>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="上传时间" width="150">
        <template #default="{ row }">
          {{ formatDate(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewDocument(row.id)">查看</el-button>
          <el-button size="small" type="primary" @click="downloadDocument(row.id)">下载</el-button>
          <el-button 
            size="small" 
            type="danger" 
            @click="deleteDocument(row.id)"
            v-if="canDelete(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 网格视图 -->
    <div v-else class="document-grid" v-loading="loading">
      <div
        v-for="doc in documents"
        :key="doc.id"
        class="document-card"
        @click="viewDocument(doc.id)"
      >
        <div class="card-header">
          <el-icon class="file-icon-large">
            <component :is="getFileIcon(doc.file_type)" />
          </el-icon>
          <div class="document-actions" @click.stop>
            <el-dropdown>
              <el-icon><More /></el-icon>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="viewDocument(doc.id)">查看</el-dropdown-item>
                  <el-dropdown-item @click="downloadDocument(doc.id)">下载</el-dropdown-item>
                  <el-dropdown-item 
                    v-if="canDelete(doc)"
                    @click="deleteDocument(doc.id)"
                    divided
                  >
                    删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
        <div class="card-body">
          <h4 class="document-title">{{ doc.title }}</h4>
          <p class="document-description">{{ doc.description || '暂无描述' }}</p>
          <div class="document-meta">
            <el-tag size="small">{{ getFileTypeText(doc.file_type) }}</el-tag>
            <span class="file-size">{{ formatFileSize(doc.file_size) }}</span>
          </div>
          <div class="upload-info">
            <span>{{ doc.upload_user?.name }}</span>
            <span>{{ formatDate(doc.created_at) }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="pagination" v-if="total > pageSize">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="fetchDocuments"
        @size-change="fetchDocuments"
      />
    </div>

    <!-- 上传对话框 -->
    <el-dialog v-model="uploadDialogVisible" title="上传文档" width="600px">
      <el-form
        ref="uploadFormRef"
        :model="uploadForm"
        :rules="uploadRules"
        label-width="100px"
      >
        <el-form-item label="文档标题" prop="title">
          <el-input v-model="uploadForm.title" placeholder="请输入文档标题" />
        </el-form-item>
        <el-form-item label="文档描述">
          <el-input
            v-model="uploadForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入文档描述"
          />
        </el-form-item>
        <el-form-item label="关联课题">
          <el-select v-model="uploadForm.project_id" placeholder="选择课题" clearable>
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="文档分类" prop="category">
          <el-select v-model="uploadForm.category" placeholder="选择分类" allow-create>
            <el-option
              v-for="category in categories"
              :key="category"
              :label="category"
              :value="category"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="选择文件" prop="file">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            :on-change="handleFileChange"
            :before-upload="beforeUpload"
          >
            <el-button>选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持 PDF, Word, Excel, PPT, 图片等格式，文件大小不超过50MB
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="uploadDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleUpload" :loading="uploading">
            上传
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { documentService, researchService } from '@/services/api'
import type { Document, ResearchProject } from '@/types'
import { ElMessage, ElMessageBox, type FormInstance, type UploadInstance } from 'element-plus'
import {
  Upload, Search, Refresh, More,
  Document as DocumentIcon,
  Picture, VideoPlay, Headset
} from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const documents = ref<Document[]>([])
const projects = ref<ResearchProject[]>([])
const categories = ref<string[]>([])
const loading = ref(false)
const syncing = ref(false)
const uploading = ref(false)
const searchQuery = ref('')
const projectFilter = ref('')
const categoryFilter = ref('')
const typeFilter = ref('')
const viewMode = ref('list')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const uploadDialogVisible = ref(false)
const uploadFormRef = ref<FormInstance>()
const uploadRef = ref<UploadInstance>()

const uploadForm = ref({
  title: '',
  description: '',
  project_id: '',
  category: '',
  file: null as File | null
})

const uploadRules = {
  title: [{ required: true, message: '请输入文档标题', trigger: 'blur' }],
  category: [{ required: true, message: '请选择文档分类', trigger: 'change' }],
  file: [{ required: true, message: '请选择文件', trigger: 'change' }]
}

// 计算属性
const todayUploads = computed(() => {
  const today = new Date().toDateString()
  return documents.value.filter(doc => 
    new Date(doc.created_at).toDateString() === today
  ).length
})

const totalSize = computed(() => {
  return Math.round(
    documents.value.reduce((sum, doc) => sum + (doc.file_size || 0), 0) / 1024 / 1024
  )
})

// 方法
const fetchDocuments = async () => {
  loading.value = true
  try {
    const response: any = await documentService.getDocuments({
      page: currentPage.value,
      pageSize: pageSize.value,
      search: searchQuery.value || undefined,
      projectId: projectFilter.value || undefined,
      category: categoryFilter.value || undefined
    })
    documents.value = response.documents || []
    total.value = response.pagination?.total || 0
  } catch (error) {
    console.error('获取文档列表失败:', error)
    ElMessage.error('获取文档列表失败')
  } finally {
    loading.value = false
  }
}

const fetchProjects = async () => {
  try {
    const response: any = await researchService.getProjects()
    projects.value = response.projects || []
  } catch (error) {
    console.error('获取课题列表失败:', error)
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

const handleSearch = () => {
  currentPage.value = 1
  fetchDocuments()
}

const syncDocuments = async () => {
  if (!projectFilter.value) {
    ElMessage.warning('请先选择课题')
    return
  }
  
  syncing.value = true
  try {
    await documentService.scanProjectDocuments(projectFilter.value)
    ElMessage.success('文档同步成功')
    fetchDocuments()
  } catch (error) {
    console.error('同步文档失败:', error)
    ElMessage.error('同步文档失败')
  } finally {
    syncing.value = false
  }
}

const viewDocument = (id: string) => {
  router.push(`/documents/${id}`)
}

const downloadDocument = async (id: string) => {
  try {
    const response = await documentService.getDocument(id)
    // TODO: 实现文件下载
    ElMessage.info('下载功能待实现')
  } catch (error) {
    console.error('下载文档失败:', error)
    ElMessage.error('下载文档失败')
  }
}

const deleteDocument = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除这个文档吗？', '确认删除', {
      type: 'warning'
    })
    
    await documentService.deleteDocument(id)
    ElMessage.success('文档删除成功')
    fetchDocuments()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除文档失败:', error)
      ElMessage.error('删除文档失败')
    }
  }
}

const canDelete = (doc: Document) => {
  return userStore.hasRole('admin') || doc.upload_user?.id === userStore.user?.id
}

const showUploadDialog = () => {
  uploadDialogVisible.value = true
  resetUploadForm()
}

const resetUploadForm = () => {
  uploadForm.value = {
    title: '',
    description: '',
    project_id: '',
    category: '',
    file: null
  }
}

const handleFileChange = (file: any) => {
  uploadForm.value.file = file.raw
  if (!uploadForm.value.title && file.name) {
    const nameWithoutExt = file.name.substring(0, file.name.lastIndexOf('.'))
    uploadForm.value.title = nameWithoutExt
  }
}

const beforeUpload = (file: File) => {
  const maxSize = 50 * 1024 * 1024 // 50MB
  if (file.size > maxSize) {
    ElMessage.error('文件大小不能超过50MB')
    return false
  }
  return true
}

const handleUpload = async () => {
  if (!uploadFormRef.value || !uploadForm.value.file) return
  
  try {
    await uploadFormRef.value.validate()
    uploading.value = true
    
    const formData = new FormData()
    formData.append('file', uploadForm.value.file)
    formData.append('title', uploadForm.value.title)
    formData.append('description', uploadForm.value.description)
    formData.append('category', uploadForm.value.category)
    if (uploadForm.value.project_id) {
      formData.append('project_id', uploadForm.value.project_id)
    }
    
    await documentService.createDocument(uploadForm.value)
    ElMessage.success('文档上传成功')
    uploadDialogVisible.value = false
    fetchDocuments()
  } catch (error) {
    console.error('上传文档失败:', error)
    ElMessage.error('上传文档失败')
  } finally {
    uploading.value = false
  }
}

const getFileIcon = (fileType: string) => {
  const iconMap: Record<string, any> = {
    pdf: DocumentIcon,
    doc: DocumentIcon,
    excel: DocumentIcon,
    ppt: DocumentIcon,
    image: Picture,
    video: VideoPlay,
    audio: Headset
  }
  return iconMap[fileType] || DocumentIcon
}

const getFileTypeText = (fileType: string) => {
  const textMap: Record<string, string> = {
    pdf: 'PDF',
    doc: 'Word',
    excel: 'Excel',
    ppt: 'PPT',
    image: '图片',
    video: '视频',
    audio: '音频'
  }
  return textMap[fileType] || '其他'
}

const formatFileSize = (size: number) => {
  if (size < 1024) return `${size}B`
  if (size < 1024 * 1024) return `${Math.round(size / 1024)}KB`
  return `${Math.round(size / 1024 / 1024)}MB`
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('zh-CN')
}

// 生命周期
onMounted(() => {
  fetchDocuments()
  fetchProjects()
  fetchCategories()
})
</script>

<style scoped>
.document-list {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.actions {
  display: flex;
  gap: 10px;
}

.filters {
  margin-bottom: 20px;
}

.document-stats {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.view-switch {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 20px;
}

.document-name {
  display: flex;
  align-items: center;
  gap: 10px;
}

.file-icon {
  font-size: 16px;
  color: #409eff;
}

.document-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.document-card {
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 15px;
  cursor: pointer;
  transition: all 0.3s;
}

.document-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.file-icon-large {
  font-size: 32px;
  color: #409eff;
}

.document-actions {
  cursor: pointer;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.document-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.4;
}

.document-description {
  color: #666;
  font-size: 12px;
  margin: 0;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.document-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.file-size {
  font-size: 12px;
  color: #999;
}

.upload-info {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #999;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 30px;
}
</style>
