#!/bin/bash

# MsgMate 消息推送系统一键启动脚本
# 支持启动Docker组件、后端服务、前端服务

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
BACKEND_DIR="$PROJECT_ROOT/src"
FRONTEND_DIR="$PROJECT_ROOT/msgmate-frontend"

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
    echo -e "${CYAN}MsgMate 消息推送系统一键启动脚本${NC}"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -c, --config CONFIG     指定配置文件 (docker|local|test|prod)"
    echo "  -m, --mode MODE         启动模式:"
    echo "                            all     - 启动所有服务 (默认)"
    echo "                            docker  - 仅启动Docker组件"
    echo "                            backend - 仅启动后端服务"
    echo "                            frontend - 仅启动前端服务"
    echo "                            restart - 重启后端服务"
    echo "  -d, --detach           后台运行模式"
    echo "  --no-build             跳过构建步骤"
    echo "  --clean                清理并重新启动"
    echo ""
    echo "示例:"
    echo "  $0                      # 启动所有服务，使用docker配置"
    echo "  $0 -c local -m backend  # 使用local配置启动后端"
    echo "  $0 -m docker            # 仅启动Docker组件"
    echo "  $0 -m restart           # 重启后端服务"
    echo "  $0 --clean              # 清理并重新启动所有服务"
}

# 检查依赖
check_dependencies() {
    log_step "检查系统依赖..."

    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi

    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi

    # 检查Go
    if ! command -v go &> /dev/null; then
        log_warn "Go 未安装，将跳过后端编译"
        SKIP_BACKEND=true
    fi

    # 检查Node.js
    if ! command -v node &> /dev/null; then
        log_warn "Node.js 未安装，将跳过前端启动"
        SKIP_FRONTEND=true
    fi

    # 检查npm
    if ! command -v npm &> /dev/null; then
        log_warn "npm 未安装，将跳过前端启动"
        SKIP_FRONTEND=true
    fi

    log_success "依赖检查完成"
}

