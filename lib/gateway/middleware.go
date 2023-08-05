package gateway

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"
)

var (
	ValidateUrlMiddleware    = GatewayUrlRequestMiddleware{Name: "validateUrl", Handler: validateUrlHandler}
	VerifyFileTypeMiddleware = GatewayUrlRequestMiddleware{Name: "verifyFileType", Handler: verifyFileTypeHandler}
	VerifyContentType        = GatewayUrlResponseMiddleware{Name: "verifyContentType", Handler: verifyContentTypeHandler}
)

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

	if g.checkArrayForMatch(&blockedTypes, contentType) {
		return errors.New("content type blocked")
	}

	allowedTypes := g.GetDomainContentTypes(resp.Request.URL.Hostname())
	if g.checkArrayForMatch(&allowedTypes, contentType) || len(allowedTypes) == 0 {
		return nil
	}

	return errors.New("content type '" + contentType + "' not allowed")
}

func validateUrlHandler(g *Gateway, link string) error {
	if !g.Enabled {
		return nil
	}

	parsedUrl, err := url.Parse(link)
	if err != nil {
		return err
	}

	// Check if URL is in the denied list
	if g.checkArrayForMatch(&g.DeniedDomains, parsedUrl.Host) {
		return errors.New("the url is in the denied list")
	}

	if g.checkArrayForMatch(&g.AllowedDomains, parsedUrl.Host) {
		return nil
	}

	return errors.New("url domain has not been explicitly allowed")
}

func verifyFileTypeHandler(g *Gateway, link string) error {
	if !g.Enabled {
		return nil
	}

	parsedUrl, err := url.Parse(link)
	if err != nil {
		return err
	}

	baseName := path.Base(parsedUrl.Path)
	fileExt := path.Ext(baseName)

	if fileExt == "" {
		return nil
	}

	allowedFileNames := []string{"checksums.txt", "checksums.sha256.txt", "checksums.sha512.txt", "sha256sum", "sha512sum"}
	allowedExts := []string{".yaml", ".yml", ".txt", ".sha256", ".sha512", ".json", ".js"}

	for _, name := range allowedFileNames {
		if strings.EqualFold(baseName, name) {
			return nil
		}
	}

	for _, ext := range allowedExts {
		if strings.EqualFold(fileExt, ext) {
			return nil
		}
	}

	return errors.New("file name or extension has not been explicitly allowed")
}
