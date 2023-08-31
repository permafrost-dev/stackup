package telemetry

import (
	//"github.com/Infisical/infisical-merge/packages/util"
	"crypto/sha256"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/posthog/posthog-go"
	"github.com/stackup-app/stackup/lib/gateway"
)

// this is not a secret, it is a public key
var POSTHOG_API_KEY_FOR_CLI = "V6yB20dWAVDZfG1yDOFkSQPeGachBJst3UPkVJJdkIS"

type Telemetry struct {
	IsEnabled     bool
	posthogClient posthog.Client
}

type NoOpLogger struct{}

func (NoOpLogger) Logf(format string, args ...interface{}) {
	//log.Info().Msgf(format, args...)
}

func (NoOpLogger) Errorf(format string, args ...interface{}) {
	//log.Debug().Msgf(format, args...)
	//fmt.Printf("  Error: %v\n", args)
}

func New(telemetryIsEnabled bool, gw *gateway.Gateway) *Telemetry {
	if POSTHOG_API_KEY_FOR_CLI != "" {
		trans := CustomTransport{Gateway: gw, Transport: http.DefaultTransport}
		client, _ := posthog.NewWithConfig(
			"phc_"+POSTHOG_API_KEY_FOR_CLI,
			posthog.Config{
				Transport: &trans,
				Logger:    NoOpLogger{},
			},
		)

		return &Telemetry{IsEnabled: telemetryIsEnabled, posthogClient: client}
	} else {
		return &Telemetry{IsEnabled: false}
	}
}

func (t *Telemetry) EventOnly(name string) {
	if !t.IsEnabled {
		return
	}

	t.Event(name, map[string]interface{}{})
}

func (t *Telemetry) Event(name string, properties map[string]interface{}) {
	if !t.IsEnabled {
		return
	}

	userId, _ := t.GetDistinctId()
	posthogProps := posthog.NewProperties()
	for key, value := range properties {
		posthogProps.Set(key, value)
	}

	t.posthogClient.Enqueue(posthog.Capture{
		DistinctId: userId,
		Event:      name,
		Timestamp:  time.Now(),
		Properties: posthogProps,
	})

	defer t.posthogClient.Close()
}

func (t *Telemetry) CaptureEvent(eventName string, properties posthog.Properties) {
	userIdentity, err := t.GetDistinctId()
	if err != nil {
		return
	}

	if t.IsEnabled {
		t.posthogClient.Enqueue(posthog.Capture{
			DistinctId: userIdentity,
			Event:      eventName,
			Properties: properties,
		})

		defer t.posthogClient.Close()
	}
}

func (t *Telemetry) GetDistinctId() (string, error) {
	var distinctId string
	var outputErr error

	machineId, err := machineid.ProtectedID("stackup")
	if err != nil {
		outputErr = err
	}

	if machineId == "" {
		machineId, _ = os.Hostname()
		interfaces, _ := net.Interfaces()
		for _, inter := range interfaces {
			if inter.HardwareAddr.String() != "" {
				machineId += inter.HardwareAddr.String() + ","
			}
		}
	}

	machineId = string(sha256.New().Sum([]byte(machineId)))[0:16]
	distinctId = "anonymous_cli_" + machineId

	return distinctId, outputErr
}
