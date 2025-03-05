package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const deepseekAPIEndpoint = "https://api.deepseek.com/v1/chat/completions"

type DeepseekProvider struct {
	*BaseProvider
}

func NewDeepseekProvider(base *BaseProvider) *DeepseekProvider {
	return &DeepseekProvider{
		BaseProvider: base,
	}
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

	// 分割标题和正文
	parts := strings.SplitN(result.Choices[0].Message.Content, "\n\n", 2)
	message := &CommitMessage{
		Title: strings.TrimSpace(parts[0]),
	}
	if len(parts) > 1 {
		message.Body = strings.TrimSpace(parts[1])
	}

	return message, nil
}
