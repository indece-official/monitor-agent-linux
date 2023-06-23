package agent

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
)

const CheckerTypeAptUpdates = "com.indece.agent.linux.v1.checker.aptupdates"

type AptUpdatesChecker struct {
}

func (c *AptUpdatesChecker) GetType() string {
	return CheckerTypeAptUpdates
}

func (c *AptUpdatesChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
		Name:    "APT-Updates",
		Type:    CheckerTypeAptUpdates,
		Version: "",
		Params: []*apiagent.CheckerV1Param{
			{
				Name:     "exec_apt_update",
				Label:    "Execute APT-Update",
				Type:     apiagent.CheckerV1ParamType_CheckerV1ParamTypeBoolean,
				Required: true,
			},
		},
		Values: []*apiagent.CheckerV1Value{
			{
				Name: "count_available",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
			},
			{
				Name:    "count_security",
				Type:    apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				MaxWarn: "1",
			},
		},
		// Run only once per day
		DefaultSchedule: "0 4 1 * * *",
		DefaultTimeout:  "120s",
	}, nil
}

func (c *AptUpdatesChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	checks := []*apiagent.CheckV1{}

	_, err := os.Stat("/usr/bin/apt")
	if !os.IsNotExist(err) {
		check := &apiagent.CheckV1{
			Name:        "APT-Updates",
			Type:        CheckerTypeAptUpdates,
			CheckerType: CheckerTypeAptUpdates,
			Params: []*apiagent.CheckV1Param{
				{
					Name:  "exec_apt_update",
					Value: "true",
				},
			},
		}

		checks = append(checks, check)
	}

	return checks, nil
}

func (c *AptUpdatesChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	var err error

	paramExecAptUpdate := true

	for _, param := range params {
		if param.Value == "" {
			continue
		}

		switch param.Name {
		case "exec_apt_update":
			paramExecAptUpdate, err = strconv.ParseBool(param.Value)
			if err != nil {
				return "", nil, fmt.Errorf("error parsing parameter '%s': %s", param.Name, err)
			}
		default:
			return "", nil, fmt.Errorf("unknown parameter '%s'", param.Name)
		}
	}

	if paramExecAptUpdate {
		cmdUpdate := exec.Command("/usr/bin/apt", "update")
		out, err := cmdUpdate.Output()
		if err != nil {
			return "", nil, fmt.Errorf("error running apt update: %s (%s)", err, string(out))
		}
	}

	cmdCheck := exec.Command("/usr/bin/apt", "list", "--upgradable")
	stdoutCheck := bytes.Buffer{}
	stderrCheck := bytes.Buffer{}

	cmdCheck.Stdout = &stdoutCheck
	cmdCheck.Stderr = &stderrCheck
	err = cmdCheck.Run()
	out := stdoutCheck.String() + stderrCheck.String()

	if err != nil {
		return "", nil, fmt.Errorf("error running apt-check: %s (%s)", err, out)
	}

	countAvailable := int64(0)
	countSecurity := int64(0)

	outRows := strings.Split(stdoutCheck.String(), "\n")
	for _, row := range outRows {
		if !strings.Contains(row, "[upgradable from:") {
			continue
		}

		rowParts := strings.Split(row, " ")
		if len(rowParts) < 3 {
			continue
		}

		identifierParts := strings.Split(rowParts[0], "/")
		if len(identifierParts) < 2 {
			continue
		}

		repos := identifierParts[len(identifierParts)-1]
		if strings.Contains(repos, "-security") {
			countSecurity++
		}

		countAvailable++
	}

	values := []*apiagent.CheckV1Value{}

	values = append(values, &apiagent.CheckV1Value{
		Name:  "count_available",
		Value: fmt.Sprintf("%d", countAvailable),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "count_security",
		Value: fmt.Sprintf("%d", countSecurity),
	})

	message := ""

	if countAvailable == 0 {
		message = "No updates available"
	} else {
		message = fmt.Sprintf(
			"%d update(s) are available (including %d security updates)",
			countAvailable,
			countSecurity,
		)
	}

	return message, values, nil
}

var _ IChecker = (*AptUpdatesChecker)(nil)

func NewAptUpdatesChecker() *AptUpdatesChecker {
	return &AptUpdatesChecker{}
}
