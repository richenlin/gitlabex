<template>
  <div class="topic-list">
    <div class="header">
      <h1>话题讨论</h1>
      <el-button v-if="userStore.isLoggedIn" type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        发布话题
      </el-button>
    </div>

    <div class="filters">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-input
            v-model="searchQuery"
            placeholder="搜索话题..."
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="6">
          <el-select v-model="projectFilter" placeholder="选择课题" clearable @change="fetchTopics">
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="statusFilter" placeholder="状态" @change="fetchTopics">
            <el-option label="全部" value="" />
            <el-option label="开放" value="open" />
            <el-option label="关闭" value="closed" />
          </el-select>
        </el-col>
      </el-row>
    </div>

    <div class="topic-content" v-loading="loading">
      <div class="topic-stats">
        <el-statistic title="总话题数" :value="total" />
        <el-statistic title="活跃话题" :value="activeTopics" />
      </div>

      <div class="topic-items">
        <div
          v-for="topic in topics"
          :key="topic.id"
          class="topic-item"
          @click="viewTopic(topic.id)"
        >
          <div class="topic-main">
            <div class="topic-header">
              <h3 class="topic-title">{{ topic.title }}</h3>
              <div class="topic-badges">
                <el-tag v-if="topic.is_pinned" type="warning" size="small">置顶</el-tag>
                <el-tag :type="topic.status === 'open' ? 'success' : 'info'" size="small">
                  {{ topic.status === 'open' ? '开放' : '关闭' }}
                </el-tag>
                <el-tag v-if="topic.priority !== 'medium'" :type="getPriorityType(topic.priority)" size="small">
                  {{ getPriorityText(topic.priority) }}
                </el-tag>
              </div>
            </div>
            
            <div class="topic-content-preview">
              {{ topic.content ? topic.content.substring(0, 150) + '...' : '暂无内容' }}
            </div>
            
            <div class="topic-labels" v-if="topic.tags && topic.tags.length">
              <el-tag
                v-for="tag in topic.tags.slice(0, 3)"
                :key="tag"
                size="small"
                class="topic-label"
              >
                {{ tag }}
              </el-tag>
            </div>
          </div>

          <div class="topic-sidebar">
            <div class="topic-stats-item">
              <el-icon><Star /></el-icon>
              <span>{{ topic.like_count || 0 }}</span>
            </div>
            <div class="topic-stats-item">
              <el-icon><ChatDotRound /></el-icon>
              <span>{{ topic.comments_count || 0 }}</span>
            </div>
          </div>

          <div class="topic-meta">
            <div class="author-info">
              <el-avatar :size="32" :src="topic.author?.avatar_url">
                {{ topic.author?.name?.charAt(0) || 'U' }}
              </el-avatar>
              <div class="author-details">
                <span class="author-name">{{ topic.author?.name || '未知用户' }}</span>
                <span class="topic-time">{{ formatDate(topic.created_at) }}</span>
              </div>
            </div>
            <div class="project-info" v-if="topic.project">
              <el-link :href="`/scenes/${topic.project.id}`" :underline="false">
                {{ topic.project.name }}
              </el-link>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="pagination" v-if="total > pageSize">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="fetchTopics"
        @size-change="fetchTopics"
      />
    </div>

    <!-- 创建话题对话框 -->
    <el-dialog v-model="createDialogVisible" title="发布话题" width="700px">
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="100px"
      >
        <el-form-item label="话题标题" prop="title">
          <el-input v-model="createForm.title" placeholder="请输入话题标题" />
        </el-form-item>
        <el-form-item label="话题内容" prop="content">
          <el-input
            v-model="createForm.content"
            type="textarea"
            :rows="6"
            placeholder="请输入话题内容"
          />
        </el-form-item>
        <el-form-item label="关联课题">
          <el-select v-model="createForm.project_id" placeholder="选择课题" clearable>
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="优先级">
          <el-select v-model="createForm.priority">
            <el-option label="低" value="low" />
            <el-option label="中" value="medium" />
            <el-option label="高" value="high" />
            <el-option label="紧急" value="urgent" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-input
            v-model="labelInput"
            placeholder="输入标签后回车添加"
            @keyup.enter="addLabel"
          />
          <div class="labels-container" v-if="createForm.labels.length">
            <el-tag
              v-for="label in createForm.labels"
              :key="label"
              closable
              @close="removeLabel(label)"
              style="margin: 5px 5px 0 0"
            >
              {{ label }}
            </el-tag>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleCreate" :loading="creating">
            发布
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
import { topicService, researchService } from '@/services/api'
import type { Topic, ResearchProject } from '@/types'
import { ElMessage, type FormInstance } from 'element-plus'
import { Plus, Search, Star, ChatDotRound } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const topics = ref<Topic[]>([])
const projects = ref<ResearchProject[]>([])
const loading = ref(false)
const creating = ref(false)
const searchQuery = ref('')
const projectFilter = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const createDialogVisible = ref(false)
const createFormRef = ref<FormInstance>()
const labelInput = ref('')

