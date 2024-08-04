package main

import (
	"context"
	"strings"
	"testing"

	"github.com/bagaking/botheater/driver/coze"
)

func TestAutoCommentRejectsInvalidInputBeforeModelCall(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		ak       string
		sk       string
		ep       string
		wantErr  string
		clearEnv bool
	}{
		{
			name:    "empty diff",
			ak:      "test-access-key",
			sk:      "test-secret-key",
			ep:      "test-endpoint",
			wantErr: "Please provide the diff information",
		},
		{
			name:     "missing credentials",
			diff:     "diff --git a/a.txt b/a.txt\n",
			ep:       "test-endpoint",
			wantErr:  "Please provide the access key and secret key",
			clearEnv: true,
		},
		{
			name:     "missing endpoint",
			diff:     "diff --git a/a.txt b/a.txt\n",
			ak:       "test-access-key",
			sk:       "test-secret-key",
			wantErr:  "Please provide the endpoint",
			clearEnv: true,
		},
	}

	ask := func(context.Context, string, string, string) (string, error) {
		t.Fatal("autoComment called the model before validating required input")
		return "", nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.clearEnv {
				t.Setenv("VOLC_ACCESSKEY", "")
				t.Setenv("VOLC_SECRETKEY", "")
				t.Setenv("DOUBAO_ENDPOINT", "")
			}

			err := autoCommentWithAsk(context.Background(), tt.diff, tt.ak, tt.sk, tt.ep, "", ask)
			if err == nil {
				t.Fatal("autoComment() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("autoComment() error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestAutoCommentPassesCallerContextToModel(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "caller-context")

	var gotContextValue interface{}
	ask := func(ctx context.Context, endpoint, prompt, question string) (string, error) {
		gotContextValue = ctx.Value(contextKey{})
		if endpoint != "test-endpoint" {
			t.Fatalf("endpoint = %q, want %q", endpoint, "test-endpoint")
		}
		if prompt == "" {
			t.Fatal("prompt is empty")
		}
		if !strings.Contains(question, "DiffInfo") {
			t.Fatalf("question = %q, want built diff question", question)
		}
		return "fix(test): keep caller context", nil
	}

	err := autoCommentWithAsk(ctx, "diff --git a/a.txt b/a.txt\n", "test-access-key", "test-secret-key", "test-endpoint", "custom prompt", ask)
	if err != nil {
		t.Fatalf("autoCommentWithAsk returned error: %v", err)
	}
	if gotContextValue != "caller-context" {
		t.Fatalf("model context value = %v, want caller context value", gotContextValue)
	}
}

func TestAutoCommentUsesResolvedFlags(t *testing.T) {
	coze.VOLC_ACCESSKEY = "previous-access-key"
	coze.VOLC_SECRETKEY = "previous-sk"

	t.Setenv("VOLC_ACCESSKEY", "env-access-key")
	t.Setenv("VOLC_SECRETKEY", "env-sk")
	t.Setenv("DOUBAO_ENDPOINT", "env-endpoint")

	ask := func(_ context.Context, endpoint, _, _ string) (string, error) {
		if coze.VOLC_ACCESSKEY != "flag-access-key" {
			t.Errorf("resolved access key = %q, want %q", coze.VOLC_ACCESSKEY, "flag-access-key")
		}
		if coze.VOLC_SECRETKEY != "flag-sk" {
			t.Errorf("resolved secret key = %q, want %q", coze.VOLC_SECRETKEY, "flag-sk")
		}
		if endpoint != "flag-endpoint" {
			t.Errorf("resolved endpoint = %q, want %q", endpoint, "flag-endpoint")
		}
		return "fix(comment): use flag credentials", nil
	}

	err := autoCommentWithAsk(context.Background(), "diff --git a/a.txt b/a.txt\n", "flag-access-key", "flag-sk", "flag-endpoint", "", ask)
	if err != nil {
		t.Fatalf("autoCommentWithAsk() error = %v, want nil", err)
	}
	if coze.VOLC_ACCESSKEY != "previous-access-key" {
		t.Fatalf("access key after autoCommentWithAsk = %q, want restored previous value", coze.VOLC_ACCESSKEY)
	}
	if coze.VOLC_SECRETKEY != "previous-sk" {
		t.Fatalf("secret key after autoCommentWithAsk = %q, want restored previous value", coze.VOLC_SECRETKEY)
	}
}

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
