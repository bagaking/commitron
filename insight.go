package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
	logOutput, err := executeGitCommand("log", "--pretty=format:%H %an %cd %s", "--date=short")
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

// getUserStats 获取指定用户的代码量统计
func getUserStats(user string) (int, int, int, error) {
	logOutput, err := executeGitCommand("log", "--author="+user, "--pretty=tformat:", "--numstat")
	if err != nil {
		return 0, 0, 0, err
	}

	var totalAdded, totalRemoved, totalCommits int
	lines := strings.Split(logOutput, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		totalCommits++
		stats := strings.Fields(line)
		if len(stats) < 3 {
			continue
		}
		added, _ := strconv.Atoi(stats[0])
		removed, _ := strconv.Atoi(stats[1])
		totalAdded += added
		totalRemoved += removed
	}
	return totalAdded, totalRemoved, totalCommits, nil
}

// getCommitChanges 获取指定提交的变更行数
func getCommitChanges(commitHash string) (int, int, error) {
	logOutput, err := executeGitCommand("show", "--numstat", "--pretty=format:", commitHash)
	if err != nil {
		return 0, 0, err
	}

	var added, removed int
	lines := strings.Split(logOutput, "\n")
	for _, line := range lines {
		stats := strings.Fields(line)
		if len(stats) < 3 {
			continue
		}
		add, _ := strconv.Atoi(stats[0])
		rem, _ := strconv.Atoi(stats[1])
		added += add
		removed += rem
	}
	return added, removed, nil
}

// getUserCommitHabits 获取指定用户的提交习惯
func getUserCommitHabits(user string) (map[string][]string, error) {
	logOutput, err := executeGitCommand("log", "--author="+user, "--pretty=format:%H %cd %s", "--date=short")
	if err != nil {
		return nil, err
	}

	commitHabits := make(map[string][]string)
	lines := strings.Split(logOutput, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			continue
		}
		date, err := time.Parse("2006-01-02", parts[1])
		if err != nil {
			continue
		}
		week := date.Format("2006-W01")
		commitHabits[week] = append(commitHabits[week], line)
	}
	return commitHabits, nil
}

// insight 显示用户的提交记录和状态信息
func insight(commiter string) error {
	// 获取用户的提交记录
	commits, err := getUserCommits(commiter)
	if err != nil {
		return err
	}

	// 获取用户的代码量统计
	totalAdded, totalRemoved, totalCommits, err := getUserStats(commiter)
	if err != nil {
		return err
	}

	// 获取用户的提交习惯
	commitHabits, err := getUserCommitHabits(commiter)
	if err != nil {
		return err
	}

	// 输出统计结果和提交记录
	sb := strings.Builder{}
	sb.WriteString("[[ User status ]]\n")
	sb.WriteString(fmt.Sprintf("User %s has made %d commits:\n", commiter, len(commits)))
	sb.WriteString(fmt.Sprintf("Total lines added: %d\n", totalAdded))
	sb.WriteString(fmt.Sprintf("Total lines removed: %d\n", totalRemoved))
	sb.WriteString(fmt.Sprintf("Total commits: %d\n", totalCommits))
	sb.WriteString("[[ Commit habits (week-wise) ]]\n")
	for week, commits := range commitHabits {
		sb.WriteString(fmt.Sprintf("%s: %d commits\n", week, len(commits)))
		for _, commit := range commits {
			parts := strings.Fields(commit)
			if len(parts) < 3 {
				continue
			}
			hash := parts[0][:7]
			date := parts[1]
			message := strings.Join(parts[2:], " ")
			added, removed, err := getCommitChanges(parts[0])
			if err != nil {
				return err
			}
			sb.WriteString(fmt.Sprintf("\t%s %s add:%d rem:%d | %s\n", hash, date, added, removed, message))
		}
	}
	fmt.Println(sb.String())
	return nil
}
