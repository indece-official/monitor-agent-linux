//go:generate go run assets/generate.go
//go:generate /bin/sh -c "mkdir -p generated/model/apiagent && protoc --go_out=./generated/model/ --go-grpc_out=./generated/model/ --proto_path=../assets/grpc/ ../assets/grpc/agent.proto"
package main

import (
	"fmt"

	"github.com/indece-official/go-gousu/v2/gousu"
	"github.com/indece-official/monitor-agent-linux/src/buildvars"
	"github.com/indece-official/monitor-agent-linux/src/controller/agent"
)

func main() {
	runner := gousu.NewRunner(buildvars.ProjectName, fmt.Sprintf("%s (Build %s)", buildvars.BuildVersion, buildvars.BuildDate))

	runner.CreateController(agent.NewController)

	runner.Run()
}
