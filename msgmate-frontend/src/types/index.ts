// 通用响应类型
export interface ApiResponse<T = any> {
  code: number
  msg: string
  data?: T
}

// 用户相关类型
export interface User {
  id: number
  user_id: string
  name: string
  nickname?: string
  mobile?: string
  email?: string
  lark_id?: string
  tags: string[]
  status: number
  create_time: string
  modify_time: string
}

export interface CreateUserReq {
  user_id: string
  name: string
  nickname?: string
  mobile?: string
  email?: string
  lark_id?: string
  tags?: string[]
}

export interface UpdateUserReq {
  user_id: string
  name?: string
  nickname?: string
  mobile?: string
  email?: string
  lark_id?: string
  tags?: string[]
  status?: number
}

export interface FindUsersByTagsReq {
  tags: string[]
  match_type?: 'any' | 'all'
  page?: number
  page_size?: number
}

export interface TagStatistic {
  tag: string
  count: number
}

// 消息相关类型
export interface SendMsgReq {
  to?: string                           // 直接指定接收者（手机号/邮箱等）
  user_ids?: string[]                   // 目标用户ID列表
  tags?: string[]                       // 目标标签列表
  subject?: string
  priority?: number
  templateID: string
  templateData: Record<string, string>
  sendTimestamp?: number
}

export interface CreateTemplateReq {
  sourceID: string
  name: string
  subject?: string
  signName?: string
  channel: number
  content: string
}

export interface Template {
  ID: number
  TemplateID: string
  RelTemplateID: string
  Name: string
  Content: string
  Subject: string
  Channel: number
  SourceID: string
  SignName: string
  Status: number
  Ext: string
  CreateTime: string
  ModifyTime: string
}

// 定时消息相关类型
export interface ScheduledMessage {
  id: number
  schedule_id: string
  user_ids: string[]
  tags: string[]
  template_id: string
  template_data: string
  scheduled_time: string
  status: number
  actual_send_time?: string
  create_time: string
  modify_time: string
}

export interface CreateScheduledMessageReq {
  to?: string                           // 直接指定接收者（手机号/邮箱等）
  user_ids?: string[]                   // 目标用户ID列表
  tags?: string[]                       // 目标标签列表
  template_id: string
  template_data: Record<string, string>
  scheduled_time: string
}

// 分页相关类型
export interface PaginationReq {
  page?: number
  page_size?: number
}

export interface PaginationResp<T> {
  list: T[]
  total: number
  page: number
}
