package commands

import (
	"fmt"
	"log"

	v2 "github.com/urfave/cli/v2"

	"github.com/cmattoon/dockerenv/pkg/inspector"
)

func GetValue() *v2.Command {
	return &v2.Command{
		Name:  "get",
		Usage: "returns a plain value suitable for scripting",
		Action: func(c *v2.Context) error {
			var containerId, varName string

			if containerId = c.String("container-id"); containerId == "" {
				fmt.Println("Must specify --container-id")
				return v2.ShowSubcommandHelp(c)
			}

			if varName = c.String("var"); varName == "" {
				fmt.Println("Must specify --var")
				return v2.ShowSubcommandHelp(c)
			}

			ins, err := inspector.New()
			if err != nil {
				log.Fatal(err)
			}

			val, err := ins.GetValue(containerId, varName)
			if err != nil {
				log.Fatalf("unable to get value: %s", err)
			}

			if val == "" {
				fmt.Println("<empty>")
				return nil
			}

			fmt.Printf("%s\n", val)
			return nil
		},
	}
}
