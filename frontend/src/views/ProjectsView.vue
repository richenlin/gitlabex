<template>
  <div class="projects-view">
    <div class="page-header">
      <div class="header-content">
        <h1>课题管理</h1>
        <p>创建和管理研究课题、项目</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog = true">
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
              <el-dropdown @command="handleAction">
                <el-button size="small" circle>
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item :command="`edit:${project.id}`">
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
                    <el-dropdown-item :command="`delete:${project.id}`" divided>
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
            <div class="meta-row">
              <div class="meta-item">
                <el-icon><User /></el-icon>
                <span>导师：{{ project.supervisor }}</span>
              </div>
                             <div class="meta-item">
                 <el-icon><UserFilled /></el-icon>
                 <span>{{ project.memberCount }} 名成员</span>
               </div>
            </div>
            
            <div class="meta-row">
              <div class="meta-item">
                <el-icon><Calendar /></el-icon>
                <span>开始：{{ formatDate(project.startDate) }}</span>
              </div>
              <div class="meta-item">
                <el-icon><Clock /></el-icon>
                <span>预期：{{ formatDate(project.expectedEndDate) }}</span>
              </div>
            </div>
          </div>

          <!-- 进度条 -->
          <div class="project-progress">
            <div class="progress-info">
              <span class="progress-label">作业完成进度</span>
              <span class="progress-percentage">{{ project.completedAssignments }}/{{ project.totalAssignments }} ({{ project.progress }}%)</span>
            </div>
            <el-progress 
              :percentage="project.progress" 
              :color="getProgressColor(project.progress)"
              :stroke-width="8"
            />
            <div class="progress-detail">
              <small>基于学生作业提交并通过教师评审的比例</small>
            </div>
          </div>

          <!-- 里程碑 -->
          <div class="project-milestones" v-if="project.milestones.length > 0">
            <div class="milestones-title">近期里程碑</div>
            <div class="milestones-list">
              <div 
                v-for="milestone in project.milestones.slice(0, 2)"
                :key="milestone.id"
                class="milestone-item"
                :class="{ 'completed': milestone.completed }"
              >
                <el-icon>
                  <Check v-if="milestone.completed" />
                  <Clock v-else />
                </el-icon>
                <span>{{ milestone.title }}</span>
                <span class="milestone-date">{{ formatDate(milestone.dueDate) }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 空状态 -->
      <div v-if="filteredProjects.length === 0" class="empty-state">
        <el-empty description="暂无课题">
          <el-button type="primary" @click="showCreateDialog = true">
            创建第一个课题
          </el-button>
        </el-empty>
      </div>
    </div>

    <!-- 创建课题对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建课题"
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
          创建课题
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { ApiService } from '@/services/api'
import { 
  Plus, 
  Search, 
  MoreFilled, 
  User,
  UserFilled,
  Calendar, 
  Clock,
  Check
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
}

// 路由
const router = useRouter()

// 响应式数据
const projects = ref<Project[]>([])
const showCreateDialog = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()

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

// 计算属性 - 过滤后的课题列表
const filteredProjects = computed(() => {
  let result = projects.value

  if (filters.status) {
    result = result.filter(p => p.status === filters.status)
  }

  if (filters.type) {
    result = result.filter(p => p.type === filters.type)
  }

  if (filters.search) {
    const searchTerm = filters.search.toLowerCase()
    result = result.filter(p => 
      p.title.toLowerCase().includes(searchTerm) ||
      p.description.toLowerCase().includes(searchTerm)
    )
  }

  return result
})

// 组件挂载时加载数据
onMounted(() => {
  loadProjects()
})

// 加载课题列表
const loadProjects = async () => {
  try {
    const response = await ApiService.getProjects()
    projects.value = response.data || []
  } catch (error) {
    console.error('加载课题列表失败:', error)
    // 使用模拟数据作为备用
    projects.value = [
    {
      id: 1,
      title: '基于Vue3的在线教育平台设计与实现',
      description: '设计并实现一个现代化的在线教育平台，支持视频播放、在线测试、作业管理等功能',
      type: 'graduation',
      status: 'ongoing',
      supervisor: '张教授',
      memberCount: 3,
      maxMembers: 4,
      totalAssignments: 6,
      completedAssignments: 4,
      progress: Math.round((4 / 6) * 100), // 基于作业完成率计算
      startDate: '2024-02-01',
      expectedEndDate: '2024-06-01',
      milestones: [
        { id: 1, title: '需求分析完成', dueDate: '2024-02-15', completed: true },
        { id: 2, title: '系统设计完成', dueDate: '2024-03-15', completed: true },
        { id: 3, title: '前端开发完成', dueDate: '2024-04-15', completed: false }
      ],
      createdAt: '2024-01-15'
    },
    {
      id: 2,
      title: '机器学习在图像识别中的应用研究',
      description: '研究深度学习算法在图像分类和目标检测中的应用，提出改进方案',
      type: 'research',
      status: 'ongoing',
      supervisor: '李研究员',
      memberCount: 2,
      maxMembers: 3,
      totalAssignments: 4,
      completedAssignments: 2,
      progress: Math.round((2 / 4) * 100), // 基于作业完成率计算
      startDate: '2024-01-01',
      expectedEndDate: '2024-12-01',
      milestones: [
        { id: 4, title: '文献调研完成', dueDate: '2024-01-31', completed: true },
        { id: 5, title: '数据集准备完成', dueDate: '2024-02-28', completed: true },
        { id: 6, title: '模型训练完成', dueDate: '2024-04-30', completed: false }
      ],
      createdAt: '2023-12-15'
    },
    {
      id: 3,
      title: '智能物流管理系统',
      description: '为参加全国大学生软件设计大赛开发的智能物流管理系统，包含路径优化、库存管理等功能',
      type: 'competition',
      status: 'planning',
      supervisor: '王老师',
      memberCount: 4,
      maxMembers: 5,
      totalAssignments: 3,
      completedAssignments: 1,
      progress: Math.round((1 / 3) * 100), // 基于作业完成率计算
      startDate: '2024-03-01',
      expectedEndDate: '2024-08-01',
      milestones: [
        { id: 7, title: '团队组建完成', dueDate: '2024-03-10', completed: true },
        { id: 8, title: '技术选型完成', dueDate: '2024-03-20', completed: false }
      ],
      createdAt: '2024-02-20'
    }
  ]
  }
}

// 创建课题
const createProject = async () => {
  if (!createFormRef.value) return

  try {
    const isValid = await createFormRef.value.validate()
    if (!isValid) return

    creating.value = true

    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))

    const newProject: Project = {
      id: Date.now(),
      title: createForm.title,
      description: createForm.description,
      type: createForm.type as any,
      status: 'planning',
      supervisor: createForm.supervisor,
      memberCount: 1, // 创建者
      maxMembers: createForm.maxMembers,
      totalAssignments: 0,
      completedAssignments: 0,
      progress: 0,
      startDate: createForm.startDate,
      expectedEndDate: createForm.expectedEndDate,
      milestones: [],
      createdAt: new Date().toISOString().split('T')[0]
    }

    projects.value.unshift(newProject)
    showCreateDialog.value = false
    ElMessage.success('课题创建成功')
  } catch (error) {
    ElMessage.error('创建课题失败')
  } finally {
    creating.value = false
  }
}

// 查看课题详情
const viewProject = (project: Project) => {
  router.push(`/projects/${project.id}`)
}

// 处理操作
const handleAction = (command: string) => {
  const [action, id] = command.split(':')
  const project = projects.value.find(p => p.id === Number(id))
  
  if (!project) return

  switch (action) {
    case 'edit':
      ElMessage.info(`编辑课题: ${project.title}`)
      break
    case 'members':
      ElMessage.info(`管理 ${project.title} 的成员`)
      break
    case 'progress':
      ElMessage.info(`管理 ${project.title} 的进度`)
      break
    case 'files':
      ElMessage.info(`管理 ${project.title} 的文档`)
      break
    case 'delete':
      handleDelete(project)
      break
  }
}

// 删除课题
const handleDelete = async (project: Project) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除课题"${project.title}"吗？此操作不可撤销。`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    const index = projects.value.findIndex(p => p.id === project.id)
    if (index > -1) {
      projects.value.splice(index, 1)
      ElMessage.success('课题删除成功')
    }
  } catch {
    // 用户取消删除
  }
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

.meta-row {
  display: flex;
  gap: 20px;
  margin-bottom: 8px;
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

.progress-info {
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

.empty-state {
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