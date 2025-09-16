<template>
  <div class="user-list">
    <!-- 搜索和操作栏 -->
    <el-card class="operation-card">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-input
            v-model="searchForm.keyword"
            placeholder="搜索用户名、邮箱、手机号"
            clearable
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select v-model="searchForm.status" placeholder="用户状态" clearable>
            <el-option label="全部" :value="undefined" />
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-col>
        <el-col :span="6">
          <el-select
            v-model="searchForm.tags"
            placeholder="选择标签筛选"
            multiple
            clearable
            collapse-tags
            collapse-tags-tooltip
          >
            <el-option
              v-for="tag in availableTags"
              :key="tag.tag"
              :label="`${tag.tag} (${tag.count})`"
              :value="tag.tag"
            />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-col>
        <el-col :span="4" class="text-right">
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            创建用户
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 用户表格 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="userList"
        style="width: 100%"
      >
        <el-table-column prop="user_id" label="用户ID" width="150" />
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column prop="nickname" label="昵称" width="120" />
        <el-table-column prop="email" label="邮箱" width="180" />
        <el-table-column prop="mobile" label="手机号" width="130" />
        <el-table-column prop="lark_id" label="飞书ID" width="150" />
        <el-table-column label="标签" width="200">
          <template #default="{ row }">
            <div class="tags-container">
              <el-tag
                v-for="tag in row.tags"
                :key="tag"
                size="small"
                style="margin-right: 4px; margin-bottom: 4px"
              >
                {{ tag }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleToggleStatus(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.create_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" @click="handleView(row)">
                <el-icon><View /></el-icon>
                查看
              </el-button>
              <el-button size="small" type="primary" @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
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

    <!-- 创建/编辑用户对话框 -->
    <el-dialog
      v-model="showUserDialog"
      :title="editingUser ? '编辑用户' : '创建用户'"
      width="600px"
      @close="resetUserForm"
    >
      <el-form
        ref="userFormRef"
        :model="userForm"
        :rules="userRules"
        label-width="100px"
      >
        <el-form-item label="用户ID" prop="user_id" required>
          <el-input
            v-model="userForm.user_id"
            placeholder="请输入用户ID"
            :disabled="!!editingUser"
          />
        </el-form-item>

        <el-form-item label="姓名" prop="name" required>
          <el-input v-model="userForm.name" placeholder="请输入姓名" />
        </el-form-item>

        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="userForm.nickname" placeholder="请输入昵称" />
        </el-form-item>

        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>

        <el-form-item label="手机号" prop="mobile">
          <el-input v-model="userForm.mobile" placeholder="请输入手机号" />
        </el-form-item>

        <el-form-item label="飞书ID" prop="lark_id">
          <el-input v-model="userForm.lark_id" placeholder="请输入飞书ID" />
        </el-form-item>

        <el-form-item label="标签" prop="tags">
          <el-select
            v-model="userForm.tags"
            placeholder="选择或输入标签"
            multiple
            filterable
            allow-create
            default-first-option
            style="width: 100%"
          >
            <el-option
              v-for="tag in availableTags"
              :key="tag.tag"
              :label="tag.tag"
              :value="tag.tag"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showUserDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitUser" :loading="submitting">
          {{ editingUser ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 查看用户详情对话框 -->
    <el-dialog v-model="showViewDialog" title="用户详情" width="600px">
      <div v-if="viewingUser" class="user-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="用户ID">
            {{ viewingUser.user_id }}
          </el-descriptions-item>
          <el-descriptions-item label="姓名">
            {{ viewingUser.name }}
          </el-descriptions-item>
          <el-descriptions-item label="昵称">
            {{ viewingUser.nickname || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="邮箱">
            {{ viewingUser.email || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="手机号">
            {{ viewingUser.mobile || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="飞书ID">
            {{ viewingUser.lark_id || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="viewingUser.status === 1 ? 'success' : 'danger'">
              {{ viewingUser.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(viewingUser.create_time) }}
          </el-descriptions-item>
          <el-descriptions-item label="标签" :span="2">
            <div class="tags-container">
              <el-tag
                v-for="tag in viewingUser.tags"
                :key="tag"
                size="small"
                style="margin-right: 8px; margin-bottom: 4px"
              >
                {{ tag }}
              </el-tag>
              <span v-if="!viewingUser.tags || viewingUser.tags.length === 0">-</span>
            </div>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>

    <!-- 批量操作对话框 -->
    <el-dialog v-model="showBatchDialog" title="按标签查找用户" width="500px">
      <el-form :model="batchForm" label-width="100px">
        <el-form-item label="标签">
          <el-select
            v-model="batchForm.tags"
            placeholder="选择标签"
            multiple
            style="width: 100%"
          >
            <el-option
              v-for="tag in availableTags"
              :key="tag.tag"
              :label="`${tag.tag} (${tag.count})`"
              :value="tag.tag"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配方式">
          <el-radio-group v-model="batchForm.matchType">
            <el-radio label="any">任意匹配</el-radio>
            <el-radio label="all">全部匹配</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBatchDialog = false">取消</el-button>
        <el-button type="primary" @click="handleBatchSearch" :loading="batchSearching">
          查找
        </el-button>
      </template>
    </el-dialog>

    <!-- 浮动操作按钮 -->
    <el-affix :offset="80" position="bottom">
      <div class="float-actions">
        <el-button
          type="success"
          circle
          size="large"
          @click="showBatchDialog = true"
        >
          <el-icon><Filter /></el-icon>
        </el-button>
      </div>
    </el-affix>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { Search, Plus, View, Edit, Delete, Filter } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import {
  createUser,
  getUser,
  updateUser,
  getUserList,
  deleteUser,
  findUsersByTags,
  getTagStatistics
} from '@/api/user'
import type { User, CreateUserReq, UpdateUserReq, FindUsersByTagsReq, TagStatistic } from '@/types'

// 表单引用
const userFormRef = ref<FormInstance>()

// 搜索表单
const searchForm = reactive({
  keyword: '',
  status: undefined as number | undefined,
  tags: [] as string[]
})

// 批量操作表单
const batchForm = reactive({
  tags: [] as string[],
  matchType: 'any' as 'any' | 'all'
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 状态
const loading = ref(false)
const submitting = ref(false)
const batchSearching = ref(false)
const showUserDialog = ref(false)
const showViewDialog = ref(false)
const showBatchDialog = ref(false)

// 数据
const userList = ref<User[]>([])
const availableTags = ref<TagStatistic[]>([])
const editingUser = ref<User>()
const viewingUser = ref<User>()

// 用户表单
const userForm = reactive<CreateUserReq>({
  user_id: '',
  name: '',
  nickname: '',
  email: '',
  mobile: '',
  lark_id: '',
  tags: []
})

// 表单验证规则
const userRules: FormRules = {
  user_id: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  email: [
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  mobile: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号格式', trigger: 'blur' }
  ]
}

// 格式化时间
const formatTime = (time: string) => {
  if (!time) return '-'
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 加载用户列表
const loadUserList = async () => {
  loading.value = true
  try {
    const response = await getUserList({
      page: pagination.page,
      page_size: pagination.pageSize
    })
    userList.value = response.data?.users || []
    pagination.total = response.data?.total || 0
  } catch (error) {
    ElMessage.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

// 加载标签统计
const loadTagStatistics = async () => {
  try {
    console.log('用户列表页面：开始加载标签统计...')
    const response = await getTagStatistics()
    console.log('用户列表页面：标签统计响应:', response)

    const responseData = response.data as any

    if (Array.isArray(responseData)) {
      console.log('用户列表页面：数组长度:', responseData.length)
      availableTags.value = responseData
    } else if (responseData && responseData.data && Array.isArray(responseData.data)) {
      console.log('用户列表页面：嵌套数组长度:', responseData.data.length)
      availableTags.value = responseData.data
    } else {
      console.log('用户列表页面：数据格式不正确，设置为空数组')
      availableTags.value = []
    }

    console.log('用户列表页面：最终availableTags:', availableTags.value)
  } catch (error) {
    console.error('用户列表页面：加载标签统计失败:', error)
    availableTags.value = []
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  loadUserList()
}

// 重置搜索
const resetSearch = () => {
  searchForm.keyword = ''
  searchForm.status = undefined
  searchForm.tags = []
  pagination.page = 1
  loadUserList()
}

// 创建用户
const handleCreate = () => {
  editingUser.value = undefined
  resetUserForm()
  showUserDialog.value = true
}

// 编辑用户
const handleEdit = (user: User) => {
  editingUser.value = user
  Object.assign(userForm, {
    user_id: user.user_id,
    name: user.name,
    nickname: user.nickname || '',
    email: user.email || '',
    mobile: user.mobile || '',
    lark_id: user.lark_id || '',
    tags: user.tags || []
  })
  showUserDialog.value = true
}

// 查看用户
const handleView = (user: User) => {
  viewingUser.value = user
  showViewDialog.value = true
}

// 切换用户状态
const handleToggleStatus = async (user: User) => {
  try {
    await updateUser({
      user_id: user.user_id,
      status: user.status
    })
    ElMessage.success('状态更新成功')
  } catch (error) {
    // 恢复原状态
    user.status = user.status === 1 ? 0 : 1
    ElMessage.error('状态更新失败')
  }
}

// 删除用户
const handleDelete = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await deleteUser(user.user_id)
    ElMessage.success('删除成功')
    loadUserList()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 提交用户表单
const handleSubmitUser = async () => {
  if (!userFormRef.value) return

  try {
    await userFormRef.value.validate()
    submitting.value = true

    if (editingUser.value) {
      // 更新用户
      await updateUser({
        user_id: userForm.user_id,
        name: userForm.name,
        nickname: userForm.nickname,
        email: userForm.email,
        mobile: userForm.mobile,
        lark_id: userForm.lark_id,
        tags: userForm.tags
      })
      ElMessage.success('更新成功')
    } else {
      // 创建用户
      await createUser(userForm)
      ElMessage.success('创建成功')
    }

    showUserDialog.value = false
    loadUserList()
    loadTagStatistics() // 重新加载标签统计
  } catch (error) {
    ElMessage.error(editingUser.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

// 重置用户表单
const resetUserForm = () => {
  editingUser.value = undefined
  Object.assign(userForm, {
    user_id: '',
    name: '',
    nickname: '',
    email: '',
    mobile: '',
    lark_id: '',
    tags: []
  })
  userFormRef.value?.resetFields()
}

// 批量搜索
const handleBatchSearch = async () => {
  if (batchForm.tags.length === 0) {
    ElMessage.warning('请选择标签')
    return
  }

  try {
    batchSearching.value = true
    const response = await findUsersByTags({
      tags: batchForm.tags,
      match_type: batchForm.matchType
    })

    userList.value = response.users || []
    pagination.total = response.count || 0
    pagination.page = 1

    showBatchDialog.value = false
    ElMessage.success(`找到 ${userList.value.length} 个用户`)
  } catch (error) {
    ElMessage.error('查找失败')
  } finally {
    batchSearching.value = false
  }
}

// 分页变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  loadUserList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  loadUserList()
}

// 初始化
onMounted(() => {
  loadUserList()
  loadTagStatistics()
})
</script>

<style scoped>
.user-list {
  padding: 20px;
}

.operation-card {
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

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.user-detail {
  padding: 20px 0;
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

.text-right {
  text-align: right;
}

.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: flex-start;
  align-items: center;
}

.action-buttons .el-button {
  margin: 0 !important;
  min-width: 60px;
  padding: 4px 8px;
  font-size: 12px;
}

.action-buttons .el-button .el-icon {
  margin-right: 2px;
}
</style>
