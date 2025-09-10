<template>
  <div class="homework-detail" v-loading="loading">
    <div v-if="homework" class="homework-container">
      <!-- 作业头部信息 -->
      <div class="homework-header">
        <div class="header-left">
          <h1 class="homework-title">{{ homework.title }}</h1>
          <div class="homework-badges">
            <el-tag :type="getStatusType(homework.status)" size="large">
              {{ getStatusText(homework.status) }}
            </el-tag>
            <el-tag v-if="isOverdue" type="danger" size="large">
              已逾期
            </el-tag>
            <el-tag v-if="isDueSoon" type="warning" size="large">
              即将截止
            </el-tag>
          </div>
        </div>
        <div class="header-actions">
          <el-button v-if="canEdit" @click="editHomework">
            <el-icon><Edit /></el-icon>
            编辑
          </el-button>
          <el-button v-if="canSubmit" type="primary" @click="submitHomework">
            <el-icon><Upload /></el-icon>
            提交作业
          </el-button>
          <el-button v-if="canGrade" type="success" @click="gradeHomework">
            <el-icon><Check /></el-icon>
            批改作业
          </el-button>
        </div>
      </div>

      <!-- 作业基本信息 -->
      <el-row :gutter="20" class="homework-info">
        <el-col :span="16">
          <el-card title="作业详情">
            <template #header>
              <span>作业详情</span>
            </template>
            
            <div class="homework-description">
              <h3>作业描述</h3>
              <div class="content" v-html="formatDescription(homework.description)"></div>
            </div>
            
            <div class="homework-requirements" v-if="homework.requirements?.length">
              <h3>作业要求</h3>
              <ul>
                <li v-for="req in homework.requirements" :key="req">{{ req }}</li>
              </ul>
            </div>
            
            <div class="homework-instructions" v-if="homework.instructions">
              <h3>提交说明</h3>
              <div class="content" v-html="formatDescription(homework.instructions)"></div>
            </div>
          </el-card>
        </el-col>
        
        <el-col :span="8">
          <el-card title="作业信息">
            <template #header>
              <span>作业信息</span>
            </template>
            
            <div class="info-item">
              <span class="label">发布者：</span>
              <span class="value">{{ homework.creator?.name }}</span>
            </div>
            <div class="info-item">
              <span class="label">关联课题：</span>
              <el-link 
                :href="`/scenes/${homework.project?.id}`" 
                type="primary"
                :underline="false"
              >
                {{ homework.project?.name }}
              </el-link>
            </div>
            <div class="info-item">
              <span class="label">发布时间：</span>
              <span class="value">{{ formatDate(homework.created_at) }}</span>
            </div>
            <div class="info-item">
              <span class="label">截止时间：</span>
              <span class="value" :class="{ 'overdue': isOverdue }">
                {{ formatDate(homework.due_date) }}
              </span>
            </div>
            <div class="info-item">
              <span class="label">满分：</span>
              <span class="value">{{ homework.max_grade }}分</span>
            </div>
            <div class="info-item" v-if="isTeacher">
              <span class="label">提交数：</span>
              <span class="value">{{ submissions.length }}人</span>
            </div>
            <div class="info-item" v-if="homework.tags?.length">
              <span class="label">标签：</span>
              <div class="tags">
                <el-tag 
                  v-for="tag in homework.tags" 
                  :key="tag" 
                  size="small"
                  class="tag"
                >
                  {{ tag }}
                </el-tag>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 学生视图：我的提交 -->
      <div v-if="isStudent" class="my-submission">
        <el-card title="我的提交">
          <template #header>
            <span>我的提交</span>
          </template>
          
          <div v-if="mySubmission">
            <div class="submission-info">
              <div class="submission-status">
                <el-tag :type="getSubmissionStatusType(mySubmission.status)" size="large">
                  {{ getSubmissionStatusText(mySubmission.status) }}
                </el-tag>
              </div>
              <div class="submission-meta">
                <span>提交时间：{{ formatDate(mySubmission.submitted_at) }}</span>
                <span v-if="mySubmission.grade">成绩：{{ mySubmission.grade }}分</span>
              </div>
            </div>
            
            <div class="submission-content" v-if="mySubmission.content">
              <h4>提交内容</h4>
              <div class="content">{{ mySubmission.content }}</div>
            </div>
            
            <div class="submission-feedback" v-if="mySubmission.feedback">
              <h4>教师反馈</h4>
              <div class="feedback">{{ mySubmission.feedback }}</div>
            </div>
            
            <div class="submission-branch" v-if="mySubmission.gitlab_branch">
              <h4>提交分支</h4>
              <el-link 
                :href="getBranchUrl(mySubmission.gitlab_branch)" 
                type="primary"
                target="_blank"
              >
                {{ mySubmission.gitlab_branch }}
                <el-icon><Link /></el-icon>
              </el-link>
            </div>
          </div>
          
          <div v-else class="no-submission">
            <el-empty description="尚未提交作业">
              <el-button v-if="canSubmit" type="primary" @click="submitHomework">
                立即提交
              </el-button>
            </el-empty>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { homeworkService } from '@/services/api'
