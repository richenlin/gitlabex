<template>
  <div class="settings-container">
    <div class="settings-header">
      <h1>账户设置</h1>
      <p>管理您的SSH密钥和密码</p>
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
            <el-menu-item index="ssh-keys">
              <el-icon><Key /></el-icon>
              <span>SSH密钥</span>
            </el-menu-item>
            <el-menu-item index="password">
              <el-icon><Lock /></el-icon>
              <span>修改密码</span>
            </el-menu-item>
          </el-menu>
        </el-card>
      </el-col>

      <!-- 右侧内容区域 -->
      <el-col :span="18">
        <!-- SSH密钥管理 -->
        <el-card v-show="activeSection === 'ssh-keys'" class="settings-content">
          <template #header>
            <div class="card-header">
              <h3>SSH密钥管理</h3>
              <el-button type="primary" @click="showAddSSHKeyDialog = true">
                <el-icon><Plus /></el-icon>
                添加SSH密钥
              </el-button>
            </div>
          </template>

          <div class="ssh-keys-section">
            <!-- SSH密钥列表 -->
            <div v-if="sshKeys.length === 0" class="empty-state">
              <el-empty description="暂无SSH密钥">
                <el-button type="primary" @click="showAddSSHKeyDialog = true">
                  添加第一个SSH密钥
                </el-button>
              </el-empty>
            </div>

            <div v-else class="ssh-keys-list">
              <div
                v-for="key in sshKeys"
                :key="key.id"
                class="ssh-key-item"
              >
                <div class="key-info">
                  <div class="key-title">
                    <el-icon><Key /></el-icon>
                    <span>{{ key.title }}</span>
                  </div>
                  <div class="key-fingerprint">{{ key.fingerprint }}</div>
                  <div class="key-created">创建时间: {{ formatDate(key.created_at) }}</div>
                </div>
                <div class="key-actions">
                  <el-button type="danger" size="small" @click="deleteSSHKey(key.id)">
                    删除
                  </el-button>
                </div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 修改密码 -->
        <el-card v-show="activeSection === 'password'" class="settings-content">
          <template #header>
            <div class="card-header">
              <h3>修改密码</h3>
              <el-text type="info">更改您的登录密码</el-text>
            </div>
          </template>

          <div class="password-section">
            <el-form
              ref="passwordFormRef"
              :model="passwordForm"
              :rules="passwordRules"
              label-width="120px"
              class="password-form"
            >
              <el-form-item label="当前密码" prop="currentPassword">
                <el-input
                  v-model="passwordForm.currentPassword"
                  type="password"
                  placeholder="请输入当前密码"
                  show-password
                />
              </el-form-item>
              
              <el-form-item label="新密码" prop="newPassword">
                <el-input
                  v-model="passwordForm.newPassword"
                  type="password"
                  placeholder="请输入新密码"
                  show-password
                />
              </el-form-item>
              
              <el-form-item label="确认新密码" prop="confirmPassword">
                <el-input
                  v-model="passwordForm.confirmPassword"
                  type="password"
                  placeholder="请再次输入新密码"
                  show-password
                />
              </el-form-item>
              
              <el-form-item>
                <el-button
                  type="primary"
                  :loading="passwordLoading"
                  @click="changePassword"
                >
                  修改密码
                </el-button>
                <el-button @click="resetPasswordForm">
                  重置
                </el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 添加SSH密钥对话框 -->
    <el-dialog
      v-model="showAddSSHKeyDialog"
      title="添加SSH密钥"
      width="600px"
      :before-close="handleSSHKeyDialogClose"
    >
      <el-form
        ref="sshKeyFormRef"
        :model="sshKeyForm"
        :rules="sshKeyRules"
        label-width="100px"
      >
        <el-form-item label="标题" prop="title">
          <el-input
            v-model="sshKeyForm.title"
            placeholder="请输入密钥标题（如：My MacBook Pro）"
          />
        </el-form-item>
        
        <el-form-item label="公钥" prop="key">
          <el-input
            v-model="sshKeyForm.key"
            type="textarea"
            :rows="6"
            placeholder="请粘贴您的SSH公钥内容（以ssh-rsa、ssh-ed25519等开头）"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="handleSSHKeyDialogClose">取消</el-button>
          <el-button type="primary" :loading="sshKeyLoading" @click="addSSHKey">
            添加密钥
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Key,
  Lock,
  Plus
} from '@element-plus/icons-vue'
import { authService } from '@/services/api'

// 响应式数据
const activeSection = ref('ssh-keys')

// SSH密钥相关
const sshKeys = ref<any[]>([])
const showAddSSHKeyDialog = ref(false)
const sshKeyLoading = ref(false)

const sshKeyFormRef = ref()
const sshKeyForm = ref({
  title: '',
  key: ''
})

