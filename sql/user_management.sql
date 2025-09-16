-- 用户管理功能数据库表结构
-- 创建时间: 2025-09-11
-- 说明: 新增用户管理、定时发送功能的数据库表

-- 用户信息表
CREATE TABLE t_user (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL UNIQUE COMMENT '用户唯一标识',
    name VARCHAR(100) NOT NULL COMMENT '用户姓名',
    nickname VARCHAR(100) COMMENT '用户昵称',
    mobile VARCHAR(20) COMMENT '手机号',
    email VARCHAR(100) COMMENT '邮箱地址',
    lark_id VARCHAR(100) COMMENT '飞书用户ID',
    tags JSON COMMENT '用户标签列表，格式：["朋友","家人","同事"]',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modify_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_mobile (mobile),
    INDEX idx_email (email),
    INDEX idx_status (status)
) COMMENT='用户信息表';

-- 定时消息表
CREATE TABLE t_scheduled_message (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    schedule_id VARCHAR(64) NOT NULL UNIQUE COMMENT '定时消息唯一标识',
    user_ids TEXT COMMENT '目标用户ID列表（JSON格式）',
    tags TEXT COMMENT '目标标签列表（JSON格式）',
    template_id VARCHAR(64) NOT NULL COMMENT '消息模板ID',
    template_data TEXT COMMENT '模板数据（JSON格式）',
    scheduled_time TIMESTAMP NOT NULL COMMENT '计划发送时间',
    status TINYINT DEFAULT 1 COMMENT '状态：1-待发送，2-已发送，3-已取消，4-发送失败',
    actual_send_time TIMESTAMP NULL COMMENT '实际发送时间',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modify_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_schedule_id (schedule_id),
    INDEX idx_scheduled_time (scheduled_time),
    INDEX idx_status (status)
) COMMENT='定时消息表';

-- 插入一些测试数据
INSERT INTO t_user (user_id, name, nickname, mobile, email, lark_id, tags) VALUES
('user_001', '张三', '小张', '13800138001', 'zhangsan@example.com', 'lark_zhangsan', '["朋友", "同事"]'),
('user_002', '李四', '小李', '13800138002', 'lisi@example.com', 'lark_lisi', '["家人"]'),
('user_003', '王五', '小王', '13800138003', 'wangwu@example.com', 'lark_wangwu', '["朋友", "客户"]'),
('user_004', '赵六', '小赵', '13800138004', 'zhaoliu@example.com', 'lark_zhaoliu', '["同事", "客户"]');
