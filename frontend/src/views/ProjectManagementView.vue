<template>
  <div class="project-management">
    <div class="page-header">
      <h1>课题管理</h1>
      <div class="header-actions">
        <el-select 
          v-model="selectedClassId" 
          placeholder="选择班级"
          @change="loadProjects"
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
          创建课题
        </el-button>
      </div>
    </div>

    <!-- 课题列表 -->
    <div class="project-list" v-if="selectedClassId">
      <el-row :gutter="20">
        <el-col 
          v-for="project in projectList" 
          :key="project.id" 
          :xs="24" :sm="12" :md="8" :lg="6"
        >
          <el-card 
            class="project-card" 
            shadow="hover"
            @click="viewProjectDetails(project)"
          >
            <div class="project-header">
              <h3>{{ project.title }}</h3>
              <el-tag :type="getProjectStatusType(project)">
                {{ getProjectStatus(project) }}
              </el-tag>
            </div>
            
            <div class="project-description">
              {{ project.description || '暂无描述' }}
            </div>
            
            <div class="project-info">
              <div class="info-item">
                <el-icon><Calendar /></el-icon>
                <span>创建: {{ formatDate(project.created_at) }}</span>
              </div>
              <div class="info-item" v-if="project.due_date">
                <el-icon><Clock /></el-icon>
                <span>截止: {{ formatDate(project.due_date) }}</span>
              </div>
              <div class="info-item">
                <el-icon><Users /></el-icon>
                <span>成员: {{ project.memberCount || 0 }} 人</span>
              </div>
              <div class="info-item">
                <el-icon><TrendCharts /></el-icon>
                <span>进度: {{ project.progress || 0 }}%</span>
              </div>
            </div>
            
            <div class="project-actions">
              <el-button-group size="small">
                <el-button @click.stop="viewProgress(project)">
                  查看进度
                </el-button>
                <el-button @click.stop="manageMembers(project)">
                  成员管理
                </el-button>
              </el-button-group>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <el-empty description="请先选择班级查看课题"></el-empty>
    </div>

    <!-- 创建课题对话框 -->
    <el-dialog 
      v-model="createDialogVisible" 
      title="创建课题" 
      width="600px"
    >
      <el-form 
        ref="createFormRef" 
        :model="createForm" 
        :rules="createRules" 
        label-width="100px"
      >
        <el-form-item label="课题标题" prop="title">
          <el-input 
            v-model="createForm.title" 
            placeholder="请输入课题标题"
          />
        </el-form-item>
        
        <el-form-item label="课题描述" prop="description">
          <el-input 
            v-model="createForm.description" 
            type="textarea" 
            :rows="4"
            placeholder="请输入课题描述和要求"
          />
        </el-form-item>
        
        <el-form-item label="截止时间" prop="dueDate">
          <el-date-picker
            v-model="createForm.dueDate"
            type="date"
            placeholder="选择截止时间"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%;"
          />
        </el-form-item>
        
        <el-form-item label="课题类型" prop="type">
          <el-select v-model="createForm.type" placeholder="选择课题类型" style="width: 100%;">
            <el-option label="个人项目" value="individual" />
            <el-option label="小组项目" value="group" />
            <el-option label="研究课题" value="research" />
            <el-option label="实践项目" value="practice" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="最大成员数" prop="maxMembers" v-if="createForm.type === 'group'">
          <el-input-number 
            v-model="createForm.maxMembers"
            :min="2"
            :max="10"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createProject" :loading="createLoading">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 课题详情对话框 -->
    <el-dialog 
      v-model="detailDialogVisible" 
      :title="selectedProject?.title" 
      width="1000px"
    >
      <div v-if="selectedProject" class="project-details">
        <el-tabs v-model="activeTab">
          <!-- 课题信息 -->
          <el-tab-pane label="课题信息" name="info">
            <div class="project-info-detail">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="课题标题">
                  {{ selectedProject.title }}
                </el-descriptions-item>
                <el-descriptions-item label="状态">
                  <el-tag :type="getProjectStatusType(selectedProject)">
                    {{ getProjectStatus(selectedProject) }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="课题类型">
                  {{ getProjectTypeText(selectedProject.type) }}
                </el-descriptions-item>
                <el-descriptions-item label="创建时间">
                  {{ formatDate(selectedProject.created_at) }}
                </el-descriptions-item>
                <el-descriptions-item label="截止时间">
                  {{ formatDate(selectedProject.due_date) }}
                </el-descriptions-item>
                <el-descriptions-item label="当前进度">
                  <el-progress 
                    :percentage="selectedProject.progress || 0" 
                    :stroke-width="8"
                  />
                </el-descriptions-item>
                <el-descriptions-item label="课题描述" :span="2">
                  <div class="project-description-detail">
                    {{ selectedProject.description }}
                  </div>
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </el-tab-pane>

          <!-- 成员管理 -->
          <el-tab-pane label="成员管理" name="members">
            <div class="members-management">
              <div class="members-header">
                <el-button 
                  type="primary" 
                  size="small" 
                  @click="showAssignMemberDialog"
                  v-if="selectedProject.type === 'group'"
                >
                  <el-icon><Plus /></el-icon>
                  分配成员
                </el-button>
              </div>
              
              <el-table :data="projectMembers" style="width: 100%">
                <el-table-column prop="name" label="姓名" />
                <el-table-column prop="username" label="学号" />
                <el-table-column prop="email" label="邮箱" />
                <el-table-column prop="role" label="角色">
                  <template #default="{ row }">
                    <el-tag :type="getMemberRoleTagType(row.role)">
                      {{ getMemberRoleText(row.role) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="join_date" label="加入时间">
                  <template #default="{ row }">
                    {{ formatDate(row.join_date) }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="150">
                  <template #default="{ row }">
                    <el-button 
                      size="small" 
                      type="danger" 
                      @click="removeMember(row)"
                      v-if="row.role !== 'leader'"
                    >
                      移除
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-tab-pane>

          <!-- 进度跟踪 -->
          <el-tab-pane label="进度跟踪" name="progress">
            <div class="progress-tracking">
              <div class="progress-overview">
                <el-row :gutter="20">
                  <el-col :span="8">
                    <el-card>
                      <el-statistic title="总体进度" :value="selectedProject.progress || 0" suffix="%" />
                    </el-card>
                  </el-col>
                  <el-col :span="8">
                    <el-card>
                      <el-statistic title="里程碑完成" :value="progressStats.completedMilestones" :suffix="`/${progressStats.totalMilestones}`" />
                    </el-card>
                  </el-col>
                  <el-col :span="8">
                    <el-card>
                      <el-statistic title="剩余天数" :value="progressStats.remainingDays" suffix="天" />
                    </el-card>
                  </el-col>
                </el-row>
              </div>
              
              <!-- 里程碑列表 -->
              <div class="milestones-list">
                <h4>项目里程碑</h4>
                <el-button size="small" @click="showAddMilestoneDialog" style="margin-bottom: 10px;">
                  <el-icon><Plus /></el-icon>
                  添加里程碑
                </el-button>
                
                <el-timeline>
                  <el-timeline-item
                    v-for="milestone in milestones"
                    :key="milestone.id"
                    :timestamp="formatDate(milestone.due_date)"
                    :type="getMilestoneStatusType(milestone.status)"
                  >
                    <el-card>
                      <div class="milestone-content">
                        <div class="milestone-header">
                          <h4>{{ milestone.title }}</h4>
                          <el-tag :type="getMilestoneStatusType(milestone.status)">
                            {{ getMilestoneStatusText(milestone.status) }}
                          </el-tag>
                        </div>
                        <p>{{ milestone.description }}</p>
                        <div class="milestone-actions">
                          <el-button 
                            size="small" 
                            @click="updateMilestoneStatus(milestone)"
                            v-if="milestone.status !== 'completed'"
                          >
                            标记完成
                          </el-button>
                        </div>
                      </div>
                    </el-card>
                  </el-timeline-item>
                </el-timeline>
              </div>
            </div>
          </el-tab-pane>

          <!-- 相关文档 -->
          <el-tab-pane label="相关文档" name="documents">
            <div class="project-documents">
              <div class="documents-header">
                <el-button type="primary" size="small" @click="createProjectDocument">
                  <el-icon><Plus /></el-icon>
                  创建文档
                </el-button>
              </div>
              
              <el-table :data="projectDocuments" style="width: 100%">
                <el-table-column prop="title" label="文档标题" />
                <el-table-column prop="type" label="文档类型">
                  <template #default="{ row }">
                    <el-tag size="small">{{ row.type }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="updated_at" label="更新时间">
                  <template #default="{ row }">
                    {{ formatDate(row.updated_at) }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="150">
                  <template #default="{ row }">
                    <el-button size="small" @click="openDocument(row)">
                      查看
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>

    <!-- 分配成员对话框 -->
    <el-dialog 
      v-model="assignMemberDialogVisible" 
      title="分配成员" 
      width="500px"
    >
      <el-form 
        ref="assignMemberFormRef" 
        :model="assignMemberForm" 
        :rules="assignMemberRules" 
        label-width="80px"
      >
        <el-form-item label="选择学生" prop="studentIds">
          <el-select 
            v-model="assignMemberForm.studentIds" 
            placeholder="选择要分配的学生"
            multiple
            filterable
            style="width: 100%;"
          >
            <el-option
              v-for="student in availableStudents"
              :key="student.id"
              :label="`${student.name} (${student.username})`"
              :value="student.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="assignMemberDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="assignMembers" :loading="assignLoading">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 添加里程碑对话框 -->
    <el-dialog 
      v-model="addMilestoneDialogVisible" 
      title="添加里程碑" 
      width="500px"
    >
      <el-form 
        ref="milestoneFormRef" 
        :model="milestoneForm" 
        :rules="milestoneRules" 
        label-width="80px"
      >
        <el-form-item label="里程碑标题" prop="title">
          <el-input 
            v-model="milestoneForm.title" 
            placeholder="请输入里程碑标题"
          />
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="milestoneForm.description" 
            type="textarea" 
            :rows="3"
            placeholder="请输入里程碑描述"
          />
        </el-form-item>
        
        <el-form-item label="截止时间" prop="dueDate">
          <el-date-picker
            v-model="milestoneForm.dueDate"
            type="date"
            placeholder="选择截止时间"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%;"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="addMilestoneDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="addMilestone" :loading="milestoneLoading">
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
  Users, 
  TrendCharts
} from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// 响应式数据
const classList = ref([])
const projectList = ref([])
const projectMembers = ref([])
const projectDocuments = ref([])
const milestones = ref([])
const availableStudents = ref([])

const selectedClassId = ref(null)
const selectedProject = ref(null)

const createDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const assignMemberDialogVisible = ref(false)
const addMilestoneDialogVisible = ref(false)

const createLoading = ref(false)
const assignLoading = ref(false)
const milestoneLoading = ref(false)

const activeTab = ref('info')

// 表单数据
const createForm = reactive({
  title: '',
  description: '',
  dueDate: null,
  type: 'individual',
  maxMembers: 2
})

const assignMemberForm = reactive({
  studentIds: []
})

const milestoneForm = reactive({
  title: '',
  description: '',
  dueDate: null
})

// 表单引用
const createFormRef = ref()
const assignMemberFormRef = ref()
const milestoneFormRef = ref()

// 表单验证规则
const createRules = {
  title: [
    { required: true, message: '请输入课题标题', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入课题描述', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择课题类型', trigger: 'change' }
  ]
}

const assignMemberRules = {
  studentIds: [
    { required: true, message: '请选择学生', trigger: 'change' }
  ]
}

const milestoneRules = {
  title: [
    { required: true, message: '请输入里程碑标题', trigger: 'blur' }
  ],
  dueDate: [
    { required: true, message: '请选择截止时间', trigger: 'change' }
  ]
}

// 计算属性
const progressStats = computed(() => {
  if (!selectedProject.value || !milestones.value.length) {
    return {
      completedMilestones: 0,
      totalMilestones: 0,
      remainingDays: 0
    }
  }
  
  const totalMilestones = milestones.value.length
  const completedMilestones = milestones.value.filter(m => m.status === 'completed').length
  
  let remainingDays = 0
  if (selectedProject.value.due_date) {
    const dueDate = new Date(selectedProject.value.due_date)
    const now = new Date()
    remainingDays = Math.max(0, Math.ceil((dueDate - now) / (1000 * 60 * 60 * 24)))
  }
  
  return {
    completedMilestones,
    totalMilestones,
    remainingDays
  }
})

// 页面初始化
onMounted(() => {
  loadClassList()
  loadAvailableStudents()
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

// 加载课题列表
const loadProjects = async () => {
  if (!selectedClassId.value) return
  
  try {
    const response = await fetch(`/api/education/projects?group_id=${selectedClassId.value}`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      projectList.value = data.data || []
    }
  } catch (error) {
    console.error('加载课题列表失败:', error)
    ElMessage.error('加载课题列表失败')
  }
}

// 加载可用学生
const loadAvailableStudents = async () => {
  if (!selectedClassId.value) return
  
  try {
    const response = await fetch(`/api/teams/${selectedClassId.value}/members?role=student`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      availableStudents.value = data.data || []
    }
  } catch (error) {
    console.error('加载学生列表失败:', error)
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  createDialogVisible.value = true
  Object.assign(createForm, {
    title: '',
    description: '',
    dueDate: null,
    type: 'individual',
    maxMembers: 2
  })
}

// 创建课题
const createProject = async () => {
  if (!createFormRef.value) return
  
  const valid = await createFormRef.value.validate()
  if (!valid) return
  
  createLoading.value = true
  
  try {
    const response = await fetch('/api/education/projects', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        ...createForm,
        group_id: selectedClassId.value
      })
    })
    
    if (response.ok) {
      ElMessage.success('课题创建成功')
      createDialogVisible.value = false
      loadProjects()
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '创建课题失败')
    }
  } catch (error) {
    console.error('创建课题失败:', error)
    ElMessage.error('创建课题失败')
  } finally {
    createLoading.value = false
  }
}

// 查看课题详情
const viewProjectDetails = async (project) => {
  selectedProject.value = project
  detailDialogVisible.value = true
  activeTab.value = 'info'
  
  // 加载相关数据
  await loadProjectMembers(project.id)
  await loadProjectDocuments(project.id)
  await loadMilestones(project.id)
}

// 加载课题成员
const loadProjectMembers = async (projectId) => {
  try {
    const response = await fetch(`/api/education/projects/${projectId}/members`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      projectMembers.value = data.data || []
    }
  } catch (error) {
    console.error('加载课题成员失败:', error)
  }
}

// 加载课题文档
const loadProjectDocuments = async (projectId) => {
  try {
    const response = await fetch(`/api/education/projects/${projectId}/documents`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      projectDocuments.value = data.data || []
    }
  } catch (error) {
    console.error('加载课题文档失败:', error)
  }
}

// 加载里程碑
const loadMilestones = async (projectId) => {
  try {
    const response = await fetch(`/api/education/projects/${projectId}/milestones`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      milestones.value = data.data || []
    }
  } catch (error) {
    console.error('加载里程碑失败:', error)
  }
}

// 工具函数
const getProjectStatusType = (project) => {
  if (!project.due_date) return 'info'
  
  const now = new Date()
  const dueDate = new Date(project.due_date)
  
  if (now > dueDate) {
    return 'danger'
  } else if (project.progress >= 100) {
    return 'success'
  } else if (now > new Date(dueDate.getTime() - 7 * 24 * 60 * 60 * 1000)) {
    return 'warning'
  } else {
    return 'primary'
  }
}

const getProjectStatus = (project) => {
  if (project.progress >= 100) return '已完成'
  if (!project.due_date) return '进行中'
  
  const now = new Date()
  const dueDate = new Date(project.due_date)
  
  if (now > dueDate) {
    return '已过期'
  } else if (now > new Date(dueDate.getTime() - 7 * 24 * 60 * 60 * 1000)) {
    return '即将到期'
  } else {
    return '进行中'
  }
}

const getProjectTypeText = (type) => {
  const typeMap = {
    individual: '个人项目',
    group: '小组项目',
    research: '研究课题',
    practice: '实践项目'
  }
  return typeMap[type] || type
}

const getMemberRoleTagType = (role) => {
  switch (role) {
    case 'leader':
      return 'danger'
    case 'member':
      return 'primary'
    default:
      return 'info'
  }
}

const getMemberRoleText = (role) => {
  switch (role) {
    case 'leader':
      return '组长'
    case 'member':
      return '成员'
    default:
      return '未知'
  }
}

const getMilestoneStatusType = (status) => {
  switch (status) {
    case 'completed':
      return 'success'
    case 'in_progress':
      return 'primary'
    case 'pending':
      return 'info'
    case 'overdue':
      return 'danger'
    default:
      return 'info'
  }
}

const getMilestoneStatusText = (status) => {
  switch (status) {
    case 'completed':
      return '已完成'
    case 'in_progress':
      return '进行中'
    case 'pending':
      return '待开始'
    case 'overdue':
      return '已过期'
    default:
      return '未知'
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('zh-CN')
}

// 其他功能函数
const viewProgress = (project) => {
  viewProjectDetails(project)
  activeTab.value = 'progress'
}

const manageMembers = (project) => {
  viewProjectDetails(project)
  activeTab.value = 'members'
}

const showAssignMemberDialog = () => {
  assignMemberDialogVisible.value = true
  assignMemberForm.studentIds = []
  loadAvailableStudents()
}

const assignMembers = async () => {
  if (!assignMemberFormRef.value) return
  
  const valid = await assignMemberFormRef.value.validate()
  if (!valid) return
  
  assignLoading.value = true
  
  try {
    const response = await fetch(`/api/education/projects/${selectedProject.value.id}/members`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        student_ids: assignMemberForm.studentIds
      })
    })
    
    if (response.ok) {
      ElMessage.success('成员分配成功')
      assignMemberDialogVisible.value = false
      loadProjectMembers(selectedProject.value.id)
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '分配成员失败')
    }
  } catch (error) {
    console.error('分配成员失败:', error)
    ElMessage.error('分配成员失败')
  } finally {
    assignLoading.value = false
  }
}

const showAddMilestoneDialog = () => {
  addMilestoneDialogVisible.value = true
  Object.assign(milestoneForm, {
    title: '',
    description: '',
    dueDate: null
  })
}

const addMilestone = async () => {
  if (!milestoneFormRef.value) return
  
  const valid = await milestoneFormRef.value.validate()
  if (!valid) return
  
  milestoneLoading.value = true
  
  try {
    const response = await fetch(`/api/education/projects/${selectedProject.value.id}/milestones`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(milestoneForm)
    })
    
    if (response.ok) {
      ElMessage.success('里程碑添加成功')
      addMilestoneDialogVisible.value = false
      loadMilestones(selectedProject.value.id)
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '添加里程碑失败')
    }
  } catch (error) {
    console.error('添加里程碑失败:', error)
    ElMessage.error('添加里程碑失败')
  } finally {
    milestoneLoading.value = false
  }
}

