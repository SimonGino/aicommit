package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const openaiAPIEndpoint = "https://api.openai.com/v1/chat/completions"

type OpenAIProvider struct {
	*BaseProvider
}

func NewOpenAIProvider(base *BaseProvider) *OpenAIProvider {
	return &OpenAIProvider{
		BaseProvider: base,
	}
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
	// 构建文件列表
	var filesList strings.Builder
	for _, file := range info.FilesChanged {
		filesList.WriteString("- ")
		filesList.WriteString(file)
		filesList.WriteString("\n")
	}

	prompt := fmt.Sprintf(`Please generate a standardized commit message for the following Git changes:

Branch: %s

Files changed:
%s
Changes:
%s

Please strictly follow the format requirements in the system prompt.`,
		info.BranchName,
		filesList.String(),
		info.DiffContent)

	reqBody := openaiRequest{
		Model: "gpt-4",
		Messages: []openaiMessage{
			{Role: "system", Content: p.GetSystemPrompt()},
			{Role: "user", Content: prompt},
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
