<template>
  <div class="file-upload">
    <el-upload
      ref="uploadRef"
      :action="uploadUrl"
      :headers="headers"
      :data="uploadData"
      :file-list="fileList"
      :before-upload="beforeUpload"
      :on-success="onSuccess"
      :on-error="onError"
      :on-progress="onProgress"
      :on-remove="onRemove"
      :multiple="multiple"
      :accept="accept"
      :limit="limit"
      :auto-upload="autoUpload"
      drag
    >
      <div class="upload-content">
        <el-icon class="upload-icon"><UploadFilled /></el-icon>
        <div class="upload-text">
          <p>将文件拖拽到此处，或<em>点击上传</em></p>
          <p class="upload-tip">
            支持 {{ acceptText }}，单个文件不超过 {{ maxSizeText }}
          </p>
        </div>
      </div>
    </el-upload>

    <!-- 上传进度 -->
    <div v-if="uploading" class="upload-progress">
      <el-progress 
        :percentage="uploadProgress" 
        :status="uploadStatus"
        :stroke-width="6"
      />
      <p class="progress-text">{{ progressText }}</p>
    </div>

    <!-- 操作按钮 -->
    <div v-if="!autoUpload && fileList.length > 0" class="upload-actions">
      <el-button @click="clearFiles">清空</el-button>
      <el-button type="primary" @click="submitUpload" :loading="uploading">
        上传文件
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox, type UploadInstance, type UploadFile, type UploadFiles, type UploadRawFile } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

interface Props {
  projectId?: string
  multiple?: boolean
  accept?: string
  limit?: number
  maxSize?: number // MB
  autoUpload?: boolean
  category?: string
  title?: string
  description?: string
}

interface Emits {
  (e: 'success', files: any[]): void
  (e: 'error', error: any): void
  (e: 'progress', progress: number): void
}

const props = withDefaults(defineProps<Props>(), {
  multiple: true,
  accept: '.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.md,.jpg,.jpeg,.png,.gif,.zip,.rar,.7z',
  limit: 10,
  maxSize: 50,
  autoUpload: true,
  category: '',
  title: '',
  description: ''
})

const emit = defineEmits<Emits>()

const userStore = useUserStore()
const uploadRef = ref<UploadInstance>()
const fileList = ref<UploadFile[]>([])
const uploading = ref(false)
const uploadProgress = ref(0)
const uploadStatus = ref<'success' | 'exception' | 'warning' | ''>('')

// 计算属性
const uploadUrl = computed(() => {
  return `${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'}/documents/upload`
})

const headers = computed(() => {
  return {
    'Authorization': `Bearer ${userStore.token}`
  }
})

const uploadData = computed(() => {
  return {
    project_id: props.projectId,
    category: props.category,
    title: props.title,
    description: props.description
  }
})

const acceptText = computed(() => {
  const types: { [key: string]: string } = {
    '.pdf': 'PDF',
    '.doc': 'Word',
    '.docx': 'Word',
    '.xls': 'Excel',
    '.xlsx': 'Excel',
    '.ppt': 'PPT',
    '.pptx': 'PPT',
    '.txt': '文本',
    '.md': 'Markdown',
    '.jpg': '图片',
    '.jpeg': '图片',
    '.png': '图片',
    '.gif': '图片',
    '.zip': '压缩包',
    '.rar': '压缩包',
    '.7z': '压缩包'
  }
  
  const extensions = props.accept.split(',')
  const typeNames = [...new Set(extensions.map(ext => types[ext.trim()] || ext.trim()))]
  return typeNames.join('、')
})

const maxSizeText = computed(() => {
  return `${props.maxSize}MB`
})

const progressText = computed(() => {
  if (uploadStatus.value === 'success') {
    return '上传成功'
  } else if (uploadStatus.value === 'exception') {
    return '上传失败'
  } else {
    return `上传中... ${uploadProgress.value}%`
  }
})

// 方法
const beforeUpload = (rawFile: UploadRawFile) => {
  // 检查文件类型
  const fileExtension = '.' + rawFile.name.split('.').pop()?.toLowerCase()
  const allowedTypes = props.accept.split(',').map(type => type.trim())
  
  if (!allowedTypes.includes(fileExtension)) {
    ElMessage.error(`不支持的文件类型: ${fileExtension}`)
    return false
  }

  // 检查文件大小
  const maxSizeBytes = props.maxSize * 1024 * 1024
  if (rawFile.size > maxSizeBytes) {
    ElMessage.error(`文件大小不能超过 ${props.maxSize}MB`)
    return false
  }

  // 检查项目ID
  if (!props.projectId) {
    ElMessage.error('请选择项目')
    return false
  }

  uploading.value = true
  uploadProgress.value = 0
  uploadStatus.value = ''
  
  return true
}

const onProgress = (event: any) => {
  uploadProgress.value = Math.round(event.percent)
  emit('progress', uploadProgress.value)
}

const onSuccess = (response: any, file: UploadFile, files: UploadFiles) => {
  uploading.value = false
  uploadProgress.value = 100
  uploadStatus.value = 'success'
  
  ElMessage.success('文件上传成功')
  emit('success', [response])
  
  // 清空文件列表
  setTimeout(() => {
    fileList.value = []
    uploadProgress.value = 0
    uploadStatus.value = ''
  }, 2000)
}

const onError = (error: any, file: UploadFile, files: UploadFiles) => {
  uploading.value = false
  uploadStatus.value = 'exception'
  
  let errorMessage = '上传失败'
  if (error?.response?.data?.error) {
    errorMessage = error.response.data.error
  } else if (error.message) {
    errorMessage = error.message
  }
  
  ElMessage.error(errorMessage)
  emit('error', error)
}

const onRemove = (file: UploadFile, files: UploadFiles) => {
  // 文件移除处理
  console.log('文件已移除:', file.name)
}

const clearFiles = () => {
  uploadRef.value?.clearFiles()
  fileList.value = []
  uploadProgress.value = 0
  uploadStatus.value = ''
}

const submitUpload = () => {
  if (!props.projectId) {
    ElMessage.error('请选择项目')
    return
  }
  
  uploadRef.value?.submit()
}

// 监听项目ID变化
watch(() => props.projectId, (newVal) => {
  if (!newVal) {
    clearFiles()
  }
})
</script>

<style scoped>
.file-upload {
  width: 100%;
}

.upload-content {
  padding: 40px 20px;
  text-align: center;
}

.upload-icon {
  font-size: 48px;
  color: #c0c4cc;
  margin-bottom: 16px;
}

.upload-text p {
  margin: 8px 0;
  color: #606266;
}

.upload-text em {
  color: #409eff;
  font-style: normal;
}

.upload-tip {
  font-size: 12px;
  color: #909399;
}

.upload-progress {
  margin-top: 20px;
}

.progress-text {
  text-align: center;
  margin-top: 8px;
  font-size: 14px;
  color: #606266;
}

.upload-actions {
  margin-top: 20px;
  text-align: right;
}

.upload-actions .el-button {
  margin-left: 8px;
}

:deep(.el-upload-dragger) {
  border: 2px dashed #d9d9d9;
  border-radius: 6px;
  width: 100%;
  height: auto;
  text-align: center;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: border-color 0.2s cubic-bezier(0.645, 0.045, 0.355, 1);
}

:deep(.el-upload-dragger:hover) {
  border-color: #409eff;
}

:deep(.el-upload-dragger.is-dragover) {
  background-color: rgba(64, 158, 255, 0.06);
  border-color: #409eff;
}
</style>
