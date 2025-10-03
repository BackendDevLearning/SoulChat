<template>
  <div class="chat-container">
    <!-- 侧边栏 -->
    <div class="chat-sidebar">
      <div class="sidebar-header">
        <div class="user-info">
          <el-avatar :src="userStore.user?.image" :size="40">
            {{ userStore.user?.username?.charAt(0) }}
          </el-avatar>
          <div class="user-details">
            <h3>{{ userStore.user?.username }}</h3>
            <p :class="{ 'online': chatStore.isConnected, 'offline': !chatStore.isConnected }">
              {{ chatStore.isConnected ? '在线' : '离线' }}
            </p>
          </div>
        </div>
        <el-dropdown @command="handleUserAction">
          <el-button type="text" :icon="Setting" />
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人资料</el-dropdown-item>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      
      <div class="sidebar-content">
        <el-tabs v-model="activeTab" class="chat-tabs">
          <el-tab-pane label="聊天" name="chat">
            <div class="chat-list">
              <div
                v-for="chat in chatList"
                :key="chat.id"
                class="chat-item"
                :class="{ active: currentChat?.id === chat.id }"
                @click="selectChat(chat)"
              >
                <el-avatar :src="chat.avatar" :size="40">
                  {{ chat.name?.charAt(0) }}
                </el-avatar>
                <div class="chat-info">
                  <h4>{{ chat.name }}</h4>
                  <p>{{ chat.lastMessage }}</p>
                </div>
                <div class="chat-meta">
                  <span class="time">{{ formatTime(chat.lastTime) }}</span>
                  <el-badge v-if="chat.unread > 0" :value="chat.unread" />
                </div>
              </div>
            </div>
          </el-tab-pane>
          
          <el-tab-pane label="群组" name="group">
            <div class="group-list">
              <el-button
                type="primary"
                :icon="Plus"
                class="create-group-btn"
                @click="showCreateGroupDialog = true"
              >
                创建群组
              </el-button>
              <div
                v-for="group in chatStore.groups"
                :key="group.id"
                class="group-item"
                @click="selectGroup(group)"
              >
                <el-avatar :size="40">{{ group.name?.charAt(0) }}</el-avatar>
                <div class="group-info">
                  <h4>{{ group.name }}</h4>
                  <p>{{ group.memberCount }} 成员</p>
                </div>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>
    
    <!-- 聊天区域 -->
    <div class="chat-main">
      <div v-if="!currentChat" class="chat-placeholder">
        <el-empty description="选择一个聊天开始对话" />
      </div>
      
      <div v-else class="chat-content">
        <!-- 聊天头部 -->
        <div class="chat-header">
          <div class="chat-title">
            <el-avatar :src="currentChat.avatar" :size="32">
              {{ currentChat.name?.charAt(0) }}
            </el-avatar>
            <div class="title-info">
              <h3>{{ currentChat.name }}</h3>
              <p>{{ currentChat.type === 'group' ? '群聊' : '私聊' }}</p>
            </div>
          </div>
          <div class="chat-actions">
            <el-button type="text" :icon="Phone" />
            <el-button type="text" :icon="VideoCamera" />
            <el-button type="text" :icon="More" />
          </div>
        </div>
        
        <!-- 消息列表 -->
        <div class="message-list" ref="messageListRef">
          <div
            v-for="message in chatStore.messages"
            :key="message.id"
            class="message-item"
            :class="{ 'own': message.from === userStore.user?.username }"
          >
            <div class="message-content">
              <div class="message-bubble">
                <p>{{ message.content }}</p>
                <span class="message-time">{{ formatTime(message.createdAt) }}</span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 消息输入 -->
        <div class="message-input">
          <el-input
            v-model="messageText"
            placeholder="输入消息..."
            @keyup.enter="sendMessage"
          >
            <template #append>
              <el-button :icon="Paperclip" @click="handleFileUpload" />
              <el-button :icon="Picture" @click="handleImageUpload" />
              <el-button type="primary" :icon="Position" @click="sendMessage" />
            </template>
          </el-input>
        </div>
      </div>
    </div>
    
    <!-- 创建群组对话框 -->
    <el-dialog
      v-model="showCreateGroupDialog"
      title="创建群组"
      width="400px"
    >
      <el-form :model="groupForm" :rules="groupRules" ref="groupFormRef">
        <el-form-item label="群组名称" prop="name">
          <el-input v-model="groupForm.name" placeholder="请输入群组名称" />
        </el-form-item>
        <el-form-item label="群组公告" prop="notice">
          <el-input
            v-model="groupForm.notice"
            type="textarea"
            placeholder="请输入群组公告"
            :rows="3"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateGroupDialog = false">取消</el-button>
        <el-button type="primary" @click="createGroup">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useChatStore } from '@/stores/chat'
import { ElMessage } from 'element-plus'
import {
  Setting,
  Plus,
  Phone,
  VideoCamera,
  More,
  Paperclip,
  Picture,
  Position
} from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()

// 响应式数据
const activeTab = ref('chat')
const currentChat = ref(null)
const messageText = ref('')
const messageListRef = ref()
const showCreateGroupDialog = ref(false)

