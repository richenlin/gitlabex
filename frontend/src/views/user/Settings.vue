<template>
  <div class="settings-container">
    <div class="settings-header">
      <h1>账户设置</h1>
      <p>管理您的账户信息和偏好设置</p>
    </div>

    <el-row :gutter="20">
      <!-- 左侧导航 -->
      <el-col :span="6">
        <el-card class="settings-nav">
          <el-menu
            v-model="activeSection"
            class="settings-menu"
            :default-active="activeSection"
            @select="handleSectionChange"
          >
            <el-menu-item index="profile">
              <el-icon><User /></el-icon>
              <span>个人信息</span>
            </el-menu-item>
            <el-menu-item index="account">
              <el-icon><Key /></el-icon>
              <span>账户安全</span>
            </el-menu-item>
            <el-menu-item index="notifications">
              <el-icon><Bell /></el-icon>
              <span>通知设置</span>
            </el-menu-item>
            <el-menu-item index="preferences">
              <el-icon><Setting /></el-icon>
              <span>偏好设置</span>
            </el-menu-item>
          </el-menu>
        </el-card>
      </el-col>

      <!-- 右侧内容区域 -->
      <el-col :span="18">
        <!-- 个人信息 -->
        <el-card v-show="activeSection === 'profile'" class="settings-content">
          <template #header>
            <div class="card-header">
              <h3>个人信息</h3>
              <el-text type="info">更新您的基本信息</el-text>
            </div>
          </template>
          
          <el-form
            ref="profileFormRef"
            :model="profileForm"
            :rules="profileFormRules"
            label-width="120px"
            v-loading="profileLoading"
          >
            <el-form-item label="头像">
              <div class="avatar-section">
                <el-avatar :size="100" :src="profileForm.avatar_url">
                  <el-icon><User /></el-icon>
                </el-avatar>
                <div class="avatar-actions">
                  <el-input
                    v-model="profileForm.avatar_url"
                    placeholder="请输入头像URL"
                    clearable
                  />
                  <el-text type="info" size="small">
                    支持 JPG、PNG 格式，建议尺寸 200x200 像素
                  </el-text>
                </div>
              </div>
            </el-form-item>
            
            <el-form-item label="用户名" prop="username">
              <el-input v-model="profileForm.username" disabled>
                <template #suffix>
                  <el-text type="info" size="small">用户名不可修改</el-text>
                </template>
              </el-input>
            </el-form-item>
            
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="profileForm.email" disabled>
                <template #suffix>
                  <el-text type="info" size="small">邮箱与GitLab同步</el-text>
                </template>
              </el-input>
            </el-form-item>
            
            <el-form-item label="姓名" prop="name">
              <el-input v-model="profileForm.name" placeholder="请输入姓名" />
            </el-form-item>
            
            <el-form-item label="角色">
              <el-tag :type="getRoleColor(profileForm.role)" size="large">
                {{ getRoleText(profileForm.role) }}
              </el-tag>
              <el-text type="info" size="small" style="margin-left: 10px;">
                角色由系统管理员分配
              </el-text>
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="handleSaveProfile" :loading="profileLoading">
                保存更改
              </el-button>
              <el-button @click="handleResetProfile">重置</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 账户安全 -->
        <el-card v-show="activeSection === 'account'" class="settings-content">
          <template #header>
            <div class="card-header">
              <h3>账户安全</h3>
              <el-text type="info">管理您的账户安全设置</el-text>
            </div>
          </template>
          
          <div class="security-section">
            <div class="security-item">
              <div class="security-info">
                <h4>GitLab 账户绑定</h4>
                <p>您的账户已与 GitLab 绑定，登录状态与 GitLab 同步</p>
              </div>
              <div class="security-status">
                <el-tag type="success">已绑定</el-tag>
              </div>
            </div>
            
            <el-divider />
            
            <div class="security-item">
              <div class="security-info">
                <h4>访问令牌</h4>
                <p>管理您的 API 访问令牌，用于第三方应用访问</p>
              </div>
              <div class="security-actions">
                <el-button type="primary" @click="showTokenDialog = true">
                  查看令牌信息
                </el-button>
              </div>
            </div>
            
            <el-divider />
            
            <div class="security-item">
              <div class="security-info">
                <h4>登录历史</h4>
                <p>查看最近的登录活动</p>
                <el-text type="info" size="small">
                  上次登录: {{ formatDate(userInfo?.last_login_at) }}
                </el-text>
              </div>
              <div class="security-actions">
                <el-button @click="fetchLoginHistory">查看历史</el-button>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 通知设置 -->
        <el-card v-show="activeSection === 'notifications'" class="settings-content">
          <template #header>
            <div class="card-header">
              <h3>通知设置</h3>
              <el-text type="info">管理您希望接收的通知类型</el-text>
            </div>
          </template>
          
          <el-form :model="notificationSettings" label-width="200px">
            <div class="notification-group">
              <h4>系统通知</h4>
              <el-form-item label="系统维护通知">
                <el-switch v-model="notificationSettings.system.maintenance" />
                <el-text type="info" size="small" style="margin-left: 10px;">
                  接收系统维护和更新通知
                </el-text>
              </el-form-item>
              <el-form-item label="安全提醒">
                <el-switch v-model="notificationSettings.system.security" />
                <el-text type="info" size="small" style="margin-left: 10px;">
                  接收账户安全相关提醒
                </el-text>
              </el-form-item>
            </div>
            
            <el-divider />
            
            <div class="notification-group">
              <h4>课题通知</h4>
              <el-form-item label="新课题创建">
                <el-switch v-model="notificationSettings.project.created" />
              </el-form-item>
              <el-form-item label="课题更新">
                <el-switch v-model="notificationSettings.project.updated" />
              </el-form-item>
              <el-form-item label="成员变更">
                <el-switch v-model="notificationSettings.project.memberChanged" />
              </el-form-item>
            </div>
            
            <el-divider />
            
            <div class="notification-group">
              <h4>作业通知</h4>
              <el-form-item label="新作业发布">
                <el-switch v-model="notificationSettings.homework.created" />
              </el-form-item>
              <el-form-item label="作业截止提醒">
                <el-switch v-model="notificationSettings.homework.deadline" />
              </el-form-item>
              <el-form-item label="作业评分">
                <el-switch v-model="notificationSettings.homework.graded" />
              </el-form-item>
            </div>
            
            <el-form-item style="margin-top: 30px;">
              <el-button type="primary" @click="handleSaveNotificationSettings" :loading="notificationLoading">
                保存设置
              </el-button>
              <el-button @click="handleResetNotificationSettings">重置为默认</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 偏好设置 -->
        <el-card v-show="activeSection === 'preferences'" class="settings-content">
          <template #header>
            <div class="card-header">
              <h3>偏好设置</h3>
              <el-text type="info">个性化您的使用体验</el-text>
            </div>
          </template>
          
          <el-form :model="preferences" label-width="150px">
            <div class="preference-group">
              <h4>界面设置</h4>
              <el-form-item label="主题模式">
                <el-radio-group v-model="preferences.theme">
                  <el-radio value="light">浅色主题</el-radio>
                  <el-radio value="dark">深色主题</el-radio>
                  <el-radio value="auto">跟随系统</el-radio>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="语言">
                <el-select v-model="preferences.language" style="width: 200px;">
                  <el-option label="简体中文" value="zh-CN" />
                  <el-option label="English" value="en-US" />
                </el-select>
              </el-form-item>
            </div>
            
            <el-divider />
            
            <div class="preference-group">
              <h4>显示设置</h4>
              <el-form-item label="每页显示数量">
                <el-select v-model="preferences.pageSize" style="width: 200px;">
                  <el-option label="10 条" :value="10" />
                  <el-option label="20 条" :value="20" />
                  <el-option label="50 条" :value="50" />
                  <el-option label="100 条" :value="100" />
                </el-select>
              </el-form-item>
              <el-form-item label="自动保存">
                <el-switch v-model="preferences.autoSave" />
                <el-text type="info" size="small" style="margin-left: 10px;">
                  编辑内容时自动保存草稿
                </el-text>
              </el-form-item>
            </div>
            
            <el-form-item style="margin-top: 30px;">
              <el-button type="primary" @click="handleSavePreferences" :loading="preferencesLoading">
                保存设置
              </el-button>
              <el-button @click="handleResetPreferences">重置为默认</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>

    <!-- 访问令牌对话框 -->
    <el-dialog
      v-model="showTokenDialog"
      title="访问令牌信息"
      width="600px"
    >
      <div class="token-info">
        <el-alert
          title="安全提示"
          type="warning"
          description="请妥善保管您的访问令牌，不要与他人分享。如果令牌泄露，请立即重新生成。"
          :closable="false"
          style="margin-bottom: 20px;"
        />
        
        <el-descriptions :column="1" border>
          <el-descriptions-item label="令牌状态">
            <el-tag type="success">有效</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatDate(userInfo?.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="过期时间">
            {{ formatDate(userInfo?.token_expiry) || '永不过期' }}
          </el-descriptions-item>
          <el-descriptions-item label="权限范围">
            <el-tag size="small" style="margin-right: 5px;">read_api</el-tag>
            <el-tag size="small" style="margin-right: 5px;">api</el-tag>
            <el-tag size="small">openid</el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showTokenDialog = false">关闭</el-button>
          <el-button type="danger" @click="handleRevokeToken">撤销令牌</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { authService } from '@/services/api'
import type { User } from '@/types'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  User as UserIcon,
  Key,
  Bell,
  Setting
} from '@element-plus/icons-vue'

