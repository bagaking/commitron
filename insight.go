package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type (
	gitStatisticGroup map[string]int
)

type UserStats struct {
	TotalAdded, TotalRemoved, TotalCommits          int
	FileTypeChanges, FileTypeAdded, FileTypeRemoved gitStatisticGroup
	FileChanges, FileAdded, FileRemoved             gitStatisticGroup
	DirChanges, DirAdded, DirRemoved                gitStatisticGroup
}

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
	logOutput, err := executeGitCommand("log", "--author="+user, "--pretty=format:%H %an %cd %s", "--date=short")
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

// getFileExtension 获取文件扩展名
func getFileExtension(file string) string {
	parts := strings.Split(file, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// getDirectory 获取文件目录
func getDirectory(file string) string {
	parts := strings.Split(file, "/")
	if len(parts) > 1 {
		return parts[0]
	}
	return "root"
}

// getUserStats 获取指定用户的代码量统计
func getUserStats(user string) (UserStats, error) {
	logOutput, err := executeGitCommand("log", "--author="+user, "--pretty=tformat:", "--numstat")
	if err != nil {
		return UserStats{}, err
	}

	stats := UserStats{
		FileTypeChanges: make(gitStatisticGroup),
		FileTypeAdded:   make(gitStatisticGroup),
		FileTypeRemoved: make(gitStatisticGroup),
		FileChanges:     make(gitStatisticGroup),
		FileAdded:       make(gitStatisticGroup),
		FileRemoved:     make(gitStatisticGroup),
		DirChanges:      make(gitStatisticGroup),
		DirAdded:        make(gitStatisticGroup),
		DirRemoved:      make(gitStatisticGroup),
	}

	lines := strings.Split(logOutput, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		stats.TotalCommits++
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		added, _ := strconv.Atoi(fields[0])
		removed, _ := strconv.Atoi(fields[1])
		file := fields[2]
		ext := getFileExtension(file)
		dir := getDirectory(file)

		stats.TotalAdded += added
		stats.TotalRemoved += removed
		stats.FileTypeChanges[ext]++
		stats.FileTypeAdded[ext] += added
		stats.FileTypeRemoved[ext] += removed
		stats.FileChanges[file]++
		stats.FileAdded[file] += added
		stats.FileRemoved[file] += removed
		stats.DirChanges[dir]++
		stats.DirAdded[dir] += added
		stats.DirRemoved[dir] += removed
	}
	return stats, nil
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

// getTopCommits 获取指定用户代码量最大的前 10 次提交
func getTopCommits(user string) ([]string, error) {
	logOutput, err := executeGitCommand("log", "--author="+user, "--pretty=format:%H %cd %s", "--date=short")
	if err != nil {
		return nil, err
	}

	type commitInfo struct {
		hash    string
		date    string
		message string
		added   int
		removed int
	}

	var commits []commitInfo
	lines := strings.Split(logOutput, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			continue
		}
		hash := parts[0]
		date := parts[1]
		message := parts[2]
		added, removed, err := getCommitChanges(hash)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commitInfo{hash, date, message, added, removed})
	}

	// 按变更行数排序
	sort.Slice(commits, func(i, j int) bool {
		return (commits[i].added + commits[i].removed) > (commits[j].added + commits[j].removed)
	})

	// 取前 10 个提交
	topCommits := commits
	if len(commits) > 10 {
		topCommits = commits[:10]
	}

	var result []string
	for _, commit := range topCommits {
		result = append(result, fmt.Sprintf("%s %s [add:%d rem:%d] %s", commit.hash[:7], commit.date, commit.added, commit.removed, commit.message))
	}
	return result, nil
}

// getTopFiles 获取变更最多的前 10 个文件
func getTopFiles(fileChanges, fileAdded, fileRemoved gitStatisticGroup) []string {
	type fileInfo struct {
		file    string
		changes int
		added   int
		removed int
	}

	var files []fileInfo
	for file, changes := range fileChanges {
		files = append(files, fileInfo{file, changes, fileAdded[file], fileRemoved[file]})
	}

	// 按变更次数排序
	sort.Slice(files, func(i, j int) bool {
		return files[i].changes > files[j].changes
	})

	// 取前 10 个文件
	topFiles := files
	if len(files) > 10 {
		topFiles = files[:10]
	}

	var result []string
	for _, file := range topFiles {
		result = append(result, fmt.Sprintf("%s [changes:%d add:%d rem:%d]", file.file, file.changes, file.added, file.removed))
	}
	return result
}

// getTopDirectories 获取变更最多的前三个目录
func getTopDirectories(dirChanges gitStatisticGroup, dirAdded gitStatisticGroup, dirRemoved gitStatisticGroup) []string {
	type dirInfo struct {
		dir     string
		changes int
		added   int
		removed int
	}

	var dirs []dirInfo
	for dir, changes := range dirChanges {
		dirs = append(dirs, dirInfo{dir, changes, dirAdded[dir], dirRemoved[dir]})
	}

	// 按变更次数排序
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].changes > dirs[j].changes
	})

	// 取前三个目录
	topDirs := dirs
	if len(dirs) > 3 {
		topDirs = dirs[:3]
	}

	var result []string
	for _, dir := range topDirs {
		result = append(result, fmt.Sprintf("%s [changes:%d add:%d rem:%d]", dir.dir, dir.changes, dir.added, dir.removed))
	}
	return result
}

