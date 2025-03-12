package ai

import (
	"context"
	"fmt"
	"strings"
)

// CommitInfo 包含生成提交消息所需的信息
type CommitInfo struct {
	FilesChanged []string
	DiffContent  string
	BranchName   string
}

// CommitMessage 表示生成的提交消息
type CommitMessage struct {
	Title string
	Body  string
}

// CommitType 定义提交类型
type CommitType struct {
	Type        string
	Description string
}

// 提交类型定义
var commitTypes = map[string][]CommitType{
	"en": {
		{"feat", "New feature"},
		{"fix", "Bug fix"},
		{"refactor", "Code refactoring"},
		{"docs", "Documentation changes"},
		{"style", "Code style changes (formatting, missing semicolons, etc)"},
		{"test", "Adding or modifying tests"},
		{"chore", "Maintenance tasks, dependencies, build changes"},
	},
	"zh-CN": {
		{"feat", "新功能"},
		{"fix", "修复缺陷"},
		{"refactor", "代码重构"},
		{"docs", "文档更新"},
		{"style", "代码格式"},
		{"test", "测试相关"},
		{"chore", "其他更新"},
	},
	"zh-TW": {
		{"feat", "新功能"},
		{"fix", "修復缺陷"},
		{"refactor", "代碼重構"},
		{"docs", "文檔更新"},
		{"style", "代碼格式"},
		{"test", "測試相關"},
		{"chore", "其他更新"},
	},
}

// Provider 定义了AI提供商的接口
type Provider interface {
	GenerateCommitMessage(ctx context.Context, info *CommitInfo) (*CommitMessage, error)
}

// BaseProvider 提供基本的AI提供商实现
type BaseProvider struct {
	APIKey   string
	Language string
}

// BuildFilesList 构建文件列表字符串
func (p *BaseProvider) BuildFilesList(files []string) string {
	var filesList strings.Builder
	for _, file := range files {
		filesList.WriteString("- ")
		filesList.WriteString(file)
		filesList.WriteString("\n")
	}
	return filesList.String()
}

// CleanMarkdownFormatting 清理Markdown格式标记
func (p *BaseProvider) CleanMarkdownFormatting(content string) string {
	// 移除 ```plaintext 和 ``` 标记
	content = strings.TrimPrefix(content, "```plaintext")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")

	// 移除开头的空行
	content = strings.TrimLeft(content, "\n")

	// 移除随机添加的issue引用（如果不是用户明确要求的）
	// 匹配中文的"修复 #数字"或英文的"Fixes #数字"等格式
	lines := strings.Split(content, "\n")
	filteredLines := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		// 跳过匹配"修复 #数字"、"Fixes #数字"等格式的行
		if strings.HasPrefix(trimmedLine, "修复 #") ||
			strings.HasPrefix(trimmedLine, "Fixes #") ||
			strings.HasPrefix(trimmedLine, "fixes #") ||
			strings.HasPrefix(trimmedLine, "Fix #") ||
			strings.HasPrefix(trimmedLine, "fix #") ||
			strings.HasPrefix(trimmedLine, "Closes #") ||
			strings.HasPrefix(trimmedLine, "closes #") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}

	return strings.Join(filteredLines, "\n")
}

// GetUserPrompt 根据语言返回用户提示
func (p *BaseProvider) GetUserPrompt(info *CommitInfo, filesList string) string {
	switch p.Language {
	case "zh-CN":
		return fmt.Sprintf(`请为以下Git更改生成标准化的提交信息：

分支：%s

更改的文件：
%s
更改内容：
%s

请严格按照系统提示中的格式要求生成提交信息。`,
			info.BranchName,
			filesList,
			info.DiffContent)
	case "zh-TW":
		return fmt.Sprintf(`請為以下Git更改生成標準化的提交信息：

分支：%s

更改的文件：
%s
更改內容：
%s

請嚴格按照系統提示中的格式要求生成提交信息。`,
			info.BranchName,
			filesList,
			info.DiffContent)
	default:
		return fmt.Sprintf(`Please generate a standardized commit message for the following Git changes:

Branch: %s

Files changed:
%s
Changes:
%s

Please strictly follow the format requirements in the system prompt.`,
			info.BranchName,
			filesList,
			info.DiffContent)
	}
}

