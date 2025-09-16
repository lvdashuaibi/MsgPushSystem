<template>
  <div class="scheduled-create">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>创建定时消息</span>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        class="create-form"
      >
        <!-- 接收者选择 -->
        <el-form-item label="接收者类型" required>
          <el-radio-group v-model="receiverType" @change="handleReceiverTypeChange">
            <el-radio label="direct">直接指定</el-radio>
            <el-radio label="users">指定用户</el-radio>
            <el-radio label="tags">按标签选择</el-radio>
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

        <!-- 指定用户 -->
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

        <!-- 消息模板 -->
        <el-form-item label="消息模板" prop="template_id" required>
          <el-select
            v-model="form.template_id"
            placeholder="请选择消息模板"
            style="width: 100%"
            @change="handleTemplateChange"
          >
            <el-option
              v-for="template in templates"
              :key="template.relTemplateID"
              :label="template.name"
              :value="template.relTemplateID"
            >
              <div class="template-option">
                <span class="template-name">{{ template.name }}</span>
                <span class="template-subject">{{ template.subject }}</span>
                <el-tag size="small" :type="getChannelTagType(template.channel)">
                  {{ getChannelText(template.channel) }}
                </el-tag>
              </div>
            </el-option>
          </el-select>
          <el-button
            type="text"
            @click="loadTemplates"
            style="margin-left: 8px"
          >
            刷新模板
          </el-button>
        </el-form-item>

        <!-- 模板预览 -->
        <el-form-item v-if="selectedTemplate" label="模板预览">
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

        <!-- 模板数据 -->
        <el-form-item v-if="templateVariables.length > 0" label="模板数据" required>
          <div class="template-data">
            <div
              v-for="variable in templateVariables"
              :key="variable"
              class="variable-item"
            >
              <el-input
                v-model="form.template_data[variable]"
                :placeholder="`请输入 ${variable} 的值`"
                style="width: 300px"
              >
                <template #prepend>{{ variable }}</template>
              </el-input>
            </div>
          </div>
        </el-form-item>

        <!-- 发送时间 -->
        <el-form-item label="发送时间" prop="scheduled_time" required>
          <el-date-picker
            v-model="form.scheduled_time"
            type="datetime"
            placeholder="选择发送时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            :disabled-date="disabledDate"
            :disabled-hours="disabledHours"
            :disabled-minutes="disabledMinutes"
            style="width: 300px"
          />
          <div class="time-tips">
            <el-text size="small" type="info">
              * 发送时间不能早于当前时间
            </el-text>
          </div>
        </el-form-item>

        <!-- 快捷时间选择 -->
        <el-form-item label="快捷选择">
          <el-button-group>
            <el-button size="small" @click="setQuickTime(1)">1小时后</el-button>
            <el-button size="small" @click="setQuickTime(24)">1天后</el-button>
            <el-button size="small" @click="setQuickTime(168)">1周后</el-button>
            <el-button size="small" @click="setQuickTime(720)">1个月后</el-button>
          </el-button-group>
        </el-form-item>

        <!-- 提交按钮 -->
        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            创建定时消息
          </el-button>
          <el-button @click="resetForm">重置</el-button>
          <el-button @click="saveDraft" :loading="savingDraft">
            保存草稿
          </el-button>
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

    <!-- 创建结果对话框 -->
    <el-dialog v-model="showResult" title="创建结果" width="500px">
      <div v-if="createResult">
        <el-result
          :icon="createResult.success ? 'success' : 'error'"
          :title="createResult.success ? '创建成功' : '创建失败'"
          :sub-title="createResult.message"
        >
          <template v-if="createResult.success && createResult.scheduleId" #extra>
            <p>定时消息ID: {{ createResult.scheduleId }}</p>
            <el-button type="primary" @click="goToScheduledList">
              查看定时消息列表
            </el-button>
          </template>
        </el-result>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter, useRoute } from 'vue-router'
import type { FormInstance, FormRules } from 'element-plus'
import dayjs from 'dayjs'
import { createScheduledMessage } from '@/api/scheduled'
import { getTemplate, getTemplateList } from '@/api/message'
import { getUserList, findUsersByTags, getTagStatistics } from '@/api/user'
import type { CreateScheduledMessageReq, Template, User, TagStatistic } from '@/types'

const router = useRouter()
const route = useRoute()

// 表单引用
const formRef = ref<FormInstance>()

// 接收者类型
const receiverType = ref<'direct' | 'users' | 'tags'>('direct')

