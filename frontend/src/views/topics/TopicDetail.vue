<template>
  <div class="topic-detail">
    <el-card v-if="topic" class="topic-card">
      <!-- 话题头部 -->
      <div class="topic-header">
        <div class="topic-title">
          <h1>{{ topic.title }}</h1>
          <el-tag :type="getPriorityType(topic.priority)" class="priority-tag">
            {{ getPriorityText(topic.priority) }}
          </el-tag>
        </div>
        <div class="topic-meta">
          <span class="author">作者：{{ topic.author?.name || `用户${topic.author_id}` }}</span>
          <span class="time">{{ formatTime(topic.created_at) }}</span>
          <span class="views">浏览：{{ topic.view_count || 0 }}</span>
        </div>
      </div>

      <!-- 标签 -->
      <div class="topic-tags" v-if="topic.tags && topic.tags.length > 0">
        <el-tag
          v-for="tag in topic.tags"
          :key="tag"
          size="small"
          class="tag"
        >
          {{ tag }}
        </el-tag>
      </div>

      <!-- 话题内容 -->
      <div class="topic-content">
        <div class="content-text" v-html="formatContent(topic.content)"></div>
      </div>

      <!-- 操作按钮 -->
      <div class="topic-actions">
        <el-button-group>
          <el-button
            :type="userLiked ? 'primary' : 'default'"
            @click="handleLike"
            :loading="likeLoading"
          >
            <el-icon><Star /></el-icon>
            点赞 {{ topic.like_count || 0 }}
          </el-button>
          <el-button
            :type="userDisliked ? 'danger' : 'default'"
            @click="handleDislike"
            :loading="dislikeLoading"
          >
            <el-icon><Star /></el-icon>
            反对 {{ topic.dislike_count || 0 }}
          </el-button>
        </el-button-group>

        <div class="action-buttons">
          <el-button
            v-if="canEdit"
            @click="handleEdit"
          >
            编辑
          </el-button>
          <el-button
            v-if="canEdit"
            type="danger"
            @click="handleDelete"
          >
            删除
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 评论区域 -->
    <el-card class="comments-card" v-if="topic">
      <template #header>
        <div class="card-header">
          <h3>评论 ({{ comments.length }})</h3>
        </div>
      </template>

      <!-- 发表评论 -->
      <div class="comment-form">
        <el-input
          v-model="newComment"
          type="textarea"
          :rows="3"
          placeholder="发表你的看法..."
          maxlength="1000"
          show-word-limit
        />
        <div class="comment-actions">
          <el-button
            type="primary"
            @click="submitComment"
            :loading="commentLoading"
            :disabled="!newComment.trim()"
          >
            发表评论
          </el-button>
        </div>
      </div>

      <!-- 评论列表 -->
      <div class="comments-list">
        <div
          v-for="comment in comments"
          :key="comment.id"
          class="comment-item"
        >
          <div class="comment-header">
            <span class="comment-author">{{ comment.author?.name || `用户${comment.author_id}` }}</span>
            <span class="comment-time">{{ formatTime(comment.created_at) }}</span>
          </div>
          <div class="comment-content">{{ comment.content }}</div>
        </div>
      </div>
    </el-card>

    <!-- 加载状态 -->
    <el-skeleton v-if="loading" :rows="8" animated />

    <!-- 错误状态 -->
    <el-empty v-if="error" :description="error" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Star } from '@element-plus/icons-vue'
import { topicService } from '@/services/api'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 数据
const topic = ref<any>(null)
const comments = ref<any[]>([])
const newComment = ref('')
const loading = ref(false)
const error = ref('')
const likeLoading = ref(false)
const dislikeLoading = ref(false)
const commentLoading = ref(false)
const userLiked = ref(false)
const userDisliked = ref(false)

// 计算属性
const canEdit = computed(() => {
  if (!topic.value || !userStore.user) return false
  return topic.value.author_id === userStore.user.id || userStore.isAdmin
})

// 方法
const loadTopic = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const topicId = route.params.id as string
    const projectId = route.query.project_id as string || ''
    const response = await topicService.getTopic(topicId, projectId)
    topic.value = response.data
    
    // 加载评论
    await loadComments()
  } catch (err) {
    error.value = '加载话题失败'
    console.error('加载话题失败:', err)
  } finally {
    loading.value = false
  }
}

const loadComments = async () => {
  try {
    const topicId = route.params.id as string
    const projectId = route.query.project_id as string || ''
    const response = await topicService.getComments(topicId, projectId)
    comments.value = response.data || []
  } catch (err) {
    console.error('加载评论失败:', err)
  }
}

