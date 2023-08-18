package gateway

import (
	"errors"
	"net/http"
	"strings"
)

var VerifyContentTypeMIddleware = GatewayUrlResponseMiddleware{Name: "verifyContentType", Handler: verifyContentTypeHandler}

func verifyContentTypeHandler(g *Gateway, resp *http.Response) error {
	if !g.Enabled {
		return nil
	}

	contentType, _, _ := strings.Cut(resp.Header.Get("Content-Type"), ";")

	// allowedTypes := []string{
	// 	"application/json", "application/javascript", "application/x-yaml", "application/x-yml", "application/yaml", "application/yml",
	// 	"text/*",
	// 	"application/octet-stream", "application/zip", "application/x-zip-compressed",
	// 	"application/x-gzip", "application/gzip", "application/x-tar", "application/tar",
	// 	"application/x-bzip2", "application/bzip2", "application/x-bzip", "application/bzip",
	// 	"application/x-xz", "application/x-lzma", "application/lzma",
	// 	"application/vnd.debian.binary-package", //.deb
	// 	"application/pgp-signature",             // .sig
	// }

	blockedTypes := g.GetBlockedContentTypes(resp.Request.URL.Hostname())

	if g.checkArrayForDomainMatch(&blockedTypes, contentType) {
		return errors.New("content type blocked")
	}

	allowedTypes := g.GetDomainContentTypes(resp.Request.URL.Hostname())
	if g.checkArrayForDomainMatch(&allowedTypes, contentType) || len(allowedTypes) == 0 {
		return nil
	}

	return errors.New("content type '" + contentType + "' not allowed")
}
