<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ApiService, type User } from './services/api'
import {
  Star,
  House,
  DataBoard,
  Document,
  User as UserIcon,
  ArrowDown,
  Setting,
  SwitchButton
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

// 响应式数据
const currentUser = ref<User | null>(null)

// 计算属性
const activeIndex = computed(() => {
  return route.path
})

// 生命周期
onMounted(() => {
  loadCurrentUser()
})

// 方法
const loadCurrentUser = async () => {
  try {
    // 这里应该调用获取当前用户信息的API
    // const response = await ApiService.getCurrentUser()
    // currentUser.value = response.data
    
    // 模拟数据
    currentUser.value = {
      id: 1,
      name: '演示用户',
      username: 'demo',
      email: 'demo@example.com',
      avatar: '',
      role: 2,
      gitlab_id: 1,
      is_active: true,
      last_sync_at: new Date().toISOString()
    }
  } catch (error) {
    console.error('获取用户信息失败:', error)
  }
}

const handleMenuSelect = (index: string) => {
  router.push(index)
}

const handleUserMenuCommand = (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      ElMessage.info('设置功能正在开发中')
      break
    case 'logout':
      logout()
      break
  }
}

const logout = async () => {
  try {
    // 这里应该调用退出登录API
    // await ApiService.logout()
    
    // 清除本地存储
    localStorage.removeItem('authToken')
    
    // 重定向到登录页面
    router.push('/login')
    
    ElMessage.success('已退出登录')
  } catch (error) {
    console.error('退出登录失败:', error)
    ElMessage.error('退出登录失败')
  }
}
</script>

<template>
  <el-container class="app-container">
    <!-- 顶部导航栏 -->
    <el-header class="app-header">
      <div class="header-content">
        <!-- Logo和标题 -->
        <div class="logo-section">
          <router-link to="/" class="logo-link">
            <el-icon class="logo-icon" size="32"><Star /></el-icon>
            <span class="logo-text">GitLabEx</span>
          </router-link>
        </div>

        <!-- 导航菜单 -->
        <el-menu
          :default-active="activeIndex"
          class="header-menu"
          mode="horizontal"
          @select="handleMenuSelect"
          background-color="transparent"
          text-color="#fff"
          active-text-color="#409EFF"
        >
          <el-menu-item index="/">
            <el-icon><House /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="/dashboard">
            <el-icon><DataBoard /></el-icon>
            <span>仪表板</span>
          </el-menu-item>
          <el-menu-item index="/documents">
            <el-icon><Document /></el-icon>
            <span>文档</span>
          </el-menu-item>
          <el-menu-item index="/users">
            <el-icon><User /></el-icon>
            <span>用户</span>
          </el-menu-item>
        </el-menu>

        <!-- 用户菜单 -->
        <div class="user-section">
          <el-dropdown @command="handleUserMenuCommand">
            <span class="user-info">
              <el-avatar :size="32" :src="currentUser?.avatar">
                <el-icon><User /></el-icon>
              </el-avatar>
              <span class="username">{{ currentUser?.name || '用户' }}</span>
              <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">
                  <el-icon><User /></el-icon>
                  个人资料
                </el-dropdown-item>
                <el-dropdown-item command="settings">
                  <el-icon><Setting /></el-icon>
                  设置
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-header>

    <!-- 主内容区域 -->
    <el-main class="app-main">
      <router-view />
    </el-main>

    <!-- 底部 -->
    <el-footer class="app-footer">
      <div class="footer-content">
        <div class="footer-info">
          <span>&copy; 2024 GitLabEx - 基于 GitLab + OnlyOffice 的教育协作平台</span>
        </div>
        <div class="footer-links">
          <el-link href="/about" :underline="false">关于</el-link>
          <el-divider direction="vertical" />
          <el-link href="https://github.com" target="_blank" :underline="false">GitHub</el-link>
          <el-divider direction="vertical" />
          <el-link href="https://gitlab.com" target="_blank" :underline="false">GitLab</el-link>
        </div>
      </div>
    </el-footer>
  </el-container>
</template>

<style scoped>
.app-container {
  min-height: 100vh;
}

.app-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 0;
  height: 60px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.logo-section {
  display: flex;
  align-items: center;
}

.logo-link {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: white;
  gap: 8px;
}

.logo-icon {
  color: white;
}

.logo-text {
  font-size: 24px;
  font-weight: bold;
  color: white;
}

.header-menu {
  flex: 1;
  margin: 0 40px;
  border-bottom: none;
}

.header-menu .el-menu-item {
  border-bottom: none !important;
  color: rgba(255, 255, 255, 0.8);
}

.header-menu .el-menu-item:hover {
  color: #fff;
  background-color: rgba(255, 255, 255, 0.1);
}

.header-menu .el-menu-item.is-active {
  color: #409EFF;
  background-color: rgba(255, 255, 255, 0.1);
}

.user-section {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: white;
  padding: 8px 12px;
  border-radius: 6px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.username {
  font-size: 14px;
}

.dropdown-icon {
  font-size: 12px;
}

.app-main {
  padding: 0;
  background-color: #f5f7fa;
}

.app-footer {
  background-color: #fff;
  border-top: 1px solid #ebeef5;
  height: 60px;
  padding: 0;
}

.footer-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.footer-info {
  font-size: 14px;
  color: #909399;
}

.footer-links {
  display: flex;
  align-items: center;
  gap: 8px;
}

.footer-links .el-link {
  font-size: 14px;
  color: #909399;
}

@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
  }
  
  .header-menu {
    margin: 0 20px;
  }
  
  .username {
    display: none;
  }
  
  .footer-content {
    flex-direction: column;
    gap: 8px;
    padding: 16px;
  }
  
  .footer-info,
  .footer-links {
    font-size: 12px;
  }
}
</style>
