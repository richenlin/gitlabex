<template>
  <div class="homework-list">
    <div class="header">
      <h1>作业管理</h1>
      <el-button 
        type="primary" 
        @click="showCreateDialog"
        v-if="canCreateHomework"
      >
        <el-icon><Plus /></el-icon>
        发布作业
      </el-button>
    </div>

    <div class="filters">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-input
            v-model="searchQuery"
            placeholder="搜索作业..."
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="5">
          <el-select v-model="projectFilter" placeholder="选择课题" clearable @change="fetchHomeworks">
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="statusFilter" placeholder="状态" @change="fetchHomeworks">
            <el-option label="全部" value="" />
            <el-option label="草稿" value="draft" />
            <el-option label="已发布" value="published" />
            <el-option label="已结束" value="closed" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="viewFilter" placeholder="查看" @change="fetchHomeworks">
            <el-option label="全部作业" value="all" />
            <el-option label="我发布的" value="my" />
            <el-option label="我的作业" value="assigned" />
          </el-select>
        </el-col>
      </el-row>
    </div>

    <div class="homework-stats">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-statistic title="总作业数" :value="total" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="进行中" :value="activeHomeworks" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="待提交" :value="pendingSubmissions" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="待批改" :value="pendingGrades" />
        </el-col>
      </el-row>
    </div>

    <div class="homework-content" v-loading="loading">
      <div
        v-for="homework in homeworks"
        :key="homework.id"
        class="homework-item"
        @click="viewHomework(homework.id)"
      >
        <div class="homework-header">
          <div class="homework-title-section">
            <h3 class="homework-title">{{ homework.title }}</h3>
            <div class="homework-badges">
              <el-tag :type="getStatusType(homework.status)" size="small">
                {{ getStatusText(homework.status) }}
              </el-tag>
              <el-tag v-if="isOverdue(homework)" type="danger" size="small">
                已逾期
              </el-tag>
              <el-tag v-if="isDueSoon(homework)" type="warning" size="small">
                即将截止
              </el-tag>
            </div>
          </div>
          <div class="homework-actions" @click.stop>
            <el-dropdown>
              <el-icon><More /></el-icon>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="viewHomework(homework.id)">查看详情</el-dropdown-item>
                  <el-dropdown-item 
                    v-if="canEdit(homework)"
                    @click="editHomework(homework.id)"
                  >
                    编辑
                  </el-dropdown-item>
                  <el-dropdown-item 
                    v-if="canSubmit(homework)"
                    @click="submitHomework(homework.id)"
                  >
                    提交作业
                  </el-dropdown-item>
                  <el-dropdown-item 
                    v-if="canGrade(homework)"
                    @click="gradeHomework(homework.id)"
                  >
                    批改作业
                  </el-dropdown-item>
                  <el-dropdown-item 
                    v-if="canDelete(homework)"
                    @click="deleteHomework(homework.id)"
                    divided
                  >
                    删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <div class="homework-description">
          {{ homework.description }}
        </div>

        <div class="homework-meta">
          <div class="meta-row">
            <div class="meta-item">
              <el-icon><User /></el-icon>
              <span>发布者：{{ homework.author?.name }}</span>
            </div>
            <div class="meta-item">
              <el-icon><Calendar /></el-icon>
              <span>截止时间：{{ formatDate(homework.deadline) }}</span>
            </div>
            <div class="meta-item" v-if="homework.project">
              <el-icon><Folder /></el-icon>
              <el-link :href="`/scenes/${homework.project.id}`" :underline="false">
                {{ homework.project.name }}
              </el-link>
            </div>
          </div>
          
          <div class="submission-stats" v-if="isTeacher">
            <div class="stat-item">
              <span class="stat-label">提交数：</span>
              <span class="stat-value">{{ homework.submissions?.length || 0 }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">已批改：</span>
              <span class="stat-value">{{ getGradedCount(homework) }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">平均分：</span>
              <span class="stat-value">{{ getAverageGrade(homework) }}</span>
            </div>
          </div>

          <div class="student-status" v-else>
            <div class="submission-status">
              {{ getSubmissionStatus(homework) }}
            </div>
          </div>
        </div>

        <div class="homework-progress" v-if="isTeacher">
          <el-progress 
            :percentage="getSubmissionProgress(homework)"
            :format="() => `${homework.submissions?.length || 0} 人已提交`"
          />
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
        @current-change="fetchHomeworks"
        @size-change="fetchHomeworks"
      />
    </div>

    <!-- 创建作业对话框 -->
    <el-dialog v-model="createDialogVisible" title="发布作业" width="700px">
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="100px"
      >
        <el-form-item label="作业标题" prop="title">
          <el-input v-model="createForm.title" placeholder="请输入作业标题" />
        </el-form-item>
        <el-form-item label="作业描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入作业描述和要求"
          />
        </el-form-item>
        <el-form-item label="关联课题" prop="project_id">
          <el-select v-model="createForm.project_id" placeholder="选择课题">
            <el-option
              v-for="project in projects"
              :key="project.id"
              :label="project.name"
              :value="project.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="截止时间" prop="deadline">
          <el-date-picker
            v-model="createForm.deadline"
            type="datetime"
            placeholder="选择截止时间"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>
        <el-form-item label="满分" prop="max_grade">
          <el-input-number v-model="createForm.max_grade" :min="1" :max="1000" />
        </el-form-item>
        <el-form-item label="作业类型">
          <el-radio-group v-model="createForm.type">
            <el-radio label="individual">个人作业</el-radio>
            <el-radio label="group">小组作业</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="提交格式">
          <el-checkbox-group v-model="createForm.allowed_formats">
            <el-checkbox label="text">文本</el-checkbox>
            <el-checkbox label="file">文件</el-checkbox>
            <el-checkbox label="link">链接</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button @click="saveDraft" :loading="saving">保存草稿</el-button>
          <el-button type="primary" @click="publishHomework" :loading="publishing">
            发布作业
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { homeworkService, researchService } from '@/services/api'
import type { Homework, ResearchProject } from '@/types'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import {
  Plus, Search, More, User, Calendar, Folder
} from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const homeworks = ref<Homework[]>([])
const projects = ref<ResearchProject[]>([])
const loading = ref(false)
const saving = ref(false)
const publishing = ref(false)
const searchQuery = ref('')
const projectFilter = ref('')
const statusFilter = ref('')
const viewFilter = ref('all')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const createDialogVisible = ref(false)
const createFormRef = ref<FormInstance>()

const createForm = ref({
  title: '',
  description: '',
  project_id: '',
  deadline: '',
  max_grade: 100,
  type: 'individual',
  allowed_formats: ['text', 'file']
})

const createRules = {
  title: [{ required: true, message: '请输入作业标题', trigger: 'blur' }],
  description: [{ required: true, message: '请输入作业描述', trigger: 'blur' }],
  project_id: [{ required: true, message: '请选择课题', trigger: 'change' }],
  deadline: [{ required: true, message: '请选择截止时间', trigger: 'change' }],
  max_grade: [{ required: true, message: '请设置满分', trigger: 'blur' }]
}

// 计算属性
const isTeacher = computed(() => {
  return userStore.hasRole('teacher') || userStore.hasRole('admin')
})

const canCreateHomework = computed(() => {
  return isTeacher.value
})

const activeHomeworks = computed(() => {
  return homeworks.value.filter(hw => hw.status === 'published').length
})

const pendingSubmissions = computed(() => {
  // TODO: 从用户角度计算待提交作业数
  return 0
})

const pendingGrades = computed(() => {
  // TODO: 从教师角度计算待批改作业数
  return 0
})

// 方法
const fetchHomeworks = async () => {
  loading.value = true
  try {
    const params: any = {
      page: currentPage.value,
      pageSize: pageSize.value,
      search: searchQuery.value || undefined,
      projectId: projectFilter.value || undefined,
      status: statusFilter.value || undefined
    }

    if (viewFilter.value === 'my') {
      params.authorId = userStore.user?.id
    }

    const response = await homeworkService.getHomeworks(params)
    homeworks.value = response.data?.items || []
    total.value = response.data?.total || 0
  } catch (error) {
    console.error('获取作业列表失败:', error)
    ElMessage.error('获取作业列表失败')
  } finally {
    loading.value = false
  }
}

const fetchProjects = async () => {
  try {
    const response = await researchService.getProjects()
    projects.value = response.data?.items || []
  } catch (error) {
    console.error('获取课题列表失败:', error)
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchHomeworks()
}

const viewHomework = (id: string) => {
  router.push(`/homeworks/${id}`)
}

const editHomework = (id: string) => {
  // TODO: 实现编辑功能
  ElMessage.info('编辑功能待实现')
}

const submitHomework = (id: string) => {
  router.push(`/homeworks/${id}/submit`)
}

const gradeHomework = (id: string) => {
  router.push(`/homeworks/${id}/grade`)
}

const deleteHomework = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除这个作业吗？', '确认删除', {
      type: 'warning'
    })
    
    await homeworkService.deleteHomework(id)
    ElMessage.success('作业删除成功')
    fetchHomeworks()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除作业失败:', error)
      ElMessage.error('删除作业失败')
    }
  }
}

