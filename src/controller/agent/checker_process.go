package agent

import (
	"context"
	"fmt"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/shirou/gopsutil/process"
	"gopkg.in/guregu/null.v4"
)

const CheckerTypeProcess = "com.indece.agent.linux.v1.checker.process"

type ProcessChecker struct {
}

func (c *ProcessChecker) GetType() string {
	return CheckerTypeProcess
}

func (c *ProcessChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
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
	}, nil
}

func (c *ProcessChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	return []*apiagent.CheckV1{}, nil
}

func (c *ProcessChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	var err error

	paramName := null.String{}

	for _, param := range params {
		if param.Value == "" {
			continue
		}

		switch param.Name {
		case "name":
			paramName.Scan(param.Value)
		default:
			return "", nil, fmt.Errorf("unknown parameter '%s'", param.Name)
		}
	}

	if !paramName.Valid || paramName.String == "" {
		return "", nil, fmt.Errorf("missing parameter 'name'")
	}

	processes, err := process.Processes()
	if err != nil {
		return "", nil, fmt.Errorf("error loading running processes: %s", err)
	}

	var foundProcess *process.Process

	for _, process := range processes {
		name, err := process.Name()
		if err != nil {
			// Ignore errors here (caused if the process terminates before we can read the name)
			continue
		}

		if name == paramName.String {
			foundProcess = process
			break
		}
	}

	if foundProcess == nil {
		return "", nil, fmt.Errorf("no running process with name '%s' found", paramName.String)
	}

	processName, err := foundProcess.Name()
	if err != nil {
		return "", nil, fmt.Errorf("error getting process name: %s", err)
	}

	processStatus, err := foundProcess.Status()
	if err != nil {
		return "", nil, fmt.Errorf("error getting process status: %s", err)
	}

	values := []*apiagent.CheckV1Value{}

	values = append(values, &apiagent.CheckV1Value{
		Name:  "status",
		Value: processStatus,
	})

	message := fmt.Sprintf(
		"Process %s (%d) is running (%s)",
		processName,
		foundProcess.Pid,
		processStatus,
	)

	return message, values, nil
}

var _ IChecker = (*ProcessChecker)(nil)

func NewProcessChecker() *ProcessChecker {
	return &ProcessChecker{}
}
