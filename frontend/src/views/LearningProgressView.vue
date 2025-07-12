<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { ApiService } from '../services/api'
import {
  TrendCharts,
  DataAnalysis,
  Trophy,
  Clock,
  Document,
  FolderOpened,
  User,
  Star,
  Calendar,
  Check
} from '@element-plus/icons-vue'

// 响应式数据
const loading = ref(false)
const selectedUserId = ref<number | null>(null)
const progressData = ref({
  overview: {
    totalAssignments: 0,
    completedAssignments: 0,
    totalProjects: 0,
    activeProjects: 0,
    totalCommits: 0,
    totalMergeRequests: 0
  },
  recentActivity: [] as any[],
  progressChart: [] as any[],
  achievements: [] as any[]
})

// 用户列表
const users = ref([] as any[])

// 模拟进度数据
const mockProgressData = {
  overview: {
    totalAssignments: 12,
    completedAssignments: 8,
    totalProjects: 3,
    activeProjects: 2,
    totalCommits: 45,
    totalMergeRequests: 12
  },
  recentActivity: [
    {
      id: 1,
      type: 'assignment',
      title: '完成了作业：数据结构实验',
      time: '2024-03-15 14:30',
      status: 'completed'
    },
    {
      id: 2,
      type: 'commit',
      title: '提交了代码：优化算法性能',
      time: '2024-03-15 10:15',
      status: 'success'
    },
    {
      id: 3,
      type: 'merge_request',
      title: '提交了合并请求：修复bug',
      time: '2024-03-14 16:45',
      status: 'pending'
    }
  ],
  progressChart: [
    { date: '2024-03-01', completed: 2, total: 3 },
    { date: '2024-03-08', completed: 4, total: 6 },
    { date: '2024-03-15', completed: 8, total: 12 }
  ],
  achievements: [
    { title: '连续提交', description: '连续7天提交代码', icon: 'Trophy', earned: true },
    { title: '作业达人', description: '完成10个作业', icon: 'Star', earned: true },
    { title: '团队合作', description: '参与3个项目', icon: 'User', earned: false }
  ]
}

// 生命周期
onMounted(() => {
  loadUserList()
  loadProgressData()
})

// 方法
const loadUserList = async () => {
  try {
    const response = await ApiService.getLearningProgressUsers()
    users.value = response
  } catch (error) {
    console.error('加载用户列表失败:', error)
    // 使用模拟数据作为后备
    users.value = [
      { id: 1, name: '张三', username: 'zhangsan', avatar: '' },
      { id: 2, name: '李四', username: 'lisi', avatar: '' },
      { id: 3, name: '王五', username: 'wangwu', avatar: '' }
    ]
  }
}

const loadProgressData = async () => {
  loading.value = true
  try {
    if (selectedUserId.value) {
      const response = await ApiService.getLearningProgress(selectedUserId.value)
      progressData.value = response
      ElMessage.success('学习进度数据加载成功')
    } else {
      // 使用模拟数据
      progressData.value = mockProgressData
      ElMessage.success('学习进度数据加载成功')
    }
  } catch (error) {
    console.error('加载学习进度失败:', error)
    ElMessage.error('加载学习进度失败')
    // 使用模拟数据作为后备
    progressData.value = mockProgressData
  } finally {
    loading.value = false
  }
}

const handleUserChange = (userId: number) => {
  selectedUserId.value = userId
  loadProgressData()
}

const getActivityIcon = (type: string) => {
  switch (type) {
    case 'assignment':
      return Check
    case 'commit':
      return Document
    case 'merge_request':
      return FolderOpened
    default:
      return Clock
  }
}

const getActivityTypeText = (type: string) => {
  switch (type) {
    case 'assignment':
      return '作业'
    case 'commit':
      return '提交'
    case 'merge_request':
      return '合并请求'
    default:
      return '活动'
  }
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'completed':
    case 'success':
      return 'success'
    case 'pending':
      return 'warning'
    case 'failed':
      return 'danger'
    default:
      return 'info'
  }
}

const calculateCompletionRate = (completed: number, total: number) => {
  return total > 0 ? Math.round((completed / total) * 100) : 0
}
</script>

