package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/SimonGino/aicommit/internal/ai"
	"github.com/SimonGino/aicommit/internal/config"
	"github.com/SimonGino/aicommit/internal/git"
	"github.com/SimonGino/aicommit/internal/interactive"
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
				Usage:   "配置AI设置",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "api-key",
						Aliases: []string{"k"},
						Usage:   "OpenAI API密钥",
					},
					&cli.StringFlag{
						Name:    "base-url",
						Aliases: []string{"u"},
						Usage:   "自定义API基础URL",
					},
					&cli.StringFlag{
						Name:    "model",
						Aliases: []string{"m"},
						Usage:   "指定模型名称",
					},
					&cli.StringFlag{
						Name:    "language",
						Aliases: []string{"l"},
						Usage:   "输出语言 (en, zh-CN, zh-TW)",
					},
					&cli.StringFlag{
						Name:    "provider",
						Aliases: []string{"p"},
						Usage:   "AI提供商 (openai, azure)",
					},
					&cli.StringFlag{
						Name:  "azure-api-version",
						Usage: "Azure OpenAI API版本 (默认: 2024-02-15-preview)",
					},
				},
				Action: configAction,
			},
			{
				Name:    "report",
				Aliases: []string{"r"},
				Usage:   "根据Git提交历史生成日报",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "language",
						Aliases: []string{"l"},
						Usage:   "指定日报语言 (默认使用配置)",
					},
					&cli.BoolFlag{
						Name:  "this-week",
						Usage: "生成本周的日报",
					},
					&cli.BoolFlag{
						Name:  "last-week",
						Usage: "生成上周的日报",
					},
					&cli.StringFlag{
						Name:  "since",
						Usage: "指定开始日期 (YYYY-MM-DD)",
					},
					&cli.StringFlag{
						Name:  "until",
						Usage: "指定结束日期 (YYYY-MM-DD)",
					},
					&cli.StringFlag{
						Name:  "author",
						Usage: "指定作者邮箱 (默认使用当前Git配置)",
					},
				},
				Action: reportAction,
			},
			{
				Name:   "check",
				Usage:  "检查配置和 API 连通性",
				Action: checkAction,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "message",
				Aliases: []string{"m"},
				Usage:   "使用指定的提交消息（跳过AI生成）",
			},
			&cli.StringFlag{
				Name:    "language",
				Aliases: []string{"l"},
				Usage:   "指定输出语言 (en, zh-CN, zh-TW)",
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

	if apiKey := c.String("api-key"); apiKey != "" {
		if err := cfg.UpdateAPIKey(apiKey); err != nil {
			return fmt.Errorf("配置API密钥失败: %w", err)
		}
		fmt.Printf("✓ 成功配置 API 密钥\n")
	}

	if baseURL := c.String("base-url"); baseURL != "" {
		if err := cfg.UpdateBaseURL(baseURL); err != nil {
			return fmt.Errorf("配置 URL 失败: %w", err)
		}
		fmt.Printf("✓ 成功配置 URL: %s\n", baseURL)
	}

	if model := c.String("model"); model != "" {
		if err := cfg.UpdateModel(model); err != nil {
			return fmt.Errorf("配置模型失败: %w", err)
		}
		fmt.Printf("✓ 成功配置模型: %s\n", model)
	}

	if language := c.String("language"); language != "" {
		if err := cfg.UpdateLanguage(language); err != nil {
			return fmt.Errorf("更新语言失败: %w", err)
		}
		fmt.Printf("✓ 成功设置语言为 %s\n", language)
	}

	if provider := c.String("provider"); provider != "" {
		if err := cfg.UpdateProvider(provider); err != nil {
			return fmt.Errorf("配置提供商失败: %w", err)
		}
		fmt.Printf("✓ 成功配置提供商: %s\n", provider)
	}

	if azureAPIVersion := c.String("azure-api-version"); azureAPIVersion != "" {
		if err := cfg.UpdateAzureAPIVersion(azureAPIVersion); err != nil {
			return fmt.Errorf("配置Azure API版本失败: %w", err)
		}
		fmt.Printf("✓ 成功配置 Azure API 版本: %s\n", azureAPIVersion)
	}

	fmt.Printf("配置文件: %s\n", cfg.ConfigFile())
	return nil
}

