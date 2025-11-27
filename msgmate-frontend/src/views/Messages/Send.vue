<template>
  <div class="send-message">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>发送消息</span>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        class="send-form"
      >
        <!-- 消息类型选择 -->
        <el-form-item label="消息类型" required>
          <el-radio-group v-model="messageType" @change="handleMessageTypeChange">
            <el-radio label="template">使用模板</el-radio>
            <el-radio label="direct">直接编写</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 接收者类型选择 -->
        <el-form-item label="接收者类型" required>
          <el-radio-group v-model="receiverType" @change="handleReceiverTypeChange">
            <el-radio label="direct">直接指定</el-radio>
            <el-radio label="users">按用户ID</el-radio>
            <el-radio label="tags">按标签</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 直接指定接收者 -->
        <el-form-item
          v-if="receiverType === 'direct'"
          label="接收者"
          prop="to"
          required
        >
          <el-input
            v-model="form.to"
            placeholder="请输入接收者（用户ID、邮箱、手机号等）"
            style="width: 400px"
          />
        </el-form-item>

        <!-- 按用户ID选择 -->
        <el-form-item
          v-if="receiverType === 'users'"
          label="目标用户"
          prop="user_ids"
          required
        >
          <div class="user-selector">
            <el-select
              v-model="form.user_ids"
              placeholder="搜索并选择用户"
              multiple
              filterable
              remote
              :remote-method="searchUsers"
              :loading="searchingUsers"
              style="width: 100%"
            >
              <el-option
                v-for="user in userOptions"
                :key="user.user_id"
                :label="`${user.name} (${user.user_id})`"
                :value="user.user_id"
              >
                <div class="user-option">
                  <span class="user-name">{{ user.name }}</span>
                  <span class="user-id">{{ user.user_id }}</span>
                  <span v-if="user.email" class="user-email">{{ user.email }}</span>
                </div>
              </el-option>
            </el-select>
            <div class="selected-users">
              <el-tag
                v-for="userId in form.user_ids"
                :key="userId"
                closable
                @close="removeUser(userId)"
                style="margin-right: 8px; margin-top: 8px"
              >
                {{ getUserDisplayName(userId) }}
              </el-tag>
            </div>
          </div>
        </el-form-item>

        <!-- 按标签选择 -->
        <el-form-item
          v-if="receiverType === 'tags'"
          label="目标标签"
          prop="tags"
          required
        >
          <div class="tag-selector">
            <el-select
              v-model="form.tags"
              placeholder="选择标签"
              multiple
              style="width: 100%"
            >
              <el-option
                v-for="tag in availableTags"
                :key="tag.tag || 'unknown'"
                :label="`${tag.tag || '未知标签'} (${tag.count || 0}个用户)`"
                :value="tag.tag || ''"
              />
            </el-select>
            <div v-if="form.tags.length > 0" class="tag-preview">
              <el-button
                type="text"
                @click="previewTagUsers"
                :loading="previewingUsers"
              >
                预览用户 ({{ estimatedUserCount }} 人)
              </el-button>
            </div>
          </div>
        </el-form-item>

        <!-- 使用模板模式 -->
        <template v-if="messageType === 'template'">
          <el-form-item label="消息模板" prop="templateID" required>
            <el-select
              v-model="form.templateID"
              placeholder="请选择消息模板"
              style="width: 400px"
              @change="handleTemplateChange"
            >
              <el-option
                v-for="template in templates"
                :key="template.relTemplateID"
                :label="template.name"
                :value="template.relTemplateID"
              >
                <div>
                  <span>{{ template.name }}</span>
                  <span style="color: #8492a6; font-size: 12px; margin-left: 8px">
                    {{ template.subject }}
                  </span>
                </div>
              </el-option>
            </el-select>
            <el-button
              type="text"
              @click="loadTemplates"
              style="margin-left: 8px"
            >
              刷新
            </el-button>
          </el-form-item>

          <el-form-item label="消息主题" prop="subject">
            <el-input
              v-model="form.subject"
              placeholder="消息主题（可选）"
              style="width: 400px"
            />
          </el-form-item>
        </template>

        <!-- 直接编写模式 -->
        <template v-if="messageType === 'direct'">
          <el-form-item label="消息渠道" prop="channels" required>
            <el-select
              v-model="form.channels"
              placeholder="选择消息渠道（支持多选）"
              multiple
              style="width: 100%; max-width: 400px"
            >
              <el-option label="邮件" :value="1" />
              <el-option label="短信" :value="2" />
              <el-option label="飞书" :value="3" />
              <el-option label="微信" :value="4" />
              <el-option label="钉钉" :value="5" />
            </el-select>
            <div v-if="form.channels && form.channels.length > 0" class="selected-channels" style="margin-top: 8px">
              <el-tag
                v-for="channel in form.channels"
                :key="channel"
                closable
                @close="removeChannel(channel)"
                style="margin-right: 8px"
              >
                {{ getChannelText(channel) }}
              </el-tag>
            </div>
          </el-form-item>

          <el-form-item label="消息主题" prop="subject">
            <el-input
              v-model="form.subject"
              placeholder="消息主题"
              style="width: 400px"
            />
          </el-form-item>

          <el-form-item label="消息内容" prop="content" required>
            <div style="width: 100%; max-width: 600px">
              <el-input
                v-model="form.content"
                type="textarea"
                placeholder="请输入消息内容"
                :rows="6"
                style="width: 100%"
              />
              <div style="margin-top: 8px; display: flex; gap: 8px; align-items: center;">
                <el-button
                  type="primary"
                  size="small"
                  @click="handleAIPolish"
                  :loading="aiPolishLoading"
                  :disabled="!form.content.trim()"
                >
                  <el-icon><MagicStick /></el-icon>
                  AI润色
                </el-button>
                <el-tooltip content="使用AI自动润色消息内容，使其更专业、更吸引人">
                  <el-icon style="cursor: help; color: #909399;"><QuestionFilled /></el-icon>
                </el-tooltip>

                <!-- HTML预览按钮 -->
                <el-button
                  v-if="form.channels && form.channels.includes(1)"
                  type="text"
                  size="small"
                  @mouseenter="showHtmlPreview = true"
                  @mouseleave="showHtmlPreview = false"
                  style="margin-left: auto;"
                >
                  👁️ HTML预览
                </el-button>
              </div>
            </div>
          </el-form-item>
        </template>

        <el-form-item label="优先级" prop="priority">
          <el-select v-model="form.priority" placeholder="选择优先级" style="width: 200px">
            <el-option label="低" :value="1" />
            <el-option label="普通" :value="2" />
            <el-option label="高" :value="3" />
            <el-option label="紧急" :value="4" />
          </el-select>
        </el-form-item>

        <!-- 模板预览（仅在使用模板模式显示） -->
        <el-form-item v-if="messageType === 'template' && selectedTemplate" label="模板预览">
          <el-card class="template-preview">
            <div class="template-info">
              <p><strong>模板名称：</strong>{{ selectedTemplate.name }}</p>
              <p><strong>渠道：</strong>{{ getChannelText(selectedTemplate.channel) }}</p>
              <p><strong>主题：</strong>{{ selectedTemplate.subject }}</p>
              <p><strong>内容：</strong></p>
              <div class="template-content">{{ selectedTemplate.content }}</div>
            </div>
          </el-card>
        </el-form-item>

        <!-- 模板数据（仅在使用模板模式显示） -->
        <el-form-item v-if="messageType === 'template' && templateVariables.length > 0" label="模板数据" required>
          <div class="template-data">
            <div
              v-for="variable in templateVariables"
              :key="variable"
              class="variable-item"
            >
              <el-input
                v-model="form.templateData[variable]"
                :placeholder="`请输入 ${variable} 的值`"
                style="width: 300px"
              >
                <template #prepend>{{ variable }}</template>
              </el-input>
            </div>
          </div>
        </el-form-item>

        <el-form-item label="发送时间">
          <el-radio-group v-model="sendTimeType">
            <el-radio label="now">立即发送</el-radio>
            <el-radio label="scheduled">定时发送</el-radio>
          </el-radio-group>
          <el-date-picker
            v-if="sendTimeType === 'scheduled'"
            v-model="scheduledTime"
            type="datetime"
            placeholder="选择发送时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="margin-left: 16px"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSend" :loading="sending">
            {{ sendTimeType === 'now' ? '立即发送' : '创建定时消息' }}
          </el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 用户预览对话框 -->
    <el-dialog v-model="showUserPreview" title="目标用户预览" width="800px">
      <el-table :data="previewUsers" style="width: 100%" max-height="400">
        <el-table-column prop="user_id" label="用户ID" width="150" />
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column prop="email" label="邮箱" width="180" />
        <el-table-column prop="mobile" label="手机号" width="130" />
        <el-table-column label="标签" min-width="200">
          <template #default="{ row }">
            <el-tag
              v-for="tag in row.tags"
              :key="tag"
              size="small"
              style="margin-right: 4px"
            >
              {{ tag }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="showUserPreview = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- HTML预览浮窗 -->
    <el-popover
      v-model:visible="showHtmlPreview"
      placement="right"
      :width="600"
      trigger="manual"
      title="HTML预览"
    >
      <template #reference>
        <div style="display: none;"></div>
      </template>
      <div class="html-preview-container">
        <iframe
          :srcDoc="getHtmlPreviewSrcDoc(form.content)"
          class="html-preview-iframe"
          frameborder="0"
          scrolling="auto"
        ></iframe>
      </div>
    </el-popover>

    <!-- 发送结果对话框 -->
    <el-dialog v-model="showResult" title="发送结果" width="500px">
      <div v-if="sendResult">
        <el-result
          :icon="sendResult.success ? 'success' : 'error'"
          :title="sendResult.success ? '发送成功' : '发送失败'"
          :sub-title="sendResult.message"
        >
          <template v-if="sendResult.success && sendResult.msgID" #extra>
            <p>消息ID: {{ sendResult.msgID }}</p>
          </template>
        </el-result>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { MagicStick, QuestionFilled } from '@element-plus/icons-vue'
import { sendMessage, getTemplate, getTemplateList } from '@/api/message'
import { createScheduledMessage } from '@/api/scheduled'
import { getUserList, findUsersByTags, getTagStatistics } from '@/api/user'
import type { SendMsgReq, Template, CreateScheduledMessageReq, User, TagStatistic } from '@/types'
import api from '@/api'

// 表单引用
const formRef = ref<FormInstance>()

// 消息类型
const messageType = ref<'template' | 'direct'>('template')

// 接收者类型
const receiverType = ref<'direct' | 'users' | 'tags'>('direct')

// 表单数据
const form = reactive<SendMsgReq & { subject?: string; priority?: number; channels?: number[]; content?: string }>({
  to: '',
  user_ids: [],
  tags: [],
  subject: '',
  priority: 2,
  templateID: '',
  templateData: {},
  channels: [],
  content: '',
  sendTimestamp: undefined
})

// 表单验证规则
const rules: FormRules = {
  to: [
    {
      validator: (rule, value, callback) => {
        if (receiverType.value === 'direct' && !value) {
          callback(new Error('请输入接收者'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  user_ids: [
    {
      validator: (rule, value, callback) => {
        if (receiverType.value === 'users' && (!value || value.length === 0)) {
          callback(new Error('请选择目标用户'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ],
  tags: [
    {
      validator: (rule, value, callback) => {
        if (receiverType.value === 'tags' && (!value || value.length === 0)) {
          callback(new Error('请选择目标标签'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ],
  templateID: [
    {
      validator: (rule, value, callback) => {
        if (messageType.value === 'template' && !value) {
          callback(new Error('请选择消息模板'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ],
  channels: [
    {
      validator: (rule, value, callback) => {
        if (messageType.value === 'direct' && (!value || value.length === 0)) {
          callback(new Error('请选择消息渠道'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ],
  content: [
    {
      validator: (rule, value, callback) => {
        if (messageType.value === 'direct' && !value?.trim()) {
          callback(new Error('请输入消息内容'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 状态
const sending = ref(false)
const showResult = ref(false)
const searchingUsers = ref(false)
const previewingUsers = ref(false)
const showUserPreview = ref(false)
const aiPolishLoading = ref(false)
const showHtmlPreview = ref(false)
const sendResult = ref<{
  success: boolean
  message: string
  msgID?: string
}>()

// 数据
const userOptions = ref<User[]>([])
const availableTags = ref<TagStatistic[]>([])
const templates = ref<Template[]>([])
const selectedTemplate = ref<Template>()
const templateVariables = ref<string[]>([])
const previewUsers = ref<User[]>([])

// 发送时间类型
const sendTimeType = ref<'now' | 'scheduled'>('now')
const scheduledTime = ref<string>()

// 计算属性
const estimatedUserCount = computed(() => {
  if (receiverType.value === 'tags') {
    return form.tags.reduce((total, tagName) => {
      const tag = availableTags.value.find(t => t.tag === tagName)
      return total + (tag?.count || 0)
    }, 0)
  } else if (receiverType.value === 'users') {
    return form.user_ids.length
  } else if (receiverType.value === 'direct') {
    return form.to ? 1 : 0
  }
  return 0
})

// 获取渠道文本
const getChannelText = (channel: number) => {
  const channelMap: Record<number, string> = {
    1: '邮件',
    2: '短信',
    3: '飞书',
    4: '微信',
    5: '钉钉'
  }
  return channelMap[channel] || '未知'
}

// 获取用户显示名称
const getUserDisplayName = (userId: string) => {
  const user = userOptions.value.find(u => u.user_id === userId)
  return user ? `${user.name} (${user.user_id})` : userId
}

// 消息类型变化处理
const handleMessageTypeChange = () => {
  // 清空模板相关数据
  form.templateID = ''
  form.templateData = {}
  selectedTemplate.value = undefined
  templateVariables.value = []

  // 清空直接编写相关数据
  form.content = ''
  form.channels = []
}

// 接收者类型变化处理
const handleReceiverTypeChange = () => {
  form.to = ''
  form.user_ids = []
  form.tags = []
}

// 移除渠道
const removeChannel = (channel: number) => {
  const index = form.channels!.indexOf(channel)
  if (index > -1) {
    form.channels!.splice(index, 1)
  }
}

// 搜索用户
const searchUsers = async (query: string) => {
  searchingUsers.value = true
  try {
    const response = await getUserList({
      page: 1,
      page_size: 50
    })

    let users = response.data?.users || []

    // 如果有搜索关键词，进行过滤
    if (query) {
      users = users.filter(user =>
        user.name.toLowerCase().includes(query.toLowerCase()) ||
        user.user_id.toLowerCase().includes(query.toLowerCase()) ||
        (user.email && user.email.toLowerCase().includes(query.toLowerCase())) ||
        (user.mobile && user.mobile.includes(query))
      )
    }

    userOptions.value = users
  } catch (error) {
    ElMessage.error('搜索用户失败')
  } finally {
    searchingUsers.value = false
  }
}

// 初始化加载用户列表
const loadInitialUsers = async () => {
  await searchUsers('')
}

// 移除用户
const removeUser = (userId: string) => {
  const index = form.user_ids.indexOf(userId)
  if (index > -1) {
    form.user_ids.splice(index, 1)
  }
}

// 预览标签用户
const previewTagUsers = async () => {
  if (form.tags.length === 0) return

  previewingUsers.value = true
  try {
    const response = await findUsersByTags({
      tags: form.tags,
      match_type: 'any'
    })
    previewUsers.value = response.data?.users || []
    showUserPreview.value = true
  } catch (error) {
    ElMessage.error('预览用户失败')
  } finally {
    previewingUsers.value = false
  }
}

// 加载模板列表
const loadTemplates = async () => {
  try {
    const response = await getTemplateList({
      page: 1,
      page_size: 100
    })

    // 转换数据格式
    const templateData = response.data?.templates || []
    templates.value = templateData.map((template: any) => ({
      relTemplateID: template.TemplateID,
      sourceID: template.SourceID,
      signName: template.SignName,
      name: template.Name,
      subject: template.Subject,
      channel: template.Channel,
      content: template.Content
    }))

    if (templates.value.length === 0) {
      ElMessage.info('请先在"消息模板"页面创建模板')
    }
  } catch (error) {
    ElMessage.error('加载模板失败')
  }
}

// AI润色处理
const handleAIPolish = async () => {
  if (!form.content.trim()) {
    ElMessage.warning('请先输入消息内容')
    return
  }

  aiPolishLoading.value = true

  try {
    // 根据选择的渠道确定要润色的渠道
    let channel = 1 // 默认邮件
    if (form.channels && form.channels.length === 1) {
      channel = form.channels[0]
    }

    const url = `/api/ai/polish/stream?original_intent=${encodeURIComponent(form.content)}&channel=${channel}`
    console.log('开始流式润色请求:', url)

    let polishedContent = ''
    let polishedSubject = ''

    const eventSource = new EventSource(url)

    eventSource.onopen = () => {
      console.log('SSE连接已建立')
    }

    let messageInstance = null
    let isCompleted = false
    let timeoutId = null

    // 设置超时，防止消息永久显示
    const setTimeoutHandler = () => {
      timeoutId = setTimeout(() => {
        console.warn('AI润色超时，自动关闭')
        if (messageInstance) {
          messageInstance.close()
          messageInstance = null
        }
        eventSource.close()
        aiPolishLoading.value = false
        isCompleted = true
      }, 120000) // 120秒超时
    }

    const clearTimeoutHandler = () => {
      if (timeoutId) {
        clearTimeout(timeoutId)
        timeoutId = null
      }
    }

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        console.log('收到SSE消息:', data.event, data.data)

        switch (data.event) {
          case 'start':
            console.log('开始生成内容...')
            setTimeoutHandler()
            // 显示生成中的提示
            messageInstance = ElMessage({
              message: '正在生成内容，请稍候...',
              type: 'info',
              duration: 0
            })
            break

          case 'chunk':
            // 实时累积内容
            polishedContent = data.data.total || ''
            // 实时更新表单，让用户看到生成过程
            form.content = polishedContent
            console.log('当前内容长度:', polishedContent.length)
            break

          case 'complete':
            // 生成完成
            console.log('生成完成:', data.data)
            clearTimeoutHandler()
            polishedContent = data.data.content
            polishedSubject = data.data.subject

            // 更新表单内容
            form.content = polishedContent

            // 如果有主题，也更新主题
            if (polishedSubject && !form.subject) {
              form.subject = polishedSubject
            }

            // 关闭生成中的提示
            if (messageInstance) {
              messageInstance.close()
              messageInstance = null
            }

            isCompleted = true
            ElMessage.success('内容润色成功！')
            eventSource.close()
            aiPolishLoading.value = false
            break

          case 'error':
            console.error('生成错误:', data.data)
            clearTimeoutHandler()

            // 关闭生成中的提示
            if (messageInstance) {
              messageInstance.close()
              messageInstance = null
            }

            isCompleted = true
            ElMessage.error(data.data.message || '生成失败')
            eventSource.close()
            aiPolishLoading.value = false
            break
        }
      } catch (error) {
        console.error('解析SSE数据失败:', error, '原始数据:', event.data)
      }
    }

    eventSource.onerror = (error) => {
      console.error('SSE连接错误:', error)
      clearTimeoutHandler()

      if (!isCompleted) {
        if (messageInstance) {
          messageInstance.close()
          messageInstance = null
        }
        ElMessage.error('连接中断，请重试')
      }

      eventSource.close()
      aiPolishLoading.value = false
    }
  } catch (error) {
    console.error('AI润色失败:', error)
    let errorMsg = 'AI润色失败'
    if (error instanceof Error) {
      if (error.message.includes('timeout')) {
        errorMsg = 'AI服务响应超时，请稍后重试'
      } else {
        errorMsg = error.message
      }
    }
    ElMessage.error(errorMsg)
    aiPolishLoading.value = false
  }
}

// 模板变化处理
const handleTemplateChange = async (templateId: string) => {
  if (!templateId) {
    selectedTemplate.value = undefined
    templateVariables.value = []
    form.templateData = {}
    return
  }

  try {
    const response = await getTemplate(templateId)
    selectedTemplate.value = response.data

    // 解析模板中的变量（假设使用 {{variable}} 格式）
    const content = selectedTemplate.value?.content || ''
    const matches = content.match(/\{\{(\w+)\}\}/g)
    if (matches) {
      templateVariables.value = matches.map(match => match.replace(/[{}]/g, ''))
      // 初始化模板数据
      const newTemplateData: Record<string, string> = {}
      templateVariables.value.forEach(variable => {
        newTemplateData[variable] = ''
      })
      form.templateData = newTemplateData
    } else {
      templateVariables.value = []
      form.templateData = {}
    }
  } catch (error) {
    ElMessage.error('获取模板信息失败')
  }
}

// 发送消息
const handleSend = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()

    // 验证模板数据（仅在使用模板模式）
    if (messageType.value === 'template' && templateVariables.value.length > 0) {
      const missingVariables = templateVariables.value.filter(
        variable => !form.templateData[variable]?.trim()
      )
      if (missingVariables.length > 0) {
        ElMessage.error(`请填写模板变量: ${missingVariables.join(', ')}`)
        return
      }
    }

    sending.value = true

    if (sendTimeType.value === 'now') {
      // 立即发送
      let sendData: any = {
        priority: form.priority
      }

      if (messageType.value === 'template') {
        // 使用模板模式
        sendData.templateID = form.templateID
        sendData.templateData = form.templateData
        sendData.subject = form.subject
      } else {
        // 直接编写模式
        sendData.channels = form.channels
        sendData.subject = form.subject
        sendData.content = form.content
      }

      // 根据接收者类型设置相应字段
      if (receiverType.value === 'direct') {
        sendData.to = form.to
      } else if (receiverType.value === 'users') {
        sendData.user_ids = form.user_ids
      } else if (receiverType.value === 'tags') {
        sendData.tags = form.tags
      }

      const response = await sendMessage(sendData)

      sendResult.value = {
        success: true,
        message: '消息发送成功',
        msgID: response.msgID
      }
    } else {
      // 定时发送
      if (!scheduledTime.value) {
        ElMessage.error('请选择发送时间')
        return
      }

      let scheduledData: any = {
        scheduled_time: scheduledTime.value
      }

      if (messageType.value === 'template') {
        // 使用模板模式
        scheduledData.template_id = form.templateID
        scheduledData.template_data = form.templateData
      } else {
        // 直接编写模式
        scheduledData.channels = form.channels
        scheduledData.subject = form.subject
        scheduledData.content = form.content
      }

      // 根据接收者类型设置相应字段
      if (receiverType.value === 'direct') {
        scheduledData.to = form.to
      } else if (receiverType.value === 'users') {
        scheduledData.user_ids = form.user_ids
      } else if (receiverType.value === 'tags') {
        scheduledData.tags = form.tags
      }

      const response = await createScheduledMessage(scheduledData)
      sendResult.value = {
        success: true,
        message: '定时消息创建成功',
        msgID: response.schedule_id
      }
    }

    showResult.value = true
  } catch (error: any) {
    sendResult.value = {
      success: false,
      message: error.message || '操作失败'
    }
    showResult.value = true
  } finally {
    sending.value = false
  }
}

// 重置表单
const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, {
    to: '',
    user_ids: [],
    tags: [],
    subject: '',
    priority: 2,
    templateID: '',
    templateData: {},
    channels: [],
    content: ''
  })
  selectedTemplate.value = undefined
  templateVariables.value = []
  sendTimeType.value = 'now'
  scheduledTime.value = undefined
  receiverType.value = 'direct'
  messageType.value = 'template'
}

// 加载标签统计
const loadTagStatistics = async () => {
  try {
    console.log('发送消息页面：开始加载标签统计...')
    const response = await getTagStatistics()
    console.log('发送消息页面：标签统计响应:', response)
    console.log('发送消息页面：response.data类型:', typeof response.data)
    console.log('发送消息页面：response.data内容:', response.data)
    console.log('发送消息页面：response.data是否为数组:', Array.isArray(response.data))

    const responseData = response.data as any

    if (Array.isArray(responseData)) {
      console.log('发送消息页面：数组长度:', responseData.length)
      console.log('发送消息页面：第一个元素:', responseData[0])
      availableTags.value = responseData
    } else if (responseData && responseData.data && Array.isArray(responseData.data)) {
      console.log('发送消息页面：嵌套数组长度:', responseData.data.length)
      console.log('发送消息页面：嵌套第一个元素:', responseData.data[0])
      availableTags.value = responseData.data
    } else {
      console.log('发送消息页面：数据格式不正确，设置为空数组')
      availableTags.value = []
    }

    console.log('发送消息页面：最终availableTags:', availableTags.value)
    console.log('发送消息页面：标签统计加载成功，数量:', availableTags.value.length)
  } catch (error) {
    console.error('发送消息页面：加载标签统计失败:', error)
    ElMessage.error('加载标签统计失败')
    // 设置空数组以防止页面卡住
    availableTags.value = []
  }
}

// 生成HTML预览的srcDoc
const getHtmlPreviewSrcDoc = (htmlContent: string) => {
  return `
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <style>
        body {
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
          line-height: 1.6;
          color: #333;
          margin: 0;
          padding: 20px;
          background-color: #f5f5f5;
        }
        h1, h2, h3, h4, h5, h6 {
          color: #303133;
          margin-top: 16px;
          margin-bottom: 8px;
        }
        p {
          margin: 8px 0;
          line-height: 1.8;
        }
        strong {
          color: #303133;
          font-weight: 600;
        }
        ul, ol {
          margin: 8px 0;
          padding-left: 24px;
        }
        li {
          margin: 4px 0;
        }
        a {
          color: #409EFF;
          text-decoration: none;
        }
        a:hover {
          text-decoration: underline;
        }
        table {
          border-collapse: collapse;
          width: 100%;
          margin: 16px 0;
        }
        table th, table td {
          border: 1px solid #ddd;
          padding: 8px;
          text-align: left;
        }
        table th {
          background-color: #f5f5f5;
          font-weight: 600;
        }
        code {
          background-color: #f5f5f5;
          padding: 2px 6px;
          border-radius: 3px;
          font-family: 'Courier New', monospace;
        }
        pre {
          background-color: #f5f5f5;
          padding: 12px;
          border-radius: 4px;
          overflow-x: auto;
        }
        pre code {
          background-color: transparent;
          padding: 0;
        }
      </style>
    </head>
    <body>
      ${htmlContent}
    </body>
    </html>
  `
}

// 初始化
onMounted(() => {
  loadTemplates()
  loadTagStatistics()
  loadInitialUsers()
})
</script>

<style scoped>
.send-message {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.send-form {
  max-width: 800px;
}

.user-selector {
  width: 100%;
}

.user-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-name {
  font-weight: 600;
}

.user-id {
  color: #909399;
  font-size: 12px;
}

.user-email {
  color: #606266;
  font-size: 12px;
}

.selected-users {
  margin-top: 8px;
}

.tag-selector {
  width: 100%;
}

.tag-preview {
  margin-top: 8px;
}

.template-preview {
  width: 100%;
  max-width: 600px;
}

.template-info p {
  margin: 8px 0;
}

.template-content {
  background-color: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  white-space: pre-wrap;
  font-family: monospace;
  margin-top: 8px;
}

.template-data {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.variable-item {
  display: flex;
  align-items: center;
}

.selected-channels {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.html-preview-container {
  width: 100%;
  height: 500px;
  border-radius: 4px;
  overflow: hidden;
}

.html-preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
}
</style>
