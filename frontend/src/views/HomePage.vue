<template>
  <div class="home-page">
    <!-- 快速访问区域 -->
    <div class="quick-access card">
      <h2>快速访问</h2>
      <div class="quick-access-grid">
        <router-link to="/scenes" class="quick-access-item">
          <div class="quick-access-icon">📋</div>
          <div class="quick-access-title">课题列表</div>
        </router-link>
        <router-link to="/topics" class="quick-access-item">
          <div class="quick-access-icon">💬</div>
          <div class="quick-access-title">话题列表</div>
        </router-link>
        <router-link to="/documents" class="quick-access-item">
          <div class="quick-access-icon">📄</div>
          <div class="quick-access-title">文档列表</div>
        </router-link>
      </div>
    </div>

    <!-- 主要内容区域 -->
    <div class="main-content">
      <div class="content-column">
        <!-- 热门课题 -->
        <div class="popular-scenes card">
          <div class="section-header">
            <h2>热门课题</h2>
            <div class="section-controls">
              <div class="search-box">
                <el-input
                  v-model="searchQuery"
                  placeholder="搜索课题..."
                  clearable
                  @input="handleSearch"
                >
                  <template #prefix>
                    <el-icon><Search /></el-icon>
                  </template>
                </el-input>
              </div>
              <el-button v-if="canCreateScene" type="primary" @click="createScene">
                创建课题
              </el-button>
            </div>
          </div>

          <div v-loading="loading" class="scene-list">
            <el-row :gutter="20">
              <el-col
                v-for="scene in scenes"
                :key="scene.id"
                :xs="24"
                :sm="12"
                :lg="8"
                class="scene-col"
              >
                <div class="scene-card" @click="viewScene(scene.id)">
                  <div class="scene-header">
                    <h3 class="scene-title">{{ scene.name }}</h3>
                    <el-tag :type="scene.is_public ? 'success' : 'warning'" size="small">
                      {{ scene.is_public ? '公开' : '私有' }}
                    </el-tag>
                  </div>
                  <p class="scene-description">{{ scene.description }}</p>
                  <div class="scene-meta">
                    <span>创建于 {{ formatDate(scene.created_at) }}</span>
                    <span>· {{ scene.view_count || 0 }} 次访问</span>
                  </div>
                  <div class="scene-tags">
                    <el-tag
                      v-for="tag in scene.tags?.slice(0, 3) || []"
                      :key="tag"
                      size="small"
                      class="scene-tag"
                    >
                      {{ tag }}
                    </el-tag>
                  </div>
                </div>
              </el-col>
            </el-row>

            <el-pagination
              v-if="total > pageSize"
              v-model:current-page="currentPage"
              :total="total"
              :page-size="pageSize"
              layout="prev, pager, next"
              @current-change="handlePageChange"
            />
          </div>
        </div>

        <!-- 我的课题 -->
        <div v-if="userStore.isLoggedIn" class="my-scenes card">
          <h3>我的课题</h3>
          <div v-if="myScenes.length > 0" class="my-scenes-list">
            <div
              v-for="scene in myScenes"
              :key="scene.id"
              class="my-scene-item"
            >
              <div class="scene-info">
                <h4>{{ scene.name }}</h4>
                <p>{{ scene.description }}</p>
                <div class="scene-stats">
                  <span>参与者: {{ scene.members?.length || 0 }}</span>
                  <span>创建时间: {{ formatDate(scene.created_at) }}</span>
                </div>
              </div>
              <div class="scene-actions">
                <el-button size="small" @click="viewScene(scene.id)">
                  查看
                </el-button>
                <el-button size="small" type="primary" @click="editScene(scene.id)">
                  编辑
                </el-button>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">
            <el-empty description="暂无参与的课题" />
          </div>
        </div>
      </div>

      <!-- 侧边栏 -->
      <div class="sidebar">
        <!-- 热门话题 -->
        <div class="hot-topics card">
          <h3>热门话题</h3>
          <div v-loading="topicsLoading" class="topics-list">
            <div
              v-for="topic in hotTopics"
              :key="topic.id"
              class="topic-item"
              @click="viewTopic(topic.id)"
            >
              <div class="topic-header">
                <h4 class="topic-title">{{ topic.title }}</h4>
                <div class="topic-likes">
                  <el-icon><Star /></el-icon>
                  {{ topic.likes_count }}
                </div>
              </div>
              <p class="topic-summary">{{ topic.content.substring(0, 50) }}...</p>
              <div class="topic-meta">
                <span>{{ topic.author.username }}</span>
                <span>{{ formatDate(topic.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 最近活动 -->
        <div class="recent-activities card">
          <h3>最近活动</h3>
          <div v-loading="activitiesLoading" class="activity-list">
            <div
              v-for="activity in recentActivities"
              :key="activity.id"
              class="activity-item"
              @click="viewActivity(activity)"
            >
              <el-icon class="activity-icon">
                <Document v-if="activity.type === 'document'" />
                <ChatLineRound v-else-if="activity.type === 'topic'" />
                <Collection v-else-if="activity.type === 'homework'" />
                <Message v-else-if="activity.type === 'comment'" />
                <Star v-else />
              </el-icon>
              <div class="activity-content">
                <p>{{ activity.title }}</p>
                <div class="activity-description">{{ activity.description }}</div>
                <div class="activity-meta">
                  <span class="activity-user">{{ activity.user_name }}</span>
                  <span v-if="activity.project_name" class="activity-project">
                    · {{ activity.project_name }}
                  </span>
                  <span class="activity-time">· {{ formatRelativeTime(activity.created_at) }}</span>
                </div>
              </div>
            </div>
            <div v-if="recentActivities.length === 0 && !activitiesLoading" class="empty-state">
              <el-empty description="暂无最近活动" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建课题对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      title="创建新课题"
      width="600px"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="80px"
      >
        <el-form-item label="课题名称" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入课题名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入课题描述"
          />
        </el-form-item>
        <el-form-item label="可见性" prop="visibility">
          <el-select v-model="createForm.visibility" style="width: 100%">
            <el-option label="公开" value="public" />
            <el-option label="私有" value="private" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签" prop="tags">
          <el-select
            v-model="createForm.tags"
            multiple
            filterable
            allow-create
            placeholder="请输入标签"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="confirmCreate">
          创建
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { researchService, topicService, activityService } from '@/services/api'
import type { Scene, Topic, ActivityItem } from '@/types'
import { ElMessage } from 'element-plus'
import { handleApiError, showSuccess } from '@/utils/errorHandler'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const scenes = ref<Scene[]>([])
const myScenes = ref<Scene[]>([])
const hotTopics = ref<Topic[]>([])
const recentActivities = ref<ActivityItem[]>([])
const searchQuery = ref('')
const loading = ref(false)
const topicsLoading = ref(false)
const activitiesLoading = ref(false)
const currentPage = ref(1)
const total = ref(0)
const pageSize = 12
const creating = ref(false)
const createDialogVisible = ref(false)

const createForm = ref({
  name: '',
  description: '',
  visibility: 'public' as 'public' | 'private',
  tags: [] as string[]
})

const createRules = {
  name: [{ required: true, message: '请输入课题名称', trigger: 'blur' }],
  description: [{ required: true, message: '请输入课题描述', trigger: 'blur' }]
}

// 权限相关状态
const canCreateScene = ref(false)

// 检查权限
const checkPermissions = async () => {
  if (!userStore.isLoggedIn) {
    canCreateScene.value = false
    return
  }

  try {
    const canCreate = await userStore.checkPermission('create', 'project')
    canCreateScene.value = canCreate
  } catch (error) {
    console.error('权限检查失败:', error)
    canCreateScene.value = false
  }
}

// 方法
const fetchScenes = async (page = 1) => {
  loading.value = true
  try {
    const response: any = await researchService.getProjects({
      page,
      pageSize,
      search: searchQuery.value || undefined
    })
    
    // 处理响应数据结构（axios拦截器已经返回response.data）
    if (response && response.projects) {
      scenes.value = response.projects || []
      total.value = response.pagination?.total || response.projects.length || 0
    } else if (Array.isArray(response)) {
      scenes.value = response
      total.value = response.length
    } else {
      scenes.value = []
      total.value = 0
    }
  } catch (error: any) {
    console.error('获取课题列表失败:', error)
    // 只在非404错误时显示错误提示
    if (error.response?.status !== 404) {
      handleApiError(error, '获取课题列表')
    }
    scenes.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const fetchMyScenes = async () => {
  if (!userStore.isLoggedIn) return
  
  try {
    const response: any = await researchService.getProjects({
      ownerId: userStore.user?.id.toString()
    })
    
    // 处理响应数据结构（axios拦截器已经返回response.data）
    if (response && response.projects) {
      myScenes.value = response.projects || []
    } else if (Array.isArray(response)) {
      myScenes.value = response
    } else {
      myScenes.value = []
    }
  } catch (error: any) {
    console.error('获取我的课题失败:', error)
    if (error.response?.status !== 404) {
      ElMessage.error('获取我的课题失败')
    }
    myScenes.value = []
  }
}

const fetchHotTopics = async () => {
  topicsLoading.value = true
  try {
    const response: any = await topicService.getTopics({
      page: 1,
      pageSize: 5
    })
    
    // 处理响应数据结构（axios拦截器已经返回response.data）
    if (response && response.topics) {
      hotTopics.value = response.topics || []
    } else if (Array.isArray(response)) {
      hotTopics.value = response
    } else {
      hotTopics.value = []
    } 
  } catch (error: any) {
    console.error('获取热门话题失败:', error)
    if (error.response?.status !== 404) {
      ElMessage.error('获取热门话题失败')
    }
    hotTopics.value = []
  } finally {
    topicsLoading.value = false
  }
}

const fetchRecentActivities = async () => {
  activitiesLoading.value = true
  try {
    const response = await activityService.getRecentActivities(8)
    
    // 处理响应数据结构（axios拦截器已经返回response.data）
    if (response && response.data) {
      recentActivities.value = response.data
    } else if (Array.isArray(response)) {
      recentActivities.value = response
    } else {
      recentActivities.value = []
    }
  } catch (error: any) {
    console.error('获取最近活动失败:', error)
    if (error.response?.status !== 404) {
      ElMessage.error('获取最近活动失败')
    }
    recentActivities.value = []
  } finally {
    activitiesLoading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchScenes()
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  fetchScenes(page)
}

const createScene = () => {
  createDialogVisible.value = true
}

const confirmCreate = async () => {
  creating.value = true
  try {
    await researchService.createProject(createForm.value)
    ElMessage.success('课题创建成功')
    createDialogVisible.value = false
    createForm.value = {
      name: '',
      description: '',
      visibility: 'public',
      tags: []
    }
    fetchScenes()
    fetchMyScenes()
  } catch (error) {
    handleApiError(error, '课题创建')
  } finally {
    creating.value = false
  }
}

const viewScene = (id: string) => {
  router.push(`/scenes/${id}`)
}

const editScene = (id: string) => {
  router.push(`/scenes/${id}/edit`)
}

const viewTopic = (id: string) => {
  router.push(`/topics/${id}`)
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('zh-CN')
}

const formatRelativeTime = (date: string) => {
  const now = new Date()
  const activityDate = new Date(date)
  const diffInSeconds = Math.floor((now.getTime() - activityDate.getTime()) / 1000)

  if (diffInSeconds < 60) {
    return '刚刚'
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60)
    return `${minutes}分钟前`
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600)
    return `${hours}小时前`
  } else if (diffInSeconds < 604800) {
    const days = Math.floor(diffInSeconds / 86400)
    return `${days}天前`
  } else {
    return formatDate(date)
  }
}

const viewActivity = (activity: ActivityItem) => {
  router.push(activity.url)
}

// 生命周期
onMounted(async () => {
  await checkPermissions()
  fetchScenes()
  fetchMyScenes()
  fetchHotTopics()
  fetchRecentActivities()
})
</script>

<style scoped>
.home-page {
  max-width: 1200px;
  margin: 0 auto;
}

.quick-access {
  margin-bottom: 30px;
}

.quick-access h2 {
  margin-bottom: 20px;
  color: var(--primary-color);
}

.quick-access-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 20px;
}

.quick-access-item {
  background-color: var(--card-background);
  border-radius: 8px;
  padding: 20px;
  text-align: center;
  transition: transform 0.3s, box-shadow 0.3s;
  display: flex;
  flex-direction: column;
  align-items: center;
  border: 1px solid var(--border-color);
}

.quick-access-item:hover {
  transform: translateY(-5px);
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.2);
  text-decoration: none;
}

.quick-access-icon {
  font-size: 32px;
  margin-bottom: 10px;
}

.quick-access-title {
  font-weight: 500;
  color: var(--text-color);
}

.main-content {
  display: flex;
  gap: 30px;
}

.content-column {
  flex: 3;
}

.sidebar {
  flex: 1;
  min-width: 300px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.search-box {
  max-width: 300px;
}

.scene-list {
  min-height: 200px;
}

.scene-col {
  margin-bottom: 20px;
}

.scene-card {
  background-color: var(--card-background);
  border-radius: 8px;
  padding: 20px;
  cursor: pointer;
  transition: transform 0.3s, box-shadow 0.3s;
  border: 1px solid var(--border-color);
  height: 100%;
}

.scene-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.scene-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.scene-title {
  margin: 0;
  font-size: 18px;
  color: var(--primary-color);
  flex: 1;
  margin-right: 10px;
}

.scene-description {
  margin: 0 0 12px 0;
  color: var(--light-text);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.scene-meta {
  font-size: 12px;
  color: var(--lighter-text);
  margin-bottom: 12px;
}

.scene-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.scene-tag {
  margin: 0;
}

.my-scenes {
  margin-top: 30px;
}

.my-scenes h3 {
  margin-bottom: 20px;
  color: var(--primary-color);
}

.my-scene-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
}

.my-scene-item:last-child {
  border-bottom: none;
}

.scene-info h4 {
  margin: 0 0 8px 0;
  color: var(--primary-color);
}

.scene-info p {
  margin: 0 0 8px 0;
  color: var(--light-text);
  font-size: 14px;
}

.scene-stats {
  font-size: 12px;
  color: var(--lighter-text);
  display: flex;
  gap: 16px;
}

.scene-actions {
  display: flex;
  gap: 8px;
}

.hot-topics h3,
.recent-activities h3 {
  margin-bottom: 20px;
  color: var(--primary-color);
}

.topics-list {
  min-height: 200px;
}

.topic-item {
  padding: 16px;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background-color 0.3s;
}

.topic-item:hover {
  background-color: rgba(77, 121, 255, 0.1);
}

.topic-item:last-child {
  border-bottom: none;
}

.topic-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.topic-title {
  margin: 0;
  font-size: 16px;
  color: var(--text-color);
  flex: 1;
  margin-right: 8px;
}

.topic-likes {
  color: var(--accent-color);
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
}

.topic-summary {
  margin: 0 0 8px 0;
  color: var(--light-text);
  font-size: 14px;
  line-height: 1.4;
}

.topic-meta {
  font-size: 12px;
  color: var(--lighter-text);
  display: flex;
  justify-content: space-between;
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.activity-icon {
  font-size: 20px;
  color: var(--primary-color);
}

.activity-content p {
  margin: 0 0 4px 0;
  color: var(--text-color);
  font-weight: 500;
}

.activity-content span {
  font-size: 12px;
  color: var(--lighter-text);
}

.activity-description {
  font-size: 14px;
  color: var(--light-text);
  margin: 4px 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.activity-meta {
  font-size: 12px;
  color: var(--lighter-text);
  display: flex;
  align-items: center;
  gap: 4px;
}

.activity-user {
  font-weight: 500;
}

.activity-project {
  color: var(--primary-color);
}

.activity-time {
  color: var(--lighter-text);
}

.activity-item {
  cursor: pointer;
  transition: background-color 0.3s;
}

.activity-item:hover {
  background-color: rgba(77, 121, 255, 0.05);
  border-radius: 4px;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
}

@media (max-width: 768px) {
  .main-content {
    flex-direction: column;
  }
  
  .sidebar {
    min-width: 100%;
    margin-top: 20px;
  }
  
  .quick-access-grid {
    grid-template-columns: 1fr;
  }
  
  .section-header {
    flex-direction: column;
    align-items: stretch;
  }
  
  .search-box {
    max-width: none;
  }
  
  .my-scene-item {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .scene-actions {
    justify-content: flex-end;
  }
}
</style>
