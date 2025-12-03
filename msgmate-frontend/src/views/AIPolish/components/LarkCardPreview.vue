<template>
  <div class="lark-card-container">
    <div v-if="parseError" class="error-message">
      <el-alert
        title="卡片解析失败"
        type="error"
        :description="parseError"
        show-icon
      />
    </div>

    <div v-else-if="cardData" class="lark-card" :class="{ 'wide-screen': cardData.config?.wide_screen_mode }">
      <!-- 卡片头部 -->
      <div v-if="cardData.header" class="card-header" :class="`template-${cardData.header.template || 'blue'}`">
        <div class="header-title">
          {{ cardData.header.title?.content || '标题' }}
        </div>
      </div>

      <!-- 卡片内容 -->
      <div class="card-body">
        <div v-for="(element, index) in cardData.elements" :key="index" class="card-element">
          <!-- 文本内容 -->
          <div v-if="element.tag === 'div' && element.text" class="element-div">
            <div v-if="element.text.tag === 'lark_md'" class="markdown-content" v-html="renderMarkdown(element.text.content)"></div>
            <div v-else class="plain-text">{{ element.text.content }}</div>
          </div>

          <!-- 字段列表 -->
          <div v-else-if="element.tag === 'div' && element.fields" class="element-fields">
            <div class="fields-grid">
              <div
                v-for="(field, fieldIndex) in element.fields"
                :key="fieldIndex"
                class="field-item"
                :class="{ 'is-short': field.is_short }"
              >
                <div v-if="field.text?.tag === 'lark_md'" class="field-content" v-html="renderMarkdown(field.text.content)"></div>
                <div v-else class="field-content">{{ field.text?.content }}</div>
              </div>
            </div>
          </div>

          <!-- 分割线 -->
          <div v-else-if="element.tag === 'hr'" class="element-hr"></div>

          <!-- 提示信息 -->
          <div v-else-if="element.tag === 'note'" class="element-note">
            <div v-for="(noteEl, noteIndex) in element.elements" :key="noteIndex">
              {{ noteEl.content }}
            </div>
          </div>

          <!-- 操作按钮 -->
          <div v-else-if="element.tag === 'action'" class="element-action">
            <div class="action-buttons">
              <el-button
                v-for="(action, actionIndex) in element.actions"
                :key="actionIndex"
                :type="action.type === 'primary' ? 'primary' : 'default'"
                size="default"
                @click="handleButtonClick(action)"
              >
                {{ action.text?.content }}
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="empty-card">
      <el-empty description="暂无卡片内容" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  cardJson: {
    type: String,
    required: true
  }
})

const cardData = ref(null)
const parseError = ref('')

// 解析卡片JSON
const parseCard = () => {
  try {
    parseError.value = ''
    if (!props.cardJson || props.cardJson.trim() === '') {
      cardData.value = null
      return
    }

    const parsed = JSON.parse(props.cardJson)
    cardData.value = parsed
  } catch (error) {
    parseError.value = `JSON解析错误: ${error.message}`
    cardData.value = null
  }
}

// 监听cardJson变化
watch(() => props.cardJson, () => {
  parseCard()
}, { immediate: true })

// 渲染Markdown内容
const renderMarkdown = (content) => {
  if (!content) return ''

  let html = content

  // 转义HTML特殊字符
  html = html.replace(/&/g, '&amp;')
           .replace(/</g, '&lt;')
           .replace(/>/g, '&gt;')

  // 加粗 **text**
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')

  // 斜体 *text*
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>')

  // 链接 [text](url)
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank">$1</a>')

  // 换行
  html = html.replace(/\n/g, '<br>')

  // 列表项 - item
  html = html.replace(/^- (.+)$/gm, '<li>$1</li>')
  html = html.replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>')

  return html
}

// 处理按钮点击
const handleButtonClick = (action) => {
  if (action.url) {
    ElMessage.info(`按钮点击: ${action.text?.content}\n链接: ${action.url}`)
  } else {
    ElMessage.info(`按钮点击: ${action.text?.content}`)
  }
}
</script>

<style scoped>
.lark-card-container {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

.error-message {
  margin-bottom: 20px;
}

.lark-card {
  background: #ffffff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.lark-card.wide-screen {
  max-width: 100%;
}

/* 卡片头部 */
.card-header {
  padding: 16px 20px;
  color: white;
  font-weight: 600;
  font-size: 16px;
}

.card-header.template-blue {
  background: linear-gradient(135deg, #4A90E2 0%, #357ABD 100%);
}

.card-header.template-red {
  background: linear-gradient(135deg, #F56C6C 0%, #E54D42 100%);
}

.card-header.template-green {
  background: linear-gradient(135deg, #67C23A 0%, #5DAF34 100%);
}

.card-header.template-orange {
  background: linear-gradient(135deg, #E6A23C 0%, #CF9236 100%);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 卡片主体 */
.card-body {
  padding: 20px;
}

.card-element {
  margin-bottom: 16px;
}

.card-element:last-child {
  margin-bottom: 0;
}

/* 文本内容 */
.element-div {
  line-height: 1.6;
  color: #303133;
}

.markdown-content {
  word-wrap: break-word;
}

.markdown-content :deep(strong) {
  font-weight: 600;
  color: #303133;
}

.markdown-content :deep(em) {
  font-style: italic;
  color: #606266;
}

.markdown-content :deep(a) {
  color: #409EFF;
  text-decoration: none;
}

.markdown-content :deep(a:hover) {
  text-decoration: underline;
}

.markdown-content :deep(ul) {
  margin: 8px 0;
  padding-left: 20px;
}

.markdown-content :deep(li) {
  margin: 4px 0;
}

.plain-text {
  color: #606266;
}

/* 字段列表 */
.element-fields {
  background: #F5F7FA;
  border-radius: 4px;
  padding: 16px;
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.field-item {
  min-height: 50px;
}

.field-item.is-short {
  grid-column: span 1;
}

.field-content {
  font-size: 14px;
  line-height: 1.6;
}

.field-content :deep(strong) {
  display: block;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

/* 分割线 */
.element-hr {
  height: 1px;
  background: #DCDFE6;
  margin: 16px 0;
}

/* 提示信息 */
.element-note {
  background: #F0F9FF;
  border-left: 4px solid #409EFF;
  padding: 12px 16px;
  border-radius: 4px;
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
}

/* 操作按钮 */
.element-action {
  margin-top: 16px;
}

.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.empty-card {
  padding: 40px;
  text-align: center;
}

/* 响应式 */
@media (max-width: 768px) {
  .fields-grid {
    grid-template-columns: 1fr;
  }

  .field-item.is-short {
    grid-column: span 1;
  }
}
</style>
