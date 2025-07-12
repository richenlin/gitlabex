<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { ApiService } from '../services/api'
import {
  DataAnalysis,
  TrendCharts,
  PieChart,
  Histogram,
  Document,
  User,
  Trophy,
  Calendar,
  Download,
  Refresh,
  Filter,
  School,
  Notebook,
  Clock,
  DataBoard
} from '@element-plus/icons-vue'

interface ReportData {
  overview: {
    totalStudents: number
    totalAssignments: number
    totalProjects: number
    averageCompletion: number
  }
  classActivity: {
    className: string
    activeStudents: number
    totalStudents: number
    completionRate: number
  }[]
  assignmentStats: {
    name: string
    submitted: number
    total: number
    onTime: number
    late: number
  }[]
  progressTrends: {
    date: string
    completion: number
    participation: number
  }[]
  topPerformers: {
    id: number
    name: string
    completionRate: number
    points: number
  }[]
}

// 响应式数据
const loading = ref(false)
const reportData = ref<ReportData>({
  overview: {
    totalStudents: 0,
    totalAssignments: 0,
    totalProjects: 0,
    averageCompletion: 0
  },
  classActivity: [],
  assignmentStats: [],
  progressTrends: [],
  topPerformers: []
})

const selectedTimeRange = ref('week')
const selectedClass = ref<string>('all')

// 时间范围选项
const timeRangeOptions = [
  { value: 'week', label: '本周' },
  { value: 'month', label: '本月' },
  { value: 'quarter', label: '本季度' },
  { value: 'year', label: '本年度' }
]

// 班级选项
const classOptions = ref([
  { value: 'all', label: '所有班级' },
  { value: 'class1', label: '计算机科学1班' },
  { value: 'class2', label: '计算机科学2班' },
  { value: 'class3', label: '软件工程1班' }
])

// 模拟报表数据
const mockReportData: ReportData = {
  overview: {
    totalStudents: 125,
    totalAssignments: 45,
    totalProjects: 8,
    averageCompletion: 87.5
  },
  classActivity: [
    { className: '计算机科学1班', activeStudents: 28, totalStudents: 30, completionRate: 93.3 },
    { className: '计算机科学2班', activeStudents: 25, totalStudents: 32, completionRate: 78.1 },
    { className: '软件工程1班', activeStudents: 31, totalStudents: 33, completionRate: 93.9 },
    { className: '软件工程2班', activeStudents: 27, totalStudents: 30, completionRate: 90.0 }
  ],
  assignmentStats: [
    { name: '数据结构实验', submitted: 28, total: 30, onTime: 25, late: 3 },
    { name: '算法分析作业', submitted: 32, total: 33, onTime: 30, late: 2 },
    { name: 'Web开发项目', submitted: 25, total: 30, onTime: 20, late: 5 },
    { name: '数据库设计', submitted: 29, total: 32, onTime: 26, late: 3 }
  ],
  progressTrends: [
    { date: '2024-03-01', completion: 75, participation: 82 },
    { date: '2024-03-08', completion: 82, participation: 88 },
    { date: '2024-03-15', completion: 87, participation: 91 },
    { date: '2024-03-22', completion: 89, participation: 93 }
  ],
  topPerformers: [
    { id: 1, name: '张三', completionRate: 98.5, points: 245 },
    { id: 2, name: '李四', completionRate: 95.2, points: 238 },
    { id: 3, name: '王五', completionRate: 92.8, points: 232 },
    { id: 4, name: '赵六', completionRate: 90.5, points: 226 },
    { id: 5, name: '钱七', completionRate: 88.9, points: 222 }
  ]
}

// 计算属性
const totalActivityRate = computed(() => {
  if (reportData.value.classActivity.length === 0) return 0
  const totalActive = reportData.value.classActivity.reduce((sum, cls) => sum + cls.activeStudents, 0)
  const totalStudents = reportData.value.classActivity.reduce((sum, cls) => sum + cls.totalStudents, 0)
  return totalStudents > 0 ? Math.round((totalActive / totalStudents) * 100) : 0
})