func reportAction(c *cli.Context) error {
	repo, err := git.GetRepo("")
	if err != nil {
		return fmt.Errorf("获取Git仓库失败: %w", err)
	}

	// 获取作者邮箱
	authorEmail := c.String("author")
	if authorEmail == "" {
		_, email, err := repo.GetUserInfo()
		if err != nil {
			return fmt.Errorf("获取Git用户信息失败: %w", err)
		}
		authorEmail = email
		if authorEmail == "" {
			return fmt.Errorf("无法确定作者邮箱，请使用 --author 参数指定或配置Git user.email")
		}
	}

	// 解析日期范围
	since, until, err := parseDateRange(c)
	if err != nil {
		return err
	}

	fmt.Printf("正在为 %s 获取 %s 到 %s 的提交记录...\n", authorEmail, since, until)
	commits, err := repo.GetCommits(authorEmail, since, until)
	if err != nil {
		return fmt.Errorf("获取提交记录失败: %w", err)
	}

	if len(commits) == 0 {
		fmt.Println("在指定时间范围内没有找到该作者的提交记录。")
		return nil
	}

	fmt.Printf("找到 %d 条提交记录，正在生成日报...\n", len(commits))

	// 加载配置
	cfg := config.LoadConfig()

	language := c.String("language")
	if language == "" {
		language = cfg.Language
	} else {
		if err := validateLanguage(language); err != nil {
			return err
		}
		// 修正 zh -> zh-CN
		if language == "zh" {
			language = "zh-CN"
		}
	}

	apiKey := cfg.APIKey
	if cfg.Provider == "azure" {
		if cfg.APIKey == "" {
			return fmt.Errorf("未配置 Azure OpenAI API 密钥，请先使用 'aicommit config --api-key YOUR_API_KEY' 配置")
		}
		if cfg.BaseURL == "" {
			return fmt.Errorf("未配置 Azure OpenAI endpoint URL，请先使用 'aicommit config --base-url YOUR_ENDPOINT_URL' 配置")
		}
	} else {
		if apiKey == "" {
			return fmt.Errorf("未配置 OpenAI API 密钥，请先使用 'aicommit config --api-key YOUR_API_KEY' 配置")
		}
	}

	aiProvider, err := ai.NewProvider(apiKey, cfg.BaseURL, cfg.Model, language, cfg.Provider, cfg.AzureAPIVersion)
	if err != nil {
		return fmt.Errorf("创建AI提供商实例失败: %w", err)
	}

	reportInfo := &ai.ReportInfo{
		Commits: commits,
	}

	reportContent, err := aiProvider.GenerateDailyReport(context.Background(), reportInfo, since, until)
	if err != nil {
		return fmt.Errorf("生成日报失败: %w", err)
	}

	fmt.Println("\n--- 生成的日报 ---")
	fmt.Println(reportContent)
	fmt.Println("--- 日报结束 ---")

	return nil
}

