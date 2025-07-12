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
import { useRouter, useRoute } from 'vue-router'
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
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

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
    const data = await ApiService.getGitLabOAuthUrl()
    
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
    const backendResponse = await fetch('/api/health')
    systemStatus.value.backend = backendResponse.ok

    // 检查GitLab状态（通过后端代理检查）
    try {
      const gitlabResponse = await fetch('/api/auth/gitlab')
      systemStatus.value.gitlab = gitlabResponse.ok
    } catch {
      systemStatus.value.gitlab = false
    }

    // 检查OnlyOffice状态
    try {
      const onlyofficeResponse = await fetch('/onlyoffice/healthcheck')
      systemStatus.value.onlyoffice = onlyofficeResponse.ok
    } catch {
      systemStatus.value.onlyoffice = false
    }
  } catch (error) {
    console.error('System status check failed:', error)
  }
}

const checkExistingAuth = async () => {
  // 检查本地存储的token并验证有效性
  const isLoggedIn = await authStore.checkAuth()
  
  if (isLoggedIn) {
    // 如果已经登录，跳转到目标页面或首页
    const redirect = route.query.redirect as string
    router.push(redirect || '/')
    return
  }

  // 检查URL参数，处理OAuth回调
  const code = route.query.code as string
  const state = route.query.state as string
  
  if (code) {
    await handleOAuthCallback(code, state)
  }
}

const handleOAuthCallback = async (code: string, state?: string) => {
  isLoading.value = true
  try {
    // 调用后端处理OAuth回调
    const data = await ApiService.handleOAuthCallback(code, state)
    
    // 登录成功
    const loginResult = await authStore.login(data.token)
    
    if (loginResult.success) {
      ElMessage.success('登录成功！')
      
      // 清除URL参数
      const cleanUrl = window.location.origin + window.location.pathname
      window.history.replaceState({}, '', cleanUrl)
      
      // 跳转到目标页面或首页
      const redirect = route.query.redirect as string
      router.push(redirect || '/')
    } else {
      throw new Error('登录失败')
    }
  } catch (error) {
    console.error('OAuth回调处理失败:', error)
    ElMessage.error('登录失败，请重试')
  } finally {
    isLoading.value = false
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
  padding: 40px;
  position: relative;
}

.login-card {
  background: white;
  border-radius: 20px;
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.12);
  overflow: hidden;
  width: 100%;
  max-width: 900px;
  min-height: 650px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: auto 1fr auto;
}

.login-header {
  background: linear-gradient(135deg, #409eff 0%, #36c 100%);
  color: white;
  padding: 60px 50px;
  text-align: center;
  grid-column: 1 / 3;
}

.login-icon {
  margin-bottom: 20px;
}

.login-title {
  margin: 0 0 12px 0;
  font-size: 32px;
  font-weight: 700;
  letter-spacing: -0.5px;
}

.login-subtitle {
  margin: 0;
  opacity: 0.9;
  font-size: 18px;
  font-weight: 300;
}

.login-content {
  padding: 50px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.login-description {
  margin-bottom: 40px;
}

.login-form {
  text-align: center;
}

.gitlab-login-btn {
  width: 100%;
  height: 60px;
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 40px;
  background: linear-gradient(135deg, #fc6d26 0%, #e24329 100%);
  border: none;
  border-radius: 12px;
  transition: all 0.3s ease;
}

.gitlab-login-btn:hover {
  background: linear-gradient(135deg, #e24329 0%, #c7371e 100%);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(252, 109, 38, 0.3);
}

.login-features {
  text-align: left;
  padding: 50px;
  background: #f8f9fa;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.login-features h3 {
  margin: 0 0 30px 0;
  color: #2c3e50;
  font-size: 22px;
  font-weight: 600;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.feature-list li {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px 0;
  color: #5f6368;
  font-size: 16px;
  border-bottom: 1px solid #e9ecef;
}

.feature-list li:last-child {
  border-bottom: none;
}

.feature-list li .el-icon {
  color: #409eff;
  font-size: 20px;
  background: #e3f2fd;
  padding: 8px;
  border-radius: 8px;
}

.login-footer {
  background: #f1f3f4;
  padding: 30px 50px;
  text-align: center;
  border-top: 1px solid #e4e7ed;
  grid-column: 1 / 3;
}

.footer-text {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
}

.system-status {
  position: absolute;
  top: 30px;
  right: 30px;
  min-width: 250px;
}

.status-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(15px);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.status-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  color: #2c3e50;
  font-size: 16px;
}

.status-grid {
  display: grid;
  gap: 12px;
}

.status-item {
  display: flex;
  justify-content: center;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .login-card {
    max-width: 700px;
  }
  
  .login-header,
  .login-content,
  .login-features {
    padding: 40px;
  }
}

@media (max-width: 768px) {
  .login-container {
    padding: 20px;
  }
  
  .login-card {
    grid-template-columns: 1fr;
    max-width: 500px;
    min-height: auto;
  }
  
  .login-header {
    grid-column: 1;
    padding: 40px 30px;
  }
  
  .login-title {
    font-size: 24px;
  }
  
  .login-subtitle {
    font-size: 16px;
  }
  
  .login-content,
  .login-features {
    padding: 30px;
  }
  
  .login-features {
    background: white;
  }
  
  .system-status {
    position: static;
    margin-top: 20px;
    min-width: auto;
  }
}

@media (max-width: 480px) {
  .login-container {
    padding: 10px;
  }
  
  .login-header,
  .login-content,
  .login-features,
  .login-footer {
    padding: 20px;
  }
}
</style> 