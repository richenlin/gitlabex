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
          @click="viewTopic(topic)"
        >
          <div class="topic-main">
            <div class="topic-header">
              <h3 class="topic-title">{{ topic.title }}</h3>
              <div class="topic-badges">
                <el-tag :type="topic.status === 'opened' ? 'success' : 'info'" size="small">
                  {{ topic.status === 'opened' ? '开放' : '关闭' }}
                </el-tag>
                <el-tag v-if="topic.project" type="primary" size="small">
                  {{ topic.project.name }}
                </el-tag>
              </div>
            </div>
            
            <div class="topic-content-preview">
              {{ topic.content ? topic.content.substring(0, 150) + '...' : '暂无内容' }}
            </div>
            
            <div class="topic-labels" v-if="topic.labels && topic.labels.length">
              <el-tag
                v-for="tag in topic.labels.slice(0, 3)"
                :key="tag"
                size="small"
                class="topic-label"
              >
                {{ tag }}
              </el-tag>
            </div>
            
            <!-- 话题操作区域 -->
            <div class="topic-actions" @click.stop>
              <el-button 
                size="small" 
                :type="topic.user_liked ? 'primary' : 'default'"
                :icon="topic.user_liked ? 'StarFilled' : 'Star'"
                @click="toggleLike(topic)"
                :loading="topic.liking"
              >
                点赞 {{ topic.like_count || 0 }}
              </el-button>
              
              <el-button 
                size="small" 
                :type="topic.user_disliked ? 'danger' : 'default'"
                icon="CircleClose"
                @click="toggleDislike(topic)"
                :loading="topic.disliking"
              >
                反对 {{ topic.dislike_count || 0 }}
              </el-button>
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

    <!-- 话题详情模态框 -->
    <el-dialog 
      v-model="topicDetailDialogVisible" 
      :title="currentTopicDetail?.title || '话题详情'" 
      width="80%" 
      max-width="800px"
      top="5vh"
    >
      <div v-loading="topicDetailLoading">
        <!-- 话题内容 -->
        <div v-if="currentTopicDetail" class="topic-detail">
          <div class="topic-header">
            <div class="topic-meta">
              <el-avatar 
                :src="currentTopicDetail.author?.avatar_url" 
                :size="32"
                class="author-avatar"
              />
              <div class="author-info">
                <span class="author-name">{{ currentTopicDetail.author?.name }}</span>
                <span class="publish-time">{{ formatDate(currentTopicDetail.created_at) }}</span>
              </div>
            </div>
            <div class="topic-labels" v-if="currentTopicDetail.labels?.length">
              <el-tag 
                v-for="label in currentTopicDetail.labels" 
                :key="label" 
                size="small"
                class="topic-label"
              >
                {{ label }}
              </el-tag>
            </div>
          </div>
          
          <div class="topic-content">
            <p>{{ currentTopicDetail.content }}</p>
          </div>
          
          <div class="topic-stats">
            <span class="stat-item">
              <el-icon><ChatDotRound /></el-icon>
              {{ currentTopicDetail.comments_count || 0 }} 回复
            </span>
            <span class="stat-item">
              <el-icon><Star /></el-icon>
              {{ currentTopicDetail.like_count || 0 }} 点赞
            </span>
            <span class="stat-item">
              <el-icon><CircleClose /></el-icon>
              {{ currentTopicDetail.dislike_count || 0 }} 反对
            </span>
          </div>
        </div>
        
        <!-- 回复输入框 -->
        <div class="reply-section" v-if="userStore.isLoggedIn">
          <h4 class="reply-title">添加回复</h4>
          <el-input 
            v-model="replyForm.content" 
            type="textarea" 
            :rows="3"
            placeholder="请输入回复内容..."
            maxlength="1000"
            show-word-limit
            class="reply-input"
          />
          <div class="reply-actions">
            <el-button @click="replyForm.content = ''">清空</el-button>
            <el-button type="primary" @click="submitReply" :loading="submittingReply">
              {{ submittingReply ? '提交中...' : '提交回复' }}
            </el-button>
          </div>
        </div>
        
        <!-- 回复列表 -->
        <div class="comments-section" v-if="topicComments.length > 0 || commentsTotal > 0">
          <h4 class="comments-title">回复 ({{ commentsTotal || topicComments.length }})</h4>
          <div class="comments-list" v-loading="commentsLoading">
            <div 
              v-for="comment in topicComments" 
              :key="comment.id" 
              class="comment-item"
            >
              <el-avatar 
                :src="comment.author?.avatar_url" 
                :size="28"
                class="comment-avatar"
              />
              <div class="comment-content">
                <div class="comment-header">
                  <span class="comment-author">{{ comment.author?.name }}</span>
                  <span class="comment-time">{{ formatDate(comment.created_at) }}</span>
                </div>
                <div class="comment-body">
                  {{ comment.content }}
                </div>
              </div>
            </div>
          </div>
          
          <!-- 回复分页 -->
          <div class="comments-pagination" v-if="commentsTotal > commentsPageSize">
            <el-pagination
              v-model:current-page="commentsCurrentPage"
              :total="commentsTotal"
              :page-size="commentsPageSize"
              layout="prev, pager, next"
              @current-change="handleCommentsPageChange"
            />
          </div>
        </div>
        
        <div v-else class="login-prompt">
          <p>请先登录后再回复话题</p>
        </div>
      </div>
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
import { Plus, Search, Star, ChatDotRound, CircleClose } from '@element-plus/icons-vue'

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

