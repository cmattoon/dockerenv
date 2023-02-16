package commands

import (
	"fmt"
	"log"
	"strings"
	"syscall"

	v2 "github.com/urfave/cli/v2"

	"github.com/cmattoon/dockerenv/pkg/inspector"

	"golang.org/x/crypto/ssh/terminal"
)

func ListValues() *v2.Command {
	return &v2.Command{
		Name:  "list",
		Usage: "lists environment variables",
		Action: func(c *v2.Context) error {
			var containerId string
			if containerId = c.String("container-id"); containerId == "" {
				fmt.Println("Must specify --container-id")
				return v2.ShowSubcommandHelp(c)
			}

			ins, err := inspector.New()
			if err != nil {
				log.Fatal(err)
			}

			allValues, err := ins.GetAllValues(containerId)
			if err != nil {
				log.Fatal(err)
			}

			maxlen := 0
			for k, _ := range allValues {
				if len(k) > maxlen {
					maxlen = len(k)
				}
			}

			tlen, _, _ := terminal.GetSize(syscall.Stdout)
			vlen := tlen - maxlen - 6
			for k, v := range allValues {
				s := mbsubstr(v, 0, vlen)
				if len(v) > len(s) {
					s = mbsubstr(s, 0, vlen-3) + "..."
				}
				fmt.Printf("%-*s    %s\n", maxlen, k, s)
			}
			return nil
		},
	}
}

func mbsubstr(s string, from, length int) string {
	//create array like string view
	wb := []string{}
	wb = strings.Split(s, "")

	//miss nil pointer error
	to := from + length

	if to > len(wb) {
		to = len(wb)
	}

	if from > len(wb) {
		from = len(wb)
	}
	return strings.Join(wb[from:to], "")
}
