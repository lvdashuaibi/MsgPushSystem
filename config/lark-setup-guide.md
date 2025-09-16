# 飞书消息推送配置指南

## 概述

MsgMate 支持通过飞书发送消息，需要在飞书开放平台创建应用并获取相应的凭证。

## 配置步骤

### 1. 创建飞书应用

1. 访问 [飞书开放平台](https://open.feishu.cn/)
2. 使用飞书账号登录
3. 点击"创建应用" → "自建应用"
4. 填写应用信息：
   - **应用名称**：MsgMate消息推送系统
   - **应用描述**：企业消息推送服务
   - **应用图标**：上传应用图标（可选）

### 2. 获取应用凭证

在应用详情页面：

1. 进入"凭证与基础信息"页面
2. 复制以下信息到配置文件：
   ```toml
   lark_app_id = "cli_xxxxxxxxxxxxxxxxx"      # App ID
   lark_app_secret = "xxxxxxxxxxxxxxxxxxxxxxx" # App Secret
   ```

### 3. 配置应用权限

在"权限管理"页面添加以下权限：

#### 必需权限
- `im:message` - 发送消息
- `im:message.group_at_msg` - 发送群组@消息
- `contact:user.id:readonly` - 通过手机号或邮箱获取用户ID

#### 可选权限
- `im:chat` - 获取群组信息
- `contact:user.base:readonly` - 获取用户基本信息

### 4. 发布应用

1. 在"版本管理与发布"页面
2. 点击"创建版本"
3. 填写版本信息并提交审核
4. 审核通过后发布应用

### 5. 配置回调地址（可选）

如果需要接收飞书事件回调：

1. 在"事件订阅"页面
2. 配置请求网址：`https://your-domain.com/api/lark/callback`
3. 添加需要订阅的事件类型

## 配置文件示例

```toml
[COMMON]
# 飞书配置
lark_app_id = "cli_a1b2c3d4e5f6g7h8"
lark_app_secret = "abcdefghijklmnopqrstuvwxyz123456"

# 其他配置...
port = 8109
mysql_as_mq = false
open_cache = true
```

## 使用说明

### 发送消息给用户

系统支持以下方式指定飞书用户：

1. **飞书用户ID**：直接使用用户的Open ID
2. **手机号**：系统会自动通过手机号获取用户ID
3. **邮箱**：系统会自动通过邮箱获取用户ID

### 消息格式

目前支持发送文本消息，消息内容支持：
- 纯文本
- @用户（在群组中）
- 简单的格式化文本

## 故障排除

### 常见错误

1. **"飞书应用配置未设置"**
   - 检查配置文件中的 `lark_app_id` 和 `lark_app_secret` 是否正确填写

2. **"failed to get access token"**
   - 检查App ID和App Secret是否正确
   - 确认应用已发布且状态正常

3. **"user not found"**
   - 检查用户手机号或邮箱是否正确
   - 确认用户已加入企业飞书组织

4. **权限不足错误**
   - 检查应用是否已添加必要权限
   - 确认应用已通过审核并发布

### 调试建议

1. 查看后端日志获取详细错误信息
2. 在飞书开放平台查看应用调用日志
3. 使用飞书开放平台的API调试工具测试接口

## 安全建议

1. **保护应用密钥**：
   - 不要将App Secret提交到版本控制系统
   - 定期轮换应用密钥
   - 限制应用权限范围

2. **网络安全**：
   - 使用HTTPS传输
   - 配置IP白名单（如果支持）
   - 监控异常调用

3. **数据保护**：
   - 不要记录用户敏感信息
   - 遵守数据保护法规
   - 定期清理日志文件

## 参考资料

- [飞书开放平台文档](https://open.feishu.cn/document/)
- [飞书消息API文档](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/message/create)
- [飞书用户ID获取API](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/user/batch_get_id)
