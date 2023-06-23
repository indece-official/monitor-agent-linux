package agent

import (
	"context"
	"fmt"
	"io"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
)

func (c *Controller) check(ctx context.Context, checkClient apiagent.Agent_CheckV1Client, checkRequest *apiagent.CheckV1Request) error {
	checkResult := &apiagent.CheckV1Result{}
	checkResult.ActionUID = checkRequest.ActionUID
	checkResult.CheckUID = checkRequest.CheckUID
	checkResult.Values = []*apiagent.CheckV1Value{}

	checker, ok := c.checkers[checkRequest.CheckerType]
	if !ok {
		checkResult.Error = "Unknown checker type"
		checkResult.Message = "Error: unknown checker type"
	} else {
		message, values, err := checker.Check(ctx, checkRequest.Params)
		if err != nil {
			checkResult.Error = err.Error()
			checkResult.Message = err.Error()
		} else {
			checkResult.Message = message
			checkResult.Values = values
		}
	}

	err := checkClient.Send(checkResult)
	if err != nil {
		return fmt.Errorf("error sending result: %s", err)
	}

	return nil
}

func (c *Controller) checkLoop(ctx context.Context) error {
	c.log.Infof("Starting check receiver")
	defer c.log.Infof("Stopped check receiver")

	checkClient, err := c.grpcClient.CheckV1(ctx)
	if err != nil {
		return fmt.Errorf("error receiving config: %s", err)
	}
	defer checkClient.CloseSend()

	for !c.stop {
		checkRequest, err := checkClient.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		err = c.check(
			ctx,
			checkClient,
			checkRequest,
		)
		if err != nil {
			return fmt.Errorf("error running check: %s", err)
		}
	}

	return nil
}
