<template>
  <div class="class-detail">
    <div class="class-header">
      <el-page-header @back="$router.go(-1)" content="班级详情" />
    </div>

    <div class="class-content" v-if="classInfo">
      <!-- 班级基础信息 -->
      <el-card class="class-info-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>班级信息</span>
            <el-button 
              v-if="canManageClass" 
              type="primary" 
              size="small" 
              @click="editClassDialog = true"
            >
              编辑班级
            </el-button>
          </div>
        </template>
        
        <el-descriptions :column="2" border>
          <el-descriptions-item label="班级名称">
            {{ classInfo.name }}
          </el-descriptions-item>
          <el-descriptions-item label="班级代码">
            <el-tag type="success">{{ classInfo.code }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建老师">
            {{ classInfo.teacher?.name }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatDate(classInfo.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="班级描述" :span="2">
            {{ classInfo.description || '暂无描述' }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 班级统计 -->
      <el-card class="stats-card" shadow="never">
        <template #header>
          <span>班级统计</span>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="8">
            <el-statistic title="学生数量" :value="classStats.student_count" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="课题数量" :value="classStats.project_count" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="活跃课题" :value="classStats.active_projects" />
          </el-col>
        </el-row>
      </el-card>

      <!-- 班级成员 -->
      <el-card class="members-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>班级成员</span>
            <el-button 
              v-if="canManageClass" 
              type="primary" 
              size="small" 
              @click="addMemberDialog = true"
            >
              添加学生
            </el-button>
          </div>
        </template>
        
        <el-table :data="classMembers" v-loading="loading.members">
          <el-table-column prop="name" label="姓名" />
          <el-table-column prop="email" label="邮箱" />
          <el-table-column prop="username" label="用户名" />
          <el-table-column label="加入时间">
            <template #default="{ row }">
              {{ formatDate(row.joined_at) }}
            </template>
          </el-table-column>
          <el-table-column v-if="canManageClass" label="操作" width="120">
            <template #default="{ row }">
              <el-button 
                type="danger" 
                size="small" 
                @click="removeMember(row.id)"
              >
                移除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 班级课题 -->
      <el-card class="projects-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>班级课题</span>
            <el-button 
              v-if="canManageClass" 
              type="primary" 
              size="small" 
              @click="createProjectDialog = true"
            >
              创建课题
            </el-button>
          </div>
        </template>
        
        <el-table :data="classProjects" v-loading="loading.projects">
          <el-table-column prop="name" label="课题名称" />
          <el-table-column prop="description" label="描述" show-overflow-tooltip />
          <el-table-column prop="status" label="状态">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建时间">
            <template #default="{ row }">
              {{ formatDate(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button 
                type="primary" 
                size="small" 
                @click="viewProject(row.id)"
              >
                查看
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 编辑班级对话框 -->
    <el-dialog v-model="editClassDialog" title="编辑班级" width="500px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="班级名称">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="班级描述">
          <el-input 
            v-model="editForm.description" 
            type="textarea" 
            :rows="3"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editClassDialog = false">取消</el-button>
        <el-button type="primary" @click="saveClassEdit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 添加学生对话框 -->
    <el-dialog v-model="addMemberDialog" title="添加学生" width="400px">
      <el-form :model="addMemberForm" label-width="80px">
        <el-form-item label="选择学生">
          <el-select 
            v-model="addMemberForm.student_id" 
            placeholder="请选择学生" 
            filterable
            style="width: 100%"
          >
            <el-option 
              v-for="student in availableStudents" 
              :key="student.id"
              :label="student.name" 
              :value="student.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addMemberDialog = false">取消</el-button>
        <el-button type="primary" @click="addMember">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ApiService } from '../services/api'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const classInfo = ref<any>(null)
const classStats = ref<any>({})
const classMembers = ref<any[]>([])
const classProjects = ref<any[]>([])
const availableStudents = ref<any[]>([])

const loading = ref({
  class: false,
  members: false,
  projects: false
})

const editClassDialog = ref(false)
const addMemberDialog = ref(false)
const createProjectDialog = ref(false)

const editForm = ref({
  name: '',
  description: ''
})

const addMemberForm = ref({
  student_id: null
})

const canManageClass = computed(() => {
  return authStore.userRole === 1 || // 管理员
         (authStore.userRole === 2 && classInfo.value?.teacher_id === authStore.user?.id) // 班级创建者
})

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const getStatusType = (status: string) => {
  const types: Record<string, string> = {
    'active': 'success',
    'completed': 'info',
    'archived': 'warning'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    'active': '进行中',
    'completed': '已完成',
    'archived': '已归档'
  }
  return texts[status] || status
}

const loadClassDetail = async () => {
  loading.value.class = true
  try {
    // 临时使用静态数据，等后端API实现后替换
    classInfo.value = {
      id: 1,
      name: '计算机科学基础班',
      code: 'CS2024001',
      description: '计算机科学与技术专业基础课程班级',
      teacher: { name: '张教授' },
      teacher_id: 1,
      created_at: new Date().toISOString()
    }
    
    // 设置编辑表单
    editForm.value = {
      name: classInfo.value.name,
      description: classInfo.value.description || ''
    }
  } catch (error) {
    ElMessage.error('获取班级信息失败')
    console.error(error)
  } finally {
    loading.value.class = false
  }
}

const loadClassStats = async () => {
  try {
    // 临时使用静态数据
    classStats.value = {
      student_count: 25,
      project_count: 3,
      active_projects: 2
    }
  } catch (error) {
    console.error('获取班级统计失败:', error)
  }
}

const loadClassMembers = async () => {
  loading.value.members = true
  try {
    // 临时使用静态数据
    classMembers.value = [
      { id: 1, name: '李同学', email: 'li@example.com', username: 'li_student', joined_at: new Date().toISOString() },
      { id: 2, name: '王同学', email: 'wang@example.com', username: 'wang_student', joined_at: new Date().toISOString() }
    ]
  } catch (error) {
    ElMessage.error('获取班级成员失败')
    console.error(error)
  } finally {
    loading.value.members = false
  }
}

const loadClassProjects = async () => {
  loading.value.projects = true
  try {
    // 临时使用静态数据
    classProjects.value = [
      { id: 1, name: 'Web开发项目', description: '学习前端和后端开发', status: 'active', created_at: new Date().toISOString() },
      { id: 2, name: '数据结构研究', description: '深入学习数据结构与算法', status: 'active', created_at: new Date().toISOString() }
    ]
  } catch (error) {
    console.error('获取班级课题失败:', error)
  } finally {
    loading.value.projects = false
  }
}

const loadAvailableStudents = async () => {
  try {
    // 临时使用静态数据
    availableStudents.value = [
      { id: 3, name: '陈同学' },
      { id: 4, name: '赵同学' }
    ]
  } catch (error) {
    console.error('获取学生列表失败:', error)
  }
}

const saveClassEdit = async () => {
  try {
    // TODO: 实现API调用
    ElMessage.success('班级信息更新成功')
    editClassDialog.value = false
    await loadClassDetail()
  } catch (error) {
    ElMessage.error('更新班级信息失败')
    console.error(error)
  }
}

const addMember = async () => {
  if (!addMemberForm.value.student_id) {
    ElMessage.warning('请选择学生')
    return
  }
  
  try {
    // TODO: 实现API调用
    ElMessage.success('添加学生成功')
    addMemberDialog.value = false
    addMemberForm.value.student_id = null
    await loadClassMembers()
    await loadClassStats()
  } catch (error) {
    ElMessage.error('添加学生失败')
    console.error(error)
  }
}

const removeMember = async (studentId: number) => {
  try {
    await ElMessageBox.confirm('确定要移除该学生吗？', '确认', {
      type: 'warning'
    })
    
    // TODO: 实现API调用
    ElMessage.success('移除学生成功')
    await loadClassMembers()
    await loadClassStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('移除学生失败')
      console.error(error)
    }
  }
}

const viewProject = (projectId: number) => {
  router.push(`/projects/${projectId}`)
}

onMounted(() => {
  loadClassDetail()
  loadClassStats()
  loadClassMembers()
  loadClassProjects()
  
  if (canManageClass.value) {
    loadAvailableStudents()
  }
})
</script>

<style scoped>
.class-detail {
  padding: 20px;
}

.class-header {
  margin-bottom: 20px;
}

.class-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.class-info-card,
.stats-card,
.members-card,
.projects-card {
  margin-bottom: 20px;
}

.stats-card .el-row {
  text-align: center;
}
</style> 