package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/khicago/got/util/typer"
	"github.com/khicago/irr"

	"github.com/bagaking/botheater/bot"
	"github.com/bagaking/botheater/driver/coze"
	"github.com/bagaking/botheater/history"
	"github.com/bagaking/botheater/utils"
)

// autoComment generates a commit comment based on the provided diff information.
func autoComment(ctx context.Context, diff, ak, sk, ep, pp string) error {
	// disable logrus to hide bot debug
	logrus.SetOutput(io.Discard)

	// Check if the -diff flag is provided
	if diff == "" {
		return irr.Error("Please provide the diff information using the --diff (or -d) flag")
	}

	// Use command-line flags for access key and secret key if provided
	ak = typer.Or(ak, coze.EnvKeyVOLCAccessKey.Read())
	sk = typer.Or(sk, coze.EnvKeyVOLCSecretKey.Read())
	ep = typer.Or(ep, coze.EnvKeyDoubaoEndpoint.Read())

	// Check if the access key and secret key are set
	if ak == "" || sk == "" {
		return irr.Error("Please provide the access key and secret key using flags or environment variables")
	}
	if ep == "" {
		return irr.Error("Please provide the endpoint using flags or environment variables")
	}

	// Set the prompt for the AI model
	prompt := typer.Or(pp, `# Role:你是一个训练有素的代码分析员, 请根据以下的代码差异信息，生成一个简洁的提交注释

# Constrains
- 语言简洁, 用英文输出
- 不输出 commit message 之外的任何内容
- 遵循 git commit message 的格式标准
  - Commit Message 包括必填的 Header 和可以不写的 Body 和 Footer 三部分, Header 的格式为 type (scope): subject, scope 可选
  - type 是主要的变更类型, 包含 feat(有新功能), refactor(大型重构), fix(修复), test(加测试), docs(修改文档), style(改代码格式),  perf(性能优化), build(构建), ci(持续集成), chore(非关键修改), revert(撤销)
  - scope 是变更范围, 有多个范围时用 (a,b) 分隔列举, 或者 * 代替
  - subject 是总结性质的一句话, 消息开头, 皆为不需要句号
  - body 是详细描述, 可以包含多行, 用于解释变更的原因和内容
  - footer 是备注, 如果有不兼容变更, 可以以 BREAKING CHANGE 开头, 后面是描述具体变更内容, 原因, 迁移/观测/回滚的方法;

# Example
feat (commitron): Add Git commit-msg hook installation

- Implement installAlias() function to install commitron as a Git commit-msg hook
- Check for existing commit-msg hook and append commitron hook if necessary
- Create a new commit-msg hook file with commitron hook if it doesn't exist
- Allow specifying access key, secret key, and endpoint during hook installation
`)

	comment, err := SimpleQuestion(context.Background(), ep, prompt, buildQuestion(diff))
	if err != nil {
		return irr.Wrap(err, "failed to generate comment")
	}

	// Print the generated comment
	fmt.Println(comment)
	return nil
}

func buildQuestion(diffInfo string) string {
	diffInfo = "DiffInfo 如下:\n" + diffInfo
	// 计算 diff 信息的总字数
	if utils.CountTokens(diffInfo) < maxDiffLength {
		return diffInfo
	}

	// 将 diff 信息按文件切分
	files := strings.Split(diffInfo, "diff --git")

	// 筛选后的 diff 信息
	var filteredDiff []string

	// 遍历每个文件
	for _, file := range files {
		// 计算文件的字数
		fileLength := len(file)

		// 如果单个文件变更不超过 4k,则保留
		if fileLength <= maxFileLength {
			filteredDiff = append(filteredDiff, file)
		} else {
			// 如果单个文件变更超过 4k,则只保留文件的元信息和修改信息
			lines := strings.Split(file, "\n")
			filteredLines := []string{"\n--Large Diff Start --"}
			for _, line := range lines {
				if strings.HasPrefix(line, "--- a/") || strings.HasPrefix(line, "+++ b/") ||
					strings.HasPrefix(line, "@@ ") {
					filteredLines = append(filteredLines, line)
				}
			}
			filteredLines = append(filteredLines, fmt.Sprintf("--Large Diff End -- (以上修改内容超过 %d, 已省略)。\n", maxFileLength))
			filteredFile := strings.Join(filteredLines, "\n")
			filteredDiff = append(filteredDiff, filteredFile)
		}
	}

	// 将筛选后的文件重新拼接成完整的 diff 信息
	diffInfo = strings.Join(filteredDiff, "diff --git")

	// 如果筛选后的 diff 信息仍然超过 32k,则进行截断
	if utils.CountTokens(diffInfo) > maxDiffLength {
		return string([]rune(diffInfo)[:maxDiffLength])
	}
	return diffInfo
}

// SimpleQuestion sends a question to the bot and returns the generated answer.
//
// ctx: The context for the request.
// endpoint: The endpoint of the bot service.
// prompt: The prompt for generating the answer.
// question: The question to be answered.
//
// Returns the generated answer and any error encountered.
func SimpleQuestion(ctx context.Context, endpoint, prompt, question string) (string, error) {
	driver := coze.New(coze.NewClient(ctx), endpoint)
	conf := defaultConf
	conf.Prompt.Content = prompt
	theBot := bot.New(conf, driver, nil)

	return theBot.Question(ctx, history.NewHistory(), question)
}
