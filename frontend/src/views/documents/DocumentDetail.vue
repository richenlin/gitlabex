<template>
  <div class="document-detail" v-loading="loading">
    <div v-if="document" class="document-container">
      <!-- 文档头部信息 -->
      <div class="document-header">
        <div class="header-left">
          <h1 class="document-title">{{ document.title }}</h1>
          <div class="document-badges">
            <el-tag :type="getStatusType(document.status)" size="large">
              {{ getStatusText(document.status) }}
            </el-tag>
          </div>
        </div>
        <div class="header-actions">
          <el-button v-if="canEdit" @click="editDocument">
            <el-icon><Edit /></el-icon>
            编辑
          </el-button>
          <el-button v-if="canRequestEdit" type="primary" @click="requestEdit">
            <el-icon><EditPen /></el-icon>
            申请修改
          </el-button>
          <el-button @click="downloadDocument">
            <el-icon><Download /></el-icon>
            下载
          </el-button>
        </div>
      </div>

      <!-- 文档基本信息 -->
      <el-row :gutter="20" class="document-info">
        <el-col :span="16">
          <el-card title="文档信息">
            <div class="document-description">
              <h3>文档描述</h3>
              <div class="content">{{ document.description || '暂无描述' }}</div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="8">
          <el-card title="文档属性">
            <div class="info-item">
              <span class="label">上传者：</span>
              <span class="value">{{ document.uploader?.name }}</span>
            </div>
            <div class="info-item">
              <span class="label">关联课题：</span>
              <el-link 
                :href="`/scenes/${document.project?.id}`" 
                type="primary"
                :underline="false"
              >
                {{ document.project?.name }}
              </el-link>
            </div>
            <div class="info-item">
              <span class="label">文档分类：</span>
              <span class="value">{{ document.category }}</span>
            </div>
            <div class="info-item">
              <span class="label">文件类型：</span>
              <span class="value">{{ getFileTypeText(document.file_type) }}</span>
            </div>
            <div class="info-item">
              <span class="label">文件大小：</span>
              <span class="value">{{ formatFileSize(document.file_size) }}</span>
            </div>
            <div class="info-item">
              <span class="label">上传时间：</span>
              <span class="value">{{ formatDate(document.created_at) }}</span>
            </div>
            <div class="info-item">
              <span class="label">下载次数：</span>
              <span class="value">{{ document.download_count }}次</span>
            </div>
            <div class="info-item" v-if="document.tags?.length">
              <span class="label">标签：</span>
              <div class="tags">
                <el-tag 
                  v-for="tag in document.tags" 
                  :key="tag" 
                  size="small"
                  class="tag"
                >
                  {{ tag }}
                </el-tag>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 编辑申请对话框 -->
    <el-dialog v-model="editRequestVisible" title="申请编辑文档" width="600px">
      <el-form
        ref="editRequestFormRef"
        :model="editRequestForm"
        :rules="editRequestRules"
        label-width="100px"
      >
        <el-form-item label="文档标题">
          <el-input 
            v-model="editRequestForm.title" 
            placeholder="建议的新标题（可选）"
          />
        </el-form-item>
        
        <el-form-item label="文档描述">
          <el-input
            v-model="editRequestForm.description"
            type="textarea"
            :rows="3"
            placeholder="建议的新描述（可选）"
          />
        </el-form-item>
        
        <el-form-item label="申请原因" prop="reason">
          <el-input
            v-model="editRequestForm.reason"
            type="textarea"
            :rows="4"
            placeholder="请说明申请修改的原因..."
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="editRequestVisible = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="submitEditRequest"
            :loading="submittingRequest"
          >
            提交申请
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { documentService } from '@/services/api'
import type { Document } from '@/types'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Edit, EditPen, Download } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const document = ref<Document | null>(null)
const loading = ref(false)
const submittingRequest = ref(false)
const editRequestVisible = ref(false)

