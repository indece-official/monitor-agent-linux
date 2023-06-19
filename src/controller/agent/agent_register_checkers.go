package agent

import (
	"context"
	"fmt"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
)

func (c *Controller) registerCheckers(ctx context.Context) error {
	req := &apiagent.RegisterCheckerV1Request{
		Checker: &apiagent.CheckerV1{
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
		},
	}

	_, err := c.grpcClient.RegisterCheckerV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering os-checker: %s", err)
	}

	req = &apiagent.RegisterCheckerV1Request{
		Checker: &apiagent.CheckerV1{
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
		},
	}

	_, err = c.grpcClient.RegisterCheckerV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering memory-checker: %s", err)
	}

	req = &apiagent.RegisterCheckerV1Request{
		Checker: &apiagent.CheckerV1{
			Name:    "CPU",
			Type:    CheckerTypeCPU,
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
			},
		},
	}

	_, err = c.grpcClient.RegisterCheckerV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering cpu-checker: %s", err)
	}

	req = &apiagent.RegisterCheckerV1Request{
		Checker: &apiagent.CheckerV1{
			Name:    "Disk",
			Type:    CheckerTypeDisk,
			Version: "",
			Params: []*apiagent.CheckerV1Param{
				{
					Name:     "mountpoint",
					Label:    "Mountpoint",
					Type:     apiagent.CheckerV1ParamType_CheckerV1ParamTypeText,
					Required: true,
				},
			},
			Values: []*apiagent.CheckerV1Value{
				{
					Name: "total",
					Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				},
				{
					Name: "used",
					Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				},
			},
		},
	}

	_, err = c.grpcClient.RegisterCheckerV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering disk-checker: %s", err)
	}

	req = &apiagent.RegisterCheckerV1Request{
		Checker: &apiagent.CheckerV1{
			Name:    "Uptime",
			Type:    CheckerTypeUptime,
			Version: "",
			Params:  []*apiagent.CheckerV1Param{},
			Values: []*apiagent.CheckerV1Value{
				{
					Name: "uptime",
					Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeDuration,
				},
			},
		},
	}

	_, err = c.grpcClient.RegisterCheckerV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering uptime-checker: %s", err)
	}

	req = &apiagent.RegisterCheckerV1Request{
		Checker: &apiagent.CheckerV1{
			Name:         "Process",
			Type:         CheckerTypeProcess,
			Version:      "",
			CustomChecks: true,
			Params: []*apiagent.CheckerV1Param{
				{
					Name:     "name",
					Label:    "Name",
					Type:     apiagent.CheckerV1ParamType_CheckerV1ParamTypeText,
					Required: true,
				},
			},
			Values: []*apiagent.CheckerV1Value{
				{
					Name: "status",
					Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeText,
				},
			},
		},
	}

	_, err = c.grpcClient.RegisterCheckerV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering process-checker: %s", err)
	}

	req = &apiagent.RegisterCheckerV1Request{
		Checker: &apiagent.CheckerV1{
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
		},
	}

	_, err = c.grpcClient.RegisterCheckerV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering ping-checker: %s", err)
	}

	req = &apiagent.RegisterCheckerV1Request{
		Checker: &apiagent.CheckerV1{
			Name:         "HTTP",
			Type:         CheckerTypeHTTP,
			Version:      "",
			CustomChecks: true,
			Params: []*apiagent.CheckerV1Param{
				{
					Name:     "url",
					Label:    "URL",
					Type:     apiagent.CheckerV1ParamType_CheckerV1ParamTypeText,
					Required: true,
				},
				{
					Name:  "dns",
					Label: "DNS",
					Type:  apiagent.CheckerV1ParamType_CheckerV1ParamTypeText,
				},
				{
					Name:  "status",
					Label: "Status",
					Type:  apiagent.CheckerV1ParamType_CheckerV1ParamTypeNumber,
				},
				{
					Name:  "timeout",
					Label: "Timeout",
					Type:  apiagent.CheckerV1ParamType_CheckerV1ParamTypeDuration,
				},
			},
			Values: []*apiagent.CheckerV1Value{
				{
					Name: "resp_time",
					Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeDuration,
				},
				{
					Name: "status_code",
					Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				},
				{
					Name: "resp_size",
					Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				},
				{
					Name: "tls_expiry",
					Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeDuration,
				},
			},
		},
	}

	_, err = c.grpcClient.RegisterCheckerV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering http-checker: %s", err)
	}

	return nil
}
