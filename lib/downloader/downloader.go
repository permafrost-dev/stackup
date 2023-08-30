package downloader

import (
	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stackup-app/stackup/lib/utils"
)

type Downloader struct {
	Gateway *gateway.Gateway
}

func NewDownloader(gateway *gateway.Gateway) *Downloader {
	return &Downloader{Gateway: gateway}
}

func (d *Downloader) DownloadApplicationIcon(targetPath string) {
	if utils.IsNonEmptyFile(targetPath) {
		return
	}

	if utils.IsFile(targetPath) {
		utils.RemoveFile(targetPath)
	}

	d.Gateway.SaveUrlToFile(consts.APP_ICON_URL, targetPath)
}
