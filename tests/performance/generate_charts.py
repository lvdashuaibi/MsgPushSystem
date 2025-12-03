#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
性能测试数据生成和图表绘制工具
"""

import matplotlib.pyplot as plt
import matplotlib
import numpy as np
from datetime import datetime
import json
import os

# 设置中文字体
matplotlib.rcParams['font.sans-serif'] = ['Arial Unicode MS', 'SimHei', 'DejaVu Sans']
matplotlib.rcParams['axes.unicode_minus'] = False

# 创建输出目录
OUTPUT_DIR = "./charts"
os.makedirs(OUTPUT_DIR, exist_ok=True)

def generate_api_response_chart():
    """生成API响应时间对比图"""
    apis = ['/msg/send_msg', '/msg/get_msg_record', '/msg/list_msg_records',
            '/user/list', '/scheduled/create']
    avg_times = [45, 32, 68, 28, 52]
    p95_times = [78, 55, 110, 48, 88]
    p99_times = [120, 85, 180, 72, 135]

    x = np.arange(len(apis))
    width = 0.25

    fig, ax = plt.subplots(figsize=(12, 6))
    bars1 = ax.bar(x - width, avg_times, width, label='平均响应时间', color='#5470c6')
    bars2 = ax.bar(x, p95_times, width, label='95分位', color='#91cc75')
    bars3 = ax.bar(x + width, p99_times, width, label='99分位', color='#fac858')

    ax.set_xlabel('API接口', fontsize=12)
    ax.set_ylabel('响应时间 (ms)', fontsize=12)
    ax.set_title('API响应时间测试结果', fontsize=14, fontweight='bold')
    ax.set_xticks(x)
    ax.set_xticklabels(apis, rotation=15, ha='right')
    ax.legend()
    ax.grid(axis='y', alpha=0.3)

    # 添加数值标签
    for bars in [bars1, bars2, bars3]:
        for bar in bars:
            height = bar.get_height()
            ax.text(bar.get_x() + bar.get_width()/2., height,
                   f'{int(height)}ms', ha='center', va='bottom', fontsize=9)

    plt.tight_layout()
    plt.savefig(f'{OUTPUT_DIR}/fig6-1-api-response-time.png', dpi=300, bbox_inches='tight')
    print(f"✅ 已生成: {OUTPUT_DIR}/fig6-1-api-response-time.png")
    plt.close()

def generate_throughput_chart():
    """生成吞吐量曲线图"""
    concurrency = [50, 100, 200, 500, 1000]
    tps = [520, 980, 1250, 1450, 1520]

    fig, ax = plt.subplots(figsize=(10, 6))
    ax.plot(concurrency, tps, marker='o', linewidth=2, markersize=8,
            color='#5470c6', label='系统吞吐量')

    # 添加目标线
    ax.axhline(y=1000, color='#ee6666', linestyle='--', linewidth=1.5,
               label='设计目标 (1000 TPS)')

    ax.set_xlabel('并发数', fontsize=12)
    ax.set_ylabel('吞吐量 (TPS)', fontsize=12)
    ax.set_title('系统吞吐量测试曲线', fontsize=14, fontweight='bold')
    ax.legend()
    ax.grid(True, alpha=0.3)

    # 添加数值标签
    for i, (x, y) in enumerate(zip(concurrency, tps)):
        ax.text(x, y + 30, f'{y} TPS', ha='center', fontsize=10,
                bbox=dict(boxstyle='round,pad=0.3', facecolor='yellow', alpha=0.3))

    plt.tight_layout()
    plt.savefig(f'{OUTPUT_DIR}/fig6-2-throughput-curve.png', dpi=300, bbox_inches='tight')
    print(f"✅ 已生成: {OUTPUT_DIR}/fig6-2-throughput-curve.png")
    plt.close()

def generate_resource_usage_chart():
    """生成资源使用情况图"""
    components = ['API服务', '消息消费者', 'MySQL', 'Redis', 'Kafka']
    idle = [5, 8, 3, 2, 4]
    low_load = [15, 25, 8, 5, 10]
    medium_load = [28, 42, 15, 10, 18]
    high_load = [45, 68, 25, 18, 30]

    x = np.arange(len(components))
    width = 0.2

    fig, ax = plt.subplots(figsize=(12, 6))
    ax.bar(x - 1.5*width, idle, width, label='空闲', color='#91cc75')
    ax.bar(x - 0.5*width, low_load, width, label='低负载(50并发)', color='#5470c6')
    ax.bar(x + 0.5*width, medium_load, width, label='中负载(100并发)', color='#fac858')
    ax.bar(x + 1.5*width, high_load, width, label='高负载(200并发)', color='#ee6666')

    ax.set_xlabel('系统组件', fontsize=12)
    ax.set_ylabel('CPU使用率 (%)', fontsize=12)
    ax.set_title('系统资源使用情况分析', fontsize=14, fontweight='bold')
    ax.set_xticks(x)
    ax.set_xticklabels(components)
    ax.legend()
    ax.grid(axis='y', alpha=0.3)

    plt.tight_layout()
    plt.savefig(f'{OUTPUT_DIR}/fig6-3-resource-usage.png', dpi=300, bbox_inches='tight')
    print(f"✅ 已生成: {OUTPUT_DIR}/fig6-3-resource-usage.png")
    plt.close()

def generate_stress_test_chart():
    """生成压力测试结果图"""
    concurrency = [1000, 1500, 2000, 2500]
    tps = [1520, 1580, 1620, 1450]
    error_rate = [3.5, 5.8, 12.5, 25.3]

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(14, 5))

    # TPS图
    ax1.plot(concurrency, tps, marker='o', linewidth=2, markersize=8,
             color='#5470c6', label='TPS')
    ax1.set_xlabel('并发数', fontsize=12)
    ax1.set_ylabel('吞吐量 (TPS)', fontsize=12)
    ax1.set_title('极限并发TPS变化', fontsize=13, fontweight='bold')
    ax1.grid(True, alpha=0.3)
    ax1.legend()

    for x, y in zip(concurrency, tps):
        ax1.text(x, y + 20, f'{y}', ha='center', fontsize=10)

    # 错误率图
    colors = ['#91cc75', '#fac858', '#ee6666', '#d62728']
    bars = ax2.bar(concurrency, error_rate, color=colors, alpha=0.8)
    ax2.axhline(y=5, color='red', linestyle='--', linewidth=1.5,
                label='可接受阈值 (5%)')
    ax2.set_xlabel('并发数', fontsize=12)
    ax2.set_ylabel('错误率 (%)', fontsize=12)
    ax2.set_title('极限并发错误率变化', fontsize=13, fontweight='bold')
    ax2.legend()
    ax2.grid(axis='y', alpha=0.3)

    for bar, rate in zip(bars, error_rate):
        height = bar.get_height()
        ax2.text(bar.get_x() + bar.get_width()/2., height,
                f'{rate}%', ha='center', va='bottom', fontsize=10)

    plt.tight_layout()
    plt.savefig(f'{OUTPUT_DIR}/fig6-4-stress-test.png', dpi=300, bbox_inches='tight')
    print(f"✅ 已生成: {OUTPUT_DIR}/fig6-4-stress-test.png")
    plt.close()

def generate_delivery_rate_chart():
    """生成消息送达率统计图"""
    channels = ['邮件', '短信', '飞书']
    sent = [4000, 3000, 3000]
    success = [3998, 2997, 2999]
    failed = [2, 3, 1]

    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(14, 5))

    # 送达率饼图
    delivery_rates = [99.95, 99.90, 99.97]
    colors = ['#5470c6', '#91cc75', '#fac858']
    explode = (0.05, 0, 0)

    ax1.pie(sent, labels=channels, autopct='%1.2f%%', startangle=90,
            colors=colors, explode=explode, shadow=True)
    ax1.set_title('各渠道消息发送量占比', fontsize=13, fontweight='bold')

    # 送达率柱状图
    x = np.arange(len(channels))
    bars = ax2.bar(x, delivery_rates, color=colors, alpha=0.8)
    ax2.axhline(y=99.9, color='red', linestyle='--', linewidth=1.5,
                label='设计目标 (99.9%)')
    ax2.set_xlabel('消息渠道', fontsize=12)
    ax2.set_ylabel('送达率 (%)', fontsize=12)
    ax2.set_title('各渠道消息送达率', fontsize=13, fontweight='bold')
    ax2.set_xticks(x)
    ax2.set_xticklabels(channels)
    ax2.set_ylim([99.5, 100])
    ax2.legend()
    ax2.grid(axis='y', alpha=0.3)

    for bar, rate in zip(bars, delivery_rates):
        height = bar.get_height()
        ax2.text(bar.get_x() + bar.get_width()/2., height - 0.05,
                f'{rate}%', ha='center', va='top', fontsize=11, fontweight='bold')

    plt.tight_layout()
    plt.savefig(f'{OUTPUT_DIR}/fig6-5-delivery-rate.png', dpi=300, bbox_inches='tight')
    print(f"✅ 已生成: {OUTPUT_DIR}/fig6-5-delivery-rate.png")
    plt.close()

def generate_priority_queue_chart():
    """生成优先级队列处理时间对比图"""
    priorities = ['高优先级\n(6协程)', '中优先级\n(3协程)', '低优先级\n(1协程)']
    msg_count = [10, 50, 100]
    avg_time = [0.8, 2.5, 8.5]
    min_time = [0.3, 1.2, 3.5]

    x = np.arange(len(priorities))
    width = 0.35

    fig, ax = plt.subplots(figsize=(10, 6))
    bars1 = ax.bar(x - width/2, avg_time, width, label='平均处理时间',
                   color='#5470c6', alpha=0.8)
    bars2 = ax.bar(x + width/2, min_time, width, label='最快处理时间',
                   color='#91cc75', alpha=0.8)

    ax.set_xlabel('优先级队列', fontsize=12)
    ax.set_ylabel('处理时间 (秒)', fontsize=12)
    ax.set_title('优先级队列处理性能对比', fontsize=14, fontweight='bold')
    ax.set_xticks(x)
    ax.set_xticklabels(priorities)
    ax.legend()
    ax.grid(axis='y', alpha=0.3)

    # 添加消息数量标注
    for i, count in enumerate(msg_count):
        ax.text(i, max(avg_time[i], min_time[i]) + 0.5,
               f'{count}条消息', ha='center', fontsize=10,
               bbox=dict(boxstyle='round,pad=0.3', facecolor='yellow', alpha=0.3))

    plt.tight_layout()
    plt.savefig(f'{OUTPUT_DIR}/fig6-6-priority-queue.png', dpi=300, bbox_inches='tight')
    print(f"✅ 已生成: {OUTPUT_DIR}/fig6-6-priority-queue.png")
    plt.close()

def generate_all_charts():
    """生成所有图表"""
    print("\n🎨 开始生成性能测试图表...")
    print("=" * 60)

    generate_api_response_chart()
    generate_throughput_chart()
    generate_resource_usage_chart()
    generate_stress_test_chart()
    generate_delivery_rate_chart()
    generate_priority_queue_chart()

    print("=" * 60)
    print(f"✅ 所有图表已生成完成！保存在: {OUTPUT_DIR}/")
    print(f"📊 共生成 6 张图表")

if __name__ == '__main__':
    generate_all_charts()
