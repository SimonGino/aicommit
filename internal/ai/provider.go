package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// azureTransport 用于添加Azure OpenAI特定的认证头
type azureTransport struct {
	transport http.RoundTripper
	apiKey    string
}

func (t *azureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Azure OpenAI 使用 api-key 头
	req.Header.Set("api-key", t.apiKey)
	return t.transport.RoundTrip(req)
}

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

// ReportInfo 包含生成日报所需的信息
type ReportInfo struct {
	Commits []string
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
	GenerateDailyReport(ctx context.Context, info *ReportInfo, since, until string) (string, error)
}

// OpenAIProvider 实现使用 sashabaranov/go-openai 库的 Provider
type OpenAIProvider struct {
	apiKey   string
	baseURL  string
	model    string
	language string
	client   *openai.Client
}

// NewProvider 创建统一的 Provider 实例，支持 OpenAI 和 Azure OpenAI
func NewProvider(apiKey, baseURL, model, language, provider, azureAPIVersion string) (Provider, error) {
	var config openai.ClientConfig
	var effectiveBaseURL string

	// 根据 provider 类型决定使用哪个配置
	switch provider {
	case "azure":
		if apiKey == "" {
			return nil, fmt.Errorf("Azure OpenAI API 密钥不能为空")
		}
		if baseURL == "" {
			return nil, fmt.Errorf("Azure OpenAI endpoint URL 不能为空")
		}
		if model == "" {
			return nil, fmt.Errorf("Azure OpenAI 模型/部署名称不能为空")
		}

		// 如果没有指定 API 版本，使用默认版本
		if azureAPIVersion == "" {
			azureAPIVersion = "2024-02-15-preview"
		}

		// Azure OpenAI 配置
		azureBaseURL := strings.TrimRight(baseURL, "/")
		config = openai.DefaultAzureConfig(apiKey, azureBaseURL)
		config.APIVersion = azureAPIVersion

		// 创建自定义HTTP客户端以添加正确的认证头
		config.HTTPClient = &http.Client{
			Transport: &azureTransport{
				transport: http.DefaultTransport,
				apiKey:    apiKey,
			},
		}

		effectiveBaseURL = azureBaseURL

	default:
		// 默认使用 OpenAI
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API 密钥不能为空")
		}
		config = openai.DefaultConfig(apiKey)
		effectiveBaseURL = baseURL

		// 设置自定义 URL (仅对 OpenAI)
		if baseURL != "" {
			config.BaseURL = baseURL
		}
	}

	// 检查是否需要设置代理
	if proxyURL := getEnv("HTTP_PROXY", ""); proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
			config.HTTPClient = &http.Client{
				Transport: transport,
			}
		}
	}

	// 如果没有指定模型，使用默认模型
	if model == "" {
		if provider == "azure" {
			// Azure OpenAI 通常使用部署名称作为模型名
			model = "gpt-4o"
		} else {
			model = openai.GPT4o
		}
	}

	providerInstance := &OpenAIProvider{
		apiKey:   apiKey,
		baseURL:  effectiveBaseURL,
		model:    model,
		language: language,
		client:   openai.NewClientWithConfig(config),
	}

	return providerInstance, nil
}

// 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// BuildFilesList 构建文件列表字符串
func (p *OpenAIProvider) BuildFilesList(files []string) string {
	var filesList strings.Builder
	for _, file := range files {
		filesList.WriteString("- ")
		filesList.WriteString(file)
		filesList.WriteString("\n")
	}
	return filesList.String()
}

