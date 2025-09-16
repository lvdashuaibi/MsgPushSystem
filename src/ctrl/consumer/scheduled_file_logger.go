package consumer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ScheduledFileLogger 定时任务文件日志器
type ScheduledFileLogger struct {
	logDir      string
	currentDate string
	logFile     *os.File
	logger      *log.Logger
}

// NewScheduledFileLogger 创建定时任务文件日志器
func NewScheduledFileLogger() *ScheduledFileLogger {
	logDir := "logs/scheduled"

	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("创建日志目录失败: %v", err)
		return nil
	}

	logger := &ScheduledFileLogger{
		logDir: logDir,
	}

	// 初始化当天的日志文件
	logger.rotateLogFile()

	return logger
}

// rotateLogFile 轮转日志文件（按日期）
func (l *ScheduledFileLogger) rotateLogFile() {
	currentDate := time.Now().Format("2006-01-02")

	// 如果日期没有变化，不需要轮转
	if l.currentDate == currentDate && l.logFile != nil {
		return
	}

	// 关闭旧的日志文件
	if l.logFile != nil {
		l.logFile.Close()
	}

	// 创建新的日志文件
	logFileName := fmt.Sprintf("scheduled-task-%s.log", currentDate)
	logFilePath := filepath.Join(l.logDir, logFileName)

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("创建日志文件失败: %v", err)
		return
	}

	l.logFile = file
	l.currentDate = currentDate
	l.logger = log.New(file, "", 0) // 不使用默认的时间前缀，我们自己格式化

	// 记录日志文件轮转信息
	l.writeLog("INFO", "📁 日志文件轮转: %s", logFilePath)
}

// writeLog 写入日志的通用方法
func (l *ScheduledFileLogger) writeLog(level, format string, args ...interface{}) {
	if l == nil || l.logger == nil {
		return
	}

	// 检查是否需要轮转日志文件
	l.rotateLogFile()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	prefix := fmt.Sprintf("[%s] [SCHEDULED-TASK] [%s]", timestamp, level)
	message := fmt.Sprintf(format, args...)

	l.logger.Printf("%s %s", prefix, message)
}

// Info 信息日志
func (l *ScheduledFileLogger) Info(format string, args ...interface{}) {
	l.writeLog("INFO", "ℹ️  "+format, args...)
}

// Warn 警告日志
func (l *ScheduledFileLogger) Warn(format string, args ...interface{}) {
	l.writeLog("WARN", "⚠️  "+format, args...)
}

// Error 错误日志
func (l *ScheduledFileLogger) Error(format string, args ...interface{}) {
	l.writeLog("ERROR", "❌ "+format, args...)
}

// Debug 调试日志
func (l *ScheduledFileLogger) Debug(format string, args ...interface{}) {
	l.writeLog("DEBUG", "🔍 "+format, args...)
}

// Success 成功日志
func (l *ScheduledFileLogger) Success(format string, args ...interface{}) {
	l.writeLog("SUCCESS", "✅ "+format, args...)
}

// Processing 处理中日志
func (l *ScheduledFileLogger) Processing(format string, args ...interface{}) {
	l.writeLog("PROCESSING", "🔄 "+format, args...)
}

// Scan 扫描日志
func (l *ScheduledFileLogger) Scan(format string, args ...interface{}) {
	l.writeLog("SCAN", "🔍 "+format, args...)
}

// LogSchedulerStart 记录调度器启动
func (l *ScheduledFileLogger) LogSchedulerStart() {
	l.Info("🚀 定时消息调度器启动成功，扫描间隔: 10秒，使用Redis轮询机制")
}

// LogSchedulerStop 记录调度器停止
func (l *ScheduledFileLogger) LogSchedulerStop() {
	l.Info("🛑 定时消息调度器已停止")
}

// LogScanStart 记录扫描开始
func (l *ScheduledFileLogger) LogScanStart() {
	l.Scan("开始Redis ZSET扫描待发送的定时消息...")
}

// LogScanResult 记录扫描结果
func (l *ScheduledFileLogger) LogScanResult(count int) {
	if count == 0 {
		l.Scan("Redis扫描完成，暂无待发送的定时消息")
	} else {
		l.Processing("Redis扫描完成，发现 %d 条待发送的定时消息", count)
	}
}

// LogRedisOperation 记录Redis操作详情
func (l *ScheduledFileLogger) LogRedisOperation(operation string, details string) {
	l.Debug("Redis操作 [%s]: %s", operation, details)
}

// LogMessageProcessStart 记录消息处理开始
func (l *ScheduledFileLogger) LogMessageProcessStart(scheduleID string, scheduledTime time.Time) {
	l.Processing("开始处理定时消息 [%s]，计划时间: %s",
		scheduleID, scheduledTime.Format("2006-01-02 15:04:05"))
}

// LogMessageProcessSuccess 记录消息处理成功
func (l *ScheduledFileLogger) LogMessageProcessSuccess(scheduleID string, successCount, totalCount int) {
	l.Success("定时消息 [%s] 处理完成，成功发送 %d/%d 条消息",
		scheduleID, successCount, totalCount)
}

// LogUserResolution 记录用户解析结果
func (l *ScheduledFileLogger) LogUserResolution(scheduleID string, userCount int) {
	l.Processing("定时消息 [%s] 解析到 %d 个目标用户", scheduleID, userCount)
}

// LogSendToQueue 记录发送到队列
func (l *ScheduledFileLogger) LogSendToQueue(userID string, to string) {
	l.Processing("用户 [%s] 消息已发送到队列，目标: %s", userID, to)
}

// LogSendError 记录发送错误
func (l *ScheduledFileLogger) LogSendError(userID string, err error) {
	l.Error("用户 [%s] 消息发送失败: %s", userID, err.Error())
}

// LogStatusUpdate 记录状态更新
func (l *ScheduledFileLogger) LogStatusUpdate(scheduleID string, status string) {
	l.Processing("定时消息 [%s] 状态更新为: %s", scheduleID, status)
}

// LogTimeComparison 记录时间比较调试信息
func (l *ScheduledFileLogger) LogTimeComparison(scheduledTime, currentTime time.Time) {
	l.Debug("时间比较 - 计划时间: %s, 当前时间: %s",
		scheduledTime.Format("2006-01-02 15:04:05"),
		currentTime.Format("2006-01-02 15:04:05"))
}

// LogDatabaseQuery 记录数据库查询
func (l *ScheduledFileLogger) LogDatabaseQuery(scheduleID string, operation string) {
	l.Debug("数据库操作 [%s]: %s", operation, scheduleID)
}

// LogRedisZSetScan 记录Redis ZSET扫描详情
func (l *ScheduledFileLogger) LogRedisZSetScan(currentScore float64, foundIDs []string) {
	l.Debug("Redis ZSET扫描 - 当前时间戳: %.0f, 找到消息ID: %v", currentScore, foundIDs)
}

// Close 关闭日志文件
func (l *ScheduledFileLogger) Close() {
	if l != nil && l.logFile != nil {
		l.Info("📁 关闭定时任务日志文件")
		l.logFile.Close()
	}
}
