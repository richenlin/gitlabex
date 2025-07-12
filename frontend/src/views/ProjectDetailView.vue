<template>
  <div class="project-detail">
    <div class="project-header">
      <el-page-header @back="$router.go(-1)" content="课题详情" />
    </div>

    <div class="project-content" v-if="projectInfo">
      <!-- 课题基础信息 -->
      <el-card class="project-info-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>课题信息</span>
            <el-button 
              v-if="canManageProject" 
              type="primary" 
              size="small" 
              @click="editProjectDialog = true"
            >
              编辑课题
            </el-button>
          </div>
        </template>
        
        <el-descriptions :column="2" border>
          <el-descriptions-item label="课题名称">
            {{ projectInfo.name }}
          </el-descriptions-item>
          <el-descriptions-item label="课题代码">
            <el-tag type="success">{{ projectInfo.code }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="指导老师">
            {{ projectInfo.teacher?.name }}
          </el-descriptions-item>
          <el-descriptions-item label="所属班级">
            {{ projectInfo.class?.name || '无' }}
          </el-descriptions-item>
          <el-descriptions-item label="课题状态">
            <el-tag :type="getStatusType(projectInfo.status)">
              {{ getStatusText(projectInfo.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatDate(projectInfo.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="课题描述" :span="2">
            {{ projectInfo.description || '暂无描述' }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 课题统计 -->
      <el-card class="stats-card" shadow="never">
        <template #header>
          <span>课题统计</span>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="6">
            <el-statistic title="成员数量" :value="projectStats.member_count" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="作业数量" :value="projectStats.assignment_count" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="已完成作业" :value="projectStats.completed_assignments" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="待完成作业" :value="projectStats.pending_assignments" />
          </el-col>
        </el-row>
      </el-card>

      <!-- 课题成员 -->
      <el-card class="members-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>课题成员</span>
            <el-button 
              v-if="canManageProject" 
              type="primary" 
              size="small" 
              @click="addMemberDialog = true"
            >
              添加成员
            </el-button>
          </div>
        </template>
        
        <el-table :data="projectMembers" v-loading="loading.members">
          <el-table-column prop="name" label="姓名" />
          <el-table-column prop="email" label="邮箱" />
          <el-table-column prop="username" label="用户名" />
          <el-table-column label="角色">
            <template #default="{ row }">
              <el-tag :type="row.role === 'leader' ? 'warning' : 'info'">
                {{ row.role === 'leader' ? '组长' : '成员' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="加入时间">
            <template #default="{ row }">
              {{ formatDate(row.joined_at) }}
            </template>
          </el-table-column>
          <el-table-column v-if="canManageProject" label="操作" width="150">
            <template #default="{ row }">
              <el-button 
                type="warning" 
                size="small" 
                @click="updateRole(row.id, row.role)"
              >
                {{ row.role === 'leader' ? '取消组长' : '设为组长' }}
              </el-button>
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

      <!-- 课题作业 -->
      <el-card class="assignments-card" shadow="never">
        <template #header>
          <div class="card-header">
            <span>课题作业</span>
            <el-button 
              v-if="canManageProject" 
              type="primary" 
              size="small" 
              @click="createAssignmentDialog = true"
            >
              创建作业
            </el-button>
          </div>
        </template>
        
        <el-table :data="projectAssignments" v-loading="loading.assignments">
          <el-table-column prop="title" label="作业标题" />
          <el-table-column prop="description" label="描述" show-overflow-tooltip />
          <el-table-column label="截止时间">
            <template #default="{ row }">
              {{ formatDate(row.due_date) }}
            </template>
          </el-table-column>
          <el-table-column prop="max_score" label="满分" width="80" />
          <el-table-column prop="status" label="状态">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'">
                {{ row.status === 'active' ? '进行中' : '已关闭' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button 
                type="primary" 
                size="small" 
                @click="viewAssignment(row.id)"
              >
                查看
              </el-button>
              <el-button 
                v-if="authStore.userRole === 3"
                type="success" 
                size="small" 
                @click="submitAssignment(row.id)"
              >
                提交
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 编辑课题对话框 -->
    <el-dialog v-model="editProjectDialog" title="编辑课题" width="500px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="课题名称">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="课题描述">
          <el-input 
            v-model="editForm.description" 
            type="textarea" 
            :rows="3"
          />
        </el-form-item>
        <el-form-item label="课题状态">
          <el-select v-model="editForm.status" style="width: 100%">
            <el-option label="进行中" value="active" />
            <el-option label="已完成" value="completed" />
            <el-option label="已归档" value="archived" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editProjectDialog = false">取消</el-button>
        <el-button type="primary" @click="saveProjectEdit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 添加成员对话框 -->
    <el-dialog v-model="addMemberDialog" title="添加成员" width="400px">
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
        <el-form-item label="角色">
          <el-select v-model="addMemberForm.role" style="width: 100%">
            <el-option label="成员" value="member" />
            <el-option label="组长" value="leader" />
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

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const projectInfo = ref<any>(null)
const projectStats = ref<any>({})
const projectMembers = ref<any[]>([])
const projectAssignments = ref<any[]>([])
const availableStudents = ref<any[]>([])

const loading = ref({
  project: false,
  members: false,
  assignments: false
})

const editProjectDialog = ref(false)
const addMemberDialog = ref(false)
const createAssignmentDialog = ref(false)

const editForm = ref({
  name: '',
  description: '',
  status: 'active'
})

const addMemberForm = ref({
  student_id: null,
  role: 'member'
})

const canManageProject = computed(() => {
  return authStore.userRole === 1 || // 管理员
         (authStore.userRole === 2 && projectInfo.value?.teacher_id === authStore.user?.id) // 课题创建者
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

const loadProjectDetail = async () => {
  loading.value.project = true
  try {
    // 临时使用静态数据
    projectInfo.value = {
      id: 1,
      name: 'Web开发课题',
      code: 'WEB2024001',
      description: '学习现代Web开发技术，包括前端和后端开发',
      teacher: { name: '李教授' },
      teacher_id: 1,
      class: { name: '计算机科学基础班' },
      status: 'active',
      created_at: new Date().toISOString()
    }
    
    // 设置编辑表单
    editForm.value = {
      name: projectInfo.value.name,
      description: projectInfo.value.description || '',
      status: projectInfo.value.status
    }
  } catch (error) {
    ElMessage.error('获取课题信息失败')
    console.error(error)
  } finally {
    loading.value.project = false
  }
}

const loadProjectStats = async () => {
  try {
    // 临时使用静态数据
    projectStats.value = {
      member_count: 8,
      assignment_count: 4,
      completed_assignments: 2,
      pending_assignments: 2
    }
  } catch (error) {
    console.error('获取课题统计失败:', error)
  }
}

const loadProjectMembers = async () => {
  loading.value.members = true
  try {
    // 临时使用静态数据
    projectMembers.value = [
      { id: 1, name: '李同学', email: 'li@example.com', username: 'li_student', role: 'leader', joined_at: new Date().toISOString() },
      { id: 2, name: '王同学', email: 'wang@example.com', username: 'wang_student', role: 'member', joined_at: new Date().toISOString() },
      { id: 3, name: '张同学', email: 'zhang@example.com', username: 'zhang_student', role: 'member', joined_at: new Date().toISOString() }
    ]
  } catch (error) {
    ElMessage.error('获取课题成员失败')
    console.error(error)
  } finally {
    loading.value.members = false
  }
}

const loadProjectAssignments = async () => {
  loading.value.assignments = true
  try {
    // 临时使用静态数据
    projectAssignments.value = [
      { 
        id: 1, 
        title: 'HTML基础练习', 
        description: '完成HTML页面设计', 
        due_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
        max_score: 100,
        status: 'active'
      },
      { 
        id: 2, 
        title: 'CSS样式设计', 
        description: '学习CSS布局和样式', 
        due_date: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString(),
        max_score: 100,
        status: 'active'
      }
    ]
  } catch (error) {
    console.error('获取课题作业失败:', error)
  } finally {
    loading.value.assignments = false
  }
}

const loadAvailableStudents = async () => {
  try {
    // 临时使用静态数据
    availableStudents.value = [
      { id: 4, name: '陈同学' },
      { id: 5, name: '赵同学' }
    ]
  } catch (error) {
    console.error('获取学生列表失败:', error)
  }
}

const saveProjectEdit = async () => {
  try {
    // TODO: 实现API调用
    ElMessage.success('课题信息更新成功')
    editProjectDialog.value = false
    await loadProjectDetail()
  } catch (error) {
    ElMessage.error('更新课题信息失败')
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
    ElMessage.success('添加成员成功')
    addMemberDialog.value = false
    addMemberForm.value = { student_id: null, role: 'member' }
    await loadProjectMembers()
    await loadProjectStats()
  } catch (error) {
    ElMessage.error('添加成员失败')
    console.error(error)
  }
}

const removeMember = async (studentId: number) => {
  try {
    await ElMessageBox.confirm('确定要移除该成员吗？', '确认', {
      type: 'warning'
    })
    
    // TODO: 实现API调用
    ElMessage.success('移除成员成功')
    await loadProjectMembers()
    await loadProjectStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('移除成员失败')
      console.error(error)
    }
  }
}

const updateRole = async (studentId: number, currentRole: string) => {
  const newRole = currentRole === 'leader' ? 'member' : 'leader'
  
  try {
    // TODO: 实现API调用
    ElMessage.success(`${newRole === 'leader' ? '设置组长' : '取消组长'}成功`)
    await loadProjectMembers()
  } catch (error) {
    ElMessage.error('更新角色失败')
    console.error(error)
  }
}

const viewAssignment = (assignmentId: number) => {
  router.push(`/assignments/${assignmentId}`)
}

const submitAssignment = (assignmentId: number) => {
  router.push(`/assignments/${assignmentId}/submit`)
}

onMounted(() => {
  loadProjectDetail()
  loadProjectStats()
  loadProjectMembers()
  loadProjectAssignments()
  
  if (canManageProject.value) {
    loadAvailableStudents()
  }
})
</script>

<style scoped>
.project-detail {
  padding: 20px;
}

.project-header {
  margin-bottom: 20px;
}

.project-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.project-info-card,
.stats-card,
.members-card,
.assignments-card {
  margin-bottom: 20px;
}

.stats-card .el-row {
  text-align: center;
}
</style> 