const canEdit = (homework: Homework) => {
  return isTeacher.value && (
    userStore.hasRole('admin') || homework.creator_id === userStore.user?.id
  )
}

const canSubmit = (homework: Homework) => {
  return !isTeacher.value && homework.status === 'published' && !isOverdue(homework)
}

const canGrade = (homework: Homework) => {
  return canEdit(homework) && homework.status === 'published'
}

const canDelete = (homework: Homework) => {
  return canEdit(homework)
}

const showCreateDialog = () => {
  createDialogVisible.value = true
  resetCreateForm()
}

const resetCreateForm = () => {
  createForm.value = {
    title: '',
    description: '',
    project_id: '',
    deadline: '',
    max_grade: 100,
    type: 'individual',
    allowed_formats: ['text', 'file']
  }
}

const saveDraft = async () => {
  if (!createFormRef.value) return
  
  try {
    await createFormRef.value.validate()
    saving.value = true
    
    await homeworkService.createHomework({
      ...createForm.value,
      status: 'draft'
    })
    
    ElMessage.success('草稿保存成功')
    createDialogVisible.value = false
    fetchHomeworks()
  } catch (error) {
    console.error('保存草稿失败:', error)
    ElMessage.error('保存草稿失败')
  } finally {
    saving.value = false
  }
}

