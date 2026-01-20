package interactive

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mattn/go-runewidth"
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

type KeyEvent int

const (
	KeyUnknown KeyEvent = iota
	KeyUp
	KeyDown
	KeyEnter
	KeySpace
	KeyEsc
	KeyQ
	KeyA
)

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

func readKeyEvent() KeyEvent {
	buf := make([]byte, 3)
	n, err := os.Stdin.Read(buf)
	if err != nil || n == 0 {
		return KeyUnknown
	}

	if n == 1 {
		switch buf[0] {
		case 13, 10:
			return KeyEnter
		case 32:
			return KeySpace
		case 27:
			return KeyEsc
		case 'q', 'Q':
			return KeyQ
		case 'a', 'A':
			return KeyA
		}
		return KeyUnknown
	}

	// Arrow keys escape sequence: ESC [ A/B
	if n >= 3 && buf[0] == 27 && buf[1] == '[' {
		switch buf[2] {
		case 'A':
			return KeyUp
		case 'B':
			return KeyDown
		}
	}

	return KeyUnknown
}

// displayWidth calculate the display width of a string (handles ANSI codes and CJK chars)
func displayWidth(s string) int {
	// Remove ANSI escape codes
	cleaned := s
	for {
		start := strings.Index(cleaned, "\033[")
		if start == -1 {
			break
		}
		end := strings.Index(cleaned[start:], "m")
		if end == -1 {
			break
		}
		cleaned = cleaned[:start] + cleaned[start+end+1:]
	}
	// Use runewidth to correctly calculate CJK character width (2 columns each)
	return runewidth.StringWidth(cleaned)
}

// printBox print content in a box with full borders
func printBox(title string, lines []string, width int) {
	// Calculate the actual width needed
	titleWidth := displayWidth(title)

	// Box width = "│ " + content + " │" = width + 2 for content area
	// Top border: ┌─ + title + ─...─ + ┐
	// We need: total width = 2 (┌─) + titleWidth + remaining dashes + 1 (┐)
	// Content line width = 2 (│ ) + width + 2 ( │) = width + 4
	// So top border should also be width + 4 total characters

	// Top border (title in the border)
	topBorderWidth := width + 2 - titleWidth - 1 // -1 for the initial ─ after ┌
	if topBorderWidth < 1 {
		topBorderWidth = 1
	}
	fmt.Println("┌─" + title + strings.Repeat("─", topBorderWidth) + "┐")

	// Content lines
	for _, line := range lines {
		lineWidth := displayWidth(line)
		padding := width - lineWidth
		if padding < 0 {
			padding = 0
		}
		fmt.Printf("│ %s%s │\n", line, strings.Repeat(" ", padding))
	}

	// Bottom border: width + 2 for ─ to match │ content │
	fmt.Println("└" + strings.Repeat("─", width+2) + "┘")
}

