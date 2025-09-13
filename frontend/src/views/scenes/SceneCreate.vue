<template>
  <div class="scene-create">
    <el-card>
      <template #header>
        <div class="card-header">
          <h2>创建研究课题</h2>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        @submit.prevent="handleSubmit"
      >
        <el-form-item label="课题名称" prop="name">
          <el-input
            v-model="form.name"
            placeholder="请输入课题名称"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="课题描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="6"
            placeholder="请输入课题描述"
            maxlength="1000"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="开始日期" prop="startDate">
          <el-date-picker
            v-model="form.startDate"
            type="date"
            placeholder="选择开始日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="结束日期" prop="endDate">
          <el-date-picker
            v-model="form.endDate"
            type="date"
            placeholder="选择结束日期（可选）"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
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

        <el-form-item label="公开性">
          <el-radio-group v-model="form.isPublic">
            <el-radio :label="true">公开</el-radio>
            <el-radio :label="false">私有</el-radio>
          </el-radio-group>
          <div class="form-tip">公开的课题可以被所有用户查看</div>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            @click="handleSubmit"
            :loading="submitting"
          >
            创建课题
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
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { researchService } from '@/services/api'

const router = useRouter()

// 表单数据
const form = reactive({
  name: '',
  description: '',
  startDate: '',
  endDate: '',
  tags: [] as string[],
  isPublic: true
})

// 标签输入
const tagInputVisible = ref(false)
const tagInputValue = ref('')
const tagInputRef = ref()

// 表单引用
const formRef = ref<FormInstance>()
const submitting = ref(false)

// 表单验证规则
const rules: FormRules = {
  name: [
    { required: true, message: '请输入课题名称', trigger: 'blur' },
    { min: 2, max: 200, message: '课题名称长度在 2 到 200 个字符', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入课题描述', trigger: 'blur' },
    { min: 10, max: 1000, message: '课题描述长度在 10 到 1000 个字符', trigger: 'blur' }
  ],
  startDate: [
    { required: true, message: '请选择开始日期', trigger: 'change' }
  ]
}

// 方法
const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true

    const data = {
      name: form.name,
      description: form.description,
      start_date: form.startDate,
      end_date: form.endDate || null,
      tags: form.tags,
      is_public: form.isPublic
    }

    await researchService.createProject(data)
    ElMessage.success('课题创建成功')
    router.push('/scenes')
  } catch (error) {
    console.error('创建课题失败:', error)
    ElMessage.error('创建课题失败')
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
</script>

<style scoped>
.scene-create {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.card-header h2 {
  margin: 0;
  color: #303133;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
