<template>
  <div class="ai-polish-container">
    <el-card class="header-card">
      <div class="header">
        <h2>🎨 AI内容润色</h2>
        <p class="subtitle">输入原始意图，AI自动生成适配不同渠道的专业内容</p>
      </div>
    </el-card>

    <el-row :gutter="20">
      <!-- 左侧：输入区域 -->
      <el-col :span="10">
        <el-card class="input-card">
          <template #header>
            <div class="card-header">
              <span>📝 原始意图输入</span>
            </div>
          </template>

          <el-form :model="form" label-width="100px">
            <el-form-item label="原始意图">
              <el-input
                v-model="form.originalIntent"
                type="textarea"
                :rows="8"
                placeholder="请输入您想要发送的原始内容，例如：本周五凌晨2点到4点系统维护，无法登录，请提前保存数据。"
              />
            </el-form-item>

            <el-form-item label="润色渠道">
              <el-radio-group v-model="form.channel">
                <el-radio :label="0">全部渠道</el-radio>
                <el-radio :label="1">📧 邮件</el-radio>
                <el-radio :label="2">💬 短信</el-radio>
                <el-radio :label="3">🦅 飞书</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="输出模式">
              <el-radio-group v-model="form.streamMode" :disabled="form.channel === 0">
                <el-radio :label="false">标准模式</el-radio>
                <el-radio :label="true">流式输出</el-radio>
              </el-radio-group>
              <div style="font-size: 12px; color: #909399; margin-top: 5px;">
                流式输出可实时查看生成过程（仅支持单渠道）
              </div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                @click="handlePolish"
                :loading="loading"
                :disabled="!form.originalIntent"
              >
                <el-icon v-if="!loading"><MagicStick /></el-icon>
                {{ loading ? '正在润色中...' : '开始润色' }}
              </el-button>
              <el-button @click="handleReset">重置</el-button>
              <el-button @click="loadExample">加载示例</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <!-- 右侧：结果展示区域 -->
      <el-col :span="14">
        <el-card class="result-card">
          <template #header>
            <div class="card-header">
              <span>✨ 润色结果</span>
              <el-tag v-if="result" type="success">已生成</el-tag>
            </div>
          </template>

          <div v-if="!result && !loading" class="empty-state">
            <el-empty description="请输入原始意图并点击开始润色按钮" />
          </div>

          <div v-if="loading" class="loading-state">
            <el-skeleton :rows="6" animated />
          </div>

          <!-- 全部渠道结果 -->
          <div v-if="result && form.channel === 0" class="all-channels-result">
            <!-- 邮件版本 -->
            <el-collapse v-model="activeNames" accordion>
              <el-collapse-item name="email" v-if="result.email_content">
                <template #title>
                  <div class="collapse-title">
                    <span>📧 邮件版本 (HTML)</span>
                    <el-tag size="small" type="primary">{{ result.email_content.format }}</el-tag>
                  </div>
                </template>
                <content-display :content="result.email_content" />
              </el-collapse-item>

              <!-- 短信版本 -->
              <el-collapse-item name="sms" v-if="result.sms_content">
                <template #title>
                  <div class="collapse-title">
                    <span>💬 短信版本 (纯文本)</span>
                    <el-tag size="small" type="success">{{ result.sms_content.format }}</el-tag>
                  </div>
                </template>
                <content-display :content="result.sms_content" />
              </el-collapse-item>

              <!-- 飞书版本 -->
              <el-collapse-item name="lark" v-if="result.lark_content">
                <template #title>
                  <div class="collapse-title">
                    <span>🦅 飞书版本 (JSON卡片)</span>
                    <el-tag size="small" type="warning">{{ result.lark_content.format }}</el-tag>
                  </div>
                </template>
                <content-display :content="result.lark_content" />
              </el-collapse-item>
            </el-collapse>
          </div>

          <!-- 单渠道结果 -->
          <div v-if="result && form.channel !== 0" class="single-channel-result">
            <content-display :content="result" :is-streaming="loading" :key="result.content" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { MagicStick } from '@element-plus/icons-vue'
import ContentDisplay from './components/ContentDisplay.vue'
import api from '@/api'

const form = reactive({
  originalIntent: '',
  channel: 0,
  streamMode: true  // 默认使用流式输出
})

const loading = ref(false)
const result = ref(null)
const activeNames = ref(['email'])

