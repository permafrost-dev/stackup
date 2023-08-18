package gateway

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

var (
	VerifyFileTypeMiddleware = GatewayUrlRequestMiddleware{Name: "verifyFileType", Handler: verifyFileTypeHandler}
)

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

	if fileExt == "." || fileExt == "" {
		return nil
	}

	if baseName == "." || baseName == "" {
		return nil
	}

	allowedExts := []string{".yaml", ".yml", ".txt", ".sha256", ".sha512", ".json", ".js"}

	for _, ext := range allowedExts {
		if strings.EqualFold(fileExt, ext) {
			return nil
		}
	}

	return errors.New("file name or extension has not been explicitly allowed")
}