// ShowFileStatusAndSelect 显示文件状态并让用户选择操作
// 返回: "use-staged", "select-files", "stage-all", "cancel"
func ShowFileStatusAndSelect(staged, modified, untracked []string) (string, error) {
	// 准备要显示的行
	var lines []string
	maxWidth := 60 // default box width

	if len(staged) > 0 {
		lines = append(lines, "")
		lines = append(lines, "\033[1m已暂存 (Staged):\033[0m")
		for _, f := range staged {
			line := fmt.Sprintf("  \033[32m✓\033[0m %s", f)
			lines = append(lines, line)
			// Update max width if needed
			if w := displayWidth(line) + 2; w > maxWidth {
				maxWidth = w
			}
		}
	}

	if len(modified) > 0 {
		lines = append(lines, "")
		lines = append(lines, "\033[1m未暂存 (Modified):\033[0m")
		for _, f := range modified {
			line := fmt.Sprintf("  \033[33m•\033[0m %s", f)
			lines = append(lines, line)
			if w := displayWidth(line) + 2; w > maxWidth {
				maxWidth = w
			}
		}
	}

	if len(untracked) > 0 {
		lines = append(lines, "")
		lines = append(lines, "\033[1m未跟踪 (Untracked):\033[0m")
		for _, f := range untracked {
			line := fmt.Sprintf("  \033[36m+\033[0m %s", f)
			lines = append(lines, line)
			if w := displayWidth(line) + 2; w > maxWidth {
				maxWidth = w
			}
		}
	}

	// Display the box
	if len(lines) > 0 {
		fmt.Println()
		printBox("检测到以下变更", lines, maxWidth)
	}

	// 检查是否有变更
	hasStaged := len(staged) > 0
	hasUnstaged := len(modified) > 0 || len(untracked) > 0

	if !hasStaged && !hasUnstaged {
		fmt.Println("没有检测到任何变更")
		return "cancel", nil
	}

	// 准备选项列表
	type Option struct {
		Key       string
		Label     string
		Action    string
		IsDefault bool
	}

	var options []Option
	var defaultAction string

	// Default option comes first
	if hasStaged {
		// When staged files exist, "use staged" is default
		options = append(options, Option{
			Key:       "u",
			Label:     "使用当前暂存区内容生成提交消息",
			Action:    "use-staged",
			IsDefault: true,
		})
		defaultAction = "use-staged"
	}

	if hasUnstaged {
		isDefault := !hasStaged // default only when no staged files
		options = append(options, Option{
			Key:       "a",
			Label:     "暂存所有变更并生成提交消息",
			Action:    "stage-all",
			IsDefault: isDefault,
		})
		if isDefault {
			defaultAction = "stage-all"
		}
		options = append(options, Option{
			Key:       "s",
			Label:     "选择要暂存的文件",
			Action:    "select-files",
			IsDefault: false,
		})
	}

	options = append(options, Option{
		Key:       "q",
		Label:     "取消",
		Action:    "cancel",
		IsDefault: false,
	})

	// 显示选项框
	var optionLines []string
	optionMaxWidth := 40

	for _, opt := range options {
		var line string
		if opt.IsDefault {
			line = fmt.Sprintf("  \033[1m[%s]\033[0m %s \033[90m(默认)\033[0m", opt.Key, opt.Label)
		} else {
			line = fmt.Sprintf("  [%s] %s", opt.Key, opt.Label)
		}
		optionLines = append(optionLines, line)
		if w := displayWidth(line) + 2; w > optionMaxWidth {
			optionMaxWidth = w
		}
	}

	optionLines = append(optionLines, "")
	optionLines = append(optionLines, "\033[90m提示: 输入字母或直接按回车选择默认选项\033[0m")

	fmt.Println()
	printBox("请选择操作", optionLines, optionMaxWidth)
	fmt.Print("\n请输入选择: ")

	// 读取单个字符
	key, err := readSingleKey()
	if err != nil {
		return "cancel", err
	}

	// Handle Enter key (newline)
	if key == 13 || key == 10 {
		fmt.Println("(回车)")
		return defaultAction, nil
	}

	fmt.Println(string(key)) // 回显按键

	// 根据按键返回对应的操作
	for _, opt := range options {
		if strings.EqualFold(string(key), opt.Key) {
			return opt.Action, nil
		}
	}

	// 如果按了无效键,使用 promptui 作为后备
	fmt.Println("\n\033[33m无效选择\033[0m,请使用方向键选择:")

	var fallbackItems []string
	var actions []string

	for _, opt := range options {
		fallbackItems = append(fallbackItems, opt.Label)
		actions = append(actions, opt.Action)
	}

	prompt := promptui.Select{
		Label: "请选择操作",
		Items: fallbackItems,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return "cancel", nil
	}

	return actions[idx], nil
}