const averageAssignmentCompletion = computed(() => {
  if (reportData.value.assignmentStats.length === 0) return 0
  const totalSubmitted = reportData.value.assignmentStats.reduce((sum, assignment) => sum + assignment.submitted, 0)
  const totalAssignments = reportData.value.assignmentStats.reduce((sum, assignment) => sum + assignment.total, 0)
  return totalAssignments > 0 ? Math.round((totalSubmitted / totalAssignments) * 100) : 0
})

// 生命周期
onMounted(() => {
  loadReportData()
})

// 方法
const loadReportData = async () => {
  loading.value = true
  try {
    const params = {
      time_range: selectedTimeRange.value,
      class: selectedClass.value
    }
    
    const response = await ApiService.getEducationReports(params)
    reportData.value = response
    ElMessage.success('报表数据加载成功')
  } catch (error) {
    console.error('加载报表数据失败:', error)
    ElMessage.error('加载报表数据失败')
    // 使用模拟数据作为后备
    reportData.value = mockReportData
  } finally {
    loading.value = false
  }
}

const exportReport = async () => {
  try {
    const params = {
      format: 'excel',
      time_range: selectedTimeRange.value,
      class: selectedClass.value
    }
    
    const response = await ApiService.exportEducationReport(params)
    ElMessage.success('报表导出成功')
    
    // 这里可以添加下载逻辑
    console.log('导出结果:', response)
  } catch (error) {
    console.error('导出报表失败:', error)
    ElMessage.error('导出报表失败')
  }
}

const refreshData = () => {
  loadReportData()
}

const onTimeRangeChange = () => {
  loadReportData()
}

const onClassChange = () => {
  loadReportData()
}

const getCompletionRateColor = (rate: number) => {
  if (rate >= 90) return '#67C23A'
  if (rate >= 70) return '#E6A23C'
  return '#F56C6C'
}

const getGradeLevel = (rate: number) => {
  if (rate >= 95) return 'A+'
  if (rate >= 90) return 'A'
  if (rate >= 85) return 'B+'
  if (rate >= 80) return 'B'
  if (rate >= 75) return 'C+'
  if (rate >= 70) return 'C'
  return 'D'
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit'
  })
}
</script>

