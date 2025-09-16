package data

type TaskEnum int

const (
	TASK_STATUS_PENDING    TaskEnum = 1
	TASK_STATUS_PROCESSING TaskEnum = 2
	TASK_STATUS_SUCC       TaskEnum = 3
	TASK_STATUS_FAILED     TaskEnum = 4
)

const (
	MSG_STATUS_PENDING TaskEnum = 1
	MSG_STATUS_SUCC    TaskEnum = 2
	MSG_STATUS_FAILED  TaskEnum = 3
)

const (
	TIMER_MSG_STATUS_PENDING    TaskEnum = 1
	TIMER_MSG_STATUS_PROCESSING TaskEnum = 2
	TIMER_MSG_STATUS_SUCC       TaskEnum = 3
	TIMER_MSG_STATUS_FAILED     TaskEnum = 4
)

type ChannelEnum int

const (
	Channel_EMAIL ChannelEnum = 1
	Channel_SMS   ChannelEnum = 2
	Channel_LARK  ChannelEnum = 3
)

type TemplateStatus int

const (
	TEMPLATE_STATUS_PENDING TemplateStatus = 1
	TEMPLATE_STATUS_NORMAL  TemplateStatus = 2
)

type PriorityEnum int

const (
	PRIORITY_LOW    PriorityEnum = 1
	PRIORITY_MIDDLE PriorityEnum = 2
	PRIORITY_HIGH   PriorityEnum = 3
	PRIORITY_RETRY  PriorityEnum = 4
)

func (p PriorityEnum) String() string {
	return GetPriorityStr(p)
}

const (
	REDIS_KEY_SOURCE_QUOTA           = "XMSG_source_quota_"
	REDIS_KEY_RATE_LIMIT_COUNT       = "XMSG_rate_limit_count"
	REDIS_KEY_RATE_LIMIT_COUNT_TIMER = "XMSG_rate_limit_count_timer"
	REDIS_KEY_TEMPLATE               = "XMSG_template_"
	REDIS_KEY_MES_RECORD             = "XMSG_msgrecord_"
)

func GetPriorityStr(p PriorityEnum) string {
	if p == PRIORITY_LOW {
		return "low"
	}
	if p == PRIORITY_MIDDLE {
		return "middle"
	}
	if p == PRIORITY_HIGH {
		return "high"
	}
	if p == PRIORITY_RETRY {
		return "retry"
	}
	return ""
}