const handlePolish = async () => {
  if (!form.originalIntent.trim()) {
    ElMessage.warning('请输入原始意图')
    return
  }

  // 如果是单渠道且开启流式模式
  if (form.channel !== 0 && form.streamMode) {
    handleStreamPolish()
    return
  }

  // 标准模式
  loading.value = true
  result.value = null

  // 显示加载提示
  const loadingMsg = ElMessage({
    message: 'AI正在生成内容，请稍候...',
    type: 'info',
    duration: 0,
    showClose: false
  })

  try {
    const url = form.channel === 0
      ? '/ai/polish/all'
      : '/ai/polish/single'

    const response = await api.post(url, {
      original_intent: form.originalIntent,
      channel: form.channel
    })

    loadingMsg.close()

    if (response.code === 0) {
      result.value = form.channel === 0 ? response.data : response.data
      ElMessage.success('内容润色成功！')
    } else {
      ElMessage.error(response.msg || '润色失败')
    }
  } catch (error) {
    loadingMsg.close()
    console.error('润色失败:', error)

    let errorMsg = '润色失败'
    if (error.message?.includes('timeout')) {
      errorMsg = 'AI服务响应超时，请稍后重试'
    } else if (error.message?.includes('Network Error')) {
      errorMsg = '网络连接失败，请检查网络'
    } else if (error.message) {
      errorMsg = error.message
    }

    ElMessage.error(errorMsg)
  } finally {
    loading.value = false
  }
}

// 流式润色处理
const handleStreamPolish = () => {
  loading.value = true
  result.value = null

  // 根据渠道确定格式
  const formatMap = {
    1: 'html',
    2: 'text',
    3: 'json'
  }

  // 初始化临时结果对象
  const tempResult = {
    channel: form.channel,
    subject: '正在生成...',
    content: '',
    format: formatMap[form.channel] || 'text',
    raw_content: form.originalIntent,
    description: '正在生成...'
  }
  result.value = tempResult

  const url = `/api/ai/polish/stream?original_intent=${encodeURIComponent(form.originalIntent)}&channel=${form.channel}`
  console.log('开始流式请求:', url)

  const eventSource = new EventSource(url)
  let isCompleted = false

  eventSource.onopen = () => {
    console.log('SSE连接已建立')
  }

  eventSource.onmessage = (event) => {
    try {
      console.log('收到SSE消息:', event.data)
      const data = JSON.parse(event.data)

      switch (data.event) {
        case 'start':
          console.log('开始生成:', data.data)
          tempResult.subject = data.data.message || '正在生成...'
          // 强制更新，触发组件重新渲染
          result.value = JSON.parse(JSON.stringify(tempResult))
          break

        case 'chunk':
          // 实时更新内容
          console.log('收到chunk:', data.data.content)
          tempResult.content = data.data.total || ''
          // 每次chunk都强制更新，确保UI实时显示
          result.value = JSON.parse(JSON.stringify(tempResult))
          break

        case 'complete':
          // 生成完成
          console.log('生成完成:', data.data)
          result.value = data.data
          isCompleted = true
          ElMessage.success('内容润色成功！')
          eventSource.close()
          loading.value = false
          break

        case 'error':
          console.error('生成错误:', data.data)
          isCompleted = true
          ElMessage.error(data.data.message || '生成失败')
          eventSource.close()
          loading.value = false
          break
      }
    } catch (error) {
      console.error('解析SSE数据失败:', error, '原始数据:', event.data)
    }
  }

  eventSource.onerror = (error) => {
    console.error('SSE连接错误:', error)
    if (eventSource.readyState === EventSource.CLOSED) {
      console.log('SSE连接已关闭')
      // 如果连接关闭但还没有收到complete事件，说明出错了
      if (!isCompleted) {
        loading.value = false
      }
    } else {
      ElMessage.error('连接中断，请重试')
      loading.value = false
    }
    eventSource.close()
  }
}

const handleReset = () => {
  form.originalIntent = ''
  form.channel = 0
  result.value = null
}

const loadExample = () => {
  form.originalIntent = '本周五凌晨2点到4点系统维护，无法登录，请提前保存数据。'
  ElMessage.info('已加载示例内容')
}
</script>

<style scoped>
.ai-polish-container {
  padding: 20px;
}

.header-card {
  margin-bottom: 20px;
}

.header {
  text-align: center;
}

.header h2 {
  margin: 0 0 10px 0;
  color: #409EFF;
}

.subtitle {
  color: #909399;
  margin: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.input-card, .result-card {
  min-height: 500px;
}

.empty-state, .loading-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.collapse-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding-right: 20px;
}

.all-channels-result {
  margin-top: 10px;
}

.single-channel-result {
  margin-top: 10px;
}
</style>
