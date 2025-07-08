<template>
  <div class="users-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <el-icon class="page-icon"><User /></el-icon>
          用户管理
        </h1>
        <p class="page-description">管理系统用户，分配角色和权限</p>
      </div>
      
      <!-- 操作栏 -->
      <div class="toolbar">
        <el-button 
          type="primary" 
          :icon="Plus"
          @click="createUser"
        >
          添加用户
        </el-button>
        
        <el-button 
          :icon="Refresh"
          @click="loadUsers"
          :loading="isLoading"
        >
          刷新
        </el-button>
        
        <el-button 
          :icon="Download"
          @click="exportUsers"
        >
          导出
        </el-button>
      </div>
    </div>

    <!-- 用户列表 -->
    <div class="users-content">
      <el-card class="users-card" v-loading="isLoading">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><User /></el-icon>
              用户列表
            </span>
            <div class="header-actions">
              <el-select 
                v-model="roleFilter" 
                placeholder="筛选角色"
                clearable
                @change="handleRoleFilter"
                style="width: 120px"
              >
                <el-option label="全部" value="" />
                <el-option label="管理员" value="5" />
                <el-option label="教师" value="4" />
                <el-option label="助教" value="3" />
                <el-option label="学生" value="2" />
                <el-option label="访客" value="1" />
              </el-select>
              <span class="user-count">共 {{ filteredUsers.length }} 个用户</span>
            </div>
          </div>
        </template>

        <!-- 用户表格 -->
        <el-table 
          :data="filteredUsers" 
          style="width: 100%" 
          :default-sort="{ prop: 'last_sync_at', order: 'descending' }"
          empty-text="暂无用户"
          class="users-table"
        >
          <el-table-column type="selection" width="55" />
          
          <el-table-column prop="id" label="ID" width="80" />
          
          <el-table-column prop="avatar" label="头像" width="80">
            <template #default="scope">
              <el-avatar 
                :src="scope.row.avatar" 
                :size="40"
                :alt="scope.row.name"
              >
                <el-icon><User /></el-icon>
              </el-avatar>
            </template>
          </el-table-column>
          
          <el-table-column prop="name" label="姓名" min-width="120">
            <template #default="scope">
              <div class="user-name">
                <span class="name-text">{{ scope.row.name }}</span>
                <el-tag v-if="!scope.row.is_active" type="danger" size="small">已禁用</el-tag>
              </div>
            </template>
          </el-table-column>
          
          <el-table-column prop="username" label="用户名" min-width="120" />
          
          <el-table-column prop="email" label="邮箱" min-width="200" />
          
          <el-table-column prop="role" label="角色" width="120">
            <template #default="scope">
              <el-tag :type="getRoleTagType(scope.row.role)" size="small">
                {{ getRoleText(scope.row.role) }}
              </el-tag>
            </template>
          </el-table-column>
          
          <el-table-column prop="gitlab_id" label="GitLab ID" width="100" />
          
          <el-table-column prop="last_sync_at" label="最后同步" width="180">
            <template #default="scope">
              {{ formatDate(scope.row.last_sync_at) }}
            </template>
          </el-table-column>
          
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="scope">
              <el-button-group>
                <el-button 
                  type="primary" 
                  size="small" 
                  :icon="Edit"
                  @click="editUser(scope.row)"
                >
                  编辑
                </el-button>
                <el-button 
                  :type="scope.row.is_active ? 'warning' : 'success'"
                  size="small" 
                  :icon="scope.row.is_active ? Lock : Unlock"
                  @click="toggleUserStatus(scope.row)"
                >
                  {{ scope.row.is_active ? '禁用' : '启用' }}
                </el-button>
                <el-button 
                  type="danger" 
                  size="small" 
                  :icon="Delete"
                  @click="deleteUser(scope.row)"
                >
                  删除
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 用户编辑对话框 -->
    <el-dialog
      v-model="userDialogVisible"
      :title="editingUser ? '编辑用户' : '添加用户'"
      width="600px"
      :before-close="handleDialogClose"
    >
      <el-form 
        ref="userFormRef"
        :model="userForm" 
        :rules="userFormRules"
        label-width="100px"
      >
        <el-form-item label="头像">
          <el-upload
            class="avatar-uploader"
            action=""
            :show-file-list="false"
            :auto-upload="false"
            :on-change="handleAvatarChange"
          >
            <img v-if="userForm.avatar" :src="userForm.avatar" class="avatar" />
            <el-icon v-else class="avatar-uploader-icon"><Plus /></el-icon>
          </el-upload>
        </el-form-item>
        
        <el-form-item label="姓名" prop="name">
          <el-input v-model="userForm.name" placeholder="请输入姓名" />
        </el-form-item>
        
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名" />
        </el-form-item>
        
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        
        <el-form-item label="角色" prop="role">
          <el-select v-model="userForm.role" placeholder="请选择角色">
            <el-option label="访客" :value="1" />
            <el-option label="学生" :value="2" />
            <el-option label="助教" :value="3" />
            <el-option label="教师" :value="4" />
            <el-option label="管理员" :value="5" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="GitLab ID" prop="gitlab_id">
          <el-input-number 
            v-model="userForm.gitlab_id" 
            placeholder="请输入GitLab用户ID"
            :min="1"
            style="width: 100%"
          />
        </el-form-item>
        
        <el-form-item label="状态">
          <el-switch
            v-model="userForm.is_active"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="userDialogVisible = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="saveUser"
            :loading="isSaving"
          >
            保存
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  User, 
  Plus, 
  Refresh,
  Download,
  Edit,
  Lock,
  Unlock,
  Delete
} from '@element-plus/icons-vue'
import { ApiService } from '../services/api'
import type { User as UserType } from '../services/api'

