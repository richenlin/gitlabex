<template>
  <div class="homework-submissions">
    <div class="page-header">
      <div class="header-content">
        <div class="title-section">
          <h1>{{ homework?.title }} - 学生提交</h1>
          <div class="homework-meta">
            <span>截止日期: {{ homework?.deadline ? formatDate(homework.deadline) : '无限制' }}</span>
            <span>总提交数: {{ submissions.length }}</span>
            <span>已批改数: {{ gradedCount }}</span>
          </div>
        </div>
        <div class="header-actions">
          <el-button @click="goBack">返回作业详情</el-button>
          <el-button type="primary" @click="exportGrades" v-if="submissions.length">导出成绩</el-button>
        </div>
      </div>
    </div>

    <div class="submissions-content">
      <div class="filter-section">
        <div class="filter-controls">
          <el-select v-model="statusFilter" placeholder="筛选状态" @change="applyFilters">
            <el-option label="全部" value="" />
            <el-option label="已提交" value="submitted" />
            <el-option label="已批改" value="graded" />
            <el-option label="未提交" value="pending" />
          </el-select>
          <el-input 
            v-model="searchQuery" 
            placeholder="搜索学生姓名"
            @input="applyFilters"
            style="width: 200px;"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
      </div>

      <el-table 
        :data="filteredSubmissions" 
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="student.name" label="学生姓名" width="120" />
        <el-table-column prop="student.username" label="用户名" width="120" />
        <el-table-column label="提交状态" width="100">
          <template #default="scope">
            <el-tag :type="getSubmissionStatusTagType(scope.row.status)" size="small">
              {{ getSubmissionStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="submitted_at" label="提交时间" width="150">
          <template #default="scope">
            {{ scope.row.submitted_at ? formatDate(scope.row.submitted_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="grade" label="成绩" width="100">
          <template #default="scope">
            <span v-if="scope.row.grade !== null" class="grade-text">
              {{ scope.row.grade }}/{{ homework?.max_grade }}
            </span>
            <span v-else class="no-grade">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="gitlab_branch" label="提交分支" width="180">
          <template #default="scope">
            <el-tag v-if="scope.row.gitlab_branch" type="info" size="small">
              {{ scope.row.gitlab_branch }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <div class="action-buttons">
              <el-button 
                size="small" 
                @click="viewSubmission(scope.row)"
                :disabled="scope.row.status === 'pending'"
              >
                查看
              </el-button>
              <el-button 
                v-if="canGrade"
                size="small" 
                type="primary"
                @click="gradeSubmission(scope.row)"
                :disabled="scope.row.status === 'pending'"
              >
                {{ scope.row.status === 'graded' ? '修改评分' : '批改' }}
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>


    <!-- 批改对话框 -->
    <el-dialog v-model="gradeDialogVisible" title="批改作业" width="600px">
      <div v-if="currentSubmission" class="grade-form">
        <div class="student-info">
          <h4>学生: {{ currentSubmission.student?.name }}</h4>
          <p>提交时间: {{ currentSubmission.submitted_at ? formatDate(currentSubmission.submitted_at) : '-' }}</p>
        </div>
        
        <el-form :model="gradeForm" label-width="80px">
          <el-form-item label="成绩">
            <el-input-number 
              v-model="gradeForm.grade" 
              :min="0" 
              :max="homework?.max_grade || 100"
              :step="1"
              controls-position="right"
            />
            <span class="max-grade-hint">满分: {{ homework?.max_grade || 100 }}</span>
          </el-form-item>
          <el-form-item label="反馈">
            <el-input 
              v-model="gradeForm.feedback" 
              type="textarea" 
              :rows="4"
              placeholder="请输入对学生作业的反馈..."
            />
          </el-form-item>
        </el-form>
      </div>
      
      <template #footer>
        <el-button @click="gradeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitGrade" :loading="grading">
          {{ currentSubmission?.status === 'graded' ? '更新评分' : '提交评分' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { homeworkService } from '@/services/api'
import type { Homework, HomeworkSubmission } from '@/types'
import { ElMessage } from 'element-plus'
import { formatDate } from '@/utils/date'
import { Document, Search } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const homework = ref<Homework | null>(null)
const submissions = ref<HomeworkSubmission[]>([])
const currentSubmission = ref<HomeworkSubmission | null>(null)

// 权限状态
const canManage = ref(false)
const canGrade = ref(false)

const loading = ref(false)
const grading = ref(false)

// 对话框状态
const viewDialogVisible = ref(false)
const gradeDialogVisible = ref(false)

// 筛选和搜索
const statusFilter = ref('')
const searchQuery = ref('')

// 表单数据
const gradeForm = ref({
  grade: 0,
  feedback: ''
})

// 计算属性
const homeworkId = computed(() => route.params.id as string)

const gradedCount = computed(() => {
  return submissions.value.filter(s => s.status === 'graded').length
})

const filteredSubmissions = computed(() => {
  let filtered = submissions.value

  // 状态筛选
  if (statusFilter.value) {
    filtered = filtered.filter(s => s.status === statusFilter.value)
  }

  // 姓名搜索
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(s => 
      s.student?.name.toLowerCase().includes(query) ||
      s.student?.username.toLowerCase().includes(query)
    )
  }

  return filtered
})

// 方法
const fetchHomework = async () => {
  try {
    const data = await homeworkService.getHomework(homeworkId.value) as any
    
    if (data && data.homework) {
      homework.value = data.homework
      
      // 获取权限信息
      const permissions = data.permissions || {}
      canManage.value = permissions.can_edit || false
      canGrade.value = permissions.can_grade || false
      
      console.log('提交列表页权限:', { canManage: canManage.value, canGrade: canGrade.value })
    } else {
      homework.value = data
      // 兼容旧格式，默认无权限
      canManage.value = false
      canGrade.value = false
    }
  } catch (error) {
    console.error('获取作业信息失败:', error)
    ElMessage.error('获取作业信息失败')
  }
}

const fetchSubmissions = async () => {
  loading.value = true
  try {
    const response = await homeworkService.getSubmissions(homeworkId.value)
    const data = response.data || response
    submissions.value = Array.isArray(data) ? data : (data.submissions || [])
  } catch (error) {
    console.error('获取提交列表失败:', error)
    ElMessage.error('获取提交列表失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  // 筛选逻辑在computed中自动处理
}

const viewSubmission = async (submission: HomeworkSubmission) => {
  if (submission.status === 'pending') {
    ElMessage.warning('学生尚未提交作业')
    return
  }
  
  // 直接打开GitLab仓库查看学生代码
  try {
    const response = await homeworkService.getSubmissionViewURL(submission.id)
    const data = response.data || response
    
    if (data.view_url) {
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

const gradeSubmissionFromView = () => {
  viewDialogVisible.value = false
  if (currentSubmission.value) {
    gradeSubmission(currentSubmission.value)
  }
}

const openInWebIDE = async () => {
  if (!currentSubmission.value) return
  
  try {
    const response = await homeworkService.getSubmissionViewURL(currentSubmission.value.id)
    const data = response.data || response
    
    if (data.view_url) {
      window.open(data.view_url, '_blank')
      ElMessage.success('正在打开GitLab在线编辑器')
    } else {
      ElMessage.error('无法获取作业查看链接')
    }
  } catch (error) {
    console.error('获取作业查看链接失败:', error)
    ElMessage.error('获取作业查看链接失败')
  }
}

const submitGrade = async () => {
  if (!currentSubmission.value) return
  
  grading.value = true
  try {
    await homeworkService.gradeHomework(currentSubmission.value.id, {
      grade: gradeForm.value.grade,
      feedback: gradeForm.value.feedback
    })
    
    ElMessage.success('评分提交成功')
    gradeDialogVisible.value = false
    
    // 更新本地数据
    const index = submissions.value.findIndex(s => s.id === currentSubmission.value?.id)
    if (index !== -1) {
      submissions.value[index] = {
        ...submissions.value[index],
        grade: gradeForm.value.grade,
        feedback: gradeForm.value.feedback,
        status: 'graded'
      }
    }
  } catch (error) {
    console.error('评分提交失败:', error)
    ElMessage.error('评分提交失败')
  } finally {
    grading.value = false
  }
}

const viewInGitLab = (submission: HomeworkSubmission) => {
  if (submission.gitlab_branch && homework.value?.project?.gitlab_url) {
    const gitlabUrl = `${homework.value.project.gitlab_url}/-/tree/${submission.gitlab_branch}`
    window.open(gitlabUrl, '_blank')
  } else {
    ElMessage.warning('GitLab链接不可用')
  }
}

const downloadFile = (file: any) => {
  // TODO: 实现文件下载功能
  ElMessage.info('文件下载功能开发中')
}

const exportGrades = () => {
  // TODO: 实现导出成绩功能
  ElMessage.info('导出成绩功能开发中')
}

const goBack = () => {
  router.push(`/homeworks/${homeworkId.value}`)
}

// 状态相关方法
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
onMounted(async () => {
  await fetchHomework()
  await fetchSubmissions()
})
</script>

<style scoped>
.homework-submissions {
  padding: 20px;
}

.page-header {
  margin-bottom: 24px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.title-section h1 {
  margin: 0 0 8px 0;
  color: var(--text-color);
}

.homework-meta {
  display: flex;
  gap: 16px;
  font-size: 14px;
  color: var(--light-text);
}

.header-actions {
  display: flex;
  gap: 12px;
}

.submissions-content {
  background: var(--card-background);
  border-radius: 8px;
  padding: 20px;
}

.filter-section {
  margin-bottom: 20px;
}

.filter-controls {
  display: flex;
  gap: 12px;
  align-items: center;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.grade-text {
  font-weight: 600;
  color: var(--success-color);
}

.no-grade {
  color: var(--light-text);
}

/* 对话框样式 */
.submission-view {
  max-height: 600px;
  overflow-y: auto;
}

.submission-header {
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}

.submission-header h3 {
  margin: 0 0 8px 0;
}

.submission-meta {
  display: flex;
  gap: 16px;
  font-size: 14px;
  color: var(--light-text);
}

.submission-content {
  margin-bottom: 20px;
}

.content-text {
  background: var(--background-light);
  padding: 12px;
  border-radius: 6px;
  border-left: 4px solid var(--primary-color);
  white-space: pre-wrap;
  line-height: 1.6;
}

.files-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background: var(--background-light);
  border-radius: 4px;
}

.file-name {
  flex: 1;
}

.grade-info {
  margin-top: 20px;
  padding: 16px;
  background: var(--success-light, #f0f9ff);
  border-radius: 6px;
  border-left: 4px solid var(--success-color);
}

.grade-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.grade-score {
  font-size: 18px;
  font-weight: 600;
  color: var(--success-color);
}

.grade-feedback p {
  margin: 4px 0 0 0;
  line-height: 1.6;
}

.grade-form .student-info {
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}

.grade-form .student-info h4 {
  margin: 0 0 4px 0;
}

.max-grade-hint {
  margin-left: 8px;
  font-size: 12px;
  color: var(--light-text);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .homework-submissions {
    padding: 10px;
  }
  
  .header-content {
    flex-direction: column;
    gap: 16px;
  }
  
  .homework-meta {
    flex-direction: column;
    gap: 8px;
  }
  
  .filter-controls {
    flex-direction: column;
    align-items: stretch;
  }
  
  .action-buttons {
    flex-direction: column;
  }
}
</style>
