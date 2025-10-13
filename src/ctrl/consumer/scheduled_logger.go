package consumer

import (
	"fmt"
	"time"

	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/log"
)

// ScheduledLogger 定时任务专用日志器
type ScheduledLogger struct {
	prefix string
}

// NewScheduledLogger 创建定时任务日志器
func NewScheduledLogger() *ScheduledLogger {
	return &ScheduledLogger{
		prefix: "[SCHEDULED-TASK]",
	}
}

// Info 定时任务信息日志
func (l *ScheduledLogger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Infof("%s %s %s", l.prefix, l.getTimestamp(), msg)
}

// Warn 定时任务警告日志
func (l *ScheduledLogger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Warnf("%s %s ⚠️  %s", l.prefix, l.getTimestamp(), msg)
}

// Error 定时任务错误日志
func (l *ScheduledLogger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Errorf("%s %s ❌ %s", l.prefix, l.getTimestamp(), msg)
}

// Debug 定时任务调试日志
func (l *ScheduledLogger) Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Debugf("%s %s 🔍 %s", l.prefix, l.getTimestamp(), msg)
}

// Success 定时任务成功日志
func (l *ScheduledLogger) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Infof("%s %s ✅ %s", l.prefix, l.getTimestamp(), msg)
}

// Processing 定时任务处理中日志
func (l *ScheduledLogger) Processing(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Infof("%s %s 🔄 %s", l.prefix, l.getTimestamp(), msg)
}

// Scan 定时任务扫描日志
func (l *ScheduledLogger) Scan(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Debugf("%s %s 🔍 %s", l.prefix, l.getTimestamp(), msg)
}

// getTimestamp 获取格式化时间戳
func (l *ScheduledLogger) getTimestamp() string {
	return time.Now().Format("15:04:05")
}

// LogSchedulerStart 记录调度器启动
func (l *ScheduledLogger) LogSchedulerStart() {
	l.Info("🚀 定时消息调度器启动成功，扫描间隔: 10秒")
}

// LogSchedulerStop 记录调度器停止
func (l *ScheduledLogger) LogSchedulerStop() {
	l.Info("🛑 定时消息调度器已停止")
}

// LogScanStart 记录扫描开始
func (l *ScheduledLogger) LogScanStart() {
	l.Scan("开始扫描待发送的定时消息...")
}

// LogScanResult 记录扫描结果
func (l *ScheduledLogger) LogScanResult(count int) {
	if count == 0 {
		l.Scan("扫描完成，暂无待发送的定时消息")
	} else {
		l.Processing("扫描完成，发现 %d 条待发送的定时消息", count)
	}
}

// LogMessageProcessStart 记录消息处理开始
func (l *ScheduledLogger) LogMessageProcessStart(scheduleID string, scheduledTime time.Time) {
	l.Processing("开始处理定时消息 [%s]，计划时间: %s",
		scheduleID, scheduledTime.Format("2006-01-02 15:04:05"))
}

// LogMessageProcessSuccess 记录消息处理成功
func (l *ScheduledLogger) LogMessageProcessSuccess(scheduleID string, successCount, totalCount int) {
	l.Success("定时消息 [%s] 处理完成，成功发送 %d/%d 条消息",
		scheduleID, successCount, totalCount)
}

// LogUserResolution 记录用户解析结果
func (l *ScheduledLogger) LogUserResolution(scheduleID string, userCount int) {
	l.Processing("定时消息 [%s] 解析到 %d 个目标用户", scheduleID, userCount)
}

// LogSendToQueue 记录发送到队列
func (l *ScheduledLogger) LogSendToQueue(userID string, to string) {
	l.Processing("用户 [%s] 消息已发送到队列，目标: %s", userID, to)
}

// LogSendError 记录发送错误
func (l *ScheduledLogger) LogSendError(userID string, err error) {
	l.Error("用户 [%s] 消息发送失败: %s", userID, err.Error())
}

// LogStatusUpdate 记录状态更新
func (l *ScheduledLogger) LogStatusUpdate(scheduleID string, status string) {
	l.Processing("定时消息 [%s] 状态更新为: %s", scheduleID, status)
}

// LogRedisOperation 记录Redis操作
func (l *ScheduledLogger) LogRedisOperation(operation string, scheduleID string) {
	l.Debug("Redis操作 [%s]: %s", operation, scheduleID)
}

// LogTimeComparison 记录时间比较调试信息
func (l *ScheduledLogger) LogTimeComparison(scheduledTime, currentTime time.Time) {
	l.Debug("时间比较 - 计划时间: %s, 当前时间: %s",
		scheduledTime.Format("2006-01-02 15:04:05"),
		currentTime.Format("2006-01-02 15:04:05"))
}
