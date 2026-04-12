#!/bin/bash

echo "📦 手动拉取 Docker 镜像"
echo "================================"
echo ""

# 使用镜像加速
export DOCKER_REGISTRY="docker.m.daocloud.io"

echo "1️⃣ 拉取 Node.js 镜像..."
docker pull ${DOCKER_REGISTRY}/library/node:18-alpine || docker pull node:18-alpine

echo ""
echo "2️⃣ 拉取 Nginx 镜像..."
docker pull ${DOCKER_REGISTRY}/library/nginx:alpine || docker pull nginx:alpine

echo ""
echo "3️⃣ 拉取 MySQL 镜像..."
docker pull ${DOCKER_REGISTRY}/library/mysql:8.0 || docker pull mysql:8.0

echo ""
echo "4️⃣ 拉取 Maven 镜像..."
docker pull ${DOCKER_REGISTRY}/library/maven:3.8-openjdk-8 || docker pull maven:3.8-openjdk-8

echo ""
echo "5️⃣ 拉取 OpenJDK 镜像..."
docker pull ${DOCKER_REGISTRY}/library/openjdk:8-jre-alpine || docker pull openjdk:8-jre-alpine

echo ""
echo "================================"
echo "✅ 镜像拉取完成！"
echo ""
echo "现在可以运行："
echo "  docker-compose up -d --build"
