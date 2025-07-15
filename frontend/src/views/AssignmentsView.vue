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
          <el-select v-model="filters.project" placeholder="选择课题" clearable>
            <el-option label="全部课题" value="" />
            <el-option 
              v-for="project in projectList" 
              :key="project.id" 
              :label="project.name" 
              :value="project.id" 
            />
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
    <div class="assignments-grid">
      <div v-if="filteredAssignments.length === 0" class="empty-state">
        <el-empty description="暂无作业数据" />
      </div>
      <div v-else class="assignment-cards">
        <el-card
          v-for="assignment in filteredAssignments"
          :key="assignment.id"
          class="assignment-card"
          shadow="hover"
        >
          <div class="assignment-header">
            <div class="assignment-title">
              <h3>{{ assignment.title }}</h3>
              <el-tag :type="getStatusType(assignment.status)">
                {{ getStatusText(assignment.status) }}
              </el-tag>
            </div>
            <div class="assignment-actions">
              <el-button size="small" @click="viewAssignment(assignment)">查看</el-button>
              <el-button size="small" type="primary" @click="editAssignment(assignment)">编辑</el-button>
              <el-button size="small" type="danger" @click="deleteAssignment(assignment.id)">删除</el-button>
            </div>
          </div>
          
          <div class="assignment-content">
            <p class="assignment-description">{{ assignment.description }}</p>
            
            <div class="assignment-meta">
              <div class="meta-item">
                <el-icon><School /></el-icon>
                <span>课题：{{ assignment.projectName }}</span>
              </div>
              <div class="meta-item">
                <el-icon><Calendar /></el-icon>
                <span>截止：{{ formatDate(assignment.dueDate) }}</span>
              </div>
              <div class="meta-item">
                <el-icon><User /></el-icon>
                <span>提交：{{ assignment.submittedCount }}/{{ assignment.totalStudents }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 创建作业对话框 -->
    <el-dialog v-model="showCreateDialog" title="创建作业" width="800px">
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="120px"
      >
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="作业标题" prop="title">
              <el-input v-model="createForm.title" placeholder="请输入作业标题" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属课题" prop="projectId">
              <el-select v-model="createForm.projectId" placeholder="选择课题">
                <el-option 
                  v-for="project in projectList" 
                  :key="project.id" 
                  :label="project.name" 
                  :value="project.id" 
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-form-item label="作业描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入作业描述"
          />
        </el-form-item>
        
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="截止时间" prop="dueDate">
              <el-date-picker
                v-model="createForm.dueDate"
                type="datetime"
                placeholder="选择截止时间"
                format="YYYY-MM-DD HH:mm"
                value-format="YYYY-MM-DDTHH:mm:ss"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="总分" prop="totalScore">
              <el-input-number v-model="createForm.totalScore" :min="1" :max="1000" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createAssignment" :loading="creating">
          创建作业
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
  School, 
  Calendar, 
  User
} from '@element-plus/icons-vue'

// 类型定义
interface Assignment {
  id: number
  title: string
  description: string
  projectName: string
  projectId: number
  status: 'draft' | 'published' | 'ongoing' | 'closed'
  dueDate: string
  totalScore: number
  submittedCount: number
  gradedCount: number
  totalStudents: number
  createdAt: string
}

// 响应式数据
const assignments = ref<Assignment[]>([
  {
    id: 1,
    title: '数据结构第一次作业',
    description: '完成链表相关题目',
    projectName: '数据结构与算法',
    projectId: 3,
    status: 'ongoing',
    dueDate: '2024-03-25T23:59:59',
    totalScore: 100,
    submittedCount: 25,
    gradedCount: 20,
    totalStudents: 30,
    createdAt: '2024-03-15T10:00:00'
  },
  {
    id: 2,
    title: 'Web前端开发实践',
    description: '使用Vue.js开发用户管理界面',
    projectName: 'Web开发实战项目',
    projectId: 1,
    status: 'published',
    dueDate: '2024-04-01T23:59:59',
    totalScore: 100,
    submittedCount: 18,
    gradedCount: 15,
    totalStudents: 25,
    createdAt: '2024-03-16T09:00:00'
  }
])

