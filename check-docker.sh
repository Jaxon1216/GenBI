#!/bin/bash

echo "🔍 Docker 环境诊断"
echo "================================"
echo ""

# 检查 1: Docker 是否安装
echo "📦 检查 1: Docker 是否安装"
if command -v docker &> /dev/null; then
    echo "✅ Docker 已安装"
    docker --version
else
    echo "❌ Docker 未安装"
    echo "   请访问: https://www.docker.com/products/docker-desktop"
    exit 1
fi
echo ""

# 检查 2: Docker 是否运行
echo "🐳 检查 2: Docker 是否运行"
if docker ps &> /dev/null; then
    echo "✅ Docker 正在运行"
    docker ps
else
    echo "❌ Docker 未运行"
    echo ""
    echo "解决方案："
    echo "1. 打开 Launchpad"
    echo "2. 找到 Docker 图标（蓝色鲸鱼）"
    echo "3. 点击启动"
    echo "4. 等待右上角图标静止（约 10-30 秒）"
    echo "5. 再次运行此脚本"
    exit 1
fi
echo ""

# 检查 3: Docker Compose 是否可用
echo "🔧 检查 3: Docker Compose 是否可用"
if docker-compose --version &> /dev/null; then
    echo "✅ Docker Compose 已安装"
    docker-compose --version
else
    echo "⚠️  Docker Compose 未安装（但可能内置在 Docker 中）"
    if docker compose version &> /dev/null; then
        echo "✅ 可以使用 'docker compose' 命令"
        docker compose version
    fi
fi
echo ""

# 检查 4: 配置文件是否存在
echo "📄 检查 4: 配置文件是否存在"
if [ -f "docker-compose.yml" ]; then
    echo "✅ docker-compose.yml 存在"
else
    echo "❌ docker-compose.yml 不存在"
    echo "   当前目录: $(pwd)"
    exit 1
fi
echo ""

# 总结
echo "================================"
echo "🎉 所有检查通过！"
echo ""
echo "现在可以运行："
echo "  docker-compose up -d --build"
echo ""
echo "或者使用新版命令："
echo "  docker compose up -d --build"
