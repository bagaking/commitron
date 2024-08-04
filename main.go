// Package main provides a command-line tool for generating commit comments based on diff information.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bagaking/botheater/driver/coze"
	"github.com/bagaking/easycmd"
	"github.com/urfave/cli/v2"

	"github.com/bagaking/botheater/bot"
)

const (
	maxDiffLength = 28 * 1024
	maxFileLength = 8 * 1024

	CMDNameInstallAlias = "install_alias"
	CMDNameComment      = "comment"
	CMDNameInsight      = "insight"
)

type appActions struct {
	installAlias func() error
	insight      func(string) error
	comment      func(ctx context.Context, diff, ak, sk, ep, pp string) error
}

var defaultAppActions = appActions{
	installAlias: installAlias,
	insight:      insight,
	comment:      autoComment,
}

// defaultConf is the default configuration for the bot.
var defaultConf = bot.Config{
	Endpoint:   "",
	PrefabName: "commitron",
	Prompt: &bot.Prompt{
		Content: "",
	},
}

func newAppBuilder() *easycmd.Builder {
	return newAppBuilderWithActions(defaultAppActions)
}

func newAppBuilderWithActions(actions appActions) *easycmd.Builder {
	app := easycmd.New("commitron").Set.Custom(func(command *cli.Command) {
		command.Usage = `
Commitron is an AI-powered command-line tool that automatically generates
meaningful Git commit messages based on your code changes. It analyzes 
your diff information and uses advanced language models to create concise, 
informative commit comments`
	}).End

	app.Child(CMDNameInstallAlias).Set.Usage("install the Git alias").End.Action(func(c *cli.Context) error { return actions.installAlias() })

	app.Child(CMDNameInsight).Set.Usage("insight the code changes").End.Flags(
		&cli.StringFlag{
			Name:     "committer",
			Usage:    "The committer (legacy alias: --commiter)",
			Aliases:  []string{"commiter"},
			Required: true,
		},
	).Action(func(c *cli.Context) error {
		committer := c.String("committer")
		return actions.insight(committer)
	})

	app.Child(CMDNameComment).Flags(
		&cli.StringFlag{Name: "diff", Usage: "The diff information", Aliases: []string{"d"}, Required: true},
		&cli.StringFlag{Name: "access_key", Usage: fmt.Sprintf("Access key for the API (alternative to %s)", coze.EnvKeyVOLCAccessKey), Aliases: []string{"ak"}, Required: false},
		&cli.StringFlag{Name: "secret_key", Usage: fmt.Sprintf("Secret key for the API (alternative to %s)", coze.EnvKeyVOLCSecretKey), Aliases: []string{"sk"}, Required: false},
		&cli.StringFlag{Name: "endpoint", Usage: fmt.Sprintf("Endpoint for generating the comment (alternative to  %s)", coze.EnvKeyDoubaoEndpoint), Aliases: []string{"e"}, Required: false},
		&cli.StringFlag{Name: "prompt", Usage: "Custom prompt for generating the comment", Aliases: []string{"p"}, Required: false},
	).Set.Custom(func(c *cli.Command) {
		c.Usage = fmt.Sprintf(`Generate a commit comment based on the provided diff information

Environment Variables:
   %s	Access key for the API (alternative to -ak)
   %s	Secret key for the API (alternative to -sk)
   %s	Endpoint for the API (alternative to -endpoint)

Example:
   commitron %s --access_key YOUR_ACCESS_KEY --secret_key YOUR_SECRET_KEY --endpoint YOUR_MODEL_ENDPOINT --diff \"...\"`, coze.EnvKeyVOLCAccessKey, coze.EnvKeyVOLCSecretKey, coze.EnvKeyDoubaoEndpoint, CMDNameComment)
	}).End.Action(func(c *cli.Context) error {
		diffInfo := c.String("diff")
		ak := c.String("access_key")
		sk := c.String("secret_key")
		ep := c.String("endpoint")
		pp := c.String("prompt")
		return actions.comment(c.Context, diffInfo, ak, sk, ep, pp)
	})

	return app
}

func runApp() error {
	return newAppBuilder().RunBaseAsApp()
}

func main() {
	if err := runApp(); err != nil {
		fmt.Println("=== EXECUTION FAILED===\n", err)

		os.Exit(1)
	}
}
