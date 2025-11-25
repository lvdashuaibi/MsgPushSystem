package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateUUID 生成UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateID 生成唯一ID
func GenerateID() string {
	return GenerateUUID()
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// GetCurrentTimestamp 获取当前时间戳
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// GetCurrentTimeString 获取当前时间字符串
func GetCurrentTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ParseTime 解析时间字符串
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", timeStr)
}

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return len(s) == 0
}

// IsNotEmpty 检查字符串是否不为空
func IsNotEmpty(s string) bool {
	return len(s) > 0
}

// Contains 检查切片是否包含元素
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Max 返回两个整数中的最大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min 返回两个整数中的最小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SafeString 安全地获取字符串指针的值
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StringPtr 获取字符串的指针
func StringPtr(s string) *string {
	return &s
}

// IntPtr 获取整数的指针
func IntPtr(i int) *int {
	return &i
}

// BoolPtr 获取布尔值的指针
func BoolPtr(b bool) *bool {
	return &b
}

// Ternary 三元运算符
func Ternary(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// TernaryString 字符串三元运算符
func TernaryString(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

// TernaryInt 整数三元运算符
func TernaryInt(condition bool, trueVal, falseVal int) int {
	if condition {
		return trueVal
	}
	return falseVal
}

// Retry 重试函数
func Retry(attempts int, sleep time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		if i < attempts-1 {
			time.Sleep(sleep)
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
