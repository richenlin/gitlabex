import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { title: '登录 - GitLabEx', requiresAuth: false }
    },
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: { title: 'GitLabEx - 教育协作平台', requiresAuth: true }
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: { title: '仪表板', requiresAuth: true }
    },
    {
      path: '/documents',
      name: 'documents',
      component: () => import('../views/DocumentsView.vue'),
      meta: { title: '文档管理', requiresAuth: true }
    },
    {
      path: '/documents/editor/:id',
      name: 'document-editor',
      component: () => import('../views/DocumentEditorView.vue'),
      meta: { title: '文档编辑器', requiresAuth: true }
    },
    {
      path: '/wiki',
      name: 'wiki-documents',
      component: () => import('../views/WikiDocumentsView.vue'),
      meta: { title: 'Wiki文档管理', requiresAuth: true }
    },
    
    // 用户管理 (管理员权限)
    {
      path: '/users',
      name: 'users',
      component: () => import('../views/UsersView.vue'),
      meta: { title: '用户管理', requiresAuth: true, requiresRole: 'admin' }
    },
    
    // 班级管理
    {
      path: '/classes',
      name: 'classes',
      component: () => import('../views/ClassesView.vue'),
      meta: { title: '班级管理', requiresAuth: true }
    },
    {
      path: '/classes/:id',
      name: 'class-detail',
      component: () => import('../views/ClassDetailView.vue'),
      meta: { title: '班级详情', requiresAuth: true }
    },
    
    // 课题管理
    {
      path: '/projects',
      name: 'projects',
      component: () => import('../views/ProjectsView.vue'),
      meta: { title: '课题管理', requiresAuth: true }
    },
    {
      path: '/projects/:id',
      name: 'project-detail',
      component: () => import('../views/ProjectDetailView.vue'),
      meta: { title: '课题详情', requiresAuth: true }
    },
    
    // 作业管理
    {
      path: '/assignments',
      name: 'assignments',
      component: () => import('../views/AssignmentsView.vue'),
      meta: { title: '作业管理', requiresAuth: true }
    },
    {
      path: '/assignments/:id',
      name: 'assignment-detail',
      component: () => import('../views/AssignmentDetailView.vue'),
      meta: { title: '作业详情', requiresAuth: true }
    },
    {
      path: '/assignments/:id/submit',
      name: 'assignment-submit',
      component: () => import('../views/AssignmentSubmitView.vue'),
      meta: { title: '提交作业', requiresAuth: true, requiresRole: 'student' }
    },
    
    // 评审管理 (老师权限)
    {
      path: '/reviews',
      name: 'reviews',
      component: () => import('../views/ReviewsView.vue'),
      meta: { title: '评审管理', requiresAuth: true, requiresRole: 'teacher' }
    },
    {
      path: '/reviews/:id',
      name: 'review-detail',
      component: () => import('../views/ReviewDetailView.vue'),
      meta: { title: '评审详情', requiresAuth: true, requiresRole: 'teacher' }
    },
    
    // 通知系统
    {
      path: '/notifications',
      name: 'notifications',
      component: () => import('../views/NotificationsView.vue'),
      meta: { title: '通知系统', requiresAuth: true }
    },
    
    // 个人中心
    {
      path: '/profile',
      name: 'profile',
      component: () => import('../views/ProfileView.vue'),
      meta: { title: '个人资料', requiresAuth: true }
    },
    
    // 学习进度跟踪
    {
      path: '/learning-progress',
      name: 'learning-progress',
      component: () => import('../views/LearningProgressView.vue'),
      meta: { title: '学习进度跟踪', requiresAuth: true }
    },
    
    // 教育报表 (老师和管理员权限)
    {
      path: '/education-reports',
      name: 'education-reports',
      component: () => import('../views/EducationReportsView.vue'),
      meta: { title: '教育报表', requiresAuth: true, requiresRole: 'teacher' }
    },
    
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue'),
      meta: { title: '关于', requiresAuth: false }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('../views/NotFoundView.vue'),
      meta: { title: '页面未找到', requiresAuth: false }
    }
  ],
})

// 路由守卫 - 登录状态检查和权限验证
router.beforeEach(async (to, from, next) => {
  // 更新页面标题
  document.title = to.meta.title as string || 'GitLabEx'
  
  // 如果页面不需要认证，直接通过
  if (!to.meta.requiresAuth) {
    next()
    return
  }
  
  // 检查用户登录状态
  const authStore = useAuthStore()
  
  // 如果已经登录，检查权限
  if (authStore.isAuthenticated) {
    // 检查角色权限
    if (to.meta.requiresRole) {
      const userRole = authStore.userRole
      const requiredRole = to.meta.requiresRole as string
      
      // 检查用户角色 (1: 管理员, 2: 老师, 3: 学生, 4: 访客)
      if (requiredRole === 'admin' && userRole !== 1) {
        next({ name: 'home' })
        return
      } else if (requiredRole === 'teacher' && ![1, 2].includes(userRole || 0)) {
        next({ name: 'home' })
        return
      } else if (requiredRole === 'student' && userRole === 4) {
        next({ name: 'home' })
        return
      }
    }
    
    next()
    return
  }
  
  // 尝试从localStorage恢复登录状态
  const isLoggedIn = await authStore.checkAuth()
  
  if (isLoggedIn) {
    // 登录状态有效，再次检查权限
    if (to.meta.requiresRole) {
      const userRole = authStore.userRole
      const requiredRole = to.meta.requiresRole as string
      
      if (requiredRole === 'admin' && userRole !== 1) {
        next({ name: 'home' })
        return
      } else if (requiredRole === 'teacher' && ![1, 2].includes(userRole || 0)) {
        next({ name: 'home' })
        return
      } else if (requiredRole === 'student' && userRole === 4) {
        next({ name: 'home' })
        return
      }
    }
    
    next()
  } else {
    // 未登录，重定向到登录页
    next({
      path: '/login',
      query: { redirect: to.fullPath }
    })
  }
})

export default router
