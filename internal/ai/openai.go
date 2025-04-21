package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const openaiAPIEndpoint = "https://api.openai.com/v1/chat/completions"

// OpenAIProvider 实现 OpenAI API 的 Provider
type OpenAIProvider struct {
	*BaseProvider
}

// NewOpenAIProvider 创建 OpenAIProvider 实例
func NewOpenAIProvider(base *BaseProvider) *OpenAIProvider {
	return &OpenAIProvider{BaseProvider: base}
}

type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *OpenAIProvider) GenerateCommitMessage(ctx context.Context, info *CommitInfo) (*CommitMessage, error) {
	reqBody := openaiRequest{
		Model: openai.GPT4oMini, // 或者选择其他合适的模型
		Messages: []openaiMessage{
			{Role: "system", Content: p.GetSystemPrompt()},
			{Role: "user", Content: p.GetUserPrompt(info, p.BuildFilesList(info.FilesChanged))},
		},
		Temperature: 0.7,
		MaxTokens:   1500,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openaiAPIEndpoint, strings.NewReader(string(reqData)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d", resp.StatusCode)
	}

	var result openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("API返回结果为空")
	}

	// 清理响应内容中的Markdown格式标记
	content := result.Choices[0].Message.Content
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

// GenerateDailyReport 使用OpenAI生成日报
func (p *OpenAIProvider) GenerateDailyReport(ctx context.Context, info *ReportInfo, since, until string) (string, error) {
	client := openai.NewClient(p.APIKey)

	userPrompt := p.GetUserPromptForReport(info, since, until)
	// OpenAI 不需要特定的系统提示来生成日报，依赖用户提示中的指令

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini, // 或者选择其他合适的模型
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("请求 OpenAI API 失败: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("OpenAI 未返回有效的日报内容")
	}

	reportContent := p.CleanMarkdownFormatting(resp.Choices[0].Message.Content)

	return reportContent, nil
}
