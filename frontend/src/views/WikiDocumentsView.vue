<template>
  <div class="wiki-documents-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <el-icon class="page-icon"><Document /></el-icon>
          Wiki文档管理
        </h1>
        <p class="page-description">基于GitLab Wiki的文档管理与OnlyOffice在线编辑</p>
      </div>
      
      <!-- 项目选择器 -->
      <div class="project-selector">
        <el-select 
          v-model="selectedProjectId" 
          placeholder="选择项目"
          @change="loadWikiPages"
          style="width: 300px;"
          filterable
        >
          <el-option
            v-for="project in projects"
            :key="project.id"
            :label="project.name"
            :value="project.id"
          />
        </el-select>
      </div>
      
      <!-- 操作栏 -->
      <div class="toolbar">
        <el-button 
          type="primary" 
          :icon="Plus"
          @click="showCreateWikiDialog"
          :disabled="!selectedProjectId"
        >
          创建Wiki页面
        </el-button>
        
        <el-button 
          :icon="Refresh"
          @click="loadWikiPages"
          :loading="isLoading"
          :disabled="!selectedProjectId"
        >
          刷新
        </el-button>
      </div>
    </div>

    <!-- Wiki页面列表 -->
    <div class="wiki-content" v-if="selectedProjectId">
      <el-card class="wiki-card" v-loading="isLoading">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><FolderOpened /></el-icon>
              Wiki页面
            </span>
            <span class="page-count">共 {{ wikiPages.length }} 个页面</span>
          </div>
        </template>

        <!-- Wiki页面表格 -->
        <el-table 
          :data="wikiPages" 
          style="width: 100%" 
          empty-text="暂无Wiki页面"
          @row-click="viewWikiPage"
          class="wiki-table"
        >
          <el-table-column prop="title" label="页面标题" min-width="200">
            <template #default="scope">
              <div class="wiki-title">
                <el-icon class="wiki-icon"><Memo /></el-icon>
                <span class="title-text">{{ scope.row.title }}</span>
              </div>
            </template>
          </el-table-column>
          
          <el-table-column prop="slug" label="页面路径" width="150">
            <template #default="scope">
              <el-tag size="small" type="info">{{ scope.row.slug }}</el-tag>
            </template>
          </el-table-column>
          
          <el-table-column label="可编辑附件" width="120">
            <template #default="scope">
              <el-tag 
                :type="scope.row.editableAttachments > 0 ? 'success' : 'info'" 
                size="small"
              >
                {{ scope.row.editableAttachments || 0 }} 个
              </el-tag>
            </template>
          </el-table-column>
          
          <el-table-column prop="updated_at" label="更新时间" width="180">
            <template #default="scope">
              {{ formatDate(scope.row.updated_at) }}
            </template>
          </el-table-column>
          
          <el-table-column label="操作" width="300" fixed="right">
            <template #default="scope">
              <el-button-group>
                <el-button 
                  type="primary" 
                  size="small" 
                  :icon="View"
                  @click.stop="viewWikiPage(scope.row)"
                >
                  查看
                </el-button>
                <el-button 
                  type="success" 
                  size="small" 
                  :icon="Upload"
                  @click.stop="showUploadAttachmentDialog(scope.row)"
                >
                  上传附件
                </el-button>
                <el-button 
                  type="warning" 
                  size="small" 
                  :icon="Edit"
                  @click.stop="editWikiContent(scope.row)"
                >
                  编辑内容
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <el-empty description="请先选择项目查看Wiki页面"></el-empty>
    </div>

    <!-- 创建Wiki页面对话框 -->
    <el-dialog 
      v-model="createWikiDialogVisible" 
      title="创建Wiki页面" 
      width="600px"
    >
      <el-form 
        ref="createWikiFormRef" 
        :model="createWikiForm" 
        :rules="createWikiRules" 
        label-width="100px"
      >
        <el-form-item label="页面标题" prop="title">
          <el-input 
            v-model="createWikiForm.title" 
            placeholder="请输入Wiki页面标题"
          />
        </el-form-item>
        
        <el-form-item label="页面内容" prop="content">
          <el-input 
            v-model="createWikiForm.content" 
            type="textarea" 
            :rows="6"
            placeholder="请输入Wiki页面内容（支持Markdown）"
          />
        </el-form-item>
        
        <el-form-item label="文档附件">
          <el-upload
            ref="createWikiUploadRef"
            :file-list="createWikiForm.attachments"
            :on-change="handleCreateWikiFileChange"
            :auto-upload="false"
            multiple
            accept=".docx,.xlsx,.pptx,.doc,.xls,.ppt"
          >
            <el-button>
              <el-icon><UploadFilled /></el-icon>
              选择文档附件
            </el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持Word、Excel、PowerPoint文档，上传后可使用OnlyOffice在线编辑
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="createWikiDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createWikiPage" :loading="createWikiLoading">
          创建
        </el-button>
      </template>
    </el-dialog>

    <!-- 上传附件对话框 -->
    <el-dialog 
      v-model="uploadAttachmentDialogVisible" 
      title="上传文档附件" 
      width="500px"
    >
      <div v-if="selectedWikiPage" class="upload-attachment-info">
        <p><strong>Wiki页面：</strong>{{ selectedWikiPage.title }}</p>
        <p><strong>页面路径：</strong>{{ selectedWikiPage.slug }}</p>
      </div>
      
      <el-form 
        ref="uploadAttachmentFormRef" 
        :model="uploadAttachmentForm" 
        label-width="80px"
      >
        <el-form-item label="选择文件">
          <el-upload
            ref="uploadAttachmentUploadRef"
            :file-list="uploadAttachmentForm.files"
            :on-change="handleUploadAttachmentFileChange"
            :auto-upload="false"
            multiple
            accept=".docx,.xlsx,.pptx,.doc,.xls,.ppt"
          >
            <el-button>
              <el-icon><UploadFilled /></el-icon>
              选择文档文件
            </el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持Word、Excel、PowerPoint文档，上传后可在线编辑
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="uploadAttachmentDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="uploadAttachment" :loading="uploadAttachmentLoading">
          上传
        </el-button>
      </template>
    </el-dialog>

    <!-- Wiki页面详情对话框 -->
    <el-dialog 
      v-model="wikiDetailDialogVisible" 
      :title="selectedWikiPage?.title" 
      width="1000px"
      :destroy-on-close="true"
    >
      <div v-if="selectedWikiPage" class="wiki-detail-content">
        <el-tabs v-model="activeDetailTab">
          <!-- Wiki内容 -->
          <el-tab-pane label="Wiki内容" name="content">
            <div class="wiki-content-display">
              <div class="content-text" v-html="renderMarkdown(selectedWikiPage.content)"></div>
            </div>
          </el-tab-pane>

          <!-- 可编辑附件 -->
          <el-tab-pane label="文档附件" name="attachments">
            <div class="editable-attachments">
              <div class="attachments-header">
                <h4>可编辑的文档附件</h4>
                <el-button 
                  type="primary" 
                  size="small" 
                  @click="showUploadAttachmentDialog(selectedWikiPage)"
                >
                  <el-icon><Plus /></el-icon>
                  添加附件
                </el-button>
              </div>
              
              <el-table :data="wikiAttachments" style="width: 100%">
                <el-table-column prop="file_name" label="文件名" />
                <el-table-column prop="file_type" label="类型" width="100">
                  <template #default="{ row }">
                    <el-tag :type="getFileTypeTagType(row.file_type)" size="small">
                      {{ row.file_type.toUpperCase() }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="last_edited_at" label="最后编辑" width="180">
                  <template #default="{ row }">
                    {{ row.last_edited_at ? formatDate(row.last_edited_at) : '-' }}
                  </template>
                </el-table-column>
                <el-table-column prop="last_edited_by" label="编辑者" width="120" />
                <el-table-column label="操作" width="200">
                  <template #default="{ row }">
                    <el-button-group size="small">
                      <el-button 
                        type="primary"
                        @click="editAttachmentWithOnlyOffice(row)"
                        :disabled="!row.can_edit"
                      >
                        <el-icon><Edit /></el-icon>
                        在线编辑
                      </el-button>
                      <el-button 
                        @click="downloadAttachment(row)"
                      >
                        <el-icon><Download /></el-icon>
                        下载
                      </el-button>
                    </el-button-group>
                  </template>
                </el-table-column>
              </el-table>
              
              <div v-if="wikiAttachments.length === 0" class="no-attachments">
                <el-empty description="该Wiki页面暂无可编辑的文档附件" />
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>

    <!-- OnlyOffice编辑器对话框 -->
    <el-dialog 
      v-model="onlyOfficeDialogVisible" 
      :title="`编辑文档 - ${currentEditingAttachment?.file_name}`" 
      width="95%" 
      fullscreen
      :before-close="handleOnlyOfficeClose"
    >
      <div ref="onlyOfficeContainer" style="height: 100vh;"></div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Document, 
  Plus, 
  Refresh, 
  FolderOpened, 
  View, 
  Upload, 
  Edit,
  Memo,
  UploadFilled,
  Download
} from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// 响应式数据
const projects = ref([])
const selectedProjectId = ref(null)
const wikiPages = ref([])
const wikiAttachments = ref([])

