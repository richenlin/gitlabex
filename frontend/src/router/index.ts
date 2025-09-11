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
      meta: { title: '课题详情', requiresAuth: false }
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
      meta: { title: '创建话题' }
    },
    {
      path: '/topics/:id',
      name: 'topic-detail',
      component: () => import('@/views/topics/TopicDetail.vue'),
      meta: { title: '话题详情', requiresAuth: false }
    },
    {
      path: '/topics/:id/edit',
      name: 'topic-edit',
      component: () => import('@/views/topics/TopicEdit.vue'),
      meta: { title: '编辑话题' }
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
      meta: { title: '上传文档' }
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
      meta: { title: '作业列表' }
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
      meta: { title: '作业详情' }
    },
    {
      path: '/homeworks/:id/submit',
      name: 'homework-submit',
      component: () => import('@/views/homeworks/HomeworkSubmit.vue'),
      meta: { title: '提交作业' }
    },
    {
      path: '/homeworks/:id/grade',
      name: 'homework-grade',
      component: () => import('@/views/homeworks/HomeworkGrade.vue'),
      meta: { title: '批改作业', roles: [UserRole.TEACHER, UserRole.ASSISTANT] }
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
      path: '/notifications',
      name: 'notifications',
      component: () => import('@/views/user/Notifications.vue'),
      meta: { title: '通知', requiresAuth: true }
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/admin/Dashboard.vue'),
      meta: { title: '管理后台', roles: [UserRole.ADMIN] }
    }
  ]
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  
  // 初始化用户信息
  if (!userStore.user && userStore.token) {
    await userStore.fetchCurrentUser()
  }

  // 检查是否需要登录 - 游客可访问标记为 false 的路由
  if (to.meta.requiresAuth !== false && !userStore.isLoggedIn) {
    next({
      name: 'login',
      query: { redirect: to.fullPath }
    })
    return
  }

  // 检查角色权限
  if (to.meta.roles && Array.isArray(to.meta.roles) && to.meta.roles.length > 0) {
    const hasRole = to.meta.roles.some((role: any) => userStore.hasRole(role))
    if (!hasRole) {
      ElMessage.error('权限不足')
      next({ name: 'home' })
      return
    }
  }

  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - 协同创新社区`
  }

  next()
})

export default router