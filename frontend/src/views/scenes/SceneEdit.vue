<template>
  <div class="scene-edit" v-loading="loading">
    <div class="scene-container">
      <!-- 面包屑导航 -->
      <el-breadcrumb class="breadcrumb" separator=">">
        <el-breadcrumb-item :to="{ path: '/scenes' }">课题</el-breadcrumb-item>
        <el-breadcrumb-item :to="{ path: `/scenes/${projectId}` }">{{ project?.name }}</el-breadcrumb-item>
        <el-breadcrumb-item>编辑课题</el-breadcrumb-item>
      </el-breadcrumb>

      <!-- 编辑表单 -->
      <div class="edit-form card">
        <h2>编辑研究课题</h2>
        
        <el-form 
          :model="form" 
          :rules="rules" 
          ref="formRef" 
          label-width="120px"
          @submit.prevent="handleSubmit"
        >
          <el-form-item label="课题名称" prop="name">
            <el-input 
              v-model="form.name" 
              placeholder="请输入课题名称"
              maxlength="100"
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

          <el-form-item label="可见性" prop="is_public">
            <el-radio-group v-model="form.is_public">
              <el-radio :label="true">公开课题</el-radio>
              <el-radio :label="false">专有课题</el-radio>
            </el-radio-group>
            <div class="form-help-text">
              <p>公开课题：所有用户都可以查看和参与</p>
              <p>专有课题：只有指定成员可以查看和参与</p>
            </div>
          </el-form-item>

          <el-form-item label="开始日期" prop="start_date">
            <el-date-picker
              v-model="form.start_date"
              type="date"
              placeholder="选择开始日期"
              value-format="YYYY-MM-DD"
            />
          </el-form-item>

          <el-form-item label="结束日期" prop="end_date">
            <el-date-picker
              v-model="form.end_date"
              type="date"
              placeholder="选择结束日期"
              value-format="YYYY-MM-DD"
            />
          </el-form-item>

          <el-form-item label="课题状态" prop="status">
            <el-select v-model="form.status" placeholder="请选择课题状态">
              <el-option label="活跃" value="active" />
              <el-option label="已归档" value="archived" />
              <el-option label="已暂停" value="suspended" />
            </el-select>
          </el-form-item>

          <el-form-item label="标签" prop="tags">
            <el-select
              v-model="form.tags"
              multiple
              filterable
              allow-create
              default-first-option
              placeholder="请选择或输入标签"
            >
              <el-option
                v-for="tag in commonTags"
                :key="tag"
                :label="tag"
                :value="tag"
              />
            </el-select>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="handleSubmit" :loading="submitting">
              保存修改
            </el-button>
            <el-button @click="handleCancel">取消</el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { researchService } from '@/services/api'
import type { ResearchProject } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import { WarningFilled } from '@element-plus/icons-vue'
import { handleApiError, showSuccess } from '@/utils/errorHandler'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const project = ref<ResearchProject | null>(null)
const loading = ref(false)
const submitting = ref(false)
const deleting = ref(false)
const formRef = ref()

// 计算属性
const projectId = computed(() => route.params.id as string)

const canDelete = ref(false)

// 检查删除权限
const checkDeletePermission = async () => {
  if (!project.value) {
    canDelete.value = false
    return
  }

  try {
    const canDeleteProject = await userStore.checkProjectPermission(projectId.value, 'delete')
    canDelete.value = canDeleteProject
  } catch (error) {
    console.error('删除权限检查失败:', error)
    canDelete.value = false
  }
}

// 表单数据
const form = reactive({
  name: '',
  description: '',
  is_public: true,
  start_date: '',
  end_date: '',
  status: 'active' as 'active' | 'archived' | 'suspended',
  tags: [] as string[]
})

// 常用标签
const commonTags = [
  '机器学习', '人工智能', '数据科学', '算法研究',
  '软件工程', 'Web开发', '移动开发', '系统设计',
  '数据库', '网络安全', '云计算', '区块链'
]

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入课题名称', trigger: 'blur' },
    { min: 2, max: 100, message: '课题名称长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入课题描述', trigger: 'blur' },
    { min: 10, max: 1000, message: '课题描述长度在 10 到 1000 个字符', trigger: 'blur' }
  ],
  is_public: [
    { required: true, message: '请选择可见性', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择课题状态', trigger: 'change' }
  ]
}