const createForm = ref({
  title: '',
  content: '',
  project_id: '',
  priority: 'medium',
  labels: [] as string[]
})

const createRules = {
  title: [{ required: true, message: '请输入话题标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入话题内容', trigger: 'blur' }]
}

// 计算属性
const activeTopics = computed(() => {
  return topics.value.filter(topic => topic.status === 'open').length
})

// 方法
const fetchTopics = async () => {
  loading.value = true
  try {
    const response: any = await topicService.getTopics({
      page: currentPage.value,
      pageSize: pageSize.value,
      search: searchQuery.value || undefined,
      projectId: projectFilter.value || undefined
    })
    
    // 处理响应数据结构（axios拦截器已经返回response.data）
    if (response && response.topics) {
      topics.value = response.topics || []
      total.value = response.pagination?.total || response.topics.length || 0
    } else if (Array.isArray(response)) {
      topics.value = response
      total.value = response.length
    } else {
      topics.value = []
      total.value = 0
    }
  } catch (error: any) {
    console.error('获取话题列表失败:', error)
    // 只在非404错误时显示错误提示
    if (error.response?.status !== 404) {
      ElMessage.error('获取话题列表失败')
    }
    topics.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const fetchProjects = async () => {
  try {
    const response: any = await researchService.getProjects()
    projects.value = response.projects || []
  } catch (error) {
    console.error('获取课题列表失败:', error)
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchTopics()
}

const viewTopic = (id: string) => {
  router.push(`/topics/${id}`)
}

const showCreateDialog = () => {
  createDialogVisible.value = true
  resetCreateForm()
}

const resetCreateForm = () => {
  createForm.value = {
    title: '',
    content: '',
    project_id: '',
    priority: 'medium',
    labels: []
  }
  labelInput.value = ''
}

const addLabel = () => {
  if (labelInput.value.trim() && !createForm.value.labels.includes(labelInput.value.trim())) {
    createForm.value.labels.push(labelInput.value.trim())
    labelInput.value = ''
  }
}

const removeLabel = (label: string) => {
  const index = createForm.value.labels.indexOf(label)
  if (index > -1) {
    createForm.value.labels.splice(index, 1)
  }
}

const handleCreate = async () => {
  if (!createFormRef.value) return
  
  try {
    await createFormRef.value.validate()
    creating.value = true
    
    await topicService.createTopic({
      title: createForm.value.title,
      content: createForm.value.content,
      project_id: createForm.value.project_id || undefined,
      priority: createForm.value.priority as 'low' | 'medium' | 'high' | 'urgent',
      labels: createForm.value.labels
    })
    
    ElMessage.success('话题发布成功')
    createDialogVisible.value = false
    fetchTopics()
  } catch (error) {
    console.error('发布话题失败:', error)
    ElMessage.error('发布话题失败')
  } finally {
    creating.value = false
  }
}

const getPriorityType = (priority: string) => {
  const types: Record<string, string> = {
    low: 'info',
    medium: '',
    high: 'warning',
    urgent: 'danger'
  }
  return types[priority] || ''
}

const getPriorityText = (priority: string) => {
  const texts: Record<string, string> = {
    low: '低优先级',
    medium: '中优先级',
    high: '高优先级',
    urgent: '紧急'
  }
  return texts[priority] || priority
}

const formatDate = (date: string) => {
  const now = new Date()
  const time = new Date(date)
  const diff = now.getTime() - time.getTime()
  
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return time.toLocaleDateString('zh-CN')
}

// 生命周期
onMounted(() => {
  fetchTopics()
  fetchProjects()
})
</script>

<style scoped>
.topic-list {
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

.topic-content {
  background: white;
  border-radius: 8px;
  padding: 20px;
}

.topic-stats {
  display: flex;
  gap: 30px;
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f0f0f0;
}

.topic-items {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.topic-item {
  display: flex;
  align-items: flex-start;
  gap: 15px;
  padding: 15px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.topic-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.topic-main {
  flex: 1;
}

.topic-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.topic-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.topic-badges {
  display: flex;
  gap: 5px;
}

.topic-content-preview {
  color: #606266;
  margin-bottom: 10px;
  line-height: 1.5;
}

.topic-labels {
  display: flex;
  gap: 5px;
  flex-wrap: wrap;
}

.topic-label {
  margin: 0;
}

.topic-sidebar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: center;
  min-width: 60px;
}

.topic-stats-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  color: #909399;
}

.topic-meta {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 150px;
}

.author-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.author-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.author-name {
  font-size: 14px;
  font-weight: 500;
}

.topic-time {
  font-size: 12px;
  color: #909399;
}

.project-info {
  font-size: 12px;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 30px;
}

.labels-container {
  margin-top: 10px;
}
</style>
