package telemetry

import (
	"errors"
	"net/http"

	"github.com/stackup-app/stackup/lib/gateway"
)

type CustomTransport struct {
	Gateway   *gateway.Gateway
	Transport http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface, allowing us to modify the request before sending it
func (c *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !c.Gateway.Allowed(req.URL.String()) {
		return nil, errors.New("access to " + req.URL.String() + " is not allowed.")
	}

	return c.Transport.RoundTrip(req)
}
