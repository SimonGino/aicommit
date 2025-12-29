package interactive

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

// FileStatus 文件状态类型
type FileStatus int

const (
	StatusStaged FileStatus = iota
	StatusModified
	StatusUntracked
)

// FileItem 文件选择项
type FileItem struct {
	Name     string
	Status   FileStatus
	Selected bool
}

// StatusLabel 返回状态标签
func (f FileItem) StatusLabel() string {
	switch f.Status {
	case StatusStaged:
		return "✓ 已暂存"
	case StatusModified:
		return "• 已修改"
	case StatusUntracked:
		return "+ 未跟踪"
	default:
		return ""
	}
}

// readSingleKey 读取单个按键
func readSingleKey() (byte, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return 0, err
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	buf := make([]byte, 1)
	_, err = os.Stdin.Read(buf)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

// ShowFileStatusAndSelect 显示文件状态并让用户选择操作
// 返回: "use-staged", "select-files", "stage-all", "cancel"
func ShowFileStatusAndSelect(staged, modified, untracked []string) (string, error) {
	// 显示文件状态
	fmt.Println("\n检测到以下变更:")
	fmt.Println()

	if len(staged) > 0 {
		fmt.Println("已暂存 (Staged):")
		for _, f := range staged {
			fmt.Printf("  \033[32m✓\033[0m %s\n", f)
		}
		fmt.Println()
	}

	if len(modified) > 0 {
		fmt.Println("未暂存 (Modified):")
		for _, f := range modified {
			fmt.Printf("  \033[33m•\033[0m %s\n", f)
		}
		fmt.Println()
	}

	if len(untracked) > 0 {
		fmt.Println("未跟踪 (Untracked):")
		for _, f := range untracked {
			fmt.Printf("  \033[36m+\033[0m %s\n", f)
		}
		fmt.Println()
	}

	// 检查是否有变更
	hasStaged := len(staged) > 0
	hasUnstaged := len(modified) > 0 || len(untracked) > 0

	if !hasStaged && !hasUnstaged {
		fmt.Println("没有检测到任何变更")
		return "cancel", nil
	}

	// 显示选项
	fmt.Println("请选择操作:")
	if hasStaged {
		fmt.Println("  [a] 使用当前暂存区内容生成提交消息")
	}
	if hasUnstaged {
		fmt.Println("  [s] 选择要暂存的文件")
		fmt.Println("  [A] 暂存所有变更 (git add .)")
	}
	fmt.Println("  [c] 取消")
	fmt.Print("\n请按键选择: ")

	// 读取单个字符
	key, err := readSingleKey()
	if err != nil {
		return "cancel", err
	}
	fmt.Println(string(key)) // 回显按键

	switch key {
	case 'a':
		if hasStaged {
			return "use-staged", nil
		}
	case 's':
		if hasUnstaged {
			return "select-files", nil
		}
	case 'A':
		if hasUnstaged {
			return "stage-all", nil
		}
	case 'c', 'C', 3: // 3 是 Ctrl+C
		return "cancel", nil
	}

	// 如果按了无效键，使用 promptui 作为后备
	fmt.Println("\n无效选择，请使用方向键选择:")
	items := []string{}
	if hasStaged {
		items = append(items, "使用当前暂存区内容生成提交消息")
	}
	if hasUnstaged {
		items = append(items, "选择要暂存的文件")
		items = append(items, "暂存所有变更 (git add .)")
	}
	items = append(items, "取消")

	prompt := promptui.Select{
		Label: "请选择操作",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "cancel", nil
	}

	selected := items[idx]
	if strings.Contains(selected, "使用当前暂存区") {
		return "use-staged", nil
	} else if strings.Contains(selected, "选择要暂存") {
		return "select-files", nil
	} else if strings.Contains(selected, "暂存所有变更") {
		return "stage-all", nil
	}
	return "cancel", nil
}

// SelectFilesToStage 多选文件进行暂存
func SelectFilesToStage(staged, modified, untracked []string) ([]string, error) {
	// 构建所有文件列表
	var allFiles []FileItem

	for _, f := range staged {
		allFiles = append(allFiles, FileItem{Name: f, Status: StatusStaged, Selected: true})
	}
	for _, f := range modified {
		allFiles = append(allFiles, FileItem{Name: f, Status: StatusModified, Selected: false})
	}
	for _, f := range untracked {
		allFiles = append(allFiles, FileItem{Name: f, Status: StatusUntracked, Selected: false})
	}

	if len(allFiles) == 0 {
		return nil, fmt.Errorf("没有可选择的文件")
	}

	fmt.Println("\n选择要暂存的文件 (按空格切换, 回车确认):")

	// 使用简化的多选实现
	selected := make([]bool, len(allFiles))
	for i, f := range allFiles {
		selected[i] = f.Selected
	}

	for {
		// 显示当前状态
		fmt.Println()
		for i, f := range allFiles {
			checkbox := "[ ]"
			if selected[i] {
				checkbox = "[x]"
			}
			statusColor := "\033[0m"
			switch f.Status {
			case StatusStaged:
				statusColor = "\033[32m"
			case StatusModified:
				statusColor = "\033[33m"
			case StatusUntracked:
				statusColor = "\033[36m"
			}
			fmt.Printf("  %s %s%s\033[0m (%s)\n", checkbox, statusColor, f.Name, f.StatusLabel())
		}

		// 提供选项：使用 checkbox 样式
		items := []string{}
		for i, f := range allFiles {
			checkbox := "[ ]"
			if selected[i] {
				checkbox = "[x]"
			}
			items = append(items, fmt.Sprintf("%s %s", checkbox, f.Name))
		}
		items = append(items, "────────────────────")
		items = append(items, "✓ 确认选择")
		items = append(items, "✗ 取消")

		prompt := promptui.Select{
			Label: "切换选择 (回车切换状态)",
			Items: items,
			Size:  15,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, nil
			}
			return nil, err
		}

		if idx == len(items)-2 {
			// 确认选择
			break
		} else if idx == len(items)-1 {
			// 取消
			return nil, nil
		} else if idx < len(allFiles) {
			// 切换选择状态
			selected[idx] = !selected[idx]
		}
	}

	// 返回选中的文件
	var result []string
	for i, f := range allFiles {
		if selected[i] {
			result = append(result, f.Name)
		}
	}

	return result, nil
}