// getMainBranchCommits 获取主干分支的提交记录
func getMainBranchCommits(user string) ([]string, error) {
	logOutput, err := executeGitCommand("log", "main", "--author="+user, "--pretty=format:%H")
	if err != nil {
		logOutput, err = executeGitCommand("log", "master", "--author="+user, "--pretty=format:%H")
		if err != nil {
			return nil, err
		}
	}

	commits := strings.Split(logOutput, "\n")
	return commits, nil
}

// insight 显示用户的提交记录和状态信息
func insight(commiter string) error {
	// 获取用户的提交记录
	commits, err := getUserCommits(commiter)
	if err != nil {
		return err
	}

	// 获取用户的代码量统计
	stats, err := getUserStats(commiter)
	if err != nil {
		return err
	}

	// 获取用户的提交习惯
	commitHabits, err := getUserCommitHabits(commiter)
	if err != nil {
		return err
	}

	// 获取用户代码量最大的前 10 次提交
	topCommits, err := getTopCommits(commiter)
	if err != nil {
		return err
	}

	// 获取变更最多的前 10 个文件
	topFiles := getTopFiles(stats.FileChanges, stats.FileAdded, stats.FileRemoved)

	// 获取变更最多的前三个目录
	topDirs := getTopDirectories(stats.DirChanges, stats.DirAdded, stats.DirRemoved)

	// 获取主干分支的提交记录
	mainBranchCommits, err := getMainBranchCommits(commiter)
	if err != nil {
		fmt.Printf("Warning: error getting main branch commits: %v\n", err)
		// return err
	}
	mainBranchCommitsSet := make(map[string]struct{})
	for _, commit := range mainBranchCommits {
		mainBranchCommitsSet[commit] = struct{}{}
	}

	// 区分在主干和不在主干的提交
	var mainCommits, nonMainCommits []string
	for _, commit := range commits {
		commitHash := strings.Fields(commit)[0]
		if _, exists := mainBranchCommitsSet[commitHash]; exists {
			mainCommits = append(mainCommits, commit)
		} else {
			nonMainCommits = append(nonMainCommits, commit)
		}
	}

	// 输出统计结果和提交记录
	sb := strings.Builder{}
	sb.WriteString("\n\033[1;34m# User status \033[0m\n")
	sb.WriteString("\n\033[1;34m## Overall \033[0m\n\n")
	sb.WriteString(fmt.Sprintf("User %s has made \033[1;32m%d\033[0m commits\n\n", commiter, len(commits)))
	sb.WriteString(fmt.Sprintf("- Main branch commits: \033[1;32m%d\033[0m\n", len(mainCommits)))
	sb.WriteString(fmt.Sprintf("- Non-main branch commits: \033[1;31m%d\033[0m\n", len(nonMainCommits)))
	sb.WriteString(fmt.Sprintf("- Total lines added: \033[1;32m%d\033[0m\n", stats.TotalAdded))
	sb.WriteString(fmt.Sprintf("- Total lines removed: \033[1;31m%d\033[0m\n", stats.TotalRemoved))
	sb.WriteString("\n\033[1;34m## File changes by type \033[0m\n\n")
	for ext, count := range stats.FileTypeChanges {
		sb.WriteString(fmt.Sprintf("- *.%s: %d files [add:%d rem:%d]\n", ext, count, stats.FileTypeAdded[ext], stats.FileTypeRemoved[ext]))
	}
	sb.WriteString("\n\033[1;34m## Top files \033[0m\n\n")
	for _, file := range topFiles {
		sb.WriteString(fmt.Sprintf("- %s\n", file))
	}
	sb.WriteString("\n\033[1;34m## Top directories \033[0m\n\n")
	for _, dir := range topDirs {
		sb.WriteString(fmt.Sprintf("- %s\n", dir))
	}
	sb.WriteString("\n\033[1;34m## Top commits \033[0m\n\n")
	for _, commit := range topCommits {
		sb.WriteString(fmt.Sprintf("- %s\n", commit))
	}
	sb.WriteString("\n\033[1;34m## Commit habits (week-wise) \033[0m\n\n")
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
			sb.WriteString(fmt.Sprintf("\t%s %s [add:%d rem:%d] %s\n", hash, date, added, removed, message))
		}
	}
	fmt.Println(sb.String())
	return nil
}
