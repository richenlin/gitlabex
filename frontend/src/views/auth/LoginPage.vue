<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-form-wrapper">
        <div class="login-header">
          <h1>欢迎回来</h1>
          <p>登录协同创新社区，开始您的教育协作之旅</p>
        </div>
        
        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          class="login-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item prop="username">
            <el-input
              v-model="loginForm.username"
              placeholder="请输入用户名"
              prefix-icon="User"
              size="large"
            />
          </el-form-item>
          
          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="请输入密码"
              prefix-icon="Lock"
              show-password
              size="large"
            />
          </el-form-item>
          
          <el-form-item>
            <el-button
              type="primary"
              size="large"
              style="width: 100%"
              :loading="loading"
              @click="handleLogin"
            >
              登录
            </el-button>
          </el-form-item>
        </el-form>
        
        <div class="login-footer">
          <p>
            还没有账号？
            <router-link to="/register">立即注册</router-link>
          </p>
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
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loginFormRef = ref<FormInstance>()
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '长度在 6 到 20 个字符', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  try {
    await loginFormRef.value.validate()
    loading.value = true
    
    await userStore.login(loginForm.username, loginForm.password)
    
    ElMessage.success('登录成功')
    
    const redirect = route.query.redirect as string
    if (redirect) {
      router.push(redirect)
    } else {
      router.push('/')
    }
  } catch (error) {
    console.error('登录失败:', error)
  } finally {
    loading.value = false
  }
}
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

.login-form {
  margin-bottom: 20px;
}

.login-footer {
  text-align: center;
}

.login-footer p {
  color: var(--light-text);
  margin: 0;
}

.login-footer a {
  color: var(--primary-color);
  text-decoration: none;
}

.login-footer a:hover {
  text-decoration: underline;
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