const projectList = ref([
  { id: 1, name: 'Web开发实战项目' },
  { id: 2, name: 'Java后端开发' },
  { id: 3, name: '数据结构与算法' }
])

const showCreateDialog = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()

// 筛选条件
const filters = reactive({
  status: '',
  project: '',
  search: ''
})

// 创建表单
const createForm = reactive({
  title: '',
  description: '',
  projectId: '',
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
  projectId: [
    { required: true, message: '请选择课题', trigger: 'change' }
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

  if (filters.project) {
    result = result.filter(a => a.projectId === Number(filters.project))
  }

  if (filters.search) {
    result = result.filter(a => 
      a.title.toLowerCase().includes(filters.search.toLowerCase())
    )
  }

  return result
})

// 方法
const handleSearch = () => {
  // 搜索已经通过计算属性自动处理
}

const getStatusType = (status: string) => {
  const typeMap: Record<string, string> = {
    draft: 'info',
    published: 'success',
    ongoing: 'warning',
    closed: 'danger'
  }
  return typeMap[status] || 'info'
}

const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    draft: '草稿',
    published: '已发布',
    ongoing: '进行中',
    closed: '已截止'
  }
  return textMap[status] || '未知'
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const viewAssignment = (assignment: Assignment) => {
  // 查看作业详情
  console.log('查看作业:', assignment)
}

const editAssignment = (assignment: Assignment) => {
  // 编辑作业
  console.log('编辑作业:', assignment)
}

const deleteAssignment = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这个作业吗？', '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 删除作业
    assignments.value = assignments.value.filter(a => a.id !== id)
    ElMessage.success('作业删除成功')
  } catch {
    // 用户取消删除
  }
}

const createAssignment = async () => {
  if (!createFormRef.value) return
  
  try {
    const valid = await createFormRef.value.validate()
    if (!valid) return
    
    creating.value = true
    
    // 模拟创建作业
    const newAssignment: Assignment = {
      id: Date.now(),
      title: createForm.title,
      description: createForm.description,
      projectName: projectList.value.find(p => p.id === Number(createForm.projectId))?.name || '',
      projectId: Number(createForm.projectId),
      status: 'draft',
      dueDate: createForm.dueDate,
      totalScore: createForm.totalScore,
      submittedCount: 0,
      gradedCount: 0,
      totalStudents: 30,
      createdAt: new Date().toISOString()
    }
    
    assignments.value.unshift(newAssignment)
    showCreateDialog.value = false
    ElMessage.success('作业创建成功')
    
    // 重置表单
    createFormRef.value.resetFields()
  } catch (error) {
    console.error('创建作业失败:', error)
    ElMessage.error('创建作业失败')
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  // 初始化时可以加载真实数据
})
</script>

<style scoped>
.assignments-view {
  padding: 24px;
  background-color: #f5f6fa;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 24px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-content h1 {
  margin: 0 0 8px 0;
  color: #1f2937;
  font-size: 28px;
  font-weight: 600;
}

.header-content p {
  margin: 0;
  color: #6b7280;
  font-size: 16px;
}

.filters-section {
  margin-bottom: 24px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.assignments-grid {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  padding: 24px;
}

.assignment-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.assignment-card {
  border: 1px solid #e5e7eb;
  transition: all 0.3s;
}

.assignment-card:hover {
  border-color: #3b82f6;
  transform: translateY(-2px);
}

.assignment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.assignment-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.assignment-title h3 {
  margin: 0;
  font-size: 18px;
  color: #1f2937;
}

.assignment-description {
  color: #6b7280;
  margin-bottom: 16px;
  line-height: 1.5;
}

.assignment-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #6b7280;
  font-size: 14px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}
</style> 