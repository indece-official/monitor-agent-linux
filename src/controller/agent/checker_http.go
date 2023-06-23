package agent

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/indece-official/monitor-agent-linux/src/utils"
	"gopkg.in/guregu/null.v4"
)

const CheckerTypeHttp = "com.indece.agent.linux.v1.checker.http"

type HttpChecker struct {
}

func (c *HttpChecker) GetType() string {
	return CheckerTypeHttp
}

func (c *HttpChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
		Name:         "HTTP",
		Type:         CheckerTypeHttp,
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
	}, nil
}

func (c *HttpChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	return []*apiagent.CheckV1{}, nil
}

func (c *HttpChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
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

					addrParts[0], err = utils.ResolveDNS(addrParts[0], paramDNS.String)
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

var _ IChecker = (*HttpChecker)(nil)

func NewHttpChecker() *HttpChecker {
	return &HttpChecker{}
}
