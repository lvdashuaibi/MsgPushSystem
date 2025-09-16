<template>
  <div class="scheduled-list">
    <!-- 搜索和操作栏 -->
    <el-card class="search-card">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-select v-model="searchForm.status" placeholder="消息状态" clearable>
            <el-option label="全部" :value="undefined" />
            <el-option label="待发送" :value="1" />
            <el-option label="已发送" :value="2" />
            <el-option label="已取消" :value="3" />
            <el-option label="发送失败" :value="4" />
          </el-select>
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
        <el-col :span="6">
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-col>
        <el-col :span="4" class="text-right">
          <el-button type="primary" @click="$router.push('/scheduled/create')">
            <el-icon><Plus /></el-icon>
            创建定时消息
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 定时消息表格 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="messageList"
        style="width: 100%"
      >
        <el-table-column prop="schedule_id" label="调度ID" width="200" />
        <el-table-column label="目标用户" width="150">
          <template #default="{ row }">
            <div v-if="row.user_ids && row.user_ids.length > 0">
              <el-tag size="small">用户: {{ row.user_ids.length }}个</el-tag>
            </div>
            <div v-if="row.tags && row.tags.length > 0" style="margin-top: 4px">
              <el-tag size="small" type="success">标签: {{ row.tags.join(', ') }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="template_id" label="模板ID" width="180" />
        <el-table-column label="计划时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.scheduled_time) }}
          </template>
        </el-table-column>
        <el-table-column label="实际发送时间" width="160">
          <template #default="{ row }">
            {{ row.actual_send_time ? formatTime(row.actual_send_time) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.create_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleView(row)">查看</el-button>
            <el-button
              v-if="row.status === 1"
              size="small"
              type="danger"
              @click="handleCancel(row)"
            >
              取消
            </el-button>
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

    <!-- 查看详情对话框 -->
    <el-dialog
      v-model="showDetailDialog"
      title="定时消息详情"
      width="800px"
    >
      <div v-if="selectedMessage" class="message-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="调度ID">
            {{ selectedMessage.schedule_id }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(selectedMessage.status)">
              {{ getStatusText(selectedMessage.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="模板ID">
            {{ selectedMessage.template_id }}
          </el-descriptions-item>
          <el-descriptions-item label="计划发送时间">
            {{ formatTime(selectedMessage.scheduled_time) }}
          </el-descriptions-item>
          <el-descriptions-item label="实际发送时间">
            {{ selectedMessage.actual_send_time ? formatTime(selectedMessage.actual_send_time) : '未发送' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(selectedMessage.create_time) }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="detail-section">
          <h4>目标用户</h4>
          <div v-if="selectedMessage.user_ids && selectedMessage.user_ids.length > 0">
            <el-tag
              v-for="userId in selectedMessage.user_ids"
              :key="userId"
              style="margin-right: 8px; margin-bottom: 8px"
            >
              {{ userId }}
            </el-tag>
          </div>
          <div v-else>无指定用户</div>
        </div>

        <div class="detail-section">
          <h4>目标标签</h4>
          <div v-if="selectedMessage.tags && selectedMessage.tags.length > 0">
            <el-tag
              v-for="tag in selectedMessage.tags"
              :key="tag"
              type="success"
              style="margin-right: 8px; margin-bottom: 8px"
            >
              {{ tag }}
            </el-tag>
          </div>
          <div v-else>无指定标签</div>
        </div>

        <div class="detail-section">
          <h4>模板数据</h4>
          <el-input
            :value="formatTemplateData(selectedMessage.template_data)"
            type="textarea"
            :rows="6"
            readonly
          />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus } from '@element-plus/icons-vue'
import { getScheduledMessageList, cancelScheduledMessage } from '@/api/scheduled'
import type { ScheduledMessage } from '@/types'
import dayjs from 'dayjs'

// 响应式数据
const loading = ref(false)
const showDetailDialog = ref(false)
const messageList = ref<ScheduledMessage[]>([])
const selectedMessage = ref<ScheduledMessage | null>(null)

// 搜索表单
const searchForm = reactive({
  status: undefined as number | undefined,
  dateRange: [] as string[]
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 获取定时消息列表
const fetchMessageList = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    }

    if (searchForm.status !== undefined) {
      params.status = searchForm.status
    }

    const response = await getScheduledMessageList(params)
    messageList.value = response.data?.messages || []
    pagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error('获取定时消息列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchMessageList()
}

// 重置搜索
const resetSearch = () => {
  searchForm.status = undefined
  searchForm.dateRange = []
  handleSearch()
}

// 查看详情
const handleView = (message: ScheduledMessage) => {
  selectedMessage.value = message
  showDetailDialog.value = true
}

// 取消定时消息
const handleCancel = async (message: ScheduledMessage) => {
  try {
    await ElMessageBox.confirm(
      `确定要取消定时消息 "${message.schedule_id}" 吗？`,
      '确认取消',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await cancelScheduledMessage(message.schedule_id)
    ElMessage.success('取消成功')
    fetchMessageList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消失败')
    }
  }
}

// 分页变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchMessageList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchMessageList()
}

// 获取状态文本
const getStatusText = (status: number) => {
  const statusMap: Record<number, string> = {
    1: '待发送',
    2: '已发送',
    3: '已取消',
    4: '发送失败'
  }
  return statusMap[status] || '未知'
}

// 获取状态类型
const getStatusType = (status: number) => {
  const typeMap: Record<number, string> = {
    1: 'warning',
    2: 'success',
    3: 'info',
    4: 'danger'
  }
  return typeMap[status] || ''
}

// 格式化时间
const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 格式化模板数据
const formatTemplateData = (data: string) => {
  try {
    return JSON.stringify(JSON.parse(data), null, 2)
  } catch {
    return data
  }
}

// 初始化
onMounted(() => {
  fetchMessageList()
})
</script>

<style scoped>
.scheduled-list {
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

.text-right {
  text-align: right;
}

.message-detail {
  padding: 16px 0;
}

.detail-section {
  margin-top: 24px;
}

.detail-section h4 {
  margin-bottom: 12px;
  color: var(--el-text-color-primary);
  font-weight: 500;
}
</style>
