<template>
  <div class="callback-page">
    <div class="callback-container">
      <div class="callback-content">
        <el-icon class="loading-icon" size="48">
          <Loading />
        </el-icon>
        <h2>正在处理登录...</h2>
        <p>请稍候，我们正在验证您的GitLab账号信息</p>
        
        <div v-if="error" class="error-message">
          <el-alert
            :title="error"
            type="error"
            :closable="false"
            show-icon
          />
          <div class="error-actions">
            <el-button @click="retryLogin">重新登录</el-button>
            <el-button type="primary" @click="goHome">返回首页</el-button>
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
import { Loading } from '@element-plus/icons-vue'
import { authService } from '@/services/api'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const error = ref('')

// 处理OAuth回调
const handleOAuthCallback = async () => {
  const code = route.query.code as string
  const state = route.query.state as string
  
  if (!code || !state) {
    error.value = '缺少必要的认证参数，请重新登录'
    return
  }
  
  try {
    // 发送授权码到后端进行验证和登录
    const response = await authService.gitLabCallback(code, state)
    
    const data = response.data || response
    if (data.token && data.user) {
      // 保存用户信息和令牌
      userStore.setUser(data.user)
      userStore.setToken(data.token)
      
      ElMessage.success('登录成功，欢迎回来！')
      
      // 重定向到目标页面
      const redirect = route.query.redirect as string || '/'
      router.push(redirect)
    } else {
      error.value = '登录失败，请重试'
    }
  } catch (err: any) {
    console.error('OAuth回调处理失败:', err)
    
    let errorMessage = '登录失败，请重试'
    if (err.response?.data?.error) {
      errorMessage = err.response.data.error
    } else if (err.message) {
      errorMessage = err.message
    }
    
    error.value = errorMessage
  }
}

// 重新登录
const retryLogin = () => {
  router.push('/auth/login')
}

// 返回首页
const goHome = () => {
  router.push('/')
}

// 组件挂载时处理回调
onMounted(() => {
  handleOAuthCallback()
})
</script>

<style scoped>
.callback-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--space-gradient);
  padding: 20px;
}

.callback-container {
  max-width: 500px;
  width: 100%;
  background-color: var(--card-background);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  overflow: hidden;
}

.callback-content {
  padding: 60px 40px;
  text-align: center;
}

.loading-icon {
  color: var(--primary-color);
  margin-bottom: 20px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.callback-content h2 {
  font-size: 24px;
  color: var(--text-color);
  margin-bottom: 12px;
}

.callback-content p {
  color: var(--light-text);
  font-size: 16px;
  margin-bottom: 30px;
}

.error-message {
  margin-top: 30px;
}

.error-actions {
  margin-top: 20px;
  display: flex;
  justify-content: center;
  gap: 12px;
}

@media (max-width: 480px) {
  .callback-content {
    padding: 40px 20px;
  }
  
  .error-actions {
    flex-direction: column;
  }
}
</style>
