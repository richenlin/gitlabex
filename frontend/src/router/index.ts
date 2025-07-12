import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { title: '登录 - GitLabEx' }
    },
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: { title: 'GitLabEx - 教育协作平台' }
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: { title: '仪表板' }
    },
    {
      path: '/documents',
      name: 'documents',
      component: () => import('../views/DocumentsView.vue'),
      meta: { title: '文档管理' }
    },
    {
      path: '/documents/editor/:id',
      name: 'document-editor',
      component: () => import('../views/DocumentEditorView.vue'),
      meta: { title: '文档编辑器' }
    },
    {
      path: '/wiki',
      name: 'wiki-documents',
      component: () => import('../views/WikiDocumentsView.vue'),
      meta: { title: 'Wiki文档管理' }
    },
    {
      path: '/users',
      name: 'users',
      component: () => import('../views/UsersView.vue'),
      meta: { title: '用户管理' }
    },
    {
      path: '/classes',
      name: 'classes',
      component: () => import('../views/ClassesView.vue'),
      meta: { title: '班级管理' }
    },
    {
      path: '/assignments',
      name: 'assignments',
      component: () => import('../views/AssignmentsView.vue'),
      meta: { title: '作业管理' }
    },
    {
      path: '/projects',
      name: 'projects',
      component: () => import('../views/ProjectsView.vue'),
      meta: { title: '课题管理' }
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('../views/ProfileView.vue'),
      meta: { title: '个人资料' }
    },
    {
      path: '/learning-progress',
      name: 'learning-progress',
      component: () => import('../views/LearningProgressView.vue'),
      meta: { title: '学习进度跟踪' }
    },
    {
      path: '/notifications',
      name: 'notifications',
      component: () => import('../views/NotificationsView.vue'),
      meta: { title: '通知系统' }
    },
    {
      path: '/education-reports',
      name: 'education-reports',
      component: () => import('../views/EducationReportsView.vue'),
      meta: { title: '教育报表' }
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../views/AboutView.vue'),
      meta: { title: '关于' }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('../views/NotFoundView.vue'),
      meta: { title: '页面未找到' }
    }
  ],
})

// 路由守卫 - 更新页面标题
router.beforeEach((to, from, next) => {
  document.title = to.meta.title as string || 'GitLabEx'
  next()
})

export default router
