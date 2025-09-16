package tools

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 地址
		Password: "niuge",          // Redis 密码，如果没有密码设置为空
		DB:       0,                // 使用的数据库编号，默认 0
	})
	lm := NewRateLimiter(client, 1000, 1)
	// 模拟用户请求
	userID := "user123"
	for i := 0; i < 15; i++ {
		allowed, err := lm.IsRequestAllowed(userID)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		if allowed {
			fmt.Println("Request allowed")
		} else {
			fmt.Println("Request denied")
		}
		time.Sleep(100 * time.Millisecond) // 模拟请求间隔
	}
	return
}

func TestTemplateReplace(t *testing.T) {
	result, err := TemplateReplace(
		"亲爱的 {{.user_name}}，您的订单 {{.order_id}} 已发货！",
		map[string]string{"user_name": "张三", "order_id": "12345678"})
	//	result, err := templateReplace(
	//		"亲爱的 ${UserName}，您的订单 ${OrderID} 已发货！",
	//		map[string]string{"UserName": "张三", "OrderID": "12345678"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)
}
