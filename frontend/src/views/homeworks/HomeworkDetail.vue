<template>
  <div class="homework-detail" v-loading="loading">
    <div v-if="homework" class="homework-container">
      <!-- 面包屑导航 -->
      <el-breadcrumb class="breadcrumb" separator=">">
        <el-breadcrumb-item :to="{ path: '/scenes' }">课题</el-breadcrumb-item>
        <el-breadcrumb-item :to="{ path: `/scenes/${homework.project_id}` }">
          {{ homework.project?.name }}
        </el-breadcrumb-item>
        <el-breadcrumb-item>{{ homework.title }}</el-breadcrumb-item>
      </el-breadcrumb>

      <!-- 作业头部信息 -->
      <div class="homework-header card">
        <div class="header-content">
          <div class="title-section">
            <h1 class="homework-title">{{ homework.title }}</h1>
            <div class="homework-meta">
              <span>创建时间: {{ formatDate(homework.created_at) }}</span>
              <span>创建者: {{ homework.author?.name }}</span>
              <span>截止日期: {{ homework.deadline ? formatDate(homework.deadline) : '无限制' }}</span>
              <el-tag :type="getStatusTagType(homework.status)" size="small">
                {{ getHomeworkStatusText(homework.status) }}
              </el-tag>
            </div>
          </div>
          <div class="header-actions" v-if="canManage">
            <el-button @click="editHomework">编辑作业</el-button>
            <el-button @click="showGradeDialog" v-if="canGrade">批改作业</el-button>
          </div>
        </div>
        <div class="homework-description">
          <h3>作业要求</h3>
          <p>{{ homework.description }}</p>
        </div>
        <div class="homework-stats">
          <div class="stat-item">
            <span class="stat-label">满分</span>
            <span class="stat-value">{{ homework.max_grade }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">已提交</span>
            <span class="stat-value">{{ submittedCount }}/{{ totalStudents }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">平均分</span>
            <span class="stat-value">{{ averageGrade.toFixed(1) }}</span>
          </div>
        </div>
      </div>

      <!-- 标签页容器 -->
      <div class="tab-container card">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <!-- 我的提交标签页 (学生视图) -->
          <el-tab-pane v-if="isStudent" label="我的提交" name="my-submission">
            <div class="my-submission-content">
              <div v-if="mySubmission" class="submission-info">
                <div class="submission-header">
                  <h3>提交状态</h3>
                  <el-tag :type="getSubmissionStatusTagType(mySubmission.status)">
                    {{ getSubmissionStatusText(mySubmission.status) }}
                  </el-tag>
                </div>
                
                <div class="submission-details">
                  <div class="detail-item">
                    <span class="label">提交时间:</span>
                    <span class="value">{{ mySubmission.submitted_at ? formatDate(mySubmission.submitted_at) : '-' }}</span>
                  </div>
                  <div class="detail-item" v-if="mySubmission.grade !== null">
                    <span class="label">得分:</span>
                    <span class="value grade">{{ mySubmission.grade }}/{{ homework.max_grade }}</span>
                  </div>
                  <div class="detail-item" v-if="mySubmission.feedback">
                    <span class="label">反馈:</span>
                    <div class="feedback-content">{{ mySubmission.feedback }}</div>
                  </div>
                </div>

                <div class="submission-content">
                  <h4>提交内容</h4>
                  <div class="content-text">{{ mySubmission.content }}</div>
                </div>

                <div class="submission-files" v-if="mySubmission.files?.length">
                  <h4>提交文件</h4>
                  <div class="file-list">
                    <div 
                      v-for="file in mySubmission.files" 
                      :key="file.id"
                      class="file-item"
                    >
                      <el-icon><Document /></el-icon>
                      <span class="file-name">{{ file.title }}</span>
                      <el-button size="small" @click="downloadFile(file)">下载</el-button>
                    </div>
                  </div>
                </div>

                <div class="submission-actions">
                  <el-button 
                    v-if="mySubmission" 
                    type="primary" 
                    @click="openHomeworkEditor"
                    :loading="submitting"
                  >
                    查看/编辑代码
                  </el-button>
                </div>
              </div>

              <div v-else class="no-submission">
                <el-empty description="尚未提交作业">
                  <el-button 
                    type="primary" 
                    @click="openHomeworkEditor"
                    v-if="canSubmit"
                    :loading="submitting"
                  >
                    打开作业仓库
                  </el-button>
                </el-empty>
              </div>
            </div>
          </el-tab-pane>

          <!-- 所有提交标签页 (教师视图) -->
          <el-tab-pane v-if="canGrade" label="所有提交" name="all-submissions">
            <div class="submissions-content">
              <div class="submissions-header">
                <h3>学生提交列表</h3>
                <div class="submissions-actions">
                  <el-select v-model="submissionFilter" placeholder="筛选状态" @change="fetchSubmissions">
                    <el-option label="全部" value="" />
                    <el-option label="已提交" value="submitted" />
                    <el-option label="已批改" value="graded" />
                    <el-option label="未提交" value="pending" />
                  </el-select>
                  <el-button @click="exportGrades" v-if="submissions.length">
                    导出成绩
                  </el-button>
                </div>
              </div>

              <div class="submissions-list" v-loading="submissionsLoading">
                <el-table :data="submissions" style="width: 100%">
                  <el-table-column prop="student.name" label="学生姓名" width="120" />
                  <el-table-column prop="student.username" label="用户名" width="120" />
                  <el-table-column label="提交状态" width="100">
                    <template #default="scope">
                      <el-tag :type="getSubmissionStatusTagType(scope.row.status)">
                        {{ getSubmissionStatusText(scope.row.status) }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="submitted_at" label="提交时间" width="150">
                    <template #default="scope">
                      {{ scope.row.submitted_at ? formatDate(scope.row.submitted_at) : '-' }}
                    </template>
                  </el-table-column>
                  <el-table-column label="得分" width="100">
                    <template #default="scope">
                      <span v-if="scope.row.grade !== null" class="grade">
                        {{ scope.row.grade }}/{{ homework.max_grade }}
                      </span>
                      <span v-else>-</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="200">
                    <template #default="scope">
                      <el-button 
                        size="small" 
                        @click="viewSubmission(scope.row)"
                        v-if="scope.row.status !== 'pending'"
                      >
                        查看
                      </el-button>
                      <el-button 
                        size="small" 
                        type="primary" 
                        @click="gradeSubmission(scope.row)"
                        v-if="scope.row.status === 'submitted' || scope.row.status === 'graded'"
                      >
                        {{ scope.row.status === 'graded' ? '修改评分' : '批改' }}
                      </el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </div>
          </el-tab-pane>

          <!-- 统计分析标签页 -->
          <el-tab-pane v-if="canGrade" label="统计分析" name="statistics">
            <div class="statistics-content">
              <div class="stats-overview">
                <div class="stat-card">
                  <h4>提交率</h4>
                  <div class="stat-number">{{ submissionRate.toFixed(1) }}%</div>
                </div>
                <div class="stat-card">
                  <h4>平均分</h4>
                  <div class="stat-number">{{ averageGrade.toFixed(1) }}</div>
                </div>
                <div class="stat-card">
                  <h4>最高分</h4>
                  <div class="stat-number">{{ maxGrade }}</div>
                </div>
                <div class="stat-card">
                  <h4>最低分</h4>
                  <div class="stat-number">{{ minGrade }}</div>
                </div>
              </div>

              <div class="grade-distribution">
                <h4>成绩分布</h4>
                <div class="distribution-chart">
                  <div 
                    v-for="range in gradeDistribution" 
                    :key="range.range"
                    class="distribution-item"
                  >
                    <span class="range-label">{{ range.range }}</span>
                    <div class="range-bar">
                      <div 
                        class="range-fill" 
                        :style="{ width: `${range.percentage}%` }"
                      ></div>
                    </div>
                    <span class="range-count">{{ range.count }}人</span>
                  </div>
                </div>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>


    <!-- 批改作业对话框 -->
    <el-dialog v-model="gradeDialogVisible" title="批改作业" width="600px">
      <div v-if="currentSubmission" class="grade-content">
        <div class="student-info">
          <h4>学生: {{ currentSubmission.student?.name }}</h4>
          <p>提交时间: {{ currentSubmission.submitted_at ? formatDate(currentSubmission.submitted_at) : '-' }}</p>
        </div>
        
        <div class="submission-content">
          <h4>提交内容</h4>
          <div class="content-text">{{ currentSubmission.content }}</div>
        </div>

        <el-form :model="gradeForm" label-width="80px">
          <el-form-item label="评分">
            <el-input-number 
              v-model="gradeForm.grade" 
              :min="0" 
              :max="homework?.max_grade || 100"
              style="width: 200px"
            />
            <span class="grade-suffix">/ {{ homework?.max_grade || 100 }}</span>
          </el-form-item>
          <el-form-item label="反馈">
            <el-input 
              v-model="gradeForm.feedback" 
              type="textarea" 
              :rows="4"
              placeholder="请输入评价和建议"
            />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="gradeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveGrade" :loading="grading">
          保存评分
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { homeworkService } from '@/services/api'
import type { Homework, HomeworkSubmission } from '@/types'
import { ElMessage } from 'element-plus'
import { Document } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const homework = ref<Homework | null>(null)
const mySubmission = ref<HomeworkSubmission | null>(null)
const submissions = ref<HomeworkSubmission[]>([])
const currentSubmission = ref<HomeworkSubmission | null>(null)

const loading = ref(false)
const submissionsLoading = ref(false)
const submitting = ref(false)
const grading = ref(false)

const activeTab = ref('my-submission')
const submissionFilter = ref('')

// 对话框状态
const gradeDialogVisible = ref(false)

// 表单数据
const gradeForm = ref({
  grade: 0,
  feedback: ''
})

// 计算属性
const homeworkId = computed(() => route.params.id as string)

// 权限相关状态
const canManage = ref(false)
const canGrade = ref(false)
const canSubmit = ref(false)

// 用户角色
const isStudent = ref(false)
const isTeacher = ref(false)

const canResubmit = computed(() => {
  return canSubmit.value && mySubmission.value?.status !== 'graded'
})

const submittedCount = computed(() => {
  return submissions.value.filter(s => s.status !== 'pending').length
})

const totalStudents = computed(() => {
  return submissions.value.length
})

const submissionRate = computed(() => {
  return totalStudents.value > 0 ? (submittedCount.value / totalStudents.value) * 100 : 0
})

const averageGrade = computed(() => {
  const gradedSubmissions = submissions.value.filter(s => s.grade !== null)
  if (gradedSubmissions.length === 0) return 0
  const total = gradedSubmissions.reduce((sum, s) => sum + (s.grade || 0), 0)
  return total / gradedSubmissions.length
})

const maxGrade = computed(() => {
  const grades = submissions.value.map(s => s.grade).filter(g => g !== null) as number[]
  return grades.length > 0 ? Math.max(...grades) : 0
})

const minGrade = computed(() => {
  const grades = submissions.value.map(s => s.grade).filter(g => g !== null) as number[]
  return grades.length > 0 ? Math.min(...grades) : 0
})

const gradeDistribution = computed(() => {
  const ranges = [
    { range: '90-100', min: 90, max: 100 },
    { range: '80-89', min: 80, max: 89 },
    { range: '70-79', min: 70, max: 79 },
    { range: '60-69', min: 60, max: 69 },
    { range: '0-59', min: 0, max: 59 }
  ]

  const gradedSubmissions = submissions.value.filter(s => s.grade !== null)
  const total = gradedSubmissions.length

  return ranges.map(range => {
    const count = gradedSubmissions.filter(s => 
      s.grade! >= range.min && s.grade! <= range.max
    ).length
    const percentage = total > 0 ? (count / total) * 100 : 0
    
    return {
      range: range.range,
      count,
      percentage
    }
  })
})

// 方法
const fetchHomework = async () => {
  loading.value = true
  try {
    // 响应拦截器已经返回了 response.data，所以这里直接就是后端返回的数据
    const data = await homeworkService.getHomework(homeworkId.value) as any
    
    console.log('从后端获取的原始数据:', data)
    
    // 后端返回格式: { homework: {...}, permissions: {...} }
    if (data && data.homework) {
      homework.value = data.homework
      
      // 从后端返回的权限信息中获取权限
      const permissions = data.permissions || {}
      canManage.value = permissions.can_edit || false
      canGrade.value = permissions.can_grade || false
      canSubmit.value = permissions.can_submit || false
      
      // 设置用户角色
      // 只有管理员或有管理权限的才是纯教师角色
      isTeacher.value = userStore.isAdmin || canManage.value
      // 可以提交作业的就显示学生视图（包括研究员）
      isStudent.value = canSubmit.value && userStore.isLoggedIn
      
      console.log('权限信息:', {
        canManage: canManage.value,
        canGrade: canGrade.value,
        canSubmit: canSubmit.value,
        isTeacher: isTeacher.value,
        isStudent: isStudent.value,
        isLoggedIn: userStore.isLoggedIn,
        permissions: permissions
      })
    } else {
      // 兼容旧格式（直接返回作业对象）
      console.warn('使用旧格式数据，没有权限信息')
      homework.value = data
      canManage.value = false
      canGrade.value = false
      canSubmit.value = false
      isStudent.value = false
      isTeacher.value = false
    }
    
    // 设置默认标签页
    if (canSubmit.value) {
      activeTab.value = 'my-submission'
      fetchMySubmission()
    } else if (canGrade.value) {
      activeTab.value = 'all-submissions'
      fetchSubmissions()
    }
  } catch (error) {
    console.error('获取作业详情失败:', error)
    ElMessage.error('获取作业详情失败')
  } finally {
    loading.value = false
  }
}

const fetchMySubmission = async () => {
  if (!isStudent.value) return
  
  try {
    const response = await homeworkService.getMySubmission(homeworkId.value)
    mySubmission.value = response.data || response
  } catch (error) {
    console.error('获取我的提交失败:', error)
  }
}

const fetchSubmissions = async () => {
  if (!canGrade.value) return
  
  submissionsLoading.value = true
  try {
    const response = await homeworkService.getSubmissions(homeworkId.value)
    const data = response.data || response
    submissions.value = Array.isArray(data) ? data : (data.submissions || [])
  } catch (error) {
    console.error('获取提交列表失败:', error)
  } finally {
    submissionsLoading.value = false
  }
}

const handleTabChange = (tabName: string) => {
  activeTab.value = tabName
  switch (tabName) {
    case 'my-submission':
      fetchMySubmission()
      break
    case 'all-submissions':
      fetchSubmissions()
      break
  }
}

const openHomeworkEditor = async () => {
  submitting.value = true
  try {
    const response = await homeworkService.submitHomework({
      homework_id: homeworkId.value,
      content: '',
      files: []
    }) as any
    
    if (response.web_ide_url) {
      ElMessage.success('正在打开作业仓库...')
      // 在新窗口打开GitLab仓库
      window.open(response.web_ide_url, '_blank')
    } else {
      ElMessage.error('无法获取仓库链接')
    }
  } catch (error: any) {
    console.error('打开仓库失败:', error)
    ElMessage.error(error.response?.data?.error || '打开仓库失败')
  } finally {
    submitting.value = false
  }
}

const viewSubmission = async (submission: HomeworkSubmission) => {
  if (!submission || submission.status === 'pending') {
    ElMessage.warning('学生尚未提交作业')
    return
  }
  
  try {
    const response = await homeworkService.getSubmissionViewURL(submission.id)
    const data = response.data || response
    
    if (data.view_url) {
      // 在新窗口打开GitLab仓库
      window.open(data.view_url, '_blank')
      ElMessage.success(`正在打开 ${data.student_name || '学生'} 的作业仓库`)
    } else {
      ElMessage.error('无法获取仓库链接')
    }
  } catch (error) {
    console.error('获取仓库链接失败:', error)
    ElMessage.error('获取仓库链接失败')
  }
}

const gradeSubmission = (submission: HomeworkSubmission) => {
  currentSubmission.value = submission
  gradeForm.value = {
    grade: submission.grade || 0,
    feedback: submission.feedback || ''
  }
  gradeDialogVisible.value = true
}

const saveGrade = async () => {
  if (!currentSubmission.value) return
  
  grading.value = true
  try {
    await homeworkService.gradeSubmission(
      currentSubmission.value.id,
      gradeForm.value.grade,
      gradeForm.value.feedback
    )
    
    ElMessage.success('评分保存成功')
    gradeDialogVisible.value = false
    fetchSubmissions()
  } catch (error) {
    console.error('保存评分失败:', error)
    ElMessage.error('保存评分失败')
  } finally {
    grading.value = false
  }
}

const editHomework = () => {
  // 跳转到作业编辑页面（数据库存储模式）
  if (homework.value?.id) {
    router.push(`/homeworks/${homework.value.id}/edit`)
  } else {
    ElMessage.error('无法获取作业信息')
  }
}

const showGradeDialog = () => {
  // 显示批改对话框，但这个功能应该针对具体的提交
  // 跳转到所有提交列表页面
  router.push(`/homeworks/${homeworkId.value}/submissions`)
}

const exportGrades = () => {
  // 导出成绩的逻辑
  ElMessage.info('导出功能开发中')
}

const downloadFile = (file: any) => {
  // 下载文件的逻辑
  console.log('下载文件:', file)
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getHomeworkStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    draft: '草稿',
    published: '已发布',
    closed: '已关闭'
  }
  return statusMap[status] || status
}

const getStatusTagType = (status: string) => {
  const typeMap: Record<string, string> = {
    draft: 'info',
    published: 'success',
    closed: 'danger'
  }
  return typeMap[status] || 'info'
}

const getSubmissionStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    pending: '未提交',
    submitted: '已提交',
    graded: '已批改'
  }
  return statusMap[status] || status
}