const handleLike = async () => {
  if (!userStore.user) {
    ElMessage.warning('请先登录')
    return
  }
  
  likeLoading.value = true
  try {
    const topicId = route.params.id as string
    const projectId = route.query.project_id as string || ''
    
    if (userLiked.value) {
      await topicService.unlikeTopic(topicId, projectId)
      if (topic.value) {
        topic.value.like_count = Math.max(0, topic.value.like_count - 1)
      }
      userLiked.value = false
    } else {
      await topicService.likeTopic(topicId, projectId)
      if (topic.value) {
        topic.value.like_count = (topic.value.like_count || 0) + 1
      }
      userLiked.value = true
      
      // 如果之前反对过，取消反对
      if (userDisliked.value) {
        if (topic.value) {
          topic.value.dislike_count = Math.max(0, topic.value.dislike_count - 1)
        }
        userDisliked.value = false
      }
    }
  } catch (err) {
    ElMessage.error('操作失败')
    console.error('点赞失败:', err)
  } finally {
    likeLoading.value = false
  }
}

const handleDislike = async () => {
  if (!userStore.user) {
    ElMessage.warning('请先登录')
    return
  }
  
  dislikeLoading.value = true
  try {
    const topicId = route.params.id as string
    const projectId = route.query.project_id as string || ''
    
    if (userDisliked.value) {
      await topicService.undislikeTopic(topicId, projectId)
      if (topic.value) {
        topic.value.dislike_count = Math.max(0, topic.value.dislike_count - 1)
      }
      userDisliked.value = false
    } else {
      await topicService.dislikeTopic(topicId, projectId)
      if (topic.value) {
        topic.value.dislike_count = (topic.value.dislike_count || 0) + 1
      }
      userDisliked.value = true
      
      // 如果之前点赞过，取消点赞
      if (userLiked.value) {
        if (topic.value) {
          topic.value.like_count = Math.max(0, topic.value.like_count - 1)
        }
        userLiked.value = false
      }
    }
  } catch (err) {
    ElMessage.error('操作失败')
    console.error('反对失败:', err)
  } finally {
    dislikeLoading.value = false
  }
}

const handleEdit = () => {
  router.push(`/topics/${route.params.id}/edit`)
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm('确定要删除这个话题吗？', '确认删除', {
      type: 'warning'
    })
    
    const topicId = route.params.id as string
    await topicService.deleteTopic(topicId)
    ElMessage.success('话题删除成功')
    router.push('/topics')
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除话题失败:', err)
    }
  }
}

const submitComment = async () => {
  if (!userStore.user) {
    ElMessage.warning('请先登录')
    return
  }
  
  if (!newComment.value.trim()) {
    ElMessage.warning('请输入评论内容')
    return
  }
  
  commentLoading.value = true
  try {
    const topicId = route.params.id as string
    const projectId = route.query.project_id as string || ''
    
    await topicService.createComment(
      topicId,
      newComment.value.trim(),
      projectId
    )
    
    ElMessage.success('评论发表成功')
    newComment.value = ''
    await loadComments()
  } catch (err) {
    ElMessage.error('发表评论失败')
    console.error('发表评论失败:', err)
  } finally {
    commentLoading.value = false
  }
}

const getPriorityType = (priority: string) => {
  const types: Record<string, string> = {
    low: '',
    normal: 'info',
    high: 'warning',
    urgent: 'danger'
  }
  return types[priority] || 'info'
}

const getPriorityText = (priority: string) => {
  const texts: Record<string, string> = {
    low: '低',
    medium: '普通',
    high: '高',
    urgent: '紧急'
  }
  return texts[priority] || '普通'
}

const formatTime = (time: string) => {
  const date = new Date(time)
  return date.toLocaleString('zh-CN')
}

const formatContent = (content: string) => {
  // 简单的换行处理
  return content.replace(/\n/g, '<br>')
}

// 生命周期
onMounted(() => {
  loadTopic()
})
</script>

<style scoped>
.topic-detail {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

.topic-card {
  margin-bottom: 20px;
}

.topic-header {
  margin-bottom: 16px;
}

.topic-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.topic-title h1 {
  margin: 0;
  color: #303133;
  font-size: 24px;
}

.priority-tag {
  flex-shrink: 0;
}

.topic-meta {
  display: flex;
  gap: 16px;
  color: #909399;
  font-size: 14px;
}

.topic-tags {
  margin-bottom: 16px;
}

.tag {
  margin-right: 8px;
  margin-bottom: 8px;
}

.topic-content {
  margin-bottom: 20px;
  padding: 16px 0;
  border-top: 1px solid #ebeef5;
  border-bottom: 1px solid #ebeef5;
}

.content-text {
  line-height: 1.6;
  color: #303133;
}

.topic-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.comments-card {
  margin-bottom: 20px;
}

.comment-form {
  margin-bottom: 20px;
}

.comment-actions {
  margin-top: 12px;
  text-align: right;
}

.comment-item {
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.comment-item:last-child {
  border-bottom: none;
}

.comment-header {
  display: flex;
  gap: 12px;
  margin-bottom: 8px;
  font-size: 14px;
}

.comment-author {
  color: #303133;
  font-weight: 500;
}

.comment-time {
  color: #909399;
}

.comment-content {
  color: #606266;
  line-height: 1.5;
}
</style>
