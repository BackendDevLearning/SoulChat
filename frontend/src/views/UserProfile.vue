<template>
  <div class="user-profile-container">
    <div class="profile-header">
      <el-button :icon="ArrowLeft" @click="$router.back()">返回</el-button>
      <h1>用户资料</h1>
    </div>
    
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="5" animated />
    </div>
    
    <div v-else-if="userProfile" class="profile-content">
      <el-card class="profile-card">
        <div class="profile-header-info">
          <div class="cover-image" :style="{ backgroundImage: `url(${userProfile.coverImage})` }">
            <div class="profile-avatar">
              <el-avatar :size="120" :src="userProfile.image">
                {{ userProfile.username?.charAt(0) }}
              </el-avatar>
            </div>
          </div>
          
          <div class="profile-info">
            <h2>{{ userProfile.username }}</h2>
            <p v-if="userProfile.bio" class="bio">{{ userProfile.bio }}</p>
            
            <div class="stats">
              <div class="stat-item">
                <span class="stat-number">{{ userProfile.followCount || 0 }}</span>
                <span class="stat-label">关注</span>
              </div>
              <div class="stat-item">
                <span class="stat-number">{{ userProfile.fanCount || 0 }}</span>
                <span class="stat-label">粉丝</span>
              </div>
            </div>
            
            <div class="actions">
              <el-button
                v-if="!isFollowing"
                type="primary"
                :loading="followLoading"
                @click="handleFollow"
              >
                关注
              </el-button>
              <el-button
                v-else
                :loading="followLoading"
                @click="handleUnfollow"
              >
                取消关注
              </el-button>
              <el-button @click="startChat">私聊</el-button>
            </div>
          </div>
        </div>
      </el-card>
      
      <el-card class="details-card">
        <template #header>
          <h3>详细信息</h3>
        </template>
        
        <div class="details-grid">
          <div class="detail-item">
            <span class="label">用户名</span>
            <span class="value">{{ userProfile.username }}</span>
          </div>
          <div class="detail-item">
            <span class="label">性别</span>
            <span class="value">{{ getGenderText(userProfile.gender) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">生日</span>
            <span class="value">{{ userProfile.birthday || '未设置' }}</span>
          </div>
          <div class="detail-item">
            <span class="label">注册时间</span>
            <span class="value">{{ formatDate(userProfile.createdAt) }}</span>
          </div>
        </div>
      </el-card>
    </div>
    
    <div v-else class="error-container">
      <el-empty description="用户不存在" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { authAPI } from '@/api'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const userId = computed(() => route.params.userId)
const loading = ref(true)
const followLoading = ref(false)
const userProfile = ref(null)
const isFollowing = ref(false)

// 获取用户资料
const fetchUserProfile = async () => {
  try {
    loading.value = true
    const response = await authAPI.getUserProfile(userId.value)
    userProfile.value = response.data
    
    // 检查关注状态
    await checkFollowStatus()
  } catch (error) {
    console.error('获取用户资料失败:', error)
    ElMessage.error('获取用户资料失败')
  } finally {
    loading.value = false
  }
}

// 检查关注状态
const checkFollowStatus = async () => {
  try {
    const response = await authAPI.getRelationship(userId.value)
    isFollowing.value = response.data?.isFollowing || false
  } catch (error) {
    console.error('检查关注状态失败:', error)
  }
}

// 关注用户
const handleFollow = async () => {
  try {
    followLoading.value = true
    await authAPI.followUser(userId.value)
    isFollowing.value = true
    userProfile.value.fanCount = (userProfile.value.fanCount || 0) + 1
    ElMessage.success('关注成功')
  } catch (error) {
    ElMessage.error('关注失败')
  } finally {
    followLoading.value = false
  }
}

// 取消关注
const handleUnfollow = async () => {
  try {
    followLoading.value = true
    await authAPI.unfollowUser(userId.value)
    isFollowing.value = false
    userProfile.value.fanCount = Math.max((userProfile.value.fanCount || 0) - 1, 0)
    ElMessage.success('取消关注成功')
  } catch (error) {
    ElMessage.error('取消关注失败')
  } finally {
    followLoading.value = false
  }
}

// 开始私聊
const startChat = () => {
  router.push({
    name: 'Chat',
    query: { userId: userId.value }
  })
}

// 获取性别文本
const getGenderText = (gender) => {
  const genderMap = {
    0: '未知',
    1: '男',
    2: '女'
  }
  return genderMap[gender] || '未知'
}

// 格式化日期
const formatDate = (timestamp) => {
  if (!timestamp) return '未知'
  return new Date(timestamp).toLocaleDateString()
}

onMounted(() => {
  fetchUserProfile()
})
</script>

<style scoped>
.user-profile-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 30px;
}

.profile-header h1 {
  margin: 0;
  color: #333;
}

.loading-container {
  padding: 20px;
}

.profile-content {
  display: flex;
  flex-direction: column;
  gap: 30px;
}

.profile-card {
  padding: 0;
  overflow: hidden;
}

.profile-header-info {
  position: relative;
}

.cover-image {
  height: 200px;
  background-size: cover;
  background-position: center;
  background-color: #f0f0f0;
  position: relative;
}

.profile-avatar {
  position: absolute;
  bottom: -60px;
  left: 50%;
  transform: translateX(-50%);
}

.profile-info {
  padding: 80px 30px 30px;
  text-align: center;
}

.profile-info h2 {
  margin: 0 0 10px 0;
  color: #333;
  font-size: 24px;
}

.bio {
  color: #666;
  margin-bottom: 20px;
  font-size: 14px;
}

.stats {
  display: flex;
  justify-content: center;
  gap: 40px;
  margin-bottom: 30px;
}

.stat-item {
  text-align: center;
}

.stat-number {
  display: block;
  font-size: 20px;
  font-weight: bold;
  color: #333;
}

.stat-label {
  font-size: 12px;
  color: #666;
}

.actions {
  display: flex;
  justify-content: center;
  gap: 12px;
}

.details-card {
  padding: 30px;
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-item .label {
  font-size: 12px;
  color: #666;
  text-transform: uppercase;
}

.detail-item .value {
  font-size: 14px;
  color: #333;
}

.error-container {
  padding: 40px;
  text-align: center;
}
</style>
