package agent

import (
	"context"
	"fmt"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
)

func (c *Controller) registerChecks(ctx context.Context) error {
	for _, checker := range c.checkers {
		reqChecks, err := checker.GetChecks()
		if err != nil {
			return fmt.Errorf("error loading checks registration for checker %s: %s", checker.GetType(), err)
		}

		for _, reqCheck := range reqChecks {
			req := &apiagent.RegisterCheckV1Request{}
			req.Check = reqCheck

			_, err = c.grpcClient.RegisterCheckV1(ctx, req)
			if err != nil {
				return fmt.Errorf("error registering check %s for checker %s: %s", reqCheck.CheckerType, checker.GetType(), err)
			}
		}
	}

	return nil
}
