package main

import (
	"bytes"
	"os/exec"
	"strings"
)

// executeGitCommand 执行 Git 命令并返回输出
func executeGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

// getUserCommits 获取指定用户的提交记录
func getUserCommits(user string) ([]string, error) {
	logOutput, err := executeGitCommand("log", "--pretty=format:%H %an %s")
	if err != nil {
		return nil, err
	}

	var userCommits []string
	commits := strings.Split(logOutput, "\n")
	for _, commit := range commits {
		if strings.Contains(commit, user) {
			userCommits = append(userCommits, commit)
		}
	}
	return userCommits, nil
}