const sshKeyRules = {
  title: [
    { required: true, message: '请输入密钥标题', trigger: 'blur' },
    { min: 2, max: 50, message: '标题长度应在 2 到 50 个字符', trigger: 'blur' }
  ],
  key: [
    { required: true, message: '请输入SSH公钥', trigger: 'blur' },
    { 
      validator: (rule: any, value: string, callback: any) => {
        if (!value.startsWith('ssh-rsa') && !value.startsWith('ssh-ed25519') && !value.startsWith('ssh-dss')) {
          callback(new Error('请输入有效的SSH公钥格式'))
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    }
  ]
}

// 密码修改相关
const passwordLoading = ref(false)

const passwordFormRef = ref()
const passwordForm = ref({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const passwordRules = {
  currentPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 8, message: '密码长度至少为8个字符', trigger: 'blur' },
    { 
      pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
      message: '密码必须包含大小写字母和数字',
      trigger: 'blur'
    }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { 
      validator: (rule: any, value: string, callback: any) => {
        if (value !== passwordForm.value.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      }, 
      trigger: 'blur' 
    }
  ]
}

// 方法
const handleSectionChange = (section: string) => {
  activeSection.value = section
}

const fetchSSHKeys = async () => {
  try {
    console.log('获取SSH密钥列表...')
    const response = await authService.getSSHKeys()
    console.log('SSH密钥API响应:', response)
    
    // 由于响应拦截器已经返回了data，所以response就是密钥列表
    sshKeys.value = response?.data || response || []
  } catch (error) {
    console.error('获取SSH密钥失败:', error)
    ElMessage.error('获取SSH密钥失败')
  }
}

const addSSHKey = async () => {
  if (!sshKeyFormRef.value) return
  
  try {
    await sshKeyFormRef.value.validate()
    sshKeyLoading.value = true
    
    console.log('添加SSH密钥:', sshKeyForm.value)
    await authService.addSSHKey(sshKeyForm.value)
    
    ElMessage.success('SSH密钥添加成功')
    showAddSSHKeyDialog.value = false
    handleSSHKeyDialogClose()
    await fetchSSHKeys()
  } catch (error) {
    console.error('添加SSH密钥失败:', error)
    ElMessage.error('添加SSH密钥失败')
  } finally {
    sshKeyLoading.value = false
  }
}

const deleteSSHKey = async (keyId: number) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除这个SSH密钥吗？删除后将无法撤销。',
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    console.log('删除SSH密钥:', keyId)
    await authService.deleteSSHKey(keyId)
    
    ElMessage.success('SSH密钥删除成功')
    await fetchSSHKeys()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除SSH密钥失败:', error)
      ElMessage.error('删除SSH密钥失败')
    }
  }
}

const handleSSHKeyDialogClose = () => {
  showAddSSHKeyDialog.value = false
  sshKeyFormRef.value?.resetFields()
  sshKeyForm.value = {
    title: '',
    key: ''
  }
}

const changePassword = async () => {
  if (!passwordFormRef.value) return
  
  try {
    await passwordFormRef.value.validate()
    passwordLoading.value = true
    
    console.log('修改密码')
    await authService.changePassword({
      currentPassword: passwordForm.value.currentPassword,
      newPassword: passwordForm.value.newPassword
    })
    
    ElMessage.success('密码修改成功')
    resetPasswordForm()
  } catch (error) {
    console.error('修改密码失败:', error)
    ElMessage.error('修改密码失败')
  } finally {
    passwordLoading.value = false
  }
}

const resetPasswordForm = () => {
  passwordFormRef.value?.resetFields()
  passwordForm.value = {
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 生命周期
onMounted(() => {
  fetchSSHKeys()
})
</script>

<style scoped>
.settings-container {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.settings-header {
  margin-bottom: 30px;
}

.settings-header h1 {
  margin: 0 0 10px 0;
  color: #303133;
  font-size: 28px;
}

.settings-header p {
  margin: 0;
  color: #606266;
  font-size: 16px;
}

.settings-nav {
  height: fit-content;
}

.settings-menu {
  border: none;
}

.settings-menu .el-menu-item {
  height: 50px;
  line-height: 50px;
  border-radius: 6px;
  margin-bottom: 5px;
}

.settings-menu .el-menu-item.is-active {
  background-color: #ecf5ff;
  color: #409eff;
}

.settings-content {
  min-height: 400px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  color: #303133;
  font-size: 18px;
}

.ssh-keys-section,
.password-section {
  padding: 20px 0;
}

.empty-state {
  text-align: center;
  padding: 40px;
}

.ssh-keys-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.ssh-key-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  background: #fafafa;
}

.key-info {
  flex: 1;
}

.key-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 8px;
}

.key-fingerprint {
  font-family: monospace;
  font-size: 12px;
  color: #606266;
  margin-bottom: 4px;
}

.key-created {
  font-size: 12px;
  color: #909399;
}

.key-actions {
  flex-shrink: 0;
}

.password-form {
  max-width: 500px;
}

.dialog-footer {
  text-align: right;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .settings-container {
    padding: 15px;
  }
  
  .el-col:first-child {
    margin-bottom: 20px;
  }
  
  .settings-nav {
    margin-bottom: 20px;
  }
  
  .ssh-key-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }
  
  .key-actions {
    align-self: flex-end;
  }
}
</style>