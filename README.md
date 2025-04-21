# AI Commit

[English](README_en.md) | 简体中文

一个基于AI的Git提交消息生成工具，支持多个AI提供商（Qwen、OpenAI、DeepSeek），自动生成符合Conventional Commits规范的提交消息。

## 功能特点

- 自动生成标准化的Git提交消息
- 支持多个AI提供商：
  - Qwen (通义千问)
  - OpenAI (GPT-4)
  - DeepSeek
- 遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范
- 支持多语言（英文、简体中文、繁体中文）
- 美观的命令行界面
- 交互式提交确认

## 提交消息格式

生成的提交消息严格遵循以下格式：

```
<类型>(<范围>): <主题>

<正文>

<脚注>
```

支持的提交类型：
- feat: 新功能
- fix: 修复缺陷
- refactor: 代码重构
- docs: 文档更新
- style: 代码格式
- test: 测试相关
- chore: 其他更新

## 安装

### 快速安装（推荐）

#### Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/install.sh | sudo bash
```

#### Windows (以管理员身份运行 PowerShell)
```powershell
iwr -useb https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/install.ps1 | iex
```

### 手动安装

1. 下载最新版本的发布包：
   - 访问 [Releases](https://github.com/SimonGino/aicommit/releases) 页面
   - 选择适合你系统的版本下载

2. 解压并安装：
```bash
# Linux/macOS
tar xzf aicommit_*.tar.gz
sudo mv aicommit /usr/local/bin/
chmod +x /usr/local/bin/aicommit

# Windows
# 解压zip文件，并将aicommit.exe添加到系统PATH
```

## 配置

首次使用前需要配置AI提供商的API密钥：

```bash
# 配置Qwen API
aicommit config --provider qwen --api-key your-api-key-here

# 配置OpenAI API
aicommit config --provider openai --api-key your-api-key-here

# 配置DeepSeek API
aicommit config --provider deepseek --api-key your-api-key-here
```

设置输出语言（可选）：
```bash
aicommit config --language zh-CN  # 简体中文（默认）
aicommit config --language en     # 英文
aicommit config --language zh-TW  # 繁体中文
```

## 使用方法

1. 暂存要提交的更改：
```bash
git add .  # 或指定文件
```

2. 生成提交消息：
```bash
aicommit  # 使用默认AI提供商
```

指定AI提供商：
```bash
aicommit --provider qwen     # 使用Qwen
aicommit --provider openai   # 使用OpenAI
aicommit --provider deepseek # 使用DeepSeek
```

使用自定义提交消息：
```bash
aicommit -m "feat(auth): 添加用户认证功能"
```

临时指定输出语言：
```bash
aicommit -l en     # 使用英文生成提交消息
aicommit -l zh-CN  # 使用简体中文生成提交消息
aicommit -l zh-TW  # 使用繁体中文生成提交消息
aicommit -l zh     # 使用简体中文生成提交消息（简写）
```

3. 使用 `aicommit report` 生成日报

   根据你的 Git 提交历史生成工作日报。

   ```bash
   # 生成本周日报 (默认作者为当前 Git 配置)
   aicommit report --this-week

   # 生成上周日报
   aicommit report --last-week

   # 生成指定日期范围的日报
   aicommit report --since 2023-10-01 --until 2023-10-31

   # 为指定作者生成本周日报
   aicommit report --this-week --author "user@example.com"
   ```

## 卸载

```bash
# Linux/macOS
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/uninstall.sh)"

# Windows (以管理员身份运行 PowerShell)
iwr -useb https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/uninstall.ps1 | iex
```

## 开发

1. 克隆仓库：
```bash
git clone https://github.com/SimonGino/aicommit.git
cd aicommit
```

2. 安装依赖：
```bash
go mod download
```

3. 运行测试：
```bash
go test ./...
```

## 贡献

欢迎提交Pull Request或Issue！

## 许可证

MIT 