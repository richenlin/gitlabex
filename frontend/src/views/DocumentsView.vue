<template>
  <div class="documents-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <el-icon class="page-icon"><Document /></el-icon>
          文档管理
        </h1>
        <p class="page-description">管理您的文档，支持在线编辑和协作</p>
      </div>
      
      <!-- 操作栏 -->
      <div class="toolbar">
        <el-upload
          ref="uploadRef"
          class="upload-demo"
          action=""
          :auto-upload="false"
          :on-change="handleFileSelect"
          :show-file-list="false"
          accept=".docx,.xlsx,.pptx,.txt,.pdf"
        >
          <el-button type="primary" :icon="Upload">
            上传文档
          </el-button>
        </el-upload>
        
        <el-button 
          type="success" 
          :icon="Plus"
          @click="createTestDocument"
          :loading="isCreating"
        >
          创建测试文档
        </el-button>
        
        <el-button 
          :icon="Refresh"
          @click="loadDocuments"
          :loading="isLoading"
        >
          刷新
        </el-button>
      </div>
    </div>

    <!-- 文档列表 -->
    <div class="documents-content">
      <el-card class="documents-card" v-loading="isLoading">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><FolderOpened /></el-icon>
              我的文档
            </span>
            <span class="document-count">共 {{ documents.length }} 个文档</span>
          </div>
        </template>

        <!-- 文档表格 -->
        <el-table 
          :data="documents" 
          style="width: 100%" 
          :default-sort="{ prop: 'created_at', order: 'descending' }"
          empty-text="暂无文档"
          @row-click="handleRowClick"
          class="documents-table"
        >
          <el-table-column prop="id" label="ID" width="80" />
          
          <el-table-column prop="title" label="文档名称" min-width="200">
            <template #default="scope">
              <div class="document-title">
                <el-icon class="file-icon" :color="getFileTypeColor(scope.row.file_type)">
                  <component :is="getFileTypeIcon(scope.row.file_type)" />
                </el-icon>
                <span class="title-text">{{ scope.row.title || `文档_${scope.row.id}` }}</span>
              </div>
            </template>
          </el-table-column>
          
          <el-table-column prop="file_type" label="类型" width="100">
            <template #default="scope">
              <el-tag :type="getFileTypeTagType(scope.row.file_type)" size="small">
                {{ scope.row.file_type?.toUpperCase() || 'DOCX' }}
              </el-tag>
            </template>
          </el-table-column>
          
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'" size="small">
                {{ scope.row.status === 'active' ? '活跃' : '非活跃' }}
              </el-tag>
            </template>
          </el-table-column>
          
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="scope">
              {{ formatDate(scope.row.created_at) }}
            </template>
          </el-table-column>
          
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="scope">
              <el-button-group>
                <el-button 
                  type="primary" 
                  size="small" 
                  :icon="Edit"
                  @click.stop="editDocument(scope.row)"
                >
                  编辑
                </el-button>
                <el-button 
                  type="info" 
                  size="small" 
                  :icon="View"
                  @click.stop="viewDocument(scope.row)"
                >
                  预览
                </el-button>
                <el-button 
                  type="danger" 
                  size="small" 
                  :icon="Delete"
                  @click.stop="deleteDocument(scope.row)"
                >
                  删除
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 上传对话框 -->
    <el-dialog
      v-model="uploadDialogVisible"
      title="上传文档"
      width="500px"
      :before-close="handleUploadClose"
    >
      <div class="upload-form">
        <el-form :model="uploadForm" label-width="80px">
          <el-form-item label="文件">
            <div class="file-info" v-if="selectedFile">
              <el-icon class="file-icon"><Document /></el-icon>
              <span class="file-name">{{ selectedFile.name }}</span>
              <span class="file-size">({{ formatFileSize(selectedFile.size) }})</span>
            </div>
          </el-form-item>
          
          <el-form-item label="编辑模式">
            <el-radio-group v-model="uploadForm.mode">
              <el-radio label="edit">编辑模式</el-radio>
              <el-radio label="view">只读模式</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-form>
      </div>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="uploadDialogVisible = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="handleUpload"
            :loading="isUploading"
          >
            上传
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Document, 
  Upload, 
  Plus, 
  Refresh, 
  FolderOpened, 
  Edit, 
  View, 
  Delete,
  DocumentChecked,
  Memo,
  Files
} from '@element-plus/icons-vue'
import { ApiService } from '../services/api'
import type { UploadFile } from 'element-plus'

const router = useRouter()

// 响应式数据
const documents = ref<any[]>([])
const isLoading = ref(false)
const isCreating = ref(false)
const isUploading = ref(false)
const uploadDialogVisible = ref(false)
const selectedFile = ref<File | null>(null)
const uploadForm = ref({
  mode: 'edit'
})

// 模拟文档数据
const mockDocuments = ref([
  {
    id: 1,
    title: '项目需求文档.docx',
    file_type: 'docx',
    status: 'active',
    created_at: new Date().toISOString(),
  },
  {
    id: 2,
    title: '数据分析报告.xlsx',
    file_type: 'xlsx',
    status: 'active',
    created_at: new Date(Date.now() - 86400000).toISOString(),
  },
  {
    id: 3,
    title: '产品演示.pptx',
    file_type: 'pptx',
    status: 'inactive',
    created_at: new Date(Date.now() - 172800000).toISOString(),
  }
])

