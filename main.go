// Package main provides a command-line tool for generating commit comments based on diff information.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/khicago/irr"

	"github.com/sirupsen/logrus"

	"github.com/bagaking/botheater/bot"
	"github.com/bagaking/botheater/driver/coze"
	"github.com/bagaking/botheater/history"
	"github.com/bagaking/botheater/utils"
)

const (
	maxDiffLength = 28 * 1024
	maxFileLength = 8 * 1024
)

// defaultConf is the default configuration for the bot.
var defaultConf = bot.Config{
	Endpoint:   "",
	PrefabName: "commitron",
	Prompt: &bot.Prompt{
		Content: "",
	},
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "install_alias" {
			installAlias()
		} else if os.Args[1] == "-h" || os.Args[1] == "--help" {
			showHelp()
		} else {
			autoComment()
			// fmt.Println("Unknown command. Use -h or --help for help.")
			// os.Exit(1)
		}
	} else {
		showHelp()
	}
}

// showHelp displays the help information for the commitron command.
func showHelp() {
	fmt.Println(`Usage: commitron [options]
Options:
  -diff        The diff information
  -ak          Access key for the API
  -sk          Secret key for the API
  -prompt      Custom prompt for generating the comment (optional)
Environment Variables:
  VOLC_ACCESSKEY     Access key for the API (alternative to -ak)
  VOLC_SECRETKEY     Secret key for the API (alternative to -sk)
  DOUBAO_ENDPOINT    Endpoint for the API (alternative to -endpoint)
Example:
  commitron -ak your_access_key -sk your_secret_key -diff \"...\"
To install the Git alias, run:
  commitron install_alias`)
}

