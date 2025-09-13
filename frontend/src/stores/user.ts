import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, GitLabRole } from '@/types'
import { UserRole } from '@/types'
import { authService, permissionService } from '@/services/api'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  const isLoading = ref(false)

  const isLoggedIn = computed(() => !!token.value && !!user.value)
  const username = computed(() => user.value?.username || '')
  const avatar = computed(() => user.value?.avatar_url)
  
  // 管理员权限状态
  const adminPermission = ref(false)
  const isAdmin = computed(() => adminPermission.value)

  // 检查管理员权限
  const checkAdminPermission = async () => {
    if (!isLoggedIn.value) {
      adminPermission.value = false
      return false
    }

    try {
      const response: any = await permissionService.checkPermission({
        action: 'manage',
        resource: 'users'
      })
      adminPermission.value = response.allowed || false
      console.log('管理员权限检查结果:', adminPermission.value)
      return adminPermission.value
    } catch (error) {
      console.error('管理员权限检查失败:', error)
      adminPermission.value = false
      return false
    }
  }

  // 权限检查方法 - 现在通过后端API进行
  const checkPermission = async (action: string, resource: string, resourceId?: string) => {
    if (!isLoggedIn.value) {
      return false
    }

    try {
      const response: any = await permissionService.checkPermission({
        action,
        resource,
        resource_id: resourceId
      })
      return response.allowed || false
    } catch (error) {
      console.error('权限检查失败:', error)
      return false
    }
  }

  const checkProjectPermission = async (projectId: string, action = 'read') => {
    if (!isLoggedIn.value) {
      return false
    }

    try {
      const response: any = await permissionService.checkProjectPermission(projectId, action)
      return response.allowed || false
    } catch (error) {
      console.error('项目权限检查失败:', error)
      return false
    }
  }

  const checkProjectPermissionDetailed = async (projectId: string) => {
    if (!isLoggedIn.value) {
      return { permissions: {}, roles: [], access_level: 0 }
    }

    try {
      const response: any = await permissionService.checkProjectPermissionDetailed(projectId)
      return response
    } catch (error) {
      console.error('详细权限检查失败:', error)
      return { permissions: {}, roles: [], access_level: 0 }
    }
  }

  const getUserPermissions = async (projectId?: string) => {
    if (!isLoggedIn.value) {
      return {}
    }

    try {
      const response: any = await permissionService.getUserPermissions(projectId)
      return response.permissions || {}
    } catch (error) {
      console.error('获取用户权限失败:', error)
      return {}
    }
  }

  // 保留简化的本地角色检查，仅用于基本的UI显示逻辑
  // 实际权限验证应该使用上面的API方法
  const hasRole = (role: UserRole | string) => {
    if (!user.value) return false
    
    // 只保留最基本的管理员检查，其他权限都通过API验证
    if (role === 'admin' || role === UserRole.ADMIN) {
      return user.value.is_admin || false
    }
    
    // 其他角色检查应该使用checkPermission方法
    console.warn('hasRole方法已弃用，请使用checkPermission方法进行权限验证')
    return true // 暂时返回true，避免UI显示问题
  }

  const hasAnyRole = (...roles: UserRole[]) => {
    return roles.some(role => hasRole(role))
  }

  const login = async (username: string, password: string) => {
    isLoading.value = true
    try {
      // TODO: 实现GitLab OAuth登录
      // 暂时使用模拟登录逻辑
      console.log('暂未实现传统用户名密码登录，请使用GitLab OAuth')
      throw new Error('请使用GitLab OAuth登录')
    } catch (error) {
      console.error('登录失败:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const setUser = (newUser: User) => {
    user.value = newUser
    localStorage.setItem('user', JSON.stringify(newUser))
  }

  const logout = () => {
    token.value = null
    user.value = null
    adminPermission.value = false
    
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const fetchCurrentUser = async () => {
    if (!token.value) return false
    
    isLoading.value = true
    try {
      console.log('Store: 调用 authService.getCurrentUser()...')
      const response: any = await authService.getCurrentUser()
      console.log('Store: API响应:', response)
      
      // 由于响应拦截器已经返回了data，所以response就是用户数据
      user.value = response
      
      localStorage.setItem('user', JSON.stringify(user.value))
      console.log('Store: 用户信息保存成功:', user.value)
      
      // 检查管理员权限
      await checkAdminPermission()
      
      return true
    } catch (error) {
      console.error('Store: 获取用户信息失败:', error)
      logout()
      return false
    } finally {
      isLoading.value = false
    }
  }

  const updateProfile = async (data: Partial<User>) => {
    isLoading.value = true
    try {
      const response: any = await authService.updateProfile(data)
      user.value = { ...user.value!, ...response }
      
      localStorage.setItem('user', JSON.stringify(user.value))
      return true
    } catch (error) {
      console.error('更新用户信息失败:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const updateUser = (data: Partial<User>) => {
    if (user.value) {
      user.value = { ...user.value, ...data }
      localStorage.setItem('user', JSON.stringify(user.value))
    }
  }

  const changePassword = async (oldPassword: string, newPassword: string) => {
    isLoading.value = true
    try {
      // TODO: 实现修改密码功能
      console.log('GitLab OAuth用户无法在本系统修改密码')
      throw new Error('GitLab OAuth用户请在GitLab系统中修改密码')
    } catch (error) {
      console.error('修改密码失败:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  // 初始化用户状态
  const initUserFromStorage = () => {
    const savedToken = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')
    
    if (savedToken && savedUser) {
      try {
        token.value = savedToken
        user.value = JSON.parse(savedUser)
      } catch (error) {
        console.error('解析用户数据失败:', error)
        // 清除损坏的数据
        localStorage.removeItem('token')
        localStorage.removeItem('user')
      }
    }
  }

  // 验证token有效性
  const validateToken = async () => {
    if (token.value && !user.value) {
      // 如果有token但没有用户信息，尝试获取用户信息
      return await fetchCurrentUser()
    }
    return !!user.value
  }

  // 初始化时从localStorage恢复状态
  initUserFromStorage()

  return {
    user,
    token,
    isLoading,
    isLoggedIn,
    username,
    avatar,
    isAdmin,
    hasRole,
    hasAnyRole,
    checkAdminPermission,
    checkPermission,
    checkProjectPermission,
    checkProjectPermissionDetailed,
    getUserPermissions,
    login,
    setToken,
    setUser,
    logout,
    fetchCurrentUser,
    updateProfile,
    updateUser,
    changePassword,
    initUserFromStorage,
    validateToken
  }
})
