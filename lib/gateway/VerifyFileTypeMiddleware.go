package gateway

import (
	"errors"
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

		parsedUrl, err := url.Parse(link)
		if err != nil {
			return err
		}

		fileExt := path.Ext(parsedUrl.Path)

		if fileExt == "." || fileExt == "" {
			return nil
		}

		for _, ext := range g.AllowedFileExts {
			if utils.GlobMatch(ext, fileExt, false) || strings.EqualFold(fileExt, ext) {
				return nil
			}
		}

		for _, ext := range g.BlockedFileExts {
			if utils.GlobMatch(ext, fileExt, false) || strings.EqualFold(fileExt, ext) {
				return errors.New("access to file extension '" + fileExt + "' is blocked")
			}
		}

		return errors.New("access to file extension '" + fileExt + "' has not been explicitly allowed")
	},
}
