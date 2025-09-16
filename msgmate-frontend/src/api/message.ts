import api from './index'
import type {
  SendMsgReq,
  CreateTemplateReq,
  Template,
  PaginationReq,
  ApiResponse
} from '@/types'

// 发送消息
export const sendMessage = (data: SendMsgReq): Promise<ApiResponse<{ msgID: string }>> => {
  return api.post('/msg/send_msg', data)
}

// 获取消息记录
export const getMessageRecord = (msgId: string): Promise<ApiResponse<any>> => {
  return api.get(`/msg/get_msg_record?msgID=${msgId}`)
}

// 获取消息记录列表
export const getMsgRecordList = (params: PaginationReq & {
  msg_id?: string;
  to?: string;
  status?: number;
  start_time?: string;
  end_time?: string
}): Promise<ApiResponse<{ records: any[], total: number, page: number }>> => {
  return api.get('/msg/list_msg_records', { params })
}

// 创建消息模板
export const createTemplate = (data: CreateTemplateReq): Promise<ApiResponse<{ templateID: string }>> => {
  return api.post('/msg/create_template', data)
}

// 获取模板信息
export const getTemplate = (templateId: string): Promise<ApiResponse<Template>> => {
  return api.get(`/msg/get_template?templateID=${templateId}`)
}

// 更新模板
export const updateTemplate = (data: any): Promise<ApiResponse> => {
  return api.post('/msg/update_template', data)
}

// 删除模板
export const deleteTemplate = (templateId: string): Promise<ApiResponse> => {
  return api.post('/msg/del_template', { templateID: templateId })
}

// 获取模板列表
export const getTemplateList = (params: PaginationReq & { source_id?: string; channel?: number }): Promise<ApiResponse<{ templates: any[], total: number, page: number }>> => {
  return api.get('/msg/list_templates', { params })
}
