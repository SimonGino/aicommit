package ai

import (
	"strings"
	"testing"
)

// 创建测试用的 OpenAIProvider
func newTestProvider(language string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:   "test-key",
		baseURL:  "https://api.openai.com/v1",
		model:    "gpt-4o",
		language: language,
	}
}

func TestBuildFilesList(t *testing.T) {
	p := newTestProvider("en")

	testCases := []struct {
		name     string
		files    []string
		expected string
	}{
		{
			name:     "空列表",
			files:    []string{},
			expected: "",
		},
		{
			name:     "单个文件",
			files:    []string{"main.go"},
			expected: "- main.go\n",
		},
		{
			name:     "多个文件",
			files:    []string{"main.go", "config.go", "utils.go"},
			expected: "- main.go\n- config.go\n- utils.go\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := p.BuildFilesList(tc.files)
			if result != tc.expected {
				t.Errorf("期望:\n%s\n实际:\n%s", tc.expected, result)
			}
		})
	}
}

func TestCleanMarkdownFormatting(t *testing.T) {
	p := newTestProvider("en")

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "无 Markdown 标记",
			input:    "feat(auth): add login",
			expected: "feat(auth): add login",
		},
		{
			name:     "带 plaintext 标记",
			input:    "```plaintext\nfeat(auth): add login\n```",
			expected: "feat(auth): add login\n",
		},
		{
			name:     "带普通代码块标记",
			input:    "```\nfeat(auth): add login\n```",
			expected: "feat(auth): add login\n",
		},
		{
			name:     "过滤 Fixes 引用",
			input:    "feat(auth): add login\n\nFixes #123",
			expected: "feat(auth): add login\n",
		},
		{
			name:     "过滤中文修复引用",
			input:    "feat(auth): add login\n\n修复 #123",
			expected: "feat(auth): add login\n",
		},
		{
			name:     "过滤 Closes 引用",
			input:    "feat(auth): add login\n\nCloses #456",
			expected: "feat(auth): add login\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := p.CleanMarkdownFormatting(tc.input)
			if result != tc.expected {
				t.Errorf("期望:\n'%s'\n实际:\n'%s'", tc.expected, result)
			}
		})
	}
}

func TestGetCommitTypes(t *testing.T) {
	testCases := []struct {
		language     string
		expectedLen  int
		expectedType string // 第一个类型
	}{
		{"en", 7, "feat"},
		{"zh-CN", 7, "feat"},
		{"zh-TW", 7, "feat"},
		{"unknown", 7, "feat"}, // 应该回退到英文
	}

	for _, tc := range testCases {
		t.Run(tc.language, func(t *testing.T) {
			p := newTestProvider(tc.language)
			types := p.GetCommitTypes()

			if len(types) != tc.expectedLen {
				t.Errorf("期望 %d 个类型, 实际 %d 个", tc.expectedLen, len(types))
			}
			if len(types) > 0 && types[0].Type != tc.expectedType {
				t.Errorf("期望第一个类型='%s', 实际='%s'", tc.expectedType, types[0].Type)
			}
		})
	}
}

func TestGetSystemPrompt(t *testing.T) {
	testCases := []struct {
		language    string
		shouldMatch string
	}{
		{"en", "You are a helpful assistant"},
		{"zh-CN", "您是一个帮助生成标准化git提交信息的助手"},
		{"zh-TW", "您是一個幫助生成標準化git提交信息的助手"},
	}

	for _, tc := range testCases {
		t.Run(tc.language, func(t *testing.T) {
			p := newTestProvider(tc.language)
			prompt := p.GetSystemPrompt()

			if !strings.Contains(prompt, tc.shouldMatch) {
				t.Errorf("系统提示应包含 '%s'", tc.shouldMatch)
			}
		})
	}
}

