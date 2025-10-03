// 时间格式化工具
export const formatTime = (timestamp) => {
  if (!timestamp) return ''
  
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date
  
  // 小于1分钟
  if (diff < 1000 * 60) {
    return '刚刚'
  }
  
  // 小于1小时
  if (diff < 1000 * 60 * 60) {
    const minutes = Math.floor(diff / (1000 * 60))
    return `${minutes}分钟前`
  }
  
  // 小于1天
  if (diff < 1000 * 60 * 60 * 24) {
    const hours = Math.floor(diff / (1000 * 60 * 60))
    return `${hours}小时前`
  }
  
  // 小于1周
  if (diff < 1000 * 60 * 60 * 24 * 7) {
    const days = Math.floor(diff / (1000 * 60 * 60 * 24))
    return `${days}天前`
  }
  
  // 超过1周，显示具体日期
  return date.toLocaleDateString()
}

// 日期格式化
export const formatDate = (timestamp, format = 'YYYY-MM-DD') => {
  if (!timestamp) return ''
  
  const date = new Date(timestamp)
  
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  
  switch (format) {
    case 'YYYY-MM-DD':
      return `${year}-${month}-${day}`
    case 'YYYY-MM-DD HH:mm':
      return `${year}-${month}-${day} ${hours}:${minutes}`
    case 'YYYY-MM-DD HH:mm:ss':
      return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
    case 'MM-DD HH:mm':
      return `${month}-${day} ${hours}:${minutes}`
    default:
      return date.toLocaleString()
  }
}

// 文件大小格式化
export const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 获取文件扩展名
export const getFileExtension = (filename) => {
  if (!filename) return ''
  return filename.split('.').pop().toLowerCase()
}

// 判断是否为图片文件
export const isImageFile = (filename) => {
  const imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg']
  const ext = getFileExtension(filename)
  return imageExtensions.includes(ext)
}

// 判断是否为视频文件
export const isVideoFile = (filename) => {
  const videoExtensions = ['mp4', 'avi', 'mov', 'wmv', 'flv', 'webm', 'mkv']
  const ext = getFileExtension(filename)
  return videoExtensions.includes(ext)
}

// 判断是否为音频文件
export const isAudioFile = (filename) => {
  const audioExtensions = ['mp3', 'wav', 'flac', 'aac', 'ogg', 'wma']
  const ext = getFileExtension(filename)
  return audioExtensions.includes(ext)
}

// 生成随机字符串
export const generateRandomString = (length = 8) => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let result = ''
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

// 防抖函数
export const debounce = (func, wait) => {
  let timeout
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout)
      func(...args)
    }
    clearTimeout(timeout)
    timeout = setTimeout(later, wait)
  }
}

// 节流函数
export const throttle = (func, limit) => {
  let inThrottle
  return function(...args) {
    if (!inThrottle) {
      func.apply(this, args)
      inThrottle = true
      setTimeout(() => inThrottle = false, limit)
    }
  }
}

// 深拷贝
export const deepClone = (obj) => {
  if (obj === null || typeof obj !== 'object') return obj
  if (obj instanceof Date) return new Date(obj.getTime())
  if (obj instanceof Array) return obj.map(item => deepClone(item))
  if (typeof obj === 'object') {
    const clonedObj = {}
    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
        clonedObj[key] = deepClone(obj[key])
      }
    }
    return clonedObj
  }
}

// 验证手机号
export const validatePhone = (phone) => {
  const phoneRegex = /^1[3-9]\d{9}$/
  return phoneRegex.test(phone)
}

// 验证邮箱
export const validateEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

// 验证密码强度
export const validatePasswordStrength = (password) => {
  if (password.length < 6) return { valid: false, message: '密码长度不能少于6位' }
  if (password.length > 20) return { valid: false, message: '密码长度不能超过20位' }
  
  const hasNumber = /\d/.test(password)
  const hasLetter = /[a-zA-Z]/.test(password)
  const hasSpecial = /[!@#$%^&*(),.?":{}|<>]/.test(password)
  
  if (hasNumber && hasLetter && hasSpecial) {
    return { valid: true, strength: 'strong', message: '密码强度：强' }
  } else if ((hasNumber && hasLetter) || (hasNumber && hasSpecial) || (hasLetter && hasSpecial)) {
    return { valid: true, strength: 'medium', message: '密码强度：中等' }
  } else {
    return { valid: true, strength: 'weak', message: '密码强度：弱' }
  }
}

// 获取浏览器信息
export const getBrowserInfo = () => {
  const ua = navigator.userAgent
  const isChrome = ua.includes('Chrome')
  const isFirefox = ua.includes('Firefox')
  const isSafari = ua.includes('Safari') && !ua.includes('Chrome')
  const isEdge = ua.includes('Edge')
  const isIE = ua.includes('MSIE') || ua.includes('Trident')
  
  return {
    isChrome,
    isFirefox,
    isSafari,
    isEdge,
    isIE,
    userAgent: ua
  }
}

// 检查是否为移动设备
export const isMobile = () => {
  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
}

// 获取URL参数
export const getUrlParams = (url = window.location.href) => {
  const params = {}
  const urlObj = new URL(url)
  for (const [key, value] of urlObj.searchParams) {
    params[key] = value
  }
  return params
}

// 设置URL参数
export const setUrlParams = (params) => {
  const url = new URL(window.location.href)
  Object.keys(params).forEach(key => {
    if (params[key] !== null && params[key] !== undefined) {
      url.searchParams.set(key, params[key])
    } else {
      url.searchParams.delete(key)
    }
  })
  window.history.replaceState({}, '', url)
}