const editRequestFormRef = ref<FormInstance>()
const editRequestForm = ref({
  title: '',
  description: '',
  reason: ''
})

const editRequestRules = {
  reason: [{ required: true, message: '请说明申请原因', trigger: 'blur' }]
}

// 计算属性
const documentId = computed(() => route.params.id as string)

const isTeacher = computed(() => 
  userStore.hasRole('teacher') || userStore.hasRole('admin')
)

const isStudent = computed(() => userStore.hasRole('student'))

const canEdit = computed(() => {
  if (!document.value) return false
  return isTeacher.value && (
    userStore.hasRole('admin') || 
    document.value.uploader_id === userStore.user?.id
  )
})

const canRequestEdit = computed(() => {
  if (!document.value) return false
  return isStudent.value && document.value.status === 'approved'
})

// 方法
const fetchDocument = async () => {
  loading.value = true
  try {
    const response = await documentService.getDocumentById(documentId.value)
    document.value = response.data
  } catch (error) {
    ElMessage.error('获取文档详情失败')
    router.go(-1)
  } finally {
    loading.value = false
  }
}

const editDocument = () => {
  ElMessage.info('直接编辑功能待实现')
}

const requestEdit = () => {
  editRequestForm.value = {
    title: document.value?.title || '',
    description: document.value?.description || '',
    reason: ''
  }
  editRequestVisible.value = true
}

const submitEditRequest = async () => {
  if (!editRequestFormRef.value) return
  
  try {
    await editRequestFormRef.value.validate()
    submittingRequest.value = true
    
    const proposedChanges: any = {}
    if (editRequestForm.value.title !== document.value?.title) {
      proposedChanges.title = editRequestForm.value.title
    }
    if (editRequestForm.value.description !== document.value?.description) {
      proposedChanges.description = editRequestForm.value.description
    }
    
    await documentService.submitEditRequest(documentId.value, {
      proposed_changes: proposedChanges,
      reason: editRequestForm.value.reason
    })
    
    ElMessage.success('编辑申请提交成功')
    editRequestVisible.value = false
    
  } catch (error) {
    ElMessage.error('编辑申请提交失败')
  } finally {
    submittingRequest.value = false
  }
}

const downloadDocument = () => {
  if (!document.value) return
  
  const downloadUrl = `${import.meta.env.VITE_API_BASE_URL}/api/v1/documents/${documentId.value}/download`
  const link = document.createElement('a')
  link.href = downloadUrl
  link.download = document.value.title
  link.click()
  
  if (document.value.download_count !== undefined) {
    document.value.download_count += 1
  }
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    pending: 'warning',
    approved: 'success',
    rejected: 'danger'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    pending: '待审核',
    approved: '已通过',
    rejected: '已拒绝'
  }
  return texts[status] || status
}

const getFileTypeText = (fileType: string) => {
  const types: Record<string, string> = {
    pdf: 'PDF',
    word: 'Word',
    excel: 'Excel',
    ppt: 'PowerPoint',
    image: '图片',
    video: '视频',
    code: '代码',
    other: '其他'
  }
  return types[fileType] || '未知'
}

const formatFileSize = (size: number) => {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`
}

const formatDate = (date: string | Date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

// 生命周期
onMounted(() => {
  fetchDocument()
})
</script>

<style scoped>
.document-detail {
  padding: 20px;
}

.document-container {
  max-width: 1200px;
  margin: 0 auto;
}

.document-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e8e8e8;
}

.document-title {
  margin: 0 0 12px 0;
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.document-badges {
  display: flex;
  gap: 8px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.document-info {
  margin-bottom: 24px;
}

.document-description h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.content {
  line-height: 1.6;
  color: #606266;
}

.info-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
}

.info-item .label {
  font-weight: 500;
  color: #909399;
  min-width: 80px;
}

.info-item .value {
  color: #303133;
}

.tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.tag {
  margin: 0;
}

@media (max-width: 768px) {
  .document-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .header-actions {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>