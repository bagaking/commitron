package main

import (
	"context"
	"strings"
	"testing"

	"github.com/bagaking/easycmd"
	"github.com/urfave/cli/v2"
)

func TestInsightFlagSupportsCommitterAndLegacyAlias(t *testing.T) {
	var insightCmd *cli.Command
	for _, cmd := range newAppBuilder().BuildBase().Subcommands {
		if cmd.Name == CMDNameInsight {
			insightCmd = cmd
			break
		}
	}
	if insightCmd == nil {
		t.Fatal("insight command is not registered")
	}
	if len(insightCmd.Flags) != 1 {
		t.Fatalf("insight command has %d flags, want 1", len(insightCmd.Flags))
	}

	flag, ok := insightCmd.Flags[0].(*cli.StringFlag)
	if !ok {
		t.Fatalf("insight flag has type %T, want *cli.StringFlag", insightCmd.Flags[0])
	}
	if flag.Name != "committer" {
		t.Fatalf("insight flag name = %q, want %q", flag.Name, "committer")
	}
	if len(flag.Aliases) != 1 || flag.Aliases[0] != "commiter" {
		t.Fatalf("insight flag aliases = %v, want [commiter]", flag.Aliases)
	}
	if !flag.Required {
		t.Fatal("insight committer flag is not required")
	}
	if !strings.Contains(flag.Usage, "--commiter") {
		t.Fatalf("insight flag usage = %q, want legacy alias mention", flag.Usage)
	}
}

func TestInsightCommandPassesCommitterToAction(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "canonical flag",
			args: []string{"commitron", CMDNameInsight, "--committer", "Test Author"},
			want: "Test Author",
		},
		{
			name: "legacy flag alias",
			args: []string{"commitron", CMDNameInsight, "--commiter", "Legacy Author"},
			want: "Legacy Author",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			actions := stubAppActions(t)
			actions.insight = func(committer string) error {
				got = committer
				return nil
			}

			err := runAppBuilderForTest(t, newAppBuilderWithActions(actions), tt.args)
			if err != nil {
				t.Fatalf("commitron insight action path error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("commitron insight action received committer = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommentCommandPassesFlagsToAction(t *testing.T) {
	var gotDiff, gotAccessKey, gotSecretKey, gotEndpoint, gotPrompt string
	actions := stubAppActions(t)
	actions.comment = func(ctx context.Context, diff, ak, sk, ep, pp string) error {
		if ctx == nil {
			t.Error("commitron comment action context = nil, want non-nil context")
		}
		gotDiff = diff
		gotAccessKey = ak
		gotSecretKey = sk
		gotEndpoint = ep
		gotPrompt = pp
		return nil
	}

	args := []string{
		"commitron",
		CMDNameComment,
		"--diff", "diff --git a/a.txt b/a.txt",
		"--ak", "test-ak",
		"--sk", "test-sk",
		"--endpoint", "test-endpoint",
		"--prompt", "test prompt",
	}
	err := runAppBuilderForTest(t, newAppBuilderWithActions(actions), args)
	if err != nil {
		t.Fatalf("commitron comment action path error = %v, want nil", err)
	}

	want := map[string]string{
		"diff":     "diff --git a/a.txt b/a.txt",
		"ak":       "test-ak",
		"sk":       "test-sk",
		"endpoint": "test-endpoint",
		"prompt":   "test prompt",
	}
	got := map[string]string{
		"diff":     gotDiff,
		"ak":       gotAccessKey,
		"sk":       gotSecretKey,
		"endpoint": gotEndpoint,
		"prompt":   gotPrompt,
	}
	for field, wantValue := range want {
		if got[field] != wantValue {
			t.Errorf("commitron comment action received %s = %q, want %q", field, got[field], wantValue)
		}
	}
}

func runAppBuilderForTest(t *testing.T, builder interface{ BuildBase() *cli.Command }, args []string) error {
	t.Helper()

	oldHelpPrinter := cli.HelpPrinter
	t.Cleanup(func() {
		cli.HelpPrinter = oldHelpPrinter
	})

	app, err := easycmd.ToApp(builder.BuildBase())
	if err != nil {
		t.Fatalf("easycmd.ToApp() error = %v", err)
	}
	return app.Run(args)
}

func stubAppActions(t *testing.T) appActions {
	t.Helper()

	return appActions{
		installAlias: func() error {
			t.Fatal("installAlias action called unexpectedly")
			return nil
		},
		insight: func(committer string) error {
			t.Fatalf("insight action called unexpectedly with committer %q", committer)
			return nil
		},
		comment: func(ctx context.Context, diff, ak, sk, ep, pp string) error {
			t.Fatalf("comment action called unexpectedly with diff %q, ak %q, sk %q, endpoint %q, prompt %q", diff, ak, sk, ep, pp)
			return nil
		},
	}
}
