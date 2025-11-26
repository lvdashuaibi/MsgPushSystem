#!/bin/bash

# MsgMate 消息推送系统停止脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT=$(cd "$(dirname "$0")" && pwd)

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

log_success() {
    echo -e "${PURPLE}[SUCCESS]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo -e "${CYAN}MsgMate 消息推送系统停止脚本${NC}"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -m, --mode MODE         停止模式:"
    echo "                            all     - 停止所有服务 (默认)"
    echo "                            docker  - 仅停止Docker组件"
    echo "                            backend - 仅停止后端服务"
    echo "                            frontend - 仅停止前端服务"
    echo "  --clean                 清理Docker容器和卷"
    echo ""
    echo "示例:"
    echo "  $0                      # 停止所有服务"
    echo "  $0 -m backend           # 仅停止后端服务"
    echo "  $0 --clean              # 停止并清理所有Docker资源"
}

# 停止后端服务
stop_backend() {
    log_step "停止后端服务..."

    # 通过PID文件停止
    if [ -f "$PROJECT_ROOT/log/backend.pid" ]; then
        local pid=$(cat "$PROJECT_ROOT/log/backend.pid")
        if kill -0 "$pid" 2>/dev/null; then
            log_info "停止后端进程 (PID: $pid)..."
            kill "$pid" 2>/dev/null || true
            sleep 2
            # 如果进程仍在运行，强制杀死
            if kill -0 "$pid" 2>/dev/null; then
                log_warn "强制停止后端进程..."
                kill -9 "$pid" 2>/dev/null || true
            fi
        fi
        rm -f "$PROJECT_ROOT/log/backend.pid"
    fi

    # 通过端口停止
    if lsof -ti:8109 &> /dev/null; then
        log_info "停止占用8109端口的进程..."
        lsof -ti:8109 | xargs kill -9 2>/dev/null || true
    fi

    log_success "后端服务已停止"
}

# 停止前端服务
stop_frontend() {
    log_step "停止前端服务..."

    # 通过PID文件停止
    if [ -f "$PROJECT_ROOT/log/frontend.pid" ]; then
        local pid=$(cat "$PROJECT_ROOT/log/frontend.pid")
        if kill -0 "$pid" 2>/dev/null; then
            log_info "停止前端进程 (PID: $pid)..."
            kill "$pid" 2>/dev/null || true
            sleep 2
            # 如果进程仍在运行，强制杀死
            if kill -0 "$pid" 2>/dev/null; then
                log_warn "强制停止前端进程..."
                kill -9 "$pid" 2>/dev/null || true
            fi
        fi
        rm -f "$PROJECT_ROOT/log/frontend.pid"
    fi

    # 通过端口停止
    if lsof -ti:3000 &> /dev/null; then
        log_info "停止占用3000端口的进程..."
        lsof -ti:3000 | xargs kill -9 2>/dev/null || true
    fi

    log_success "前端服务已停止"
}

# 停止Docker服务
stop_docker() {
    log_step "停止Docker服务..."

    cd "$PROJECT_ROOT"

    if [ "$CLEAN_MODE" = true ]; then
        log_info "停止并清理Docker容器和卷..."
        docker-compose down -v --remove-orphans 2>/dev/null || true

        # 清理Kafka和Zookeeper数据目录以避免集群ID不匹配问题
        log_info "清理Kafka和Zookeeper数据目录..."
        rm -rf "./docker-compose/kafka/data" "./docker-compose/zookeeper/data" "./docker-compose/zookeeper/logs" 2>/dev/null || true

        # 清理未使用的镜像和网络
        log_info "清理未使用的Docker资源..."
        docker system prune -f 2>/dev/null || true
    else
        log_info "停止Docker容器..."
        docker-compose down 2>/dev/null || true

        # 检查是否有Kafka集群ID不匹配的问题
        if docker logs msgcenter_kafka 2>&1 | grep -q "InconsistentClusterIdException"; then
            log_warn "检测到Kafka集群ID不匹配问题"
            log_warn "建议使用 './stop.sh --clean' 清理数据后重新启动"
        fi
    fi

    log_success "Docker服务已停止"
}

# 显示状态
show_status() {
    echo ""
    log_success "🛑 MsgMate 消息推送系统已停止"
    echo ""

    # 检查端口状态
    local ports=(3306 6379 8109 3000 9092 8899)
    local running_services=()

    for port in "${ports[@]}"; do
        if lsof -ti:$port &> /dev/null; then
            running_services+=($port)
        fi
    done

    if [ ${#running_services[@]} -gt 0 ]; then
        log_warn "以下端口仍有服务运行: ${running_services[*]}"
        echo "如需强制停止，请运行: ./stop.sh --clean"
    else
        log_success "所有服务端口已释放"
    fi

    echo ""
    echo -e "${CYAN}重新启动系统: ${NC}./start.sh"
}

# 主函数
main() {
    # 默认参数
    MODE="all"
    CLEAN_MODE=false

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -m|--mode)
                MODE="$2"
                shift 2
                ;;
            --clean)
                CLEAN_MODE=true
                shift
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 验证参数
    if [[ ! "$MODE" =~ ^(all|docker|backend|frontend)$ ]]; then
        log_error "无效的模式: $MODE"
        exit 1
    fi

    # 显示停止信息
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    MsgMate 消息推送系统                      ║"
    echo "║                      停止服务脚本                           ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo -e "停止模式: ${GREEN}$MODE${NC}"
    echo -e "清理模式: ${GREEN}$CLEAN_MODE${NC}"
    echo ""

    # 执行停止流程
    if [ "$MODE" = "all" ] || [ "$MODE" = "backend" ]; then
        stop_backend
    fi

    if [ "$MODE" = "all" ] || [ "$MODE" = "frontend" ]; then
        stop_frontend
    fi

    if [ "$MODE" = "all" ] || [ "$MODE" = "docker" ]; then
        stop_docker
    fi

    show_status
}

# 运行主函数
main "$@"
