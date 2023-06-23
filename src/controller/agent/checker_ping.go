package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	probing "github.com/prometheus-community/pro-bing"
	"gopkg.in/guregu/null.v4"
)

const CheckerTypePing = "com.indece.agent.linux.v1.checker.ping"

type PingChecker struct {
}

func (c *PingChecker) GetType() string {
	return CheckerTypePing
}

func (c *PingChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
		Name:         "Ping",
		Type:         CheckerTypePing,
		Version:      "",
		CustomChecks: true,
		Params: []*apiagent.CheckerV1Param{
			{
				Name:     "host",
				Label:    "Host",
				Type:     apiagent.CheckerV1ParamType_CheckerV1ParamTypeText,
				Required: true,
			},
			{
				Name:  "timeout",
				Label: "Timeout",
				Type:  apiagent.CheckerV1ParamType_CheckerV1ParamTypeDuration,
			},
		},
		Values: []*apiagent.CheckerV1Value{
			{
				Name: "time",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeDuration,
			},
		},
	}, nil
}

func (c *PingChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	return []*apiagent.CheckV1{}, nil
}

func (c *PingChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	var err error

	paramHost := null.String{}
	paramTimeout := 3 * time.Second

	for _, param := range params {
		if param.Value == "" {
			continue
		}

		switch param.Name {
		case "host":
			paramHost.Scan(param.Value)
		case "timeout":
			paramTimeout, err = time.ParseDuration(param.Value)
			if err != nil {
				return "", nil, fmt.Errorf("error parsing parameter 'timeout': %s", err)
			}
		default:
			return "", nil, fmt.Errorf("unknown parameter '%s'", param.Name)
		}
	}

	if !paramHost.Valid || paramHost.String == "" {
		return "", nil, fmt.Errorf("missing parameter 'host'")
	}

	pinger, err := probing.NewPinger(paramHost.String)
	if err != nil {
		return "", nil, fmt.Errorf("error initializing ping: %s", err)
	}
	pinger.Count = 1
	pinger.Timeout = paramTimeout
	err = pinger.Run()
	if err != nil {
		return "", nil, fmt.Errorf("error pinging %s: %s", paramHost.String, err)
	}

	stats := pinger.Statistics()

	values := []*apiagent.CheckV1Value{}

	values = append(values, &apiagent.CheckV1Value{
		Name:  "time",
		Value: stats.AvgRtt.String(),
	})

	message := fmt.Sprintf(
		"Ping %s (%dms)",
		paramHost.String,
		stats.AvgRtt/time.Millisecond,
	)

	return message, values, nil
}

var _ IChecker = (*PingChecker)(nil)

func NewPingChecker() *PingChecker {
	return &PingChecker{}
}
