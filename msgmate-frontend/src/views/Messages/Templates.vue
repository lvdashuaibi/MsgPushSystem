<template>
  <div class="templates">
    <!-- 操作栏 -->
    <el-card class="operation-card">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-input
            v-model="searchForm.keyword"
            placeholder="搜索模板名称或内容"
            clearable
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select v-model="searchForm.channel" placeholder="选择渠道" clearable>
            <el-option label="邮件" :value="1" />
            <el-option label="短信" :value="2" />
            <el-option label="飞书" :value="3" />
            <el-option label="微信" :value="4" />
            <el-option label="钉钉" :value="5" />
          </el-select>
        </el-col>
        <el-col :span="6">
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-col>
        <el-col :span="8" class="text-right">
          <el-button type="primary" @click="handleCreate">
            <el-icon><Plus /></el-icon>
            创建模板
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 模板列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="templateList"
        style="width: 100%"
      >
        <el-table-column prop="Name" label="模板名称" width="200" />
        <el-table-column prop="Subject" label="主题" width="200" />
        <el-table-column label="渠道" width="100">
          <template #default="{ row }">
            <el-tag :type="getChannelTagType(row.Channel)">
              {{ getChannelText(row.Channel) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="SourceID" label="来源ID" width="150" />
        <el-table-column prop="SignName" label="签名" width="120" />
        <el-table-column label="内容预览" min-width="200">
          <template #default="{ row }">
            <div class="content-preview">
              {{ row.Content.length > 50 ? row.Content.substring(0, 50) + '...' : row.Content }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleView(row)">查看</el-button>
            <el-button size="small" type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
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

    <!-- 创建/编辑模板对话框 -->
    <el-dialog
      v-model="showTemplateDialog"
      :title="editingTemplate ? '编辑模板' : '创建模板'"
      width="800px"
      @close="resetTemplateForm"
    >
      <el-form
        ref="templateFormRef"
        :model="templateForm"
        :rules="templateRules"
        label-width="100px"
      >
        <el-form-item label="模板名称" prop="name" required>
          <el-input v-model="templateForm.name" placeholder="请输入模板名称" />
        </el-form-item>

        <el-form-item label="来源ID" prop="sourceID" required>
          <el-input v-model="templateForm.sourceID" placeholder="请输入来源ID，如：user-service、order-system" />
          <div class="form-help">
            <el-text size="small" type="info">
              * 来源ID用于标识模板的业务来源，建议格式：系统名-模块名
            </el-text>
          </div>
        </el-form-item>

        <el-form-item label="主题" prop="subject">
          <el-input v-model="templateForm.subject" placeholder="请输入主题" />
        </el-form-item>

        <el-form-item label="渠道" prop="channel" required>
          <el-select v-model="templateForm.channel" placeholder="选择渠道">
            <el-option label="邮件" :value="1" />
            <el-option label="短信" :value="2" />
            <el-option label="飞书" :value="3" />
            <el-option label="微信" :value="4" />
            <el-option label="钉钉" :value="5" />
          </el-select>
        </el-form-item>

        <el-form-item label="签名" prop="signName">
          <el-input v-model="templateForm.signName" placeholder="请输入签名（短信渠道必填）" />
        </el-form-item>

        <el-form-item label="模板内容" prop="content" required>
          <el-input
            v-model="templateForm.content"
            type="textarea"
            :rows="8"
            placeholder="请输入模板内容，使用 {{变量名}} 格式定义变量"
          />
          <div class="content-help">
            <p>提示：使用 {{变量名}} 格式定义变量，例如：</p>
            <p>尊敬的 {{name}}，您的订单 {{orderNo}} 已发货。</p>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showTemplateDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitTemplate" :loading="submitting">
          {{ editingTemplate ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 查看模板对话框 -->
    <el-dialog v-model="showViewDialog" title="模板详情" width="600px">
      <div v-if="viewingTemplate" class="template-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="模板ID">
            {{ viewingTemplate.TemplateID }}
          </el-descriptions-item>
          <el-descriptions-item label="模板名称">
            {{ viewingTemplate.Name }}
          </el-descriptions-item>
          <el-descriptions-item label="来源ID">
            {{ viewingTemplate.SourceID }}
          </el-descriptions-item>
          <el-descriptions-item label="主题">
            {{ viewingTemplate.Subject }}
          </el-descriptions-item>
          <el-descriptions-item label="渠道">
            <el-tag :type="getChannelTagType(viewingTemplate.Channel)">
              {{ getChannelText(viewingTemplate.Channel) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="签名">
            {{ viewingTemplate.SignName || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="模板内容" :span="2">
            <div class="template-content-view">
              {{ viewingTemplate.Content }}
            </div>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createTemplate, getTemplate, updateTemplate, deleteTemplate, getTemplateList } from '@/api/message'
import type { CreateTemplateReq, Template } from '@/types'

// 表单引用
const templateFormRef = ref<FormInstance>()

// 搜索表单
const searchForm = reactive({
  keyword: '',
  channel: undefined as number | undefined
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
const showTemplateDialog = ref(false)
const showViewDialog = ref(false)

// 数据
const templateList = ref<Template[]>([])
const editingTemplate = ref<Template>()
const viewingTemplate = ref<Template>()

// 模板表单
const templateForm = reactive<CreateTemplateReq>({
  sourceID: '',
  name: '',
  subject: '',
  signName: '',
  channel: 1,
  content: ''
})

// 表单验证规则
const templateRules: FormRules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  sourceID: [{ required: true, message: '请输入来源ID', trigger: 'blur' }],
  channel: [{ required: true, message: '请选择渠道', trigger: 'change' }],
  content: [{ required: true, message: '请输入模板内容', trigger: 'blur' }]
}

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

// 加载模板列表
const loadTemplateList = async () => {
  loading.value = true
  try {
    const response = await getTemplateList({
      page: pagination.page,
      page_size: pagination.pageSize,
      source_id: searchForm.keyword || undefined,
      channel: searchForm.channel || undefined
    })

    // 从data字段中获取模板数据
    templateList.value = response.data?.templates || []
    pagination.total = response.data?.total || 0
  } catch (error) {
    console.error('加载模板列表失败:', error)
    ElMessage.error('加载模板列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  loadTemplateList()
}

// 重置搜索
const resetSearch = () => {
  searchForm.keyword = ''
  searchForm.channel = undefined
  pagination.page = 1
  loadTemplateList()
}

// 创建模板
const handleCreate = () => {
  editingTemplate.value = undefined
  resetTemplateForm()
  showTemplateDialog.value = true
}

// 编辑模板
const handleEdit = (template: Template) => {
  editingTemplate.value = template
  Object.assign(templateForm, {
    sourceID: template.SourceID,
    name: template.Name,
    subject: template.Subject,
    signName: template.SignName,
    channel: template.Channel,
    content: template.Content
  })
  showTemplateDialog.value = true
}

// 查看模板
const handleView = (template: Template) => {
  viewingTemplate.value = template
  showViewDialog.value = true
}

// 删除模板
const handleDelete = async (template: Template) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模板 "${template.Name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await deleteTemplate(template.TemplateID)
    ElMessage.success('删除成功')
    loadTemplateList()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 提交模板表单
const handleSubmitTemplate = async () => {
  if (!templateFormRef.value) return

  try {
    await templateFormRef.value.validate()
    submitting.value = true

    if (editingTemplate.value) {
      // 更新模板
      await updateTemplate({
        templateID: editingTemplate.value.TemplateID,
        ...templateForm
      })
      ElMessage.success('更新成功')
    } else {
      // 创建模板
      await createTemplate(templateForm)
      ElMessage.success('创建成功')
    }

    showTemplateDialog.value = false
    loadTemplateList()
  } catch (error) {
    ElMessage.error(editingTemplate.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

// 重置模板表单
const resetTemplateForm = () => {
  editingTemplate.value = undefined
  Object.assign(templateForm, {
    sourceID: '',
    name: '',
    subject: '',
    signName: '',
    channel: 1,
    content: ''
  })
  templateFormRef.value?.resetFields()
}

// 分页变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  loadTemplateList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  loadTemplateList()
}

// 初始化
onMounted(() => {
  loadTemplateList()
})
</script>

<style scoped>
.templates {
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

.content-preview {
  color: #606266;
  line-height: 1.4;
}

.template-detail {
  padding: 20px 0;
}

.template-content-view {
  background-color: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  white-space: pre-wrap;
  font-family: monospace;
  max-height: 200px;
  overflow-y: auto;
}

.content-help {
  margin-top: 8px;
  padding: 8px;
  background-color: #f0f9ff;
  border-radius: 4px;
  font-size: 12px;
  color: #606266;
}

.content-help p {
  margin: 4px 0;
}

.form-help {
  margin-top: 4px;
}

.text-right {
  text-align: right;
}
</style>
