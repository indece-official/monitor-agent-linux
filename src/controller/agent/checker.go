package agent

import (
	"context"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
)

type IChecker interface {
	GetType() string
	GetChecker() (*apiagent.CheckerV1, error)
	GetChecks() ([]*apiagent.CheckV1, error)
	Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error)
}
