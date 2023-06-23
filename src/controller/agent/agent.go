package agent

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/indece-official/go-gousu/v2/gousu"
	"github.com/indece-official/go-gousu/v2/gousu/logger"
	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/namsral/flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const ControllerName = "agent"

var (
	serverHost      = flag.String("server_host", "0.0.0.0", "")
	serverPort      = flag.Int("server_port", 9440, "")
	serverCACrt     = flag.String("ca_crt", "", "")
	serverClientCrt = flag.String("client_crt", "", "")
	serverClientKey = flag.String("client_key", "", "")
)

type IController interface {
	gousu.IController
}

type Controller struct {
	log           *logger.Log
	grpcConn      *grpc.ClientConn
	checkers      map[string]IChecker
	grpcClient    apiagent.AgentClient
	stop          bool
	error         error
	waitGroupStop sync.WaitGroup
}

var _ IController = (*Controller)(nil)

func (c *Controller) Name() string {
	return ControllerName
}

func (c *Controller) addChecker(checker IChecker) {
	c.checkers[checker.GetType()] = checker
}

func (c *Controller) Start() error {
	c.checkers = map[string]IChecker{}

	c.addChecker(NewAptUpdatesChecker())
	c.addChecker(NewDockerContainerChecker())
	c.addChecker(NewCpuChecker())
	c.addChecker(NewDiskChecker())
	c.addChecker(NewHttpChecker())
	c.addChecker(NewMemoryChecker())
	c.addChecker(NewOSChecker())
	c.addChecker(NewPingChecker())
	c.addChecker(NewProcessChecker())
	c.addChecker(NewUptimeChecker())

	caCrtRaw, err := base64.StdEncoding.DecodeString(*serverCACrt)
	if err != nil {
		return fmt.Errorf("error base64-decoding ca crt: %s", err)
	}

	clientCrtRaw, err := base64.StdEncoding.DecodeString(*serverClientCrt)
	if err != nil {
		return fmt.Errorf("error base64-decoding client crt: %s", err)
	}

	clientKeyRaw, err := base64.StdEncoding.DecodeString(*serverClientKey)
	if err != nil {
		return fmt.Errorf("error base64-decoding client key: %s", err)
	}

	clientCrt, err := tls.X509KeyPair(
		clientCrtRaw,
		clientKeyRaw,
	)
	if err != nil {
		return err
	}

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM(caCrtRaw) {
		return fmt.Errorf("credentials: failed to append certificates")
	}

	creds := credentials.NewTLS(&tls.Config{
		RootCAs: rootCAs,
		Certificates: []tls.Certificate{
			clientCrt,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to load tls certs: %s", err)
	}

	c.grpcConn, err = grpc.Dial(
		fmt.Sprintf("%s:%d", *serverHost, *serverPort),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return fmt.Errorf(
			"connecting to grpc server on %s:%d failed: %s",
			*serverHost,
			*serverPort,
			err,
		)
	}

	c.grpcClient = apiagent.NewAgentClient(c.grpcConn)

	ctx := context.Background()

	c.waitGroupStop.Add(1)
	go func() {
		for !c.stop {
			c.error = nil

			err := c.pingLoop(context.Background())
			if err != nil {
				c.error = err

				c.log.Errorf("Error in piing loop: %s", err)
			}

			time.Sleep(5 * time.Second)
		}

		c.waitGroupStop.Done()
	}()

	c.waitGroupStop.Add(1)
	go func() {
		for !c.stop {
			time.Sleep(5 * time.Second)

			c.error = nil

			err = c.registerAgent(ctx)
			if err != nil {
				c.error = err

				c.log.Errorf("Error in registering agent: %s", err)

				continue
			}

			err = c.registerCheckers(ctx)
			if err != nil {
				c.error = err

				c.log.Errorf("Error registering checkers: %s", err)

				continue
			}

			err = c.registerChecks(ctx)
			if err != nil {
				c.error = err

				c.log.Errorf("Error registering checks: %s", err)

				continue
			}

			err := c.checkLoop(context.Background())
			if err != nil {
				c.error = err

				c.log.Errorf("Error in check loop: %s", err)
			}
		}

		c.waitGroupStop.Done()
	}()

	return nil
}

func (c *Controller) Health() error {
	return c.error
}

func (c *Controller) Stop() error {
	c.stop = true

	if c.grpcConn != nil {
		err := c.grpcConn.Close()
		if err != nil {
			c.log.Warnf("Error closing connection: %s", err)
		}
	}

	c.waitGroupStop.Wait()

	c.grpcConn = nil

	return nil
}

// NewController creates a new preinitialized instance of Controller
func NewController(ctx gousu.IContext) gousu.IController {
	log := logger.GetLogger(fmt.Sprintf("controller.%s", ControllerName))

	return &Controller{
		log: log,
	}
}

var _ gousu.ControllerFactory = NewController
