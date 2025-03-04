package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Repository struct {
	path string
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

// GetUnstagedChanges 获取未暂存的更改
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

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	return files, nil
}

// GetStagedChanges 获取已暂存的更改
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

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
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
func (r *Repository) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = r.path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("获取当前分支失败: %w", err)
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
