<template>
  <div class="login-success-container">
    <div class="login-success-card">
      <div v-if="loading" class="loading">
        <el-icon class="rotating"><Loading /></el-icon>
        <p>正在处理登录信息...</p>
      </div>
      <div v-else-if="error" class="error">
        <el-icon class="error-icon"><Warning /></el-icon>
        <h2>登录失败</h2>
        <p>{{ errorMessage }}</p>
        <el-button @click="backToLogin" type="primary">返回登录</el-button>
      </div>
      <div v-else class="success">
        <el-icon class="success-icon"><SuccessFilled /></el-icon>
        <h2>登录成功</h2>
        <p>正在跳转到首页...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Loading, Warning, SuccessFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(true)
const error = ref(false)
const errorMessage = ref('')

const errorMessages: { [key: string]: string } = {
  missing_code: '缺少认证代码，请重新登录',
  token_exchange_failed: '令牌交换失败，请重新登录',
  fetch_user_failed: '获取用户信息失败，请重新登录',
  sync_user_failed: '用户信息同步失败，请重新登录',
  jwt_generation_failed: 'JWT令牌生成失败，请重新登录',
  invalid_token: '无效的登录令牌，请重新登录'
}

onMounted(async () => {
  try {
    // 检查是否有错误参数
    const errorParam = route.query.error as string
    if (errorParam) {
      error.value = true
      errorMessage.value = errorMessages[errorParam] || '登录过程中发生未知错误'
      loading.value = false
      return
    }

    // 获取token参数
    const token = route.query.token as string
    if (!token) {
      error.value = true
      errorMessage.value = '未收到登录令牌，请重新登录'
      loading.value = false
      return
    }

    // 存储token到localStorage
    localStorage.setItem('authToken', token)
    
    // 更新auth store状态
    await authStore.updateUserInfo()
    
    loading.value = false
    
    // 显示成功消息
    ElMessage.success('登录成功！')
    
    // 延迟跳转到首页
    setTimeout(() => {
      router.push('/')
    }, 1000)
    
  } catch (err) {
    console.error('Login success handling error:', err)
    error.value = true
    errorMessage.value = '处理登录信息时发生错误'
    loading.value = false
  }
})

const backToLogin = () => {
  // 清除可能存在的无效token
  localStorage.removeItem('authToken')
  router.push('/login')
}
</script>

<style scoped>
.login-success-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f5f5f5;
}

.login-success-card {
  background: white;
  padding: 3rem;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  text-align: center;
  max-width: 400px;
  width: 100%;
}

.loading, .error, .success {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.rotating {
  animation: rotate 2s linear infinite;
  font-size: 2rem;
  color: #409eff;
}

.error-icon {
  font-size: 3rem;
  color: #f56c6c;
}

.success-icon {
  font-size: 3rem;
  color: #67c23a;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

h2 {
  margin: 0;
  color: #333;
}

p {
  margin: 0;
  color: #666;
}
</style> 