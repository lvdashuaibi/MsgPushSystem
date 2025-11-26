import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建axios实例
const api = axios.create({
  baseURL: '/api',
  timeout: 60000, // 增加到60秒，适应AI请求
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 添加用户ID头（用于测试）- 后端期望的是Source-Id
    config.headers['Source-Id'] = 'frontend-user'
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    const { data } = response
    if (data.code === 0) {
      // 后端直接返回数据在根级别，需要适配前端期望的格式
      const { code, msg, ...rest } = data
      return {
        code,
        msg,
        data: Object.keys(rest).length > 0 ? rest : undefined
      }
    } else {
      ElMessage.error(data.msg || '请求失败')
      return Promise.reject(new Error(data.msg || '请求失败'))
    }
  },
  (error) => {
    ElMessage.error(error.message || '网络错误')
    return Promise.reject(error)
  }
)

export default api
