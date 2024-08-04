package main

import (
	"strings"
	"testing"
)

func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxRunes int
		want     string
	}{
		{
			name:     "shorter than limit",
			input:    "hello",
			maxRunes: 8,
			want:     "hello",
		},
		{
			name:     "unicode boundary",
			input:    "ab世界cd",
			maxRunes: 4,
			want:     "ab世界",
		},
		{
			name:     "zero limit",
			input:    "hello",
			maxRunes: 0,
			want:     "",
		},
		{
			name:     "negative limit",
			input:    "hello",
			maxRunes: -1,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateRunes(tt.input, tt.maxRunes)
			if got != tt.want {
				t.Errorf("truncateRunes(%q, %d) = %q, want %q", tt.input, tt.maxRunes, got, tt.want)
			}
		})
	}
}

func TestBuildQuestionLargeDiffDoesNotPanicWhenFilteredDiffIsShort(t *testing.T) {
	largeFile := strings.Repeat("x", maxDiffLength+1)
	diff := "diff --git a/large.txt b/large.txt\n" +
		"--- a/large.txt\n" +
		"+++ b/large.txt\n" +
		"@@ -1 +1 @@\n" +
		largeFile

	got := buildQuestion(diff)
	if got == "" {
		t.Errorf("buildQuestion(%q) = empty string, want summarized diff", diff[:64])
	}
	if !strings.Contains(got, "--Large Diff Start --") {
		t.Errorf("buildQuestion(%q) = %q, want large diff marker", diff[:64], got)
	}
}

func TestBuildQuestionTruncatesFilteredDiffWhenSummaryStillTooLarge(t *testing.T) {
	var diff strings.Builder
	largeFile := strings.Repeat("x", maxFileLength+1)
	for i := 0; i < 360; i++ {
		diff.WriteString("diff --git a/file")
		diff.WriteString(string(rune('a' + i%26)))
		diff.WriteString(".txt b/file")
		diff.WriteString(string(rune('a' + i%26)))
		diff.WriteString(".txt\n")
		diff.WriteString("--- a/file.txt\n")
		diff.WriteString("+++ b/file.txt\n")
		diff.WriteString("@@ -1 +1 @@\n")
		diff.WriteString(largeFile)
		diff.WriteString("\n")
	}

	got := buildQuestion(diff.String())
	if got == "" {
		t.Fatal("buildQuestion returned empty string, want truncated summary")
	}
	if len([]rune(got)) != maxDiffLength {
		t.Fatalf("buildQuestion returned %d runes, want %d", len([]rune(got)), maxDiffLength)
	}
	if !strings.Contains(got, "--Large Diff Start --") {
		t.Fatalf("buildQuestion output does not include large diff marker")
	}
}
