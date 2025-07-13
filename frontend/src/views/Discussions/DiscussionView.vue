<template>
  <div class="discussion-container">
    <div class="discussion-header">
      <div class="header-left">
        <h1 class="discussion-title">
          <i class="el-icon-chat-dot-round"></i>
          话题讨论
        </h1>
        <p class="discussion-subtitle">基于GitLab Issues的话题讨论功能</p>
      </div>
      <div class="header-right">
        <el-button 
          type="primary" 
          :icon="Plus"
          @click="showCreateDialog = true"
          v-if="canCreateDiscussion"
        >
          创建话题
        </el-button>
      </div>
    </div>

    <!-- 过滤器 -->
    <div class="discussion-filters">
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="项目">
          <el-select v-model="filters.projectId" placeholder="选择项目" @change="handleProjectChange">
            <el-option 
              v-for="project in projects" 
              :key="project.id" 
              :label="project.name" 
              :value="project.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="分类">
          <el-select v-model="filters.category" placeholder="选择分类" @change="loadDiscussions">
            <el-option label="全部" value="" />
            <el-option 
              v-for="category in categories" 
              :key="category" 
              :label="getCategoryLabel(category)" 
              :value="category"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="选择状态" @change="loadDiscussions">
            <el-option label="全部" value="" />
            <el-option label="开放" value="open" />
            <el-option label="关闭" value="closed" />
          </el-select>
        </el-form-item>
        
        <el-form-item>
          <el-button :icon="Refresh" @click="loadDiscussions">刷新</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 话题列表 -->
    <div class="discussion-list">
      <el-card v-if="loading" class="loading-card">
        <div class="loading-content">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>加载中...</span>
        </div>
      </el-card>
      
      <el-empty v-else-if="discussions.length === 0" description="暂无话题" />
      
      <div v-else class="discussion-items">
        <div 
          v-for="discussion in discussions" 
          :key="discussion.id"
          class="discussion-item"
          :class="{ 'is-pinned': discussion.is_pinned }"
          @click="goToDiscussion(discussion.id)"
        >
          <div class="discussion-content">
            <div class="discussion-meta">
              <el-tag 
                v-if="discussion.is_pinned" 
                type="warning" 
                size="small"
                class="pin-tag"
              >
                <el-icon><Top /></el-icon>
                置顶
              </el-tag>
              <el-tag 
                :type="getCategoryType(discussion.category)" 
                size="small"
                class="category-tag"
              >
                {{ getCategoryLabel(discussion.category) }}
              </el-tag>
              <el-tag 
                :type="discussion.status === 'open' ? 'success' : 'info'" 
                size="small"
              >
                {{ discussion.status === 'open' ? '开放' : '关闭' }}
              </el-tag>
            </div>
            
            <h3 class="discussion-title">{{ discussion.title }}</h3>
            <p class="discussion-description">{{ truncate(discussion.content, 200) }}</p>
            
            <div class="discussion-stats">
              <div class="stat-item">
                <el-icon><User /></el-icon>
                <span>{{ discussion.author.name }}</span>
              </div>
              <div class="stat-item">
                <el-icon><View /></el-icon>
                <span>{{ discussion.view_count }} 浏览</span>
              </div>
              <div class="stat-item">
                <el-icon><ChatDotRound /></el-icon>
                <span>{{ discussion.reply_count }} 回复</span>
              </div>
              <div class="stat-item">
                <el-icon><Star /></el-icon>
                <span>{{ discussion.like_count }} 点赞</span>
              </div>
              <div class="stat-item">
                <el-icon><Clock /></el-icon>
                <span>{{ formatDate(discussion.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination-container" v-if="total > 0">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <!-- 创建话题对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建话题"
      width="60%"
      @close="resetCreateForm"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="80px"
      >
        <el-form-item label="项目" prop="project_id">
          <el-select v-model="createForm.project_id" placeholder="选择项目">
            <el-option 
              v-for="project in projects" 
              :key="project.id" 
              :label="project.name" 
              :value="project.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="标题" prop="title">
          <el-input v-model="createForm.title" placeholder="输入话题标题" />
        </el-form-item>
        
        <el-form-item label="分类" prop="category">
          <el-select v-model="createForm.category" placeholder="选择分类">
            <el-option 
              v-for="category in categories" 
              :key="category" 
              :label="getCategoryLabel(category)" 
              :value="category"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="内容" prop="content">
          <el-input
            v-model="createForm.content"
            type="textarea"
            :rows="8"
            placeholder="输入话题内容，支持Markdown格式"
          />
        </el-form-item>
        
        <el-form-item label="标签">
          <el-input v-model="createForm.tags" placeholder="输入标签，用逗号分隔" />
        </el-form-item>
        
        <el-form-item>
          <el-checkbox v-model="createForm.is_public">公开话题</el-checkbox>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="createDiscussion" :loading="creating">
            创建
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, 
  Refresh, 
  Loading, 
  Top, 
  User, 
  View, 
  ChatDotRound, 
  Star, 
  Clock 
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { ApiService } from '@/services/api'
import type { 
  Discussion, 
  Project, 
  DiscussionFilters, 
  CreateDiscussionRequest,
  DiscussionCategory 
} from '@/types/discussion'

const router = useRouter()
const authStore = useAuthStore()

// 响应式数据
const loading = ref(false)
const creating = ref(false)
const showCreateDialog = ref(false)
const discussions = ref<Discussion[]>([])
const projects = ref<Project[]>([])
const categories = ref<string[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

// 过滤器
const filters = reactive<DiscussionFilters>({
  projectId: '',
  category: '',
  status: ''
})

// 创建表单
const createForm = reactive<CreateDiscussionRequest>({
  project_id: 0,
  title: '',
  content: '',
  category: 'general',
  tags: '',
  is_public: true
})

// 创建表单验证规则
const createRules = {
  project_id: [
    { required: true, message: '请选择项目', trigger: 'change' }
  ],
  title: [
    { required: true, message: '请输入标题', trigger: 'blur' },
    { min: 1, max: 200, message: '标题长度在1到200个字符', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入内容', trigger: 'blur' },
    { min: 10, message: '内容至少10个字符', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择分类', trigger: 'change' }
  ]
}

// 计算属性
const canCreateDiscussion = computed(() => {
  return authStore.userRole === 1 || authStore.userRole === 2 || authStore.userRole === 3
})

// 组件挂载时加载数据
onMounted(async () => {
  await loadProjects()
  await loadCategories()
  if (filters.projectId) {
    await loadDiscussions()
  }
})

// 加载项目列表
const loadProjects = async () => {
  try {
    const response = await ApiService.getProjects()
    projects.value = response.data || []
    if (projects.value.length > 0) {
      filters.projectId = projects.value[0].id.toString()
    }
  } catch (error) {
    console.error('加载项目失败:', error)
    ElMessage.error('加载项目失败')
  }
}

// 加载分类列表
const loadCategories = async () => {
  try {
    const response = await ApiService.getDiscussionCategories()
    categories.value = response.data?.categories || []
  } catch (error) {
    console.error('加载分类失败:', error)
    // 使用默认分类
    categories.value = ['general', 'question', 'announcement', 'help', 'feedback']
  }
}

// 加载话题列表
const loadDiscussions = async () => {
  if (!filters.projectId) return
  
  loading.value = true
  try {
    const params = {
      project_id: Number(filters.projectId),
      page: currentPage.value,
      page_size: pageSize.value,
      category: filters.category,
      status: filters.status
    }
    
    const response = await ApiService.getDiscussions(params)
    discussions.value = response.data?.discussions || []
    total.value = response.data?.total || 0
  } catch (error) {
    console.error('加载话题失败:', error)
    ElMessage.error('加载话题失败')
  } finally {
    loading.value = false
  }
}

// 项目变化处理
const handleProjectChange = () => {
  currentPage.value = 1
  loadDiscussions()
}

// 分页处理
const handleCurrentChange = (page: number) => {
  currentPage.value = page
  loadDiscussions()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  currentPage.value = 1
  loadDiscussions()
}

// 跳转到话题详情
const goToDiscussion = (id: number) => {
  router.push(`/discussions/${id}`)
}

// 创建话题
const createDiscussion = async () => {
  // 表单验证
  const formRef = ref()
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch (error) {
    return
  }
  
  creating.value = true
  try {
    const requestData = {
      ...createForm,
      project_id: Number(createForm.project_id)
    }
    await ApiService.createDiscussion(requestData)
    ElMessage.success('话题创建成功')
    showCreateDialog.value = false
    resetCreateForm()
    await loadDiscussions()
  } catch (error) {
    console.error('创建话题失败:', error)
    ElMessage.error('创建话题失败')
  } finally {
    creating.value = false
  }
}

// 重置创建表单
const resetCreateForm = () => {
  Object.assign(createForm, {
    project_id: 0,
    title: '',
    content: '',
    category: 'general',
    tags: '',
    is_public: true
  })
}

// 获取分类标签
const getCategoryLabel = (category: string) => {
  const labels: Record<string, string> = {
    general: '通用',
    question: '问题',
    announcement: '公告',
    help: '求助',
    feedback: '反馈'
  }
  return labels[category] || category
}

// 获取分类类型
const getCategoryType = (category: string) => {
  const types: Record<string, string> = {
    general: '',
    question: 'warning',
    announcement: 'danger',
    help: 'info',
    feedback: 'success'
  }
  return types[category] || ''
}

// 格式化日期
const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN') + ' ' + date.toLocaleTimeString('zh-CN')
}

// 过滤器：截断文本
const truncate = (text: string, length: number) => {
  if (text.length <= length) return text
  return text.substring(0, length) + '...'
}
</script>

<style scoped>
.discussion-container {
  padding: 20px;
  background-color: #f8f9fa;
  min-height: 100vh;
}

.discussion-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.header-left {
  flex: 1;
}

.discussion-title {
  margin: 0;
  color: #303133;
  font-size: 24px;
  font-weight: 600;
}

.discussion-title i {
  margin-right: 8px;
  color: #409eff;
}

.discussion-subtitle {
  margin: 8px 0 0 0;
  color: #909399;
  font-size: 14px;
}

.discussion-filters {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.filter-form {
  margin: 0;
}

.discussion-list {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  overflow: hidden;
}

.loading-card {
  border: none;
  box-shadow: none;
}

.loading-content {
  text-align: center;
  padding: 40px;
  color: #909399;
}

.loading-content .el-icon {
  font-size: 24px;
  margin-right: 8px;
}

.discussion-items {
  padding: 0;
}

.discussion-item {
  padding: 20px;
  border-bottom: 1px solid #f0f2f5;
  cursor: pointer;
  transition: background-color 0.2s;
}

.discussion-item:hover {
  background-color: #f8f9fa;
}

.discussion-item:last-child {
  border-bottom: none;
}

.discussion-item.is-pinned {
  background-color: #fff9e6;
  border-left: 4px solid #e6a23c;
}

.discussion-content {
  width: 100%;
}

.discussion-meta {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.pin-tag {
  margin-right: 8px;
}

.category-tag {
  margin-right: 8px;
}

.discussion-item .discussion-title {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  line-height: 1.4;
}

.discussion-description {
  margin: 0 0 12px 0;
  color: #606266;
  line-height: 1.5;
  font-size: 14px;
}

.discussion-stats {
  display: flex;
  align-items: center;
  color: #909399;
  font-size: 12px;
}

.stat-item {
  display: flex;
  align-items: center;
  margin-right: 16px;
}

.stat-item .el-icon {
  margin-right: 4px;
}

.pagination-container {
  padding: 20px;
  text-align: center;
  border-top: 1px solid #f0f2f5;
}

.dialog-footer {
  text-align: right;
}

.el-select {
  width: 100%;
}
</style> 