import type { Homework, Submission } from '@/types'
import { ElMessage } from 'element-plus'
import { 
  Edit, Upload, Check, Link 
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const homework = ref<Homework | null>(null)
const submissions = ref<Submission[]>([])
const mySubmission = ref<Submission | null>(null)
const loading = ref(false)

// 计算属性
const homeworkId = computed(() => route.params.id as string)

const isTeacher = computed(() => 
  userStore.hasRole('teacher') || userStore.hasRole('admin')
)

const isStudent = computed(() => userStore.hasRole('student'))

const canEdit = computed(() => {
  if (!homework.value) return false
  return isTeacher.value && (
    userStore.hasRole('admin') || 
    homework.value.creator_id === userStore.user?.id
  )
})

const canSubmit = computed(() => {
  if (!homework.value) return false
  return isStudent.value && 
         homework.value.status === 'published' && 
         !isOverdue.value &&
         !mySubmission.value
})

const canGrade = computed(() => {
  if (!homework.value) return false
  return canEdit.value && homework.value.status === 'published'
})

const isOverdue = computed(() => {
  if (!homework.value?.due_date) return false
  return new Date() > new Date(homework.value.due_date)
})

const isDueSoon = computed(() => {
  if (!homework.value?.due_date) return false
  const deadline = new Date(homework.value.due_date)
  const now = new Date()
  const timeDiff = deadline.getTime() - now.getTime()
  const daysDiff = timeDiff / (1000 * 3600 * 24)
  return daysDiff > 0 && daysDiff <= 3
})

// 方法
const fetchHomework = async () => {
  loading.value = true
  try {
    const response = await homeworkService.getHomeworkById(homeworkId.value)
    homework.value = response.data
  } catch (error) {
    ElMessage.error('获取作业详情失败')
  } finally {
    loading.value = false
  }
}

const fetchMySubmission = async () => {
  if (!isStudent.value) return
  
  try {
    const response = await homeworkService.getMySubmission(homeworkId.value)
    mySubmission.value = response.data
  } catch (error) {
    mySubmission.value = null
  }
}

const editHomework = () => {
  router.push(`/homeworks/${homeworkId.value}/edit`)
}

const submitHomework = () => {
  router.push(`/homeworks/${homeworkId.value}/submit`)
}

const gradeHomework = () => {
  router.push(`/homeworks/${homeworkId.value}/grade`)
}

const getBranchUrl = (branchName: string) => {
  if (!homework.value?.project?.gitlab_url) return '#'
  return `${homework.value.project.gitlab_url}/-/tree/${branchName}`
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    draft: 'info',
    published: 'success',
    closed: 'warning'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    draft: '草稿',
    published: '已发布',
    closed: '已结束'
  }
  return texts[status] || status
}

const getSubmissionStatusType = (status: string) => {
  const types: Record<string, string> = {
    pending: 'info',
    submitted: 'warning',
    graded: 'success'
  }
  return types[status] || 'info'
}

const getSubmissionStatusText = (status: string) => {
  const texts: Record<string, string> = {
    pending: '未提交',
    submitted: '已提交',
    graded: '已批改'
  }
  return texts[status] || status
}

const formatDate = (date: string | Date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const formatDescription = (text: string) => {
  if (!text) return ''
  return text.replace(/\n/g, '<br>')
}

// 生命周期
onMounted(async () => {
  await fetchHomework()
  if (isStudent.value) {
    await fetchMySubmission()
  }
})
</script>

<style scoped>
.homework-detail {
  padding: 20px;
}

.homework-container {
  max-width: 1200px;
  margin: 0 auto;
}

.homework-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e8e8e8;
}

.homework-title {
  margin: 0 0 12px 0;
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.homework-badges {
  display: flex;
  gap: 8px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.homework-info {
  margin-bottom: 24px;
}

.homework-description,
.homework-requirements,
.homework-instructions {
  margin-bottom: 20px;
}

.homework-description h3,
.homework-requirements h3,
.homework-instructions h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.content {
  line-height: 1.6;
  color: #606266;
}

.homework-requirements ul {
  margin: 0;
  padding-left: 20px;
}

.homework-requirements li {
  margin-bottom: 8px;
  color: #606266;
}

.info-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
}

.info-item .label {
  font-weight: 500;
  color: #909399;
  min-width: 80px;
}

.info-item .value {
  color: #303133;
}

.info-item .value.overdue {
  color: #f56c6c;
  font-weight: 500;
}

.tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.tag {
  margin: 0;
}

.my-submission {
  margin-top: 24px;
}

.submission-info {
  margin-bottom: 16px;
}

.submission-status {
  margin-bottom: 8px;
}

.submission-meta {
  font-size: 14px;
  color: #909399;
}

.submission-meta span {
  margin-right: 16px;
}

.submission-content,
.submission-feedback,
.submission-branch {
  margin-top: 16px;
}

.submission-content h4,
.submission-feedback h4,
.submission-branch h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.submission-content .content,
.submission-feedback .feedback {
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  line-height: 1.6;
}

.no-submission {
  text-align: center;
  padding: 40px 0;
}

@media (max-width: 768px) {
  .homework-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .header-actions {
    width: 100%;
    justify-content: flex-start;
  }
  
  .homework-info .el-col {
    margin-bottom: 16px;
  }
}
</style>