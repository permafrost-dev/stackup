package checksums

import (
	"net/url"
	"path"
	"strings"
)

func GetChecksumUrls(fullUrl string) []string {
	url, _ := url.Parse(fullUrl)
	reqFn := path.Base(url.Path)
	url.Path = path.Dir(url.Path)

	return []string{
		url.JoinPath("checksums.txt").String(),
		url.JoinPath("checksums.sha256.txt").String(),
		url.JoinPath("checksums.sha512.txt").String(),
		url.JoinPath("sha256sum").String(),
		url.JoinPath("sha512sum").String(),
		url.JoinPath("sha512sum").String(),
		url.JoinPath(reqFn + ".sha256").String(),
		url.JoinPath(reqFn + ".sha512").String(),
	}
}

func stringToLines(contents string) []string {
	contents = strings.TrimSpace(contents)
	contents = strings.ReplaceAll(contents, "\\n", "  # \n")
	contents = strings.ReplaceAll(contents, "\t", " ")

	lines := strings.Split(contents, "\\n")
	lines = strings.Split(lines[0], "\n")

	return lines
}