const userStore = useUserStore()

// 响应式数据
const activeSection = ref('profile')
const userInfo = ref<User | null>(null)
const showTokenDialog = ref(false)

// 加载状态
const profileLoading = ref(false)
const notificationLoading = ref(false)
const preferencesLoading = ref(false)

// 表单引用
const profileFormRef = ref()

// 个人信息表单
const profileForm = ref({
  username: '',
  email: '',
  name: '',
  avatar_url: '',
  role: ''
})

const profileFormRules = {
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    { min: 2, max: 50, message: '姓名长度应在 2 到 50 个字符', trigger: 'blur' }
  ]
}

// 通知设置
const notificationSettings = ref({
  system: {
    maintenance: true,
    security: true
  },
  project: {
    created: true,
    updated: true,
    memberChanged: true
  },
  homework: {
    created: true,
    deadline: true,
    graded: true
  }
})

// 偏好设置
const preferences = ref({
  theme: 'light',
  language: 'zh-CN',
  pageSize: 20,
  autoSave: true
})

// 方法
const fetchUserInfo = async () => {
  try {
    const response = await authService.getCurrentUser()
    userInfo.value = response.data || response
    
    // 更新表单数据
    if (userInfo.value) {
      profileForm.value = {
        username: userInfo.value.username,
        email: userInfo.value.email,
        name: userInfo.value.name,
        avatar_url: userInfo.value.avatar_url || '',
        role: userInfo.value.role
      }
    }
  } catch (error) {
    console.error('获取用户信息失败:', error)
    ElMessage.error('获取用户信息失败')
  }
}

