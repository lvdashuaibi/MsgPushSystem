package msgpush

import (
	"fmt"
	"testing"
)

func TestSendEmail(t *testing.T) {
	SendEmail("871283069@qq.com", "这是测试邮件", "你的商品已发货")
	return
}

func TestSendSMS(t *testing.T) {
	SendSMS("18676382530", "牛哥web", "SMS_478635046",
		"{\"user_name\":\"niu\"}, \"order_id\" : \"12345678\"")
	return
}

func TestSendLark(t *testing.T) {
	accessToken, err := GetAccessToken()
	if err != nil {
		fmt.Println("Error getting access token:", err)
		return
	}
	userID, err := getUserOpenID(accessToken, "18676382530")
	if err != nil {
		fmt.Println("Error getting user ID:", err)
		return
	}
	fmt.Println("User ID:", userID)
	//	userID := "5efg94ff"
	content := "老虎，欢迎加入飞书组织"
	err = SendMessage(accessToken, userID, content)
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
	return
}
