<template>
  <div class="projects-view">
    <div class="page-header">
      <div class="header-content">
        <h1>课题管理</h1>
        <p>创建和管理研究课题、项目</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog = true" v-if="canCreateProject">
          <el-icon><Plus /></el-icon>
          创建课题
        </el-button>
      </div>
    </div>

    <!-- 筛选和搜索 -->
    <div class="filters-section">
      <el-row :gutter="16">
        <el-col :xs="24" :sm="6">
          <el-select v-model="filters.status" placeholder="课题状态" clearable>
            <el-option label="全部" value="" />
            <el-option label="规划中" value="planning" />
            <el-option label="进行中" value="ongoing" />
            <el-option label="已完成" value="completed" />
            <el-option label="已暂停" value="paused" />
          </el-select>
        </el-col>
        <el-col :xs="24" :sm="6">
          <el-select v-model="filters.type" placeholder="课题类型" clearable>
            <el-option label="全部类型" value="" />
            <el-option label="毕业设计" value="graduation" />
            <el-option label="科研项目" value="research" />
            <el-option label="竞赛项目" value="competition" />
            <el-option label="实践项目" value="practice" />
          </el-select>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-input
            v-model="filters.search"
            placeholder="搜索课题名称或描述"
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
      </el-row>
    </div>

    <!-- 课题列表 -->
    <div class="projects-grid">
      <div v-if="projects.length === 0" class="loading-state">
        <el-empty description="暂无课题数据">
          <el-button type="primary" @click="loadProjects">重新加载</el-button>
        </el-empty>
      </div>
      
      <el-card
        v-for="project in filteredProjects"
        :key="project.id"
        class="project-card"
        shadow="hover"
        @click="viewProject(project)"
      >
        <template #header>
          <div class="project-header">
            <div class="project-title">
              <h3>{{ project.title }}</h3>
              <div class="project-tags">
                <el-tag :type="getStatusType(project.status)" size="small">
                  {{ getStatusText(project.status) }}
                </el-tag>
                <el-tag :type="getTypeColor(project.type)" size="small">
                  {{ getTypeText(project.type) }}
                </el-tag>
              </div>
            </div>
            <div class="project-actions" @click.stop>
              <el-dropdown @command="handleAction" trigger="click">
                <el-button size="small" circle>
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item :command="`edit:${project.id}`" v-if="canEditProject">
                      编辑课题
                    </el-dropdown-item>
                    <el-dropdown-item :command="`members:${project.id}`">
                      管理成员
                    </el-dropdown-item>
                    <el-dropdown-item :command="`progress:${project.id}`">
                      进度管理
                    </el-dropdown-item>
                    <el-dropdown-item :command="`files:${project.id}`">
                      文档管理
                    </el-dropdown-item>
                    <el-dropdown-item :command="`delete:${project.id}`" divided v-if="canDeleteProject">
                      删除课题
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </template>

        <div class="project-content">
          <p class="project-description">{{ project.description }}</p>
          
          <div class="project-meta">
            <div class="meta-item">
              <el-icon><User /></el-icon>
              <span>指导老师：{{ project.supervisor }}</span>
            </div>
            <div class="meta-item">
              <el-icon><User /></el-icon>
              <span>成员：{{ project.memberCount }}/{{ project.maxMembers }}</span>
            </div>
            <div class="meta-item">
              <el-icon><FolderOpened /></el-icon>
              <span>班级：{{ project.className || '未分配' }}</span>
            </div>
          </div>

          <div class="project-progress">
            <div class="progress-header">
              <span>作业进度</span>
              <span>{{ project.completedAssignments }}/{{ project.totalAssignments }}</span>
            </div>
            <el-progress 
              :percentage="project.progress" 
              :color="getProgressColor(project.progress)"
              :show-text="false"
            />
          </div>

          <div class="project-dates">
            <div class="date-item">
              <el-icon><Calendar /></el-icon>
              <span>开始：{{ formatDate(project.startDate) }}</span>
            </div>
            <div class="date-item">
              <el-icon><Calendar /></el-icon>
              <span>结束：{{ formatDate(project.expectedEndDate) }}</span>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 创建课题对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingProject ? '编辑课题' : '创建课题'"
      width="900px"
      @close="resetForm"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="120px"
      >
        <el-row :gutter="20">
          <el-col :span="16">
            <el-form-item label="课题标题" prop="title">
              <el-input
                v-model="createForm.title"
                placeholder="请输入课题标题"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="课题类型" prop="type">
              <el-select v-model="createForm.type" placeholder="选择课题类型">
                <el-option label="毕业设计" value="graduation" />
                <el-option label="科研项目" value="research" />
                <el-option label="竞赛项目" value="competition" />
                <el-option label="实践项目" value="practice" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="课题描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            placeholder="详细描述课题的目标、要求和预期成果"
            :rows="4"
            maxlength="1000"
            show-word-limit
          />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="指导教师" prop="supervisor">
              <el-input
                v-model="createForm.supervisor"
                placeholder="请输入指导教师姓名"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大成员数">
              <el-input-number
                v-model="createForm.maxMembers"
                :min="1"
                :max="20"
                placeholder="课题最大成员数"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="开始时间" prop="startDate">
              <el-date-picker
                v-model="createForm.startDate"
                type="date"
                placeholder="选择开始时间"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="预期结束时间" prop="expectedEndDate">
              <el-date-picker
                v-model="createForm.expectedEndDate"
                type="date"
                placeholder="选择预期结束时间"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="技术要求">
          <el-checkbox-group v-model="createForm.techRequirements">
            <el-checkbox value="frontend">前端开发</el-checkbox>
            <el-checkbox value="backend">后端开发</el-checkbox>
            <el-checkbox value="mobile">移动开发</el-checkbox>
            <el-checkbox value="ai">人工智能</el-checkbox>
            <el-checkbox value="database">数据库设计</el-checkbox>
            <el-checkbox value="ui">UI/UX设计</el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item label="评估标准">
          <el-input
            v-model="createForm.evaluationCriteria"
            type="textarea"
            placeholder="描述课题的评估标准和要求"
            :rows="3"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createProject" :loading="creating">
          {{ editingProject ? '更新课题' : '创建课题' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { ApiService } from '@/services/api'
import { 
  Plus, 
  Search, 
  MoreFilled, 
  User,
  UserFilled,
  Calendar, 
  Clock,
  Check,
  FolderOpened
} from '@element-plus/icons-vue'

// 类型定义
interface Milestone {
  id: number
  title: string
  dueDate: string
  completed: boolean
}

interface Project {
  id: number
  title: string
  description: string
  type: 'graduation' | 'research' | 'competition' | 'practice'
  status: 'planning' | 'ongoing' | 'completed' | 'paused'
  supervisor: string
  memberCount: number
  maxMembers: number
  totalAssignments: number
  completedAssignments: number
  progress: number
  startDate: string
  expectedEndDate: string
  milestones: Milestone[]
  createdAt: string
  classId: number
  className: string
}

// 路由
const router = useRouter()
const authStore = useAuthStore()

// 响应式数据
const projects = ref<Project[]>([])
const showCreateDialog = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()
const editingProject = ref<Project | null>(null)

// 权限控制计算属性
const isAdmin = computed(() => authStore.userRole === 1)
const isTeacher = computed(() => authStore.userRole === 2)
const isStudent = computed(() => authStore.userRole === 3)
const canCreateProject = computed(() => isAdmin.value || isTeacher.value)
const canEditProject = computed(() => isAdmin.value || isTeacher.value)
const canDeleteProject = computed(() => isAdmin.value || isTeacher.value)

// 筛选条件
const filters = reactive({
  status: '',
  type: '',
  search: ''
})

// 创建表单
const createForm = reactive({
  title: '',
  description: '',
  type: '',
  supervisor: '',
  maxMembers: 5,
  startDate: '',
  expectedEndDate: '',
  techRequirements: [] as string[],
  evaluationCriteria: ''
})

// 表单验证规则
const createRules: FormRules = {
  title: [
    { required: true, message: '请输入课题标题', trigger: 'blur' },
    { min: 5, max: 100, message: '长度在 5 到 100 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择课题类型', trigger: 'change' }
  ],
  description: [
    { required: true, message: '请输入课题描述', trigger: 'blur' },
    { min: 20, max: 1000, message: '长度在 20 到 1000 个字符', trigger: 'blur' }
  ],
  supervisor: [
    { required: true, message: '请输入指导教师', trigger: 'blur' }
  ],
  startDate: [
    { required: true, message: '请选择开始时间', trigger: 'change' }
  ],
  expectedEndDate: [
    { required: true, message: '请选择预期结束时间', trigger: 'change' }
  ]
}

// 筛选后的课题列表
const filteredProjects = computed(() => {
  console.log('=== filteredProjects 计算 ===')
  console.log('原始 projects.value:', projects.value)
  console.log('filters:', filters)
  
  let result = projects.value
  
  // 状态筛选
  if (filters.status) {
    result = result.filter(project => project.status === filters.status)
    console.log('状态筛选后:', result.length, '个项目')
  }
  
  // 类型筛选
  if (filters.type) {
    result = result.filter(project => project.type === filters.type)
    console.log('类型筛选后:', result.length, '个项目')
  }
  
  // 搜索筛选
  if (filters.search.trim()) {
    const searchTerm = filters.search.toLowerCase()
    result = result.filter(project => 
      project.title.toLowerCase().includes(searchTerm) ||
      project.description.toLowerCase().includes(searchTerm) ||
      project.supervisor.toLowerCase().includes(searchTerm)
    )
    console.log('搜索筛选后:', result.length, '个项目')
  }
  
  console.log('最终筛选结果:', result)
  console.log('=== filteredProjects 计算结束 ===')
  return result
})

// 组件挂载时加载数据
onMounted(() => {
  console.log('ProjectsView mounted, loading projects...')
  console.log('Auth store:', authStore.user, 'Role:', authStore.userRole)
  loadProjects()
})

// 加载课题列表
const loadProjects = async () => {
  console.log('=== loadProjects 开始 ===')
  console.log('当前用户角色:', authStore.userRole)
  console.log('权限检查 - canCreateProject:', canCreateProject.value)
  
  try {
    console.log('调用 ApiService.getProjects()...')
    const response = await ApiService.getProjects()
    console.log('=== API响应详情 ===')
    console.log('Response type:', typeof response)
    console.log('Response:', response)
    console.log('Response.data type:', typeof response?.data)
    console.log('Response.data:', response?.data)
    
    // 处理不同的响应格式
    let projectsData: any[] = []
    
    if (Array.isArray(response)) {
      // 后端直接返回数组
      console.log('处理直接数组响应')
      projectsData = response
    } else if (response && response.data && Array.isArray(response.data)) {
      // 后端返回包装对象 { data: [...], total: number }
      console.log('处理包装对象响应')
      projectsData = response.data
    } else {
      console.warn('未知的API响应格式:', response)
      projectsData = []
    }
    
    if (projectsData.length > 0) {
      console.log('开始处理', projectsData.length, '个项目...')
      
      projects.value = projectsData.map((project: any, index: number) => {
        console.log(`处理项目 ${index + 1}:`, project)
        
        const mappedProject = {
          id: project.id,
          title: project.name, // 后端字段是name，前端期望title
          description: project.description,
          type: project.type,
          status: project.status,
          supervisor: project.teacher_name || '未知导师',
          memberCount: 1, // 简化数据中没有成员信息，使用默认值
          maxMembers: project.max_members,
          totalAssignments: project.total_assignments,
          completedAssignments: project.completed_assignments,
          progress: project.total_assignments > 0 
            ? Math.round((project.completed_assignments / project.total_assignments) * 100) 
            : 0,
          startDate: project.start_date,
          expectedEndDate: project.end_date,
          milestones: [], // 简化数据中没有milestones，使用空数组
          createdAt: project.created_at,
          classId: project.class_id,
          className: project.class_name
        }
        
        console.log(`映射后的项目 ${index + 1}:`, mappedProject)
        return mappedProject
      })
      
      console.log('=== 最终项目列表 ===')
      console.log('Projects count:', projects.value.length)
      console.log('Projects:', projects.value)
      
      ElMessage.success(`成功加载 ${projects.value.length} 个课题`)
    } else {
      console.log('没有找到项目数据')
      projects.value = []
      ElMessage.warning('未获取到课题数据')
    }
  } catch (error) {
    console.error('=== 加载课题列表失败 ===')
    console.error('Error:', error)
    console.error('Error details:', error instanceof Error ? error.message : 'Unknown error')
    ElMessage.error('加载课题列表失败: ' + (error instanceof Error ? error.message : '未知错误'))
    // 当API失败时，不使用模拟数据，让用户知道真实状态
    projects.value = []
  }
  
  console.log('=== loadProjects 结束 ===')
  console.log('Final projects.value:', projects.value)
}

// 课题操作处理
const handleAction = (command: string) => {
  const [action, projectId] = command.split(':')
  const project = projects.value.find(p => p.id === parseInt(projectId))
  
  switch (action) {
    case 'edit':
      editProject(project)
      break
    case 'members':
      manageMembers(project)
      break
    case 'progress':
      manageProgress(project)
      break
    case 'files':
      manageFiles(project)
      break
    case 'delete':
      deleteProject(project)
      break
    default:
      console.warn('未知操作:', action)
  }
}

// 查看课题详情
const viewProject = (project: any) => {
  router.push(`/projects/${project.id}`)
}

// 编辑课题
const editProject = (project: any) => {
  // 检查权限：只有管理员和教师可以编辑课题
  if (!canEditProject.value) {
    ElMessage.warning('您没有权限编辑课题')
    return
  }
  
  // 检查是否是课题的创建者（教师只能编辑自己创建的课题）
  if (isTeacher.value && project.supervisor !== authStore.user?.name) {
    ElMessage.warning('您只能编辑自己创建的课题')
    return
  }
  
  // 填充编辑表单
  Object.assign(createForm, {
    title: project.title,
    description: project.description,
    type: project.type,
    supervisor: project.supervisor,
    maxMembers: project.maxMembers,
    startDate: project.startDate,
    expectedEndDate: project.expectedEndDate,
    techRequirements: [],
    evaluationCriteria: ''
  })
  
  editingProject.value = project
  showCreateDialog.value = true
}

// 管理成员
const manageMembers = (project: any) => {
  if (!canEditProject.value) {
    ElMessage.warning('您没有权限管理课题成员')
    return
  }
  
  ElMessage.info('课题成员管理功能开发中')
  // TODO: 实现成员管理功能
}

// 管理进度
const manageProgress = (project: any) => {
  ElMessage.info('进度管理功能开发中')
  // TODO: 实现进度管理功能
}

// 管理文档
const manageFiles = (project: any) => {
  router.push(`/projects/${project.id}/documents`)
}

// 删除课题
const deleteProject = async (project: any) => {
  if (!canDeleteProject.value) {
    ElMessage.warning('您没有权限删除课题')
    return
  }
  
  // 检查是否是课题的创建者（教师只能删除自己创建的课题）
  if (isTeacher.value && project.supervisor !== authStore.user?.name) {
    ElMessage.warning('您只能删除自己创建的课题')
    return
  }
  
  // 检查是否有作业，如果有则不能删除
  if (project.totalAssignments > 0) {
    ElMessage.warning('该课题已有作业，无法删除')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除课题 "${project.title}" 吗？此操作无法撤销。`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    // 调用删除API
    await ApiService.deleteProject(project.id)
    
    // 从列表中移除
    const index = projects.value.findIndex(p => p.id === project.id)
    if (index > -1) {
      projects.value.splice(index, 1)
      ElMessage.success('课题删除成功')
    }
  } catch (error) {
    if (error !== 'cancel') { // 不是用户取消操作
      console.error('删除课题失败:', error)
      ElMessage.error('删除课题失败')
    }
  }
}

// 创建课题
const createProject = async () => {
  if (!createFormRef.value) return
  
  await createFormRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    
    creating.value = true
    try {
      if (editingProject.value) {
        // 更新现有课题
        const updateData = {
          name: createForm.title,
          description: createForm.description,
          status: editingProject.value.status,
          start_date: createForm.startDate,
          end_date: createForm.expectedEndDate
        }
        
        const updatedProject = await ApiService.updateProject(editingProject.value.id, updateData)
        
        // 更新列表中的项目
        const index = projects.value.findIndex(p => p.id === editingProject.value!.id)
        if (index > -1) {
          projects.value[index] = {
            ...projects.value[index],
            title: updatedProject.name,
            description: updatedProject.description,
            startDate: updatedProject.start_date,
            expectedEndDate: updatedProject.end_date
          }
        }
        
        ElMessage.success('课题更新成功')
      } else {
        // 创建新课题
        const createData = {
          name: createForm.title,
          description: createForm.description,
          type: createForm.type,
          start_date: createForm.startDate,
          end_date: createForm.expectedEndDate,
          max_members: createForm.maxMembers,
          wiki_enabled: true,
          issues_enabled: true,
          mr_enabled: true
        }
        
        const newProject = await ApiService.createProject(createData)
        
        // 重新加载项目列表以获取最新数据
        await loadProjects()
        
        ElMessage.success('课题创建成功')
      }
      
      showCreateDialog.value = false
      resetForm()
    } catch (error) {
      console.error('操作课题失败:', error)
      ElMessage.error(editingProject.value ? '课题更新失败' : '课题创建失败')
    } finally {
      creating.value = false
    }
  })
}

// 搜索处理
const handleSearch = () => {
  // 搜索逻辑已在计算属性中处理
}

// 重置表单
const resetForm = () => {
  if (createFormRef.value) {
    createFormRef.value.resetFields()
  }
  Object.assign(createForm, {
    title: '',
    description: '',
    type: '',
    supervisor: '',
    maxMembers: 5,
    startDate: '',
    expectedEndDate: '',
    techRequirements: [],
    evaluationCriteria: ''
  })
  editingProject.value = null
}

// 获取状态类型
const getStatusType = (status: string) => {
  const statusMap = {
    planning: 'info',
    ongoing: 'success',
    completed: 'primary',
    paused: 'warning'
  }
  return statusMap[status as keyof typeof statusMap] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap = {
    planning: '规划中',
    ongoing: '进行中', 
    completed: '已完成',
    paused: '已暂停'
  }
  return statusMap[status as keyof typeof statusMap] || '未知'
}

// 获取类型颜色
const getTypeColor = (type: string) => {
  const typeMap = {
    graduation: 'danger',
    research: 'primary',
    competition: 'warning',
    practice: 'success'
  }
  return typeMap[type as keyof typeof typeMap] || 'info'
}

// 获取类型文本
const getTypeText = (type: string) => {
  const typeMap = {
    graduation: '毕业设计',
    research: '科研项目',
    competition: '竞赛项目',
    practice: '实践项目'
  }
  return typeMap[type as keyof typeof typeMap] || '未知'
}

// 获取进度条颜色
const getProgressColor = (progress: number) => {
  if (progress >= 80) return '#67c23a'
  if (progress >= 60) return '#409eff'
  if (progress >= 40) return '#e6a23c'
  return '#f56c6c'
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-CN', {
    month: 'short',
    day: 'numeric'
  })
}
</script>

<style scoped>
.projects-view {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-content h1 {
  font-size: 28px;
  color: #303133;
  margin: 0 0 8px 0;
}

.header-content p {
  color: #909399;
  margin: 0;
}

.filters-section {
  margin-bottom: 24px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.projects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 24px;
}

.project-card {
  cursor: pointer;
  transition: all 0.3s ease;
  height: fit-content;
}

.project-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.project-title h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: #303133;
  line-height: 1.4;
}

.project-tags {
  display: flex;
  gap: 6px;
}

.project-content {
  padding-top: 8px;
}

.project-description {
  color: #606266;
  margin: 0 0 16px 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.project-meta {
  margin-bottom: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #909399;
  font-size: 14px;
}

.project-progress {
  margin-bottom: 16px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
}

.progress-detail {
  margin-top: 4px;
  text-align: center;
}

.progress-detail small {
  color: #909399;
  font-size: 12px;
}

.progress-label {
  font-size: 14px;
  color: #606266;
}

.progress-percentage {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.project-milestones {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}

.milestones-title {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
}

.milestones-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.milestone-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #909399;
}

.milestone-item.completed {
  color: #67c23a;
}

.milestone-date {
  margin-left: auto;
  font-size: 12px;
}

.project-dates {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
  display: flex;
  justify-content: space-between;
}

.date-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #909399;
  font-size: 14px;
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
}

.loading-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
}

@media (max-width: 768px) {
  .projects-view {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .projects-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  
  .meta-row {
    flex-direction: column;
    gap: 8px;
  }
}
</style> 