// 表单数据
const form = reactive<CreateScheduledMessageReq & { template_data: Record<string, string> }>({
  to: '',
  user_ids: [],
  tags: [],
  template_id: '',
  template_data: {},
  scheduled_time: ''
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
  template_id: [{ required: true, message: '请选择消息模板', trigger: 'change' }],
  scheduled_time: [{ required: true, message: '请选择发送时间', trigger: 'change' }]
}

// 状态
const submitting = ref(false)
const savingDraft = ref(false)
const searchingUsers = ref(false)
const previewingUsers = ref(false)
const showUserPreview = ref(false)
const showResult = ref(false)

// 数据
const userOptions = ref<User[]>([])
const availableTags = ref<TagStatistic[]>([])
const templates = ref<Template[]>([])
const selectedTemplate = ref<Template>()
const templateVariables = ref<string[]>([])
const previewUsers = ref<User[]>([])
const createResult = ref<{
  success: boolean
  message: string
  scheduleId?: string
}>()

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

// 获取渠道标签类型
const getChannelTagType = (channel: number) => {
  const typeMap: Record<number, string> = {
    1: 'primary',
    2: 'success',
    3: 'info',
    4: 'warning',
    5: 'danger'
  }
  return typeMap[channel] || ''
}

// 获取用户显示名称
const getUserDisplayName = (userId: string) => {
  const user = userOptions.value.find(u => u.user_id === userId)
  return user ? `${user.name} (${user.user_id})` : userId
}

// 接收者类型变化处理
const handleReceiverTypeChange = () => {
  form.to = ''
  form.user_ids = []
  form.tags = []
}

// 搜索用户
const searchUsers = async (query: string) => {
  searchingUsers.value = true
  try {
    console.log('定时消息页面：开始搜索用户，关键词:', query)
    const response = await getUserList({
      page: 1,
      page_size: 50
    })
    console.log('定时消息页面：用户列表响应:', response)

    const responseData = response.data as any
    let users: any[] = []

    // 智能数据访问
    if (responseData && responseData.users && Array.isArray(responseData.users)) {
      users = responseData.users
    } else if (Array.isArray(responseData)) {
      users = responseData
    } else {
      console.log('定时消息页面：用户数据格式不正确')
      users = []
    }

    console.log('定时消息页面：获取到用户数量:', users.length)

    // 如果有搜索关键词，进行过滤
    if (query) {
      users = users.filter(user =>
        user.name.toLowerCase().includes(query.toLowerCase()) ||
        user.user_id.toLowerCase().includes(query.toLowerCase()) ||
        (user.email && user.email.toLowerCase().includes(query.toLowerCase())) ||
        (user.mobile && user.mobile.includes(query))
      )
      console.log('定时消息页面：过滤后用户数量:', users.length)
    }

    userOptions.value = users
    console.log('定时消息页面：最终用户选项:', userOptions.value)
  } catch (error) {
    console.error('定时消息页面：搜索用户失败:', error)
    ElMessage.error('搜索用户失败')
    userOptions.value = []
  } finally {
    searchingUsers.value = false
  }
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
    templates.value = templateData.map(template => ({
      relTemplateID: template.TemplateID,
      sourceID: template.SourceID,
      signName: template.SignName,
      name: template.Name,
      subject: template.Subject,
      channel: template.Channel,
      content: template.Content
    }))

    if (templates.value.length === 0) {
      ElMessage.info('请先创建消息模板')
    }
  } catch (error) {
    ElMessage.error('加载模板失败')
  }
}

// 模板变化处理
const handleTemplateChange = async (templateId: string) => {
  if (!templateId) {
    selectedTemplate.value = undefined
    templateVariables.value = []
    form.template_data = {}
    return
  }

  try {
    const response = await getTemplate(templateId)
    selectedTemplate.value = response.data

    // 解析模板中的变量
    const content = selectedTemplate.value?.content || ''
    const matches = content.match(/\{\{(\w+)\}\}/g)
    if (matches) {
      templateVariables.value = matches.map(match => match.replace(/[{}]/g, ''))
      // 初始化模板数据
      const newTemplateData: Record<string, string> = {}
      templateVariables.value.forEach(variable => {
        newTemplateData[variable] = ''
      })
      form.template_data = newTemplateData
    } else {
      templateVariables.value = []
      form.template_data = {}
    }
  } catch (error) {
    ElMessage.error('获取模板信息失败')
  }
}

// 禁用日期
const disabledDate = (time: Date) => {
  return time.getTime() < Date.now() - 24 * 60 * 60 * 1000
}

// 禁用小时
const disabledHours = () => {
  const now = new Date()
  const selectedDate = form.scheduled_time ? new Date(form.scheduled_time) : null

  if (!selectedDate || selectedDate.toDateString() !== now.toDateString()) {
    return []
  }

  const hours = []
  for (let i = 0; i < now.getHours(); i++) {
    hours.push(i)
  }
  return hours
}

