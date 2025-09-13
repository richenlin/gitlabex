<template>
  <div class="topic-create">
    <el-card>
      <template #header>
        <div class="card-header">
          <h2>创建话题</h2>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        @submit.prevent="handleSubmit"
      >
        <el-form-item label="话题标题" prop="title">
          <el-input
            v-model="form.title"
            placeholder="请输入话题标题"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="所属项目" prop="projectId">
          <el-select
            v-model="form.projectId"
            placeholder="选择所属项目（可选）"
            style="width: 100%"
            clearable
          >
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="话题内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="8"
            placeholder="请输入话题内容"
            maxlength="5000"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="优先级" prop="priority">
          <el-radio-group v-model="form.priority">
            <el-radio label="low">低</el-radio>
            <el-radio label="medium">普通</el-radio>
            <el-radio label="high">高</el-radio>
            <el-radio label="urgent">紧急</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="标签">
          <el-tag
            v-for="tag in form.tags"
            :key="tag"
            closable
            @close="removeTag(tag)"
            style="margin-right: 8px; margin-bottom: 8px"
          >
            {{ tag }}
          </el-tag>
          <el-input
            v-if="tagInputVisible"
            ref="tagInputRef"
            v-model="tagInputValue"
            size="small"
            style="width: 120px"
            @keyup.enter="addTag"
            @blur="addTag"
          />
          <el-button
            v-else
            size="small"
            @click="showTagInput"
          >
            + 添加标签
          </el-button>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            @click="handleSubmit"
            :loading="submitting"
          >
            创建话题
          </el-button>
          <el-button @click="handleCancel">
            取消
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { topicService, researchService } from '@/services/api'

const router = useRouter()

// 表单数据
const form = reactive({
  title: '',
  projectId: '',
  content: '',
  priority: 'medium' as 'low' | 'medium' | 'high' | 'urgent',
  tags: [] as string[]
})

// 项目列表
const projects = ref<any[]>([])

// 标签输入
const tagInputVisible = ref(false)
const tagInputValue = ref('')
const tagInputRef = ref()

// 表单引用
const formRef = ref<FormInstance>()
const submitting = ref(false)

// 表单验证规则
const rules: FormRules = {
  title: [
    { required: true, message: '请输入话题标题', trigger: 'blur' },
    { min: 2, max: 200, message: '话题标题长度在 2 到 200 个字符', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入话题内容', trigger: 'blur' },
    { min: 10, max: 5000, message: '话题内容长度在 10 到 5000 个字符', trigger: 'blur' }
  ]
}

// 方法
const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true

    const data = {
      title: form.title,
      content: form.content,
      priority: form.priority,
      tags: form.tags,
      project_id: form.projectId || undefined
    }

    await topicService.createTopic(data)
    ElMessage.success('话题创建成功')
    router.push('/topics')
  } catch (error) {
    console.error('创建话题失败:', error)
    ElMessage.error('创建话题失败')
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  router.back()
}

const removeTag = (tag: string) => {
  const index = form.tags.indexOf(tag)
  if (index > -1) {
    form.tags.splice(index, 1)
  }
}

const showTagInput = () => {
  tagInputVisible.value = true
  setTimeout(() => {
    tagInputRef.value?.focus()
  }, 100)
}

const addTag = () => {
  const tag = tagInputValue.value.trim()
  if (tag && !form.tags.includes(tag)) {
    form.tags.push(tag)
  }
  tagInputVisible.value = false
  tagInputValue.value = ''
}

const loadProjects = async () => {
  try {
    const response = await researchService.getProjects()
    projects.value = response.data || []
  } catch (error) {
    console.error('加载项目列表失败:', error)
  }
}

// 生命周期
onMounted(() => {
  loadProjects()
})
</script>

<style scoped>
.topic-create {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.card-header h2 {
  margin: 0;
  color: #303133;
}
</style>