func SelectFilesToStage(staged, modified, untracked []string) ([]string, error) {
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

	selected := make([]bool, len(allFiles))
	for i, f := range allFiles {
		selected[i] = f.Selected
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	cursorPos := 0
	totalItems := len(allFiles) + 2
	totalLines := len(allFiles) + 5
	firstRender := true

	renderList := func() {
		if !firstRender {
			fmt.Printf("\033[%dA\r", totalLines)
		}
		firstRender = false

		fmt.Print("\033[2K选择要暂存的文件 (↑↓移动, 空格切换, a全选/取消全选, 回车确认, q取消):\r\n")
		fmt.Print("\033[2K\r\n")

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

			fmt.Print("\033[2K")
			if i == cursorPos {
				fmt.Printf("\033[7m▸ %s %s%s\033[0m (%s)\033[0m\r\n", checkbox, statusColor, f.Name, f.StatusLabel())
			} else {
				fmt.Printf("  %s %s%s\033[0m (%s)\r\n", checkbox, statusColor, f.Name, f.StatusLabel())
			}
		}

		fmt.Print("\033[2K\r\n")
		fmt.Print("\033[2K────────────────────\r\n")

		confirmIdx := len(allFiles)
		cancelIdx := len(allFiles) + 1

		fmt.Print("\033[2K")
		if cursorPos == confirmIdx {
			fmt.Print("\033[7m▸ ✓ 确认选择\033[0m\r\n")
		} else {
			fmt.Print("  ✓ 确认选择\r\n")
		}

		fmt.Print("\033[2K")
		if cursorPos == cancelIdx {
			fmt.Print("\033[7m▸ ✗ 取消\033[0m\r\n")
		} else {
			fmt.Print("  ✗ 取消\r\n")
		}
	}

	renderList()

	for {
		key := readKeyEvent()

		switch key {
		case KeyUp:
			if cursorPos > 0 {
				cursorPos--
			}
		case KeyDown:
			if cursorPos < totalItems-1 {
				cursorPos++
			}
		case KeySpace:
			if cursorPos < len(allFiles) {
				selected[cursorPos] = !selected[cursorPos]
			}
		case KeyA:
			allSelected := true
			for _, s := range selected {
				if !s {
					allSelected = false
					break
				}
			}
			for i := range selected {
				selected[i] = !allSelected
			}
		case KeyEnter:
			if cursorPos == len(allFiles) {
				fmt.Print("\r\n")
				var result []string
				for i, f := range allFiles {
					if selected[i] {
						result = append(result, f.Name)
					}
				}
				return result, nil
			} else if cursorPos == len(allFiles)+1 {
				fmt.Print("\r\n")
				return nil, nil
			} else {
				selected[cursorPos] = !selected[cursorPos]
			}
		case KeyQ, KeyEsc:
			fmt.Print("\r\n")
			return nil, nil
		}

		renderList()
	}
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
	// 准备要显示的行
	var lines []string
	maxWidth := 60 // default width

	// Add title (bold)
	titleLine := fmt.Sprintf("\033[1m%s\033[0m", title)
	lines = append(lines, titleLine)
	if w := displayWidth(titleLine) + 2; w > maxWidth {
		maxWidth = w
	}

	// Add body if present
	if body != "" {
		lines = append(lines, "")
		for _, line := range strings.Split(body, "\n") {
			lines = append(lines, line)
			if w := displayWidth(line) + 2; w > maxWidth {
				maxWidth = w
			}
		}
	}

	// Display the box
	fmt.Println()
	printBox("✔ 生成的提交消息", lines, maxWidth)

	// 准备选项框
	type Option struct {
		Key       string
		Label     string
		IsDefault bool
	}

	options := []Option{
		{Key: "a", Label: "接受并提交", IsDefault: true},
		{Key: "e", Label: "编辑后提交", IsDefault: false},
		{Key: "r", Label: "重新生成", IsDefault: false},
		{Key: "c", Label: "取消", IsDefault: false},
	}

	var optionLines []string
	optionMaxWidth := 40

	for _, opt := range options {
		var line string
		if opt.IsDefault {
			line = fmt.Sprintf("  \033[1m[%s]\033[0m %s \033[90m(默认)\033[0m", opt.Key, opt.Label)
		} else {
			line = fmt.Sprintf("  [%s] %s", opt.Key, opt.Label)
		}
		optionLines = append(optionLines, line)
		if w := displayWidth(line) + 2; w > optionMaxWidth {
			optionMaxWidth = w
		}
	}

	optionLines = append(optionLines, "")
	optionLines = append(optionLines, "\033[90m提示: 输入字母或直接按回车选择默认选项\033[0m")

	fmt.Println()
	printBox("请选择操作", optionLines, optionMaxWidth)
	fmt.Print("\n请按键选择: ")

	// 读取单个字符
	key, err := readSingleKey()
	if err != nil {
		return ActionCancel, err
	}

	// Handle Enter key - default to accept
	if key == 13 || key == 10 {
		fmt.Println("(回车)")
		return ActionAccept, nil
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
		fmt.Println("\n\033[33m无效选择\033[0m，请使用方向键选择:")
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
