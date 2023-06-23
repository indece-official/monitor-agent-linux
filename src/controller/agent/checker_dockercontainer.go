package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"gopkg.in/guregu/null.v4"
)

const CheckerTypeDockerContainer = "com.indece.agent.linux.v1.checker.dockercontainer"

type DockerContainerChecker struct {
}

func (c *DockerContainerChecker) GetType() string {
	return CheckerTypeDockerContainer
}

func (c *DockerContainerChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
		Name:    "DockerContainer",
		Type:    CheckerTypeDockerContainer,
		Version: "",
		Params: []*apiagent.CheckerV1Param{
			{
				Name:     "name",
				Label:    "Name",
				Type:     apiagent.CheckerV1ParamType_CheckerV1ParamTypeText,
				Required: true,
			},
		},
		Values: []*apiagent.CheckerV1Value{
			{
				Name: "status",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeText,
			},
		},
		CustomChecks: true,
	}, nil
}

func (c *DockerContainerChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	return []*apiagent.CheckV1{}, nil
}

func (c *DockerContainerChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	var err error

	paramName := null.String{}

	for _, param := range params {
		if param.Value == "" {
			continue
		}

		switch param.Name {
		case "name":
			paramName.Scan(param.Value)
		default:
			return "", nil, fmt.Errorf("unknown parameter '%s'", param.Name)
		}
	}

	if !paramName.Valid || paramName.String == "" {
		return "", nil, fmt.Errorf("missing parameter 'name'")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", nil, fmt.Errorf("error connecting to docker daemon: %s", err)
	}

	opts := types.ContainerListOptions{All: true}
	opts.Filters = filters.NewArgs()
	opts.Filters.Add("name", paramName.String)

	containers, err := cli.ContainerList(ctx, opts)
	if err != nil {
		return "", nil, fmt.Errorf("error loading containers from docker daemon: %s", err)
	}

	if len(containers) == 0 {
		return "", nil, fmt.Errorf("docker container %s not found", paramName.String)
	}

	container := containers[0]

	containerID := container.ID[0:12]
	name := containerID

	if len(container.Names) > 0 {
		name = strings.TrimLeft(container.Names[0], "/")
	}

	state, err := cli.ContainerInspect(ctx, container.ID)
	if err != nil {
		return "", nil, fmt.Errorf("error inspecting container %s (%s): %s", name, containerID, err)
	}

	// "created", "running", "paused", "restarting", "removing", "exited", "dead":
	if state.State.Status != "running" {
		return "", nil, fmt.Errorf("docker container %s (%s) is in state %s (%s): %s", name, containerID, state.State.Status, container.Image, state.State.Error)
	}

	// Starting, Healthy or Unhealthy
	if state.State.Health != nil && state.State.Health.Status != "Healthy" {
		return "", nil, fmt.Errorf("docker container %s (%s) is running but has health status %s (%s): %s", name, containerID, state.State.Health.Status, container.Image, state.State.Error)
	}

	values := []*apiagent.CheckV1Value{}

	healthMessage := ""
	if state.State.Health != nil && state.State.Health.Status == "Healthy" {
		healthMessage = "and healthy "
	}

	message := fmt.Sprintf(
		"Docker container %s (%s) is running %s(%s)",
		name,
		containerID,
		healthMessage,
		container.Image,
	)

	return message, values, nil
}

var _ IChecker = (*DockerContainerChecker)(nil)

func NewDockerContainerChecker() *DockerContainerChecker {
	return &DockerContainerChecker{}
}
