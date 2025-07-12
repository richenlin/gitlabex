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
  SwitchButton,
  School,
  Notebook,
  FolderOpened,
  Bell,
  TrendCharts
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

// 响应式数据
const currentUser = ref<User | null>(null)

// 计算属性
const activeIndex = computed(() => {
  return route.path
})

// 是否显示导航栏（登录页面不显示）
const showNavigation = computed(() => {
  return route.path !== '/login'
})

// 生命周期
onMounted(() => {
  loadCurrentUser()
})

// 方法
const loadCurrentUser = async () => {
  try {
    // 调用真实的API获取当前用户信息
    const response = await ApiService.getCurrentUser()
    currentUser.value = response
    console.log('当前用户:', currentUser.value)
  } catch (error) {
    console.error('获取用户信息失败:', error)
    ElMessage.error('获取用户信息失败')
    
    // 如果API调用失败，使用默认用户作为后备
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
    <!-- 登录页面：全屏显示 -->
    <div v-if="!showNavigation" class="login-layout">
      <router-view />
    </div>
    
    <!-- 其他页面：带导航栏布局 -->
    <template v-else>
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

          <!-- 占位符，保持header布局 -->
          <div class="header-spacer"></div>

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
      <el-container class="app-body">
        <!-- 左侧菜单 -->
        <el-aside class="app-aside">
        <el-menu
          :default-active="activeIndex"
          class="sidebar-menu"
          @select="handleMenuSelect"
          background-color="#ffffff"
          text-color="#333333"
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
          <el-menu-item index="/classes">
            <el-icon><School /></el-icon>
            <span>班级管理</span>
          </el-menu-item>
          <el-menu-item index="/projects">
            <el-icon><FolderOpened /></el-icon>
            <span>课题管理</span>
          </el-menu-item>
          <el-menu-item index="/assignments">
            <el-icon><Notebook /></el-icon>
            <span>作业管理</span>
          </el-menu-item>
          <el-menu-item index="/learning-progress">
            <el-icon><DataBoard /></el-icon>
            <span>学习进度跟踪</span>
          </el-menu-item>
          <el-menu-item index="/notifications">
            <el-icon><Bell /></el-icon>
            <span>通知系统</span>
          </el-menu-item>
          <el-menu-item index="/education-reports">
            <el-icon><TrendCharts /></el-icon>
            <span>教育报表</span>
          </el-menu-item>
          <el-menu-item index="/documents">
            <el-icon><Document /></el-icon>
            <span>文档</span>
          </el-menu-item>
          <el-menu-item index="/wiki">
            <el-icon><FolderOpened /></el-icon>
            <span>Wiki文档</span>
          </el-menu-item>
          <el-menu-item index="/users">
            <el-icon><User /></el-icon>
            <span>用户</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

        <!-- 主内容区域 -->
        <el-main class="app-main">
          <router-view />
        </el-main>
      </el-container>

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
    </template>
  </el-container>
</template>

<style scoped>
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.login-layout {
  min-height: 100vh;
  width: 100%;
}

.app-header {
  background: #fff;
  padding: 0;
  height: 64px;
  border-bottom: 1px solid #e6e6e6;
  position: fixed;
  width: 100%;
  top: 0;
  z-index: 1000;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 24px;
}

.logo-section {
  display: flex;
  align-items: center;
  min-width: 200px;
}

.logo-link {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: #333;
  gap: 12px;
}

.logo-icon {
  color: #409EFF;
}

.logo-text {
  font-size: 20px;
  font-weight: 600;
  color: #333;
}

.header-spacer {
  flex: 1;
}

.app-body {
  flex: 1;
  margin-top: 64px;
  display: flex;
}

.app-aside {
  width: 256px;
  background-color: #fff;
  border-right: 1px solid #e6e6e6;
  position: fixed;
  top: 64px;
  bottom: 0;
  overflow-y: auto;
}

.sidebar-menu {
  height: calc(100vh - 64px);
  border-right: none;
  padding-top: 16px;
}

.sidebar-menu .el-menu-item {
  height: 48px;
  line-height: 48px;
  margin: 4px 16px;
  border-radius: 4px;
  border: none;
}

.sidebar-menu .el-menu-item:hover {
  background-color: #f5f7fa;
}

.sidebar-menu .el-menu-item.is-active {
  background-color: #ecf5ff;
  color: #409EFF;
}

.sidebar-menu .el-menu-item .el-icon {
  margin-right: 12px;
  font-size: 18px;
}

.user-section {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  color: #333;
  padding: 8px 16px;
  border-radius: 4px;
  transition: all 0.3s;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.username {
  font-size: 14px;
  color: #333;
}

.dropdown-icon {
  font-size: 12px;
  color: #909399;
}

.app-main {
  flex: 1;
  margin-left: 256px;
  padding: 24px;
  background-color: #f5f7fa;
  min-height: calc(100vh - 124px);
}

.app-footer {
  background-color: #fff;
  border-top: 1px solid #e6e6e6;
  height: 60px;
  padding: 0;
  margin-left: 256px;
}

.footer-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 24px;
}

.footer-info {
  font-size: 14px;
  color: #909399;
}

.footer-links {
  display: flex;
  align-items: center;
  gap: 16px;
}

.footer-links .el-link {
  font-size: 14px;
  color: #606266;
  }
  
/* 移除移动端适配 */
</style>