const handleSectionChange = (section: string) => {
  activeSection.value = section
}

const handleSaveProfile = async () => {
  if (!profileFormRef.value) return
  
  try {
    await profileFormRef.value.validate()
    profileLoading.value = true
    
    await authService.updateProfile({
      name: profileForm.value.name,
      avatar_url: profileForm.value.avatar_url
    })
    
    // 更新本地用户信息
    if (userInfo.value) {
      userInfo.value.name = profileForm.value.name
      userInfo.value.avatar_url = profileForm.value.avatar_url
    }
    
    // 更新store中的用户信息
    userStore.updateUser({
      name: profileForm.value.name,
      avatar_url: profileForm.value.avatar_url
    })
    
    ElMessage.success('个人信息更新成功')
  } catch (error) {
    console.error('更新个人信息失败:', error)
    ElMessage.error('更新个人信息失败')
  } finally {
    profileLoading.value = false
  }
}

const handleResetProfile = () => {
  if (userInfo.value) {
    profileForm.value = {
      username: userInfo.value.username,
      email: userInfo.value.email,
      name: userInfo.value.name,
      avatar_url: userInfo.value.avatar_url || '',
      role: userInfo.value.role
    }
  }
}

const handleSaveNotificationSettings = async () => {
  notificationLoading.value = true
  try {
    // 这里应该调用保存通知设置的API
    await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟API调用
    ElMessage.success('通知设置保存成功')
  } catch (error) {
    console.error('保存通知设置失败:', error)
    ElMessage.error('保存通知设置失败')
  } finally {
    notificationLoading.value = false
  }
}

