package gateway

import (
	"errors"
	"net/url"
)

var ValidateUrlMiddleware = GatewayUrlRequestMiddleware{Name: "validateUrl", Handler: validateUrlHandler}

func validateUrlHandler(g *Gateway, link string) error {
	if !g.Enabled {
		return nil
	}

	parsedUrl, err := url.Parse(link)
	if err != nil {
		return err
	}

	if g.checkArrayForDomainMatch(&g.DeniedDomains, parsedUrl.Host) {
		return errors.New("access to domain" + parsedUrl.Host + " is not allowed")
	}

	if g.checkArrayForDomainMatch(&g.AllowedDomains, parsedUrl.Host) {
		return nil
	}

	return errors.New("url domain has not been explicitly allowed")
}
