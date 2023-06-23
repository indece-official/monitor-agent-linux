package agent

import (
	"context"
	"fmt"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/shirou/gopsutil/host"
)

const CheckerTypeOS = "com.indece.agent.linux.v1.checker.os"

type OSChecker struct {
}

func (c *OSChecker) GetType() string {
	return CheckerTypeOS
}

func (c *OSChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
		Name:    "OS",
		Type:    CheckerTypeOS,
		Version: "",
		Params:  []*apiagent.CheckerV1Param{},
		Values: []*apiagent.CheckerV1Value{
			{
				Name: "os",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeText,
			},
			{
				Name: "platform",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeText,
			},
			{
				Name: "platform_version",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeText,
			},
		},
		// Run only every 5 min
		DefaultSchedule: "6 */5 * * * *",
	}, nil
}

func (c *OSChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	return []*apiagent.CheckV1{
		{
			Name:        "OS",
			Type:        CheckerTypeOS,
			CheckerType: CheckerTypeOS,
			Params:      []*apiagent.CheckV1Param{},
		},
	}, nil
}

func (c *OSChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return "", nil, fmt.Errorf("error loading host info: %s", err)
	}

	values := []*apiagent.CheckV1Value{}

	values = append(values, &apiagent.CheckV1Value{
		Name:  "os",
		Value: hostInfo.OS,
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "platform",
		Value: hostInfo.Platform,
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "platform_version",
		Value: hostInfo.PlatformVersion,
	})

	message := fmt.Sprintf(
		"Running %s %s (kernel %s)",
		hostInfo.Platform,
		hostInfo.PlatformVersion,
		hostInfo.KernelVersion,
	)

	return message, values, nil
}

var _ IChecker = (*OSChecker)(nil)

func NewOSChecker() *OSChecker {
	return &OSChecker{}
}