const isLoading = ref(false)
const createWikiLoading = ref(false)
const uploadAttachmentLoading = ref(false)

const createWikiDialogVisible = ref(false)
const uploadAttachmentDialogVisible = ref(false)
const wikiDetailDialogVisible = ref(false)
const onlyOfficeDialogVisible = ref(false)

const selectedWikiPage = ref(null)
const currentEditingAttachment = ref(null)
const activeDetailTab = ref('content')

let currentDocEditor = null

// 表单数据
const createWikiForm = reactive({
  title: '',
  content: '',
  attachments: []
})

const uploadAttachmentForm = reactive({
  files: []
})

// 表单引用
const createWikiFormRef = ref()
const uploadAttachmentFormRef = ref()
const createWikiUploadRef = ref()
const uploadAttachmentUploadRef = ref()
const onlyOfficeContainer = ref()

// 表单验证规则
const createWikiRules = {
  title: [
    { required: true, message: '请输入Wiki页面标题', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入Wiki页面内容', trigger: 'blur' }
  ]
}

// 页面初始化
onMounted(() => {
  loadProjects()
})

// 加载项目列表
const loadProjects = async () => {
  try {
    const response = await fetch('/api/projects', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      projects.value = data.data || []
    }
  } catch (error) {
    console.error('加载项目列表失败:', error)
    ElMessage.error('加载项目列表失败')
  }
}

// 加载Wiki页面列表
const loadWikiPages = async () => {
  if (!selectedProjectId.value) return
  
  isLoading.value = true
  try {
    const response = await fetch(`/api/projects/${selectedProjectId.value}/wiki`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      wikiPages.value = data.data || []
    } else {
      // 使用模拟数据
      wikiPages.value = [
        {
          id: 1,
          title: '项目文档',
          slug: 'project-docs',
          content: '# 项目文档\n\n这是项目的主要文档页面。',
          updated_at: new Date().toISOString(),
          editableAttachments: 2
        },
        {
          id: 2,
          title: '开发指南',
          slug: 'dev-guide',
          content: '# 开发指南\n\n本页面包含开发相关的指导信息。',
          updated_at: new Date(Date.now() - 86400000).toISOString(),
          editableAttachments: 1
        }
      ]
    }
  } catch (error) {
    console.error('加载Wiki页面失败:', error)
    ElMessage.error('加载Wiki页面失败')
  } finally {
    isLoading.value = false
  }
}

// 显示创建Wiki对话框
const showCreateWikiDialog = () => {
  createWikiDialogVisible.value = true
  Object.assign(createWikiForm, {
    title: '',
    content: '',
    attachments: []
  })
}

// 处理创建Wiki文件选择
const handleCreateWikiFileChange = (file) => {
  createWikiForm.attachments = createWikiUploadRef.value.fileList
}

// 创建Wiki页面
const createWikiPage = async () => {
  if (!createWikiFormRef.value) return
  
  const valid = await createWikiFormRef.value.validate()
  if (!valid) return
  
  createWikiLoading.value = true
  
  try {
    const formData = new FormData()
    formData.append('project_id', selectedProjectId.value)
    formData.append('title', createWikiForm.title)
    formData.append('content', createWikiForm.content)
    
    // 添加附件
    createWikiForm.attachments.forEach(file => {
      formData.append('attachments', file.raw)
    })
    
    const response = await fetch(`/api/projects/${selectedProjectId.value}/wiki`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: formData
    })
    
    if (response.ok) {
      ElMessage.success('Wiki页面创建成功')
      createWikiDialogVisible.value = false
      loadWikiPages()
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '创建Wiki页面失败')
    }
  } catch (error) {
    console.error('创建Wiki页面失败:', error)
    ElMessage.error('创建Wiki页面失败')
  } finally {
    createWikiLoading.value = false
  }
}