// 话题详情模态框相关状态
const topicDetailDialogVisible = ref(false)
const topicDetailLoading = ref(false)
const currentTopicDetail = ref<any>(null)
const topicComments = ref<any[]>([])
const replyingTopic = ref<Topic | null>(null)
const submittingReply = ref(false)
const replyForm = ref({
  content: ''
})

// 回复分页相关
const commentsCurrentPage = ref(1)
const commentsPageSize = ref(10)
const commentsTotal = ref(0)
const commentsLoading = ref(false)

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
  return topics.value.filter(topic => topic.status === 'opened').length
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

const viewTopic = async (topic: Topic) => {
  // 从话题列表获取项目ID
  if (!topic.project_id) {
    ElMessage.error('无法获取话题所属项目信息')
    return
  }
  
  topicDetailLoading.value = true
  topicDetailDialogVisible.value = true
  
  try {
    const response: any = await topicService.getTopic(topic.id, topic.project_id)
    currentTopicDetail.value = response.topic
    topicComments.value = response.comments || []
    replyingTopic.value = currentTopicDetail.value
  } catch (error) {
    console.error('获取话题详情失败:', error)
    ElMessage.error('获取话题详情失败')
    topicDetailDialogVisible.value = false
  } finally {
    topicDetailLoading.value = false
  }
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

const submitReply = async () => {
  if (!replyForm.value.content.trim()) {
    ElMessage.warning('请输入回复内容')
    return
  }
  
  if (!replyingTopic.value || !replyingTopic.value.project_id) {
    ElMessage.error('无法获取话题信息')
    return
  }
  
  submittingReply.value = true
  try {
    const response: any = await topicService.createComment(
      replyingTopic.value.id, 
      replyForm.value.content, 
      replyingTopic.value.project_id
    )
    ElMessage.success('回复成功')
    replyForm.value.content = ''
    
    // 更新评论列表 - 刷新评论并可能跳转到最后一页
    if (response.comment) {
      if (currentTopicDetail.value) {
        currentTopicDetail.value.comments_count = (currentTopicDetail.value.comments_count || 0) + 1
      }
      
      // 计算新评论可能所在的页面
      const newTotal = currentTopicDetail.value?.comments_count || 1
      const lastPage = Math.ceil(newTotal / commentsPageSize.value)
      
      // 如果新评论在最后一页，跳转到最后一页
      if (lastPage > commentsCurrentPage.value) {
        commentsCurrentPage.value = lastPage
      }
      
      await fetchComments()
    }
    
    // 更新话题列表中的回复数量
    const topicIndex = topics.value.findIndex(t => t.id === replyingTopic.value?.id)
    if (topicIndex >= 0) {
      topics.value[topicIndex].comments_count = (topics.value[topicIndex].comments_count || 0) + 1
    }
  } catch (error) {
    console.error('提交回复失败:', error)
    ElMessage.error('提交回复失败')
  } finally {
    submittingReply.value = false
  }
}

const toggleLike = async (topic: Topic) => {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    return
  }
  
  if (!topic.project_id) {
    ElMessage.error('无法获取话题所属项目信息')
    return
  }
  
  topic.liking = true
  try {
    if (topic.user_liked) {
      await topicService.unlikeTopic(topic.id, topic.project_id)
      topic.user_liked = false
      topic.like_count = Math.max(0, (topic.like_count || 0) - 1)
      ElMessage.success('取消点赞成功')
    } else {
      await topicService.likeTopic(topic.id, topic.project_id)
      topic.user_liked = true
      topic.like_count = (topic.like_count || 0) + 1
      // 如果之前反对了，取消反对
      if (topic.user_disliked) {
        await topicService.undislikeTopic(topic.id, topic.project_id)
        topic.user_disliked = false
        topic.dislike_count = Math.max(0, (topic.dislike_count || 0) - 1)
      }
      ElMessage.success('点赞成功')
    }
  } catch (error) {
    console.error('点赞操作失败:', error)
    ElMessage.error('点赞操作失败')
  } finally {
    topic.liking = false
  }
}

const toggleDislike = async (topic: Topic) => {
  if (!userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    return
  }
  
  if (!topic.project_id) {
    ElMessage.error('无法获取话题所属项目信息')
    return
  }
  
  topic.disliking = true
  try {
    if (topic.user_disliked) {
      await topicService.undislikeTopic(topic.id, topic.project_id)
      topic.user_disliked = false
      topic.dislike_count = Math.max(0, (topic.dislike_count || 0) - 1)
      ElMessage.success('取消反对成功')
    } else {
      await topicService.dislikeTopic(topic.id, topic.project_id)
      topic.user_disliked = true
      topic.dislike_count = (topic.dislike_count || 0) + 1
      // 如果之前点赞了，取消点赞
      if (topic.user_liked) {
        await topicService.unlikeTopic(topic.id, topic.project_id)
        topic.user_liked = false
        topic.like_count = Math.max(0, (topic.like_count || 0) - 1)
      }
      ElMessage.success('反对成功')
    }
  } catch (error) {
    console.error('反对操作失败:', error)
    ElMessage.error('反对操作失败')
  } finally {
    topic.disliking = false
  }
}

