<template>
  <div class="homework-create">
    <div class="create-container">
      <div class="header">
        <el-button @click="$router.go(-1)" text>
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>创建作业</h1>
      </div>

      <el-card title="作业信息">
        <el-form
          ref="createFormRef"
          :model="createForm"
          :rules="createRules"
          label-width="120px"
        >
          <el-form-item label="作业标题" prop="title">
            <el-input 
              v-model="createForm.title" 
              placeholder="请输入作业标题"
              maxlength="100"
              show-word-limit
            />
          </el-form-item>
          
          <el-form-item label="作业描述" prop="description">
            <el-input
              v-model="createForm.description"
              type="textarea"
              :rows="6"
              placeholder="请输入作业描述和要求..."
              maxlength="2000"
              show-word-limit
            />
          </el-form-item>
          
          <el-form-item label="关联课题" prop="project_id">
            <el-select 
              v-model="createForm.project_id" 
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
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="截止时间" prop="due_date">
                <el-date-picker
                  v-model="createForm.due_date"
                  type="datetime"
                  placeholder="选择截止时间"
                  style="width: 100%"
                  :disabled-date="disabledDate"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="满分" prop="max_grade">
                <el-input-number 
                  v-model="createForm.max_grade" 
                  :min="1" 
                  :max="1000"
                  controls-position="right"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
          </el-row>
          
          <el-form-item label="作业说明">
            <el-input
              v-model="createForm.instructions"
              type="textarea"
              :rows="4"
              placeholder="提交说明、格式要求等..."
              maxlength="1000"
              show-word-limit
            />
          </el-form-item>
          
          <el-form-item label="作业要求">
            <div class="requirements-section">
              <el-input
                v-model="requirementInput"
                placeholder="输入要求后按回车添加"
                @keyup.enter="addRequirement"
              >
                <template #append>
                  <el-button @click="addRequirement">添加</el-button>
                </template>
              </el-input>
              
              <div class="requirements-list" v-if="createForm.requirements.length">
                <div 
                  v-for="(req, index) in createForm.requirements" 
                  :key="index"
                  class="requirement-item"
                >
                  <span>{{ index + 1 }}. {{ req }}</span>
                  <el-button 
                    size="small" 
                    type="danger" 
                    text
                    @click="removeRequirement(index)"
                  >
                    <el-icon><Close /></el-icon>
                  </el-button>
                </div>
              </div>
            </div>
          </el-form-item>
          
          <el-form-item label="标签">
            <el-input
              v-model="tagInput"
              placeholder="输入标签后按回车添加"
              @keyup.enter="addTag"
            />
            <div class="tags-container" v-if="createForm.tags.length">
              <el-tag
                v-for="tag in createForm.tags"
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

        <div class="form-actions">
          <el-button @click="$router.go(-1)">取消</el-button>
          <el-button @click="saveDraft" :loading="saving">保存草稿</el-button>
          <el-button 
            type="primary" 
            @click="publishHomework"
            :loading="publishing"
          >
            发布作业
          </el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { homeworkService, researchService } from '@/services/api'
import type { ResearchProject } from '@/types'
import { ElMessage, type FormInstance } from 'element-plus'
import { ArrowLeft, Close } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const projects = ref<ResearchProject[]>([])
const saving = ref(false)
const publishing = ref(false)
const requirementInput = ref('')
const tagInput = ref('')

const createFormRef = ref<FormInstance>()
const createForm = ref({
  title: '',
  description: '',
  project_id: '',
  due_date: '',
  max_grade: 100,
  instructions: '',
  requirements: [] as string[],
  tags: [] as string[]
})

const createRules = {
  title: [{ required: true, message: '请输入作业标题', trigger: 'blur' }],
  description: [{ required: true, message: '请输入作业描述', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择课题', trigger: 'change' }],
  due_date: [{ required: true, message: '请选择截止时间', trigger: 'change' }],
  max_grade: [{ required: true, message: '请设置满分', trigger: 'blur' }]
}

// 方法
const fetchProjects = async () => {
  try {
    const response = await researchService.getProjects()
    projects.value = response.projects || []
  } catch (error) {
    ElMessage.error('获取课题列表失败')
  }
}

const disabledDate = (date: Date) => {
  return date < new Date()
}

const addRequirement = () => {
  const req = requirementInput.value.trim()
  if (req && !createForm.value.requirements.includes(req)) {
    createForm.value.requirements.push(req)
    requirementInput.value = ''
  }
}

const removeRequirement = (index: number) => {
  createForm.value.requirements.splice(index, 1)
}

const addTag = () => {
  const tag = tagInput.value.trim()
  if (tag && !createForm.value.tags.includes(tag)) {
    createForm.value.tags.push(tag)
    tagInput.value = ''
  }
}

const removeTag = (tag: string) => {
  const index = createForm.value.tags.indexOf(tag)
  if (index !== -1) {
    createForm.value.tags.splice(index, 1)
  }
}

const saveDraft = async () => {
  if (!createFormRef.value) return
  
  try {
    await createFormRef.value.validate()
    saving.value = true
    
    await homeworkService.createHomework({
      ...createForm.value,
      status: 'draft'
    })
    
    ElMessage.success('草稿保存成功')
    router.push('/homeworks')
  } catch (error) {
    ElMessage.error('保存草稿失败')
  } finally {
    saving.value = false
  }
}

const publishHomework = async () => {
  if (!createFormRef.value) return
  
  try {
    await createFormRef.value.validate()
    publishing.value = true
    
    await homeworkService.createHomework({
      ...createForm.value,
      status: 'published'
    })
    
    ElMessage.success('作业发布成功')
    router.push('/homeworks')
  } catch (error) {
    ElMessage.error('发布作业失败')
  } finally {
    publishing.value = false
  }
}

// 生命周期
onMounted(() => {
  fetchProjects()
})
</script>

<style scoped>
.homework-create {
  padding: 20px;
}

.create-container {
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

.requirements-section {
  width: 100%;
}

.requirements-list {
  margin-top: 12px;
}

.requirement-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 8px;
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
</style>