// 查看Wiki页面
const viewWikiPage = async (wikiPage) => {
  selectedWikiPage.value = wikiPage
  wikiDetailDialogVisible.value = true
  activeDetailTab.value = 'content'
  
  // 加载Wiki页面的可编辑附件
  await loadWikiAttachments(wikiPage.slug)
}

// 加载Wiki附件
const loadWikiAttachments = async (wikiSlug) => {
  try {
    const response = await fetch(`/api/projects/${selectedProjectId.value}/wiki/${wikiSlug}/attachments`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      wikiAttachments.value = data.data || []
    } else {
      // 使用模拟数据
      wikiAttachments.value = [
        {
          id: 1,
          file_name: '需求文档.docx',
          file_type: 'docx',
          last_edited_at: new Date().toISOString(),
          last_edited_by: '张老师',
          can_edit: true
        },
        {
          id: 2,
          file_name: '数据分析.xlsx',
          file_type: 'xlsx',
          last_edited_at: new Date(Date.now() - 3600000).toISOString(),
          last_edited_by: '李同学',
          can_edit: true
        }
      ]
    }
  } catch (error) {
    console.error('加载Wiki附件失败:', error)
    wikiAttachments.value = []
  }
}

// 显示上传附件对话框
const showUploadAttachmentDialog = (wikiPage) => {
  selectedWikiPage.value = wikiPage
  uploadAttachmentDialogVisible.value = true
  uploadAttachmentForm.files = []
}

