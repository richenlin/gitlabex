import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { ApiService, type User } from '../services/api'

export interface AuthState {
  isAuthenticated: boolean
  user: User | null
  token: string | null
  loading: boolean
}

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const isAuthenticated = ref(false)
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(false)

  // 计算属性
  const userRole = computed(() => {
    if (!user.value) return null
    return user.value.role
  })

  const isAdmin = computed(() => {
    return userRole.value === 1
  })

  const isTeacher = computed(() => {
    return userRole.value === 2
  })

  const isStudent = computed(() => {
    return userRole.value === 3
  })

  // 方法
  const setToken = (newToken: string | null) => {
    token.value = newToken
    if (newToken) {
      localStorage.setItem('authToken', newToken)
      isAuthenticated.value = true
    } else {
      localStorage.removeItem('authToken')
      isAuthenticated.value = false
    }
  }

  const setUser = (newUser: User | null) => {
    user.value = newUser
  }

  const login = async (authToken: string) => {
    loading.value = true
    try {
      // 设置token
      setToken(authToken)
      
      // 获取用户信息
      const userData = await ApiService.getCurrentUser()
      setUser(userData)
      
      return { success: true, user: userData }
    } catch (error) {
      console.error('登录失败:', error)
      // 清除无效token
      logout()
      return { success: false, error }
    } finally {
      loading.value = false
    }
  }

  const logout = () => {
    setToken(null)
    setUser(null)
    isAuthenticated.value = false
  }

  const checkAuth = async () => {
    // 检查本地存储的token
    const storedToken = localStorage.getItem('authToken')
    if (!storedToken) {
      return false
    }

    loading.value = true
    try {
      // 验证token有效性
      const userData = await ApiService.getCurrentUser()
      setToken(storedToken)
      setUser(userData)
      return true
    } catch (error) {
      console.error('Token验证失败:', error)
      // 清除无效token
      logout()
      return false
    } finally {
      loading.value = false
    }
  }



  const updateUserInfo = async () => {
    if (!isAuthenticated.value) return

    try {
      const userData = await ApiService.getCurrentUser()
      setUser(userData)
      return userData
    } catch (error) {
      console.error('更新用户信息失败:', error)
      throw error
    }
  }

  // 手动更新用户信息（用于资料修改后同步）
  const refreshUserInfo = async () => {
    return updateUserInfo()
  }

  return {
    // 状态
    isAuthenticated,
    user,
    token,
    loading,
    // 计算属性
    userRole,
    isAdmin,
    isTeacher,
    isStudent,
    // 方法
    setToken,
    setUser,
    login,
    logout,
    checkAuth,
    updateUserInfo,
    refreshUserInfo
  }
}) 