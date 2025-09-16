import api from './index'
import type {
  User,
  CreateUserReq,
  UpdateUserReq,
  FindUsersByTagsReq,
  TagStatistic,
  PaginationReq,
  PaginationResp,
  ApiResponse
} from '@/types'

// 创建用户
export const createUser = (data: CreateUserReq): Promise<ApiResponse<{ user_id: string }>> => {
  return api.post('/user/create', data)
}

// 获取用户信息
export const getUser = (userId: string): Promise<ApiResponse<{ user: User }>> => {
  return api.get(`/user/get?user_id=${userId}`)
}

// 更新用户信息
export const updateUser = (data: UpdateUserReq): Promise<ApiResponse> => {
  return api.post('/user/update', data)
}

// 获取用户列表
export const getUserList = (params: PaginationReq): Promise<ApiResponse<{ users: User[], total: number, page: number }>> => {
  return api.get('/user/list', { params })
}

// 删除用户
export const deleteUser = (userId: string): Promise<ApiResponse> => {
  return api.post('/user/delete', { user_id: userId })
}

// 按标签查找用户
export const findUsersByTags = (data: FindUsersByTagsReq): Promise<ApiResponse<{ users: User[], count: number }>> => {
  return api.post('/user/find_by_tags', data)
}

// 获取标签统计
export const getTagStatistics = (): Promise<ApiResponse<TagStatistic[]>> => {
  return api.get('/user/tag_statistics')
}