# 检查端口占用
check_ports() {
    log_step "检查端口占用..."

    local ports=(3306 6379 8109 3000 9092 8899)
    local occupied_ports=()

    for port in "${ports[@]}"; do
        if lsof -ti:$port &> /dev/null; then
            occupied_ports+=($port)
        fi
    done

    if [ ${#occupied_ports[@]} -gt 0 ]; then
        log_warn "以下端口被占用: ${occupied_ports[*]}"
        echo "是否要停止占用这些端口的进程? (y/N)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            for port in "${occupied_ports[@]}"; do
                log_info "停止端口 $port 上的进程..."
                lsof -ti:$port | xargs kill -9 2>/dev/null || true
            done
        fi
    fi
}

# 启动Docker组件
start_docker() {
    log_step "启动Docker组件..."

    cd "$PROJECT_ROOT"

    if [ "$CLEAN_MODE" = true ]; then
        log_info "清理Docker容器和卷..."
        docker-compose down -v --remove-orphans 2>/dev/null || true
        # 清理Kafka数据目录以避免集群ID不匹配问题
        log_info "清理Kafka和Zookeeper数据目录..."
        rm -rf docker-compose/kafka/data docker-compose/zookeeper/data docker-compose/zookeeper/logs 2>/dev/null || true
    else
        # 检查Kafka是否因为集群ID不匹配而无法启动
        if docker ps -a | grep -q msgcenter_kafka; then
            local kafka_status=$(docker inspect -f '{{.State.Status}}' msgcenter_kafka 2>/dev/null || echo "not_found")
            if [ "$kafka_status" = "exited" ]; then
                log_warn "检测到Kafka容器异常退出，检查是否需要清理数据..."
                local last_error=$(docker logs msgcenter_kafka 2>&1 | grep -i "InconsistentClusterIdException" || echo "")
                if [ ! -z "$last_error" ]; then
                    log_warn "检测到Kafka集群ID不匹配，自动清理数据目录..."
                    docker-compose down 2>/dev/null || true
                    rm -rf docker-compose/kafka/data docker-compose/zookeeper/data docker-compose/zookeeper/logs 2>/dev/null || true
                    log_info "Kafka数据已清理，将重新初始化..."
                fi
            fi
        fi
    fi

    # 在启动前再次检查并清理可能存在的旧数据
    if [ -d "docker-compose/kafka/data" ]; then
        # 检查meta.properties文件是否存在
        if [ -f "docker-compose/kafka/data/meta.properties" ]; then
            log_info "检测到Kafka旧数据，预防性清理..."
            rm -rf docker-compose/kafka/data docker-compose/zookeeper/data docker-compose/zookeeper/logs 2>/dev/null || true
        fi
    fi

    log_info "启动MySQL、Redis、Kafka等基础服务..."
    if [ "$DETACH_MODE" = true ]; then
        docker-compose up -d
    else
        docker-compose up -d
        # 等待服务启动
        sleep 10
    fi

    # 检查服务状态
    log_info "检查Docker服务状态..."
    docker-compose ps

    # 等待MySQL启动完成
    log_info "等待MySQL启动完成..."
    local retry_count=0
    while ! docker-compose exec -T mysql mysqladmin ping -h localhost -u root -prootpass &> /dev/null; do
        if [ $retry_count -ge 30 ]; then
            log_error "MySQL启动超时"
            log_error "请检查Docker日志: docker-compose logs mysql"
            exit 1
        fi
        echo -n "."
        sleep 2
        ((retry_count++))
    done
    echo ""

    # 等待Kafka启动完成
    log_info "等待Kafka启动完成..."
    retry_count=0
    while ! docker-compose exec -T kafka bash -c "kafka-topics --bootstrap-server kafka:9093 --list" &> /dev/null; do
        if [ $retry_count -ge 30 ]; then
            log_warn "Kafka启动超时，但继续启动其他服务..."
            break
        fi
        echo -n "."
        sleep 2
        ((retry_count++))
    done
    echo ""

    log_success "Docker组件启动完成"
}

# 初始化数据库
init_database() {
    log_step "初始化数据库..."

    # 检查SQL文件是否存在
    if [ -f "$PROJECT_ROOT/sql/init.sql" ]; then
        log_info "导入数据库结构..."
        docker-compose exec -T mysql mysql -u root -prootpass msgcenter_db < "$PROJECT_ROOT/sql/init.sql" 2>/dev/null || true
    fi

    if [ -f "$PROJECT_ROOT/sql/user_management.sql" ]; then
        log_info "导入用户管理数据..."
        docker-compose exec -T mysql mysql -u root -prootpass msgcenter_db < "$PROJECT_ROOT/sql/user_management.sql" 2>/dev/null || true
    fi

    log_success "数据库初始化完成"
}

# 构建后端
build_backend() {
    if [ "$SKIP_BACKEND" = true ] || [ "$NO_BUILD" = true ]; then
        return
    fi

    log_step "构建后端项目..."

    cd "$PROJECT_ROOT"

    log_info "编译Go项目..."
    go build -o msgcenter src/main.go

    log_success "后端构建完成"
}

# 启动后端
start_backend() {
    if [ "$SKIP_BACKEND" = true ]; then
        log_warn "跳过后端启动（Go未安装）"
        return
    fi

    log_step "启动后端服务..."

    cd "$PROJECT_ROOT"

    # 检查配置文件
    local config_file="./config/config-${CONFIG}.toml"
    if [ ! -f "$config_file" ]; then
        log_error "配置文件不存在: $config_file"
        exit 1
    fi

    log_info "使用配置: $CONFIG"
    log_info "后端服务将在 http://localhost:8109 启动"

    # 检查是否已有后端进程在运行
    if [ -f "log/backend.pid" ]; then
        local old_pid=$(cat log/backend.pid)
        if kill -0 "$old_pid" 2>/dev/null; then
            log_warn "后端进程已在运行 (PID: $old_pid)，先停止旧进程..."
            kill "$old_pid" 2>/dev/null || true
            sleep 2
            if kill -0 "$old_pid" 2>/dev/null; then
                kill -9 "$old_pid" 2>/dev/null || true
            fi
        fi
    fi

    if [ "$DETACH_MODE" = true ]; then
        nohup ./bin/main --config="$config_file" > log/backend.log 2>&1 &
        echo $! > log/backend.pid
        log_info "后端服务已在后台启动，PID: $(cat log/backend.pid)"
    else
        ./bin/main --config="$config_file" &
        BACKEND_PID=$!
        echo $BACKEND_PID > log/backend.pid
    fi

    # 等待后端启动
    log_info "等待后端服务启动..."
    local retry_count=0
    while ! curl -s http://localhost:8109/user/tag_statistics > /dev/null 2>&1; do
        if [ $retry_count -ge 30 ]; then
            log_error "后端服务启动超时"
            log_error "请查看日志: tail -f log/backend.log"
            exit 1
        fi
        echo -n "."
        sleep 2
        ((retry_count++))
    done
    echo ""

    log_success "后端服务启动完成"
}

# 重启后端
restart_backend() {
    if [ "$SKIP_BACKEND" = true ]; then
        log_warn "跳过后端重启（Go未安装）"
        return
    fi

    log_step "重启后端服务..."

    cd "$PROJECT_ROOT"

    # 检查配置文件
    local config_file="./config/config-${CONFIG}.toml"
    if [ ! -f "$config_file" ]; then
        log_error "配置文件不存在: $config_file"
        exit 1
    fi

    log_info "使用配置: $CONFIG"

    # 停止旧的后端进程
    if [ -f "log/backend.pid" ]; then
        local old_pid=$(cat log/backend.pid)
        if kill -0 "$old_pid" 2>/dev/null; then
            log_info "停止旧的后端进程 (PID: $old_pid)..."
            kill "$old_pid" 2>/dev/null || true
            sleep 2
            if kill -0 "$old_pid" 2>/dev/null; then
                log_warn "进程未正常退出，强制杀死..."
                kill -9 "$old_pid" 2>/dev/null || true
            fi
            log_success "旧进程已停止"
        else
            log_info "旧进程不存在或已停止"
        fi
    fi

    # 等待端口释放
    log_info "等待端口释放..."
    local retry_count=0
    while lsof -ti:8109 &> /dev/null; do
        if [ $retry_count -ge 10 ]; then
            log_warn "端口仍被占用，强制清理..."
            lsof -ti:8109 | xargs kill -9 2>/dev/null || true
            sleep 2
            break
        fi
        echo -n "."
        sleep 1
        ((retry_count++))
    done
    echo ""

    # 重新编译后端
    if [ "$NO_BUILD" != true ]; then
        log_info "重新编译后端..."
        go build -o bin/main src/main.go
        log_success "后端编译完成"
    fi

    # 启动新的后端进程
    log_info "启动新的后端进程..."
    log_info "后端服务将在 http://localhost:8109 启动"

    if [ "$DETACH_MODE" = true ]; then
        nohup ./bin/main --config="$config_file" > log/backend.log 2>&1 &
        echo $! > log/backend.pid
        log_info "后端服务已在后台启动，PID: $(cat log/backend.pid)"
    else
        ./bin/main --config="$config_file" &
        BACKEND_PID=$!
        echo $BACKEND_PID > log/backend.pid
    fi

    # 等待后端启动
    log_info "等待后端服务启动..."
    local retry_count=0
    while ! curl -s http://localhost:8109/user/tag_statistics > /dev/null 2>&1; do
        if [ $retry_count -ge 30 ]; then
            log_error "后端服务启动超时"
            log_error "请查看日志: tail -f log/backend.log"
            exit 1
        fi
        echo -n "."
        sleep 2
        ((retry_count++))
    done
    echo ""

    log_success "后端服务重启完成"
}

# 安装前端依赖
install_frontend_deps() {
    if [ "$SKIP_FRONTEND" = true ] || [ "$NO_BUILD" = true ]; then
        return
    fi

    log_step "安装前端依赖..."

    cd "$FRONTEND_DIR"

    if [ ! -d "node_modules" ] || [ "$CLEAN_MODE" = true ]; then
        log_info "安装npm依赖..."
        npm install
    else
        log_info "依赖已存在，跳过安装"
    fi

    log_success "前端依赖安装完成"
}

# 启动前端
start_frontend() {
    if [ "$SKIP_FRONTEND" = true ]; then
        log_warn "跳过前端启动（Node.js/npm未安装）"
        return
    fi

    log_step "启动前端服务..."

    cd "$FRONTEND_DIR"

    log_info "前端服务将在 http://localhost:3000 启动"

    # 检查是否已有前端进程在运行
    if [ -f "$PROJECT_ROOT/log/frontend.pid" ]; then
        local old_pid=$(cat "$PROJECT_ROOT/log/frontend.pid")
        if kill -0 "$old_pid" 2>/dev/null; then
            log_warn "前端进程已在运行 (PID: $old_pid)，先停止旧进程..."
            kill "$old_pid" 2>/dev/null || true
            sleep 2
            if kill -0 "$old_pid" 2>/dev/null; then
                kill -9 "$old_pid" 2>/dev/null || true
            fi
        fi
    fi

    if [ "$DETACH_MODE" = true ]; then
        nohup npm run dev > "$PROJECT_ROOT/log/frontend.log" 2>&1 &
        echo $! > "$PROJECT_ROOT/log/frontend.pid"
        log_info "前端服务已在后台启动，PID: $(cat "$PROJECT_ROOT/log/frontend.pid")"
    else
        npm run dev &
        FRONTEND_PID=$!
        echo $FRONTEND_PID > "$PROJECT_ROOT/log/frontend.pid"
    fi

    # 等待前端启动
    log_info "等待前端服务启动..."
    local retry_count=0
    while ! curl -s http://localhost:3000 > /dev/null 2>&1; do
        if [ $retry_count -ge 30 ]; then
            log_error "前端服务启动超时"
            log_error "请查看日志: tail -f log/frontend.log"
            exit 1
        fi
        echo -n "."
        sleep 2
        ((retry_count++))
    done
    echo ""

    log_success "前端服务启动完成"
}

# 显示服务状态
show_status() {
    echo ""
    log_success "🎉 MsgMate 消息推送系统启动完成！"
    echo ""
    echo -e "${CYAN}服务访问地址:${NC}"
    echo -e "  📱 前端界面:    ${GREEN}http://localhost:3000${NC}"
    echo -e "  🔧 后端API:     ${GREEN}http://localhost:8109${NC}"
    echo -e "  🗄️  Kafka UI:    ${GREEN}http://localhost:8899${NC}"
    echo ""
    echo -e "${CYAN}数据库连接信息:${NC}"
    echo -e "  🐬 MySQL:       ${GREEN}localhost:3306${NC} (用户名: root, 密码: rootpass)"
    echo -e "  🔴 Redis:       ${GREEN}localhost:6379${NC} (密码: redispass)"
    echo -e "  📨 Kafka:       ${GREEN}localhost:9092${NC}"
    echo ""
    echo -e "${CYAN}日志文件:${NC}"
    echo -e "  📋 后端日志:    ${GREEN}$PROJECT_ROOT/log/backend.log${NC}"
    echo -e "  📋 前端日志:    ${GREEN}$PROJECT_ROOT/log/frontend.log${NC}"
    echo ""
    echo -e "${YELLOW}停止服务: ${NC}./stop.sh"
    echo -e "${YELLOW}查看日志: ${NC}tail -f log/backend.log"
}

# 清理函数
cleanup() {
    if [ "$DETACH_MODE" != true ]; then
        log_info "正在停止服务..."
        if [ ! -z "$BACKEND_PID" ]; then
            kill $BACKEND_PID 2>/dev/null || true
        fi
        if [ ! -z "$FRONTEND_PID" ]; then
            kill $FRONTEND_PID 2>/dev/null || true
        fi
    fi
}

# 主函数
main() {
    # 默认参数
    CONFIG="docker"
    MODE="all"
    DETACH_MODE=false
    NO_BUILD=false
    CLEAN_MODE=false
    SKIP_BACKEND=false
    SKIP_FRONTEND=false

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -c|--config)
                CONFIG="$2"
                shift 2
                ;;
            -m|--mode)
                MODE="$2"
                shift 2
                ;;
            -d|--detach)
                DETACH_MODE=true
                shift
                ;;
            --no-build)
                NO_BUILD=true
                shift
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
    if [[ ! "$CONFIG" =~ ^(docker|local|test|prod)$ ]]; then
        log_error "无效的配置: $CONFIG"
        exit 1
    fi

    if [[ ! "$MODE" =~ ^(all|docker|backend|frontend|restart)$ ]]; then
        log_error "无效的模式: $MODE"
        exit 1
    fi

    # 创建日志目录
    mkdir -p "$PROJECT_ROOT/log"

    # 设置信号处理
    trap cleanup EXIT INT TERM

    # 显示启动信息
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    MsgMate 消息推送系统                      ║"
    echo "║                      一键启动脚本                           ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo -e "配置文件: ${GREEN}$CONFIG${NC}"
    echo -e "启动模式: ${GREEN}$MODE${NC}"
    echo -e "后台运行: ${GREEN}$DETACH_MODE${NC}"
    echo ""

    # 执行启动流程
    check_dependencies

    # 处理restart模式
    if [ "$MODE" = "restart" ]; then
        restart_backend
        show_status
        # 如果不是后台模式，等待用户中断
        if [ "$DETACH_MODE" != true ]; then
            log_info "按 Ctrl+C 停止所有服务"
            wait
        fi
        return
    fi

    if [ "$MODE" = "all" ] || [ "$MODE" = "docker" ]; then
        check_ports
        start_docker
        init_database
    fi

    if [ "$MODE" = "all" ] || [ "$MODE" = "backend" ]; then
        build_backend
        start_backend
    fi

    if [ "$MODE" = "all" ] || [ "$MODE" = "frontend" ]; then
        install_frontend_deps
        start_frontend
    fi

    show_status

    # 如果不是后台模式，等待用户中断
    if [ "$DETACH_MODE" != true ]; then
        log_info "按 Ctrl+C 停止所有服务"
        wait
    fi
}

# 运行主函数
main "$@"
