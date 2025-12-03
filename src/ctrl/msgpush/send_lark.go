package msgpush

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lvdashuaibi/MsgPushSystem/src/config"
)

func GetAccessToken() (string, error) {
	url := "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/"

	// 从配置中获取飞书应用信息
	appID := config.Conf.Common.LarkAppID
	appSecret := config.Conf.Common.LarkAppSecret

	// 检查配置是否为空
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("飞书应用配置未设置，请在配置文件中设置 lark_app_id 和 lark_app_secret")
	}

	body := map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	}
	bodyJSON, _ := json.Marshal(body)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if result["code"].(float64) != 0 {
		return "", fmt.Errorf("failed to get access token: %v", result["msg"])
	}

	return result["tenant_access_token"].(string), nil
}

// SendMessage 发送普通文本消息（兼容旧接口）
func SendMessage(accessToken, to, content string) error {
	return SendTextMessage(accessToken, to, content)
}

// SendTextMessage 发送普通文本消息
func SendTextMessage(accessToken, to, content string) error {
	url := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=user_id"

	body := map[string]interface{}{
		"receive_id": to,
		"content":    fmt.Sprintf("{\"text\":\"%s\"}", content),
		"msg_type":   "text",
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if code, ok := result["code"].(float64); ok && code != 0 {
		return fmt.Errorf("发送消息失败: %v", result["msg"])
	}

	fmt.Println("飞书消息发送成功:", string(respBody))
	return nil
}

// SendRichTextMessage 发送富文本消息
func SendRichTextMessage(accessToken, to, title, content string) error {
	url := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=user_id"

	richTextContent := map[string]interface{}{
		"zh_cn": map[string]interface{}{
			"title": title,
			"content": [][]map[string]interface{}{
				{
					{
						"tag":  "text",
						"text": content,
					},
				},
			},
		},
	}

	contentJSON, _ := json.Marshal(richTextContent)

	body := map[string]interface{}{
		"receive_id": to,
		"content":    string(contentJSON),
		"msg_type":   "post",
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if code, ok := result["code"].(float64); ok && code != 0 {
		return fmt.Errorf("发送富文本消息失败: %v", result["msg"])
	}

	fmt.Println("飞书富文本消息发送成功:", string(respBody))
	return nil
}

// SendCardMessage 发送卡片消息（推荐使用，显示效果最好）
func SendCardMessage(accessToken, to, title, content string, fields map[string]string) error {
	url := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=user_id"

	// 构建字段列表
	fieldElements := []map[string]interface{}{}
	for key, value := range fields {
		fieldElements = append(fieldElements, map[string]interface{}{
			"is_short": true,
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": fmt.Sprintf("**%s**\n%s", key, value),
			},
		})
	}

	// 卡片消息内容
	cardContent := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": title,
			},
			"template": "blue",
		},
		"elements": []map[string]interface{}{
			{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": content,
				},
			},
		},
	}

	// 如果有字段，添加分割线和字段
	if len(fieldElements) > 0 {
		elements := cardContent["elements"].([]map[string]interface{})
		elements = append(elements, map[string]interface{}{
			"tag": "hr",
		})
		elements = append(elements, map[string]interface{}{
			"tag":    "div",
			"fields": fieldElements,
		})
		cardContent["elements"] = elements
	}

	contentJSON, _ := json.Marshal(cardContent)

	body := map[string]interface{}{
		"receive_id": to,
		"content":    string(contentJSON),
		"msg_type":   "interactive",
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if code, ok := result["code"].(float64); ok && code != 0 {
		return fmt.Errorf("发送卡片消息失败: %v", result["msg"])
	}

	fmt.Println("飞书卡片消息发送成功:", string(respBody))
	return nil
}

// SendCardMessageFromJSON 直接使用JSON字符串发送卡片消息（用于AI生成的卡片）
func SendCardMessageFromJSON(accessToken, to, cardJSON string) error {
	url := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=user_id"

	// 验证JSON格式
	var cardContent map[string]interface{}
	if err := json.Unmarshal([]byte(cardJSON), &cardContent); err != nil {
		return fmt.Errorf("无效的卡片JSON格式: %v", err)
	}

	body := map[string]interface{}{
		"receive_id": to,
		"content":    cardJSON,
		"msg_type":   "interactive",
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if code, ok := result["code"].(float64); ok && code != 0 {
		return fmt.Errorf("发送AI卡片消息失败: %v", result["msg"])
	}

	fmt.Println("飞书AI卡片消息发送成功:", string(respBody))
	return nil
}

// 根据手机号获取用户 OpenID
func getUserOpenID(accessToken, phone string) (string, error) {
	url := "https://open.feishu.cn/open-apis/contact/v3/users/batch_get_id"

	body := map[string]interface{}{
		"mobiles": []string{phone},
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	fmt.Println("result ", result)

	if result["code"].(float64) != 0 {
		return "", fmt.Errorf("failed to get user info: %v", result["msg"])
	}

	// 获取用户 OpenID
	data := result["data"].(map[string]interface{})
	userList := data["user_list"].([]interface{})
	if len(userList) == 0 {
		return "", fmt.Errorf("user not found")
	}
	user := userList[0].(map[string]interface{})
	return user["user_id"].(string), nil
}