const publishHomework = async () => {
  if (!createFormRef.value) return
  
  try {
    await createFormRef.value.validate()
    publishing.value = true
    
    await homeworkService.createHomework({
      ...createForm.value,
      status: 'published'
    })
    
    ElMessage.success('作业发布成功')
    createDialogVisible.value = false
    fetchHomeworks()
  } catch (error) {
    console.error('发布作业失败:', error)
    ElMessage.error('发布作业失败')
  } finally {
    publishing.value = false
  }
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

const isOverdue = (homework: Homework) => {
  return new Date() > new Date(homework.deadline)
}

const isDueSoon = (homework: Homework) => {
  const deadline = new Date(homework.deadline)
  const now = new Date()
  const timeDiff = deadline.getTime() - now.getTime()
  const daysDiff = timeDiff / (1000 * 3600 * 24)
  return daysDiff > 0 && daysDiff <= 3
}

const getGradedCount = (homework: Homework) => {
  return homework.submissions?.filter(sub => sub.grade !== null).length || 0
}

const getAverageGrade = (homework: Homework) => {
  const gradedSubmissions = homework.submissions?.filter(sub => sub.grade !== null) || []
  if (gradedSubmissions.length === 0) return '-'
  
  const total = gradedSubmissions.reduce((sum, sub) => sum + (sub.grade || 0), 0)
  return Math.round(total / gradedSubmissions.length)
}

const getSubmissionProgress = (homework: Homework) => {
  // TODO: 根据课题成员数计算提交进度
  const submitted = homework.submissions?.length || 0
  const total = 50 // 假设值，应该从课题成员数获取
  return Math.round((submitted / total) * 100)
}

const getSubmissionStatus = (homework: Homework) => {
  // TODO: 根据当前用户的提交状态返回
  if (isOverdue(homework)) return '已逾期'
  return '未提交'
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN')
}

// 生命周期
onMounted(() => {
  fetchHomeworks()
  fetchProjects()
})
</script>

<style scoped>
.homework-list {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filters {
  margin-bottom: 20px;
}

.homework-stats {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.homework-content {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.homework-item {
  background: white;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s;
}

.homework-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.homework-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 15px;
}

.homework-title-section {
  flex: 1;
}

.homework-title {
  margin: 0 0 10px 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.homework-badges {
  display: flex;
  gap: 8px;
}

.homework-actions {
  cursor: pointer;
}

.homework-description {
  color: #606266;
  margin-bottom: 15px;
  line-height: 1.6;
}

.homework-meta {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.meta-row {
  display: flex;
  gap: 30px;
  flex-wrap: wrap;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 14px;
  color: #606266;
}

.submission-stats {
  display: flex;
  gap: 20px;
}

.stat-item {
  font-size: 14px;
}

.stat-label {
  color: #909399;
}

.stat-value {
  color: #303133;
  font-weight: 500;
}

.student-status {
  display: flex;
  align-items: center;
}

.submission-status {
  font-size: 14px;
  font-weight: 500;
  color: #e6a23c;
}

.homework-progress {
  margin-top: 15px;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 30px;
}
</style>
