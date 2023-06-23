package agent

import (
	"context"
	"fmt"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/indece-official/monitor-agent-linux/src/utils"
	"github.com/shirou/gopsutil/disk"
	"gopkg.in/guregu/null.v4"
)

const CheckerTypeDisk = "com.indece.agent.linux.v1.checker.disk"

type DiskChecker struct {
}

func (c *DiskChecker) GetType() string {
	return CheckerTypeDisk
}

func (c *DiskChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
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
			{
				Name:    "used_percent",
				Type:    apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				MaxWarn: "80",
				MaxCrit: "90",
			},
		},
	}, nil
}

func (c *DiskChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	checks := []*apiagent.CheckV1{}

	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("error loading partition list: %s", err)
	}

	mapDevices := map[string]bool{}

	for _, partition := range partitions {
		if partition.Fstype == "squashfs" {
			// Ignore squashfs
			continue
		}

		if processed, ok := mapDevices[partition.Device]; processed && ok {
			// Ignore repeated mount points
			continue
		}

		mapDevices[partition.Device] = true

		check := &apiagent.CheckV1{
			Name:        fmt.Sprintf("Disk %s", partition.Mountpoint),
			Type:        fmt.Sprintf("%s:%s", CheckerTypeDisk, partition.Mountpoint),
			CheckerType: CheckerTypeDisk,
			Params: []*apiagent.CheckV1Param{
				{
					Name:  "mountpoint",
					Value: partition.Mountpoint,
				},
			},
		}

		checks = append(checks, check)
	}

	return checks, nil
}

func (c *DiskChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	paramMountpoint := null.String{}

	for _, param := range params {
		if param.Value == "" {
			continue
		}

		switch param.Name {
		case "mountpoint":
			paramMountpoint.Scan(param.Value)
		default:
			return "", nil, fmt.Errorf("unknown parameter '%s'", param.Name)
		}
	}

	if !paramMountpoint.Valid || paramMountpoint.String == "" {
		return "", nil, fmt.Errorf("missing parameter 'mountpoint'")
	}

	values := []*apiagent.CheckV1Value{}

	partitions, err := disk.Partitions(false)
	if err != nil {
		return "", values, fmt.Errorf("error loading partition list: %s", err)
	}

	var partition *disk.PartitionStat

	for _, p := range partitions {
		if p.Mountpoint == paramMountpoint.String {
			partition = &p
			break
		}
	}

	if partition == nil {
		return "", values, fmt.Errorf("error partition not found")
	}

	usage, err := disk.Usage(paramMountpoint.String)
	if err != nil {
		return "", values, fmt.Errorf("error loading partition stats: %s", err)
	}

	values = append(values, &apiagent.CheckV1Value{
		Name:  "total",
		Value: fmt.Sprintf("%d", usage.Total),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "used",
		Value: fmt.Sprintf("%d", usage.Used),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "used_percent",
		Value: fmt.Sprintf("%.0f", usage.UsedPercent),
	})

	message := fmt.Sprintf(
		"%.1f%% (%s of %s) used of filesystem %s mounted on %s (%s)",
		usage.UsedPercent,
		utils.FormatBytes(int64(usage.Used)),
		utils.FormatBytes(int64(usage.Total)),
		partition.Device,
		usage.Path,
		usage.Fstype,
	)

	return message, values, nil
}

var _ IChecker = (*DiskChecker)(nil)

func NewDiskChecker() *DiskChecker {
	return &DiskChecker{}
}