const handleResetNotificationSettings = () => {
  notificationSettings.value = {
    system: {
      maintenance: true,
      security: true
    },
    project: {
      created: true,
      updated: true,
      memberChanged: true
    },
    homework: {
      created: true,
      deadline: true,
      graded: true
    }
  }
}

const handleSavePreferences = async () => {
  preferencesLoading.value = true
  try {
    // 这里应该调用保存偏好设置的API
    await new Promise(resolve => setTimeout(resolve, 1000)) // 模拟API调用
    ElMessage.success('偏好设置保存成功')
  } catch (error) {
    console.error('保存偏好设置失败:', error)
    ElMessage.error('保存偏好设置失败')
  } finally {
    preferencesLoading.value = false
  }
}

const handleResetPreferences = () => {
  preferences.value = {
    theme: 'light',
    language: 'zh-CN',
    pageSize: 20,
    autoSave: true
  }
}

const fetchLoginHistory = () => {
  ElMessage.info('登录历史功能开发中...')
}

const handleRevokeToken = async () => {
  try {
    await ElMessageBox.confirm(
      '撤销令牌后，使用该令牌的所有应用将无法访问您的账户。确定要继续吗？',
      '撤销访问令牌',
      {
        type: 'warning',
        confirmButtonText: '确定撤销',
        cancelButtonText: '取消'
      }
    )
    
    // 这里应该调用撤销令牌的API
    ElMessage.success('访问令牌已撤销')
    showTokenDialog.value = false
  } catch (error) {
    if (error !== 'cancel') {
      console.error('撤销令牌失败:', error)
      ElMessage.error('操作失败')
    }
  }
}

// 工具方法
const getRoleColor = (role?: string) => {
  const colorMap: Record<string, string> = {
    admin: 'danger',
    teacher: 'warning',
    assistant: 'info',
    student: 'success',
    guest: ''
  }
  return colorMap[role || ''] || ''
}

const getRoleText = (role?: string) => {
  const textMap: Record<string, string> = {
    admin: '管理员',
    teacher: '教师',
    assistant: '助教',
    student: '学生',
    guest: '访客'
  }
  return textMap[role || ''] || role || ''
}

const formatDate = (date?: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 生命周期
onMounted(() => {
  fetchUserInfo()
})
</script>

<style scoped>
.settings-container {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.settings-header {
  margin-bottom: 30px;
}

.settings-header h1 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 32px;
}

.settings-header p {
  margin: 0;
  color: #606266;
  font-size: 16px;
}

.settings-nav .settings-menu {
  border: none;
}

.settings-nav .el-menu-item {
  margin-bottom: 4px;
  border-radius: 8px;
}

.settings-content {
  min-height: 600px;
}

.card-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.card-header h3 {
  margin: 0;
  color: #303133;
  font-size: 20px;
}

.avatar-section {
  display: flex;
  align-items: center;
  gap: 20px;
}

.avatar-actions {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.security-section {
  display: flex;
  flex-direction: column;
}

.security-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 0;
}

.security-info h4 {
  margin: 0 0 8px 0;
  color: #303133;
  font-size: 16px;
}

.security-info p {
  margin: 0 0 4px 0;
  color: #606266;
  line-height: 1.5;
}

.security-status,
.security-actions {
  flex-shrink: 0;
}

.notification-group,
.preference-group {
  margin-bottom: 30px;
}

.notification-group h4,
.preference-group h4 {
  margin: 0 0 20px 0;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.token-info {
  padding: 10px 0;
}

.dialog-footer {
  text-align: right;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .settings-container {
    padding: 10px;
  }
  
  .security-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }
  
  .avatar-section {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }
}
</style>
