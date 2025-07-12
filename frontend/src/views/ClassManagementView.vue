<template>
  <div class="class-management">
    <div class="page-header">
      <h1>班级管理</h1>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        创建班级
      </el-button>
    </div>

    <!-- 班级列表 -->
    <div class="class-list">
      <el-row :gutter="20">
        <el-col 
          v-for="classItem in classList" 
          :key="classItem.id" 
          :xs="24" :sm="12" :md="8" :lg="6"
        >
          <el-card 
            class="class-card" 
            shadow="hover"
            @click="viewClassDetails(classItem)"
          >
            <div class="class-header">
              <h3>{{ classItem.name }}</h3>
              <el-tag :type="getClassStatusType(classItem)">
                {{ classItem.visibility }}
              </el-tag>
            </div>
            
            <div class="class-description">
              {{ classItem.description || '暂无描述' }}
            </div>
            
            <div class="class-stats">
              <div class="stat-item">
                <el-icon><User /></el-icon>
                <span>{{ classItem.memberCount || 0 }} 人</span>
              </div>
              <div class="stat-item">
                <el-icon><FolderOpened /></el-icon>
                <span>{{ classItem.projectCount || 0 }} 项目</span>
              </div>
              <div class="stat-item">
                <el-icon><Document /></el-icon>
                <span>{{ classItem.assignmentCount || 0 }} 作业</span>
              </div>
            </div>
            
            <div class="class-actions">
              <el-button-group>
                <el-button size="small" @click.stop="manageMembers(classItem)">
                  成员管理
                </el-button>
                <el-button size="small" @click.stop="viewAssignments(classItem)">
                  作业管理
                </el-button>
              </el-button-group>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 创建班级对话框 -->
    <el-dialog 
      v-model="createDialogVisible" 
      title="创建班级" 
      width="500px"
    >
      <el-form 
        ref="createFormRef" 
        :model="createForm" 
        :rules="createRules" 
        label-width="80px"
      >
        <el-form-item label="班级名称" prop="name">
          <el-input 
            v-model="createForm.name" 
            placeholder="请输入班级名称"
          />
        </el-form-item>
        
        <el-form-item label="班级描述" prop="description">
          <el-input 
            v-model="createForm.description" 
            type="textarea" 
            :rows="3"
            placeholder="请输入班级描述"
          />
        </el-form-item>
        
        <el-form-item label="班主任" prop="teacherId">
          <el-select 
            v-model="createForm.teacherId" 
            placeholder="选择班主任"
            filterable
          >
            <el-option
              v-for="teacher in teacherList"
              :key="teacher.id"
              :label="teacher.name"
              :value="teacher.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createClass" :loading="createLoading">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 班级详情对话框 -->
    <el-dialog 
      v-model="detailDialogVisible" 
      :title="selectedClass?.name" 
      width="900px"
    >
      <div v-if="selectedClass" class="class-details">
        <el-tabs v-model="activeTab">
          <!-- 基本信息 -->
          <el-tab-pane label="基本信息" name="basic">
            <div class="basic-info">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="班级名称">
                  {{ selectedClass.name }}
                </el-descriptions-item>
                <el-descriptions-item label="创建时间">
                  {{ formatDate(selectedClass.created_at) }}
                </el-descriptions-item>
                <el-descriptions-item label="班级描述" :span="2">
                  {{ selectedClass.description || '暂无描述' }}
                </el-descriptions-item>
                <el-descriptions-item label="成员数量">
                  {{ classMembers.length }}
                </el-descriptions-item>
                <el-descriptions-item label="项目数量">
                  {{ classProjects.length }}
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </el-tab-pane>

          <!-- 成员管理 -->
          <el-tab-pane label="成员管理" name="members">
            <div class="members-management">
              <div class="members-header">
                <el-button type="primary" size="small" @click="showAddMemberDialog">
                  <el-icon><Plus /></el-icon>
                  添加成员
                </el-button>
              </div>
              
              <el-table :data="classMembers" style="width: 100%">
                <el-table-column prop="name" label="姓名" />
                <el-table-column prop="username" label="用户名" />
                <el-table-column prop="email" label="邮箱" />
                <el-table-column prop="role" label="角色">
                  <template #default="{ row }">
                    <el-tag :type="getRoleTagType(row.role)">
                      {{ getRoleName(row.role) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="200">
                  <template #default="{ row }">
                    <el-button-group size="small">
                      <el-button @click="editMemberRole(row)">编辑角色</el-button>
                      <el-button type="danger" @click="removeMember(row)">移除</el-button>
                    </el-button-group>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-tab-pane>

          <!-- 项目列表 -->
          <el-tab-pane label="项目列表" name="projects">
            <div class="projects-list">
              <el-table :data="classProjects" style="width: 100%">
                <el-table-column prop="name" label="项目名称" />
                <el-table-column prop="description" label="描述" />
                <el-table-column prop="created_at" label="创建时间">
                  <template #default="{ row }">
                    {{ formatDate(row.created_at) }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="150">
                  <template #default="{ row }">
                    <el-button size="small" @click="openProject(row)">
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

    <!-- 添加成员对话框 -->
    <el-dialog 
      v-model="addMemberDialogVisible" 
      title="添加成员" 
      width="400px"
    >
      <el-form 
        ref="addMemberFormRef" 
        :model="addMemberForm" 
        :rules="addMemberRules" 
        label-width="80px"
      >
        <el-form-item label="选择用户" prop="userId">
          <el-select 
            v-model="addMemberForm.userId" 
            placeholder="选择要添加的用户"
            filterable
          >
            <el-option
              v-for="user in availableUsers"
              :key="user.id"
              :label="`${user.name} (${user.username})`"
              :value="user.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="分配角色" prop="role">
          <el-select v-model="addMemberForm.role" placeholder="选择角色">
            <el-option label="学生" :value="20" />
            <el-option label="助教" :value="30" />
            <el-option label="教师" :value="40" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="addMemberDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="addMember" :loading="addMemberLoading">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, User, FolderOpened, Document } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// 响应式数据
const classList = ref([])
const classMembers = ref([])
const classProjects = ref([])
const teacherList = ref([])
const availableUsers = ref([])

const createDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const addMemberDialogVisible = ref(false)
const createLoading = ref(false)
const addMemberLoading = ref(false)

const selectedClass = ref(null)
const activeTab = ref('basic')

// 表单数据
const createForm = reactive({
  name: '',
  description: '',
  teacherId: null
})

const addMemberForm = reactive({
  userId: null,
  role: 20 // 默认学生角色
})

// 表单引用
const createFormRef = ref()
const addMemberFormRef = ref()

// 表单验证规则
const createRules = {
  name: [
    { required: true, message: '请输入班级名称', trigger: 'blur' }
  ],
  teacherId: [
    { required: true, message: '请选择班主任', trigger: 'change' }
  ]
}

const addMemberRules = {
  userId: [
    { required: true, message: '请选择用户', trigger: 'change' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// 页面初始化
onMounted(() => {
  loadClassList()
  loadTeacherList()
  loadAvailableUsers()
})

// 加载班级列表
const loadClassList = async () => {
  try {
    // 从当前用户的团队中获取班级
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

// 加载教师列表
const loadTeacherList = async () => {
  try {
    const response = await fetch('/api/users?role=teacher', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      teacherList.value = data.data || []
    }
  } catch (error) {
    console.error('加载教师列表失败:', error)
  }
}

// 加载可用用户列表
const loadAvailableUsers = async () => {
  try {
    const response = await fetch('/api/users', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      availableUsers.value = data.data || []
    }
  } catch (error) {
    console.error('加载用户列表失败:', error)
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  createDialogVisible.value = true
  // 重置表单
  Object.assign(createForm, {
    name: '',
    description: '',
    teacherId: null
  })
}

// 创建班级
const createClass = async () => {
  if (!createFormRef.value) return
  
  const valid = await createFormRef.value.validate()
  if (!valid) return
  
  createLoading.value = true
  
  try {
    const response = await fetch('/api/teams/classes', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(createForm)
    })
    
    if (response.ok) {
      ElMessage.success('班级创建成功')
      createDialogVisible.value = false
      loadClassList()
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '创建班级失败')
    }
  } catch (error) {
    console.error('创建班级失败:', error)
    ElMessage.error('创建班级失败')
  } finally {
    createLoading.value = false
  }
}

// 查看班级详情
const viewClassDetails = async (classItem) => {
  selectedClass.value = classItem
  detailDialogVisible.value = true
  activeTab.value = 'basic'
  
  // 加载班级成员和项目
  await loadClassMembers(classItem.id)
  await loadClassProjects(classItem.id)
}

// 加载班级成员
const loadClassMembers = async (classId) => {
  try {
    const response = await fetch(`/api/teams/${classId}/members`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      classMembers.value = data.data || []
    }
  } catch (error) {
    console.error('加载班级成员失败:', error)
  }
}

// 加载班级项目
const loadClassProjects = async (classId) => {
  try {
    const response = await fetch(`/api/teams/${classId}/projects`, {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      classProjects.value = data.data || []
    }
  } catch (error) {
    console.error('加载班级项目失败:', error)
  }
}

// 显示添加成员对话框
const showAddMemberDialog = () => {
  addMemberDialogVisible.value = true
  Object.assign(addMemberForm, {
    userId: null,
    role: 20
  })
}

// 添加成员
const addMember = async () => {
  if (!addMemberFormRef.value) return
  
  const valid = await addMemberFormRef.value.validate()
  if (!valid) return
  
  addMemberLoading.value = true
  
  try {
    const response = await fetch(`/api/teams/${selectedClass.value.id}/members`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        user_id: addMemberForm.userId,
        role: addMemberForm.role
      })
    })
    
    if (response.ok) {
      ElMessage.success('添加成员成功')
      addMemberDialogVisible.value = false
      loadClassMembers(selectedClass.value.id)
    } else {
      const errorData = await response.json()
      ElMessage.error(errorData.error || '添加成员失败')
    }
  } catch (error) {
    console.error('添加成员失败:', error)
    ElMessage.error('添加成员失败')
  } finally {
    addMemberLoading.value = false
  }
}

// 移除成员
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
    
    const response = await fetch(`/api/teams/${selectedClass.value.id}/members/${member.id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    if (response.ok) {
      ElMessage.success('移除成员成功')
      loadClassMembers(selectedClass.value.id)
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

// 管理成员
const manageMembers = (classItem) => {
  viewClassDetails(classItem)
  activeTab.value = 'members'
}

// 查看作业
const viewAssignments = (classItem) => {
  router.push(`/assignments?classId=${classItem.id}`)
}

// 打开项目
const openProject = (project) => {
  // 在新窗口打开GitLab项目
  window.open(project.web_url, '_blank')
}

// 工具函数
const getClassStatusType = (classItem) => {
  switch (classItem.visibility) {
    case 'public':
      return 'success'
    case 'internal':
      return 'warning'
    case 'private':
      return 'info'
    default:
      return 'info'
  }
}

const getRoleTagType = (role) => {
  switch (role) {
    case 50: return 'danger'  // 管理员
    case 40: return 'warning' // 教师
    case 30: return 'success' // 助教
    case 20: return 'info'    // 学生
    default: return 'info'
  }
}

const getRoleName = (role) => {
  switch (role) {
    case 50: return '管理员'
    case 40: return '教师'
    case 30: return '助教'
    case 20: return '学生'
    case 10: return '访客'
    default: return '未知'
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.class-management {
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

.class-list {
  margin-top: 20px;
}

.class-card {
  cursor: pointer;
  transition: all 0.3s;
  height: 280px;
}

.class-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.class-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.class-header h3 {
  margin: 0;
  color: #303133;
  font-size: 16px;
  font-weight: 600;
}

.class-description {
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

.class-stats {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
}

.stat-item {
  display: flex;
  align-items: center;
  color: #909399;
  font-size: 12px;
}

.stat-item .el-icon {
  margin-right: 4px;
}

.class-actions {
  text-align: center;
}

.class-details {
  margin-top: 20px;
}

.basic-info {
  padding: 20px 0;
}

.members-management {
  padding: 20px 0;
}

.members-header {
  margin-bottom: 16px;
}

.projects-list {
  padding: 20px 0;
}

@media (max-width: 768px) {
  .class-management {
    padding: 10px;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }
  
  .class-stats {
    flex-direction: column;
    gap: 8px;
  }
}
</style> 