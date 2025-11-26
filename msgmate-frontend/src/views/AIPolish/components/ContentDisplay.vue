<template>
  <div class="content-display">
    <el-descriptions :column="1" border>
      <el-descriptions-item label="主题">
        <el-tag type="primary">{{ content.subject }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="格式">
        <el-tag :type="formatType">{{ content.format }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="说明">
        {{ content.description }}
      </el-descriptions-item>
    </el-descriptions>

    <el-divider />

    <div class="content-section">
      <div class="section-header">
        <span class="section-title">📝 润色后的内容</span>
        <el-button-group>
          <el-button size="small" @click="copyContent">
            <el-icon><DocumentCopy /></el-icon>
            复制
          </el-button>
          <el-button size="small" @click="downloadContent">
            <el-icon><Download /></el-icon>
            下载
          </el-button>
        </el-button-group>
      </div>

      <!-- HTML格式预览 -->
      <div v-if="content.format === 'html'" class="html-preview">
        <el-tabs v-model="activeTab">
          <el-tab-pane label="预览" name="preview">
            <div class="preview-container" v-html="content.content"></div>
          </el-tab-pane>
          <el-tab-pane label="源代码" name="source">
            <el-input
              v-model="content.content"
              type="textarea"
              :rows="15"
              readonly
              class="code-textarea"
            />
          </el-tab-pane>
        </el-tabs>
      </div>

      <!-- 纯文本格式 -->
      <div v-else-if="content.format === 'text'" class="text-preview">
        <div class="streaming-text">
          {{ content.content }}
          <span v-if="isStreaming" class="cursor-blink">|</span>
        </div>
        <div class="text-stats">
          <el-tag size="small">字数: {{ textLength }}</el-tag>
          <el-tag v-if="isStreaming" size="small" type="warning">生成中...</el-tag>
        </div>
      </div>

      <!-- JSON格式 -->
      <div v-else-if="content.format === 'json'" class="json-preview">
        <el-tabs v-model="activeTab">
          <el-tab-pane label="格式化" name="formatted">
            <pre class="json-code">{{ formattedJson }}</pre>
          </el-tab-pane>
          <el-tab-pane label="原始" name="raw">
            <el-input
              v-model="content.content"
              type="textarea"
              :rows="15"
              readonly
              class="code-textarea"
            />
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { DocumentCopy, Download } from '@element-plus/icons-vue'

const props = defineProps({
  content: {
    type: Object,
    required: true
  },
  isStreaming: {
    type: Boolean,
    default: false
  }
})

const activeTab = ref('preview')

const formatType = computed(() => {
  const types = {
    html: 'primary',
    text: 'success',
    json: 'warning'
  }
  return types[props.content.format] || 'info'
})

const textLength = computed(() => {
  return props.content.content ? props.content.content.length : 0
})

const formattedJson = computed(() => {
  try {
    const jsonObj = JSON.parse(props.content.content)
    return JSON.stringify(jsonObj, null, 2)
  } catch (e) {
    return props.content.content
  }
})

const copyContent = () => {
  navigator.clipboard.writeText(props.content.content).then(() => {
    ElMessage.success('内容已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

const downloadContent = () => {
  const extensions = {
    html: 'html',
    text: 'txt',
    json: 'json'
  }
  const ext = extensions[props.content.format] || 'txt'
  const filename = `${props.content.subject}.${ext}`

  const blob = new Blob([props.content.content], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)

  ElMessage.success('文件已下载')
}
</script>

<style scoped>
.content-display {
  padding: 10px;
}

.content-section {
  margin-top: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.section-title {
  font-size: 16px;
  font-weight: bold;
  color: #303133;
}

.html-preview, .text-preview, .json-preview {
  margin-top: 10px;
}

.preview-container {
  border: 1px solid #DCDFE6;
  border-radius: 4px;
  padding: 20px;
  background-color: #FFFFFF;
  min-height: 300px;
}

.code-textarea, .text-textarea {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.json-code {
  background-color: #F5F7FA;
  border: 1px solid #DCDFE6;
  border-radius: 4px;
  padding: 15px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  overflow-x: auto;
  max-height: 500px;
  overflow-y: auto;
}

.text-stats {
  margin-top: 10px;
  text-align: right;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.streaming-text {
  background-color: #F5F7FA;
  border: 1px solid #DCDFE6;
  border-radius: 4px;
  padding: 15px;
  min-height: 200px;
  font-family: 'PingFang SC', 'Microsoft YaHei', sans-serif;
  font-size: 14px;
  line-height: 1.8;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.cursor-blink {
  animation: blink 1s infinite;
  color: #409EFF;
  font-weight: bold;
}

@keyframes blink {
  0%, 50% {
    opacity: 1;
  }
  51%, 100% {
    opacity: 0;
  }
}
</style>
