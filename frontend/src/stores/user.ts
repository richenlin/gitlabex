import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, EducationRole } from '@/types'
import { UserRole } from '@/types'
import { authService } from '@/services/api'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  const isLoading = ref(false)

  const isLoggedIn = computed(() => !!token.value && !!user.value)
  const username = computed(() => user.value?.username || '')
  const avatar = computed(() => user.value?.avatar_url)
  const userRole = computed(() => user.value?.role || UserRole.GUEST)

  const hasRole = (role: UserRole | string) => {
    // 支持字符串和枚举值的角色检查
    if (typeof role === 'string') {
      return userRole.value === role || 
             (role === 'admin' && userRole.value === UserRole.ADMIN) ||
             (role === 'teacher' && userRole.value === UserRole.TEACHER) ||
             (role === 'assistant' && userRole.value === UserRole.ASSISTANT) ||
             (role === 'student' && userRole.value === UserRole.STUDENT)
    }
    // 对于枚举值的比较，只能用于特定的枚举值
    return false // 暂时返回false，实际应该基于业务逻辑判断
  }

  const hasAnyRole = (...roles: UserRole[]) => {
    return roles.some(role => userRole.value >= role)
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
    
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const fetchCurrentUser = async () => {
    if (!token.value) return false
    
    isLoading.value = true
    try {
      const response: any = await authService.getCurrentUser()
      user.value = response
      
      localStorage.setItem('user', JSON.stringify(user.value))
      return true
    } catch (error) {
      console.error('获取用户信息失败:', error)
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

  return {
    user,
    token,
    isLoading,
    isLoggedIn,
    username,
    avatar,
    userRole,
    hasRole,
    hasAnyRole,
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