// 响应式数据
const users = ref<UserType[]>([])
const isLoading = ref(false)
const isSaving = ref(false)
const userDialogVisible = ref(false)
const roleFilter = ref('')
const editingUser = ref<UserType | null>(null)

// 表单数据
const userForm = ref({
  id: 0,
  name: '',
  username: '',
  email: '',
  avatar: '',
  role: 2,
  gitlab_id: 0,
  is_active: true
})

const userFormRules = {
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' }
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ],
  gitlab_id: [
    { required: true, message: '请输入GitLab用户ID', trigger: 'blur' }
  ]
}

// 计算属性
const filteredUsers = computed(() => {
  if (!roleFilter.value) return users.value
  return users.value.filter(user => user.role.toString() === roleFilter.value)
})

// 生命周期
onMounted(() => {
  loadUsers()
})

// 方法
const loadUsers = async () => {
  isLoading.value = true
  try {
    const response = await ApiService.getActiveUsers()
    users.value = response.data
    ElMessage.success('用户列表加载成功')
  } catch (error) {
    console.error('加载用户失败:', error)
    ElMessage.error('加载用户失败')
  } finally {
    isLoading.value = false
  }
}

const handleRoleFilter = () => {
  // 过滤逻辑已在计算属性中处理
}

const createUser = () => {
  editingUser.value = null
  userForm.value = {
    id: 0,
    name: '',
    username: '',
    email: '',
    avatar: '',
    role: 2,
    gitlab_id: 0,
    is_active: true
  }
  userDialogVisible.value = true
}

const editUser = (user: UserType) => {
  editingUser.value = user
  userForm.value = {
    id: user.id,
    name: user.name,
    username: user.username,
    email: user.email,
    avatar: user.avatar,
    role: user.role,
    gitlab_id: user.gitlab_id,
    is_active: user.is_active
  }
  userDialogVisible.value = true
}

const saveUser = async () => {
  isSaving.value = true
  try {
    // 这里应该调用保存用户API
    // if (editingUser.value) {
    //   await ApiService.updateUser(userForm.value)
    // } else {
    //   await ApiService.createUser(userForm.value)
    // }
    
    // 模拟保存
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    ElMessage.success(editingUser.value ? '用户更新成功' : '用户创建成功')
    userDialogVisible.value = false
    loadUsers()
  } catch (error) {
    console.error('保存用户失败:', error)
    ElMessage.error('保存用户失败')
  } finally {
    isSaving.value = false
  }
}

