package inspector

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// DockerInspector implements Inspector for Docker CE.
type DockerInspector struct {
	c *client.Client
}

func newDockerInspector() (Inspector, error) {
	di := &DockerInspector{}

	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil || c == nil {
		return di, fmt.Errorf("error initializing docker client: %s", err)
	}

	return &DockerInspector{
		c: c,
	}, nil
}

// GetValue implements Inspector.
func (di *DockerInspector) GetValue(containerId, varName string) (string, error) {
	values, err := di.GetAllValues(containerId)
	if err != nil {
		return "", err
	}
	if v, ok := values[varName]; ok {
		return v, nil
	}
	return "", fmt.Errorf("Variable %s not set in container %s", varName, containerId)
}

func (di *DockerInspector) GetAllValues(containerId string) (map[string]string, error) {
	values := map[string]string{}
	data, err := di.inspect(containerId)
	if err != nil {
		return values, fmt.Errorf("error inspecting container '%s': %s", containerId, err)
	}
	for _, kv := range data.Config.Env {
		x := strings.SplitN(kv, "=", 2)
		values[x[0]] = x[1]
	}
	return values, nil
}

func (di *DockerInspector) inspect(containerId string) (types.ContainerJSON, error) {
	return di.c.ContainerInspect(context.TODO(), containerId)
}