// parseDateRange 解析日期范围标志
func parseDateRange(c *cli.Context) (since, until string, err error) {
	dateFormat := "2006-01-02"
	now := time.Now()

	if c.Bool("this-week") {
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7 // Adjust Sunday to be the 7th day
		}
		startOfWeek := now.AddDate(0, 0, -int(weekday)+1)
		endOfWeek := startOfWeek.AddDate(0, 0, 6)
		since = startOfWeek.Format(dateFormat)
		until = endOfWeek.Format(dateFormat)
	} else if c.Bool("last-week") {
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7 // Adjust Sunday to be the 7th day
		}
		startOfLastWeek := now.AddDate(0, 0, -int(weekday)+1-7)
		endOfLastWeek := startOfLastWeek.AddDate(0, 0, 6)
		since = startOfLastWeek.Format(dateFormat)
		until = endOfLastWeek.Format(dateFormat)
	} else {
		since = c.String("since")
		until = c.String("until")

		// 如果只提供了 since，until 默认为今天
		if since != "" && until == "" {
			until = now.Format(dateFormat)
		}
		// 如果只提供了 until，since 默认为空（git log 会处理）
		// 如果都没有提供，则 since 和 until 都为空，git log 会获取所有历史
		// 我们需要一个默认行为，比如默认获取本周？或者报错？这里先按不提供则获取全部处理
		// 增加校验：如果提供了 since 或 until，必须是 YYYY-MM-DD 格式
		if since != "" {
			if _, pErr := time.Parse(dateFormat, since); pErr != nil {
				err = fmt.Errorf("无效的开始日期格式: %s，请使用 YYYY-MM-DD", since)
				return
			}
		}
		if until != "" {
			if _, pErr := time.Parse(dateFormat, until); pErr != nil {
				err = fmt.Errorf("无效的结束日期格式: %s，请使用 YYYY-MM-DD", until)
				return
			}
		}

		// 默认行为：如果 since 和 until 都为空，默认获取本周
		if since == "" && until == "" {
			weekday := now.Weekday()
			if weekday == time.Sunday {
				weekday = 7 // Adjust Sunday to be the 7th day
			}
			startOfWeek := now.AddDate(0, 0, -int(weekday)+1)
			endOfWeek := startOfWeek.AddDate(0, 0, 6)
			since = startOfWeek.Format(dateFormat)
			until = endOfWeek.Format(dateFormat)
			fmt.Printf("未指定日期范围，默认使用本周 (%s - %s)\n", since, until)
		}
	}

	return
}

// validateLanguage 验证语言是否支持
func validateLanguage(lang string) error {
	switch lang {
	case "en", "zh-CN", "zh-TW", "zh":
		return nil
	default:
		return fmt.Errorf("不支持的语言: %s，请使用 en, zh-CN, zh-TW", lang)
	}
}

func checkAction(c *cli.Context) error {
	cfg := config.LoadConfig()

	// 创建 AI 提供商实例
	provider, err := ai.NewProvider(
		cfg.APIKey, cfg.BaseURL, cfg.Model,
		cfg.Language, cfg.Provider, cfg.AzureAPIVersion,
	)
	if err != nil {
		fmt.Printf("\n✗ 创建 AI 提供商失败: %v\n", err)
		return nil
	}

	// 执行检测
	result := provider.(*ai.OpenAIProvider).Check(c.Context)
	ai.PrintCheckResult(result)
	return nil
}