// NewProvider 创建指定类型的AI提供商实例
func NewProvider(providerType, apiKey, language string) (Provider, error) {
	base := &BaseProvider{
		APIKey:   apiKey,
		Language: language,
	}

	switch providerType {
	case "qwen":
		return NewQwenProvider(base), nil
	case "openai":
		return NewOpenAIProvider(base), nil
	case "deepseek":
		return NewDeepseekProvider(base), nil
	default:
		return nil, fmt.Errorf("不支持的AI提供商: %s", providerType)
	}
}

// GetCommitTypes 返回指定语言的提交类型
func (p *BaseProvider) GetCommitTypes() []CommitType {
	types, ok := commitTypes[p.Language]
	if !ok {
		return commitTypes["en"]
	}
	return types
}

// GetSystemPrompt 根据语言返回系统提示
func (p *BaseProvider) GetSystemPrompt() string {
	// 获取提交类型列表
	types := p.GetCommitTypes()

	// 构建类型说明
	var typeDesc string
	for _, t := range types {
		typeDesc += fmt.Sprintf("- %s: %s\n", t.Type, t.Description)
	}

	switch p.Language {
	case "zh-CN":
		return fmt.Sprintf(`您是一个帮助生成标准化git提交信息的助手。
请严格遵循以下提交信息格式规则：

1. 格式：<类型>(<范围>): <主题>

<正文>

<脚注>

2. 类型必须是以下之一：
%s
3. 范围：可选，描述影响的区域（如：router、auth、db）
4. 主题：简短摘要（不超过50个字符）
5. 正文：详细说明（每行不超过72个字符）
6. 脚注：可选，用于说明重大变更或引用问题编号

示例：
feat(认证): 实现JWT认证系统

添加基于JWT的认证系统，支持刷新令牌
- 实现令牌生成和验证
- 添加用户会话管理
- 设置安全Cookie处理

重大变更：需要新的认证头
修复 #123`, typeDesc)

	case "zh-TW":
		return fmt.Sprintf(`您是一個幫助生成標準化git提交信息的助手。
請嚴格遵循以下提交信息格式規則：

1. 格式：<類型>(<範圍>): <主題>

<正文>

<腳註>

2. 類型必須是以下之一：
%s
3. 範圍：可選，描述影響的區域（如：router、auth、db）
4. 主題：簡短摘要（不超過50個字符）
5. 正文：詳細說明（每行不超過72個字符）
6. 腳註：可選，用於說明重大變更或引用問題編號

示例：
feat(認證): 實現JWT認證系統

添加基於JWT的認證系統，支持刷新令牌
- 實現令牌生成和驗證
- 添加用戶會話管理
- 設置安全Cookie處理

重大變更：需要新的認證頭
修復 #123`, typeDesc)

	default:
		return fmt.Sprintf(`You are a helpful assistant that generates standardized git commit messages.
Follow these strict rules for commit message format:

1. Format: <type>(<scope>): <subject>

<body>

<footer>

2. Types must be one of:
%s
3. Scope: Optional, describes the affected area (e.g., router, auth, db)
4. Subject: Short summary (50 chars or less)
5. Body: Detailed explanation (72 chars per line)
6. Footer: Optional, for breaking changes or issue references

Example:
feat(auth): implement JWT authentication

Add JWT-based authentication system with refresh tokens
- Implement token generation and validation
- Add user session management
- Set up secure cookie handling

BREAKING CHANGE: New authentication headers required
Fixes #123`, typeDesc)
	}
}
