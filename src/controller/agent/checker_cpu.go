package agent

import (
	"context"
	"fmt"
	"math"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

const CheckerTypeCpu = "com.indece.agent.linux.v1.checker.cpu"

type CpuChecker struct {
}

func (c *CpuChecker) GetType() string {
	return CheckerTypeCpu
}

func (c *CpuChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
		Name:    "CPU",
		Type:    CheckerTypeCpu,
		Version: "",
		Params:  []*apiagent.CheckerV1Param{},
		Values: []*apiagent.CheckerV1Value{
			{
				Name: "count",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
			},
			{
				Name: "load_1",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
			},
			{
				Name: "load_5",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
			},
			{
				Name: "load_15",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
			},
			{
				Name: "load_1_percent",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
			},
			{
				Name:    "load_5_percent",
				Type:    apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				MaxWarn: "80",
				MaxCrit: "90",
			},
			{
				Name:    "load_15_percent",
				Type:    apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				MaxWarn: "80",
				MaxCrit: "90",
			},
		},
	}, nil
}

func (c *CpuChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	return []*apiagent.CheckV1{
		{
			Name:        "CPU",
			Type:        CheckerTypeCpu,
			CheckerType: CheckerTypeCpu,
			Params:      []*apiagent.CheckV1Param{},
		},
	}, nil
}

func (c *CpuChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	l, err := load.Avg()
	if err != nil {
		return "", nil, fmt.Errorf("error loading load stats: %s", err)
	}

	count, err := cpu.Counts(true)
	if err != nil {
		return "", nil, fmt.Errorf("error loading number of cpus: %s", err)
	}

	values := []*apiagent.CheckV1Value{}

	values = append(values, &apiagent.CheckV1Value{
		Name:  "count",
		Value: fmt.Sprintf("%d", count),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "load_1",
		Value: fmt.Sprintf("%.2f", l.Load1),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "load_5",
		Value: fmt.Sprintf("%.2f", l.Load5),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "load_15",
		Value: fmt.Sprintf("%.2f", l.Load15),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "load_1_percent",
		Value: fmt.Sprintf("%.2f", (l.Load1/math.Max(float64(count), 1))*100.0),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "load_5_percent",
		Value: fmt.Sprintf("%.2f", (l.Load5/math.Max(float64(count), 1))*100.0),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "load_15_percent",
		Value: fmt.Sprintf("%.2f", (l.Load15/math.Max(float64(count), 1))*100.0),
	})

	message := fmt.Sprintf(
		"Load(1) = %.2f, Load(5) = %.2f, Load(15) = %.2f for %d cores",
		l.Load1,
		l.Load5,
		l.Load15,
		count,
	)

	return message, values, nil
}

var _ IChecker = (*CpuChecker)(nil)

func NewCpuChecker() *CpuChecker {
	return &CpuChecker{}
}