const fetchComments = async () => {
  if (!currentTopicDetail.value || !currentTopicDetail.value.project_id) {
    return
  }
  
  commentsLoading.value = true
  try {
    const response: any = await topicService.getTopic(
      currentTopicDetail.value.id, 
      currentTopicDetail.value.project_id
    )
    const allComments = response.comments || []
    commentsTotal.value = allComments.length
    
    // 前端分页逻辑
    const startIndex = (commentsCurrentPage.value - 1) * commentsPageSize.value
    const endIndex = startIndex + commentsPageSize.value
    topicComments.value = allComments.slice(startIndex, endIndex)
  } catch (error) {
    console.error('获取评论失败:', error)
  } finally {
    commentsLoading.value = false
  }
}

const handleCommentsPageChange = (page: number) => {
  commentsCurrentPage.value = page
  fetchComments()
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
  padding: 0; /* 由外层容器控制 padding */
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header h1 {
  font-size: 24px;
  color: var(--text-color);
  margin: 0;
}

.filters {
  margin-bottom: 24px;
  background: #fff;
  padding: 20px;
  border-radius: var(--border-radius-md);
  box-shadow: var(--shadow-sm);
}

/* 统计卡片优化 */
.topic-content {
  background: transparent; /* 移除背景，让列表项独立成为卡片 */
  padding: 0;
}

.topic-stats {
  display: flex;
  gap: 40px;
  margin-bottom: 24px;
  padding: 20px;
  background: #fff;
  border-radius: var(--border-radius-md);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color-light);
}

.topic-items {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 话题列表项 - 卡片化 */
.topic-item {
  display: flex;
  align-items: flex-start;
  gap: 20px;
  padding: 24px;
  background-color: #fff;
  border: 1px solid var(--border-color-light);
  border-radius: var(--border-radius-md);
  cursor: pointer;
  transition: all 0.2s ease-in-out;
  position: relative;
}

.topic-item:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
  border-color: var(--primary-color);
}

.topic-main {
  flex: 1;
  min-width: 0; /* 修复 flex 子项溢出 */
}

.topic-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.topic-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-color);
  line-height: 1.4;
}

.topic-title:hover {
  color: var(--primary-color);
}

.topic-badges {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.topic-content-preview {
  color: var(--text-color-regular);
  margin-bottom: 16px;
  line-height: 1.6;
  font-size: 14px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.topic-labels {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.topic-sidebar {
  display: flex;
  flex-direction: column;
  gap: 16px;
  align-items: center;
  min-width: 60px;
  padding-left: 20px;
  border-left: 1px solid var(--border-color-light);
}

.topic-stats-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  color: var(--text-color-secondary);
  font-size: 12px;
}

.topic-stats-item .el-icon {
  font-size: 18px;
  margin-bottom: 2px;
}

.topic-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 4px;
  padding-top: 16px;
  border-top: 1px solid var(--border-color-light);
}

.author-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.author-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.author-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
}

.topic-time {
  font-size: 12px;
  color: var(--text-color-secondary);
}

.project-info {
  font-size: 13px;
}

/* 话题操作区域 */
.topic-actions {
  display: flex;
  gap: 12px;
  opacity: 0.8;
  transition: opacity 0.2s;
}

.topic-item:hover .topic-actions {
  opacity: 1;
}

.topic-actions .el-button {
  padding: 8px 16px;
  height: 32px;
}

/* 详情页样式优化 */
.topic-detail {
  margin-bottom: 30px;
}

.topic-detail .topic-header {
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--border-color);
}

.topic-content {
  padding: 24px;
  background: #f9fafc; /* 极浅的灰色背景 */
  border-radius: 8px;
  border: 1px solid var(--border-color-light);
  margin-bottom: 24px;
}

.topic-content p {
  margin: 0;
  color: var(--text-color);
  line-height: 1.8;
  font-size: 16px;
}

/* 评论区美化 */
.comments-title {
  font-size: 18px;
  margin-bottom: 20px;
  padding-left: 12px;
  border-left: 4px solid var(--primary-color);
}

.comment-item {
  display: flex;
  gap: 16px;
  padding: 20px;
  background: #fff;
  border-bottom: 1px solid var(--border-color-light);
  transition: background-color 0.2s;
}

.comment-item:hover {
  background-color: #f9fafc;
}

.comment-item:last-child {
  border-bottom: none;
}

.comment-header {
  margin-bottom: 8px;
}

.comment-author {
  font-weight: 600;
  margin-right: 12px;
}

.comment-body {
  color: var(--text-color-regular);
  line-height: 1.6;
}

/* 分页 */
.pagination {
  margin-top: 40px;
  display: flex;
  justify-content: center;
}
</style>
