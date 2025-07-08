<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <el-icon class="login-icon" :size="48" color="#409eff">
          <Lock />
        </el-icon>
        <h1 class="login-title">GitLabEx 教育平台</h1>
        <p class="login-subtitle">基于GitLab的教育增强平台</p>
      </div>

      <div class="login-content">
        <div class="login-description">
          <el-alert
            title="欢迎使用GitLabEx"
            type="info"
            :closable="false"
            show-icon
          >
            <template #default>
              <p>本平台通过GitLab OAuth2.0进行身份认证，</p>
              <p>将最大化复用GitLab的用户管理、团队协作功能。</p>
            </template>
          </el-alert>
        </div>

        <div class="login-form">
          <el-button 
            type="primary" 
            size="large" 
            :icon="Platform"
            @click="loginWithGitLab"
            :loading="isLoading"
            class="gitlab-login-btn"
          >
            使用 GitLab 登录
          </el-button>

          <div class="login-features">
            <h3>平台特色功能</h3>
            <ul class="feature-list">
              <li>
                <el-icon><Document /></el-icon>
                <span>OnlyOffice 在线文档协作</span>
              </li>
              <li>
                <el-icon><User /></el-icon>
                <span>GitLab 用户权限管理</span>
              </li>
              <li>
                <el-icon><ChatDotRound /></el-icon>
                <span>教育场景优化界面</span>
              </li>
              <li>
                <el-icon><TrendCharts /></el-icon>
                <span>学习进度跟踪分析</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <div class="login-footer">
        <p class="footer-text">
          © 2024 GitLabEx 教育平台 - 专注于教育场景的GitLab增强方案
        </p>
      </div>
    </div>

    <!-- 系统状态检查 -->
    <div class="system-status" v-if="systemStatus">
      <el-card class="status-card">
        <template #header>
          <span class="status-title">
            <el-icon><Monitor /></el-icon>
            系统状态
          </span>
        </template>
        <div class="status-grid">
          <div class="status-item">
            <el-tag :type="systemStatus.backend ? 'success' : 'danger'" size="small">
              后端服务: {{ systemStatus.backend ? '正常' : '异常' }}
            </el-tag>
          </div>
          <div class="status-item">
            <el-tag :type="systemStatus.gitlab ? 'success' : 'danger'" size="small">
              GitLab: {{ systemStatus.gitlab ? '正常' : '异常' }}
            </el-tag>
          </div>
          <div class="status-item">
            <el-tag :type="systemStatus.onlyoffice ? 'success' : 'danger'" size="small">
              OnlyOffice: {{ systemStatus.onlyoffice ? '正常' : '异常' }}
            </el-tag>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Lock,
  Document,
  User,
  ChatDotRound,
  TrendCharts,
  Monitor,
  Platform
} from '@element-plus/icons-vue'
import { ApiService } from '../services/api'

const router = useRouter()

// 响应式数据
const isLoading = ref(false)
const systemStatus = ref({
  backend: false,
  gitlab: false,
  onlyoffice: false
})

// 生命周期
onMounted(() => {
  checkSystemStatus()
  // 检查是否已经登录
  checkExistingAuth()
})

// 方法
const loginWithGitLab = async () => {
  isLoading.value = true
  try {
    // 获取GitLab OAuth URL
    const response = await fetch('http://localhost:8080/api/auth/gitlab')
    
    if (!response.ok) {
      throw new Error('Failed to get GitLab OAuth URL')
    }
    
    const data = await response.json()
    
    if (data.url) {
      // 跳转到GitLab OAuth页面
      window.location.href = data.url
    } else {
      throw new Error('No OAuth URL returned')
    }
  } catch (error) {
    console.error('GitLab login failed:', error)
    ElMessage.error('GitLab登录失败，请检查配置或稍后重试')
  } finally {
    isLoading.value = false
  }
}

const checkSystemStatus = async () => {
  try {
    // 检查后端状态
    const backendResponse = await fetch('http://localhost:8080/api/health')
    systemStatus.value.backend = backendResponse.ok

    // 检查GitLab状态（通过后端代理检查）
    try {
      const gitlabResponse = await fetch('http://localhost:8080/api/auth/gitlab')
      systemStatus.value.gitlab = gitlabResponse.ok
    } catch {
      systemStatus.value.gitlab = false
    }

    // 检查OnlyOffice状态
    try {
      const onlyofficeResponse = await fetch('http://localhost:8000/healthcheck')
      systemStatus.value.onlyoffice = onlyofficeResponse.ok
    } catch {
      systemStatus.value.onlyoffice = false
    }
  } catch (error) {
    console.error('System status check failed:', error)
  }
}

const checkExistingAuth = () => {
  // 检查本地存储的token
  const token = localStorage.getItem('authToken')
  if (token) {
    // TODO: 验证token有效性
    // 如果有效，直接跳转到仪表板
    // router.push('/dashboard')
  }

  // 检查URL参数，处理OAuth回调
  const urlParams = new URLSearchParams(window.location.search)
  const code = urlParams.get('code')
  const state = urlParams.get('state')
  
  if (code) {
    handleOAuthCallback(code, state)
  }
}

const handleOAuthCallback = async (code: string, state: string | null) => {
  try {
    // 这里应该通过后端处理OAuth回调
    // 由于我们在登录页面，GitLab会重定向到后端的callback URL
    // 这个函数主要用于处理可能的前端回调场景
    ElMessage.info('正在处理GitLab认证回调...')
  } catch (error) {
    console.error('OAuth callback handling failed:', error)
    ElMessage.error('认证回调处理失败')
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
  position: relative;
}

.login-card {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  width: 100%;
  max-width: 450px;
  min-height: 600px;
  display: flex;
  flex-direction: column;
}

.login-header {
  background: linear-gradient(135deg, #409eff 0%, #36c 100%);
  color: white;
  padding: 40px 30px;
  text-align: center;
}

.login-icon {
  margin-bottom: 16px;
}

.login-title {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
}

.login-subtitle {
  margin: 0;
  opacity: 0.9;
  font-size: 14px;
}

.login-content {
  flex: 1;
  padding: 30px;
}

.login-description {
  margin-bottom: 30px;
}

.login-form {
  text-align: center;
}

.gitlab-login-btn {
  width: 100%;
  height: 50px;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 30px;
  background: linear-gradient(135deg, #fc6d26 0%, #e24329 100%);
  border: none;
}

.gitlab-login-btn:hover {
  background: linear-gradient(135deg, #e24329 0%, #c7371e 100%);
}

.login-features {
  text-align: left;
}

.login-features h3 {
  margin: 0 0 16px 0;
  color: #2c3e50;
  font-size: 16px;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.feature-list li {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  color: #5f6368;
  font-size: 14px;
}

.feature-list li .el-icon {
  color: #409eff;
  font-size: 16px;
}

.login-footer {
  background: #f8f9fa;
  padding: 20px 30px;
  text-align: center;
  border-top: 1px solid #e4e7ed;
}

.footer-text {
  margin: 0;
  color: #7f8c8d;
  font-size: 12px;
}

.system-status {
  position: absolute;
  top: 20px;
  right: 20px;
  min-width: 200px;
}

.status-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
}

.status-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #2c3e50;
}

.status-grid {
  display: grid;
  gap: 8px;
}

.status-item {
  display: flex;
  justify-content: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .login-container {
    padding: 10px;
  }
  
  .login-card {
    min-height: auto;
  }
  
  .system-status {
    position: static;
    margin-top: 20px;
    min-width: auto;
  }
}
</style> 