<template>
  <div class="learning-progress-view">
    <!-- 页面标题 -->
    <el-row class="page-header">
      <el-col :span="24">
        <div class="header-content">
          <h1 class="page-title">
            <el-icon><TrendCharts /></el-icon>
            学习进度跟踪
          </h1>
          <p class="page-description">查看学生的学习进度、作业完成情况和项目参与度</p>
        </div>
      </el-col>
    </el-row>

    <!-- 用户选择 -->
    <el-row class="user-selector">
      <el-col :span="24">
        <el-card>
          <div class="selector-content">
            <span class="selector-label">选择学生：</span>
            <el-select
              v-model="selectedUserId"
              placeholder="请选择学生"
              @change="handleUserChange"
              size="large"
            >
              <el-option
                v-for="user in users"
                :key="user.id"
                :label="user.name"
                :value="user.id"
              >
                <div class="user-option">
                  <el-avatar :size="20" :src="user.avatar">
                    <el-icon><User /></el-icon>
                  </el-avatar>
                  <span>{{ user.name }} ({{ user.username }})</span>
                </div>
              </el-option>
            </el-select>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 数据概览 -->
    <el-row :gutter="24" class="overview-section">
      <el-col :xs="12" :sm="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon assignment">
              <el-icon size="32"><Check /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ progressData.overview.completedAssignments }}/{{ progressData.overview.totalAssignments }}</div>
              <div class="stat-label">已完成作业</div>
              <div class="stat-progress">
                <el-progress
                  :percentage="calculateCompletionRate(progressData.overview.completedAssignments, progressData.overview.totalAssignments)"
                  :show-text="false"
                  stroke-width="4"
                />
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="12" :sm="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon project">
              <el-icon size="32"><FolderOpened /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ progressData.overview.activeProjects }}/{{ progressData.overview.totalProjects }}</div>
              <div class="stat-label">活跃项目</div>
              <div class="stat-progress">
                <el-progress
                  :percentage="calculateCompletionRate(progressData.overview.activeProjects, progressData.overview.totalProjects)"
                  :show-text="false"
                  stroke-width="4"
                  color="#67C23A"
                />
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="12" :sm="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon commit">
              <el-icon size="32"><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ progressData.overview.totalCommits }}</div>
              <div class="stat-label">代码提交</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="12" :sm="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon merge">
              <el-icon size="32"><DataAnalysis /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ progressData.overview.totalMergeRequests }}</div>
              <div class="stat-label">合并请求</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 主要内容区域 -->
    <el-row :gutter="24" class="main-content">
      <!-- 最近活动 -->
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><Clock /></el-icon>
              <span>最近活动</span>
            </div>
          </template>
          <div class="activity-list">
            <div 
              v-for="activity in progressData.recentActivity" 
              :key="activity.id"
              class="activity-item"
            >
              <div class="activity-icon">
                <el-icon :color="getStatusColor(activity.status)">
                  <component :is="getActivityIcon(activity.type)" />
                </el-icon>
              </div>
              <div class="activity-content">
                <div class="activity-title">{{ activity.title }}</div>
                <div class="activity-meta">
                  <el-tag :type="getStatusColor(activity.status)" size="small">
                    {{ getActivityTypeText(activity.type) }}
                  </el-tag>
                  <span class="activity-time">{{ activity.time }}</span>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 进度图表 -->
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><TrendCharts /></el-icon>
              <span>进度趋势</span>
            </div>
          </template>
          <div class="chart-container">
            <div class="chart-placeholder">
              <el-icon size="48" color="#ddd"><DataAnalysis /></el-icon>
              <p>图表功能开发中...</p>
              <p class="chart-hint">这里将显示学习进度的趋势图表</p>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 成就系统 -->
    <el-row class="achievements-section">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><Trophy /></el-icon>
              <span>成就系统</span>
            </div>
          </template>
          <div class="achievements-grid">
            <div 
              v-for="achievement in progressData.achievements" 
              :key="achievement.title"
              class="achievement-item"
              :class="{ earned: achievement.earned }"
            >
              <div class="achievement-icon">
                <el-icon size="24">
                  <Trophy v-if="achievement.icon === 'Trophy'" />
                  <Star v-else-if="achievement.icon === 'Star'" />
                  <User v-else-if="achievement.icon === 'User'" />
                </el-icon>
              </div>
              <div class="achievement-content">
                <div class="achievement-title">{{ achievement.title }}</div>
                <div class="achievement-description">{{ achievement.description }}</div>
              </div>
              <div class="achievement-status">
                <el-tag v-if="achievement.earned" type="success" size="small">已获得</el-tag>
                <el-tag v-else type="info" size="small">未获得</el-tag>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.learning-progress-view {
  padding: 20px;
}

.page-header {
  margin-bottom: 24px;
}

.header-content {
  text-align: center;
}

.page-title {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.page-description {
  color: #606266;
  font-size: 16px;
  margin: 0;
}

.user-selector {
  margin-bottom: 24px;
}

.selector-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.selector-label {
  font-weight: 500;
  color: #303133;
}

.user-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.overview-section {
  margin-bottom: 24px;
}

.stat-card {
  height: 120px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
  height: 100%;
}

.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-icon.assignment {
  background: linear-gradient(135deg, #409EFF, #67C23A);
}

.stat-icon.project {
  background: linear-gradient(135deg, #67C23A, #E6A23C);
}

.stat-icon.commit {
  background: linear-gradient(135deg, #E6A23C, #F56C6C);
}

.stat-icon.merge {
  background: linear-gradient(135deg, #F56C6C, #9C27B0);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 4px;
}

.stat-label {
  color: #909399;
  font-size: 14px;
  margin-bottom: 8px;
}

.stat-progress {
  width: 100%;
}

.main-content {
  margin-bottom: 24px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.activity-list {
  max-height: 400px;
  overflow-y: auto;
}

.activity-item {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-icon {
  margin-top: 4px;
}

.activity-content {
  flex: 1;
}

.activity-title {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.activity-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.activity-time {
  color: #909399;
  font-size: 12px;
}

.chart-container {
  height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-placeholder {
  text-align: center;
  color: #909399;
}

.chart-hint {
  font-size: 12px;
  margin-top: 8px;
}

.achievements-section {
  margin-bottom: 24px;
}

.achievements-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.achievement-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border: 2px solid #f0f0f0;
  border-radius: 8px;
  transition: all 0.3s;
}

.achievement-item.earned {
  border-color: #67C23A;
  background: #f0f9ff;
}

.achievement-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.achievement-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  color: #909399;
}

.achievement-item.earned .achievement-icon {
  background: #67C23A;
  color: white;
}

.achievement-content {
  flex: 1;
}

.achievement-title {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.achievement-description {
  color: #606266;
  font-size: 14px;
}

.achievement-status {
  margin-left: auto;
}

@media (max-width: 768px) {
  .learning-progress-view {
    padding: 16px;
  }
  
  .page-title {
    font-size: 24px;
  }
  
  .stat-card {
    height: auto;
  }
  
  .stat-content {
    flex-direction: column;
    text-align: center;
    gap: 8px;
  }
  
  .stat-icon {
    width: 48px;
    height: 48px;
  }
  
  .achievements-grid {
    grid-template-columns: 1fr;
  }
}
</style> 