// 生命周期
onMounted(() => {
  loadDocuments()
})

// 方法
const loadDocuments = async () => {
  isLoading.value = true
  try {
    // 由于后端暂时没有文档列表API，使用模拟数据
    await new Promise(resolve => setTimeout(resolve, 1000))
    documents.value = mockDocuments.value
    ElMessage.success('文档列表加载成功')
  } catch (error) {
    console.error('加载文档失败:', error)
    ElMessage.error('加载文档失败')
    // 使用模拟数据作为备用
    documents.value = mockDocuments.value
  } finally {
    isLoading.value = false
  }
}

const createTestDocument = async () => {
  isCreating.value = true
  try {
    const document = await ApiService.createTestDocument()
    ElMessage.success('测试文档创建成功')
    // 跳转到编辑器
    router.push(`/documents/editor/${document.document_id}`)
  } catch (error) {
    console.error('创建测试文档失败:', error)
    ElMessage.error('创建测试文档失败')
  } finally {
    isCreating.value = false
  }
}

const handleFileSelect = (file: UploadFile) => {
  selectedFile.value = file.raw as File
  uploadDialogVisible.value = true
}

const handleUpload = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请选择文件')
    return
  }
  
  isUploading.value = true
  try {
    const document = await ApiService.uploadDocument(selectedFile.value, uploadForm.value.mode)
    ElMessage.success('文档上传成功')
    uploadDialogVisible.value = false
    // 刷新文档列表
    loadDocuments()
    // 跳转到编辑器
    router.push(`/documents/editor/${document.document_id}`)
  } catch (error) {
    console.error('上传文档失败:', error)
    ElMessage.error('上传文档失败')
  } finally {
    isUploading.value = false
  }
}

const handleUploadClose = () => {
  selectedFile.value = null
  uploadForm.value.mode = 'edit'
}

const editDocument = (document: any) => {
  router.push(`/documents/editor/${document.id}`)
}

const viewDocument = (document: any) => {
  // 以只读模式打开文档
  router.push(`/documents/editor/${document.id}?mode=view`)
}

const deleteDocument = async (document: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除文档 "${document.title}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    // 这里应该调用删除API
    // await ApiService.deleteDocument(document.id)
    
    // 模拟删除
    const index = documents.value.findIndex(d => d.id === document.id)
    if (index > -1) {
      documents.value.splice(index, 1)
    }
    
    ElMessage.success('文档删除成功')
  } catch (error) {
    console.log('取消删除')
  }
}

const handleRowClick = (row: any) => {
  editDocument(row)
}

// 工具方法
const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN')
}

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getFileTypeIcon = (fileType: string) => {
  switch (fileType?.toLowerCase()) {
    case 'docx':
    case 'doc':
      return DocumentChecked
    case 'xlsx':
    case 'xls':
      return Memo
    case 'pptx':
    case 'ppt':
      return Files
    default:
      return Document
  }
}

const getFileTypeColor = (fileType: string) => {
  switch (fileType?.toLowerCase()) {
    case 'docx':
    case 'doc':
      return '#1976d2'
    case 'xlsx':
    case 'xls':
      return '#388e3c'
    case 'pptx':
    case 'ppt':
      return '#f57c00'
    default:
      return '#616161'
  }
}

const getFileTypeTagType = (fileType: string) => {
  switch (fileType?.toLowerCase()) {
    case 'docx':
    case 'doc':
      return 'primary'
    case 'xlsx':
    case 'xls':
      return 'success'
    case 'pptx':
    case 'ppt':
      return 'warning'
    default:
      return 'info'
  }
}
</script>

<style scoped>
.documents-container {
  padding: 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  background: white;
  padding: 24px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.header-content h1 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-icon {
  font-size: 28px;
  color: #409eff;
}

.page-description {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
}

.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
}

.documents-content {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.documents-card {
  border: none;
  box-shadow: none;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #2c3e50;
}

.document-count {
  color: #7f8c8d;
  font-size: 14px;
}

.documents-table {
  --el-table-border-color: #f1f2f6;
}

.documents-table :deep(.el-table__row:hover > td) {
  background-color: #f8f9ff !important;
}

.document-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-icon {
  font-size: 16px;
}

.title-text {
  color: #2c3e50;
  font-weight: 500;
}

.upload-form {
  padding: 20px 0;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: #f8f9ff;
  border-radius: 8px;
  border: 1px solid #e1e6ff;
}

.file-name {
  font-weight: 500;
  color: #2c3e50;
}

.file-size {
  color: #7f8c8d;
  font-size: 12px;
}

.dialog-footer {
  display: flex;
  gap: 12px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .documents-container {
    padding: 10px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .toolbar {
    justify-content: flex-start;
    flex-wrap: wrap;
  }
  
  .documents-table {
    font-size: 14px;
  }
}
</style> 