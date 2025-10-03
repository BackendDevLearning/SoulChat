import { defineStore } from 'pinia'
import { ref } from 'vue'
import { chatAPI } from '@/api'
import { websocketService } from '@/utils/websocket'

export const useChatStore = defineStore('chat', () => {
  // 状态
  const messages = ref([])
  const currentChat = ref(null)
  const onlineUsers = ref([])
  const groups = ref([])
  const isConnected = ref(false)

  // WebSocket 连接
  const connectWebSocket = (username) => {
    websocketService.connect(username)
    
    websocketService.on('connected', () => {
      isConnected.value = true
    })

    websocketService.on('disconnected', () => {
      isConnected.value = false
    })

    websocketService.on('message', (message) => {
      messages.value.push(message)
    })

    websocketService.on('userJoined', (user) => {
      onlineUsers.value.push(user)
    })

    websocketService.on('userLeft', (username) => {
      onlineUsers.value = onlineUsers.value.filter(u => u !== username)
    })
  }

  // 断开 WebSocket
  const disconnectWebSocket = () => {
    websocketService.disconnect()
    isConnected.value = false
  }

  // 发送消息
  const sendMessage = async (messageData) => {
    try {
      if (isConnected.value) {
        // 通过 WebSocket 发送
        websocketService.sendMessage(messageData)
      } else {
        // 通过 HTTP API 发送
        await chatAPI.sendMessage(messageData)
      }
    } catch (error) {
      console.error('发送消息失败:', error)
    }
  }

  // 获取消息历史
  const getMessages = async (params) => {
    try {
      const response = await chatAPI.getMessages(params)
      messages.value = response.messages || []
      return response
    } catch (error) {
      console.error('获取消息失败:', error)
      return []
    }
  }

  // 创建群组
  const createGroup = async (groupData) => {
    try {
      const response = await chatAPI.createGroup(groupData)
      groups.value.push(response.group)
      return response
    } catch (error) {
      console.error('创建群组失败:', error)
      return null
    }
  }

  // 加入群组
  const joinGroup = async (groupId) => {
    try {
      const response = await chatAPI.joinGroup(groupId)
      return response
    } catch (error) {
      console.error('加入群组失败:', error)
      return false
    }
  }

  // 获取群组消息
  const getGroupMessages = async (groupId) => {
    try {
      const response = await chatAPI.getGroupMessages(groupId)
      return response.messages || []
    } catch (error) {
      console.error('获取群组消息失败:', error)
      return []
    }
  }

  // 设置当前聊天
  const setCurrentChat = (chat) => {
    currentChat.value = chat
  }

  // 清空消息
  const clearMessages = () => {
    messages.value = []
  }

  return {
    messages,
    currentChat,
    onlineUsers,
    groups,
    isConnected,
    connectWebSocket,
    disconnectWebSocket,
    sendMessage,
    getMessages,
    createGroup,
    joinGroup,
    getGroupMessages,
    setCurrentChat,
    clearMessages
  }
})
