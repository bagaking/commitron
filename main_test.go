package main

import (
	"strings"
	"testing"

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