const toggleUserStatus = async (user: UserType) => {
  try {
    const action = user.is_active ? '禁用' : '启用'
    await ElMessageBox.confirm(
      `确定要${action}用户 "${user.name}" 吗？`,
      '状态确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    // 这里应该调用切换状态API
    // await ApiService.toggleUserStatus(user.id)
    
    // 模拟切换状态
    user.is_active = !user.is_active
    
    ElMessage.success(`用户${action}成功`)
  } catch (error) {
    console.log('取消操作')
  }
}

const deleteUser = async (user: UserType) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.name}" 吗？此操作不可恢复！`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'error',
      }
    )
    
    // 这里应该调用删除API
    // await ApiService.deleteUser(user.id)
    
    // 模拟删除
    const index = users.value.findIndex(u => u.id === user.id)
    if (index > -1) {
      users.value.splice(index, 1)
    }
    
    ElMessage.success('用户删除成功')
  } catch (error) {
    console.log('取消删除')
  }
}

const exportUsers = () => {
  // 模拟导出功能
  const data = users.value.map(user => ({
    ID: user.id,
    姓名: user.name,
    用户名: user.username,
    邮箱: user.email,
    角色: getRoleText(user.role),
    GitLab_ID: user.gitlab_id,
    状态: user.is_active ? '启用' : '禁用',
    最后同步: new Date(user.last_sync_at).toLocaleString('zh-CN')
  }))
  
  const csv = convertToCSV(data)
  downloadCSV(csv, '用户列表.csv')
  ElMessage.success('用户列表导出成功')
}

const handleAvatarChange = (file: any) => {
  // 这里应该实现头像上传逻辑
  const reader = new FileReader()
  reader.onload = (e: any) => {
    userForm.value.avatar = e.target.result
  }
  reader.readAsDataURL(file.raw)
}

const handleDialogClose = () => {
  editingUser.value = null
  userForm.value = {
    id: 0,
    name: '',
    username: '',
    email: '',
    avatar: '',
    role: 2,
    gitlab_id: 0,
    is_active: true
  }
}

// 工具方法
const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN')
}

const getRoleText = (role: number) => {
  const roleMap: { [key: number]: string } = {
    1: '访客',
    2: '学生',
    3: '助教',
    4: '教师',
    5: '管理员'
  }
  return roleMap[role] || '未知'
}

const getRoleTagType = (role: number) => {
  const typeMap: { [key: number]: string } = {
    1: 'info',
    2: 'primary',
    3: 'success',
    4: 'warning',
    5: 'danger'
  }
  return typeMap[role] || 'info'
}

const convertToCSV = (data: any[]) => {
  if (!data.length) return ''
  
  const headers = Object.keys(data[0])
  const csvContent = [
    headers.join(','),
    ...data.map(row => headers.map(header => JSON.stringify(row[header])).join(','))
  ].join('\n')
  
  return csvContent
}

const downloadCSV = (content: string, filename: string) => {
  const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)
  link.setAttribute('href', url)
  link.setAttribute('download', filename)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
</script>

<style scoped>
.users-container {
  padding: 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  background: white;
  padding: 24px;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.header-content h1 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-icon {
  font-size: 28px;
  color: #409eff;
}

.page-description {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
}

.toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
}

.users-content {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.users-card {
  border: none;
  box-shadow: none;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #2c3e50;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-count {
  color: #7f8c8d;
  font-size: 14px;
}

.users-table {
  --el-table-border-color: #f1f2f6;
}

.users-table :deep(.el-table__row:hover > td) {
  background-color: #f8f9ff !important;
}

.user-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.name-text {
  color: #2c3e50;
  font-weight: 500;
}

.avatar-uploader {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
}

.avatar-uploader:hover {
  border-color: #409eff;
}

.avatar {
  width: 80px;
  height: 80px;
  display: block;
  object-fit: cover;
}

.avatar-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 80px;
  height: 80px;
  text-align: center;
  line-height: 80px;
}

.dialog-footer {
  display: flex;
  gap: 12px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .users-container {
    padding: 10px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .toolbar {
    justify-content: flex-start;
    flex-wrap: wrap;
  }
  
  .card-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .users-table {
    font-size: 14px;
  }
}
</style> 