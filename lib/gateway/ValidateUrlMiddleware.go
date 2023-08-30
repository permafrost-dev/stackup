package gateway

import (
	"errors"
	"net/url"

	"github.com/stackup-app/stackup/lib/messages"
	"github.com/stackup-app/stackup/lib/types"
)

var ValidateUrlMiddleware = GatewayUrlRequestMiddleware{
	Name: "validateUrl",
	Handler: func(g *Gateway, link string) error {
		if !g.Enabled {
			return nil
		}

		parsedUrl, err := url.Parse(link)
		if err != nil {
			return err
		}

		if g.checkArrayForDomainMatch(&g.AllowedDomains, parsedUrl.Host) {
			return nil
		}

		if g.checkArrayForDomainMatch(&g.DeniedDomains, parsedUrl.Host) {
			return errors.New(messages.AccessBlocked(types.AccessTypeDomain, parsedUrl.Host))
		}

		return errors.New(messages.NotExplicitlyAllowed(types.AccessTypeDomain, parsedUrl.Host))
	},
}
