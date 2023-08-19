package gateway

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/stackup-app/stackup/lib/utils"
)

var VerifyFileTypeMiddleware = GatewayUrlRequestMiddleware{
	Name: "verifyFileType",
	Handler: func(g *Gateway, link string) error {
		if !g.Enabled {
			return nil
		}

		allowedExts := []string{"*"}
		blockedExts := []string{"*"}

		if len(g.AllowedFileExts) > 0 {
			allowedExts = g.AllowedFileExts
		}

		if len(g.BlockedFileExts) > 0 {
			blockedExts = g.BlockedFileExts
		}

		fmt.Printf("Allowedexts: %v\n", allowedExts)
		fmt.Printf("Blockedexts: %v\n", blockedExts)

		parsedUrl, err := url.Parse(link)
		if err != nil {
			return err
		}

		baseName := path.Base(parsedUrl.Path)
		fileExt := path.Ext(baseName)

		if fileExt == "." || fileExt == "" {
			return nil
		}

		for _, ext := range allowedExts {
			if utils.GlobMatch(ext, fileExt, false) || strings.EqualFold(fileExt, ext) {
				return nil
			}
		}

		for _, ext := range blockedExts {
			if utils.GlobMatch(ext, fileExt, false) || strings.EqualFold(fileExt, ext) {
				return errors.New("access to file extension '" + fileExt + "' is blocked")
			}
		}

		return errors.New("access to file extension '" + fileExt + "' has not been explicitly allowed")
	},
}
