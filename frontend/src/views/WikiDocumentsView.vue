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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Document, Plus, Refresh, FolderOpened, Memo, View, Upload, Edit, UploadFilled } from '@element-plus/icons-vue'
import type { FormInstance, UploadInstance, UploadUserFile } from 'element-plus'
import { formatDate } from '@/utils/date'
import type { Project, WikiPage, WikiAttachment } from '../types/wiki'
import api from '@/services/api'

// 响应式数据
const selectedProjectId = ref<number | null>(null)
const projects = ref<Project[]>([])
const wikiPages = ref<WikiPage[]>([])
const wikiAttachments = ref<WikiAttachment[]>([])
const isLoading = ref(false)
const createWikiDialogVisible = ref(false)
const uploadAttachmentDialogVisible = ref(false)
const selectedWikiPage = ref<WikiPage | null>(null)
const createWikiLoading = ref(false)
const uploadAttachmentLoading = ref(false)

// 表单引用
const createWikiFormRef = ref<FormInstance>()
const uploadAttachmentFormRef = ref<FormInstance>()
const createWikiUploadRef = ref<UploadInstance>()
const uploadAttachmentUploadRef = ref<UploadInstance>()

// 表单数据
const createWikiForm = reactive({
  title: '',
  content: '',
  attachments: [] as UploadUserFile[]
})

const uploadAttachmentForm = reactive({
  files: [] as UploadUserFile[]
})

// 表单验证规则
const createWikiRules = {
  title: [
    { required: true, message: '请输入页面标题', trigger: 'blur' },
    { min: 2, max: 50, message: '标题长度应在2-50个字符之间', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入页面内容', trigger: 'blur' }
  ]
}

// 加载项目列表
const loadProjects = async () => {
  try {
    const response = await api.get('/projects')
    projects.value = response.data || []
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
    const response = await api.get(`/projects/${selectedProjectId.value}/wiki`)
    wikiPages.value = response.data || []
  } catch (error) {
    console.error('加载Wiki页面失败:', error)
    ElMessage.error('加载Wiki页面失败')
  } finally {
    isLoading.value = false
  }
}

// 处理创建Wiki页面的文件变更
const handleCreateWikiFileChange = (file: UploadUserFile) => {
  createWikiForm.attachments = createWikiUploadRef.value?.fileList || []
}

// 创建Wiki页面
const createWikiPage = async () => {
  if (!createWikiFormRef.value) return
  
  await createWikiFormRef.value.validate(async (valid) => {
  if (!valid) return
  
  createWikiLoading.value = true
  
  try {
    const formData = new FormData()
    formData.append('title', createWikiForm.title)
    formData.append('content', createWikiForm.content)
    
      createWikiForm.attachments.forEach((file) => {
        if (file.raw) {
      formData.append('attachments', file.raw)
        }
    })
    
      const response = await api.post(`/projects/${selectedProjectId.value}/wiki`, formData)
    
      if (response.status === 200) {
      ElMessage.success('Wiki页面创建成功')
      createWikiDialogVisible.value = false
      loadWikiPages()
    } else {
        ElMessage.error(response.data.error || '创建Wiki页面失败')
    }
  } catch (error) {
    console.error('创建Wiki页面失败:', error)
    ElMessage.error('创建Wiki页面失败')
  } finally {
    createWikiLoading.value = false
  }
  })
}

// 显示创建Wiki页面对话框
const showCreateWikiDialog = () => {
  createWikiForm.title = ''
  createWikiForm.content = ''
  createWikiForm.attachments = []
  createWikiDialogVisible.value = true
}

// 查看Wiki页面
const viewWikiPage = async (wikiPage: WikiPage) => {
  selectedWikiPage.value = wikiPage
  await loadWikiAttachments(wikiPage.slug)
}

// 加载Wiki附件列表
const loadWikiAttachments = async (wikiSlug: string) => {
  try {
    const response = await api.get(`/projects/${selectedProjectId.value}/wiki/${wikiSlug}/attachments`)
    wikiAttachments.value = response.data || []
  } catch (error) {
    console.error('加载Wiki附件失败:', error)
    ElMessage.error('加载Wiki附件失败')
  }
}

// 显示上传附件对话框
const showUploadAttachmentDialog = (wikiPage: WikiPage) => {
  selectedWikiPage.value = wikiPage
  uploadAttachmentForm.files = []
  uploadAttachmentDialogVisible.value = true
}

// 处理上传附件的文件变更
const handleUploadAttachmentFileChange = (file: UploadUserFile) => {
  uploadAttachmentForm.files = uploadAttachmentUploadRef.value?.fileList || []
}

// 上传附件
const uploadAttachment = async () => {
  if (!selectedWikiPage.value) return
  
  uploadAttachmentLoading.value = true
  
  try {
    const formData = new FormData()
    uploadAttachmentForm.files.forEach((file) => {
      if (file.raw) {
        formData.append('files', file.raw)
      }
    })
    
    const response = await api.post(
      `/projects/${selectedProjectId.value}/wiki/${selectedWikiPage.value.slug}/attachments`,
      formData
    )
    
    if (response.status === 200) {
      ElMessage.success('附件上传成功')
      uploadAttachmentDialogVisible.value = false
      // 重新加载附件列表
      await loadWikiAttachments(selectedWikiPage.value.slug)
    } else {
      ElMessage.error(response.data.error || '上传附件失败')
    }
  } catch (error) {
    console.error('上传附件失败:', error)
    ElMessage.error('上传附件失败')
  } finally {
    uploadAttachmentLoading.value = false
  }
}

// 编辑Wiki内容
const editWikiContent = (wikiPage: WikiPage) => {
  const project = projects.value.find(p => p.id === selectedProjectId.value)
  window.open(`/gitlab/${project?.path_with_namespace}/-/wikis/${wikiPage.slug}/edit`, '_blank')
}

// 生命周期钩子
onMounted(() => {
  loadProjects()
})
</script>

<style scoped>
.wiki-documents-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.header-content {
  margin-bottom: 20px;
}

.page-title {
  display: flex;
  align-items: center;
  font-size: 24px;
  margin: 0;
}

.page-icon {
  margin-right: 10px;
  font-size: 24px;
}

.page-description {
  color: #666;
  margin: 10px 0;
}

.project-selector {
  margin-bottom: 20px;
}

.toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.wiki-content {
  background: #fff;
  border-radius: 4px;
}

.wiki-card {
  margin-bottom: 20px;
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
}

.page-count {
  color: #666;
  font-size: 14px;
}

.wiki-table {
  margin-top: 10px;
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
  color: #409EFF;
  cursor: pointer;
}

.title-text:hover {
  text-decoration: underline;
}

.empty-state {
  padding: 40px;
  text-align: center;
}

.upload-attachment-info {
  margin-bottom: 20px;
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.upload-attachment-info p {
  margin: 5px 0;
}
</style> 