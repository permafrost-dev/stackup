package gateway_test

import (
	"testing"

	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stretchr/testify/assert"
)

func TestGatewayEnable(t *testing.T) {
	g := gateway.New(nil)
	g.Enabled = false
	g.Enable()
	assert.True(t, g.Enabled, "gateway should be enabled")
}

func TestGatewayDisable(t *testing.T) {
	g := gateway.New(nil)
	g.Enabled = true
	g.Disable()
	assert.False(t, g.Enabled, "gateway should be disabled")
}

// func TestGatewayAllowed(t *testing.T) {
// 	g := gateway.New([]string{}, []string{"*.example.com", "*.one.example.net", "api.**.com"}, []string{}, []string{})
// 	verifyChecksums := true
// 	enableStats := false
// 	gatewayAllow := "allow"

// 	s := &settings.Settings{}
// 	gateway.GatewayMiddleware.AddPreMiddleware(&gateway.ValidateUrlMiddleware)
// 	gateway.GatewayMiddleware.AddPreMiddleware(&gateway.V)

// 	g.Initialize(s, nil, nil)
// 	assert.True(t, g.Allowed("https://www.example.com"), "www.example.com should be allowed")
// 	assert.True(t, g.Allowed("https://example.com"), "example.com should be allowed")
// 	assert.False(t, g.Allowed("https://www.example.net"), "www.example.net should not be allowed")
// 	assert.True(t, g.Allowed("https://one.example.net"), "one.example.net should be allowed")
// 	assert.True(t, g.Allowed("https://a.one.example.net"), "a.one.example.net should be allowed")
// 	assert.True(t, g.Allowed("https://api.test.com"), "api.test.com should be allowed")
// 	assert.True(t, g.Allowed("https://api.one.test.com"), "api.one.test.com should be allowed")
// 	assert.False(t, g.Allowed("https://api.test.example.org"), "api.one.example.org should not be allowed")
// // }
