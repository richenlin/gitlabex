<template>
  <div class="classes-view">
    <div class="page-header">
      <h1>班级管理</h1>
      <p>创建和管理您的班级</p>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        创建班级
      </el-button>
    </div>

    <!-- 班级列表 -->
    <div class="classes-grid">
      <el-card
        v-for="classItem in classes"
        :key="classItem.id"
        class="class-card"
        shadow="hover"
        @click="viewClass(classItem)"
      >
        <template #header>
          <div class="class-header">
            <h3>{{ classItem.name }}</h3>
            <el-tag :type="getClassStatusType(classItem.status)">
              {{ getClassStatusText(classItem.status) }}
            </el-tag>
          </div>
        </template>
        <div class="class-content">
          <p class="class-description">{{ classItem.description }}</p>
          <div class="class-stats">
            <div class="stat-item">
              <el-icon><User /></el-icon>
              <span>{{ classItem.studentsCount }} 学生</span>
            </div>
            <div class="stat-item">
              <el-icon><Document /></el-icon>
              <span>{{ classItem.assignmentsCount }} 作业</span>
            </div>
            <div class="stat-item">
              <el-icon><Calendar /></el-icon>
              <span>{{ formatDate(classItem.createdAt) }}</span>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 空状态 -->
      <div v-if="classes.length === 0" class="empty-state">
        <el-empty description="暂无班级">
          <el-button type="primary" @click="showCreateDialog = true">
            创建第一个班级
          </el-button>
        </el-empty>
      </div>
    </div>

    <!-- 创建班级对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建班级"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="120px"
      >
        <el-form-item label="班级名称" prop="name">
          <el-input
            v-model="createForm.name"
            placeholder="请输入班级名称"
            maxlength="50"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="班级描述" prop="description">
          <el-input
            v-model="createForm.description"
            type="textarea"
            placeholder="请输入班级描述"
            :rows="4"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="学期">
          <el-select v-model="createForm.semester" placeholder="选择学期">
            <el-option label="2024春季" value="2024春季" />
            <el-option label="2024秋季" value="2024秋季" />
            <el-option label="2025春季" value="2025春季" />
          </el-select>
        </el-form-item>
        <el-form-item label="课程代码">
          <el-input
            v-model="createForm.courseCode"
            placeholder="如：CS101"
            maxlength="20"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createClass" :loading="creating">
          创建
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Plus, User, Document, Calendar } from '@element-plus/icons-vue'

// 类型定义
interface ClassItem {
  id: number
  name: string
  description: string
  status: 'active' | 'archived' | 'draft'
  studentsCount: number
  assignmentsCount: number
  semester: string
  courseCode: string
  createdAt: string
}

// 响应式数据
const classes = ref<ClassItem[]>([])
const showCreateDialog = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()

// 创建表单
const createForm = reactive({
  name: '',
  description: '',
  semester: '',
  courseCode: ''
})

// 表单验证规则
const createRules: FormRules = {
  name: [
    { required: true, message: '请输入班级名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  description: [
    { max: 200, message: '描述不能超过 200 个字符', trigger: 'blur' }
  ]
}

// 组件挂载时加载数据
onMounted(() => {
  loadClasses()
})

// 加载班级列表
const loadClasses = () => {
  // 模拟数据，实际应从API获取
  classes.value = [
    {
      id: 1,
      name: '高等数学A',
      description: '理工科高等数学课程，包含微积分、线性代数等内容',
      status: 'active',
      studentsCount: 45,
      assignmentsCount: 8,
      semester: '2024春季',
      courseCode: 'MATH101',
      createdAt: '2024-02-15'
    },
    {
      id: 2,
      name: 'Java程序设计',
      description: 'Java语言基础与面向对象编程',
      status: 'active',
      studentsCount: 38,
      assignmentsCount: 12,
      semester: '2024春季',
      courseCode: 'CS201',
      createdAt: '2024-02-20'
    },
    {
      id: 3,
      name: '数据结构与算法',
      description: '计算机专业核心课程',
      status: 'draft',
      studentsCount: 0,
      assignmentsCount: 0,
      semester: '2024秋季',
      courseCode: 'CS301',
      createdAt: '2024-03-01'
    }
  ]
}

// 创建班级
const createClass = async () => {
  if (!createFormRef.value) return

  try {
    const isValid = await createFormRef.value.validate()
    if (!isValid) return

    creating.value = true

    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000))

    const newClass: ClassItem = {
      id: Date.now(),
      name: createForm.name,
      description: createForm.description,
      status: 'draft',
      studentsCount: 0,
      assignmentsCount: 0,
      semester: createForm.semester,
      courseCode: createForm.courseCode,
      createdAt: new Date().toISOString().split('T')[0]
    }

    classes.value.unshift(newClass)
    showCreateDialog.value = false
    ElMessage.success('班级创建成功')
  } catch (error) {
    ElMessage.error('创建班级失败')
  } finally {
    creating.value = false
  }
}

// 查看班级详情
const viewClass = (classItem: ClassItem) => {
  ElMessage.info(`查看班级: ${classItem.name}`)
  // TODO: 路由到班级详情页
}

// 重置表单
const resetForm = () => {
  if (createFormRef.value) {
    createFormRef.value.resetFields()
  }
  Object.assign(createForm, {
    name: '',
    description: '',
    semester: '',
    courseCode: ''
  })
}

// 获取班级状态类型
const getClassStatusType = (status: string) => {
  const statusMap = {
    active: 'success',
    archived: 'info',
    draft: 'warning'
  }
  return statusMap[status as keyof typeof statusMap] || 'info'
}

// 获取班级状态文本
const getClassStatusText = (status: string) => {
  const statusMap = {
    active: '进行中',
    archived: '已归档',
    draft: '草稿'
  }
  return statusMap[status as keyof typeof statusMap] || '未知'
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.classes-view {
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

.page-header h1 {
  font-size: 28px;
  color: #303133;
  margin: 0 0 8px 0;
}

.page-header p {
  color: #909399;
  margin: 0;
}

.classes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 24px;
}

.class-card {
  cursor: pointer;
  transition: all 0.3s ease;
}

.class-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
}

.class-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.class-header h3 {
  margin: 0;
  font-size: 18px;
  color: #303133;
}

.class-content {
  padding-top: 8px;
}

.class-description {
  color: #606266;
  margin: 0 0 16px 0;
  line-height: 1.5;
  min-height: 42px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.class-stats {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #909399;
  font-size: 14px;
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
}

@media (max-width: 768px) {
  .classes-view {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .classes-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }
}
</style> 