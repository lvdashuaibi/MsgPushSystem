<template>
  <div class="dashboard-content">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card" v-loading="loading">
          <div class="stat-item">
            <div class="stat-icon success">
              <el-icon><Message /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.totalMessages }}</div>
              <div class="stat-label">总消息数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card" v-loading="loading">
          <div class="stat-item">
            <div class="stat-icon primary">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.totalUsers }}</div>
              <div class="stat-label">总用户数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card" v-loading="loading">
          <div class="stat-item">
            <div class="stat-icon warning">
              <el-icon><Timer /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.scheduledMessages }}</div>
              <div class="stat-label">定时消息</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card" v-loading="loading">
          <div class="stat-item">
            <div class="stat-icon info">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.templates }}</div>
              <div class="stat-label">消息模板</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :span="12">
        <el-card title="消息发送统计">
          <div class="chart-container">
            <div class="chart-placeholder">
              📈 消息发送统计图表
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card title="用户增长统计">
          <div class="chart-container">
            <div class="chart-placeholder">
              📊 用户增长统计图表
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Message,
  User,
  Timer,
  Document
} from '@element-plus/icons-vue'
import { getUserList } from '@/api/user'
import { getScheduledMessageList } from '@/api/scheduled'
import { getTemplateList, getMsgRecordList } from '@/api/message'

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
    console.log('仪表盘：开始加载统计数据...')

    // 并行获取所有统计数据
    const [userResponse, scheduledResponse, templateResponse, msgRecordResponse] = await Promise.all([
      getUserList({ page: 1, page_size: 1 }),
      getScheduledMessageList({ page: 1, page_size: 1, status: 1 }),
      getTemplateList({ page: 1, page_size: 1 }),
      getMsgRecordList({ page: 1, page_size: 1 })
    ])

    console.log('仪表盘：用户响应:', userResponse)
    console.log('仪表盘：定时消息响应:', scheduledResponse)
    console.log('仪表盘：模板响应:', templateResponse)
    console.log('仪表盘：消息记录响应:', msgRecordResponse)

    // 处理用户数据
    const userData = userResponse.data as any
    if (userData && userData.total !== undefined) {
      stats.value.totalUsers = userData.total
    } else if (userData && userData.users && Array.isArray(userData.users)) {
      stats.value.totalUsers = userData.users.length
    }

    // 处理定时消息数据
    const scheduledData = scheduledResponse.data as any
    if (scheduledData && scheduledData.total !== undefined) {
      stats.value.scheduledMessages = scheduledData.total
    } else if (scheduledData && scheduledData.scheduled_messages && Array.isArray(scheduledData.scheduled_messages)) {
      stats.value.scheduledMessages = scheduledData.scheduled_messages.length
    }

    // 处理模板数据
    const templateData = templateResponse.data as any
    if (templateData && templateData.total !== undefined) {
      stats.value.templates = templateData.total
    } else if (templateData && templateData.templates && Array.isArray(templateData.templates)) {
      stats.value.templates = templateData.templates.length
    }

    // 处理消息记录数据
    const msgRecordData = msgRecordResponse.data as any
    if (msgRecordData && msgRecordData.total !== undefined) {
      stats.value.totalMessages = msgRecordData.total
    } else if (msgRecordData && msgRecordData.records && Array.isArray(msgRecordData.records)) {
      stats.value.totalMessages = msgRecordData.records.length
    } else {
      // 如果没有消息记录数据，使用默认值
      stats.value.totalMessages = 0
    }

    console.log('仪表盘：最终统计数据:', stats.value)
  } catch (error) {
    console.error('仪表盘：获取统计数据失败:', error)
    ElMessage.error('获取统计数据失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  console.log('Dashboard mounted')
  fetchStats()
})
</script>

<style scoped>
.dashboard-content {
  padding: 20px;
}

.stats-row {
  margin-bottom: 20px;
}

.charts-row {
  margin-bottom: 20px;
}

.stat-card {
  height: 120px;
}

.stat-item {
  display: flex;
  align-items: center;
  height: 100%;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  font-size: 24px;
  color: white;
}

.stat-icon.success {
  background: linear-gradient(135deg, #67c23a, #85ce61);
}

.stat-icon.primary {
  background: linear-gradient(135deg, #409eff, #66b1ff);
}

.stat-icon.warning {
  background: linear-gradient(135deg, #e6a23c, #ebb563);
}

.stat-icon.info {
  background: linear-gradient(135deg, #909399, #a6a9ad);
}

.stat-content {
  flex: 1;
}

.stat-number {
  font-size: 28px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: var(--el-text-color-regular);
}

.chart-container {
  height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-placeholder {
  font-size: 18px;
  color: var(--el-text-color-placeholder);
  text-align: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .dashboard-content {
    padding: 10px;
  }

  .stats-row .el-col {
    margin-bottom: 16px;
  }

  .charts-row .el-col {
    margin-bottom: 16px;
  }
}
</style>
