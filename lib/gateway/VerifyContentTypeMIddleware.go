package gateway

import (
	"errors"
	"net/http"
	"strings"
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
			return errors.New("content type blocked")
		}

		return errors.New("content type '" + contentType + "' has not been explicitly allowed")
	},
}
