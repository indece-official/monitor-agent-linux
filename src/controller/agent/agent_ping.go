package agent

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
)

func (c *Controller) pingLoop(ctx context.Context) error {
	var err error

	c.log.Infof("Starting ping sender")
	defer c.log.Infof("Stopped ping sender")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	tickerPing := time.NewTicker(10 * time.Second)
	defer tickerPing.Stop()

	for !c.stop {
		select {
		case <-ticker.C:
			continue
		case <-tickerPing.C:
			req := &apiagent.PingV1Request{}

			_, err = c.grpcClient.PingV1(ctx, req)
			if err == io.EOF {
				return nil
			}

			if err != nil {
				return fmt.Errorf("error sending ping: %s", err)
			}
		}
	}

	return nil
}