// CleanMarkdownFormatting 清理Markdown格式标记
func (p *OpenAIProvider) CleanMarkdownFormatting(content string) string {
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
func (p *OpenAIProvider) GetUserPrompt(info *CommitInfo, filesList string) string {
	switch p.language {
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

// GetUserPromptForReport 根据语言返回生成日报的用户提示
func (p *OpenAIProvider) GetUserPromptForReport(info *ReportInfo, since, until string) string {
	// 将提交列表格式化为 "- YYYY-MM-DD -- Subject"
	var commitsFormatted strings.Builder
	for _, commit := range info.Commits {
		commitsFormatted.WriteString("- ")
		commitsFormatted.WriteString(commit)
		commitsFormatted.WriteString("\n")
	}
	commitsList := strings.TrimSpace(commitsFormatted.String())

	switch p.language {
	case "zh-CN":
		return fmt.Sprintf(`请根据以下 Git commit 记录（格式为 "- YYYY-MM-DD -- Commit Subject"），为日期范围 %s 至 %s 总结生成一份简洁的工作日报。

要求：
1.  使用 Markdown 格式。
2.  按日期**总结**当天完成的主要工作，**不要**罗列单个 commit message。
3.  忽略所有 "Merge branch" 或 "Merge remote-tracking branch" 相关的提交。
4.  报告标题或开头应明确指出报告的时间范围是 %s 到 %s。
5.  语言为简体中文。

Commit 记录:
%s

请生成日报内容：`, since, until, since, until, commitsList)
	case "zh-TW":
		return fmt.Sprintf(`請根據以下 Git commit 記錄（格式為 "- YYYY-MM-DD -- Commit Subject"），為日期範圍 %s 至 %s 總結生成一份簡潔的工作日報。

要求：
1.  使用 Markdown 格式。
2.  按日期**總結**當天完成的主要工作，**不要**羅列單個 commit message。
3.  忽略所有 "Merge branch" 或 "Merge remote-tracking branch" 相關的提交。
4.  報告標題或開頭應明確指出報告的時間範圍是 %s 到 %s。
5.  語言為繁體中文。

Commit 記錄:
%s

請生成日報內容：`, since, until, since, until, commitsList)
	default:
		return fmt.Sprintf(`Please summarize the following Git commit records (formatted as "- YYYY-MM-DD -- Commit Subject") into a concise work report for the period %s to %s.

Requirements:
1.  Use Markdown format.
2.  Summarize the main work completed **per day**. **Do not** list individual commit messages.
3.  Ignore any commits related to "Merge branch" or "Merge remote-tracking branch".
4.  The report title or beginning should clearly state the reporting period is from %s to %s.
5.  The language should be English.

Commit Records:
%s

Please generate the report content:`, since, until, since, until, commitsList)
	}
}

// GetCommitTypes 返回指定语言的提交类型
func (p *OpenAIProvider) GetCommitTypes() []CommitType {
	types, ok := commitTypes[p.language]
	if !ok {
		return commitTypes["en"]
	}
	return types
}

// GetSystemPrompt 根据语言返回系统提示
func (p *OpenAIProvider) GetSystemPrompt() string {
	// 获取提交类型列表
	types := p.GetCommitTypes()

	// 构建类型说明
	var typeDesc string
	for _, t := range types {
		typeDesc += fmt.Sprintf("- %s: %s\n", t.Type, t.Description)
	}

	switch p.language {
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

// GenerateCommitMessage 使用 OpenAI API 生成提交消息
func (p *OpenAIProvider) GenerateCommitMessage(ctx context.Context, info *CommitInfo) (*CommitMessage, error) {
	// 准备用户提示
	userPrompt := p.GetUserPrompt(info, p.BuildFilesList(info.FilesChanged))
	systemPrompt := p.GetSystemPrompt()

	// 创建聊天请求
	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: p.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			Temperature: 0.7,
			MaxTokens:   1500,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("请求 OpenAI API 失败: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return nil, fmt.Errorf("OpenAI 未返回有效的提交信息内容")
	}

	// 清理响应内容中的Markdown格式标记
	content := resp.Choices[0].Message.Content
	content = p.CleanMarkdownFormatting(content)

	// 分割标题和正文
	parts := strings.SplitN(content, "\n\n", 2)
	message := &CommitMessage{
		Title: strings.TrimSpace(parts[0]),
	}
	if len(parts) > 1 {
		message.Body = strings.TrimSpace(parts[1])
	}

	return message, nil
}

// GenerateDailyReport 使用 OpenAI API 生成日报
func (p *OpenAIProvider) GenerateDailyReport(ctx context.Context, info *ReportInfo, since, until string) (string, error) {
	// 准备用户提示
	userPrompt := p.GetUserPromptForReport(info, since, until)

	// 创建聊天请求
	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: p.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
			Temperature: 0.7,
			MaxTokens:   2000,
		},
	)

	if err != nil {
		return "", fmt.Errorf("请求 OpenAI API 失败: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("OpenAI 未返回有效的日报内容")
	}

	reportContent := p.CleanMarkdownFormatting(resp.Choices[0].Message.Content)

	return reportContent, nil
}
