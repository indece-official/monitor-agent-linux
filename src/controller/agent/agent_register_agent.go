package agent

import (
	"context"
	"fmt"

	"github.com/indece-official/monitor-agent-linux/src/buildvars"
	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
)

func (c *Controller) registerAgent(ctx context.Context) error {
	req := &apiagent.RegisterAgentV1Request{
		Type:    "com.indece.agent.linux.v1",
		Version: buildvars.BuildVersion,
	}

	_, err := c.grpcClient.RegisterAgentV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering agent: %s", err)
	}

	return nil
}
