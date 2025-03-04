#!/bin/bash

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]; then
    echo "请使用sudo运行此脚本"
    exit 1
fi

# 设置变量
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="aicommit"
CONFIG_DIR="$HOME/.config/aicommit"

# 检查二进制文件是否存在
if [ ! -f "$BINARY_NAME" ]; then
    echo "错误：找不到 $BINARY_NAME 二进制文件"
    echo "请先运行 'go build -o aicommit cmd/aicommit/main.go'"
    exit 1
fi

# 创建配置目录
mkdir -p "$CONFIG_DIR"

# 复制二进制文件到安装目录
cp "$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "✓ 安装完成！"
echo "现在你可以使用 'aicommit' 命令了" 