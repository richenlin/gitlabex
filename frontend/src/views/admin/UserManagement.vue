<template>
  <div class="user-management">
    <div class="page-header">
      <h1>用户管理</h1>
      <p class="page-description">管理系统用户，包括用户信息、角色权限等</p>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="search-section">
        <el-input
          v-model="searchQuery"
          placeholder="搜索用户名、邮箱或姓名"
          clearable
          style="width: 300px"
          @input="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
      
      <div class="action-buttons">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          新增用户
        </el-button>
        <el-button @click="refreshUsers">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 用户列表 -->
    <el-card class="user-list-card">
      <el-table
        v-loading="loading"
        :data="filteredUsers"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="avatar_url" label="头像" width="80">
          <template #default="{ row }">
            <el-avatar :size="32" :src="row.avatar_url || defaultAvatar" />
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="name" label="姓名" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="is_admin" label="管理员" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_admin ? 'danger' : 'info'">
              {{ row.is_admin ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="editUser(row)">编辑</el-button>
            <el-button size="small" type="warning" @click="manageRoles(row)">角色</el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteUser(row)"
              :disabled="row.id === userStore.user?.id"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="totalUsers"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 创建/编辑用户对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingUser ? '编辑用户' : '新增用户'"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="userFormRef"
        :model="userForm"
        :rules="userFormRules"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="userForm.name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" type="email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item v-if="!editingUser" label="密码" prop="password">
          <el-input 
            v-model="userForm.password" 
            type="password" 
            placeholder="请输入密码"
            show-password
          />
          <div class="password-tips">
            <el-alert
              title="GitLab密码要求"
              type="info"
              :closable="false"
              show-icon
              style="margin-top: 8px;"
            >
              <template #default>
                <ul style="margin: 0; padding-left: 16px;">
                  <li>密码长度至少8个字符</li>
                  <li>必须包含大小写字母和数字</li>
                  <li>不能包含常用单词（如password、admin、user等）</li>
                  <li>不能包含常见字母组合（如123456、abcdef等）</li>
                </ul>
                <div style="margin-top: 8px;">
                  <strong>推荐格式：</strong>MyStr0ng!P@ss、T3st!Us3r#2024
                </div>
              </template>
            </el-alert>
          </div>
        </el-form-item>
        <el-form-item label="管理员" prop="is_admin">
          <el-switch v-model="userForm.is_admin" />
        </el-form-item>
        <el-form-item label="默认角色" prop="default_role">
          <el-select v-model="userForm.default_role" placeholder="选择默认角色">
            <el-option label="访客 (Guest)" value="guest" />
            <el-option label="普通用户 (Reporter)" value="reporter" />
            <el-option label="研究员 (Developer)" value="developer" />
            <el-option label="教师 (Maintainer)" value="maintainer" />
            <el-option label="管理员 (Owner)" value="owner" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="submitUser" :loading="submitting">
            {{ editingUser ? '更新' : '创建' }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 角色管理对话框 -->
    <el-dialog
      v-model="showRoleDialog"
      title="角色管理"
      width="500px"
    >
      <div v-if="selectedUser" class="role-management">
        <div class="user-info">
          <el-avatar :size="48" :src="selectedUser.avatar_url || defaultAvatar" />
          <div class="user-details">
            <h3>{{ selectedUser.name }}</h3>
            <p>{{ selectedUser.username }}</p>
          </div>
        </div>
        
        <el-divider />
        
        <div class="role-sections">
          <h4>全局角色</h4>
          <el-form :model="roleForm">
            <el-form-item label="管理员权限">
              <el-switch v-model="roleForm.is_admin" />
            </el-form-item>
          </el-form>
          
          <h4>项目角色</h4>
          <div class="project-roles">
            <el-table :data="projectRoles" size="small">
              <el-table-column prop="project_name" label="项目" />
              <el-table-column prop="role" label="角色" />
              <el-table-column label="操作" width="120">
                <template #default="{ row }">
                  <el-button size="small" @click="editProjectRole(row)">编辑</el-button>
                  <el-button size="small" type="danger" @click="removeProjectRole(row)">移除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showRoleDialog = false">关闭</el-button>
          <el-button type="primary" @click="saveRoles" :loading="savingRoles">保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, Refresh } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { userManagementService } from '@/services/api'
import type { User } from '@/types'
import { formatDate } from '@/utils/date'

const userStore = useUserStore()

// 响应式数据
const loading = ref(false)
const submitting = ref(false)
const savingRoles = ref(false)
const users = ref<User[]>([])
const filteredUsers = ref<User[]>([])
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const totalUsers = ref(0)

// 对话框状态
const showCreateDialog = ref(false)
const showRoleDialog = ref(false)
const editingUser = ref<User | null>(null)
const selectedUser = ref<User | null>(null)

// 表单数据
const userForm = ref({
  username: '',
  name: '',
  email: '',
  password: '',
  is_admin: false,
  default_role: 'reporter'
})

const roleForm = ref({
  is_admin: false
})

const projectRoles = ref([
  // 示例数据，实际应从API获取
  { project_id: '1', project_name: '示例项目', role: 'developer' }
])

// 表单验证规则
const userFormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码长度不能少于 8 个字符', trigger: 'blur' },
    { 
      validator: (rule: any, value: string, callback: Function) => {
        if (!value) {
          callback()
          return
        }
        
        // GitLab密码要求验证
        const commonPatterns = [
          /password/i,
          /admin/i,
          /user/i,
          /test/i,
          /123456/,
          /abcdef/,
          /qwerty/i,
          /asdfgh/i,
          /zxcvbn/i
        ]
        
        const hasCommonPattern = commonPatterns.some(pattern => pattern.test(value))
        if (hasCommonPattern) {
          callback(new Error('密码不能包含常用的单词和字母组合'))
          return
        }
        
        // 检查是否包含至少一个数字、一个大写字母、一个小写字母
        const hasNumber = /\d/.test(value)
        const hasUpper = /[A-Z]/.test(value)
        const hasLower = /[a-z]/.test(value)
        
        if (!hasNumber || !hasUpper || !hasLower) {
          callback(new Error('密码必须包含大小写字母和数字'))
          return
        }
        
        callback()
      },
      trigger: 'blur'
    }
  ]
}

