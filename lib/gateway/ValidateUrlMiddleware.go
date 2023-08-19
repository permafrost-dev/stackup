package gateway

import (
	"errors"
	"net/url"
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
			return errors.New("access to domain" + parsedUrl.Host + " is blocked")
		}

		return errors.New("access to domain" + parsedUrl.Host + " has not been explicitly allowed")
	},
}
