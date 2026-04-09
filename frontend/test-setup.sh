#!/bin/bash

# 前端工程化测试脚本

echo "🚀 前端工程化测试脚本"
echo "================================"
echo ""

# 检查当前目录
if [ ! -f "package.json" ]; then
    echo "❌ 错误：请在 frontend 目录下运行此脚本"
    echo "   cd frontend && ./test-setup.sh"
    exit 1
fi

echo "✅ 当前目录正确"
echo ""

# 测试 1: 检查配置文件
echo "📝 测试 1: 检查配置文件"
echo "--------------------------------"
files=(".eslintrc.js" ".prettierrc.js" "tsconfig.json" "Dockerfile" ".dockerignore" ".vscode/settings.json")
for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "   ✅ $file"
    else
        echo "   ❌ $file 不存在"
    fi
done
echo ""

# 测试 2: 检查 Husky
echo "🪝 测试 2: 检查 Husky"
echo "--------------------------------"
if [ -d ".husky" ]; then
    echo "✅ Husky 目录存在"
    if [ -f ".husky/pre-commit" ]; then
        echo "✅ pre-commit 钩子存在"
    else
        echo "⚠️  pre-commit 钩子不存在，运行: npm run prepare"
    fi
else
    echo "⚠️  Husky 未配置，运行: npm run prepare"
fi
echo ""

# 测试 3: 运行代码检查
echo "🔍 测试 3: 运行代码检查"
echo "--------------------------------"
echo "运行 ESLint..."
if npm run lint:js > /dev/null 2>&1; then
    echo "✅ ESLint 检查通过"
else
    echo "⚠️  ESLint 发现问题（这是正常的，可以运行 npm run lint:fix 修复）"
fi

echo "运行 TypeScript 检查..."
if npm run tsc > /dev/null 2>&1; then
    echo "✅ TypeScript 检查通过"
else
    echo "⚠️  TypeScript 发现类型错误"
fi
echo ""

# 测试 4: 测试 Docker
echo "🐳 测试 4: 测试 Docker"
echo "--------------------------------"
if command -v docker &> /dev/null; then
    echo "✅ Docker 已安装"
    echo "   可以运行: docker build -t genbi-frontend ."
else
    echo "❌ Docker 未安装"
    echo "   请访问: https://www.docker.com/get-started"
fi
echo ""

# 总结
echo "================================"
echo "🎉 测试完成！"
echo ""
echo "📖 下一步："
echo "   1. 阅读 QUICK_START.md（5 分钟快速上手）"
echo "   2. 阅读 CODE_QUALITY_GUIDE.md（详细指南）"
echo "   3. 运行命令验证效果"
echo ""
echo "🚀 快速命令："
echo "   npm run lint              # 完整检查"
echo "   npm run lint:fix          # 自动修复"
echo "   docker build -t genbi-frontend .  # 构建 Docker 镜像"
echo ""