const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

// 计算属性
const userFormRef = ref()

// 方法
const fetchUsers = async () => {
  loading.value = true
  try {
    const response = await userManagementService.getUsers({
      page: currentPage.value,
      pageSize: pageSize.value,
      search: searchQuery.value
    })
    
    // 处理API响应
    if (response && Array.isArray(response)) {
      users.value = response
      filteredUsers.value = response
      totalUsers.value = response.length
    } else if (response && (response as any).users) {
      users.value = (response as any).users
      filteredUsers.value = (response as any).users
      totalUsers.value = (response as any).total || (response as any).users.length
    } else {
      // 如果API还没有实现，使用模拟数据
      const mockUsers: User[] = [
        {
          id: 1,
          username: 'admin',
          email: 'admin@example.com',
          name: '管理员',
          is_admin: true,
          created_at: '2024-01-01T00:00:00Z'
        },
        {
          id: 2,
          username: 'teacher1',
          email: 'teacher1@example.com',
          name: '张老师',
          is_admin: false,
          created_at: '2024-01-02T00:00:00Z'
        },
        {
          id: 3,
          username: 'student1',
          email: 'student1@example.com',
          name: '李同学',
          is_admin: false,
          created_at: '2024-01-03T00:00:00Z'
        }
      ]
      
      users.value = mockUsers
      filteredUsers.value = mockUsers
      totalUsers.value = mockUsers.length
    }
  } catch (error) {
    console.error('获取用户列表失败:', error)
    ElMessage.error('获取用户列表失败')
    
    // 发生错误时使用模拟数据
    const mockUsers: User[] = [
      {
        id: 1,
        username: 'admin',
        email: 'admin@example.com',
        name: '管理员',
        is_admin: true,
        created_at: '2024-01-01T00:00:00Z'
      }
    ]
    
    users.value = mockUsers
    filteredUsers.value = mockUsers
    totalUsers.value = mockUsers.length
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  if (!searchQuery.value) {
    filteredUsers.value = users.value
    return
  }
  
  const query = searchQuery.value.toLowerCase()
  filteredUsers.value = users.value.filter(user => 
    user.username.toLowerCase().includes(query) ||
    user.name.toLowerCase().includes(query) ||
    user.email.toLowerCase().includes(query)
  )
}

const refreshUsers = () => {
  fetchUsers()
}

const editUser = (user: User) => {
  editingUser.value = user
  userForm.value = {
    username: user.username,
    name: user.name,
    email: user.email,
    password: '',
    is_admin: user.is_admin,
    default_role: 'reporter'
  }
  showCreateDialog.value = true
}

const deleteUser = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.name}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await userManagementService.deleteUser(user.id.toString())
    ElMessage.success('用户删除成功')
    await fetchUsers()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除用户失败:', error)
      ElMessage.error('删除用户失败')
    }
  }
}