// 禁用分钟
const disabledMinutes = (hour: number) => {
  const now = new Date()
  const selectedDate = form.scheduled_time ? new Date(form.scheduled_time) : null

  if (!selectedDate ||
      selectedDate.toDateString() !== now.toDateString() ||
      hour !== now.getHours()) {
    return []
  }

  const minutes = []
  for (let i = 0; i <= now.getMinutes(); i++) {
    minutes.push(i)
  }
  return minutes
}

// 设置快捷时间
const setQuickTime = (hours: number) => {
  const now = dayjs()
  form.scheduled_time = now.add(hours, 'hour').format('YYYY-MM-DD HH:mm:ss')
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()

    // 验证模板数据
    if (templateVariables.value.length > 0) {
      const missingVariables = templateVariables.value.filter(
        variable => !form.template_data[variable]?.trim()
      )
      if (missingVariables.length > 0) {
        ElMessage.error(`请填写模板变量: ${missingVariables.join(', ')}`)
        return
      }
    }

    submitting.value = true

    const requestData: CreateScheduledMessageReq = {
      template_id: form.template_id,
      template_data: form.template_data,
      scheduled_time: form.scheduled_time
    }

    if (receiverType.value === 'direct') {
      requestData.to = form.to
    } else if (receiverType.value === 'users') {
      requestData.user_ids = form.user_ids
    } else if (receiverType.value === 'tags') {
      requestData.tags = form.tags
    }

    const response = await createScheduledMessage(requestData)

    createResult.value = {
      success: true,
      message: '定时消息创建成功',
      scheduleId: response.data?.schedule_id
    }
    showResult.value = true
  } catch (error: any) {
    createResult.value = {
      success: false,
      message: error.message || '创建失败'
    }
    showResult.value = true
  } finally {
    submitting.value = false
  }
}

// 重置表单
const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, {
    to: '',
    user_ids: [],
    tags: [],
    template_id: '',
    template_data: {},
    scheduled_time: ''
  })
  selectedTemplate.value = undefined
  templateVariables.value = []
  receiverType.value = 'direct'
}

// 保存草稿
const saveDraft = async () => {
  savingDraft.value = true
  try {
    // 这里可以实现保存草稿的逻辑
    localStorage.setItem('scheduled_message_draft', JSON.stringify({
      receiverType: receiverType.value,
      form: form
    }))
    ElMessage.success('草稿保存成功')
  } catch (error) {
    ElMessage.error('保存草稿失败')
  } finally {
    savingDraft.value = false
  }
}

// 跳转到定时消息列表
const goToScheduledList = () => {
  router.push('/scheduled/list')
  showResult.value = false
}

// 加载标签统计
const loadTagStatistics = async () => {
  try {
    console.log('开始加载标签统计...')
    const response = await getTagStatistics()
    console.log('标签统计响应:', response)
    console.log('response.data类型:', typeof response.data)
    console.log('response.data内容:', response.data)
    console.log('response.data是否为数组:', Array.isArray(response.data))

    const responseData = response.data as any

    if (Array.isArray(responseData)) {
      console.log('数组长度:', responseData.length)
      console.log('第一个元素:', responseData[0])
      availableTags.value = responseData
    } else if (responseData && responseData.data && Array.isArray(responseData.data)) {
      console.log('嵌套数组长度:', responseData.data.length)
      console.log('嵌套第一个元素:', responseData.data[0])
      availableTags.value = responseData.data
    } else {
      console.log('数据格式不正确，设置为空数组')
      availableTags.value = []
    }

    console.log('最终availableTags:', availableTags.value)
    console.log('标签统计加载成功，数量:', availableTags.value.length)
  } catch (error) {
    console.error('加载标签统计失败:', error)
    ElMessage.error('加载标签统计失败')
    // 设置空数组以防止页面卡住
    availableTags.value = []
  }
}

// 初始化加载用户列表
const loadInitialUsers = async () => {
  await searchUsers('')
}

// 初始化
onMounted(() => {
  loadTemplates()
  loadTagStatistics()
  loadInitialUsers()

  // 从路由参数中获取标签信息
  if (route.query.tags) {
    receiverType.value = 'tags'
    form.tags = Array.isArray(route.query.tags) ? route.query.tags : [route.query.tags as string]
  }

  // 尝试加载草稿
  const draft = localStorage.getItem('scheduled_message_draft')
  if (draft) {
    try {
      const draftData = JSON.parse(draft)
      receiverType.value = draftData.receiverType
      Object.assign(form, draftData.form)
    } catch (error) {
      console.error('加载草稿失败:', error)
    }
  }
})
</script>

<style scoped>
.scheduled-create {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.create-form {
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

.template-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.template-name {
  font-weight: 600;
}

.template-subject {
  color: #909399;
  font-size: 12px;
  margin: 0 8px;
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

.time-tips {
  margin-left: 12px;
}
</style>
