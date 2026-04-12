#!/bin/bash

echo "🔧 配置 Docker 镜像加速"
echo "================================"
echo ""

echo "请按照以下步骤操作："
echo ""
echo "1️⃣ 打开 Docker Desktop"
echo ""
echo "2️⃣ 点击右上角 ⚙️ 设置图标"
echo ""
echo "3️⃣ 选择 'Docker Engine'"
echo ""
echo "4️⃣ 在 JSON 配置中添加以下内容："
echo ""
cat << 'EOF'
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://docker.nju.edu.cn",
    "https://mirror.ccs.tencentyun.com"
  ]
}
EOF
echo ""
echo "5️⃣ 点击 'Apply & Restart' 重启 Docker"
echo ""
echo "6️⃣ 等待重启完成（约 10-30 秒）"
echo ""
echo "7️⃣ 再次运行："
echo "   docker-compose up -d --build"
echo ""
echo "================================"
echo ""
echo "💡 提示：如果还是失败，使用简化版："
echo "   docker-compose -f docker-compose-simple.yml up -d --build"
