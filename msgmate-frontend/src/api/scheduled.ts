import api from './index'
import type {
  CreateScheduledMessageReq,
  ScheduledMessage,
  PaginationReq,
  ApiResponse
} from '@/types'

// 创建定时消息
export const createScheduledMessage = (data: CreateScheduledMessageReq): Promise<ApiResponse<{ schedule_id: string }>> => {
  return api.post('/scheduled/create', data)
}

// 获取定时消息
export const getScheduledMessage = (scheduleId: string): Promise<ApiResponse<{ message: ScheduledMessage }>> => {
  return api.get(`/scheduled/get?schedule_id=${scheduleId}`)
}

// 获取定时消息列表
export const getScheduledMessageList = (params: PaginationReq & { status?: number }): Promise<ApiResponse<{ messages: ScheduledMessage[], total: number, page: number }>> => {
  return api.get('/scheduled/list', { params })
}

// 取消定时消息
export const cancelScheduledMessage = (scheduleId: string): Promise<ApiResponse> => {
  return api.post('/scheduled/cancel', { schedule_id: scheduleId })
}
