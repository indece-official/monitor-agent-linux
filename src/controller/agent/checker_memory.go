package agent

import (
	"context"
	"fmt"
	"math"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/indece-official/monitor-agent-linux/src/utils"
	"github.com/shirou/gopsutil/mem"
)

const CheckerTypeMemory = "com.indece.agent.linux.v1.checker.memory"

type MemoryChecker struct {
}

func (c *MemoryChecker) GetType() string {
	return CheckerTypeMemory
}

func (c *MemoryChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
		Name:    "Memory",
		Type:    CheckerTypeMemory,
		Version: "",
		Params:  []*apiagent.CheckerV1Param{},
		Values: []*apiagent.CheckerV1Value{
			{
				Name: "total",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
			},
			{
				Name: "used",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
			},
			{
				Name:    "used_percent",
				Type:    apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				MaxWarn: "80",
				MaxCrit: "90",
			},
		},
	}, nil
}

func (c *MemoryChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	return []*apiagent.CheckV1{
		{
			Name:        "Memory",
			Type:        CheckerTypeMemory,
			CheckerType: CheckerTypeMemory,
			Params:      []*apiagent.CheckV1Param{},
		},
	}, nil
}

func (c *MemoryChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return "", nil, fmt.Errorf("error loading memory stats: %s", err)
	}

	values := []*apiagent.CheckV1Value{}

	values = append(values, &apiagent.CheckV1Value{
		Name:  "total",
		Value: fmt.Sprintf("%d", memStats.Total),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "used",
		Value: fmt.Sprintf("%d", memStats.Used),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "used_percent",
		Value: fmt.Sprintf("%.1f", float64(memStats.Used)/math.Max(float64(memStats.Total), 1)*100.0),
	})

	message := fmt.Sprintf(
		"%.1f%% (%s of %s) used of memory",
		float64(memStats.Used)/math.Max(float64(memStats.Total), 1)*100.0,
		utils.FormatBytes(int64(memStats.Used)),
		utils.FormatBytes(int64(memStats.Total)),
	)

	return message, values, nil
}

var _ IChecker = (*MemoryChecker)(nil)

func NewMemoryChecker() *MemoryChecker {
	return &MemoryChecker{}
}