<template>
  <div class="education-reports-view">
    <!-- 页面标题 -->
    <el-row class="page-header">
      <el-col :span="24">
        <div class="header-content">
          <h1 class="page-title">
            <el-icon><DataAnalysis /></el-icon>
            教育报表
          </h1>
          <p class="page-description">查看教育数据统计和分析报表</p>
        </div>
      </el-col>
    </el-row>

    <!-- 筛选控制栏 -->
    <el-row class="filter-bar">
      <el-col :span="24">
        <el-card>
          <div class="filter-content">
            <div class="filters">
              <div class="filter-item">
                <label>时间范围：</label>
                <el-select v-model="selectedTimeRange" @change="onTimeRangeChange">
                  <el-option
                    v-for="option in timeRangeOptions"
                    :key="option.value"
                    :label="option.label"
                    :value="option.value"
                  />
                </el-select>
              </div>
              <div class="filter-item">
                <label>班级：</label>
                <el-select v-model="selectedClass" @change="onClassChange">
                  <el-option
                    v-for="option in classOptions"
                    :key="option.value"
                    :label="option.label"
                    :value="option.value"
                  />
                </el-select>
              </div>
            </div>
            <div class="actions">
              <el-button @click="refreshData" :loading="loading">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
              <el-button @click="exportReport" type="primary">
                <el-icon><Download /></el-icon>
                导出报表
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 数据概览 -->
    <el-row :gutter="24" class="overview-section">
      <el-col :xs="12" :sm="6">
        <el-card class="overview-card">
          <div class="overview-content">
            <div class="overview-icon students">
              <el-icon size="32"><User /></el-icon>
            </div>
            <div class="overview-info">
              <div class="overview-value">{{ reportData.overview.totalStudents }}</div>
              <div class="overview-label">学生总数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="12" :sm="6">
        <el-card class="overview-card">
          <div class="overview-content">
            <div class="overview-icon assignments">
              <el-icon size="32"><Document /></el-icon>
            </div>
            <div class="overview-info">
              <div class="overview-value">{{ reportData.overview.totalAssignments }}</div>
              <div class="overview-label">作业总数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="12" :sm="6">
        <el-card class="overview-card">
          <div class="overview-content">
            <div class="overview-icon projects">
              <el-icon size="32"><Notebook /></el-icon>
            </div>
            <div class="overview-info">
              <div class="overview-value">{{ reportData.overview.totalProjects }}</div>
              <div class="overview-label">项目总数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="12" :sm="6">
        <el-card class="overview-card">
          <div class="overview-content">
            <div class="overview-icon completion">
              <el-icon size="32"><Trophy /></el-icon>
            </div>
            <div class="overview-info">
              <div class="overview-value">{{ reportData.overview.averageCompletion }}%</div>
              <div class="overview-label">平均完成率</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 主要报表区域 -->
    <el-row :gutter="24" class="main-reports">
      <!-- 班级活跃度 -->
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><School /></el-icon>
              <span>班级活跃度</span>
            </div>
          </template>
          <div class="class-activity-chart">
            <div v-loading="loading" class="activity-list">
              <div v-for="cls in reportData.classActivity" :key="cls.className" class="activity-item">
                <div class="activity-info">
                  <div class="class-name">{{ cls.className }}</div>
                  <div class="activity-stats">
                    <span>{{ cls.activeStudents }}/{{ cls.totalStudents }} 人</span>
                    <span :style="{ color: getCompletionRateColor(cls.completionRate) }">
                      {{ cls.completionRate }}%
                    </span>
                  </div>
                </div>
                <div class="activity-progress">
                  <el-progress
                    :percentage="cls.completionRate"
                    :stroke-width="8"
                    :color="getCompletionRateColor(cls.completionRate)"
                  />
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 作业完成统计 -->
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><Document /></el-icon>
              <span>作业完成统计</span>
            </div>
          </template>
          <div class="assignment-stats-chart">
            <div v-loading="loading" class="stats-table">
              <el-table :data="reportData.assignmentStats" stripe>
                <el-table-column prop="name" label="作业名称" />
                <el-table-column prop="submitted" label="已提交" align="center" />
                <el-table-column prop="total" label="总数" align="center" />
                <el-table-column label="完成率" align="center">
                  <template #default="scope">
                    <el-tag :type="scope.row.submitted / scope.row.total >= 0.8 ? 'success' : 'warning'">
                      {{ Math.round((scope.row.submitted / scope.row.total) * 100) }}%
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 进度趋势和排行榜 -->
    <el-row :gutter="24" class="secondary-reports">
      <!-- 进度趋势 -->
      <el-col :xs="24" :lg="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><TrendCharts /></el-icon>
              <span>进度趋势</span>
            </div>
          </template>
          <div class="progress-trends-chart">
            <div v-loading="loading" class="chart-placeholder">
              <el-icon size="48" color="#ddd"><TrendCharts /></el-icon>
              <p>趋势图表开发中...</p>
              <div class="trend-data">
                <div v-for="trend in reportData.progressTrends" :key="trend.date" class="trend-item">
                  <span>{{ formatDate(trend.date) }}</span>
                  <span>完成率: {{ trend.completion }}%</span>
                  <span>参与度: {{ trend.participation }}%</span>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 优秀学生排行 -->
      <el-col :xs="24" :lg="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><Trophy /></el-icon>
              <span>优秀学生排行</span>
            </div>
          </template>
          <div class="top-performers">
            <div v-loading="loading" class="performers-list">
              <div v-for="(student, index) in reportData.topPerformers" :key="student.id" class="performer-item">
                <div class="performer-rank">
                  <span class="rank-number" :class="{ 'top-three': index < 3 }">{{ index + 1 }}</span>
                </div>
                <div class="performer-info">
                  <div class="performer-name">{{ student.name }}</div>
                  <div class="performer-stats">
                    <span>完成率: {{ student.completionRate }}%</span>
                    <span>积分: {{ student.points }}</span>
                  </div>
                </div>
                <div class="performer-grade">
                  <el-tag :type="student.completionRate >= 90 ? 'success' : 'warning'">
                    {{ getGradeLevel(student.completionRate) }}
                  </el-tag>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 统计摘要 -->
    <el-row class="summary-section">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <el-icon><DataBoard /></el-icon>
              <span>统计摘要</span>
            </div>
          </template>
          <div class="summary-content">
            <div class="summary-stats">
              <div class="stat-item">
                <div class="stat-label">总体活跃度</div>
                <div class="stat-value">{{ totalActivityRate }}%</div>
                <div class="stat-description">学生参与学习活动的比例</div>
              </div>
              <div class="stat-item">
                <div class="stat-label">平均作业完成率</div>
                <div class="stat-value">{{ averageAssignmentCompletion }}%</div>
                <div class="stat-description">所有作业的平均完成情况</div>
              </div>
              <div class="stat-item">
                <div class="stat-label">班级平均表现</div>
                <div class="stat-value">{{ getGradeLevel(reportData.overview.averageCompletion) }}</div>
                <div class="stat-description">基于完成率的综合评级</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.education-reports-view {
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

.filter-bar {
  margin-bottom: 24px;
}

.filter-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.filters {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-item label {
  font-weight: 500;
  color: #303133;
}

.actions {
  display: flex;
  gap: 12px;
}

.overview-section {
  margin-bottom: 24px;
}

.overview-card {
  height: 120px;
}

.overview-content {
  display: flex;
  align-items: center;
  gap: 16px;
  height: 100%;
}

.overview-icon {
  width: 64px;
  height: 64px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.overview-icon.students {
  background: linear-gradient(135deg, #409EFF, #67C23A);
}

.overview-icon.assignments {
  background: linear-gradient(135deg, #67C23A, #E6A23C);
}

.overview-icon.projects {
  background: linear-gradient(135deg, #E6A23C, #F56C6C);
}

.overview-icon.completion {
  background: linear-gradient(135deg, #F56C6C, #9C27B0);
}

.overview-info {
  flex: 1;
}

.overview-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 4px;
}

.overview-label {
  color: #909399;
  font-size: 14px;
}

.main-reports {
  margin-bottom: 24px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.class-activity-chart {
  height: 300px;
  overflow-y: auto;
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.activity-item {
  padding: 12px 0;
}

.activity-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.class-name {
  font-weight: 500;
  color: #303133;
}

.activity-stats {
  display: flex;
  gap: 12px;
  font-size: 14px;
  color: #606266;
}

.activity-progress {
  margin-top: 8px;
}

.assignment-stats-chart {
  height: 300px;
  overflow-y: auto;
}

.stats-table {
  height: 100%;
}

.secondary-reports {
  margin-bottom: 24px;
}

.progress-trends-chart {
  height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-placeholder {
  text-align: center;
  color: #909399;
}

.trend-data {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.trend-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 16px;
  background: #f8f9fa;
  border-radius: 4px;
  font-size: 14px;
}

.top-performers {
  height: 300px;
  overflow-y: auto;
}

.performers-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.performer-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  transition: all 0.3s;
}

.performer-item:hover {
  background: #e8f4fd;
}

.performer-rank {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.rank-number {
  font-weight: bold;
  color: #606266;
}

.rank-number.top-three {
  color: #F56C6C;
}

.performer-info {
  flex: 1;
}

.performer-name {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.performer-stats {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: #909399;
}

.performer-grade {
  margin-left: auto;
}

.summary-section {
  margin-bottom: 24px;
}

.summary-content {
  padding: 16px 0;
}

.summary-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 24px;
}

.stat-item {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 8px;
}

.stat-description {
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
}

@media (max-width: 768px) {
  .education-reports-view {
    padding: 16px;
  }
  
  .page-title {
    font-size: 24px;
  }
  
  .filter-content {
    flex-direction: column;
    align-items: stretch;
  }
  
  .filters {
    justify-content: center;
  }
  
  .actions {
    justify-content: center;
  }
  
  .overview-card {
    height: auto;
  }
  
  .overview-content {
    flex-direction: column;
    text-align: center;
    gap: 8px;
  }
  
  .overview-icon {
    width: 48px;
    height: 48px;
  }
  
  .summary-stats {
    grid-template-columns: 1fr;
  }
}
</style> 