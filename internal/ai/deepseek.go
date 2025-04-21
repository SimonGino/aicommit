package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const deepseekAPIEndpoint = "https://api.deepseek.com/v1/chat/completions"

// DeepseekProvider 实现 Deepseek API 的 Provider
type DeepseekProvider struct {
	*BaseProvider
}

// NewDeepseekProvider 创建 DeepseekProvider 实例
func NewDeepseekProvider(base *BaseProvider) *DeepseekProvider {
	return &DeepseekProvider{BaseProvider: base}
}

type deepseekRequest struct {
	Model       string            `json:"model"`
	Messages    []deepseekMessage `json:"messages"`
	Temperature float64           `json:"temperature"`
	MaxTokens   int               `json:"max_tokens"`
}

type deepseekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepseekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *DeepseekProvider) GenerateCommitMessage(ctx context.Context, info *CommitInfo) (*CommitMessage, error) {
	reqBody := deepseekRequest{
		Model: "deepseek-chat",
		Messages: []deepseekMessage{
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

	req, err := http.NewRequestWithContext(ctx, "POST", deepseekAPIEndpoint, strings.NewReader(string(reqData)))
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

	var result deepseekResponse
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

// GenerateDailyReport 使用 Deepseek 生成日报
func (p *DeepseekProvider) GenerateDailyReport(ctx context.Context, info *ReportInfo, since, until string) (string, error) {
	client := &http.Client{}
	url := "https://api.deepseek.com/chat/completions"

	userPrompt := p.GetUserPromptForReport(info, since, until)
	// Deepseek 同样不需要特定系统提示生成日报，依赖用户提示

	payload := map[string]interface{}{
		"model": "deepseek-chat", // 或其他模型
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
		"stream": false, // 确保不使用流式输出
	}
	payloadBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 Deepseek API 失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 Deepseek 响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Deepseek API 返回错误状态 %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("解析 Deepseek 响应 JSON 失败: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("Deepseek 响应格式错误: 缺少 choices 字段或为空")
	}
	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Deepseek 响应格式错误: choices 元素格式错误")
	}
	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Deepseek 响应格式错误: 缺少 message 字段")
	}
	content, ok := message["content"].(string)
	if !ok || content == "" {
		return "", fmt.Errorf("Deepseek 未返回有效的日报内容")
	}

	reportContent := p.CleanMarkdownFormatting(content)

	return reportContent, nil
}
