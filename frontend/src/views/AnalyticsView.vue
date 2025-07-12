<template>
  <div class="analytics-view">
    <div class="page-header">
      <h1>统计分析</h1>
      <p>学习数据统计和分析报告</p>
    </div>

    <!-- 统计概览 -->
    <div class="analytics-overview">
      <el-row :gutter="20">
        <el-col :xs="12" :sm="6" :lg="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ overviewStats.totalProjects }}</div>
              <div class="stat-label">总课题数</div>
            </div>
            <el-icon class="stat-icon">
              <FolderOpened />
            </el-icon>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :lg="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ overviewStats.totalAssignments }}</div>
              <div class="stat-label">总作业数</div>
            </div>
            <el-icon class="stat-icon">
              <Notebook />
            </el-icon>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :lg="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ overviewStats.totalStudents }}</div>
              <div class="stat-label">学生总数</div>
            </div>
            <el-icon class="stat-icon">
              <User />
            </el-icon>
          </el-card>
        </el-col>
        <el-col :xs="12" :sm="6" :lg="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <div class="stat-number">{{ overviewStats.completionRate }}%</div>
              <div class="stat-label">完成率</div>
            </div>
            <el-icon class="stat-icon">
              <TrendCharts />
            </el-icon>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 图表区域 -->
    <div class="analytics-charts">
      <el-row :gutter="20">
        <!-- 作业提交趋势 -->
        <el-col :xs="24" :lg="12">
          <el-card class="chart-card">
            <template #header>
              <div class="chart-header">
                <span>作业提交趋势</span>
                <el-date-picker
                  v-model="dateRange"
                  type="daterange"
                  range-separator="至"
                  start-placeholder="开始日期"
                  end-placeholder="结束日期"
                  size="small"
                  @change="updateCharts"
                />
              </div>
            </template>
            <div class="chart-container">
              <SimpleChart
                :chart-data="submissionTrendData"
                chart-type="line"
                :height="300"
              />
            </div>
          </el-card>
        </el-col>

        <!-- 课题参与分布 -->
        <el-col :xs="24" :lg="12">
          <el-card class="chart-card">
            <template #header>
              <span>课题参与分布</span>
            </template>
            <div class="chart-container">
              <SimpleChart
                :chart-data="projectDistributionData"
                chart-type="doughnut"
                :height="300"
              />
            </div>
          </el-card>
        </el-col>

        <!-- 学生成绩分布 -->
        <el-col :xs="24" :lg="12">
          <el-card class="chart-card">
            <template #header>
              <span>学生成绩分布</span>
            </template>
            <div class="chart-container">
              <SimpleChart
                :chart-data="gradeDistributionData"
                chart-type="bar"
                :height="300"
              />
            </div>
          </el-card>
        </el-col>

        <!-- 活跃度统计 -->
        <el-col :xs="24" :lg="12">
          <el-card class="chart-card">
            <template #header>
              <span>活跃度统计</span>
            </template>
            <div class="chart-container">
              <SimpleChart
                :chart-data="activityData"
                chart-type="radar"
                :height="300"
              />
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 详细报表 -->
    <div class="analytics-tables">
      <el-tabs v-model="activeTab" type="card">
        <!-- 课题统计 -->
        <el-tab-pane label="课题统计" name="projects">
          <el-table
            :data="projectStats"
            style="width: 100%"
            v-loading="loading"
            stripe
          >
            <el-table-column prop="name" label="课题名称" min-width="200" />
            <el-table-column prop="student_count" label="学生数" width="100" align="center" />
            <el-table-column prop="assignment_count" label="作业数" width="100" align="center" />
            <el-table-column prop="completion_rate" label="完成率" width="100" align="center">
              <template #default="{ row }">
                <el-progress
                  :percentage="row.completion_rate"
                  :color="getProgressColor(row.completion_rate)"
                  :show-text="true"
                  :stroke-width="8"
                />
              </template>
            </el-table-column>
            <el-table-column prop="average_score" label="平均分" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getScoreType(row.average_score)">
                  {{ row.average_score || '-' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="last_activity" label="最后活动" width="150" align="center" />
          </el-table>
        </el-tab-pane>

        <!-- 学生统计 -->
        <el-tab-pane label="学生统计" name="students">
          <el-table
            :data="studentStats"
            style="width: 100%"
            v-loading="loading"
            stripe
          >
            <el-table-column prop="name" label="学生姓名" min-width="120" />
            <el-table-column prop="class_name" label="班级" width="120" />
            <el-table-column prop="project_count" label="参与课题" width="100" align="center" />
            <el-table-column prop="assignment_count" label="作业总数" width="100" align="center" />
            <el-table-column prop="completed_count" label="已完成" width="100" align="center" />
            <el-table-column prop="completion_rate" label="完成率" width="100" align="center">
              <template #default="{ row }">
                <el-progress
                  :percentage="row.completion_rate"
                  :color="getProgressColor(row.completion_rate)"
                  :show-text="true"
                  :stroke-width="8"
                />
              </template>
            </el-table-column>
            <el-table-column prop="average_score" label="平均分" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getScoreType(row.average_score)">
                  {{ row.average_score || '-' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 作业统计 -->
        <el-tab-pane label="作业统计" name="assignments">
          <el-table
            :data="assignmentStats"
            style="width: 100%"
            v-loading="loading"
            stripe
          >
            <el-table-column prop="title" label="作业标题" min-width="200" />
            <el-table-column prop="project_name" label="所属课题" width="150" />
            <el-table-column prop="due_date" label="截止时间" width="150" align="center" />
            <el-table-column prop="submission_count" label="提交数" width="100" align="center" />
            <el-table-column prop="total_students" label="总学生数" width="100" align="center" />
            <el-table-column prop="completion_rate" label="完成率" width="100" align="center">
              <template #default="{ row }">
                <el-progress
                  :percentage="row.completion_rate"
                  :color="getProgressColor(row.completion_rate)"
                  :show-text="true"
                  :stroke-width="8"
                />
              </template>
            </el-table-column>
            <el-table-column prop="average_score" label="平均分" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="getScoreType(row.average_score)">
                  {{ row.average_score || '-' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  FolderOpened,
  Notebook,
  User,
  TrendCharts
} from '@element-plus/icons-vue'
import { ApiService } from '../services/api'
import { useAuthStore } from '../stores/auth'
import SimpleChart from '../components/SimpleChart.vue'

const authStore = useAuthStore()

// 响应式数据
const loading = ref(false)
const activeTab = ref('projects')
const dateRange = ref<[Date, Date]>([
  new Date(Date.now() - 30 * 24 * 60 * 60 * 1000), // 30天前
  new Date()
])

const overviewStats = ref({
  totalProjects: 0,
  totalAssignments: 0,
  totalStudents: 0,
  completionRate: 0
})

const projectStats = ref([])
const studentStats = ref([])
const assignmentStats = ref([])

// 图表数据
const submissionTrendData = ref({
  labels: [],
  datasets: [{
    label: '作业提交数',
    data: [],
    borderColor: '#409EFF',
    backgroundColor: '#409EFF20',
    tension: 0.4
  }]
})

const projectDistributionData = ref({
  labels: [],
  datasets: [{
    data: [],
    backgroundColor: [
      '#409EFF',
      '#67C23A',
      '#E6A23C',
      '#F56C6C',
      '#909399'
    ]
  }]
})

const gradeDistributionData = ref({
  labels: ['90-100', '80-89', '70-79', '60-69', '60以下'],
  datasets: [{
    label: '学生数',
    data: [],
    backgroundColor: '#409EFF'
  }]
})

const activityData = ref({
  labels: ['作业提交', '文档编辑', '讨论参与', '代码提交', '项目活跃'],
  datasets: [{
    label: '活跃度',
    data: [],
    backgroundColor: '#409EFF20',
    borderColor: '#409EFF'
  }]
})

// 计算属性
const userRole = computed(() => authStore.userRole)

// 生命周期
onMounted(() => {
  loadAnalyticsData()
})

// 方法
const loadAnalyticsData = async () => {
  try {
    loading.value = true
    await Promise.all([
      loadOverviewStats(),
      loadProjectStats(),
      loadStudentStats(),
      loadAssignmentStats(),
      loadChartData()
    ])
  } catch (error) {
    console.error('加载统计数据失败:', error)
    ElMessage.error('加载统计数据失败')
  } finally {
    loading.value = false
  }
}

const loadOverviewStats = async () => {
  try {
    const response = await ApiService.getAnalyticsOverview()
    overviewStats.value = response.data || {}
  } catch (error) {
    console.error('加载概览统计失败:', error)
  }
}

const loadProjectStats = async () => {
  try {
    const response = await ApiService.getProjectStats()
    projectStats.value = response.data || []
  } catch (error) {
    console.error('加载课题统计失败:', error)
  }
}

const loadStudentStats = async () => {
  try {
    const response = await ApiService.getStudentStats()
    studentStats.value = response.data || []
  } catch (error) {
    console.error('加载学生统计失败:', error)
  }
}

const loadAssignmentStats = async () => {
  try {
    const response = await ApiService.getAssignmentStats()
    assignmentStats.value = response.data || []
  } catch (error) {
    console.error('加载作业统计失败:', error)
  }
}

const loadChartData = async () => {
  try {
    // 加载提交趋势数据
    const trendResponse = await ApiService.getSubmissionTrend({
      start_date: dateRange.value[0],
      end_date: dateRange.value[1]
    })
    submissionTrendData.value = trendResponse.data || submissionTrendData.value

    // 加载课题分布数据
    const distributionResponse = await ApiService.getProjectDistribution()
    projectDistributionData.value = distributionResponse.data || projectDistributionData.value

    // 加载成绩分布数据
    const gradeResponse = await ApiService.getGradeDistribution()
    gradeDistributionData.value = gradeResponse.data || gradeDistributionData.value

    // 加载活跃度数据
    const activityResponse = await ApiService.getActivityStats()
    activityData.value = activityResponse.data || activityData.value
  } catch (error) {
    console.error('加载图表数据失败:', error)
  }
}

const updateCharts = () => {
  loadChartData()
}

const getProgressColor = (percentage: number) => {
  if (percentage >= 80) return '#67C23A'
  if (percentage >= 60) return '#E6A23C'
  return '#F56C6C'
}

const getScoreType = (score: number) => {
  if (score >= 90) return 'success'
  if (score >= 80) return 'warning'
  if (score >= 60) return 'info'
  return 'danger'
}
</script>

<style scoped>
.analytics-view {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.page-header p {
  margin: 0;
  color: #606266;
  font-size: 14px;
}

.analytics-overview {
  margin-bottom: 20px;
}

.stat-card {
  height: 100px;
  position: relative;
  overflow: hidden;
}

.stat-card :deep(.el-card__body) {
  padding: 20px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.stat-content {
  flex: 1;
}

.stat-number {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: #606266;
}

.stat-icon {
  font-size: 40px;
  color: #409EFF;
  opacity: 0.3;
}

.analytics-charts {
  margin-bottom: 20px;
}

.chart-card {
  height: 400px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-container {
  height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.analytics-tables {
  margin-top: 20px;
}

.analytics-tables :deep(.el-tabs__content) {
  padding-top: 20px;
}

.analytics-tables :deep(.el-table) {
  margin-top: 0;
}
</style> 