// 方法
const fetchProject = async () => {
  loading.value = true
  try {
    const response: any = await researchService.getProject(projectId.value)
    project.value = response
    
    // 检查权限
    await checkDeletePermission()
    
    // 填充表单数据
    if (project.value) {
      form.name = project.value.name
      form.description = project.value.description
      form.is_public = project.value.is_public
      form.status = project.value.status
      form.tags = project.value.tags || []
      
      // 处理日期格式
      if (project.value.created_at) {
        form.start_date = new Date(project.value.created_at).toISOString().split('T')[0]
      }
      if (project.value.updated_at) {
        form.end_date = new Date(project.value.updated_at).toISOString().split('T')[0]
      }
    }
  } catch (error: any) {
    console.error('获取课题详情失败:', error)
    if (error.response?.status === 403) {
      ElMessage.error('权限不足，无法编辑该课题')
      router.push(`/scenes/${projectId.value}`)
    } else if (error.response?.status === 404) {
      ElMessage.error('课题不存在')
      router.push('/scenes')
    } else {
      ElMessage.error('获取课题详情失败')
    }
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    const valid = await formRef.value.validate()
    if (!valid) return

    submitting.value = true

    // 准备更新数据
    const updateData: any = {
      title: form.name,  // 后端可能使用title字段
      description: form.description,
      is_public: form.is_public,
      status: form.status,
      tags: form.tags
    }

    if (form.start_date) {
      updateData.start_date = new Date(form.start_date).toISOString()
    }
    if (form.end_date) {
      updateData.end_date = new Date(form.end_date).toISOString()
    }

    await researchService.updateProject(projectId.value, updateData)
    
    showSuccess('课题更新成功')
    router.push(`/scenes/${projectId.value}`)
  } catch (error: any) {
    console.error('更新课题失败:', error)
    handleApiError(error, '更新课题')
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  router.push(`/scenes/${projectId.value}`)
}

// 权限检查 - 现在通过API进行
const checkPermission = async () => {
  if (!userStore.isLoggedIn) {
    ElMessage.error('请先登录')
    router.push('/auth/login')
    return false
  }

  // 如果已经有项目信息，通过API检查编辑权限
  if (project.value) {
    try {
      const canEdit = await userStore.checkProjectPermission(projectId.value, 'update')
      if (!canEdit) {
        ElMessage.error('权限不足，无法编辑该课题')
        router.push(`/scenes/${projectId.value}`)
        return false
      }
    } catch (error) {
      console.error('权限检查失败:', error)
      ElMessage.error('权限检查失败')
      router.push(`/scenes/${projectId.value}`)
      return false
    }
  }

  return true
}

// 生命周期
onMounted(async () => {
  if (!(await checkPermission())) return
  await fetchProject()
  // 获取项目信息后再次检查权限
  await checkPermission()
})
</script>

<style scoped>
.scene-edit {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.breadcrumb {
  margin-bottom: 20px;
}

.edit-form {
  padding: 30px;
}

.edit-form h2 {
  margin: 0 0 30px 0;
  color: var(--primary-color);
  font-size: 24px;
  font-weight: 600;
}

.form-help-text {
  margin-top: 8px;
  font-size: 12px;
  color: var(--light-text);
}

.form-help-text p {
  margin: 2px 0;
}

.delete-warning {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px 0;
}

.warning-icon {
  font-size: 24px;
  color: var(--warning-color);
  margin-top: 2px;
}

.warning-text {
  color: var(--danger-color);
  font-weight: 500;
  margin-top: 8px;
}

.delete-warning p {
  margin: 4px 0;
}

.delete-warning strong {
  color: var(--primary-color);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .scene-edit {
    padding: 10px;
  }
  
  .edit-form {
    padding: 20px;
  }
  
  .edit-form h2 {
    font-size: 20px;
  }
}
</style>