<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-form-wrapper">
        <div class="login-header">
          <h1>欢迎来到协同创新社区</h1>
          <p>使用GitLab账号登录，开始您的教育协作之旅</p>
        </div>
        
        <div class="oauth-login">
          <el-button
            type="primary"
            size="large"
            class="gitlab-login-btn"
            :loading="loading"
            @click="handleGitLabLogin"
          >
            <el-icon class="gitlab-icon"><Platform /></el-icon>
            使用 GitLab 登录
          </el-button>
          
          <div class="login-tips">
            <p class="tip-text">
              <el-icon><InfoFilled /></el-icon>
              使用您的GitLab账号即可直接登录，系统将自动同步您的基本信息
            </p>
          </div>
        </div>
      </div>
      
      <div class="login-info">
        <div class="info-content">
          <h2>协同创新社区</h2>
          <p>基于GitLab的教育协作平台，专为教学场景设计的现代化教育管理系统。</p>
          
          <div class="features">
            <div class="feature">
              <h3>🚀 高效协作</h3>
              <p>基于Git的版本控制，多人协作开发</p>
            </div>
            <div class="feature">
              <h3>📚 教学管理</h3>
              <p>课题管理、作业布置、文档分享</p>
            </div>
            <div class="feature">
              <h3>🔒 企业级安全</h3>
              <p>完善的权限控制和数据保护</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import { Platform, InfoFilled } from '@element-plus/icons-vue'
import { authService } from '@/services/api'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loading = ref(false)

// 处理GitLab OAuth登录
const handleGitLabLogin = async () => {
  try {
    loading.value = true
    
    // 调用后端API获取GitLab授权URL
    const response = await authService.getGitLabAuthUrl()
    const authUrl = response.auth_url
    
    if (authUrl) {
      // 跳转到GitLab授权页面
      window.location.href = authUrl
    } else {
      ElMessage.error('获取授权链接失败')
    }
  } catch (error) {
    console.error('GitLab登录失败:', error)
    ElMessage.error('登录失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

// 组件挂载时检查是否已登录
onMounted(() => {
  // 如果用户已登录，重定向到首页
  if (userStore.isLoggedIn) {
    router.push('/')
  }
})
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--space-gradient);
  padding: 20px;
}

.login-container {
  display: flex;
  max-width: 1000px;
  width: 100%;
  background-color: var(--card-background);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  overflow: hidden;
}

.login-form-wrapper {
  flex: 1;
  padding: 60px 40px;
  max-width: 400px;
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.login-header h1 {
  font-size: 28px;
  color: var(--primary-color);
  margin-bottom: 8px;
}

.login-header p {
  color: var(--light-text);
  font-size: 16px;
}

.oauth-login {
  text-align: center;
}

.gitlab-login-btn {
  width: 100%;
  height: 50px;
  font-size: 16px;
  font-weight: 600;
  background: linear-gradient(135deg, #fc6d26, #fca326);
  border: none;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.gitlab-login-btn:hover {
  background: linear-gradient(135deg, #e85d15, #eb9315);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(252, 109, 38, 0.3);
}

.gitlab-icon {
  margin-right: 8px;
  font-size: 18px;
}

.login-tips {
  margin-top: 30px;
  padding: 20px;
  background-color: rgba(77, 121, 255, 0.05);
  border-radius: 8px;
  border-left: 4px solid var(--primary-color);
}

.tip-text {
  color: var(--light-text);
  font-size: 14px;
  margin: 0;
  display: flex;
  align-items: center;
  line-height: 1.5;
}

.tip-text .el-icon {
  margin-right: 8px;
  color: var(--primary-color);
  font-size: 16px;
}

.login-info {
  flex: 1;
  background: linear-gradient(135deg, rgba(77, 121, 255, 0.1), rgba(0, 225, 255, 0.1));
  padding: 60px 40px;
  display: flex;
  align-items: center;
  border-left: 1px solid var(--border-color);
}

.info-content {
  text-align: center;
}

.info-content h2 {
  font-size: 32px;
  color: var(--primary-color);
  margin-bottom: 16px;
}

.info-content p {
  color: var(--light-text);
  font-size: 18px;
  margin-bottom: 40px;
  line-height: 1.6;
}

.features {
  text-align: left;
}

.feature {
  margin-bottom: 24px;
}

.feature h3 {
  font-size: 20px;
  color: var(--text-color);
  margin-bottom: 4px;
}

.feature p {
  font-size: 14px;
  color: var(--lighter-text);
  margin: 0;
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
    margin: 20px;
  }
  
  .login-form-wrapper,
  .login-info {
    padding: 40px 30px;
  }
  
  .login-info {
    border-left: none;
    border-top: 1px solid var(--border-color);
  }
}

@media (max-width: 480px) {
  .login-form-wrapper,
  .login-info {
    padding: 30px 20px;
  }
}
</style>