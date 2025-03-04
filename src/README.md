# AI Commit

一个AI驱动的git commit消息生成器，支持多个AI提供商（Qwen、OpenAI、DeepSeek）。

## 功能特点

- 基于暂存的更改自动生成有意义的提交消息
- 支持多个AI提供商（目前支持Qwen、OpenAI、DeepSeek）
- 简单的配置和API密钥管理
- 美观的命令行界面
- 支持多语言输出（英文、简体中文、繁体中文）

## 安装

```bash
go install github.com/SimonGino/aicommit/cmd/aicommit@latest
```

## 配置

在使用工具之前，需要配置AI提供商的API密钥：

```bash
# 配置Qwen API
aicommit config --provider qwen --api-key your-api-key-here

# 配置OpenAI API
aicommit config --provider openai --api-key your-api-key-here

# 配置DeepSeek API
aicommit config --provider deepseek --api-key your-api-key-here
```

## 使用方法

1. 暂存你的更改：
```bash
git add .  # 或指定文件
```

2. 生成并提交带有AI生成的消息：
```bash
aicommit
```

你也可以指定使用特定的提供商：
```bash
aicommit --provider qwen    # 使用Qwen API
aicommit --provider openai  # 使用OpenAI API
aicommit --provider deepseek # 使用DeepSeek API
```

或者使用手动指定的消息（跳过AI）：
```bash
aicommit -m "你的消息"
```

## 语言设置

你可以设置提交消息的输出语言：
```bash
aicommit config --language en      # 英文
aicommit config --language zh-CN   # 简体中文
aicommit config --language zh-TW   # 繁体中文
```

## 许可证

MIT 