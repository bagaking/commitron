package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/khicago/irr"
)

// installAlias installs the Git cz alias.
func installAlias() error {
	// 检查配置文件可用性
	gitConfigPath, gitConfigContent, err := testGitConfig()
	if err != nil {
		return err
	}

	fmt.Println("Installing Git Alias 'cz'.")
	fmt.Println("The alias reads credentials and endpoint from the environment used by `commitron comment`.")
	fmt.Println("Set VOLC_ACCESSKEY, VOLC_SECRETKEY, and DOUBAO_ENDPOINT in your shell or secret manager before running `git cz`.")

	// 注入 Git Alias
	aliasStr := makeAliasStr()
	// fmt.Printf("start config alias in %s\n", gitConfigPath)

	// 将 Git Alias 追加到全局 Git 配置文件
	gitConfigPath = strings.TrimSpace(gitConfigPath)

	if err = writeGitConfigWithAlias(gitConfigPath, gitConfigContent, aliasStr); err != nil {
		return err
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
		return "", nil, irr.Wrap(err, "Error reading global Git config file")
	}

	// 检查是否已经存在相同的 Git Alias
	if strings.Contains(string(gitConfigContent), "[alias]") && strings.Contains(string(gitConfigContent), "cz = ") {
		return "", nil, irr.Error("skipped: Git Alias 'cz' is already configured.")
	}

	return gitConfigPath, gitConfigContent, nil
}

func writeGitConfigWithAlias(gitConfigPath string, gitConfigContent []byte, aliasStr string) error {
	if err := os.WriteFile(gitConfigPath, append(gitConfigContent, []byte(aliasStr)...), 0o600); err != nil {
		return irr.Wrap(err, "error writing global git config file")
	}
	if err := os.Chmod(gitConfigPath, 0o600); err != nil {
		return irr.Wrap(err, "error setting global git config file permissions")
	}
	return nil
}

func makeAliasStr() string {
	return `
[alias]
    cz = "!f() { \
        if ! command -v commitron >/dev/null 2>&1; then \
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
        COMMIT_MSG_CONTENT=$(commitron comment --diff \"$COMMITRON_DIFF\"); \
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
`
}
