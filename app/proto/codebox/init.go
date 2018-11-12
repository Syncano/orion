package codebox

import (
	"time"

	"github.com/Syncano/orion/app/proto/codebox/broker"
)

// Timeout is a default timeout for codebox grpc.
const Timeout = 8 * time.Minute

// Runner ...
var Runner broker.ScriptRunnerClient