func TestGetUserPrompt(t *testing.T) {
	info := &CommitInfo{
		FilesChanged: []string{"main.go", "config.go"},
		DiffContent:  "+ new line\n- old line",
		BranchName:   "feature/test",
	}

	testCases := []struct {
		language    string
		shouldMatch string
	}{
		{"en", "Branch: feature/test"},
		{"zh-CN", "分支：feature/test"},
		{"zh-TW", "分支：feature/test"},
	}

	for _, tc := range testCases {
		t.Run(tc.language, func(t *testing.T) {
			p := newTestProvider(tc.language)
			filesList := p.BuildFilesList(info.FilesChanged)
			prompt := p.GetUserPrompt(info, filesList)

			if !strings.Contains(prompt, tc.shouldMatch) {
				t.Errorf("用户提示应包含 '%s'", tc.shouldMatch)
			}
			if !strings.Contains(prompt, "main.go") {
				t.Error("用户提示应包含文件名")
			}
		})
	}
}

// ========================
// TruncateDiff 测试
// ========================

func TestTruncateDiff_Empty(t *testing.T) {
	p := newTestProvider("en")

	result := p.TruncateDiff("", 1000)
	if result != "" {
		t.Errorf("空 diff 应返回空字符串, 实际='%s'", result)
	}
}

func TestTruncateDiff_SmallDiff(t *testing.T) {
	p := newTestProvider("en")

	smallDiff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"
 func main() {
 }`

	result := p.TruncateDiff(smallDiff, 10000)
	if result != smallDiff {
		t.Error("小于限制的 diff 应原样返回")
	}
}

func TestTruncateDiff_LargeDiff(t *testing.T) {
	p := newTestProvider("en")

	// 生成一个较大的 diff（必须远大于 maxLength）
	var largeDiff strings.Builder
	largeDiff.WriteString("diff --git a/file1.go b/file1.go\n")
	largeDiff.WriteString("--- a/file1.go\n")
	largeDiff.WriteString("+++ b/file1.go\n")
	largeDiff.WriteString("@@ -1,1000 +1,1000 @@\n")
	for i := 0; i < 500; i++ {
		largeDiff.WriteString("+    this is a very long line of code that adds more content " + string(rune('a'+i%26)) + "\n")
	}

	largeDiff.WriteString("diff --git a/file2.go b/file2.go\n")
	largeDiff.WriteString("--- a/file2.go\n")
	largeDiff.WriteString("+++ b/file2.go\n")
	largeDiff.WriteString("@@ -1,1000 +1,1000 @@\n")
	for i := 0; i < 500; i++ {
		largeDiff.WriteString("-    this is a very long line of old code that will be removed " + string(rune('a'+i%26)) + "\n")
	}

	originalLength := len(largeDiff.String())
	maxLength := 2000
	result := p.TruncateDiff(largeDiff.String(), maxLength)

	// 验证结果被截断
	if len(result) >= originalLength {
		t.Errorf("大 diff 应该被截断, 原始长度=%d, 结果长度=%d", originalLength, len(result))
	}

	// 验证包含截断提示 (检查多种可能的格式)
	hasTruncateNotice := strings.Contains(result, "[truncated]") ||
		strings.Contains(result, "truncated:") ||
		strings.Contains(result, "more lines")
	if !hasTruncateNotice {
		// 显示结果的最后部分用于调试
		suffix := result
		if len(suffix) > 500 {
			suffix = result[len(result)-500:]
		}
		t.Errorf("截断后的 diff 应包含截断提示, 结果末尾:\n%s", suffix)
	}

	// 验证保留文件头
	if !strings.Contains(result, "diff --git") {
		t.Error("截断后的 diff 应保留文件头信息")
	}
}

func TestTruncateDiff_MultipleFiles(t *testing.T) {
	p := newTestProvider("en")

	multiFileDiff := `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
+new line 1
diff --git a/file2.go b/file2.go
--- a/file2.go
+++ b/file2.go
@@ -1,3 +1,4 @@
+new line 2
diff --git a/file3.go b/file3.go
--- a/file3.go
+++ b/file3.go
@@ -1,3 +1,4 @@
+new line 3`

	result := p.TruncateDiff(multiFileDiff, 10000)

	// 验证所有文件都被保留
	if strings.Count(result, "diff --git") != 3 {
		t.Error("应保留所有 3 个文件的 diff")
	}
}
