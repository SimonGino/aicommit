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

const qwenAPIEndpoint = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"

// QwenProvider 实现通义千问 API 的 Provider
type QwenProvider struct {
	*BaseProvider
}

// NewQwenProvider 创建 QwenProvider 实例
func NewQwenProvider(base *BaseProvider) *QwenProvider {
	return &QwenProvider{BaseProvider: base}
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

	// 清理响应内容中的Markdown格式标记
	content := result.Output.Text
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

// GenerateDailyReport 使用通义千问生成日报
func (p *QwenProvider) GenerateDailyReport(ctx context.Context, info *ReportInfo, since, until string) (string, error) {
	client := &http.Client{}
	url := "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"

	userPrompt := p.GetUserPromptForReport(info, since, until)
	// 通义千问同样不需要特定系统提示生成日报，依赖用户提示

	payload := map[string]interface{}{
		"model": "qwen-turbo", // 或其他版本
		"input": map[string]interface{}{
			"messages": []map[string]string{
				{"role": "user", "content": userPrompt},
			},
		},
		"parameters": map[string]interface{}{
			"result_format": "message",
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求通义千问 API 失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取通义千问响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("通义千问 API 返回错误状态 %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("解析通义千问响应 JSON 失败: %w", err)
	}

	output, ok := result["output"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("通义千问响应格式错误: 缺少 output 字段")
	}
	choices, ok := output["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("通义千问响应格式错误: 缺少 choices 字段或为空")
	}
	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("通义千问响应格式错误: choices 元素格式错误")
	}
	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("通义千问响应格式错误: 缺少 message 字段")
	}
	content, ok := message["content"].(string)
	if !ok || content == "" {
		return "", fmt.Errorf("通义千问未返回有效的日报内容")
	}

	reportContent := p.CleanMarkdownFormatting(content)

	return reportContent, nil
}
