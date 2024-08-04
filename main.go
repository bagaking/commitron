// Package main provides a command-line tool for generating commit comments based on diff information.
package main

import (
	"fmt"
	"github.com/bagaking/botheater/driver/coze"
	"github.com/urfave/cli/v2"
	"os"

	"github.com/bagaking/botheater/bot"
	"github.com/bagaking/easycmd"
)

const (
	maxDiffLength = 28 * 1024
	maxFileLength = 8 * 1024

	CMDNameInstallAlias = "install_alias"
	CMDNameComment      = "comment"
	CMDNameInsight      = "insight"
)

// defaultConf is the default configuration for the bot.
var defaultConf = bot.Config{
	Endpoint:   "",
	PrefabName: "commitron",
	Prompt: &bot.Prompt{
		Content: "",
	},
}

func runApp() error {
	app := easycmd.New("commitron").Set.Custom(func(command *cli.Command) {
		command.Usage = `
Commitron is an AI-powered command-line tool that automatically generates
meaningful Git commit messages based on your code changes. It analyzes 
your diff information and uses advanced language models to create concise, 
informative commit comments`
	}).End

	app.Child(CMDNameInstallAlias).Set.Usage("install the Git alias").End.Action(func(c *cli.Context) error { return installAlias() })

	app.Child(CMDNameInsight).Set.Usage("insight the code changes").End.Flags(
		&cli.StringFlag{Name: "commiter", Usage: "The commiter", Required: true},
	).Action(func(c *cli.Context) error {
		commiter := c.String("commiter")
		return insight(commiter)

	})

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

	return app.RunBaseAsApp()
}

func main() {
	if err := runApp(); err != nil {
		fmt.Println("=== EXECUTION FAILED===\n", err)

		os.Exit(1)
	}
}
