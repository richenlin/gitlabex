<template>
  <div class="assignments-view">
    <div class="page-header">
      <div class="header-content">
        <h1>作业管理</h1>
        <p>创建、布置和管理作业</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          创建作业
        </el-button>
      </div>
    </div>

    <!-- 筛选和搜索 -->
    <div class="filters-section">
      <el-row :gutter="16">
        <el-col :xs="24" :sm="6">
          <el-select v-model="filters.status" placeholder="作业状态" clearable>
            <el-option label="全部" value="" />
            <el-option label="草稿" value="draft" />
            <el-option label="已发布" value="published" />
            <el-option label="进行中" value="ongoing" />
            <el-option label="已截止" value="closed" />
          </el-select>
        </el-col>
        <el-col :xs="24" :sm="6">
          <el-select v-model="filters.class" placeholder="选择班级" clearable>
            <el-option label="全部班级" value="" />
            <el-option label="高等数学A" value="1" />
            <el-option label="Java程序设计" value="2" />
            <el-option label="数据结构与算法" value="3" />
          </el-select>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-input
            v-model="filters.search"
            placeholder="搜索作业名称"
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
      </el-row>
    </div>

    <!-- 作业列表 -->
    <div class="assignments-list">
      <el-card
        v-for="assignment in filteredAssignments"
        :key="assignment.id"
        class="assignment-card"
        shadow="hover"
      >
        <template #header>
          <div class="assignment-header">
            <div class="assignment-title">
              <h3>{{ assignment.title }}</h3>
              <el-tag :type="getStatusType(assignment.status)">
                {{ getStatusText(assignment.status) }}
              </el-tag>
            </div>
            <div class="assignment-actions">
              <el-button size="small" @click="viewAssignment(assignment)">
                查看详情
              </el-button>
              <el-dropdown @command="handleAction">
                <el-button size="small">
                  更多<el-icon><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item :command="`edit:${assignment.id}`">
                      编辑作业
                    </el-dropdown-item>
                    <el-dropdown-item :command="`submissions:${assignment.id}`">
                      查看提交
                    </el-dropdown-item>
                    <el-dropdown-item :command="`grade:${assignment.id}`">
                      批改作业
                    </el-dropdown-item>
                    <el-dropdown-item :command="`delete:${assignment.id}`" divided>
                      删除作业
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </template>
        
        <div class="assignment-content">
          <p class="assignment-description">{{ assignment.description }}</p>
          
          <div class="assignment-meta">
            <div class="meta-row">
              <div class="meta-item">
                <el-icon><School /></el-icon>
                <span>{{ assignment.className }}</span>
              </div>
              <div class="meta-item">
                <el-icon><Calendar /></el-icon>
                <span>截止时间：{{ formatDate(assignment.dueDate) }}</span>
              </div>
            </div>
            
            <div class="meta-row">
              <div class="meta-item">
                <el-icon><User /></el-icon>
                <span>{{ assignment.submittedCount }}/{{ assignment.totalStudents }} 已提交</span>
              </div>
              <div class="meta-item">
                <el-icon><Document /></el-icon>
                <span>{{ assignment.gradedCount }} 已批改</span>
              </div>
            </div>
          </div>

          <!-- 进度条 -->
          <div class="assignment-progress">
            <div class="progress-label">提交进度</div>
            <el-progress 
              :percentage="getSubmissionPercentage(assignment)" 
              :color="getProgressColor(assignment)"
            />
          </div>
        </div>
      </el-card>

      <!-- 空状态 -->
      <div v-if="filteredAssignments.length === 0" class="empty-state">
        <el-empty description="暂无作业">
          <el-button type="primary" @click="showCreateDialog = true">
            创建第一个作业
          </el-button>
        </el-empty>
      </div>
    </div>

    <!-- 创建作业对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建作业"
      width="800px"
      @close="resetForm"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="120px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="作业标题" prop="title">
              <el-input
                v-model="createForm.title"
                placeholder="请输入作业标题"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属班级" prop="classId">
              <el-select v-model="createForm.classId" placeholder="选择班级">
                <el-option label="高等数学A" value="1" />
                <el-option label="Java程序设计" value="2" />
                <el-option label="数据结构与算法" value="3" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="作业描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            placeholder="请输入作业描述和要求"
            :rows="4"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="截止时间" prop="dueDate">
              <el-date-picker
                v-model="createForm.dueDate"
                type="datetime"
                placeholder="选择截止时间"
                format="YYYY-MM-DD HH:mm"
                value-format="YYYY-MM-DD HH:mm:ss"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="总分" prop="totalScore">
              <el-input-number
                v-model="createForm.totalScore"
                :min="1"
                :max="1000"
                placeholder="作业总分"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="提交要求">
          <el-checkbox-group v-model="createForm.requirements">
            <el-checkbox value="code">代码文件</el-checkbox>
            <el-checkbox value="report">实验报告</el-checkbox>
            <el-checkbox value="presentation">演示文档</el-checkbox>
            <el-checkbox value="video">视频演示</el-checkbox>
          </el-checkbox-group>
        </el-form-item>

        <el-form-item label="发布选项">
          <el-radio-group v-model="createForm.publishOption">
            <el-radio value="draft">保存为草稿</el-radio>
            <el-radio value="schedule">定时发布</el-radio>
            <el-radio value="immediate">立即发布</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item 
          v-if="createForm.publishOption === 'schedule'" 
          label="发布时间"
        >
          <el-date-picker
            v-model="createForm.publishTime"
            type="datetime"
            placeholder="选择发布时间"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createAssignment" :loading="creating">
          {{ createForm.publishOption === 'immediate' ? '创建并发布' : '创建作业' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { 
  Plus, 
  Search, 
  ArrowDown, 
  School, 
  Calendar, 
  User, 
  Document 
} from '@element-plus/icons-vue'

// 类型定义
interface Assignment {
  id: number
  title: string
  description: string
  className: string
  classId: string
  status: 'draft' | 'published' | 'ongoing' | 'closed'
  dueDate: string
  totalScore: number
  submittedCount: number
  gradedCount: number
  totalStudents: number
  createdAt: string
}

// 响应式数据
const assignments = ref<Assignment[]>([])
const showCreateDialog = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()

// 筛选条件
const filters = reactive({
  status: '',
  class: '',
  search: ''
})

// 创建表单
const createForm = reactive({
  title: '',
  description: '',
  classId: '',
  dueDate: '',
  totalScore: 100,
  requirements: [] as string[],
  publishOption: 'draft',
  publishTime: ''
})

// 表单验证规则
const createRules: FormRules = {
  title: [
    { required: true, message: '请输入作业标题', trigger: 'blur' },
    { min: 2, max: 100, message: '长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  classId: [
    { required: true, message: '请选择班级', trigger: 'change' }
  ],
  description: [
    { required: true, message: '请输入作业描述', trigger: 'blur' },
    { max: 500, message: '描述不能超过 500 个字符', trigger: 'blur' }
  ],
  dueDate: [
    { required: true, message: '请选择截止时间', trigger: 'change' }
  ]
}

// 计算属性 - 过滤后的作业列表
const filteredAssignments = computed(() => {
  let result = assignments.value

  if (filters.status) {
    result = result.filter(a => a.status === filters.status)
  }

  if (filters.class) {
    result = result.filter(a => a.classId === filters.class)
  }

  if (filters.search) {
    result = result.filter(a => 
      a.title.toLowerCase().includes(filters.search.toLowerCase())
    )
  }

  return result
})

// 组件挂载时加载数据
onMounted(() => {
  loadAssignments()
})

// 加载作业列表
const loadAssignments = () => {
  // 模拟数据
  assignments.value = [
    {
      id: 1,
      title: '第一章 函数与极限 练习题',
      description: '完成教材第一章课后习题1-20题，需要详细解题步骤',
      className: '高等数学A',
      classId: '1',
      status: 'ongoing',
      dueDate: '2024-04-15 23:59:59',
      totalScore: 100,
      submittedCount: 32,
      gradedCount: 15,
      totalStudents: 45,
      createdAt: '2024-03-01'
    },
    {
      id: 2,
      title: 'Java基础语法编程作业',
      description: '使用Java实现学生信息管理系统，包含增删改查功能',
      className: 'Java程序设计',
      classId: '2',
      status: 'published',
      dueDate: '2024-04-20 23:59:59',
      totalScore: 150,
      submittedCount: 28,
      gradedCount: 28,
      totalStudents: 38,
      createdAt: '2024-03-10'
    },
    {
      id: 3,
      title: '排序算法实现与分析',
      description: '实现快速排序、归并排序、堆排序算法，并进行时间复杂度分析',
      className: '数据结构与算法',
      classId: '3',
      status: 'draft',
      dueDate: '2024-05-01 23:59:59',
      totalScore: 200,
      submittedCount: 0,
      gradedCount: 0,
      totalStudents: 42,
      createdAt: '2024-03-20'
    }
  ]
}

// 创建作业
const createAssignment = async () => {
  if (!createFormRef.value) return

  try {
    const isValid = await createFormRef.value.validate()
    if (!isValid) return

    creating.value = true

    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))

    const newAssignment: Assignment = {
      id: Date.now(),
      title: createForm.title,
      description: createForm.description,
      className: getClassName(createForm.classId),
      classId: createForm.classId,
      status: createForm.publishOption === 'immediate' ? 'published' : 'draft',
      dueDate: createForm.dueDate,
      totalScore: createForm.totalScore,
      submittedCount: 0,
      gradedCount: 0,
      totalStudents: getTotalStudents(createForm.classId),
      createdAt: new Date().toISOString().split('T')[0]
    }

    assignments.value.unshift(newAssignment)
    showCreateDialog.value = false
    
    const message = createForm.publishOption === 'immediate' 
      ? '作业创建并发布成功' 
      : '作业创建成功'
    ElMessage.success(message)
  } catch (error) {
    ElMessage.error('创建作业失败')
  } finally {
    creating.value = false
  }
}

// 获取班级名称
const getClassName = (classId: string) => {
  const classMap: Record<string, string> = {
    '1': '高等数学A',
    '2': 'Java程序设计',
    '3': '数据结构与算法'
  }
  return classMap[classId] || '未知班级'
}

// 获取班级学生总数
const getTotalStudents = (classId: string) => {
  const studentCountMap: Record<string, number> = {
    '1': 45,
    '2': 38,
    '3': 42
  }
  return studentCountMap[classId] || 0
}

// 查看作业详情
const viewAssignment = (assignment: Assignment) => {
  ElMessage.info(`查看作业: ${assignment.title}`)
  // TODO: 路由到作业详情页
}

// 处理操作
const handleAction = (command: string) => {
  const [action, id] = command.split(':')
  const assignment = assignments.value.find(a => a.id === Number(id))
  
  if (!assignment) return

  switch (action) {
    case 'edit':
      ElMessage.info(`编辑作业: ${assignment.title}`)
      break
    case 'submissions':
      ElMessage.info(`查看 ${assignment.title} 的提交情况`)
      break
    case 'grade':
      ElMessage.info(`批改 ${assignment.title}`)
      break
    case 'delete':
      handleDelete(assignment)
      break
  }
}

// 删除作业
const handleDelete = async (assignment: Assignment) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除作业"${assignment.title}"吗？此操作不可撤销。`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    const index = assignments.value.findIndex(a => a.id === assignment.id)
    if (index > -1) {
      assignments.value.splice(index, 1)
      ElMessage.success('作业删除成功')
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
    classId: '',
    dueDate: '',
    totalScore: 100,
    requirements: [],
    publishOption: 'draft',
    publishTime: ''
  })
}

// 获取状态类型
const getStatusType = (status: string) => {
  const statusMap = {
    draft: 'info',
    published: 'warning', 
    ongoing: 'success',
    closed: 'danger'
  }
  return statusMap[status as keyof typeof statusMap] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap = {
    draft: '草稿',
    published: '已发布',
    ongoing: '进行中',
    closed: '已截止'
  }
  return statusMap[status as keyof typeof statusMap] || '未知'
}

// 获取提交百分比
const getSubmissionPercentage = (assignment: Assignment) => {
  if (assignment.totalStudents === 0) return 0
  return Math.round((assignment.submittedCount / assignment.totalStudents) * 100)
}

// 获取进度条颜色
const getProgressColor = (assignment: Assignment) => {
  const percentage = getSubmissionPercentage(assignment)
  if (percentage >= 80) return '#67c23a'
  if (percentage >= 60) return '#e6a23c'
  return '#f56c6c'
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.assignments-view {
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

.assignments-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.assignment-card {
  transition: all 0.3s ease;
}

.assignment-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.assignment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.assignment-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.assignment-title h3 {
  margin: 0;
  font-size: 18px;
  color: #303133;
}

.assignment-actions {
  display: flex;
  gap: 8px;
}

.assignment-content {
  padding-top: 8px;
}

.assignment-description {
  color: #606266;
  margin: 0 0 16px 0;
  line-height: 1.5;
}

.assignment-meta {
  margin-bottom: 16px;
}

.meta-row {
  display: flex;
  gap: 24px;
  margin-bottom: 8px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #909399;
  font-size: 14px;
}

.assignment-progress {
  margin-top: 16px;
}

.progress-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

@media (max-width: 768px) {
  .assignments-view {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .assignment-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .meta-row {
    flex-direction: column;
    gap: 8px;
  }
}
</style> 