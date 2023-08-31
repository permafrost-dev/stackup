package downloader

import (
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stackup-app/stackup/lib/utils"
)

type Downloader struct {
	Gateway *gateway.Gateway
}

func New(gateway *gateway.Gateway) *Downloader {
	return &Downloader{Gateway: gateway}
}

func (d *Downloader) Download(url string, targetPath string) {
	if utils.IsNonEmptyFile(targetPath) {
		return
	}

	if utils.IsFile(targetPath) {
		utils.RemoveFile(targetPath)
	}

	d.Gateway.SaveUrlToFile(url, targetPath)
}
