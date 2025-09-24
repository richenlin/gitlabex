import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { UserRole } from '@/types'
import type { RouteMeta } from '@/types'
import { ElMessage } from 'element-plus'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomePage.vue'),
      meta: { title: '首页', requiresAuth: false }
    },
    {
      path: '/auth/login',
      name: 'login',
      component: () => import('@/views/auth/LoginPage.vue'),
      meta: { title: '登录', requiresAuth: false }
    },
    {
      path: '/auth/gitlab/callback',
      name: 'auth-callback',
      component: () => import('@/views/auth/CallbackPage.vue'),
      meta: { title: '登录处理中', requiresAuth: false }
    },
    {
      path: '/login',
      redirect: '/auth/login'
    },
    {
      path: '/scenes',
      name: 'scenes',
      component: () => import('@/views/scenes/SceneList.vue'),
      meta: { title: '课题列表', requiresAuth: false }
    },
    {
      path: '/scenes/create',
      name: 'scene-create',
      component: () => import('@/views/scenes/SceneCreate.vue'),
      meta: { title: '创建课题', roles: [UserRole.TEACHER, UserRole.ASSISTANT] }
    },
    {
      path: '/scenes/:id',
      name: 'scene-detail',
      component: () => import('@/views/scenes/SceneDetail.vue'),
      meta: { title: '课题详情', requiresAuth: true }
    },
    {
      path: '/scenes/:id/edit',
      name: 'scene-edit',
      component: () => import('@/views/scenes/SceneEdit.vue'),
      meta: { title: '编辑课题', roles: [UserRole.TEACHER, UserRole.ASSISTANT] }
    },
    {
      path: '/topics',
      name: 'topics',
      component: () => import('@/views/topics/TopicList.vue'),
      meta: { title: '话题列表', requiresAuth: false }
    },
    {
      path: '/topics/create',
      name: 'topic-create',
      component: () => import('@/views/topics/TopicCreate.vue'),
      meta: { title: '创建话题', requiresAuth: true }
    },
    {
      path: '/topics/:id',
      name: 'topic-detail',
      component: () => import('@/views/topics/TopicDetail.vue'),
      meta: { title: '话题详情', requiresAuth: true }
    },
    {
      path: '/topics/:id/edit',
      name: 'topic-edit',
      component: () => import('@/views/topics/TopicEdit.vue'),
      meta: { title: '编辑话题', requiresAuth: true }
    },
    {
      path: '/documents',
      name: 'documents',
      component: () => import('@/views/documents/DocumentList.vue'),
      meta: { title: '文档列表', requiresAuth: false }
    },
    {
      path: '/documents/upload',
      name: 'document-upload',
      component: () => import('@/views/documents/DocumentUpload.vue'),
      meta: { title: '上传文档', requiresAuth: true }
    },
    {
      path: '/documents/:id',
      name: 'document-detail',
      component: () => import('@/views/documents/DocumentDetail.vue'),
      meta: { title: '文档详情', requiresAuth: false }
    },
    {
      path: '/homeworks',
      name: 'homeworks',
      component: () => import('@/views/homeworks/HomeworkList.vue'),
      meta: { title: '作业列表', requiresAuth: true }
    },
    {
      path: '/homeworks/create',
      name: 'homework-create',
      component: () => import('@/views/homeworks/HomeworkCreate.vue'),
      meta: { title: '创建作业', roles: [UserRole.TEACHER, UserRole.ASSISTANT] }
    },
    {
      path: '/homeworks/:id',
      name: 'homework-detail',
      component: () => import('@/views/homeworks/HomeworkDetail.vue'),
      meta: { title: '作业详情', requiresAuth: true }
    },
    {
      path: '/homeworks/:id/edit',
      name: 'homework-edit',
      component: () => import('@/views/homeworks/HomeworkCreate.vue'),
      meta: { title: '编辑作业', roles: [UserRole.TEACHER, UserRole.ASSISTANT] }
    },
    {
      path: '/homeworks/:id/submit',
      name: 'homework-submit',
      component: () => import('@/views/homeworks/HomeworkSubmit.vue'),
      meta: { title: '提交作业', requiresAuth: true }
    },
    {
      path: '/homeworks/:id/grade',
      name: 'homework-grade',
      component: () => import('@/views/homeworks/HomeworkGrade.vue'),
      meta: { title: '批改作业', roles: [UserRole.TEACHER, UserRole.ASSISTANT] }
    },
    {
      path: '/homeworks/:id/submissions',
      name: 'homework-submissions',
      component: () => import('@/views/homeworks/HomeworkSubmissions.vue'),
      meta: { title: '作业提交列表', roles: [UserRole.TEACHER, UserRole.ASSISTANT] }
    },
    {
      path: '/homeworks/submissions/:id/grade',
      name: 'submission-grade',
      component: () => import('@/views/homeworks/HomeworkGrade.vue'),
      meta: { title: '批改提交', roles: [UserRole.TEACHER, UserRole.ASSISTANT] }
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/views/user/Profile.vue'),
      meta: { title: '个人资料', requiresAuth: true }
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/user/Settings.vue'),
      meta: { title: '设置', requiresAuth: true }
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/admin/Dashboard.vue'),
      meta: { title: '管理后台', roles: [UserRole.ADMIN] }
    },
    {
      path: '/admin/users',
      name: 'user-management',
      component: () => import('@/views/admin/UserManagement.vue'),
      meta: { title: '用户管理', roles: [UserRole.ADMIN] }
    }
  ]
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  
  // 验证token有效性（如果有token但没有用户信息，尝试获取用户信息）
  if (userStore.token && !userStore.user) {
    try {
      const isValid = await userStore.validateToken()
      if (!isValid) {
        // token无效，清除状态
        userStore.logout()
      }
    } catch (error) {
      console.error('Token validation failed:', error)
      userStore.logout()
    }
  }

  // 检查是否需要登录 - 默认不需要登录，除非明确设置为true或有角色要求
  const requiresAuth = to.meta.requiresAuth === true || (to.meta.roles && Array.isArray(to.meta.roles) && to.meta.roles.length > 0)
  
  if (requiresAuth && !userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    next({
      name: 'login',
      query: { redirect: to.fullPath }
    })
    return
  }

  // 检查角色权限 - 简化版本，具体权限由页面组件通过API验证
  if (to.meta.roles && Array.isArray(to.meta.roles) && to.meta.roles.length > 0) {
    if (!userStore.isLoggedIn) {
      ElMessage.error('请先登录')
      next({ name: 'login', query: { redirect: to.fullPath } })
      return
    }
    
    // 对于需要特殊权限的路由，只检查管理员权限
    // 其他权限检查将在页面组件中通过API进行
    const requiresAdmin = to.meta.roles.includes(UserRole.ADMIN)
    if (requiresAdmin && !userStore.hasRole('admin')) {
      ElMessage.error('需要管理员权限')
      next({ name: 'home' })
      return
    }
    
    // 其他角色权限检查交给页面组件处理
  }

  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - 协同创新社区`
  }

  next()
})

export default router