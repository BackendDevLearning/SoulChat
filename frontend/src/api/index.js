import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建 axios 实例
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    
    const message = error.response?.data?.msg || error.message || '请求失败'
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

// 认证相关 API
export const authAPI = {
  // 登录
  login: (credentials) => api.post('/users/login', credentials),
  
  // 注册
  register: (userData) => api.post('/users', userData),
  
  // 获取用户信息
  getProfile: () => api.get('/profiles/me'),
  
  // 更新用户信息
  updateProfile: (profileData) => api.put('/users/updateUserInfo', profileData),
  
  // 更新密码
  updatePassword: (passwordData) => api.post('/users/updatePassword', passwordData),
  
  // 获取其他用户信息
  getUserProfile: (userId) => api.get(`/profiles/${userId}`),
  
  // 关注用户
  followUser: (targetId) => api.post(`/profiles/${targetId}/follow`),
  
  // 取消关注
  unfollowUser: (targetId) => api.post(`/profiles/${targetId}/unfollow`),
  
  // 获取关系状态
  getRelationship: (targetId) => api.get(`/profiles/${targetId}/relationship`),
  
  // 检查是否可以添加好友
  canAddFriend: (targetId) => api.post(`/profiles/${targetId}/canAddFriend`)
}

// 聊天相关 API
export const chatAPI = {
  // 发送消息
  sendMessage: (messageData) => api.post('/chat/send', messageData),
  
  // 获取消息历史
  getMessages: (params) => api.get('/chat/messages', { params }),
  
  // 创建群组
  createGroup: (groupData) => api.post('/chat/groups', groupData),
  
  // 加入群组
  joinGroup: (groupId) => api.post(`/chat/groups/${groupId}/join`),
  
  // 获取群组消息
  getGroupMessages: (groupId, params) => api.get(`/chat/groups/${groupId}/messages`, { params })
}

export default api
