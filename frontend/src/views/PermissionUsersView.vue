<template>
  <div class="permission-users-view">
    <!-- 页面标题和操作栏 -->
    <div class="page-header">
      <div class="header-left">
        <h1>用户管理</h1>
        <p>查看和管理系统用户信息及权限</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="handleSyncUsers">
          <el-icon><Refresh /></el-icon>
          同步GitLab用户
        </el-button>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="filters">
      <el-row :gutter="16">
        <el-col :span="8">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索用户名、邮箱..."
            clearable
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="6">
          <el-select v-model="selectedRole" placeholder="筛选角色" clearable @change="handleSearch">
            <el-option label="全部" value="" />
            <el-option label="管理员" value="admin" />
            <el-option label="教师" value="teacher" />
            <el-option label="助教" value="assistant" />
            <el-option label="学生" value="student" />
            <el-option label="访客" value="guest" />
          </el-select>
        </el-col>
        <el-col :span="6">
          <el-select v-model="selectedStatus" placeholder="筛选状态" clearable @change="handleSearch">
            <el-option label="全部" value="" />
            <el-option label="活跃" value="active" />
            <el-option label="非活跃" value="inactive" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-button @click="handleExport">导出用户</el-button>
        </el-col>
      </el-row>
    </div>

    <!-- 用户列表 -->
    <el-table 
      v-loading="loading" 
      :data="users" 
      style="width: 100%"
      @sort-change="handleSortChange"
    >
      <el-table-column prop="avatar" label="头像" width="80">
        <template #default="{ row }">
          <el-avatar :src="row.avatar" :size="40">
            {{ row.name.charAt(0) }}
          </el-avatar>
        </template>
      </el-table-column>
      
      <el-table-column prop="name" label="姓名" sortable="custom" />
      <el-table-column prop="username" label="用户名" sortable="custom" />
      <el-table-column prop="email" label="邮箱" />
      
      <el-table-column prop="role_name" label="角色">
        <template #default="{ row }">
          <el-tag 
            :type="getRoleTagType(row.dynamic_role || row.role)"
            effect="light"
          >
            {{ row.dynamic_role_name || row.role_name }}
          </el-tag>
          <div v-if="row.dynamic_role && row.dynamic_role !== getRoleFromStaticRole(row.role)" class="role-info">
            <el-text size="small" type="info">
              静态: {{ row.role_name }}
            </el-text>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column prop="is_active" label="状态">
        <template #default="{ row }">
          <el-switch
            v-model="row.is_active"
            @change="handleStatusChange(row)"
            :disabled="isCurrentUser(row.id)"
          />
        </template>
      </el-table-column>
      
      <el-table-column prop="last_sync_at" label="最后同步" sortable="custom">
        <template #default="{ row }">
          <el-text size="small">
            {{ formatDate(row.last_sync_at) }}
          </el-text>
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="handleViewUser(row)">查看</el-button>
          <el-button 
            size="small" 
            type="primary" 
            @click="handleEditRole(row)"
            :disabled="isCurrentUser(row.id)"
          >
            编辑角色
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :small="false"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 用户详情对话框 -->
    <el-dialog v-model="userDetailVisible" title="用户详情" width="500px">
      <div v-if="selectedUser" class="user-detail">
        <div class="user-avatar">
          <el-avatar :src="selectedUser.avatar" :size="80">
            {{ selectedUser.name.charAt(0) }}
          </el-avatar>
        </div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="姓名">{{ selectedUser.name }}</el-descriptions-item>
          <el-descriptions-item label="用户名">{{ selectedUser.username }}</el-descriptions-item>
          <el-descriptions-item label="邮箱">{{ selectedUser.email }}</el-descriptions-item>
          <el-descriptions-item label="GitLab ID">{{ selectedUser.gitlab_id }}</el-descriptions-item>
          <el-descriptions-item label="当前角色">
            <el-tag :type="getRoleTagType(selectedUser.dynamic_role || selectedUser.role)">
              {{ selectedUser.dynamic_role_name || selectedUser.role_name }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="静态角色">{{ selectedUser.role_name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="selectedUser.is_active ? 'success' : 'danger'">
              {{ selectedUser.is_active ? '活跃' : '非活跃' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(selectedUser.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="最后同步">{{ formatDate(selectedUser.last_sync_at) }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>

    <!-- 编辑角色对话框 -->
    <el-dialog v-model="editRoleVisible" title="编辑用户角色" width="400px">
      <div v-if="selectedUser">
        <el-form :model="roleForm" label-width="100px">
          <el-form-item label="用户">
            <el-text>{{ selectedUser.name }} ({{ selectedUser.username }})</el-text>
          </el-form-item>
          <el-form-item label="当前角色">
            <el-tag :type="getRoleTagType(selectedUser.dynamic_role || selectedUser.role)">
              {{ selectedUser.dynamic_role_name || selectedUser.role_name }}
            </el-tag>
          </el-form-item>
          <el-form-item label="新角色">
            <el-select v-model="roleForm.role" placeholder="选择角色">
              <el-option label="管理员" value="1" />
              <el-option label="教师" value="2" />
              <el-option label="学生" value="3" />
              <el-option label="访客" value="4" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="editRoleVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveRole">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'

// 状态管理
const authStore = useAuthStore()

// 响应式数据
const loading = ref(false)
const users = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

// 搜索和筛选
const searchKeyword = ref('')
const selectedRole = ref('')
const selectedStatus = ref('')

// 排序
const sortField = ref('')
const sortOrder = ref('')

// 对话框
const userDetailVisible = ref(false)
const editRoleVisible = ref(false)
const selectedUser = ref(null)

// 编辑角色表单
const roleForm = ref({
  role: ''
})

// 方法
const loadUsers = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      per_page: pageSize.value,
      search: searchKeyword.value,
      role: selectedRole.value,
      status: selectedStatus.value,
      sort_field: sortField.value,
      sort_order: sortOrder.value
    }
    
    // TODO: 实际API调用
    // const response = await api.getUsers(params)
    // users.value = response.data.users
    // total.value = response.data.total
    
    // 模拟数据
    users.value = [
      {
        id: 1,
        name: '张管理员',
        username: 'admin',
        email: 'admin@example.com',
        gitlab_id: 1,
        role: 1,
        role_name: '管理员',
        dynamic_role: 'admin',
        dynamic_role_name: '管理员',
        is_active: true,
        avatar: '',
        created_at: '2024-01-01T00:00:00Z',
        last_sync_at: '2024-01-15T10:00:00Z'
      },
      {
        id: 2,
        name: '李老师',
        username: 'teacher1',
        email: 'teacher1@example.com',
        gitlab_id: 2,
        role: 2,
        role_name: '教师',
        dynamic_role: 'teacher',
        dynamic_role_name: '教师',
        is_active: true,
        avatar: '',
        created_at: '2024-01-02T00:00:00Z',
        last_sync_at: '2024-01-15T09:30:00Z'
      },
      {
        id: 3,
        name: '王学生',
        username: 'student1',
        email: 'student1@example.com',
        gitlab_id: 3,
        role: 3,
        role_name: '学生',
        dynamic_role: 'student',
        dynamic_role_name: '学生',
        is_active: true,
        avatar: '',
        created_at: '2024-01-03T00:00:00Z',
        last_sync_at: '2024-01-15T08:00:00Z'
      }
    ]
    total.value = users.value.length
  } catch (error) {
    console.error('加载用户失败:', error)
    ElMessage.error('加载用户失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  currentPage.value = 1
  loadUsers()
}

const handleSortChange = ({ prop, order }) => {
  sortField.value = prop
  sortOrder.value = order === 'ascending' ? 'asc' : 'desc'
  loadUsers()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  loadUsers()
}

const handleCurrentChange = (page: number) => {
  currentPage.value = page
  loadUsers()
}

const handleSyncUsers = async () => {
  try {
    loading.value = true
    ElMessage.info('正在同步GitLab用户...')
    // TODO: 实际API调用
    await new Promise(resolve => setTimeout(resolve, 2000)) // 模拟延迟
    ElMessage.success('用户同步完成')
    loadUsers()
  } catch (error) {
    console.error('同步用户失败:', error)
    ElMessage.error('同步用户失败')
  } finally {
    loading.value = false
  }
}

const handleExport = () => {
  ElMessage.info('导出功能开发中...')
}

const handleViewUser = (user: any) => {
  selectedUser.value = user
  userDetailVisible.value = true
}

const handleEditRole = (user: any) => {
  selectedUser.value = user
  roleForm.value.role = user.role.toString()
  editRoleVisible.value = true
}

const handleSaveRole = async () => {
  try {
    const newRole = parseInt(roleForm.value.role)
    // TODO: 实际API调用
    // await api.updateUserRole(selectedUser.value.id, newRole)
    
    // 更新本地数据
    const userIndex = users.value.findIndex(u => u.id === selectedUser.value.id)
    if (userIndex !== -1) {
      users.value[userIndex].role = newRole
      users.value[userIndex].role_name = getRoleName(newRole)
    }
    
    ElMessage.success('角色更新成功')
    editRoleVisible.value = false
  } catch (error) {
    console.error('更新角色失败:', error)
    ElMessage.error('更新角色失败')
  }
}

const handleStatusChange = async (user: any) => {
  try {
    // TODO: 实际API调用
    // await api.updateUserStatus(user.id, user.is_active)
    ElMessage.success(`用户${user.is_active ? '激活' : '禁用'}成功`)
  } catch (error) {
    console.error('更新状态失败:', error)
    ElMessage.error('更新状态失败')
    // 回退状态
    user.is_active = !user.is_active
  }
}

// 工具方法
const isCurrentUser = (userId: number) => {
  return authStore.user?.id === userId
}

const getRoleTagType = (role: string | number) => {
  const roleStr = typeof role === 'number' ? getRoleFromStaticRole(role) : role
  switch (roleStr) {
    case 'admin': return 'danger'
    case 'teacher': return 'warning'
    case 'assistant': return 'info'
    case 'student': return 'success'
    case 'guest': return ''
    default: return ''
  }
}

const getRoleFromStaticRole = (role: number) => {
  switch (role) {
    case 1: return 'admin'
    case 2: return 'teacher'
    case 3: return 'student'
    case 4: return 'guest'
    default: return 'guest'
  }
}

const getRoleName = (role: number) => {
  switch (role) {
    case 1: return '管理员'
    case 2: return '教师'
    case 3: return '学生'
    case 4: return '访客'
    default: return '未知'
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

// 生命周期
onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.permission-users-view {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.header-left h1 {
  color: #303133;
  margin-bottom: 8px;
}

.header-left p {
  color: #606266;
  margin: 0;
}

.filters {
  margin-bottom: 20px;
}

.role-info {
  margin-top: 4px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

.user-detail {
  text-align: center;
}

.user-avatar {
  margin-bottom: 20px;
}

@media (max-width: 768px) {
  .permission-users-view {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .filters .el-col {
    margin-bottom: 12px;
  }
}
</style> 