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
