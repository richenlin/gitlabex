<template>
  <div class="document-upload">
    <div class="upload-container">
      <div class="header">
        <el-button @click="$router.go(-1)" text>
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>上传文档</h1>
      </div>

      <el-card title="文档上传">
        <el-form
          ref="uploadFormRef"
          :model="uploadForm"
          :rules="uploadRules"
          label-width="120px"
        >
          <el-form-item label="关联课题" prop="project_id">
            <el-select 
              v-model="uploadForm.project_id" 
              placeholder="选择课题"
              style="width: 100%"
            >
              <el-option
                v-for="project in projects"
                :key="project.id"
                :label="project.name"
                :value="project.id"
              />
            </el-select>
          </el-form-item>
          
          <el-form-item label="文档标题" prop="title">
            <el-input 
              v-model="uploadForm.title" 
              placeholder="请输入文档标题（可选，默认使用文件名）"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>
          
          <el-form-item label="文档描述">
            <el-input
              v-model="uploadForm.description"
              type="textarea"
              :rows="4"
              placeholder="请输入文档描述..."
              maxlength="500"
              show-word-limit
            />
          </el-form-item>
          
          <el-form-item label="文档分类" prop="category">
            <el-select 
              v-model="uploadForm.category" 
              placeholder="选择分类"
              style="width: 100%"
              allow-create
              filterable
            >
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
              placeholder="输入标签后按回车添加"
              @keyup.enter="addTag"
            />
            <div class="tags-container" v-if="uploadForm.tags.length">
              <el-tag
                v-for="tag in uploadForm.tags"
                :key="tag"
                closable
                @close="removeTag(tag)"
                style="margin: 5px 5px 0 0"
              >
                {{ tag }}
              </el-tag>
            </div>
          </el-form-item>
          
          <el-form-item label="上传文件" prop="files">
            <el-upload
              ref="uploadRef"
              v-model:file-list="fileList"
              :action="uploadAction"
              :headers="uploadHeaders"
              :on-success="handleUploadSuccess"
              :on-error="handleUploadError"
              :on-remove="handleFileRemove"
              :before-upload="beforeUpload"
              multiple
              drag
              :limit="5"
            >
              <el-icon class="el-icon--upload"><upload-filled /></el-icon>
              <div class="el-upload__text">
                将文件拖到此处，或<em>点击上传</em>
              </div>
              <template #tip>
                <div class="el-upload__tip">
                  支持 PDF/DOC/DOCX/PPT/PPTX/XLS/XLSX/ZIP/RAR 等格式，单个文件不超过 100MB，最多5个文件
                </div>
              </template>
            </el-upload>
          </el-form-item>
        </el-form>

        <div class="form-actions">
          <el-button @click="$router.go(-1)">取消</el-button>
          <el-button 
            type="primary" 
            @click="uploadDocument"
            :loading="uploading"
            :disabled="uploadForm.files.length === 0"
          >
            保存文档
          </el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { documentService, researchService } from '@/services/api'
import type { ResearchProject, UploadFile } from '@/types'
import { ElMessage, type FormInstance, type UploadInstance } from 'element-plus'
import { ArrowLeft, UploadFilled } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const projects = ref<ResearchProject[]>([])
const categories = ref(['技术文档', '研究报告', '数据集', '代码示例', '参考资料', '其他'])
const uploading = ref(false)
const tagInput = ref('')
const fileList = ref<UploadFile[]>([])

const uploadFormRef = ref<FormInstance>()
const uploadRef = ref<UploadInstance>()

const uploadForm = ref({
  project_id: '',
  title: '',
  description: '',
  category: '',
  tags: [] as string[],
  files: [] as string[]
})

const uploadRules = {
  project_id: [{ required: true, message: '请选择课题', trigger: 'change' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }]
}

// 计算属性
const uploadAction = computed(() => {
  return `${import.meta.env.VITE_API_BASE_URL}/api/v1/files/upload`
})

const uploadHeaders = computed(() => {
  return {
    'Authorization': `Bearer ${userStore.token}`
  }
})

// 方法
const fetchProjects = async () => {
  try {
    const response = await researchService.getProjects()
    projects.value = response.data?.items || []
  } catch (error) {
    ElMessage.error('获取课题列表失败')
  }
}

const beforeUpload = (file: File) => {
  // 检查文件类型
  const allowedExtensions = ['.pdf', '.doc', '.docx', '.ppt', '.pptx', '.xls', '.xlsx', '.zip', '.rar']
  const fileName = file.name.toLowerCase()
  const isValidType = allowedExtensions.some(ext => fileName.endsWith(ext))
  
  if (!isValidType) {
    ElMessage.error('不支持的文件格式')
    return false
  }
  
  // 检查文件大小
  if (file.size > 100 * 1024 * 1024) {
    ElMessage.error('文件大小不能超过 100MB')
    return false
  }
  
  return true
}

const handleUploadSuccess = (response: any, file: UploadFile) => {
  uploadForm.value.files.push(response.data.file_path)
  
  // 如果没有设置标题，使用第一个文件名
  if (!uploadForm.value.title && uploadForm.value.files.length === 1) {
    uploadForm.value.title = file.name.replace(/\.[^/.]+$/, '')
  }
  
  ElMessage.success(`${file.name} 上传成功`)
}

const handleUploadError = (error: any, file: UploadFile) => {
  ElMessage.error(`${file.name} 上传失败`)
}

const handleFileRemove = (file: UploadFile) => {
  // 从文件列表中移除
  const index = uploadForm.value.files.findIndex(f => f.includes(file.name))
  if (index !== -1) {
    uploadForm.value.files.splice(index, 1)
  }
}

const addTag = () => {
  const tag = tagInput.value.trim()
  if (tag && !uploadForm.value.tags.includes(tag)) {
    uploadForm.value.tags.push(tag)
    tagInput.value = ''
  }
}

const removeTag = (tag: string) => {
  const index = uploadForm.value.tags.indexOf(tag)
  if (index !== -1) {
    uploadForm.value.tags.splice(index, 1)
  }
}

const uploadDocument = async () => {
  if (!uploadFormRef.value) return
  
  try {
    await uploadFormRef.value.validate()
    uploading.value = true
    
    for (const filePath of uploadForm.value.files) {
      await documentService.createDocument({
        project_id: uploadForm.value.project_id,
        title: uploadForm.value.title,
        description: uploadForm.value.description,
        category: uploadForm.value.category,
        tags: uploadForm.value.tags,
        file_path: filePath
      })
    }
    
    ElMessage.success('文档保存成功')
    router.push('/documents')
  } catch (error) {
    ElMessage.error('文档保存失败')
  } finally {
    uploading.value = false
  }
}

// 生命周期
onMounted(() => {
  fetchProjects()
})
</script>

<style scoped>
.document-upload {
  padding: 20px;
}

.upload-container {
  max-width: 800px;
  margin: 0 auto;
}

.header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}

.header h1 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.tags-container {
  margin-top: 8px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.el-upload {
  width: 100%;
}

.el-upload__tip {
  text-align: center;
  color: #606266;
}
</style>