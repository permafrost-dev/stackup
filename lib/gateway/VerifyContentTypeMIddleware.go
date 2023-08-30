package gateway

import (
	"errors"
	"net/http"
	"strings"

	"github.com/stackup-app/stackup/lib/messages"
	"github.com/stackup-app/stackup/lib/types"
)

var VerifyContentTypeMIddleware = GatewayUrlResponseMiddleware{
	Name: "verifyContentType",
	Handler: func(g *Gateway, resp *http.Response) error {
		if !g.Enabled {
			return nil
		}

		contentType, _, _ := strings.Cut(resp.Header.Get("Content-Type"), ";")

		allowedTypes := g.GetDomainContentTypes(resp.Request.URL.Hostname())
		if g.checkArrayForDomainMatch(&allowedTypes, contentType) || len(allowedTypes) == 0 {
			return nil
		}

		blockedTypes := g.GetBlockedContentTypes(resp.Request.URL.Hostname())
		if g.checkArrayForDomainMatch(&blockedTypes, contentType) {
			return errors.New(messages.AccessBlocked(types.AccessTypeContentType, contentType))
		}

		return errors.New(messages.NotExplicitlyAllowed(types.AccessTypeContentType, contentType))
	},
}
