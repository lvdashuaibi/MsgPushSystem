#!/bin/bash

# 性能测试脚本
# 使用方法: ./run_tests.sh [test_type]
# test_type: api | throughput | stress | all

set -e

BASE_URL="http://localhost:8109"
RESULTS_DIR="./results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# 创建结果目录
mkdir -p "$RESULTS_DIR"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查服务是否运行
check_service() {
    echo_info "检查服务状态..."
    if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        echo_info "服务运行正常"
        return 0
    else
        echo_error "服务未运行，请先启动服务"
        return 1
    fi
}

# API响应时间测试
test_api_response() {
    echo_info "=== API响应时间测试 ==="

    local result_file="$RESULTS_DIR/api_response_$TIMESTAMP.txt"

    echo_info "使用wrk进行测试..."
    echo_info "配置: 12线程, 100并发, 60秒"

    wrk -t12 -c100 -d60s \
        -s send_message.lua \
        "$BASE_URL/msg/send_msg" \
        | tee "$result_file"

    echo_info "结果已保存到: $result_file"
}

# 吞吐量测试
test_throughput() {
    echo_info "=== 吞吐量测试 ==="

    local concurrencies=(50 100 200 500)

    for concurrency in "${concurrencies[@]}"; do
        echo_info "测试并发数: $concurrency"

        local result_file="$RESULTS_DIR/throughput_c${concurrency}_$TIMESTAMP.txt"

        wrk -t12 -c"$concurrency" -d60s \
            -s send_message.lua \
            "$BASE_URL/msg/send_msg" \
            | tee "$result_file"

        echo_info "等待10秒后进行下一轮测试..."
        sleep 10
    done
}

# 压力测试
test_stress() {
    echo_info "=== 压力测试 ==="

    local concurrencies=(1000 1500 2000)

    for concurrency in "${concurrencies[@]}"; do
        echo_info "测试并发数: $concurrency"

        local result_file="$RESULTS_DIR/stress_c${concurrency}_$TIMESTAMP.txt"

        wrk -t12 -c"$concurrency" -d30s \
            -s send_message.lua \
            "$BASE_URL/msg/send_msg" \
            | tee "$result_file"

        echo_warn "等待30秒让系统恢复..."
        sleep 30
    done
}

# 长时间稳定性测试
test_stability() {
    echo_info "=== 24小时稳定性测试 ==="

    local result_file="$RESULTS_DIR/stability_24h_$TIMESTAMP.txt"
    local duration=$((24 * 60 * 60)) # 24小时

    echo_info "开始24小时稳定性测试..."
    echo_info "并发数: 100"
    echo_info "预计结束时间: $(date -d "+24 hours" "+%Y-%m-%d %H:%M:%S")"

    wrk -t12 -c100 -d"${duration}s" \
        -s send_message.lua \
        "$BASE_URL/msg/send_msg" \
        | tee "$result_file"

    echo_info "稳定性测试完成"
}

# Go测试
test_go() {
    echo_info "=== 运行Go测试 ==="

    cd "$(dirname "$0")"

    echo_info "运行API响应时间测试..."
    go test -v -run TestAPIResponseTime -timeout 10m | tee "$RESULTS_DIR/go_api_$TIMESTAMP.txt"

    echo_info "运行吞吐量测试..."
    go test -v -run TestThroughput -timeout 30m | tee "$RESULTS_DIR/go_throughput_$TIMESTAMP.txt"

    echo_info "运行优先级队列测试..."
    go test -v -run TestPriorityQueue -timeout 10m | tee "$RESULTS_DIR/go_priority_$TIMESTAMP.txt"

    echo_info "运行压力测试..."
    go test -v -run TestStressTest -timeout 30m | tee "$RESULTS_DIR/go_stress_$TIMESTAMP.txt"
}

# 生成测试报告
generate_report() {
    echo_info "=== 生成测试报告 ==="

    local report_file="$RESULTS_DIR/report_$TIMESTAMP.md"

    cat > "$report_file" << EOF
# 性能测试报告

**测试时间:** $(date "+%Y-%m-%d %H:%M:%S")

## 测试环境

- 操作系统: $(uname -s)
- CPU: $(sysctl -n machdep.cpu.brand_string 2>/dev/null || echo "Unknown")
- 内存: $(sysctl -n hw.memsize 2>/dev/null | awk '{print $1/1024/1024/1024 "GB"}' || echo "Unknown")
- Go版本: $(go version)

## 测试结果

### API响应时间测试

\`\`\`
$(cat "$RESULTS_DIR"/api_response_*.txt 2>/dev/null | tail -20)
\`\`\`

### 吞吐量测试

\`\`\`
$(cat "$RESULTS_DIR"/throughput_*.txt 2>/dev/null | tail -50)
\`\`\`

### 压力测试

\`\`\`
$(cat "$RESULTS_DIR"/stress_*.txt 2>/dev/null | tail -50)
\`\`\`

## 结论

- 系统在正常负载下表现良好
- 高并发场景下需要进一步优化
- 建议部署多实例以提高吞吐量

EOF

    echo_info "报告已生成: $report_file"
}

# 主函数
main() {
    local test_type="${1:-all}"

    echo_info "开始性能测试..."
    echo_info "测试类型: $test_type"

    # 检查服务
    if ! check_service; then
        exit 1
    fi

    case "$test_type" in
        api)
            test_api_response
            ;;
        throughput)
            test_throughput
            ;;
        stress)
            test_stress
            ;;
        stability)
            test_stability
            ;;
        go)
            test_go
            ;;
        all)
            test_api_response
            echo_info "等待10秒..."
            sleep 10

            test_throughput
            echo_info "等待30秒..."
            sleep 30

            test_stress
            ;;
        *)
            echo_error "未知的测试类型: $test_type"
            echo_info "可用类型: api, throughput, stress, stability, go, all"
            exit 1
            ;;
    esac

    generate_report

    echo_info "所有测试完成!"
    echo_info "结果保存在: $RESULTS_DIR"
}

# 运行主函数
main "$@"
