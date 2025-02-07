#!/bin/bash

# 检测操作系统
OS="$(uname -s)"
case "${OS}" in
    Linux*)     BINARY_NAME=aicommit-linux;;
    Darwin*)    BINARY_NAME=aicommit-macos;;
    *)          echo "Unsupported operating system: ${OS}"; exit 1;;
esac

# 检测CPU架构
ARCH="$(uname -m)"
if [[ "${ARCH}" != "x86_64" ]] && [[ "${ARCH}" != "arm64" ]]; then
    echo "Unsupported architecture: ${ARCH}"
    exit 1
fi

# 设置安装目录
INSTALL_DIR="/usr/local/bin"
if [[ ! -w "${INSTALL_DIR}" ]]; then
    echo "Error: No write permission to ${INSTALL_DIR}"
    echo "Please run with sudo: sudo ./install.sh"
    exit 1
fi

# 获取最新版本
echo "Fetching latest release..."
LATEST_RELEASE_URL=$(curl -s https://api.github.com/repos/SimonGino/aicommit/releases/latest | grep "browser_download_url.*${BINARY_NAME}" | cut -d '"' -f 4)

if [[ -z "${LATEST_RELEASE_URL}" ]]; then
    echo "Error: Could not find latest release"
    exit 1
fi

# 下载二进制文件
echo "Downloading ${BINARY_NAME}..."
curl -L "${LATEST_RELEASE_URL}" -o "${INSTALL_DIR}/aicommit"

# 设置执行权限
chmod +x "${INSTALL_DIR}/aicommit"

# 添加zsh自动补全支持
if [[ -f "${HOME}/.zshrc" ]]; then
    # 检查是否已经添加过自动补全配置
    if ! grep -q "# aicommit completion" "${HOME}/.zshrc"; then
        echo -e "\n# aicommit completion\neval \"$(aicommit --completion zsh)\"" >> "${HOME}/.zshrc"
        echo "Added zsh completion support. Please restart your shell or run: source ${HOME}/.zshrc"
    fi
fi

echo "Installation completed! You can now use 'aicommit' command."
echo "Try 'aicommit --help' to get started."