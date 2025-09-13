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
          
          <el-form-item label="文档标题">
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
          
          <el-form-item label="文档分类">
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
          
          <el-form-item label="上传文件" prop="project_id">
            <FileUpload
              :project-id="uploadForm.project_id"
              :title="uploadForm.title"
              :description="uploadForm.description"
              :category="uploadForm.category"
              :multiple="true"
              :limit="10"
              :max-size="50"
              @success="handleUploadSuccess"
              @error="handleUploadError"
            />
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { researchService } from '@/services/api'
import type { ResearchProject } from '@/types'
import { ElMessage, type FormInstance } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import FileUpload from '@/components/common/FileUpload.vue'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const projects = ref<ResearchProject[]>([])
const categories = ref(['技术文档', '研究报告', '数据集', '代码示例', '参考资料', '其他'])
const uploadFormRef = ref<FormInstance>()

const uploadForm = ref({
  project_id: '',
  title: '',
  description: '',
  category: ''
})

const uploadRules = {
  project_id: [{ required: true, message: '请选择课题', trigger: 'change' }]
}

// 方法
const fetchProjects = async () => {
  try {
    const response = await researchService.getProjects()
    const data = response.data || response
    projects.value = data.projects || data.items || data || []
  } catch (error) {
    console.error('获取课题列表失败:', error)
    ElMessage.error('获取课题列表失败')
  }
}

const handleUploadSuccess = (files: any[]) => {
  ElMessage.success('文档上传成功')
  // 延迟跳转，让用户看到成功消息
  setTimeout(() => {
    router.push('/documents')
  }, 1500)
}

const handleUploadError = (error: any) => {
  console.error('文档上传失败:', error)
  ElMessage.error('文档上传失败')
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