<template>
  <div class="discussion-detail-container">
    <div class="discussion-header">
      <el-button 
        :icon="ArrowLeft" 
        @click="goBack"
        class="back-button"
      >
        返回列表
      </el-button>
      
      <div class="header-actions" v-if="discussion">
        <el-button 
          v-if="canEdit" 
          :icon="Edit" 
          @click="showEditDialog = true"
        >
          编辑
        </el-button>
        <el-button 
          v-if="canDelete" 
          :icon="Delete" 
          type="danger" 
          @click="deleteDiscussion"
        >
          删除
        </el-button>
        <el-button 
          v-if="canPin" 
          :icon="Top" 
          @click="pinDiscussion"
        >
          {{ discussion.is_pinned ? '取消置顶' : '置顶' }}
        </el-button>
      </div>
    </div>

    <!-- 加载状态 -->
    <el-card v-if="loading" class="loading-card">
      <div class="loading-content">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载中...</span>
      </div>
    </el-card>

    <!-- 话题详情 -->
    <div v-else-if="discussion" class="discussion-detail">
      <!-- 话题主体 -->
      <el-card class="discussion-card">
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
        
        <h1 class="discussion-title">{{ discussion.title }}</h1>
        
        <div class="discussion-info">
          <div class="author-info">
            <el-avatar :size="32" :src="discussion.author.avatar" />
            <div class="author-details">
              <div class="author-name">{{ discussion.author.name }}</div>
              <div class="author-time">{{ formatDate(discussion.created_at) }}</div>
            </div>
          </div>
          
          <div class="discussion-stats">
            <div class="stat-item">
              <el-icon><View /></el-icon>
              <span>{{ discussion.view_count }}</span>
            </div>
            <div class="stat-item">
              <el-icon><ChatDotRound /></el-icon>
              <span>{{ discussion.reply_count }}</span>
            </div>
            <div class="stat-item">
              <el-button 
                :type="isLiked ? 'primary' : 'default'" 
                :icon="Star"
                @click="toggleLike"
                size="small"
              >
                {{ discussion.like_count }}
              </el-button>
            </div>
          </div>
        </div>
        
        <div class="discussion-content">
          <div v-html="formatContent(discussion.content)"></div>
        </div>
        
        <div class="discussion-tags" v-if="discussion.tags">
          <el-tag 
            v-for="tag in getTagList(discussion.tags)" 
            :key="tag"
            size="small"
            class="tag-item"
          >
            {{ tag }}
          </el-tag>
        </div>
      </el-card>

      <!-- 回复列表 -->
      <div class="replies-section">
        <div class="replies-header">
          <h3>回复 ({{ replies.length }})</h3>
          <el-button 
            type="primary" 
            :icon="Plus"
            @click="showReplyDialog = true"
            v-if="canReply"
          >
            添加回复
          </el-button>
        </div>
        
        <div class="replies-list">
          <div 
            v-for="reply in replies" 
            :key="reply.id"
            class="reply-item"
            :class="{ 'is-resolved': reply.is_resolved }"
          >
            <div class="reply-avatar">
              <el-avatar :size="32" :src="reply.author.avatar" />
            </div>
            <div class="reply-content">
              <div class="reply-header">
                <div class="reply-author">{{ reply.author.name }}</div>
                <div class="reply-time">{{ formatDate(reply.created_at) }}</div>
                <div class="reply-actions">
                  <el-button 
                    v-if="canResolve" 
                    :type="reply.is_resolved ? 'success' : 'default'"
                    size="small"
                    @click="toggleResolve(reply)"
                  >
                    {{ reply.is_resolved ? '已解决' : '标记解决' }}
                  </el-button>
                </div>
              </div>
              <div class="reply-body">
                <div v-html="formatContent(reply.content)"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="showEditDialog"
      title="编辑话题"
      width="60%"
      @close="resetEditForm"
    >
      <el-form
        ref="editFormRef"
        :model="editForm"
        :rules="editRules"
        label-width="80px"
      >
        <el-form-item label="标题" prop="title">
          <el-input v-model="editForm.title" />
        </el-form-item>
        
        <el-form-item label="分类" prop="category">
          <el-select v-model="editForm.category">
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
            v-model="editForm.content"
            type="textarea"
            :rows="8"
          />
        </el-form-item>
        
        <el-form-item label="标签">
          <el-input v-model="editForm.tags" />
        </el-form-item>
        
        <el-form-item>
          <el-checkbox v-model="editForm.is_public">公开话题</el-checkbox>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="updateDiscussion" :loading="updating">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 回复对话框 -->
    <el-dialog
      v-model="showReplyDialog"
      title="添加回复"
      width="50%"
      @close="resetReplyForm"
    >
      <el-form
        ref="replyFormRef"
        :model="replyForm"
        :rules="replyRules"
        label-width="80px"
      >
        <el-form-item label="内容" prop="content">
          <el-input
            v-model="replyForm.content"
            type="textarea"
            :rows="6"
            placeholder="输入回复内容，支持Markdown格式"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showReplyDialog = false">取消</el-button>
        <el-button type="primary" @click="createReply" :loading="replying">
          发表回复
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  ArrowLeft, 
  Edit, 
  Delete, 
  Top, 
  Loading, 
  Plus, 
  View, 
  ChatDotRound, 
  Star 
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { ApiService } from '@/services/api'
import type { 
  Discussion, 
  DiscussionReply, 
  UpdateDiscussionRequest,
  CreateReplyRequest 
} from '@/types/discussion'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 响应式数据
const loading = ref(false)
const updating = ref(false)
const replying = ref(false)
const showEditDialog = ref(false)
const showReplyDialog = ref(false)
const discussion = ref<Discussion | null>(null)
const replies = ref<DiscussionReply[]>([])
const isLiked = ref(false)
const canEdit = ref(false)
const canDelete = ref(false)
const categories = ref<string[]>([])

