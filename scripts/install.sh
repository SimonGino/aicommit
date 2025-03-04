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
REPO="SimonGino/aicommit"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# 转换架构名称
case "$ARCH" in
    "x86_64")
        ARCH="x86_64"
        ;;
    "aarch64")
        ARCH="arm64"
        ;;
    "arm64")
        ARCH="arm64"
        ;;
    *)
        echo "不支持的架构: $ARCH"
        exit 1
        ;;
esac

# 获取最新版本
echo "正在获取最新版本..."
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "获取版本信息失败"
    exit 1
fi

echo "最新版本: $VERSION"

# 下载二进制文件
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
echo "正在下载: $DOWNLOAD_URL"

TMP_DIR=$(mktemp -d)
curl -L "$DOWNLOAD_URL" -o "$TMP_DIR/$BINARY_NAME.tar.gz"

# 解压文件
cd "$TMP_DIR"
tar xzf "$BINARY_NAME.tar.gz"

# 创建配置目录
mkdir -p "$CONFIG_DIR"

# 安装二进制文件
mv "$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# 清理临时文件
cd - > /dev/null
rm -rf "$TMP_DIR"

echo "✓ 安装完成！"
echo "现在你可以使用 'aicommit' 命令了"
echo "配置目录: $CONFIG_DIR" 