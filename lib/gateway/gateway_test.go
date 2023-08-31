package gateway_test

import (
	"testing"

	"github.com/stackup-app/stackup/lib/app"
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stackup-app/stackup/lib/scripting"
	"github.com/stackup-app/stackup/lib/settings"
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

func TestGatewayInitialize(t *testing.T) {
	s := &settings.Settings{
		Domains: settings.WorkflowSettingsDomains{
			Allowed: []string{"*.example.com", "*.one.example.net", "api.**.com"},
			Blocked: []string{},
			Hosts:   []settings.WorkflowSettingsDomainsHost{},
		},
		Gateway: settings.WorkflowSettingsGateway{
			Middleware: []string{"validateUrl"},
		},
	}

	g := gateway.New(nil)
	//findTask := func(id string) (any, error) { return nil, nil }
	a := app.NewApplication()
	a.Initialize()

	engine := scripting.CreateNewJavascriptEngine(a.ToInterface) // &sync.Map{}, g, findTask, func() string { return "." })
	// var engineIntf interface{} =
	// var engineContract types.JavaScriptEngineContract = engineIntf.(types.JavaScriptEngineContract)
	g.Initialize(s, engine.AsContract(), nil)

	assert.True(t, g.Enabled, "gateway should be enabled")
	assert.Equal(t, 3, len(g.AllowedDomains), "gateway should have 3 allowed domains")
	assert.NotNil(t, g.HttpClient, "gateway should have a valid HttpClient property")
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
