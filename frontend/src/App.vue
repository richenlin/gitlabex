<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ApiService, type User } from './services/api'
import { useAuthStore } from './stores/auth'
import {
  Star,
  DataBoard,
  Document,
  User as UserIcon,
  ArrowDown,
  Setting,
  SwitchButton,
  School,
  Notebook,
  FolderOpened,
  TrendCharts
} from '@element-plus/icons-vue'
import NotificationBell from './components/NotificationBell.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 响应式数据
const currentUser = computed(() => authStore.user)

// 计算属性
const activeIndex = computed(() => {
  return route.path
})

// 是否显示导航栏（登录页面不显示）
const showNavigation = computed(() => {
  return route.path !== '/login'
})

// 是否为教师或管理员
const isTeacherOrAdmin = computed(() => {
  const userRole = authStore.userRole
  return userRole === 1 || userRole === 2 // 1: 管理员, 2: 教师
})

// 生命周期
onMounted(() => {
  // 路由守卫已经处理了用户认证，这里不需要再次加载
  // 如果需要，可以调用authStore.updateUserInfo()刷新用户信息
})

// 方法
const loadCurrentUser = async () => {
  try {
    await authStore.updateUserInfo()
    console.log('当前用户:', currentUser.value)
  } catch (error) {
    console.error('获取用户信息失败:', error)
    ElMessage.error('获取用户信息失败')
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
    // 调用后端退出登录API
    await ApiService.logout()
    
    // 清除本地状态
    authStore.logout()
    
    // 重定向到登录页面
    router.push('/login')
    
    ElMessage.success('已退出登录')
  } catch (error) {
    console.error('退出登录失败:', error)
    
    // 即使API调用失败，也要清除本地状态
    authStore.logout()
    router.push('/login')
    
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

          <!-- 公告通知 -->
          <div class="notification-section">
            <NotificationBell />
          </div>

          <!-- 用户菜单 -->
          <div class="user-section">
            <el-dropdown @command="handleUserMenuCommand">
              <span class="user-info">
                <el-avatar :size="32" :src="currentUser?.avatar">
                  <el-icon><UserIcon /></el-icon>
                </el-avatar>
                <span class="username">{{ currentUser?.name || '用户' }}</span>
                <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="profile">
                    <el-icon><UserIcon /></el-icon>
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
            <el-icon><DataBoard /></el-icon>
            <span>首页（仪表盘）</span>
          </el-menu-item>
          <el-menu-item index="/classes" v-if="isTeacherOrAdmin">
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
          <el-menu-item index="/analytics">
            <el-icon><TrendCharts /></el-icon>
            <span>统计分析</span>
          </el-menu-item>
          <el-menu-item index="/documents">
            <el-icon><Document /></el-icon>
            <span>文档管理</span>
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

.notification-section {
  display: flex;
  align-items: center;
  margin-right: 16px;
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
