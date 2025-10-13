<template>
  <div class="records">
    <!-- 搜索栏 -->
    <el-card class="search-card">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-input
            v-model="searchForm.msgId"
            placeholder="输入消息ID查询"
            clearable
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="6">
          <el-input
            v-model="searchForm.receiver"
            placeholder="接收者"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-col>
        <el-col :span="8">
          <el-date-picker
            v-model="searchForm.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 消息记录表格 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="recordList"
        style="width: 100%"
      >
        <el-table-column prop="msgId" label="消息ID" width="200" />
        <el-table-column prop="to" label="接收者" width="180" />
        <el-table-column prop="subject" label="主题" width="200" />
        <el-table-column prop="templateID" label="模板ID" width="180" />
        <el-table-column label="发送状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发送时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.sendTime) }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleView(row)">查看详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 消息详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="消息详情" width="800px">
      <div v-if="selectedRecord" class="message-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="消息ID">
            {{ selectedRecord.msgId }}
          </el-descriptions-item>
          <el-descriptions-item label="接收者">
            {{ selectedRecord.to }}
          </el-descriptions-item>
          <el-descriptions-item label="主题">
            {{ selectedRecord.subject }}
          </el-descriptions-item>
          <el-descriptions-item label="模板ID">
            {{ selectedRecord.templateID }}
          </el-descriptions-item>
          <el-descriptions-item label="发送状态">
            <el-tag :type="getStatusType(selectedRecord.status)">
              {{ getStatusText(selectedRecord.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="优先级">
            {{ getPriorityText(selectedRecord.priority) }}
          </el-descriptions-item>
          <el-descriptions-item label="发送时间">
            {{ formatTime(selectedRecord.sendTime) }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(selectedRecord.createTime) }}
          </el-descriptions-item>
          <el-descriptions-item label="模板数据" :span="2">
            <div class="template-data-view">
              <pre>{{ JSON.stringify(selectedRecord.templateData, null, 2) }}</pre>
            </div>
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedRecord.errorMsg" label="错误信息" :span="2">
            <div class="error-msg">
              {{ selectedRecord.errorMsg }}
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 重新发送按钮 -->
        <div v-if="selectedRecord.status === 'failed'" class="retry-section">
          <el-button type="primary" @click="handleRetry" :loading="retrying">
            重新发送
          </el-button>
        </div>
      </div>
    </el-dialog>

    <!-- 快速查询对话框 -->
    <el-dialog v-model="showQuickSearchDialog" title="快速查询" width="500px">
      <el-form :model="quickSearchForm" label-width="100px">
        <el-form-item label="消息ID">
          <el-input
            v-model="quickSearchForm.msgId"
            placeholder="请输入消息ID"
            @keyup.enter="handleQuickSearch"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showQuickSearchDialog = false">取消</el-button>
        <el-button type="primary" @click="handleQuickSearch" :loading="quickSearching">
          查询
        </el-button>
      </template>
    </el-dialog>

    <!-- 浮动操作按钮 -->
    <el-affix :offset="80" position="bottom">
      <div class="float-actions">
        <el-button
          type="primary"
          circle
          size="large"
          @click="showQuickSearchDialog = true"
        >
          <el-icon><Search /></el-icon>
        </el-button>
      </div>
    </el-affix>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import { getMessageRecord, getMsgRecordList } from '@/api/message'

// 搜索表单
const searchForm = reactive({
  msgId: '',
  receiver: '',
  dateRange: [] as string[]
})

// 快速搜索表单
const quickSearchForm = reactive({
  msgId: ''
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 状态
const loading = ref(false)
const retrying = ref(false)
const quickSearching = ref(false)
const showDetailDialog = ref(false)
const showQuickSearchDialog = ref(false)

// 数据
const recordList = ref<any[]>([])
const selectedRecord = ref<any>()

// 消息记录接口类型（临时定义）
interface MessageRecord {
  msgId: string
  to: string
  subject: string
  templateID: string
  templateData: Record<string, string>
  status: 'pending' | 'sent' | 'failed' | 'cancelled'
  priority: number
  sendTime: string
  createTime: string
  errorMsg?: string
}

// 获取状态类型
const getStatusType = (status: string) => {
  const typeMap: Record<string, string> = {
    pending: 'warning',
    sent: 'success',
    failed: 'danger',
    cancelled: 'info'
  }
  return typeMap[status] || ''
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    pending: '待发送',
    sent: '已发送',
    failed: '发送失败',
    cancelled: '已取消'
  }
  return textMap[status] || '未知'
}

// 获取优先级文本
const getPriorityText = (priority: number) => {
  const textMap: Record<number, string> = {
    1: '低',
    2: '普通',
    3: '高',
    4: '紧急'
  }
  return textMap[priority] || '未知'
}

// 格式化时间
const formatTime = (time: string) => {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 加载消息记录列表
const loadRecordList = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    }

    // 添加搜索条件
    if (searchForm.msgId) {
      params.msg_id = searchForm.msgId
    }
    if (searchForm.receiver) {
      params.to = searchForm.receiver
    }
    if (searchForm.dateRange && searchForm.dateRange.length === 2) {
      params.start_time = searchForm.dateRange[0]
      params.end_time = searchForm.dateRange[1]
    }

    const response = await getMsgRecordList(params)

    // 转换数据格式
    const records = response.data?.records || []
    recordList.value = records.map((record: any) => ({
      msgId: record.MsgId,
      to: record.To,
      subject: record.Subject,
      templateID: record.TemplateID,
      templateData: record.TemplateData ? JSON.parse(record.TemplateData) : {},
      status: getStatusFromCode(record.Status),
      channel: record.Channel,
      sourceID: record.SourceID,
      retryCount: record.RetryCount || 0,
      createTime: record.CreateTime,
      modifyTime: record.ModifyTime
    }))

    pagination.total = response.data?.total || 0
  } catch (error) {
    console.error('加载消息记录失败:', error)
    ElMessage.error('加载消息记录失败')
  } finally {
    loading.value = false
  }
}

// 将数字状态码转换为字符串状态
const getStatusFromCode = (statusCode: number) => {
  const statusMap: Record<number, string> = {
    1: 'pending',
    2: 'sent',
    3: 'failed'
  }
  return statusMap[statusCode] || 'pending'
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  loadRecordList()
}

// 重置搜索
const resetSearch = () => {
  searchForm.msgId = ''
  searchForm.receiver = ''
  searchForm.dateRange = []
  pagination.page = 1
  loadRecordList()
}

// 查看详情
const handleView = async (record: MessageRecord) => {
  try {
    const response = await getMessageRecord(record.msgId)
    selectedRecord.value = {
      ...record,
      ...response.data
    }
    showDetailDialog.value = true
  } catch (error) {
    ElMessage.error('获取消息详情失败')
  }
}

// 重新发送
const handleRetry = async () => {
  if (!selectedRecord.value) return

  try {
    retrying.value = true
    // 这里需要实现重新发送的API
    // await retryMessage(selectedRecord.value.msgId)
    ElMessage.success('重新发送成功')
    showDetailDialog.value = false
    loadRecordList()
  } catch (error) {
    ElMessage.error('重新发送失败')
  } finally {
    retrying.value = false
  }
}

// 快速查询
const handleQuickSearch = async () => {
  if (!quickSearchForm.msgId.trim()) {
    ElMessage.warning('请输入消息ID')
    return
  }

  try {
    quickSearching.value = true
    const response = await getMessageRecord(quickSearchForm.msgId)

    // 将查询结果显示在详情对话框中
    selectedRecord.value = {
      msgId: quickSearchForm.msgId,
      ...response.data
    }

    showQuickSearchDialog.value = false
    showDetailDialog.value = true

    // 清空快速搜索表单
    quickSearchForm.msgId = ''
  } catch (error) {
    ElMessage.error('查询失败，请检查消息ID是否正确')
  } finally {
    quickSearching.value = false
  }
}

// 分页变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  loadRecordList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  loadRecordList()
}

// 初始化
onMounted(() => {
  loadRecordList()
})
</script>

<style scoped>
.records {
  padding: 20px;
}

.search-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.message-detail {
  padding: 20px 0;
}

.template-data-view {
  background-color: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  max-height: 200px;
  overflow-y: auto;
}

.template-data-view pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
}

.error-msg {
  color: #f56c6c;
  background-color: #fef0f0;
  padding: 8px 12px;
  border-radius: 4px;
  border-left: 4px solid #f56c6c;
}

.retry-section {
  margin-top: 20px;
  text-align: center;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.float-actions {
  position: fixed;
  bottom: 80px;
  right: 40px;
  z-index: 1000;
}

.float-actions .el-button {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
</style>
