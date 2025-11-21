<template>
  <header class="app-header">
    <div class="header-content">
      <div class="logo">
        <h1>协同创新社区</h1>
      </div>
      
      <nav class="main-nav">
        <RouterLink 
          v-for="item in navItems" 
          :key="item.path"
          :to="item.path" 
          class="nav-link"
          :class="{ active: $route.path === item.path }"
        >
          {{ item.label }}
        </RouterLink>
      </nav>
      
      <div class="header-actions">
        <!-- 实时通知面板 -->
        <NotificationPanel v-if="userStore.isLoggedIn" />
        
        <!-- 登录用户显示用户信息 -->
        <el-dropdown v-if="userStore.isLoggedIn" @command="handleUserAction">
          <div class="user-info">
            <span class="username">{{ userStore.user?.name || '用户名' }}</span>
            <el-avatar :size="32" :src="userStore.user?.avatar_url || defaultAvatar" />
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-if="userStore.isAdmin" command="user-management">用户管理</el-dropdown-item>
              <el-dropdown-item command="profile">个人资料</el-dropdown-item>
              <el-dropdown-item command="settings">设置</el-dropdown-item>
              <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        
        <!-- 游客显示登录按钮 -->
        <div v-else class="guest-actions">
          <el-button size="small" @click="$router.push('/auth/login')">
            登录
          </el-button>
        </div>
      </div>
    </div>
  </header>

</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import { Bell, Message, InfoFilled } from '@element-plus/icons-vue'
import NotificationPanel from '@/components/common/NotificationPanel.vue'
const userStore = useUserStore()
const router = useRouter()

const navItems = [
  { path: '/', label: '首页' },
  { path: '/scenes', label: '课题' },
  { path: '/topics', label: '话题' },
  { path: '/documents', label: '文档' }
]

const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

// 调试信息
console.log('AppHeader: userStore.user:', userStore.user)
console.log('AppHeader: userStore.isAdmin:', userStore.isAdmin)
console.log('AppHeader: userStore.user?.is_admin:', userStore.user?.is_admin)

const handleUserAction = (command: string) => {
  switch (command) {
    case 'user-management':
      router.push('/admin/users')
      break
    case 'profile':
      router.push('/profile')
      break
    case 'settings':
      router.push('/settings')
      break
    case 'logout':
      userStore.logout()
      ElMessage.success('已退出登录')
      router.push('/auth/login')
      break
  }
}

// 组件挂载时的初始化
onMounted(() => {
  // 可以在这里添加其他初始化逻辑
})
</script>

<style scoped>
.app-header {
  background-color: #1a1a4a; /* 恢复深色背景，增加对比度 */
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  position: sticky;
  top: 0;
  z-index: 100;
  color: #fff;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
}

.logo h1 {
  font-size: 20px;
  color: #fff; /* Logo 改为白色 */
  font-weight: 700;
  margin: 0;
  letter-spacing: 0.5px;
}

.main-nav {
  display: flex;
  gap: 40px;
}

.nav-link {
  color: rgba(255, 255, 255, 0.85); /* 文字改为白色 */
  font-weight: 500;
  font-size: 15px;
  padding: 20px 0;
  position: relative;
  text-decoration: none;
  transition: color 0.2s ease;
}

.nav-link:hover {
  color: #fff;
}

.nav-link.active {
  color: #fff;
  font-weight: 600;
}

/* 激活状态指示条 - 改为亮蓝色或白色 */
.nav-link::after {
  content: "";
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 3px;
  background-color: var(--primary-color); /* 保持品牌色，或者用 #fff */
  transform: scaleX(0);
  transition: transform 0.2s ease;
  transform-origin: center;
}

.nav-link.active::after {
  transform: scaleX(1);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.user-info:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.username {
  color: #fff; /* 用户名改为白色 */
  font-size: 14px;
  font-weight: 500;
}

/* Mobile Responsive */
@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
    height: 56px;
  }
  
  .main-nav {
    gap: 20px;
  }
  
  .nav-link {
    font-size: 14px;
  }
  
  .logo h1 {
    font-size: 18px;
  }
}
</style>