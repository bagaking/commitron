// Package main provides a command-line tool for generating commit comments based on diff information.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bagaking/botheater/driver/coze"
	"github.com/khicago/irr"
	"github.com/urfave/cli/v2"

	"github.com/bagaking/botheater/bot"
	"github.com/bagaking/easycmd"
)

const (
	maxDiffLength = 28 * 1024
	maxFileLength = 8 * 1024

	CMDNameComment = "comment"
)

// defaultConf is the default configuration for the bot.
var defaultConf = bot.Config{
	Endpoint:   "",
	PrefabName: "commitron",
	Prompt: &bot.Prompt{
		Content: "",
	},
}

func runApp() {
	app := easycmd.New("commitron").Set.Custom(func(command *cli.Command) {
		command.Usage = `
Commitron is an AI-powered command-line tool that automatically generates
meaningful Git commit messages based on your code changes. It analyzes 
your diff information and uses advanced language models to create concise, 
informative commit comments`
	}).End
	app.Child("install_alias").Set.Usage("install the Git alias").End.Action(func(c *cli.Context) error { return installAlias() })
	app.Child(CMDNameComment).Flags(
		&cli.StringFlag{Name: "diff", Usage: "The diff information", Aliases: []string{"d"}, Required: true},
		&cli.StringFlag{Name: "access_key", Usage: fmt.Sprintf("Access key for the API (alternative to %s)", coze.EnvKeyVOLCAccessKey), Aliases: []string{"ak"}, Required: false},
		&cli.StringFlag{Name: "secret_key", Usage: fmt.Sprintf("Secret key for the API (alternative to %s)", coze.EnvKeyVOLCAccessKey), Aliases: []string{"sk"}, Required: false},
		&cli.StringFlag{Name: "endpoint", Usage: fmt.Sprintf("Endpoint for generating the comment (alternative to  %s)", coze.EnvKeyDoubaoEndpoint), Aliases: []string{"e"}, Required: false},
		&cli.StringFlag{Name: "prompt", Usage: "Custom prompt for generating the comment", Aliases: []string{"p"}, Required: false},
	).Set.Custom(func(c *cli.Command) {
		c.Usage = fmt.Sprintf(`Generate a commit comment based on the provided diff information

Environment Variables:
   %s	Access key for the API (alternative to -ak)
   %s	Secret key for the API (alternative to -sk)
   %s	Endpoint for the API (alternative to -endpoint)

Example:
   commitron %s -ak your_access_key -sk your_secret_key -diff \"...\"`, coze.EnvKeyVOLCAccessKey, coze.EnvKeyVOLCAccessKey, coze.EnvKeyDoubaoEndpoint, CMDNameComment)
	}).End.Action(func(c *cli.Context) error {
		diffInfo := c.String("diff")
		ak := c.String("access_key")
		sk := c.String("secret_key")
		ep := c.String("endpoint")
		pp := c.String("prompt")
		return autoComment(c.Context, diffInfo, ak, sk, ep, pp)
	})
	if err := app.RunBaseAsApp(); err != nil {
		fmt.Println("=== EXECUTION FAILED===\n", err)

		os.Exit(1)
	}
}

func main() {
	runApp()
}

// installAlias installs the Git commit-msg hook.
func installAlias() error {
	// 检查配置文件可用性
	gitConfigPath, gitConfigContent, err := testGitConfig()
	if err != nil {
		return err
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
		fmt.Print("Enter Doubao Endpoint: ")
		endpointStr, _ = reader.ReadString('\n')
		endpointStr = strings.TrimSpace(endpointStr)
	}

	commentStr := "commitron " + CMDNameComment
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
		return irr.Wrap(err, "error writing global git config file")
	}

	fmt.Println("success: Git Alias 'cz' has been configured.")
	return nil
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
        COMMIT_MSG_CONTENT=$(%s --diff \"$COMMITRON_DIFF\"); \
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