// CommitAction 提交操作类型
type CommitAction string

const (
	ActionAccept     CommitAction = "accept"
	ActionEdit       CommitAction = "edit"
	ActionRegenerate CommitAction = "regenerate"
	ActionCancel     CommitAction = "cancel"
)

// ShowCommitMessage 显示提交消息并让用户选择操作
func ShowCommitMessage(title, body string) (CommitAction, error) {
	// 显示消息框
	fmt.Println("\n✔ 生成的提交消息：")
	fmt.Println("┌" + strings.Repeat("─", 60) + "┐")
	fmt.Printf("│ \033[1m%s\033[0m\n", title)
	if body != "" {
		fmt.Println("│")
		for _, line := range strings.Split(body, "\n") {
			// 使用 rune 切片来正确处理多字节字符
			runes := []rune(line)
			if len(runes) > 55 {
				line = string(runes[:52]) + "..."
			}
			fmt.Printf("│ %s\n", line)
		}
	}
	fmt.Println("└" + strings.Repeat("─", 60) + "┘")

	// 显示选项
	fmt.Println("\n请选择操作:")
	fmt.Println("  [a] 接受并提交")
	fmt.Println("  [e] 编辑后提交")
	fmt.Println("  [r] 重新生成")
	fmt.Println("  [c] 取消")
	fmt.Print("\n请按键选择: ")

	// 读取单个字符
	key, err := readSingleKey()
	if err != nil {
		return ActionCancel, err
	}
	fmt.Println(string(key)) // 回显按键

	switch key {
	case 'a', 'A':
		return ActionAccept, nil
	case 'e', 'E':
		return ActionEdit, nil
	case 'r', 'R':
		return ActionRegenerate, nil
	case 'c', 'C', 3: // 3 是 Ctrl+C
		return ActionCancel, nil
	default:
		// 无效按键，使用 promptui 后备
		fmt.Println("\n无效选择，请使用方向键选择:")
		items := []string{
			"接受并提交",
			"编辑后提交",
			"重新生成",
			"取消",
		}

		prompt := promptui.Select{
			Label: "请选择操作",
			Items: items,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			return ActionCancel, nil
		}

		switch idx {
		case 0:
			return ActionAccept, nil
		case 1:
			return ActionEdit, nil
		case 2:
			return ActionRegenerate, nil
		default:
			return ActionCancel, nil
		}
	}
}

// EditMessage 编辑消息 (使用 $EDITOR 或默认 vi)
func EditMessage(content string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// 创建临时文件
	tmpfile, err := os.CreateTemp("", "aicommit-*.txt")
	if err != nil {
		return content, err
	}
	defer os.Remove(tmpfile.Name())

	// 写入内容
	if _, err := tmpfile.WriteString(content); err != nil {
		return content, err
	}
	tmpfile.Close()

	// 打开编辑器
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return content, err
	}

	// 读取编辑后的内容
	edited, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return content, err
	}

	return strings.TrimSpace(string(edited)), nil
}

// PromptConfirm 确认提示
func PromptConfirm(message string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