// 编辑表单
const editForm = reactive<UpdateDiscussionRequest>({
  title: '',
  content: '',
  category: '',
  tags: '',
  is_public: true
})

// 回复表单
const replyForm = reactive<CreateReplyRequest>({
  content: '',
  parent_reply_id: undefined
})

// 表单验证规则
const editRules = {
  title: [
    { required: true, message: '请输入标题', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入内容', trigger: 'blur' }
  ]
}

const replyRules = {
  content: [
    { required: true, message: '请输入回复内容', trigger: 'blur' }
  ]
}

// 计算属性
const canReply = computed(() => {
  return discussion.value && discussion.value.status === 'open'
})

const canPin = computed(() => {
  return authStore.userRole === 1 || authStore.userRole === 2
})

const canResolve = computed(() => {
  return authStore.userRole === 1 || authStore.userRole === 2
})

// 组件挂载时加载数据
onMounted(async () => {
  const discussionId = route.params.id
  if (discussionId) {
    await loadDiscussion(Number(discussionId))
    await loadCategories()
  }
})

// 加载话题详情
const loadDiscussion = async (id: number) => {
  loading.value = true
  try {
    const response = await ApiService.getDiscussionDetail(id)
    discussion.value = response.data.discussion
    replies.value = response.data.replies || []
    isLiked.value = response.data.is_liked || false
    canEdit.value = response.data.can_edit || false
    canDelete.value = response.data.can_delete || false
  } catch (error) {
    console.error('加载话题详情失败:', error)
    ElMessage.error('加载话题详情失败')
  } finally {
    loading.value = false
  }
}

// 加载分类
const loadCategories = async () => {
  try {
    const response = await ApiService.getDiscussionCategories()
    categories.value = response.data?.categories || []
  } catch (error) {
    categories.value = ['general', 'question', 'announcement', 'help', 'feedback']
  }
}

// 返回列表
const goBack = () => {
  router.go(-1)
}

// 切换点赞
const toggleLike = async () => {
  if (!discussion.value) return
  
  try {
    if (isLiked.value) {
      await ApiService.unlikeDiscussion(discussion.value.id)
      discussion.value.like_count--
      isLiked.value = false
      ElMessage.success('取消点赞')
    } else {
      await ApiService.likeDiscussion(discussion.value.id)
      discussion.value.like_count++
      isLiked.value = true
      ElMessage.success('点赞成功')
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 置顶话题
const pinDiscussion = async () => {
  if (!discussion.value) return
  
  try {
    await ApiService.pinDiscussion(discussion.value.id)
    discussion.value.is_pinned = !discussion.value.is_pinned
    ElMessage.success(discussion.value.is_pinned ? '置顶成功' : '取消置顶成功')
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 删除话题
const deleteDiscussion = async () => {
  if (!discussion.value) return
  
  try {
    await ElMessageBox.confirm('确定要删除这个话题吗？', '确认删除', {
      type: 'warning'
    })
    
    await ApiService.deleteDiscussion(discussion.value.id)
    ElMessage.success('删除成功')
    router.push('/discussions')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 更新话题
const updateDiscussion = async () => {
  if (!discussion.value) return
  
  const formRef = ref()
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch (error) {
    return
  }
  
  updating.value = true
  try {
    await ApiService.updateDiscussion(discussion.value.id, editForm)
    ElMessage.success('更新成功')
    showEditDialog.value = false
    await loadDiscussion(discussion.value.id)
  } catch (error) {
    ElMessage.error('更新失败')
  } finally {
    updating.value = false
  }
}

// 创建回复
const createReply = async () => {
  if (!discussion.value) return
  
  const formRef = ref()
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch (error) {
    return
  }
  
  replying.value = true
  try {
    await ApiService.createReply(discussion.value.id, replyForm)
    ElMessage.success('回复成功')
    showReplyDialog.value = false
    resetReplyForm()
    await loadDiscussion(discussion.value.id)
  } catch (error) {
    ElMessage.error('回复失败')
  } finally {
    replying.value = false
  }
}

// 切换解决状态
const toggleResolve = async (reply: DiscussionReply) => {
  try {
    // 这里需要添加标记解决的API
    reply.is_resolved = !reply.is_resolved
    ElMessage.success(reply.is_resolved ? '标记为已解决' : '取消解决标记')
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 重置编辑表单
const resetEditForm = () => {
  if (discussion.value) {
    Object.assign(editForm, {
      title: discussion.value.title,
      content: discussion.value.content,
      category: discussion.value.category,
      tags: discussion.value.tags,
      is_public: discussion.value.is_public
    })
  }
}

// 重置回复表单
const resetReplyForm = () => {
  Object.assign(replyForm, {
    content: '',
    parent_reply_id: null
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

// 格式化内容（支持Markdown）
const formatContent = (content: string) => {
  // 简单的Markdown转换
  return content
    .replace(/\n/g, '<br>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
}

// 获取标签列表
const getTagList = (tags: string) => {
  return tags ? tags.split(',').map(tag => tag.trim()).filter(tag => tag) : []
}

// 初始化编辑表单
onMounted(() => {
  resetEditForm()
})
</script>

<style scoped>
.discussion-detail-container {
  padding: 20px;
  background-color: #f8f9fa;
  min-height: 100vh;
}

.discussion-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.back-button {
  margin-right: 20px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.loading-card {
  border: none;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.loading-content {
  text-align: center;
  padding: 40px;
  color: #909399;
}

.discussion-detail {
  max-width: 1200px;
  margin: 0 auto;
}

.discussion-card {
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.discussion-meta {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.pin-tag, .category-tag {
  margin-right: 8px;
}

.discussion-title {
  margin: 0 0 20px 0;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  line-height: 1.4;
}

.discussion-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f0f2f5;
}

.author-info {
  display: flex;
  align-items: center;
}

.author-details {
  margin-left: 12px;
}

.author-name {
  font-weight: 600;
  color: #303133;
}

.author-time {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.discussion-stats {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-item {
  display: flex;
  align-items: center;
  color: #909399;
  font-size: 14px;
}

.stat-item .el-icon {
  margin-right: 4px;
}

.discussion-content {
  margin-bottom: 20px;
  line-height: 1.6;
  color: #303133;
}

.discussion-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-item {
  margin: 0;
}

.replies-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.replies-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f2f5;
}

.replies-header h3 {
  margin: 0;
  color: #303133;
}

.replies-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.reply-item {
  display: flex;
  gap: 12px;
  padding: 16px;
  border-radius: 8px;
  background: #f8f9fa;
}

.reply-item.is-resolved {
  background: #f0f9f0;
  border-left: 4px solid #67c23a;
}

.reply-avatar {
  flex-shrink: 0;
}

.reply-content {
  flex: 1;
  min-width: 0;
}

.reply-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.reply-author {
  font-weight: 600;
  color: #303133;
}

.reply-time {
  font-size: 12px;
  color: #909399;
}

.reply-actions {
  display: flex;
  gap: 8px;
}

.reply-body {
  color: #303133;
  line-height: 1.6;
}

.el-select {
  width: 100%;
}
</style> 