const updateMilestoneStatus = async (milestone) => {
  try {
    const response = await fetch(`/api/education/milestones/${milestone.id}/complete`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      ElMessage.success('里程碑已标记为完成')
      loadMilestones(selectedProject.value.id)
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '更新里程碑状态失败')
    }
  } catch (error) {
    console.error('更新里程碑状态失败:', error)
    ElMessage.error('更新里程碑状态失败')
  }
}

const createProjectDocument = () => {
  // 跳转到文档创建页面
  router.push(`/documents/create?projectId=${selectedProject.value.id}`)
}

const openDocument = (document) => {
  // 跳转到文档查看页面
  router.push(`/documents/${document.id}`)
}

const removeMember = async (member) => {
  try {
    await ElMessageBox.confirm(
      `确定要移除成员 ${member.name} 吗？`,
      '确认移除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const response = await fetch(`/api/education/projects/${selectedProject.value.id}/members/${member.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      ElMessage.success('移除成员成功')
      loadProjectMembers(selectedProject.value.id)
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '移除成员失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('移除成员失败:', error)
      ElMessage.error('移除成员失败')
    }
  }
}
</script>

<style scoped>
.project-management {
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

.project-list {
  margin-top: 20px;
}

.project-card {
  cursor: pointer;
  transition: all 0.3s;
  height: 320px;
}

.project-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.project-header h3 {
  margin: 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.project-description {
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

.project-info {
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

.project-actions {
  text-align: center;
}

.project-details {
  margin-top: 20px;
}

.project-info-detail {
  padding: 20px 0;
}

.project-description-detail {
  white-space: pre-wrap;
  line-height: 1.6;
}

.members-management {
  padding: 20px 0;
}

.members-header {
  margin-bottom: 16px;
}

.progress-tracking {
  padding: 20px 0;
}

.progress-overview {
  margin-bottom: 30px;
}

.milestones-list h4 {
  margin-bottom: 15px;
  color: #303133;
}

.milestone-content {
  padding: 10px;
}

.milestone-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.milestone-header h4 {
  margin: 0;
  color: #303133;
}

.milestone-actions {
  margin-top: 10px;
}

.project-documents {
  padding: 20px 0;
}

.documents-header {
  margin-bottom: 16px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

@media (max-width: 768px) {
  .project-management {
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