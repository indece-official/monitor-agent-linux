package agent

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/indece-official/monitor-agent-linux/src/utils"
	"github.com/shirou/gopsutil/host"
)

const CheckerTypeUptime = "com.indece.agent.linux.v1.checker.uptime"

type UptimeChecker struct {
}

func (c *UptimeChecker) GetType() string {
	return CheckerTypeUptime
}

func (c *UptimeChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
		Name:    "Uptime",
		Type:    CheckerTypeUptime,
		Version: "",
		Params:  []*apiagent.CheckerV1Param{},
		Values: []*apiagent.CheckerV1Value{
			{
				Name: "uptime",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeDuration,
			},
			{
				Name:    "restart_required",
				Type:    apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				MaxWarn: "1",
			},
		},
	}, nil
}

func (c *UptimeChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	return []*apiagent.CheckV1{
		{
			Name:        "Uptime",
			Type:        CheckerTypeUptime,
			CheckerType: CheckerTypeUptime,
			Params:      []*apiagent.CheckV1Param{},
		},
	}, nil
}

func (c *UptimeChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	values := []*apiagent.CheckV1Value{}

	uptime, err := host.Uptime()
	if err != nil {
		return "", values, fmt.Errorf("error loading uptime: %s", err)
	}

	uptimeDuration := time.Duration(uptime) * time.Second

	values = append(values, &apiagent.CheckV1Value{
		Name:  "uptime",
		Value: uptimeDuration.String(),
	})

	restartRequired := false

	_, err = os.Stat("/var/run/reboot-required")
	if !os.IsNotExist(err) {
		restartRequired = true
	}

	values = append(values, &apiagent.CheckV1Value{
		Name:  "restart_required",
		Value: fmt.Sprintf("%d", utils.BoolToInt(restartRequired)),
	})

	message := fmt.Sprintf(
		"Host is up for %s",
		utils.FormatDurationPretty(uptimeDuration),
	)

	if restartRequired {
		message += " (system restart required)"
	}

	return message, values, nil
}

var _ IChecker = (*UptimeChecker)(nil)

func NewUptimeChecker() *UptimeChecker {
	return &UptimeChecker{}
}
