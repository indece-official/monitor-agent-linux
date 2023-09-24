package agent

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
)

func (c *Controller) check(
	ctx context.Context,
	checkClient apiagent.Agent_CheckV1Client,
	mutexCheckClientSend *sync.Mutex,
	checkRequest *apiagent.CheckV1Request,
) error {
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

	mutexCheckClientSend.Lock()
	defer mutexCheckClientSend.Unlock()

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
	mutexCheckClientSend := &sync.Mutex{}

	for !c.stop {
		checkRequest, err := checkClient.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		go func() {
			err := c.check(
				ctx,
				checkClient,
				mutexCheckClientSend,
				checkRequest,
			)
			if err != nil {
				c.log.Errorf("error running check: %s", err)
			}
		}()
	}

	return nil
}
