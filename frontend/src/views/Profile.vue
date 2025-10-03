<template>
  <div class="profile-container">
    <div class="profile-header">
      <el-button :icon="ArrowLeft" @click="$router.back()">返回</el-button>
      <h1>个人资料</h1>
    </div>
    
    <div class="profile-content">
      <el-card class="profile-card">
        <div class="profile-avatar">
          <el-avatar :size="120" :src="profileForm.image">
            {{ profileForm.username?.charAt(0) }}
          </el-avatar>
          <el-button type="primary" @click="handleAvatarUpload">
            更换头像
          </el-button>
        </div>
        
        <el-form
          ref="profileFormRef"
          :model="profileForm"
          :rules="profileRules"
          label-width="100px"
          class="profile-form"
        >
          <el-form-item label="用户名" prop="username">
            <el-input v-model="profileForm.username" />
          </el-form-item>
          
          <el-form-item label="手机号" prop="phone">
            <el-input v-model="profileForm.phone" disabled />
          </el-form-item>
          
          <el-form-item label="个人简介" prop="bio">
            <el-input
              v-model="profileForm.bio"
              type="textarea"
              :rows="3"
              placeholder="介绍一下自己吧..."
            />
          </el-form-item>
          
          <el-form-item label="性别" prop="gender">
            <el-radio-group v-model="profileForm.gender">
              <el-radio :label="0">未知</el-radio>
              <el-radio :label="1">男</el-radio>
              <el-radio :label="2">女</el-radio>
            </el-radio-group>
          </el-form-item>
          
          <el-form-item label="生日" prop="birthday">
            <el-date-picker
              v-model="profileForm.birthday"
              type="date"
              placeholder="选择生日"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
            />
          </el-form-item>
          
          <el-form-item label="封面图片" prop="coverImage">
            <el-input v-model="profileForm.coverImage" placeholder="封面图片URL" />
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" :loading="loading" @click="updateProfile">
              保存修改
            </el-button>
            <el-button @click="resetForm">重置</el-button>
          </el-form-item>
        </el-form>
      </el-card>
      
      <el-card class="password-card">
        <template #header>
          <h3>修改密码</h3>
        </template>
        
        <el-form
          ref="passwordFormRef"
          :model="passwordForm"
          :rules="passwordRules"
          label-width="100px"
        >
          <el-form-item label="当前密码" prop="oldPassword">
            <el-input
              v-model="passwordForm.oldPassword"
              type="password"
              show-password
            />
          </el-form-item>
          
          <el-form-item label="新密码" prop="newPassword">
            <el-input
              v-model="passwordForm.newPassword"
              type="password"
              show-password
            />
          </el-form-item>
          
          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input
              v-model="passwordForm.confirmPassword"
              type="password"
              show-password
            />
          </el-form-item>
          
          <el-form-item>
            <el-button type="primary" :loading="passwordLoading" @click="updatePassword">
              修改密码
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'

const userStore = useUserStore()

const profileFormRef = ref()
const passwordFormRef = ref()
const loading = ref(false)
const passwordLoading = ref(false)

// 个人资料表单
const profileForm = reactive({
  username: '',
  phone: '',
  bio: '',
  image: '',
  coverImage: '',
  gender: 0,
  birthday: ''
})

// 密码表单
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 表单验证规则
const profileRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 20, message: '用户名长度在2-20个字符', trigger: 'blur' }
  ]
}

const passwordRules = {
  oldPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 初始化表单数据
const initForm = () => {
  if (userStore.user) {
    Object.assign(profileForm, {
      username: userStore.user.username || '',
      phone: userStore.user.phone || '',
      bio: userStore.user.bio || '',
      image: userStore.user.image || '',
      coverImage: userStore.user.coverImage || '',
      gender: userStore.user.gender || 0,
      birthday: userStore.user.birthday || ''
    })
  }
}

// 更新个人资料
const updateProfile = async () => {
  if (!profileFormRef.value) return
  
  await profileFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const success = await userStore.updateProfile(profileForm)
        if (success) {
          ElMessage.success('更新成功')
        }
      } finally {
        loading.value = false
      }
    }
  })
}

// 修改密码
const updatePassword = async () => {
  if (!passwordFormRef.value) return
  
  await passwordFormRef.value.validate(async (valid) => {
    if (valid) {
      passwordLoading.value = true
      try {
        const success = await userStore.updatePassword({
          oldPassword: passwordForm.oldPassword,
          newPassword: passwordForm.newPassword
        })
        if (success) {
          ElMessage.success('密码修改成功')
          // 清空表单
          passwordForm.oldPassword = ''
          passwordForm.newPassword = ''
          passwordForm.confirmPassword = ''
        }
      } finally {
        passwordLoading.value = false
      }
    }
  })
}

// 重置表单
const resetForm = () => {
  initForm()
}

// 处理头像上传
const handleAvatarUpload = () => {
  // TODO: 实现头像上传
  ElMessage.info('头像上传功能开发中')
}

onMounted(() => {
  initForm()
})
</script>

<style scoped>
.profile-container {
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

.profile-content {
  display: flex;
  flex-direction: column;
  gap: 30px;
}

.profile-card {
  padding: 30px;
}

.profile-avatar {
  text-align: center;
  margin-bottom: 30px;
}

.profile-avatar .el-avatar {
  margin-bottom: 20px;
}

.profile-form {
  max-width: 500px;
  margin: 0 auto;
}

.password-card {
  padding: 30px;
}

.password-card .el-form {
  max-width: 500px;
  margin: 0 auto;
}
</style>
