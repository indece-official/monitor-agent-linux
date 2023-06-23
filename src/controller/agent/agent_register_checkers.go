package agent

import (
	"context"
	"fmt"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
)

func (c *Controller) registerCheckers(ctx context.Context) error {
	for _, checker := range c.checkers {
		reqChecker, err := checker.GetChecker()
		if err != nil {
			return fmt.Errorf("error loading checker registration for checker %s: %s", checker.GetType(), err)
		}

		req := &apiagent.RegisterCheckerV1Request{}
		req.Checker = reqChecker

		_, err = c.grpcClient.RegisterCheckerV1(ctx, req)
		if err != nil {
			return fmt.Errorf("error registering checker %s: %s", checker.GetType(), err)
		}
	}

	return nil
}
