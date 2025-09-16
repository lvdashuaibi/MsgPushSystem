#!/bin/bash

# MsgMate 消息推送系统状态检查脚本

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

echo -e "${CYAN}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    MsgMate 系统状态检查                      ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# 检查端口状态
check_port() {
    local port=$1
    local service=$2
    if lsof -ti:$port &> /dev/null; then
        echo -e "  ${GREEN}✅ $service${NC} - 端口 $port 正在运行"
        return 0
    else
        echo -e "  ${RED}❌ $service${NC} - 端口 $port 未运行"
        return 1
    fi
}

# 检查URL状态
check_url() {
    local url=$1
    local service=$2
    if curl -s "$url" > /dev/null 2>&1; then
        echo -e "  ${GREEN}✅ $service${NC} - $url 可访问"
        return 0
    else
        echo -e "  ${RED}❌ $service${NC} - $url 不可访问"
        return 1
    fi
}

echo -e "${BLUE}[端口状态检查]${NC}"
check_port 3306 "MySQL数据库"
check_port 6379 "Redis缓存"
check_port 9092 "Kafka消息队列"
check_port 8899 "Kafka UI"
check_port 8109 "后端API服务"
check_port 3000 "前端Web服务"

echo ""
echo -e "${BLUE}[服务可访问性检查]${NC}"
check_url "http://localhost:8109/user/tag_statistics" "后端API"
check_url "http://localhost:3000" "前端界面"
check_url "http://localhost:8899" "Kafka UI"

echo ""
echo -e "${BLUE}[Docker服务状态]${NC}"
if command -v docker-compose &> /dev/null; then
    cd "$PROJECT_ROOT"
    if docker-compose ps | grep -q "Up"; then
        echo -e "  ${GREEN}✅ Docker服务${NC}"
        docker-compose ps
    else
        echo -e "  ${RED}❌ Docker服务未运行${NC}"
    fi
else
    echo -e "  ${YELLOW}⚠️  Docker Compose 未安装${NC}"
fi

echo ""
echo -e "${BLUE}[进程状态检查]${NC}"
if [ -f "$PROJECT_ROOT/log/backend.pid" ]; then
    local backend_pid=$(cat "$PROJECT_ROOT/log/backend.pid")
    if kill -0 "$backend_pid" 2>/dev/null; then
        echo -e "  ${GREEN}✅ 后端进程${NC} - PID: $backend_pid"
    else
        echo -e "  ${RED}❌ 后端进程${NC} - PID文件存在但进程未运行"
    fi
else
    echo -e "  ${YELLOW}⚠️  后端进程${NC} - 无PID文件"
fi

if [ -f "$PROJECT_ROOT/log/frontend.pid" ]; then
    local frontend_pid=$(cat "$PROJECT_ROOT/log/frontend.pid")
    if kill -0 "$frontend_pid" 2>/dev/null; then
        echo -e "  ${GREEN}✅ 前端进程${NC} - PID: $frontend_pid"
    else
        echo -e "  ${RED}❌ 前端进程${NC} - PID文件存在但进程未运行"
    fi
else
    echo -e "  ${YELLOW}⚠️  前端进程${NC} - 无PID文件"
fi

echo ""
echo -e "${BLUE}[日志文件状态]${NC}"
if [ -f "$PROJECT_ROOT/log/backend.log" ]; then
    local backend_log_size=$(du -h "$PROJECT_ROOT/log/backend.log" | cut -f1)
    echo -e "  ${GREEN}✅ 后端日志${NC} - $backend_log_size (log/backend.log)"
else
    echo -e "  ${YELLOW}⚠️  后端日志${NC} - 文件不存在"
fi

if [ -f "$PROJECT_ROOT/log/frontend.log" ]; then
    local frontend_log_size=$(du -h "$PROJECT_ROOT/log/frontend.log" | cut -f1)
    echo -e "  ${GREEN}✅ 前端日志${NC} - $frontend_log_size (log/frontend.log)"
else
    echo -e "  ${YELLOW}⚠️  前端日志${NC} - 文件不存在"
fi

echo ""
echo -e "${CYAN}[快速操作]${NC}"
echo -e "  启动系统: ${GREEN}./start.sh${NC}"
echo -e "  停止系统: ${GREEN}./stop.sh${NC}"
echo -e "  查看后端日志: ${GREEN}tail -f log/backend.log${NC}"
echo -e "  查看前端日志: ${GREEN}tail -f log/frontend.log${NC}"
echo -e "  查看Docker日志: ${GREEN}docker-compose logs -f${NC}"
