package integ

import (
	"github.com/superdurable/iwf/service"
	"testing"
	"time"
)

// remove the underscore to run
// nolint
func _TestNothingButJustRunningTheServiceTemporalWorkerForDebug(t *testing.T) {
	startIwfServiceWithClient(service.BackendTypeTemporal)
	time.Sleep(time.Hour)
}
