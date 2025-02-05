#!/bin/bash

# 设置安装目录
INSTALL_DIR="/usr/local/bin"
BINARY_PATH="${INSTALL_DIR}/aicommit"

# 检查权限
if [[ ! -w "${INSTALL_DIR}" ]]; then
    echo "Error: No write permission to ${INSTALL_DIR}"
    echo "Please run with sudo: sudo ./uninstall.sh"
    exit 1
fi

# 删除二进制文件
if [[ -f "${BINARY_PATH}" ]]; then
    rm "${BINARY_PATH}"
    echo "aicommit has been uninstalled."
else
    echo "aicommit is not installed in ${INSTALL_DIR}"
fi

# 删除配置文件（可选）
CONFIG_DIR="${HOME}/.config/aicommit"
if [[ -d "${CONFIG_DIR}" ]]; then
    read -p "Do you want to remove configuration files? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "${CONFIG_DIR}"
        echo "Configuration files removed."
    fi
fi 