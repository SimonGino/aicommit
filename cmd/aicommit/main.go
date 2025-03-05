package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/SimonGino/aicommit/internal/ai"
	"github.com/SimonGino/aicommit/internal/config"
	"github.com/SimonGino/aicommit/internal/git"
	"github.com/urfave/cli/v2"
)

// 版本信息，由 GoReleaser 在构建时注入
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := &cli.App{
		Name:    "aicommit",
		Usage:   "AI驱动的git commit消息生成器",
		Version: getVersion(),
		Commands: []*cli.Command{
			{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "配置AI提供商设置",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "provider",
						Aliases: []string{"p"},
						Usage:   "AI提供商 (qwen, openai, deepseek)",
					},
					&cli.StringFlag{
						Name:    "api-key",
						Aliases: []string{"k"},
						Usage:   "API密钥",
					},
					&cli.StringFlag{
						Name:    "language",
						Aliases: []string{"l"},
						Usage:   "输出语言 (en, zh-CN, zh-TW)",
					},
				},
				Action: configAction,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "provider",
				Aliases: []string{"p"},
				Usage:   "使用指定的AI提供商",
			},
			&cli.StringFlag{
				Name:    "message",
				Aliases: []string{"m"},
				Usage:   "使用指定的提交消息（跳过AI生成）",
			},
		},
		Action: defaultAction,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func configAction(c *cli.Context) error {
	cfg := config.LoadConfig()

	if provider := c.String("provider"); provider != "" {
		if apiKey := c.String("api-key"); apiKey != "" {
			if err := cfg.UpdateAPIKey(provider, apiKey); err != nil {
				return fmt.Errorf("配置API密钥失败: %w", err)
			}
			fmt.Printf("✓ 成功配置 %s API密钥\n", provider)
		}
	}

	if language := c.String("language"); language != "" {
		if err := cfg.UpdateLanguage(language); err != nil {
			return fmt.Errorf("更新语言失败: %w", err)
		}
		fmt.Printf("✓ 成功设置语言为 %s\n", language)
	}

	fmt.Printf("配置文件: %s\n", cfg.ConfigFile())
	return nil
}

func defaultAction(c *cli.Context) error {
	repo, err := git.GetRepo("")
	if err != nil {
		return fmt.Errorf("获取Git仓库失败: %w", err)
	}

	// 获取未暂存的更改
	unstaged, err := repo.GetUnstagedChanges()
	if err != nil {
		return fmt.Errorf("获取未暂存更改失败: %w", err)
	}

	if len(unstaged) > 0 {
		fmt.Println("\n未暂存的更改:")
		for _, file := range unstaged {
			fmt.Printf("  • %s\n", file)
		}

		if c.String("message") == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("\n是否要暂存这些更改？[y/N] ")
			answer, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("读取用户输入失败: %w", err)
			}
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer == "y" {
				if err := repo.StageAll(); err != nil {
					return err
				}
				fmt.Println("✓ 已暂存所有更改")
			}
		}
	}

	// 获取已暂存的更改
	staged, err := repo.GetStagedChanges()
	if err != nil {
		return fmt.Errorf("获取已暂存更改失败: %w", err)
	}

	if len(staged) == 0 {
		return fmt.Errorf("没有找到已暂存的更改。使用 'git add' 来暂存你的更改")
	}

	fmt.Println("\n已暂存的更改:")
	for _, file := range staged {
		fmt.Printf("  • %s\n", file)
	}

	// 如果指定了提交消息，直接使用
	if message := c.String("message"); message != "" {
		if err := repo.Commit(message); err != nil {
			return err
		}
		fmt.Printf("✓ 已提交更改：%s\n", message)
		return nil
	}

	// 获取差异内容
	diff, err := repo.GetDiff(true)
	if err != nil {
		return fmt.Errorf("获取差异内容失败: %w", err)
	}

	// 获取当前分支
	branch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("获取当前分支失败: %w", err)
	}

	// 准备提交信息
	commitInfo := &ai.CommitInfo{
		FilesChanged: staged,
		DiffContent:  diff,
		BranchName:   branch,
	}

	// 加载配置
	cfg := config.LoadConfig()
	provider := c.String("provider")
	if provider == "" {
		provider = cfg.DefaultProvider
	}

	// 创建AI提供商实例
	aiProvider, err := ai.NewProvider(provider, cfg.GetAPIKey(provider), cfg.Language)
	if err != nil {
		return fmt.Errorf("创建AI提供商实例失败: %w", err)
	}

	fmt.Println("\n正在生成提交消息...")
	message, err := aiProvider.GenerateCommitMessage(context.Background(), commitInfo)
	if err != nil {
		return fmt.Errorf("生成提交消息失败: %w", err)
	}

	// 显示生成的消息
	fmt.Printf("\n生成的提交消息:\n")
	fmt.Printf("标题: %s\n", message.Title)
	if message.Body != "" {
		fmt.Printf("正文:\n%s\n", message.Body)
	}

	// 确认提交
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n是否要使用这个消息提交？[Y/n] ")
	answer, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("读取用户输入失败: %w", err)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "" || answer == "y" {
		commitMessage := message.Title
		if message.Body != "" {
			commitMessage += "\n\n" + message.Body
		}
		if err := repo.Commit(commitMessage); err != nil {
			return err
		}
		fmt.Println("✓ 已提交更改")
	} else {
		fmt.Println("提交已取消")
	}

	return nil
}

// 添加版本信息处理函数
func getVersion() string {
	commitHash := commit
	if len(commit) >= 7 {
		commitHash = commit[:7]
	}
	return fmt.Sprintf("%s (%s - %s)", version, commitHash, date)
}
