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

# 删除二进制文件
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    rm "$INSTALL_DIR/$BINARY_NAME"
    echo "✓ 已删除二进制文件"
else
    echo "警告：找不到二进制文件"
fi

# 询问是否删除配置文件
read -p "是否删除配置文件？这将删除所有API密钥和设置。(y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -d "$CONFIG_DIR" ]; then
        rm -rf "$CONFIG_DIR"
        echo "✓ 已删除配置文件"
    else
        echo "警告：找不到配置目录"
    fi
fi

echo "✓ 卸载完成！" 