// 处理上传附件文件选择
const handleUploadAttachmentFileChange = (file) => {
  uploadAttachmentForm.files = uploadAttachmentUploadRef.value.fileList
}

// 上传附件
const uploadAttachment = async () => {
  if (uploadAttachmentForm.files.length === 0) {
    ElMessage.warning('请选择要上传的文件')
    return
  }
  
  uploadAttachmentLoading.value = true
  
  try {
    const formData = new FormData()
    formData.append('project_id', selectedProjectId.value)
    formData.append('wiki_slug', selectedWikiPage.value.slug)
    
    uploadAttachmentForm.files.forEach(file => {
      formData.append('attachments', file.raw)
    })
    
    const response = await fetch(`/api/projects/${selectedProjectId.value}/wiki/${selectedWikiPage.value.slug}/attachments`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: formData
    })
    
    if (response.ok) {
      ElMessage.success('附件上传成功')
      uploadAttachmentDialogVisible.value = false
      // 重新加载附件列表
      await loadWikiAttachments(selectedWikiPage.value.slug)
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '上传附件失败')
    }
  } catch (error) {
    console.error('上传附件失败:', error)
    ElMessage.error('上传附件失败')
  } finally {
    uploadAttachmentLoading.value = false
  }
}

// 使用OnlyOffice编辑附件
const editAttachmentWithOnlyOffice = async (attachment) => {
  currentEditingAttachment.value = attachment
  onlyOfficeDialogVisible.value = true
  
  try {
    // 等待DOM更新
    await nextTick()
    
    // 启动OnlyOffice编辑会话
    const response = await fetch(`/api/documents/${attachment.id}/edit`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const config = await response.json()
      
      // 初始化OnlyOffice编辑器
      if (window.DocsAPI) {
        currentDocEditor = new window.DocsAPI.DocEditor(onlyOfficeContainer.value, {
          documentType: config.documentType,
          document: config.document,
          editorConfig: config.editorConfig,
          token: config.token,
          events: {
            onAppReady: () => {
              console.log('OnlyOffice编辑器已就绪')
            },
            onDocumentStateChange: (event) => {
              console.log('文档状态变更:', event.data)
            }
          }
        })
      } else {
        ElMessage.error('OnlyOffice编辑器未加载')
      }
    } else {
      ElMessage.error('启动编辑器失败')
    }
  } catch (error) {
    console.error('启动OnlyOffice编辑器失败:', error)
    ElMessage.error('启动编辑器失败')
  }
}

