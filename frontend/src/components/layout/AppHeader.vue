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
        
        <!-- 系统公告按钮 -->
        <el-button 
          text 
          class="announcement-btn"
          @click="showAnnouncements = true"
        >
          <el-icon><InfoFilled /></el-icon>
        </el-button>
        
        <!-- 登录用户显示用户信息 -->
        <el-dropdown v-if="userStore.isLoggedIn" @command="handleUserAction">
          <div class="user-info">
            <span class="username">{{ userStore.user?.name || '用户名' }}</span>
            <el-avatar :size="32" :src="userStore.user?.avatar_url || defaultAvatar" />
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人资料</el-dropdown-item>
              <el-dropdown-item command="settings">设置</el-dropdown-item>
              <el-dropdown-item command="notifications">通知</el-dropdown-item>
              <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        
        <!-- 游客显示登录/注册按钮 -->
        <div v-else class="guest-actions">
          <el-button size="small" @click="$router.push('/auth/login')">
            登录
          </el-button>
          <el-button size="small" type="primary" @click="$router.push('/auth/register')">
            注册
          </el-button>
        </div>
      </div>
    </div>
  </header>

  <!-- 公告弹窗 -->
  <el-dialog
    v-model="showAnnouncements"
    title="系统公告"
    width="600px"
  >
    <div class="announcements-list">
      <div v-for="announcement in announcements" :key="announcement.id" class="announcement-item">
        <h4>{{ announcement.title }}</h4>
        <p>{{ announcement.content }}</p>
        <span class="announcement-time">{{ announcement.createdAt }}</span>
      </div>
    </div>
  </el-dialog>
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

const showAnnouncements = ref(false)

const defaultAvatar = 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png'

const announcements = [
  {
    id: 1,
    title: '系统维护通知',
    content: '系统将于今晚22:00-24:00进行维护升级，请提前做好保存工作。',
    createdAt: '2024-01-15 09:00'
  },
  {
    id: 2,
    title: '新功能上线',
    content: '作业管理系统已升级，增加了批量评分功能。',
    createdAt: '2024-01-14 14:30'
  }
]

const handleUserAction = (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/user/profile')
      break
    case 'settings':
      router.push('/user/settings')
      break
    case 'notifications':
      router.push('/user/notifications')
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
  background-color: rgba(26, 26, 74, 0.9);
  box-shadow: 0 2px 15px rgba(77, 121, 255, 0.3);
  backdrop-filter: blur(5px);
  border-bottom: 1px solid rgba(77, 121, 255, 0.2);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
}

.logo h1 {
  font-size: 24px;
  color: var(--primary-color);
  text-shadow: 0 0 10px rgba(77, 121, 255, 0.5);
  margin: 0;
}

.main-nav {
  display: flex;
  gap: 30px;
}

.nav-link {
  color: var(--text-color);
  font-weight: 500;
  padding: 8px 16px;
  border-radius: 4px;
  transition: all 0.3s ease;
  position: relative;
}

.nav-link:hover {
  color: var(--accent-color);
  background-color: rgba(77, 121, 255, 0.1);
  text-decoration: none;
}

.nav-link.active {
  background-color: rgba(77, 121, 255, 0.2);
  color: var(--primary-color);
  box-shadow: 0 0 15px rgba(77, 121, 255, 0.3);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.announcement-btn {
  color: var(--text-color);
  font-size: 20px;
}

.announcement-btn:hover {
  color: var(--accent-color);
  transform: scale(1.1);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: rgba(77, 121, 255, 0.1);
}

.username {
  color: var(--text-color);
  font-size: 14px;
}

.guest-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.announcements-list {
  max-height: 400px;
  overflow-y: auto;
}

.announcement-item {
  padding: 16px 0;
  border-bottom: 1px solid var(--border-color);
}

.announcement-item:last-child {
  border-bottom: none;
}

.announcement-item h4 {
  margin: 0 0 8px 0;
  color: var(--primary-color);
}

.announcement-item p {
  margin: 0 0 8px 0;
  color: var(--light-text);
  line-height: 1.5;
}

.announcement-time {
  font-size: 12px;
  color: var(--lighter-text);
}

@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
  }
  
  .main-nav {
    gap: 16px;
  }
  
  .nav-link {
    padding: 6px 12px;
    font-size: 14px;
  }
  
  .logo h1 {
    font-size: 20px;
  }
}
</style>