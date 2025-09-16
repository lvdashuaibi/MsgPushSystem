 # MsgMate 配置文件说明

## 配置文件说明

本项目提供多个环境的配置文件：

- `config-docker.toml` - Docker环境配置（推荐用于快速启动）
- `config-test.toml` - 测试环境配置
- `config-local.toml` - 本地开发配置（需要自行创建）
- `config-local.toml.example` - 本地配置模板文件

## 安全配置管理

⚠️ **重要提醒**：包含敏感信息的配置文件已被添加到 `.gitignore` 中，不会被提交到版本控制系统。

### 创建本地配置

1. 复制模板文件：
   ```bash
   cp config/config-local.toml.example config/config-local.toml
   ```

2. 编辑配置文件，填入真实的配置信息：
   - 数据库密码
   - Redis密码
   - 邮箱授权码
   - 阿里云AccessKey
   - 飞书应用ID和密钥等

### 被忽略的敏感配置文件

以下配置文件包含敏感信息，已被 `.gitignore` 忽略：
- `config-local.toml`
- `config-prod.toml`
- `config-niuge.toml`
- `*-local.toml`
- `*-prod.toml`
- `*-private.toml`

### 主要配置项

#### 基础配置
```toml
[COMMON]
port = 8109                     # 服务端口
mysql_as_mq = false            # 是否使用MySQL作为消息队列
open_cache = true              # 是否启用Redis缓存
max_retry_count = 4            # 最大重试次数

# 邮件配置
email_account = "your@email.com"      # 发送邮箱
email_auth_code = "your_auth_code"    # 邮箱授权码

# 阿里云短信配置
ali_app_id = "your_access_key"        # 阿里云AccessKey
ali_app_secret = "your_secret"        # 阿里云Secret

# 飞书配置
lark_app_id = "your_lark_app_id"      # 飞书应用ID
lark_app_secret = "your_lark_secret"  # 飞书应用密钥
```

#### 飞书配置获取方法

1. **登录飞书开放平台**：访问 [https://open.feishu.cn/](https://open.feishu.cn/)
2. **创建应用**：
   - 点击"创建应用"
   - 选择"自建应用"
   - 填写应用名称和描述
3. **获取应用凭证**：
   - 在应用详情页面，找到"凭证与基础信息"
   - 复制 `App ID` 到配置文件的 `lark_app_id`
   - 复制 `App Secret` 到配置文件的 `lark_app_secret`
4. **配置应用权限**：
   - 在"权限管理"中添加所需权限
   - 常用权限：`im:message`（发送消息）、`contact:user.id:readonly`（获取用户ID）

#### Kafka配置
```toml
[kafka]
brokers = ["地址1:端口", "地址2:端口"]  # Kafka代理服务器地址列表

# 队列配置示例
[kafka.topics.队列名称]
name = "Topic名称"      # Kafka Topic名称
ack = 0                # 确认机制：0=不等待确认，1=等待leader确认，-1=等待所有副本确认
async = true           # 是否异步发送
offset = 0             # 消费者偏移量：0=从头开始消费，-1=从最新消息开始消费
group_id = "组ID"      # 消费者组ID，可选
```

## 本地开发配置

当您在本地机器上直接运行Go程序而Kafka在Docker容器中运行时，需要特别注意以下几点：

### 1. 配置brokers地址

由于Go程序不在Docker网络中，无法通过容器名称解析地址，因此需要使用主机的IP地址或localhost：

```toml
[kafka]
brokers = ["localhost:9092"]  # 使用本地地址和映射的端口
```

### 2. 确保端口映射正确

在docker-compose.yml中，确保Kafka容器的端口已正确映射到主机：

```yaml
kafka:
  ports:
    - "9092:9092"  # 主机端口:容器端口
```

### 3. 配置Kafka监听地址

在docker-compose.yml中，确保Kafka配置了正确的监听地址：

```yaml
kafka:
  environment:
    KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092,OUTSIDE://localhost:9092
    KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,OUTSIDE:PLAINTEXT
    KAFKA_LISTENERS: PLAINTEXT://0.0.0.0:9092,OUTSIDE://0.0.0.0:9093
    KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT
```


## 故障排除

如果遇到连接问题，可以尝试以下解决方案：

1. **检查端口映射**：确认docker-compose.yml中的端口映射正确
2. **使用IP地址**：使用本机实际IP地址替代localhost
3. **修改hosts文件**：在Windows的hosts文件中添加映射`127.0.0.1 msgcenter_kafka`
4. **检查防火墙**：确保防火墙未阻止Kafka端口