// 聊天列表（模拟数据）
const chatList = ref([
  {
    id: '1',
    name: '张三',
    avatar: '',
    lastMessage: '你好，最近怎么样？',
    lastTime: Date.now() - 1000 * 60 * 30,
    unread: 2,
    type: 'private'
  },
  {
    id: '2',
    name: '李四',
    avatar: '',
    lastMessage: '明天见面吧',
    lastTime: Date.now() - 1000 * 60 * 60 * 2,
    unread: 0,
    type: 'private'
  }
])

// 群组表单
const groupForm = reactive({
  name: '',
  notice: '',
  memberIds: []
})

const groupFormRef = ref()

const groupRules = {
  name: [
    { required: true, message: '请输入群组名称', trigger: 'blur' }
  ]
}

// 生命周期
onMounted(() => {
  // 连接 WebSocket
  if (userStore.user?.username) {
    chatStore.connectWebSocket(userStore.user.username)
  }
})

onUnmounted(() => {
  // 断开 WebSocket
  chatStore.disconnectWebSocket()
})

// 方法
const selectChat = (chat) => {
  currentChat.value = chat
  chatStore.setCurrentChat(chat)
  // 加载消息历史
  loadMessages()
}

const selectGroup = (group) => {
  currentChat.value = { ...group, type: 'group' }
  chatStore.setCurrentChat(currentChat.value)
  loadGroupMessages()
}

const loadMessages = async () => {
  if (currentChat.value?.type === 'private') {
    await chatStore.getMessages({
      to: currentChat.value.id,
      type: 1
    })
  }
}

const loadGroupMessages = async () => {
  if (currentChat.value?.type === 'group') {
    await chatStore.getGroupMessages(currentChat.value.id)
  }
}

const sendMessage = async () => {
  if (!messageText.value.trim() || !currentChat.value) return
  
  const messageData = {
    to: currentChat.value.id,
    content: messageText.value,
    type: currentChat.value.type === 'group' ? 2 : 1,
    contentType: 1
  }
  
  await chatStore.sendMessage(messageData)
  messageText.value = ''
  
  // 滚动到底部
  await nextTick()
  scrollToBottom()
}

const scrollToBottom = () => {
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight
  }
}

const createGroup = async () => {
  if (!groupFormRef.value) return
  
  await groupFormRef.value.validate(async (valid) => {
    if (valid) {
      const success = await chatStore.createGroup(groupForm)
      if (success) {
        showCreateGroupDialog.value = false
        groupForm.name = ''
        groupForm.notice = ''
        ElMessage.success('群组创建成功')
      }
    }
  })
}

const handleUserAction = (command) => {
  switch (command) {
    case 'profile':
      router.push('/profile')
      break
    case 'logout':
      userStore.logout()
      router.push('/login')
      break
  }
}

const handleFileUpload = () => {
  // TODO: 实现文件上传
  ElMessage.info('文件上传功能开发中')
}

const handleImageUpload = () => {
  // TODO: 实现图片上传
  ElMessage.info('图片上传功能开发中')
}

const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date
  
  if (diff < 1000 * 60) {
    return '刚刚'
  } else if (diff < 1000 * 60 * 60) {
    return `${Math.floor(diff / (1000 * 60))}分钟前`
  } else if (diff < 1000 * 60 * 60 * 24) {
    return `${Math.floor(diff / (1000 * 60 * 60))}小时前`
  } else {
    return date.toLocaleDateString()
  }
}
</script>

<style scoped>
.chat-container {
  display: flex;
  height: 100vh;
  background: #f5f5f5;
}

.chat-sidebar {
  width: 300px;
  background: white;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-details h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.user-details p {
  margin: 0;
  font-size: 12px;
}

.user-details .online {
  color: #67c23a;
}

.user-details .offline {
  color: #909399;
}

.sidebar-content {
  flex: 1;
  overflow: hidden;
}

.chat-tabs {
  height: 100%;
}

.chat-list,
.group-list {
  height: calc(100vh - 120px);
  overflow-y: auto;
}

.chat-item,
.group-item {
  padding: 12px 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
}

.chat-item:hover,
.group-item:hover {
  background: #f5f5f5;
}

.chat-item.active {
  background: #e6f7ff;
}

.chat-info,
.group-info {
  flex: 1;
  min-width: 0;
}

.chat-info h4,
.group-info h4 {
  margin: 0 0 4px 0;
  font-size: 14px;
  color: #333;
}

.chat-info p,
.group-info p {
  margin: 0;
  font-size: 12px;
  color: #666;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chat-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.chat-meta .time {
  font-size: 11px;
  color: #999;
}

.create-group-btn {
  width: 100%;
  margin: 20px;
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.chat-placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.chat-header {
  padding: 20px;
  background: white;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.chat-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-info h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.title-info p {
  margin: 0;
  font-size: 12px;
  color: #666;
}

.chat-actions {
  display: flex;
  gap: 8px;
}

.message-list {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background: #fafafa;
}

.message-item {
  margin-bottom: 16px;
  display: flex;
}

.message-item.own {
  justify-content: flex-end;
}

.message-content {
  max-width: 70%;
}

.message-bubble {
  background: white;
  padding: 12px 16px;
  border-radius: 18px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.message-item.own .message-bubble {
  background: #007bff;
  color: white;
}

.message-bubble p {
  margin: 0 0 4px 0;
  font-size: 14px;
  line-height: 1.4;
}

.message-time {
  font-size: 11px;
  opacity: 0.7;
}

.message-input {
  padding: 20px;
  background: white;
  border-top: 1px solid #e0e0e0;
}
</style>
