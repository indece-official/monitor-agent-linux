package agent

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/indece-official/monitor-agent-linux/src/utils"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"gopkg.in/guregu/null.v4"
)

const (
	CheckerTypeOS      = "com.indece.agent.linux.v1.checker.os"
	CheckerTypeMemory  = "com.indece.agent.linux.v1.checker.memory"
	CheckerTypeCPU     = "com.indece.agent.linux.v1.checker.cpu"
	CheckerTypeDisk    = "com.indece.agent.linux.v1.checker.disk"
	CheckerTypeUptime  = "com.indece.agent.linux.v1.checker.uptime"
	CheckerTypeProcess = "com.indece.agent.linux.v1.checker.process"
	CheckerTypePing    = "com.indece.agent.linux.v1.checker.ping"
	CheckerTypeHTTP    = "com.indece.agent.linux.v1.checker.http"
)

func (c *Controller) checkOS(ctx context.Context) (string, []*apiagent.CheckV1Value, error) {
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

func (c *Controller) checkMemory(ctx context.Context) (string, []*apiagent.CheckV1Value, error) {
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

func (c *Controller) checkCPU(ctx context.Context) (string, []*apiagent.CheckV1Value, error) {
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

	message := fmt.Sprintf(
		"Load(1) = %.2f, Load(5) = %.2f, Load(15) = %.2f for %d cores",
		l.Load1,
		l.Load5,
		l.Load15,
		count,
	)

	return message, values, nil
}

func (c *Controller) checkDisk(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
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

func (c *Controller) checkUptime(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
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

	message := fmt.Sprintf(
		"Host is up for %s",
		utils.FormatDurationPretty(uptimeDuration),
	)

	return message, values, nil
}

func (c *Controller) checkProcess(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
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
			return "", nil, fmt.Errorf("error getting process name: %s", err)
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

func (c *Controller) checkPing(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
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

func (c *Controller) checkHTTP(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	var err error

	paramURL := null.String{}
	paramDNS := null.String{}
	paramTimeout := 5 * time.Second
	paramExpectedStatus := int64(200)

	for _, param := range params {
		if param.Value == "" {
			continue
		}

		switch param.Name {
		case "url":
			paramURL.Scan(param.Value)
		case "dns":
			paramDNS.Scan(param.Value)
		case "timeout":
			paramTimeout, err = time.ParseDuration(param.Value)
			if err != nil {
				return "", nil, fmt.Errorf("error parsing parameter 'timeout': %s", err)
			}
		case "status":
			paramExpectedStatus, err = strconv.ParseInt(param.Value, 10, 64)
			if err != nil {
				return "", nil, fmt.Errorf("error parsing parameter 'status': %s", err)
			}
		default:
			return "", nil, fmt.Errorf("unknown parameter '%s'", param.Name)
		}
	}

	if !paramURL.Valid || paramURL.String == "" {
		return "", nil, fmt.Errorf("missing parameter 'url'")
	}

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				var err error

				if paramDNS.Valid && paramDNS.String != "" {
					addrParts := strings.Split(addr, ":")
					if len(addrParts) != 2 {
						return nil, fmt.Errorf("error parsing address '%s': must have format <host>:<port>", addr)
					}

					addrParts[0], err = c.resolveDNS(addrParts[0], paramDNS.String)
					if err != nil {
						return nil, err
					}

					addr = strings.Join(addrParts, ":")
				}

				return net.Dial(network, addr)
			},
		},
		Timeout: paramTimeout,
	}

	startAt := time.Now()

	values := []*apiagent.CheckV1Value{}

	resp, err := client.Get(paramURL.String)
	if err != nil {
		return "", values, fmt.Errorf("error getting '%s': %s", paramURL.String, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", values, fmt.Errorf("error reading response body from '%s': %s", paramURL.String, err)
	}

	responseTime := time.Since(startAt)

	values = append(values, &apiagent.CheckV1Value{
		Name:  "resp_time",
		Value: responseTime.String(),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "status_code",
		Value: fmt.Sprintf("%d", resp.StatusCode),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "resp_size",
		Value: fmt.Sprintf("%d", len(body)),
	})

	if resp.StatusCode != int(paramExpectedStatus) {
		return "", values, fmt.Errorf("error getting '%s' - %s (expected status %d)", paramURL.String, resp.Status, paramExpectedStatus)
	}

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		now := time.Now()
		duration := resp.TLS.PeerCertificates[0].NotAfter.Sub(now)

		values = append(values, &apiagent.CheckV1Value{
			Name:  "tls_expiry",
			Value: duration.String(),
		})
	}

	message := fmt.Sprintf(
		"GET '%s' - %s\n",
		paramURL.String,
		resp.Status,
	)

	return message, values, nil
}

func (c *Controller) check(ctx context.Context, checkClient apiagent.Agent_CheckV1Client, checkRequest *apiagent.CheckV1Request) error {
	checkResult := &apiagent.CheckV1Result{}
	checkResult.ActionUID = checkRequest.ActionUID
	checkResult.CheckUID = checkRequest.CheckUID
	checkResult.Values = []*apiagent.CheckV1Value{}

	switch checkRequest.CheckerType {
	case CheckerTypeOS:
		message, values, err := c.checkOS(ctx)
		if err != nil {
			checkResult.Error = err.Error()
			checkResult.Message = err.Error()
		} else {
			checkResult.Message = message
			checkResult.Values = values
		}
	case CheckerTypeMemory:
		message, values, err := c.checkMemory(ctx)
		if err != nil {
			checkResult.Error = err.Error()
			checkResult.Message = err.Error()
		} else {
			checkResult.Message = message
			checkResult.Values = values
		}
	case CheckerTypeCPU:
		message, values, err := c.checkCPU(ctx)
		if err != nil {
			checkResult.Error = err.Error()
			checkResult.Message = err.Error()
		} else {
			checkResult.Message = message
			checkResult.Values = values
		}
	case CheckerTypeDisk:
		message, values, err := c.checkDisk(ctx, checkRequest.Params)
		if err != nil {
			checkResult.Error = err.Error()
			checkResult.Message = err.Error()
		} else {
			checkResult.Message = message
			checkResult.Values = values
		}
	case CheckerTypeUptime:
		message, values, err := c.checkUptime(ctx, checkRequest.Params)
		if err != nil {
			checkResult.Error = err.Error()
			checkResult.Message = err.Error()
		} else {
			checkResult.Message = message
			checkResult.Values = values
		}
	case CheckerTypeProcess:
		message, values, err := c.checkProcess(ctx, checkRequest.Params)
		if err != nil {
			checkResult.Error = err.Error()
			checkResult.Message = err.Error()
		} else {
			checkResult.Message = message
			checkResult.Values = values
		}
	case CheckerTypePing:
		message, values, err := c.checkPing(ctx, checkRequest.Params)
		if err != nil {
			checkResult.Error = err.Error()
			checkResult.Message = err.Error()
		} else {
			checkResult.Message = message
			checkResult.Values = values
		}
	case CheckerTypeHTTP:
		message, values, err := c.checkHTTP(ctx, checkRequest.Params)
		if err != nil {
			checkResult.Error = err.Error()
			checkResult.Message = err.Error()
		} else {
			checkResult.Message = message
			checkResult.Values = values
		}
	default:
		checkResult.Error = "Unknown checker type"
		checkResult.Message = "Error: unknown checker type"
	}

	err := checkClient.Send(checkResult)
	if err != nil {
		return fmt.Errorf("error sending result: %s", err)
	}

	return nil
}

func (c *Controller) checkLoop(ctx context.Context) error {
	c.log.Infof("Starting check receiver")
	defer c.log.Infof("Stopped check receiver")

	checkClient, err := c.grpcClient.CheckV1(ctx)
	if err != nil {
		return fmt.Errorf("error receiving config: %s", err)
	}
	defer checkClient.CloseSend()

	for !c.stop {
		checkRequest, err := checkClient.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		err = c.check(
			ctx,
			checkClient,
			checkRequest,
		)
		if err != nil {
			return fmt.Errorf("error running check: %s", err)
		}
	}

	return nil
}
