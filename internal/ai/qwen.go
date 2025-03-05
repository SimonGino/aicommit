package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const qwenAPIEndpoint = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"

type QwenProvider struct {
	*BaseProvider
}

func NewQwenProvider(base *BaseProvider) *QwenProvider {
	return &QwenProvider{
		BaseProvider: base,
	}
}

type qwenRequest struct {
	Model      string         `json:"model"`
	Input      qwenInput      `json:"input"`
	Parameters qwenParameters `json:"parameters"`
}

type qwenInput struct {
	Messages []qwenMessage `json:"messages"`
}

type qwenMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type qwenParameters struct {
	ResultFormat string  `json:"result_format"`
	Temperature  float64 `json:"temperature"`
	TopP         float64 `json:"top_p"`
	TopK         int     `json:"top_k"`
	MaxTokens    int     `json:"max_tokens"`
}

type qwenResponse struct {
	Output struct {
		Text string `json:"text"`
	} `json:"output"`
}

func (p *QwenProvider) GenerateCommitMessage(ctx context.Context, info *CommitInfo) (*CommitMessage, error) {
	reqBody := qwenRequest{
		Model: "qwen-max",
		Input: qwenInput{
			Messages: []qwenMessage{
				{Role: "system", Content: p.GetSystemPrompt()},
				{Role: "user", Content: p.GetUserPrompt(info, p.BuildFilesList(info.FilesChanged))},
			},
		},
		Parameters: qwenParameters{
			ResultFormat: "text",
			Temperature:  0.7,
			TopP:         0.8,
			TopK:         50,
			MaxTokens:    1500,
		},
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", qwenAPIEndpoint, strings.NewReader(string(reqData)))
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

	var result qwenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 分割标题和正文
	parts := strings.SplitN(result.Output.Text, "\n\n", 2)
	message := &CommitMessage{
		Title: strings.TrimSpace(parts[0]),
	}
	if len(parts) > 1 {
		message.Body = strings.TrimSpace(parts[1])
	}

	return message, nil
}
