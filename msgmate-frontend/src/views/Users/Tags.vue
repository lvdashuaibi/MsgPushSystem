<template>
  <div class="user-tags">
    <!-- 标签统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon primary">
              <el-icon><Collection /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ totalTags }}</div>
              <div class="stat-label">总标签数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon success">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ totalUsers }}</div>
              <div class="stat-label">标记用户数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon warning">
              <el-icon><Star /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ mostUsedTag?.tag || '-' }}</div>
              <div class="stat-label">最热标签</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-item">
            <div class="stat-icon info">
              <el-icon><TrendCharts /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ averageTagsPerUser.toFixed(1) }}</div>
              <div class="stat-label">平均标签数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 标签管理 -->
    <el-card class="tags-card">
      <template #header>
        <div class="card-header">
          <span>标签管理</span>
          <el-button type="primary" @click="refreshTagStatistics">
            <el-icon><Refresh /></el-icon>
            刷新统计
          </el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <div class="search-bar">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索标签名称"
              clearable
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </el-col>
          <el-col :span="4">
            <el-select v-model="sortBy" placeholder="排序方式" @change="handleSort">
              <el-option label="按使用量降序" value="count_desc" />
              <el-option label="按使用量升序" value="count_asc" />
              <el-option label="按名称升序" value="name_asc" />
              <el-option label="按名称降序" value="name_desc" />
            </el-select>
          </el-col>
        </el-row>
      </div>

      <!-- 标签列表 -->
      <div class="tags-grid">
        <div
          v-for="tag in filteredTags"
          :key="tag.tag"
          class="tag-item"
          @click="handleViewTagUsers(tag)"
        >
          <div class="tag-content">
            <div class="tag-name">{{ tag.tag }}</div>
            <div class="tag-count">{{ tag.count }} 个用户</div>
          </div>
          <div class="tag-actions">
            <el-button size="small" type="primary" @click.stop="handleViewTagUsers(tag)">
              查看用户
            </el-button>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <el-empty v-if="filteredTags.length === 0" description="暂无标签数据" />
    </el-card>

    <!-- 标签用户列表对话框 -->
    <el-dialog
      v-model="showUsersDialog"
      :title="`标签 '${selectedTag?.tag}' 的用户列表`"
      width="800px"
    >
      <div v-if="selectedTag" class="tag-users">
        <div class="tag-info">
          <el-descriptions :column="3" border>
            <el-descriptions-item label="标签名称">
              {{ selectedTag.tag }}
            </el-descriptions-item>
            <el-descriptions-item label="用户数量">
              {{ selectedTag.count }}
            </el-descriptions-item>
            <el-descriptions-item label="使用率">
              {{ ((selectedTag.count / totalUsers) * 100).toFixed(1) }}%
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <div class="users-table">
          <el-table
            v-loading="loadingUsers"
            :data="tagUsers"
            style="width: 100%"
            max-height="400"
          >
            <el-table-column prop="user_id" label="用户ID" width="150" />
            <el-table-column prop="name" label="姓名" width="120" />
            <el-table-column prop="email" label="邮箱" width="180" />
            <el-table-column prop="mobile" label="手机号" width="130" />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'danger'">
                  {{ row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="其他标签" min-width="200">
              <template #default="{ row }">
                <div class="other-tags">
                  <el-tag
                    v-for="tag in row.tags.filter(t => t !== selectedTag?.tag)"
                    :key="tag"
                    size="small"
                    style="margin-right: 4px"
                  >
                    {{ tag }}
                  </el-tag>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>

      <template #footer>
        <el-button @click="showUsersDialog = false">关闭</el-button>
        <el-button type="primary" @click="handleBatchMessage">
          批量发送消息
        </el-button>
      </template>
    </el-dialog>

    <!-- 标签云图 -->
    <el-card class="tag-cloud-card">
      <template #header>
        <span>标签云图</span>
      </template>
      <div class="tag-cloud">
        <span
          v-for="tag in tagStatistics"
          :key="tag.tag"
          class="cloud-tag"
          :style="getTagCloudStyle(tag)"
          @click="handleViewTagUsers(tag)"
        >
          {{ tag.tag }}
        </span>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { getTagStatistics, findUsersByTags } from '@/api/user'
import type { TagStatistic, User } from '@/types'

const router = useRouter()

// 状态
const loading = ref(false)
const loadingUsers = ref(false)
const showUsersDialog = ref(false)

// 数据
const tagStatistics = ref<TagStatistic[]>([])
const tagUsers = ref<User[]>([])
const selectedTag = ref<TagStatistic>()

// 搜索和排序
const searchKeyword = ref('')
const sortBy = ref('count_desc')

// 计算属性
const totalTags = computed(() => tagStatistics.value.length)
const totalUsers = computed(() => {
  const uniqueUsers = new Set()
  tagStatistics.value.forEach(tag => {
    // 这里需要根据实际数据结构调整
    for (let i = 0; i < tag.count; i++) {
      uniqueUsers.add(`user_${i}`)
    }
  })
  return uniqueUsers.size
})

const mostUsedTag = computed(() => {
  return tagStatistics.value.reduce((max, tag) =>
    tag.count > (max?.count || 0) ? tag : max, null as TagStatistic | null)
})

const averageTagsPerUser = computed(() => {
  if (totalUsers.value === 0) return 0
  const totalTagUsage = tagStatistics.value.reduce((sum, tag) => sum + tag.count, 0)
  return totalTagUsage / totalUsers.value
})

// 过滤和排序后的标签
const filteredTags = computed(() => {
  let filtered = tagStatistics.value

  // 搜索过滤
  if (searchKeyword.value) {
    filtered = filtered.filter(tag =>
      tag.tag.toLowerCase().includes(searchKeyword.value.toLowerCase())
    )
  }

  // 排序
  filtered = [...filtered].sort((a, b) => {
    switch (sortBy.value) {
      case 'count_desc':
        return b.count - a.count
      case 'count_asc':
        return a.count - b.count
      case 'name_asc':
        return a.tag.localeCompare(b.tag)
      case 'name_desc':
        return b.tag.localeCompare(a.tag)
      default:
        return 0
    }
  })

  return filtered
})

// 加载标签统计
const loadTagStatistics = async () => {
  loading.value = true
  try {
    console.log('标签管理页面：开始加载标签统计...')
    const response = await getTagStatistics()
    console.log('标签管理页面：标签统计响应:', response)

    const responseData = response.data as any

    if (Array.isArray(responseData)) {
      console.log('标签管理页面：数组长度:', responseData.length)
      tagStatistics.value = responseData
    } else if (responseData && responseData.data && Array.isArray(responseData.data)) {
      console.log('标签管理页面：嵌套数组长度:', responseData.data.length)
      tagStatistics.value = responseData.data
    } else {
      console.log('标签管理页面：数据格式不正确，设置为空数组')
      tagStatistics.value = []
    }

    console.log('标签管理页面：最终tagStatistics:', tagStatistics.value)
    console.log('标签管理页面：标签统计加载成功，数量:', tagStatistics.value.length)
  } catch (error) {
    console.error('标签管理页面：加载标签统计失败:', error)
    ElMessage.error('加载标签统计失败')
    tagStatistics.value = []
  } finally {
    loading.value = false
  }
}

// 刷新统计
const refreshTagStatistics = () => {
  loadTagStatistics()
}

// 搜索处理
const handleSearch = () => {
  // 搜索逻辑已在计算属性中处理
}

// 排序处理
const handleSort = () => {
  // 排序逻辑已在计算属性中处理
}

// 查看标签用户
const handleViewTagUsers = async (tag: TagStatistic) => {
  selectedTag.value = tag
  loadingUsers.value = true
  showUsersDialog.value = true

  try {
    const response = await findUsersByTags({
      tags: [tag.tag],
      match_type: 'any'
    })
    tagUsers.value = response.users || []
  } catch (error) {
    ElMessage.error('加载用户列表失败')
    tagUsers.value = []
  } finally {
    loadingUsers.value = false
  }
}

// 批量发送消息
const handleBatchMessage = () => {
  if (!selectedTag.value) return

  // 跳转到发送消息页面，并传递标签信息
  router.push({
    path: '/messages/send',
    query: {
      tags: selectedTag.value.tag
    }
  })
  showUsersDialog.value = false
}

// 获取标签云样式
const getTagCloudStyle = (tag: TagStatistic) => {
  const maxCount = Math.max(...tagStatistics.value.map(t => t.count))
  const minCount = Math.min(...tagStatistics.value.map(t => t.count))
  const ratio = maxCount > minCount ? (tag.count - minCount) / (maxCount - minCount) : 0.5

  const fontSize = 12 + ratio * 20 // 12px - 32px
  const opacity = 0.6 + ratio * 0.4 // 0.6 - 1.0

  return {
    fontSize: `${fontSize}px`,
    opacity: opacity,
    color: `hsl(${ratio * 240}, 70%, 50%)` // 从红色到蓝色
  }
}

// 初始化
onMounted(() => {
  loadTagStatistics()
})
</script>

<style scoped>
.user-tags {
  padding: 20px;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  height: 100px;
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

.stat-icon.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.success {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.warning {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.info {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-content {
  flex: 1;
}

.stat-number {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.tags-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-bar {
  margin-bottom: 20px;
}

.tags-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.tag-item {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s;
  background: white;
}

.tag-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.tag-content {
  margin-bottom: 12px;
}

.tag-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.tag-count {
  font-size: 14px;
  color: #909399;
}

.tag-actions {
  text-align: right;
}

.tag-users {
  padding: 16px 0;
}

.tag-info {
  margin-bottom: 20px;
}

.users-table {
  margin-top: 16px;
}

.other-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tag-cloud-card {
  margin-top: 20px;
}

.tag-cloud {
  padding: 20px;
  text-align: center;
  line-height: 2;
}

.cloud-tag {
  display: inline-block;
  margin: 4px 8px;
  padding: 4px 8px;
  border-radius: 4px;
  background-color: #f5f7fa;
  cursor: pointer;
  transition: all 0.3s;
  user-select: none;
}

.cloud-tag:hover {
  background-color: #409eff;
  color: white;
  transform: scale(1.1);
}
</style>
