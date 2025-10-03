import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '@/api'
import { ElMessage } from 'element-plus'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)
  const isAuthenticated = computed(() => !!token.value)

  // 初始化认证
  const initAuth = () => {
    if (token.value) {
      // 验证 token 有效性
      validateToken()
    }
  }

  // 验证 token
  const validateToken = async () => {
    try {
      const response = await authAPI.getProfile()
      user.value = response.data
    } catch (error) {
      logout()
    }
  }

  // 登录
  const login = async (credentials) => {
    try {
      const response = await authAPI.login(credentials)
      token.value = response.token
      user.value = response.res
      localStorage.setItem('token', token.value)
      ElMessage.success('登录成功')
      return true
    } catch (error) {
      ElMessage.error(error.message || '登录失败')
      return false
    }
  }

  // 注册
  const register = async (userData) => {
    try {
      const response = await authAPI.register(userData)
      token.value = response.token
      user.value = response.res
      localStorage.setItem('token', response.token)
      ElMessage.success('注册成功')
      return true
    } catch (error) {
      ElMessage.error(error.message || '注册失败')
      return false
    }
  }

  // 登出
  const logout = () => {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  // 更新用户信息
  const updateProfile = async (profileData) => {
    try {
      const response = await authAPI.updateProfile(profileData)
      user.value = { ...user.value, ...profileData }
      ElMessage.success('更新成功')
      return true
    } catch (error) {
      ElMessage.error(error.message || '更新失败')
      return false
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    initAuth,
    login,
    register,
    logout,
    updateProfile
  }
})
