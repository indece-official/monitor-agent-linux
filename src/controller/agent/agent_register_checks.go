package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/shirou/gopsutil/disk"
)

func (c *Controller) registerChecks(ctx context.Context) error {
	req := &apiagent.RegisterCheckV1Request{
		Check: &apiagent.CheckV1{
			Name:        "OS",
			Type:        CheckerTypeOS,
			CheckerType: CheckerTypeOS,
			Params:      []*apiagent.CheckV1Param{},
		},
	}

	_, err := c.grpcClient.RegisterCheckV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering os-check: %s", err)
	}

	req = &apiagent.RegisterCheckV1Request{
		Check: &apiagent.CheckV1{
			Name:        "Memory",
			Type:        CheckerTypeMemory,
			CheckerType: CheckerTypeMemory,
			Params:      []*apiagent.CheckV1Param{},
		},
	}

	_, err = c.grpcClient.RegisterCheckV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering memory-check: %s", err)
	}

	req = &apiagent.RegisterCheckV1Request{
		Check: &apiagent.CheckV1{
			Name:        "CPU",
			Type:        CheckerTypeCPU,
			CheckerType: CheckerTypeCPU,
			Params:      []*apiagent.CheckV1Param{},
		},
	}

	_, err = c.grpcClient.RegisterCheckV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering cpu-check: %s", err)
	}

	partitions, err := disk.Partitions(false)
	if err != nil {
		return fmt.Errorf("error loading partition list: %s", err)
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

		req = &apiagent.RegisterCheckV1Request{
			Check: &apiagent.CheckV1{
				Name:        fmt.Sprintf("Disk %s", partition.Mountpoint),
				Type:        fmt.Sprintf("%s:%s", CheckerTypeDisk, partition.Mountpoint),
				CheckerType: CheckerTypeDisk,
				Params: []*apiagent.CheckV1Param{
					{
						Name:  "mountpoint",
						Value: partition.Mountpoint,
					},
				},
			},
		}

		_, err = c.grpcClient.RegisterCheckV1(ctx, req)
		if err != nil {
			return fmt.Errorf("error registering disk-check: %s", err)
		}
	}

	req = &apiagent.RegisterCheckV1Request{
		Check: &apiagent.CheckV1{
			Name:        "Uptime",
			Type:        CheckerTypeUptime,
			CheckerType: CheckerTypeUptime,
			Params:      []*apiagent.CheckV1Param{},
		},
	}

	_, err = c.grpcClient.RegisterCheckV1(ctx, req)
	if err != nil {
		return fmt.Errorf("error registering uptime-check: %s", err)
	}

	_, err = os.Stat("/usr/bin/apt")
	if !os.IsNotExist(err) {
		req = &apiagent.RegisterCheckV1Request{
			Check: &apiagent.CheckV1{
				Name:        "APT-Updates",
				Type:        CheckerTypeAptUpdates,
				CheckerType: CheckerTypeAptUpdates,
				Params: []*apiagent.CheckV1Param{
					{
						Name:  "exec_apt_update",
						Value: "true",
					},
				},
			},
		}

		_, err = c.grpcClient.RegisterCheckV1(ctx, req)
		if err != nil {
			return fmt.Errorf("error registering uptime-check: %s", err)
		}
	}

	return nil
}
