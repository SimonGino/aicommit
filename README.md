# AI Commit

[English](README_en.md) | 简体中文

一个基于AI的Git提交消息生成工具，自动生成符合Conventional Commits规范的提交消息。

## 功能特点

- 🤖 **AI驱动** - 自动分析代码变更，生成标准化提交消息
- 🎯 **交互式操作** - 支持键盘快捷键，快速选择操作
- 📁 **灵活的文件选择** - 可选择暂存区、手动选择文件或暂存全部
- ✏️ **消息编辑** - 支持编辑生成的消息或重新生成
- 🔧 **配置检测** - 内置 `check` 命令验证配置和API连通性
- 🌍 **多语言支持** - 英文、简体中文、繁体中文
- ☁️ **多平台** - 支持 OpenAI 和 Azure OpenAI
- 📊 **日报生成** - 根据Git提交历史生成工作日报

## 快速开始

### 安装

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/install.sh | sudo bash

# Windows (以管理员身份运行 PowerShell)
iwr -useb https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/install.ps1 | iex
```

### 配置

```bash
# 配置 OpenAI API 密钥
aicommit config --api-key your-openai-api-key

# 检查配置是否正确
aicommit check
```

### 使用

```bash
# 交互式提交（推荐）
aicommit

# 使用自定义消息
aicommit -m "feat: 添加新功能"
```

## 交互式流程

运行 `aicommit` 后，会显示交互式界面：

```
检测到以下变更:

已暂存 (Staged):
  ✓ src/main.go

未暂存 (Modified):
  • config.json

请选择操作:
  [a] 使用当前暂存区内容生成提交消息
  [s] 选择要暂存的文件
  [A] 暂存所有变更 (git add .)
  [c] 取消

请按键选择: a

正在生成提交消息...

✔ 生成的提交消息：
┌────────────────────────────────────────────────────────────┐
│ feat(main): 添加用户认证功能                                │
│                                                            │
│ - 实现 JWT 令牌验证                                        │
│ - 添加用户登录接口                                         │
└────────────────────────────────────────────────────────────┘

请选择操作:
  [a] 接受并提交
  [e] 编辑后提交
  [r] 重新生成
  [c] 取消

请按键选择: a

✓ 已提交更改
```

## 命令

| 命令 | 说明 |
|------|------|
| `aicommit` | 交互式生成并提交 |
| `aicommit -m "msg"` | 使用指定消息提交 |
| `aicommit check` | 检查配置和API连通性 |
| `aicommit config` | 配置设置 |
| `aicommit report` | 生成日报 |

## 配置

### OpenAI

```bash
aicommit config --provider openai
aicommit config --api-key sk-your-api-key
aicommit config --model gpt-4o  # 可选
```

### Azure OpenAI

```bash
aicommit config --provider azure
aicommit config --api-key your-azure-key
aicommit config --base-url "https://your-resource.openai.azure.com/openai/deployments/your-deployment/chat/completions"
aicommit config --azure-api-version "2024-02-15-preview"
```

### 语言设置

```bash
aicommit config --language zh-CN  # 简体中文（默认）
aicommit config --language en     # 英文
aicommit config --language zh-TW  # 繁体中文
```

## 日报生成

```bash
# 本周日报
aicommit report --this-week

# 上周日报
aicommit report --last-week

# 指定日期范围
aicommit report --since 2024-01-01 --until 2024-01-31
```

## 提交消息格式

遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<类型>(<范围>): <主题>

<正文>
```

支持的类型：`feat` | `fix` | `refactor` | `docs` | `style` | `test` | `chore`

## 开发

```bash
git clone https://github.com/SimonGino/aicommit.git
cd aicommit
go mod download
go test ./...
go build -o aicommit ./cmd/aicommit
```

## 卸载

```bash
# Linux/macOS
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/uninstall.sh)"

# Windows (以管理员身份运行 PowerShell)
iwr -useb https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/uninstall.ps1 | iex
```

## 许可证

MIT