<template>
  <div class="scene-list">
    <div class="header">
      <h1>研究课题</h1>
      <div class="actions">
        <el-button type="primary" @click="showCreateDialog" v-if="canCreate">
          <el-icon><Plus /></el-icon>
          创建课题
        </el-button>
      </div>
    </div>

    <div class="filters">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-input
            v-model="searchQuery"
            placeholder="搜索课题..."
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select v-model="visibilityFilter" placeholder="可见性" @change="fetchProjects">
            <el-option label="全部" value="" />
            <el-option label="公开" value="public" />
            <el-option label="私有" value="private" />
          </el-select>
        </el-col>
      </el-row>
    </div>

    <el-row :gutter="20" v-loading="loading">
      <el-col
        :span="8"
        v-for="project in projects"
        :key="project.id"
        class="project-col"
      >
        <el-card class="project-card" @click="viewProject(project.id)">
          <div class="project-header">
            <h3>{{ project.name }}</h3>
            <el-tag :type="project.is_public ? 'success' : 'warning'" size="small">
              {{ project.is_public ? '公开' : '私有' }}
            </el-tag>
          </div>
          <p class="project-description">{{ project.description }}</p>
          <div class="project-meta">
            <span>创建者: {{ project.creator?.name }}</span>
            <span>创建时间: {{ formatDate(project.created_at) }}</span>
            <span>成员: {{ project.members?.length || 0 }}</span>
          </div>
          <div class="project-actions" @click.stop>
            <el-button size="small" @click="viewProject(project.id)">查看</el-button>
            <el-button 
              size="small" 
              type="primary" 
              @click="editProject(project.id)"
              v-if="canEdit(project)"
            >
              编辑
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="pagination" v-if="total > pageSize">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[12, 24, 48]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="fetchProjects"
        @size-change="fetchProjects"
      />
    </div>

    <!-- 创建课题对话框 -->
    <el-dialog v-model="createDialogVisible" title="创建研究课题" width="600px">
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="100px"
      >
        <el-form-item label="课题名称" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入课题名称" />
        </el-form-item>
        <el-form-item label="课题描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入课题描述"
          />
        </el-form-item>
        <el-form-item label="可见性" prop="is_public">
          <el-switch
            v-model="createForm.is_public"
            active-text="公开"
            inactive-text="私有"
          />
        </el-form-item>
        <el-form-item label="标签">
          <el-input
            v-model="tagInput"
            placeholder="输入标签后回车添加"
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
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleCreate" :loading="creating">
            创建
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { researchService } from '@/services/api'
import type { ResearchProject } from '@/types'
import { ElMessage, type FormInstance } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const projects = ref<ResearchProject[]>([])
const loading = ref(false)
const creating = ref(false)
const searchQuery = ref('')
const visibilityFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(12)
const total = ref(0)

const createDialogVisible = ref(false)
const createFormRef = ref<FormInstance>()
const tagInput = ref('')

const createForm = ref({
  name: '',
  description: '',
  is_public: true,
  tags: [] as string[]
})

const createRules = {
  name: [{ required: true, message: '请输入课题名称', trigger: 'blur' }],
  description: [{ required: true, message: '请输入课题描述', trigger: 'blur' }]
}

// 计算属性
const canCreate = computed(() => {
  return userStore.hasRole('teacher') || userStore.hasRole('admin')
})

// 方法
const fetchProjects = async () => {
  loading.value = true
  try {
    const response: any = await researchService.getProjects({
      page: currentPage.value,
      pageSize: pageSize.value,
      search: searchQuery.value || undefined,
      visibility: visibilityFilter.value || undefined
    })
    
    // 处理响应数据结构（axios拦截器已经返回response.data）
    if (response && response.projects) {
      projects.value = response.projects || []
      total.value = response.pagination?.total || response.projects.length || 0
    } else if (Array.isArray(response)) {
      projects.value = response
      total.value = response.length
    } else {
      projects.value = []
      total.value = 0
    }
  } catch (error: any) {
    console.error('获取课题列表失败:', error)
    // 只在非404错误时显示错误提示
    if (error.response?.status !== 404) {
      ElMessage.error('获取课题列表失败')
    }
    projects.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchProjects()
}

const viewProject = (id: string) => {
  router.push(`/scenes/${id}`)
}

const editProject = (id: string) => {
  router.push(`/scenes/${id}/edit`)
}

const canEdit = (project: ResearchProject) => {
  return userStore.hasRole('admin') || project.creator_id === userStore.user?.id
}

const showCreateDialog = () => {
  createDialogVisible.value = true
  resetCreateForm()
}

const resetCreateForm = () => {
  createForm.value = {
    name: '',
    description: '',
    is_public: true,
    tags: []
  }
  tagInput.value = ''
}

const addTag = () => {
  if (tagInput.value.trim() && !createForm.value.tags.includes(tagInput.value.trim())) {
    createForm.value.tags.push(tagInput.value.trim())
    tagInput.value = ''
  }
}

const removeTag = (tag: string) => {
  const index = createForm.value.tags.indexOf(tag)
  if (index > -1) {
    createForm.value.tags.splice(index, 1)
  }
}

const handleCreate = async () => {
  if (!createFormRef.value) return
  
  try {
    await createFormRef.value.validate()
    creating.value = true
    
    await researchService.createProject({
      name: createForm.value.name,
      description: createForm.value.description,
      is_public: createForm.value.is_public,
      tags: createForm.value.tags
    })
    
    ElMessage.success('课题创建成功')
    createDialogVisible.value = false
    fetchProjects()
  } catch (error) {
    console.error('创建课题失败:', error)
    ElMessage.error('创建课题失败')
  } finally {
    creating.value = false
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('zh-CN')
}

// 生命周期
onMounted(() => {
  fetchProjects()
})
</script>

<style scoped>
.scene-list {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filters {
  margin-bottom: 20px;
}

.project-col {
  margin-bottom: 20px;
}

.project-card {
  cursor: pointer;
  transition: all 0.3s;
  height: 100%;
}

.project-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.project-header h3 {
  margin: 0;
  font-size: 16px;
}

.project-description {
  color: #666;
  margin-bottom: 15px;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.project-meta {
  font-size: 12px;
  color: #999;
  display: flex;
  flex-direction: column;
  gap: 5px;
  margin-bottom: 15px;
}

.project-actions {
  display: flex;
  gap: 10px;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 30px;
}

.tags-container {
  margin-top: 10px;
}
</style>