const getSubmissionStatusTagType = (status: string) => {
  const typeMap: Record<string, string> = {
    pending: 'info',
    submitted: 'warning',
    graded: 'success'
  }
  return typeMap[status] || 'info'
}

// 生命周期
onMounted(() => {
  fetchHomework()
})
</script>

<style scoped>
.homework-detail {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.breadcrumb {
  margin-bottom: 20px;
}

.homework-header {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.title-section h1 {
  margin: 0 0 8px 0;
  color: var(--primary-color);
}

.homework-meta {
  display: flex;
  gap: 16px;
  align-items: center;
  font-size: 14px;
  color: var(--light-text);
}

.homework-description {
  margin-bottom: 20px;
}

.homework-description h3 {
  margin: 0 0 8px 0;
  color: var(--text-color);
}

.homework-description p {
  margin: 0;
  color: var(--light-text);
  line-height: 1.6;
}

.homework-stats {
  display: flex;
  gap: 40px;
  padding: 16px 0;
  border-top: 1px solid var(--border-color);
}

.stat-item {
  text-align: center;
}

.stat-label {
  display: block;
  font-size: 12px;
  color: var(--light-text);
  margin-bottom: 4px;
}

.stat-value {
  display: block;
  font-size: 20px;
  font-weight: bold;
  color: var(--primary-color);
}

.tab-container {
  min-height: 500px;
}

/* 我的提交样式 */
.submission-info {
  max-width: 800px;
}

.submission-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.submission-details {
  margin-bottom: 20px;
}

.detail-item {
  display: flex;
  margin-bottom: 12px;
}

.detail-item .label {
  width: 100px;
  font-weight: 500;
  color: var(--text-color);
}

.detail-item .value {
  color: var(--light-text);
}

.detail-item .grade {
  font-weight: bold;
  color: var(--primary-color);
}

.feedback-content {
  background-color: var(--background-light);
  padding: 12px;
  border-radius: 4px;
  color: var(--text-color);
  line-height: 1.5;
}

.submission-content {
  margin-bottom: 20px;
}

.submission-content h4 {
  margin: 0 0 12px 0;
  color: var(--text-color);
}

.content-text {
  background-color: var(--background-light);
  padding: 16px;
  border-radius: 8px;
  color: var(--text-color);
  line-height: 1.6;
  white-space: pre-wrap;
}

.submission-files {
  margin-bottom: 20px;
}

.submission-files h4 {
  margin: 0 0 12px 0;
  color: var(--text-color);
}

.file-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: var(--background-light);
  border-radius: 4px;
}

.file-name {
  flex: 1;
  color: var(--text-color);
}

.no-submission {
  text-align: center;
  padding: 60px 0;
}

/* 提交列表样式 */
.submissions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.submissions-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.grade {
  font-weight: bold;
  color: var(--primary-color);
}

/* 统计分析样式 */
.stats-overview {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  background-color: var(--card-background);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 20px;
  text-align: center;
}

.stat-card h4 {
  margin: 0 0 8px 0;
  color: var(--text-color);
  font-size: 14px;
}

.stat-number {
  font-size: 24px;
  font-weight: bold;
  color: var(--primary-color);
}

.grade-distribution h4 {
  margin: 0 0 20px 0;
  color: var(--text-color);
}

.distribution-chart {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.distribution-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.range-label {
  width: 60px;
  font-size: 12px;
  color: var(--text-color);
}

.range-bar {
  flex: 1;
  height: 20px;
  background-color: var(--background-light);
  border-radius: 10px;
  overflow: hidden;
}

.range-fill {
  height: 100%;
  background-color: var(--primary-color);
  transition: width 0.3s;
}

.range-count {
  width: 40px;
  text-align: right;
  font-size: 12px;
  color: var(--light-text);
}

/* 批改对话框样式 */
.grade-content {
  max-height: 400px;
  overflow-y: auto;
}

.student-info {
  margin-bottom: 20px;
  padding: 12px;
  background-color: var(--background-light);
  border-radius: 8px;
}

.student-info h4 {
  margin: 0 0 4px 0;
  color: var(--text-color);
}

.student-info p {
  margin: 0;
  font-size: 12px;
  color: var(--light-text);
}

.grade-suffix {
  margin-left: 8px;
  color: var(--light-text);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .homework-detail {
    padding: 10px;
  }
  
  .header-content {
    flex-direction: column;
    gap: 16px;
  }
  
  .homework-stats {
    flex-direction: column;
    gap: 16px;
    text-align: left;
  }
  
  .stats-overview {
    grid-template-columns: 1fr;
  }
  
  .submissions-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .submissions-actions {
    justify-content: flex-end;
  }
  
  .distribution-item {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  
  .range-label,
  .range-count {
    width: auto;
    text-align: left;
  }
}
</style>
