import protobuf from 'protobufjs'

class WebSocketService {
  constructor() {
    this.ws = null
    this.listeners = {}
    this.messageType = null
    this.isConnecting = false
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
    this.reconnectInterval = 3000
  }

  // 连接 WebSocket
  async connect(username) {
    if (this.isConnecting || this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    this.isConnecting = true

    try {
      // 加载 protobuf 定义
      await this.loadProtobufDefinitions()
      
      const wsUrl = `ws://localhost:8000/ws?user=${encodeURIComponent(username)}`
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        console.log('WebSocket 连接成功')
        this.isConnecting = false
        this.reconnectAttempts = 0
        this.emit('connected')
      }

      this.ws.onmessage = (event) => {
        this.handleMessage(event.data)
      }

      this.ws.onclose = (event) => {
        console.log('WebSocket 连接关闭', event.code, event.reason)
        this.isConnecting = false
        this.emit('disconnected')
        
        // 自动重连
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          setTimeout(() => {
            this.reconnectAttempts++
            this.connect(username)
          }, this.reconnectInterval)
        }
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket 错误:', error)
        this.isConnecting = false
        this.emit('error', error)
      }

    } catch (error) {
      console.error('WebSocket 连接失败:', error)
      this.isConnecting = false
      this.emit('error', error)
    }
  }

  // 加载 protobuf 定义
  async loadProtobufDefinitions() {
    try {
      // 这里需要从后端获取 protobuf 定义文件
      // 或者使用预定义的 protobuf 结构
      const root = protobuf.Root.fromJSON({
        nested: {
          Message: {
            fields: {
              id: { type: 'string', id: 1 },
              from: { type: 'string', id: 2 },
              to: { type: 'string', id: 3 },
              content: { type: 'string', id: 4 },
              type: { type: 'int32', id: 5 },
              contentType: { type: 'int32', id: 6 },
              url: { type: 'string', id: 7 },
              createdAt: { type: 'int64', id: 8 }
            }
          }
        }
      })
      
      this.messageType = root.lookupType('Message')
    } catch (error) {
      console.error('加载 protobuf 定义失败:', error)
      throw error
    }
  }

  // 处理接收到的消息
  handleMessage(data) {
    try {
      if (data instanceof ArrayBuffer) {
        // 二进制 protobuf 消息
        const message = this.messageType.decode(new Uint8Array(data))
        const messageObj = this.messageType.toObject(message)
        this.emit('message', messageObj)
      } else {
        // JSON 消息（用于调试）
        const message = JSON.parse(data)
        this.emit('message', message)
      }
    } catch (error) {
      console.error('处理消息失败:', error)
    }
  }

  // 发送消息
  sendMessage(messageData) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      try {
        if (this.messageType) {
          // 发送 protobuf 消息
          const message = this.messageType.create(messageData)
          const buffer = this.messageType.encode(message).finish()
          this.ws.send(buffer)
        } else {
          // 发送 JSON 消息（用于调试）
          this.ws.send(JSON.stringify(messageData))
        }
      } catch (error) {
        console.error('发送消息失败:', error)
      }
    } else {
      console.warn('WebSocket 未连接')
    }
  }

  // 发送心跳
  sendHeartbeat() {
    const heartbeat = {
      type: 0, // HEAT_BEAT
      contentType: 1,
      content: 'ping'
    }
    this.sendMessage(heartbeat)
  }

  // 断开连接
  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  // 事件监听
  on(event, callback) {
    if (!this.listeners[event]) {
      this.listeners[event] = []
    }
    this.listeners[event].push(callback)
  }

  // 移除事件监听
  off(event, callback) {
    if (this.listeners[event]) {
      this.listeners[event] = this.listeners[event].filter(cb => cb !== callback)
    }
  }

  // 触发事件
  emit(event, ...args) {
    if (this.listeners[event]) {
      this.listeners[event].forEach(callback => {
        try {
          callback(...args)
        } catch (error) {
          console.error('事件回调执行失败:', error)
        }
      })
    }
  }

  // 获取连接状态
  getConnectionState() {
    if (!this.ws) return 'DISCONNECTED'
    
    switch (this.ws.readyState) {
      case WebSocket.CONNECTING:
        return 'CONNECTING'
      case WebSocket.OPEN:
        return 'CONNECTED'
      case WebSocket.CLOSING:
        return 'CLOSING'
      case WebSocket.CLOSED:
        return 'DISCONNECTED'
      default:
        return 'UNKNOWN'
    }
  }
}

// 创建单例实例
export const websocketService = new WebSocketService()