const manageRoles = (user: User) => {
  selectedUser.value = user
  roleForm.value.is_admin = user.is_admin
  showRoleDialog.value = true
}

const editProjectRole = (role: any) => {
  // TODO: 实现项目角色编辑
  ElMessage.info('项目角色编辑功能待实现')
}

const removeProjectRole = (role: any) => {
  // TODO: 实现项目角色移除
  ElMessage.info('项目角色移除功能待实现')
}

const submitUser = async () => {
  try {
    await userFormRef.value?.validate()
    submitting.value = true
    
    if (editingUser.value) {
      // 更新用户
      await userManagementService.updateUser(editingUser.value.id.toString(), {
        username: userForm.value.username,
        name: userForm.value.name,
        email: userForm.value.email,
        is_admin: userForm.value.is_admin
      })
      ElMessage.success('用户更新成功')
    } else {
      // 创建用户
      await userManagementService.createUser({
        username: userForm.value.username,
        name: userForm.value.name,
        email: userForm.value.email,
        password: userForm.value.password,
        is_admin: userForm.value.is_admin,
        default_role: userForm.value.default_role
      })
      ElMessage.success('用户创建成功')
    }
    
    showCreateDialog.value = false
    await fetchUsers()
  } catch (error) {
    console.error('提交用户信息失败:', error)
    ElMessage.error('提交用户信息失败')
  } finally {
    submitting.value = false
  }
}

const saveRoles = async () => {
  try {
    savingRoles.value = true
    
    if (selectedUser.value) {
      await userManagementService.updateUserRoles(selectedUser.value.id.toString(), {
        is_admin: roleForm.value.is_admin
      })
      ElMessage.success('角色保存成功')
      showRoleDialog.value = false
      await fetchUsers()
    }
  } catch (error) {
    console.error('保存角色失败:', error)
    ElMessage.error('保存角色失败')
  } finally {
    savingRoles.value = false
  }
}

const resetForm = () => {
  editingUser.value = null
  userForm.value = {
    username: '',
    name: '',
    email: '',
    password: '',
    is_admin: false,
    default_role: 'reporter'
  }
  userFormRef.value?.resetFields()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  fetchUsers()
}

const handleCurrentChange = (page: number) => {
  currentPage.value = page
  fetchUsers()
}

// 生命周期
onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.user-management {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0 0 8px 0;
  color: var(--primary-color);
  font-size: 28px;
}

.page-description {
  margin: 0;
  color: var(--light-text);
  font-size: 14px;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px;
  background: var(--card-bg);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.search-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.action-buttons {
  display: flex;
  gap: 12px;
}

.user-list-card {
  margin-bottom: 20px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.role-management {
  .user-info {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 16px;
  }
  
  .user-details h3 {
    margin: 0 0 4px 0;
    color: var(--primary-color);
  }
  
  .user-details p {
    margin: 0;
    color: var(--light-text);
    font-size: 14px;
  }
  
  .role-sections h4 {
    margin: 16px 0 12px 0;
    color: var(--primary-color);
  }
  
  .project-roles {
    margin-top: 12px;
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