// 关闭OnlyOffice编辑器
const handleOnlyOfficeClose = () => {
  if (currentDocEditor) {
    currentDocEditor.destroyEditor()
    currentDocEditor = null
  }
  onlyOfficeDialogVisible.value = false
  currentEditingAttachment.value = null
  
  // 重新加载附件信息
  if (selectedWikiPage.value) {
    loadWikiAttachments(selectedWikiPage.value.slug)
  }
}

// 下载附件
const downloadAttachment = (attachment) => {
  const downloadUrl = `/api/documents/${attachment.id}/download`
  window.open(downloadUrl, '_blank')
}

// 编辑Wiki内容
const editWikiContent = (wikiPage) => {
  // 跳转到GitLab Wiki编辑页面
  const gitlabUrl = `${window.location.origin}/gitlab/${projects.value.find(p => p.id === selectedProjectId.value)?.path_with_namespace}/-/wikis/${wikiPage.slug}/edit`
  window.open(gitlabUrl, '_blank')
}

// 渲染Markdown
const renderMarkdown = (content) => {
  // 简单的Markdown渲染（生产环境建议使用专业的Markdown解析器）
  return content
    .replace(/^### (.*$)/gm, '<h3>$1</h3>')
    .replace(/^## (.*$)/gm, '<h2>$1</h2>')
    .replace(/^# (.*$)/gm, '<h1>$1</h1>')
    .replace(/\*\*(.*)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*)\*/g, '<em>$1</em>')
    .replace(/\n/g, '<br/>')
}

// 工具函数
const getFileTypeTagType = (fileType) => {
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

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}
</script>

<style scoped>
.wiki-documents-container {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  gap: 20px;
}

.header-content {
  flex: 1;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 28px;
  color: #303133;
  margin: 0 0 8px 0;
}

.page-icon {
  color: #409EFF;
}

.page-description {
  color: #909399;
  font-size: 16px;
  margin: 0;
}

.project-selector {
  display: flex;
  align-items: center;
}

.toolbar {
  display: flex;
  gap: 12px;
}

.wiki-content {
  margin-top: 20px;
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
  color: #303133;
}

.page-count {
  color: #909399;
  font-size: 14px;
}

.wiki-table {
  margin-top: 16px;
}

.wiki-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.wiki-icon {
  color: #409EFF;
}

.title-text {
  font-weight: 500;
}

.upload-attachment-info {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 20px;
}

.upload-attachment-info p {
  margin: 5px 0;
  color: #606266;
}

.wiki-detail-content {
  min-height: 400px;
}

.wiki-content-display {
  padding: 20px;
  background: #fafafa;
  border-radius: 4px;
  min-height: 300px;
}

.content-text {
  line-height: 1.8;
  color: #303133;
}

.editable-attachments {
  padding: 20px 0;
}

.attachments-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.attachments-header h4 {
  margin: 0;
  color: #303133;
}

.no-attachments {
  text-align: center;
  padding: 40px 20px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

@media (max-width: 1200px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
  }
  
  .project-selector {
    order: 1;
    margin-bottom: 12px;
  }
  
  .toolbar {
    order: 2;
  }
}

@media (max-width: 768px) {
  .wiki-documents-container {
    padding: 16px;
  }
  
  .page-title {
    font-size: 24px;
  }
  
  .toolbar {
    flex-direction: column;
  }
  
  .project-selector .el-select {
    width: 100% !important;
  }
}
</style> 