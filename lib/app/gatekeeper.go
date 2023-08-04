package app

import (
	"net/url"
	"strings"

	glob "github.com/ryanuber/go-glob"
)

type Gatekeeper struct {
	Enabled bool
}

func CreateGatekeeper() *Gatekeeper {
	return &Gatekeeper{Enabled: true}
}

func (g *Gatekeeper) Initialize() {
	g.Enabled = true
}

func (g *Gatekeeper) CanAccessUrl(urlStr string) bool {
	if !g.Enabled {
		return true
	}

	if App.Workflow.Settings.Domains.Allowed == nil {
		return true
	}

	for _, domain := range App.Workflow.Settings.Domains.Allowed {
		parsedUrl, _ := url.Parse(urlStr)

		if strings.EqualFold(parsedUrl.Host, domain) {
			return true
		}

		if strings.Contains(domain, "*") && glob.Glob(domain, parsedUrl.Host) {
			return true
		}
	}

	return false
}
