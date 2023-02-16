package cli

import (
	v2 "github.com/urfave/cli/v2"

	"github.com/cmattoon/dockerenv/cli/commands"
)

func New() *v2.App {
	app := &v2.App{
		Name:  "dockerenv",
		Usage: "extract information from docker environment variables",
		Flags: []v2.Flag{
			&v2.StringFlag{
				Name:    "container-id",
				Aliases: []string{"id", "c"},
				Value:   "",
				Usage:   "The container to extract values from",
			},
			&v2.StringFlag{
				Name:    "var-name",
				Aliases: []string{"var", "v"},
				Value:   "",
				Usage:   "The variable name to extract values from",
			},
		},
		Commands: []*v2.Command{
			commands.ListValues(),
			commands.GetValue(),
			commands.TLS(),
		},
	}

	return app
}