func defaultAction(c *cli.Context) error {
	repo, err := git.GetRepo("")
	if err != nil {
		return fmt.Errorf("获取Git仓库失败: %w", err)
	}

	// 如果指定了提交消息，直接使用旧逻辑
	if message := c.String("message"); message != "" {
		staged, err := repo.GetStagedChanges()
		if err != nil {
			return fmt.Errorf("获取已暂存更改失败: %w", err)
		}
		if len(staged) == 0 {
			return fmt.Errorf("没有找到已暂存的更改。使用 'git add' 来暂存你的更改")
		}
		if err := repo.Commit(message); err != nil {
			return err
		}
		fmt.Printf("✓ 已提交更改：%s\n", message)
		return nil
	}

	// 获取所有变更
	staged, modified, untracked, err := repo.GetAllChanges()
	if err != nil {
		return fmt.Errorf("获取变更失败: %w", err)
	}

	// 检查是否有任何变更
	if len(staged) == 0 && len(modified) == 0 && len(untracked) == 0 {
		fmt.Println("没有检测到任何变更")
		return nil
	}

	// 显示文件状态并让用户选择操作
	action, err := interactive.ShowFileStatusAndSelect(staged, modified, untracked)
	if err != nil {
		return fmt.Errorf("交互式选择失败: %w", err)
	}

	switch action {
	case "use-staged":
		// 使用当前暂存区
	case "select-files":
		files, err := interactive.SelectFilesToStage(staged, modified, untracked)
		if err != nil {
			return fmt.Errorf("文件选择失败: %w", err)
		}
		if len(files) == 0 {
			fmt.Println("没有选择任何文件，操作取消")
			return nil
		}
		if err := repo.StageFiles(files); err != nil {
			return err
		}
		fmt.Printf("✓ 已暂存 %d 个文件\n", len(files))
	case "stage-all":
		if err := repo.StageAll(); err != nil {
			return err
		}
		fmt.Println("✓ 已暂存所有变更")
	case "cancel":
		fmt.Println("操作已取消")
		return nil
	default:
		return nil
	}

	// 重新获取已暂存的更改
	staged, err = repo.GetStagedChanges()
	if err != nil {
		return fmt.Errorf("获取已暂存更改失败: %w", err)
	}

	if len(staged) == 0 {
		return fmt.Errorf("没有找到已暂存的更改")
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

	// 获取语言设置，优先使用命令行参数
	language := c.String("language")
	if language == "" {
		language = cfg.Language
	} else {
		if err := validateLanguage(language); err != nil {
			return err
		}
		if language == "zh" {
			language = "zh-CN"
		}
	}

	// 创建AI提供商实例
	apiKey := cfg.APIKey
	if cfg.Provider == "azure" {
		if cfg.APIKey == "" {
			return fmt.Errorf("未配置 Azure OpenAI API 密钥，请先使用 'aicommit config --api-key YOUR_API_KEY' 配置")
		}
		if cfg.BaseURL == "" {
			return fmt.Errorf("未配置 Azure OpenAI endpoint URL，请先使用 'aicommit config --base-url YOUR_ENDPOINT_URL' 配置")
		}
	} else {
		if apiKey == "" {
			return fmt.Errorf("未配置 OpenAI API 密钥，请先使用 'aicommit config --api-key YOUR_API_KEY' 配置")
		}
	}
	aiProvider, err := ai.NewProvider(apiKey, cfg.BaseURL, cfg.Model, language, cfg.Provider, cfg.AzureAPIVersion)
	if err != nil {
		return fmt.Errorf("创建AI提供商实例失败: %w", err)
	}

	// 生成提交消息的循环 (支持重新生成)
	for {
		fmt.Println("\n正在生成提交消息...")
		message, err := aiProvider.GenerateCommitMessage(context.Background(), commitInfo)
		if err != nil {
			return fmt.Errorf("生成提交消息失败: %w", err)
		}

		// 显示生成的消息并让用户选择操作
		action, err := interactive.ShowCommitMessage(message.Title, message.Body)
		if err != nil {
			return fmt.Errorf("交互式选择失败: %w", err)
		}

		commitMessage := message.Title
		if message.Body != "" {
			commitMessage += "\n\n" + message.Body
		}

		switch action {
		case interactive.ActionAccept:
			if err := repo.Commit(commitMessage); err != nil {
				return err
			}
			fmt.Println("✓ 已提交更改")
			return nil

		case interactive.ActionEdit:
			edited, err := interactive.EditMessage(commitMessage)
			if err != nil {
				return fmt.Errorf("编辑消息失败: %w", err)
			}
			if strings.TrimSpace(edited) == "" {
				fmt.Println("提交消息为空，操作取消")
				return nil
			}
			if err := repo.Commit(edited); err != nil {
				return err
			}
			fmt.Println("✓ 已提交更改")
			return nil

		case interactive.ActionRegenerate:
			// 继续循环，重新生成
			continue

		case interactive.ActionCancel:
			fmt.Println("提交已取消")
			return nil
		}
	}
}

// 添加版本信息处理函数
func getVersion() string {
	commitHash := commit
	if len(commit) >= 7 {
		commitHash = commit[:7]
	}
	return fmt.Sprintf("%s (%s - %s)", version, commitHash, date)
}
