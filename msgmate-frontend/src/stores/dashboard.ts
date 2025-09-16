import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUserList } from '@/api/user'
import { getScheduledMessageList } from '@/api/scheduled'

export const useDashboardStore = defineStore('dashboard', () => {
  // 统计数据
  const stats = ref({
    totalMessages: 0,
    totalUsers: 0,
    scheduledMessages: 0,
    templates: 0
  })

  // 加载状态
  const loading = ref(false)

  // 获取统计数据
  const fetchStats = async () => {
    loading.value = true
    try {
      // 获取用户总数
      const userResponse = await getUserList({ page: 1, page_size: 1 })
      stats.value.totalUsers = userResponse.data?.total || 0

      // 获取定时消息数量
      const scheduledResponse = await getScheduledMessageList({ page: 1, page_size: 1, status: 1 })
      stats.value.scheduledMessages = scheduledResponse.data?.total || 0

      // 模拟其他数据
      stats.value.totalMessages = Math.floor(Math.random() * 1000) + 500
      stats.value.templates = Math.floor(Math.random() * 20) + 5
    } catch (error) {
      console.error('获取统计数据失败:', error)
    } finally {
      loading.value = false
    }
  }

  return {
    stats,
    loading,
    fetchStats
  }
})
