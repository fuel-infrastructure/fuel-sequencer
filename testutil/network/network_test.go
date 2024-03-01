package network

import (
	"testing"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/app"
)

func TestRunNetwork(t *testing.T) {
	app.InitSDKConfig()

	_ = New(t)
	for {
		time.Sleep(time.Hour)
	}
}