// autoComment generates a commit comment based on the provided diff information.
func autoComment() {
	// disable logrus to hide bot debug
	logrus.SetOutput(io.Discard)

	// Define command-line flags
	diffInfo := flag.String("diff", "", "The diff information")
	accessKey := flag.String("ak", "", "Access key for the API")
	secretKey := flag.String("sk", "", "Secret key for the API")
	endpoint := flag.String("endpoint", "", "Endpoint for generating the comment")
	customPrompt := flag.String("prompt", "", "Custom prompt for generating the comment")
	flag.Parse()

	// Check if the -diff flag is provided
	if *diffInfo == "" {
		fmt.Println("Please provide the diff information using the -diff flag")
		os.Exit(1)
	}

	// Use command-line flags for access key and secret key if provided
	if *accessKey == "" {
		*accessKey = os.Getenv("VOLC_ACCESSKEY")
	}
	if *secretKey == "" {
		*secretKey = os.Getenv("VOLC_SECRETKEY")
	}
	if *endpoint == "" {
		*endpoint = os.Getenv("DOUBAO_ENDPOINT")
	}

	// Check if the access key and secret key are set
	if *accessKey == "" || *secretKey == "" {
		fmt.Println("Please provide the access key and secret key using the -ak and -sk flags or set the VOLC_ACCESSKEY and VOLC_SECRETKEY environment variables")
		os.Exit(1)
	}
	if *endpoint == "" {
		fmt.Println("Please provide the endpoint using the -endpint or set the DOUBAO_ENDPOINT environment variables")
		os.Exit(1)
	}

	// Set the prompt for the AI model
	prompt := `# Role:你是一个训练有素的代码分析员, 请根据以下的代码差异信息，生成一个简洁的提交注释

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

`

	if *customPrompt != "" {
		prompt = *customPrompt
	}

	question := buildQuestion(*diffInfo)
	// Call the SimpleQuestion function to generate the comment

	comment, err := SimpleQuestion(context.Background(), *endpoint, prompt, question)
	if err != nil {
		fmt.Printf("Error generating comment: %v\n", err)
		os.Exit(1)
	}

	// Print the generated comment
	fmt.Println(comment)
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

// installAlias installs the Git commit-msg hook.
func installAlias() {
	// 检查配置文件可用性
	gitConfigPath, gitConfigContent, err := testGitConfig()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	// Ask user if they want to specify ak and sk
	fmt.Print("Do you want to specify access key and secret key? (y/n): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	var accessKey, secretKey string
	if response == "y" {
		fmt.Print("Enter Access Key: ")
		accessKey, _ = reader.ReadString('\n')
		accessKey = strings.TrimSpace(accessKey)

		fmt.Print("Enter Secret Key: ")
		secretKey, _ = reader.ReadString('\n')
		secretKey = strings.TrimSpace(secretKey)
	}

	fmt.Print("Do you want to specify endpoint? (y/n): ")
	response, _ = reader.ReadString('\n')
	response = strings.TrimSpace(response)

	var endpointStr string
	if response == "y" {
		fmt.Print("Enter Endpoint: ")
		endpointStr, _ = reader.ReadString('\n')
		endpointStr = strings.TrimSpace(endpointStr)
	}

	//// 获取当前工作目录
	//wd, err := os.Getwd()
	//if err != nil {
	//	fmt.Printf("Error getting current working directory: %v\n", err)
	//	os.Exit(1)
	//}

	//// 构建 commit-msg 钩子的路径
	//hookPath := filepath.Join(wd, ".git", "hooks", "commit-msg")

	commentStr := "commitron"
	if accessKey != "" {
		commentStr += fmt.Sprintf(" -ak %s ", accessKey)
	}
	if secretKey != "" {
		commentStr += fmt.Sprintf(" -sk %s", secretKey)
	}
	if endpointStr != "" {
		commentStr += fmt.Sprintf(" -endpoint %s ", endpointStr)
	}

	// 注入 Git Alias
	aliasStr := makeAliasStr(commentStr)
	// fmt.Printf("start config alias in %s\n", gitConfigPath)

	// 将 Git Alias 追加到全局 Git 配置文件

	err = os.WriteFile(strings.TrimSpace(gitConfigPath),
		append(gitConfigContent, []byte(aliasStr)...), 0o644)
	if err != nil {
		fmt.Printf("Error writing global Git config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("success: Git Alias 'cz' has been configured.")
}

func testGitConfig() (gitConfigPath string, gitConfigContent []byte, err error) {
	// git config --global --list --show-origin
	gitGlobalConfig, err := exec.Command("git", "config", "--global", "--list", "--show-origin").Output()
	if err != nil {
		return "", nil, irr.Wrap(err, "Cannot getting global Git config file path\n")
	}

	// 从输出中提取全局 Git 配置文件路径
	re := regexp.MustCompile(`file:(\S+)`)
	match := re.FindStringSubmatch(string(gitGlobalConfig))
	if len(match) < 2 {
		return "", nil, irr.Error("Global Git config file path not found.")
	}

	if gitConfigPath = strings.TrimSpace(match[1]); gitConfigPath == "" {
		return "", nil, irr.Error("Global Git config file path not found.")
	}

	// 读取现有的全局 Git 配置
	if gitConfigContent, err = os.ReadFile(strings.TrimSpace(gitConfigPath)); err != nil {
		return "", nil, irr.Wrap(err, "Error reading global Git config file: %v")
	}

	// 检查是否已经存在相同的 Git Alias
	if strings.Contains(string(gitConfigContent), "[alias]") && strings.Contains(string(gitConfigContent), "cz = ") {
		return "", nil, irr.Error("skipped: Git Alias 'cz' is already configured.")
	}

	return gitConfigPath, gitConfigContent, nil
}

func makeAliasStr(commitronCmd string) string {
	return fmt.Sprintf(`
[alias]
    cz = "!f() { \
        if [ -z \"$(which commitron)\" ]; then \
            echo 'commitron could not be found. Please install it by running:'; \
            echo 'go install github.com/bagaking/commitron@latest'; \
            exit 1; \
        fi; \
        COMMITRON_DIFF=$(git diff --cached); \
        show_animation() { \
            while true; do \
                for s in '/' '-' '\\\\' '|'; do \
                    printf '\\r> [Commitron] generating commit message... %%s' \"$s\"; \
                    sleep 0.2; \
                done; \
            done; \
        }; \
        show_animation & \
        animation_pid=$!; \
        COMMIT_MSG_CONTENT=$(%s -diff \"$COMMITRON_DIFF\"); \
        COMMITRON_EXIT_CODE=$?; \
        kill $animation_pid > /dev/null 2>&1; \
        wait $animation_pid 2>/dev/null; \
        printf '\\r> [Commitron] generating commit message... [done]\\n'; \
        if [ $COMMITRON_EXIT_CODE -ne 0 ]; then \
		  printf '\\rFailed to generate commit message. Aborting commit.\\n'; \
          echo 'Error output from commitron:'; \
          echo \"$COMMIT_MSG_CONTENT\"; \
          exit 1; \
        fi; \
        git commit -e -m \"$COMMIT_MSG_CONTENT\"; \
    }; f"
`, commitronCmd)
}
