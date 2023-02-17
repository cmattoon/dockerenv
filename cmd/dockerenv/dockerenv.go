package main

import (
	"context"
	"fmt"
	//"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	//"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	//"github.com/docker/docker/pkg/stdcopy"

	"github.com/cmattoon/dockerenv/cli"
)

func main() {
	app := cli.New()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func old_stuff() {
	containerID := os.Args[1]
	envVarName := os.Args[2]

	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create client: %s", err.Error())
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Fatalf("Failed to list containers: %s", err.Error())
	}

	for _, c := range containers {
		if strings.HasPrefix(c.ID, containerID) {
			info, err := cli.ContainerInspect(ctx, c.ID)
			if err != nil {
				log.Fatalf("Unable to inspect container %s: %s", c.ID, err.Error())
			}

			value, err := parseEnv(info.Config.Env, envVarName)
			if err != nil {
				log.Fatalf("Failed to parse env: %s", err.Error())
			}
			fmt.Printf("%s", value)
			os.Exit(0)
		}
	}
	log.Fatalf("No matching containers")
}

func parseEnv(e []string, k string) (v string, err error) {
	for _, p := range e {
		pair := strings.SplitN(p, "=", 2)
		if len(pair) == 2 && pair[0] == k {
			return pair[1], nil
		}
	}
	return "", fmt.Errorf("no such key '%s'", k)
}
