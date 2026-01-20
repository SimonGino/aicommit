package git

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Repository struct {
	path string
}

func unquoteGitPath(s string) string {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return s
	}

	s = s[1 : len(s)-1]
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result.WriteByte('\n')
				i++
			case 't':
				result.WriteByte('\t')
				i++
			case '\\':
				result.WriteByte('\\')
				i++
			case '"':
				result.WriteByte('"')
				i++
			default:
				if i+3 < len(s) && isOctalDigit(s[i+1]) && isOctalDigit(s[i+2]) && isOctalDigit(s[i+3]) {
					octal := s[i+1 : i+4]
					if val, err := strconv.ParseInt(octal, 8, 32); err == nil {
						result.WriteByte(byte(val))
						i += 3
						continue
					}
				}
				result.WriteByte(s[i])
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}

func isOctalDigit(b byte) bool {
	return b >= '0' && b <= '7'
}

// GetRepo 获取Git仓库实例
func GetRepo(path string) (*Repository, error) {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("获取当前目录失败: %w", err)
		}
	}

	// 检查是否是Git仓库
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("当前目录不是Git仓库，请先运行 'git init'")
	}

	return &Repository{path: path}, nil
}

func (r *Repository) GetUnstagedChanges() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取未暂存更改失败: %w", err)
	}

	if len(output) == 0 {
		return nil, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			files = append(files, unquoteGitPath(line))
		}
	}
	return files, nil
}

func (r *Repository) GetStagedChanges() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取已暂存更改失败: %w", err)
	}

	if len(output) == 0 {
		return nil, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			files = append(files, unquoteGitPath(line))
		}
	}
	return files, nil
}

// GetDiff 获取指定文件的差异内容
func (r *Repository) GetDiff(staged bool) (string, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--cached")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取差异内容失败: %w", err)
	}

	return string(output), nil
}

// GetCurrentBranch 获取当前分支名
// 对于刚 git init 但尚未有任何 commit 的仓库，返回 unborn 分支名
func (r *Repository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		// git init 后尚未有 commit，HEAD 无法解析
		// 尝试从 symbolic-ref 获取 unborn 分支名
		cmdSymRef := exec.Command("git", "symbolic-ref", "--short", "HEAD")
		cmdSymRef.Dir = r.path
		symRefOutput, symErr := cmdSymRef.Output()
		if symErr == nil {
			return strings.TrimSpace(string(symRefOutput)), nil
		}
		// 两种方式都失败，返回默认值 "main"
		return "main", nil
	}

	return strings.TrimSpace(string(output)), nil
}

// Commit 提交更改
func (r *Repository) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = r.path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("提交更改失败: %w", err)
	}

	return nil
}

// StageAll 暂存所有更改
func (r *Repository) StageAll() error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = r.path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("暂存更改失败: %w", err)
	}

	return nil
}

func (r *Repository) GetUntrackedFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取未跟踪文件失败: %w", err)
	}

	if len(output) == 0 {
		return nil, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	files := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			files = append(files, unquoteGitPath(line))
		}
	}
	return files, nil
}

// GetAllChanges 获取所有变更文件 (已暂存、未暂存、未跟踪)
func (r *Repository) GetAllChanges() (staged, modified, untracked []string, err error) {
	staged, err = r.GetStagedChanges()
	if err != nil {
		return nil, nil, nil, err
	}

	modified, err = r.GetUnstagedChanges()
	if err != nil {
		return nil, nil, nil, err
	}

	untracked, err = r.GetUntrackedFiles()
	if err != nil {
		return nil, nil, nil, err
	}

	return staged, modified, untracked, nil
}

// StageFiles 暂存指定文件
func (r *Repository) StageFiles(files []string) error {
	if len(files) == 0 {
		return nil
	}

	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("暂存文件失败: %w", err)
	}

	return nil
}

// GetDiffForFiles 获取指定文件的差异内容
func (r *Repository) GetDiffForFiles(files []string, staged bool) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	args := []string{"diff"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--")
	args = append(args, files...)

	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取差异内容失败: %w", err)
	}

	return string(output), nil
}

// GetUserInfo 获取 Git 用户信息
func (r *Repository) GetUserInfo() (name, email string, err error) {
	cmdName := exec.Command("git", "config", "user.name")
	cmdName.Dir = r.path
	nameBytes, err := cmdName.Output()
	if err != nil {
		// 尝试从全局配置获取
		cmdNameGlobal := exec.Command("git", "config", "--global", "user.name")
		nameBytes, err = cmdNameGlobal.Output()
		if err != nil {
			return "", "", fmt.Errorf("获取 git user.name 失败，请检查本地或全局配置: %w", err)
		}
	}
	name = strings.TrimSpace(string(nameBytes))

	cmdEmail := exec.Command("git", "config", "user.email")
	cmdEmail.Dir = r.path
	emailBytes, err := cmdEmail.Output()
	if err != nil {
		// 尝试从全局配置获取
		cmdEmailGlobal := exec.Command("git", "config", "--global", "user.email")
		emailBytes, err = cmdEmailGlobal.Output()
		if err != nil {
			return name, "", fmt.Errorf("获取 git user.email 失败，请检查本地或全局配置: %w", err)
		}
	}
	email = strings.TrimSpace(string(emailBytes))

	if name == "" || email == "" {
		return name, email, fmt.Errorf("Git 用户名或邮箱未配置，请使用 'git config user.name' 和 'git config user.email' 进行设置")
	}

	return name, email, nil
}

// GetCommits 获取指定作者在时间范围内的提交记录 (格式: "YYYY-MM-DD -- Subject")
func (r *Repository) GetCommits(authorEmail string, since, until string) ([]string, error) {
	// 使用 %cs 获取 YYYY-MM-DD 格式的日期
	args := []string{"log", fmt.Sprintf("--author=%s", authorEmail), "--pretty=format:%cs -- %s", "--date=short"}
	if since != "" {
		args = append(args, fmt.Sprintf("--since=%s", since))
	}
	if until != "" {
		args = append(args, fmt.Sprintf("--until=%s", until))
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		// 如果没有commit，git log会返回非0退出码，但output是空的
		if _, ok := err.(*exec.ExitError); ok && len(output) == 0 {
			return []string{}, nil // 没有找到commit，返回空列表，不算错误
		}
		return nil, fmt.Errorf("获取提交记录失败: %w", err)
	}

	if len(output) == 0 {
		return []string{}, nil
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	filteredCommits := make([]string, 0, len(commits))
	for _, c := range commits {
		trimmedCommit := strings.TrimSpace(c)
		if trimmedCommit == "" {
			continue
		}
		// 分割日期和主题
		parts := strings.SplitN(trimmedCommit, " -- ", 2)
		if len(parts) == 2 {
			commitSubject := strings.TrimSpace(parts[1])
			// 过滤 Merge commits
			if strings.HasPrefix(commitSubject, "Merge branch") || strings.HasPrefix(commitSubject, "Merge remote-tracking branch") {
				continue
			}
		}
		filteredCommits = append(filteredCommits, trimmedCommit)
	}

	return filteredCommits, nil
}
