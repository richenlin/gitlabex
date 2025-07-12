<template>
  <div class="assignment-management">
    <div class="page-header">
      <h1>作业管理</h1>
      <div class="header-actions">
        <el-select 
          v-model="selectedClassId" 
          placeholder="选择班级"
          @change="loadAssignments"
          style="width: 200px; margin-right: 10px;"
        >
          <el-option
            v-for="classItem in classList"
            :key="classItem.id"
            :label="classItem.name"
            :value="classItem.id"
          />
        </el-select>
        <el-button type="primary" @click="showCreateDialog" :disabled="!selectedClassId">
          <el-icon><Plus /></el-icon>
          布置作业
        </el-button>
      </div>
    </div>

    <!-- 作业列表 -->
    <div class="assignment-list" v-if="selectedClassId">
      <el-row :gutter="20">
        <el-col 
          v-for="assignment in assignmentList" 
          :key="assignment.id" 
          :xs="24" :sm="12" :md="8" :lg="6"
        >
          <el-card 
            class="assignment-card" 
            shadow="hover"
            @click="viewAssignmentDetails(assignment)"
          >
            <div class="assignment-header">
              <h3>{{ assignment.title }}</h3>
              <el-tag :type="getAssignmentStatusType(assignment)">
                {{ getAssignmentStatus(assignment) }}
              </el-tag>
            </div>
            
            <div class="assignment-description">
              {{ assignment.description || '暂无描述' }}
            </div>
            
            <div class="assignment-info">
              <div class="info-item">
                <el-icon><Calendar /></el-icon>
                <span>创建: {{ formatDate(assignment.created_at) }}</span>
              </div>
              <div class="info-item" v-if="assignment.due_date">
                <el-icon><Clock /></el-icon>
                <span>截止: {{ formatDate(assignment.due_date) }}</span>
              </div>
              <div class="info-item">
                <el-icon><User /></el-icon>
                <span>提交: {{ assignment.submitCount || 0 }}/{{ assignment.totalStudents || 0 }}</span>
              </div>
            </div>
            
            <div class="assignment-actions">
              <el-button-group size="small">
                <el-button @click.stop="viewSubmissions(assignment)">
                  查看提交
                </el-button>
                <el-button @click.stop="editAssignment(assignment)">
                  编辑
                </el-button>
              </el-button-group>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <el-empty description="请先选择班级查看作业"></el-empty>
    </div>

    <!-- 创建作业对话框 -->
    <el-dialog 
      v-model="createDialogVisible" 
      title="布置作业" 
      width="600px"
    >
      <el-form 
        ref="createFormRef" 
        :model="createForm" 
        :rules="createRules" 
        label-width="100px"
      >
        <el-form-item label="作业标题" prop="title">
          <el-input 
            v-model="createForm.title" 
            placeholder="请输入作业标题"
          />
        </el-form-item>
        
        <el-form-item label="作业描述" prop="description">
          <el-input 
            v-model="createForm.description" 
            type="textarea" 
            :rows="4"
            placeholder="请输入作业要求和说明"
          />
        </el-form-item>
        
        <el-form-item label="截止时间" prop="dueDate">
          <el-date-picker
            v-model="createForm.dueDate"
            type="datetime"
            placeholder="选择截止时间"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%;"
          />
        </el-form-item>
        
        <el-form-item label="作业附件">
          <el-upload
            ref="uploadRef"
            :file-list="createForm.attachments"
            :on-change="handleFileChange"
            :auto-upload="false"
            multiple
          >
            <el-button>
              <el-icon><UploadFilled /></el-icon>
              上传附件
            </el-button>
            <template #tip>
              <div class="el-upload__tip">
                支持上传多个文件，文件大小不超过10MB
              </div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createAssignment" :loading="createLoading">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 作业详情对话框 -->
    <el-dialog 
      v-model="detailDialogVisible" 
      :title="selectedAssignment?.title" 
      width="1000px"
    >
      <div v-if="selectedAssignment" class="assignment-details">
        <el-tabs v-model="activeTab">
          <!-- 作业信息 -->
          <el-tab-pane label="作业信息" name="info">
            <div class="assignment-info-detail">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="作业标题">
                  {{ selectedAssignment.title }}
                </el-descriptions-item>
                <el-descriptions-item label="状态">
                  <el-tag :type="getAssignmentStatusType(selectedAssignment)">
                    {{ getAssignmentStatus(selectedAssignment) }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="创建时间">
                  {{ formatDate(selectedAssignment.created_at) }}
                </el-descriptions-item>
                <el-descriptions-item label="截止时间">
                  {{ formatDate(selectedAssignment.due_date) }}
                </el-descriptions-item>
                <el-descriptions-item label="作业描述" :span="2">
                  <div class="assignment-description-detail">
                    {{ selectedAssignment.description }}
                  </div>
                </el-descriptions-item>
              </el-descriptions>
              
              <!-- 作业附件 -->
              <div v-if="assignmentAttachments.length > 0" class="assignment-attachments">
                <h4>作业附件</h4>
                <div class="attachment-list">
                  <div 
                    v-for="attachment in assignmentAttachments" 
                    :key="attachment.id"
                    class="attachment-item"
                  >
                    <el-icon><Document /></el-icon>
                    <span>{{ attachment.filename }}</span>
                    <el-button size="small" @click="downloadAttachment(attachment)">
                      下载
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </el-tab-pane>

          <!-- 学生提交 -->
          <el-tab-pane label="学生提交" name="submissions">
            <div class="submissions-management">
              <div class="submissions-header">
                <el-row :gutter="20">
                  <el-col :span="6">
                    <el-statistic title="总人数" :value="submissionStats.totalStudents" />
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="已提交" :value="submissionStats.submitted" />
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="未提交" :value="submissionStats.notSubmitted" />
                  </el-col>
                  <el-col :span="6">
                    <el-statistic title="已批改" :value="submissionStats.graded" />
                  </el-col>
                </el-row>
              </div>
              
              <el-table :data="submissionList" style="width: 100%; margin-top: 20px;">
                <el-table-column prop="student_name" label="学生姓名" />
                <el-table-column prop="student_username" label="学生学号" />
                <el-table-column label="提交状态">
                  <template #default="{ row }">
                    <el-tag :type="getSubmissionStatusType(row.status)">
                      {{ getSubmissionStatusText(row.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="submitted_at" label="提交时间">
                  <template #default="{ row }">
                    {{ row.submitted_at ? formatDate(row.submitted_at) : '-' }}
                  </template>
                </el-table-column>
                <el-table-column prop="grade" label="成绩">
                  <template #default="{ row }">
                    {{ row.grade || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="200">
                  <template #default="{ row }">
                    <el-button-group size="small">
                      <el-button 
                        v-if="row.status === 'submitted'"
                        @click="viewSubmission(row)"
                      >
                        查看提交
                      </el-button>
                      <el-button 
                        v-if="row.status === 'submitted'"
                        type="primary"
                        @click="gradeSubmission(row)"
                      >
                        批改
                      </el-button>
                    </el-button-group>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>

    <!-- 批改作业对话框 -->
    <el-dialog 
      v-model="gradeDialogVisible" 
      title="批改作业" 
      width="600px"
    >
      <div v-if="selectedSubmission" class="grade-form">
        <div class="student-info">
          <h4>学生信息</h4>
          <p>姓名: {{ selectedSubmission.student_name }}</p>
          <p>学号: {{ selectedSubmission.student_username }}</p>
          <p>提交时间: {{ formatDate(selectedSubmission.submitted_at) }}</p>
        </div>
        
        <el-form 
          ref="gradeFormRef" 
          :model="gradeForm" 
          :rules="gradeRules" 
          label-width="80px"
        >
          <el-form-item label="成绩" prop="grade">
            <el-input-number 
              v-model="gradeForm.grade"
              :min="0"
              :max="100"
              :precision="1"
            />
            <span style="margin-left: 10px;">分</span>
          </el-form-item>
          
          <el-form-item label="评语" prop="feedback">
            <el-input 
              v-model="gradeForm.feedback" 
              type="textarea" 
              :rows="4"
              placeholder="请输入批改评语"
            />
          </el-form-item>
        </el-form>
      </div>
      
      <template #footer>
        <el-button @click="gradeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitGrade" :loading="gradeLoading">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, 
  Calendar, 
  Clock, 
  User, 
  Document, 
  UploadFilled 
} from '@element-plus/icons-vue'
import { useRoute } from 'vue-router'

const route = useRoute()

// 响应式数据
const classList = ref([])
const assignmentList = ref([])
const submissionList = ref([])
const assignmentAttachments = ref([])

const selectedClassId = ref(null)
const selectedAssignment = ref(null)
const selectedSubmission = ref(null)

const createDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const gradeDialogVisible = ref(false)

const createLoading = ref(false)
const gradeLoading = ref(false)

const activeTab = ref('info')

// 表单数据
const createForm = reactive({
  title: '',
  description: '',
  dueDate: null,
  attachments: []
})

const gradeForm = reactive({
  grade: null,
  feedback: ''
})

// 表单引用
const createFormRef = ref()
const gradeFormRef = ref()
const uploadRef = ref()

// 表单验证规则
const createRules = {
  title: [
    { required: true, message: '请输入作业标题', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入作业描述', trigger: 'blur' }
  ]
}

const gradeRules = {
  grade: [
    { required: true, message: '请输入成绩', trigger: 'blur' }
  ],
  feedback: [
    { required: true, message: '请输入评语', trigger: 'blur' }
  ]
}

// 计算属性
const submissionStats = computed(() => {
  if (!submissionList.value.length) {
    return {
      totalStudents: 0,
      submitted: 0,
      notSubmitted: 0,
      graded: 0
    }
  }
  
  const totalStudents = submissionList.value.length
  const submitted = submissionList.value.filter(s => s.status === 'submitted' || s.status === 'graded').length
  const graded = submissionList.value.filter(s => s.status === 'graded').length
  
  return {
    totalStudents,
    submitted,
    notSubmitted: totalStudents - submitted,
    graded
  }
})

// 页面初始化
onMounted(() => {
  loadClassList()
  
  // 从URL参数获取班级ID
  if (route.query.classId) {
    selectedClassId.value = parseInt(route.query.classId)
    loadAssignments()
  }
})

// 加载班级列表
const loadClassList = async () => {
  try {
    const response = await fetch('/api/teams/user/current', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      classList.value = data.data || []
    }
  } catch (error) {
    console.error('加载班级列表失败:', error)
    ElMessage.error('加载班级列表失败')
  }
}

// 加载作业列表
const loadAssignments = async () => {
  if (!selectedClassId.value) return
  
  try {
    const response = await fetch(`/api/education/assignments?group_id=${selectedClassId.value}`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      assignmentList.value = data.data || []
    }
  } catch (error) {
    console.error('加载作业列表失败:', error)
    ElMessage.error('加载作业列表失败')
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  createDialogVisible.value = true
  // 重置表单
  Object.assign(createForm, {
    title: '',
    description: '',
    dueDate: null,
    attachments: []
  })
}

// 处理文件变更
const handleFileChange = (file) => {
  createForm.attachments = uploadRef.value.fileList
}

// 创建作业
const createAssignment = async () => {
  if (!createFormRef.value) return
  
  const valid = await createFormRef.value.validate()
  if (!valid) return
  
  createLoading.value = true
  
  try {
    const formData = new FormData()
    formData.append('group_id', selectedClassId.value)
    formData.append('title', createForm.title)
    formData.append('description', createForm.description)
    if (createForm.dueDate) {
      formData.append('due_date', createForm.dueDate)
    }
    
    // 添加附件
    createForm.attachments.forEach(file => {
      formData.append('attachments', file.raw)
    })
    
    const response = await fetch('/api/education/assignments', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: formData
    })
    
    if (response.ok) {
      ElMessage.success('作业创建成功')
      createDialogVisible.value = false
      loadAssignments()
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '创建作业失败')
    }
  } catch (error) {
    console.error('创建作业失败:', error)
    ElMessage.error('创建作业失败')
  } finally {
    createLoading.value = false
  }
}

// 查看作业详情
const viewAssignmentDetails = async (assignment) => {
  selectedAssignment.value = assignment
  detailDialogVisible.value = true
  activeTab.value = 'info'
  
  // 加载作业附件和提交情况
  await loadAssignmentAttachments(assignment.id)
  await loadSubmissions(assignment.id)
}

// 加载作业附件
const loadAssignmentAttachments = async (assignmentId) => {
  try {
    const response = await fetch(`/api/education/assignments/${assignmentId}/attachments`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      assignmentAttachments.value = data.data || []
    }
  } catch (error) {
    console.error('加载作业附件失败:', error)
  }
}

// 加载提交情况
const loadSubmissions = async (assignmentId) => {
  try {
    const response = await fetch(`/api/education/assignments/${assignmentId}/submissions`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      submissionList.value = data.data || []
    }
  } catch (error) {
    console.error('加载提交情况失败:', error)
  }
}

// 查看提交
const viewSubmissions = (assignment) => {
  viewAssignmentDetails(assignment)
  activeTab.value = 'submissions'
}

// 批改作业
const gradeSubmission = (submission) => {
  selectedSubmission.value = submission
  gradeDialogVisible.value = true
  Object.assign(gradeForm, {
    grade: null,
    feedback: ''
  })
}

// 提交成绩
const submitGrade = async () => {
  if (!gradeFormRef.value) return
  
  const valid = await gradeFormRef.value.validate()
  if (!valid) return
  
  gradeLoading.value = true
  
  try {
    const response = await fetch(`/api/education/submissions/${selectedSubmission.value.id}/grade`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(gradeForm)
    })
    
    if (response.ok) {
      ElMessage.success('批改完成')
      gradeDialogVisible.value = false
      loadSubmissions(selectedAssignment.value.id)
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '批改失败')
    }
  } catch (error) {
    console.error('批改失败:', error)
    ElMessage.error('批改失败')
  } finally {
    gradeLoading.value = false
  }
}

// 下载附件
const downloadAttachment = (attachment) => {
  window.open(attachment.download_url, '_blank')
}

// 工具函数
const getAssignmentStatusType = (assignment) => {
  if (!assignment.due_date) return 'info'
  
  const now = new Date()
  const dueDate = new Date(assignment.due_date)
  
  if (now > dueDate) {
    return 'danger' // 已过期
  } else if (now > new Date(dueDate.getTime() - 24 * 60 * 60 * 1000)) {
    return 'warning' // 即将过期
  } else {
    return 'success' // 正常
  }
}

const getAssignmentStatus = (assignment) => {
  if (!assignment.due_date) return '无截止时间'
  
  const now = new Date()
  const dueDate = new Date(assignment.due_date)
  
  if (now > dueDate) {
    return '已过期'
  } else if (now > new Date(dueDate.getTime() - 24 * 60 * 60 * 1000)) {
    return '即将过期'
  } else {
    return '进行中'
  }
}

const getSubmissionStatusType = (status) => {
  switch (status) {
    case 'submitted':
      return 'warning'
    case 'graded':
      return 'success'
    default:
      return 'info'
  }
}

const getSubmissionStatusText = (status) => {
  switch (status) {
    case 'submitted':
      return '已提交'
    case 'graded':
      return '已批改'
    default:
      return '未提交'
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}
</script>

<style scoped>
.assignment-management {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  color: #303133;
}

.header-actions {
  display: flex;
  align-items: center;
}

.assignment-list {
  margin-top: 20px;
}

.assignment-card {
  cursor: pointer;
  transition: all 0.3s;
  height: 280px;
}

.assignment-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.assignment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.assignment-header h3 {
  margin: 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.assignment-description {
  color: #606266;
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 16px;
  height: 42px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.assignment-info {
  margin-bottom: 16px;
}

.info-item {
  display: flex;
  align-items: center;
  color: #909399;
  font-size: 12px;
  margin-bottom: 4px;
}

.info-item .el-icon {
  margin-right: 4px;
}

.assignment-actions {
  text-align: center;
}

.assignment-details {
  margin-top: 20px;
}

.assignment-info-detail {
  padding: 20px 0;
}

.assignment-description-detail {
  white-space: pre-wrap;
  line-height: 1.6;
}

.assignment-attachments {
  margin-top: 20px;
}

.assignment-attachments h4 {
  margin-bottom: 10px;
  color: #303133;
}

.attachment-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.attachment-item {
  display: flex;
  align-items: center;
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.attachment-item .el-icon {
  margin-right: 8px;
  color: #409eff;
}

.attachment-item span {
  flex: 1;
}

.submissions-management {
  padding: 20px 0;
}

.submissions-header {
  margin-bottom: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
}

.grade-form {
  padding: 20px 0;
}

.student-info {
  padding: 15px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 20px;
}

.student-info h4 {
  margin-top: 0;
  margin-bottom: 10px;
  color: #303133;
}

.student-info p {
  margin: 5px 0;
  color: #606266;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

@media (max-width: 768px) {
  .assignment-management {
    padding: 10px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .header-actions {
    width: 100%;
  }
  
  .header-actions .el-select {
    width: 100% !important;
    margin-right: 0 !important;
    margin-bottom: 10px;
  }
}
</style> 