package projectinfo

import (
	"path/filepath"
	"regexp"
	"strings"
)

type Project struct {
	arg0 string
	cwd  string
}

func New(arg0, cwd string) *Project {
	return &Project{
		arg0: arg0,
		cwd:  cwd,
	}
}

func (p *Project) Name() string {
	binaryName := filepath.Base(p.arg0)

	if strings.Contains(binaryName, ".") {
		binaryName = strings.Split(binaryName, ".")[0]
	}

	return binaryName
}

func (p *Project) Path() string {
	cwd := p.cwd
	if cwd == "" {
		cwd = "."
	}

	if absCwd, err := filepath.Abs(cwd); err == nil {
		return absCwd
	}

	return cwd
}

func (p *Project) FsSafeName() string {
	result := strings.TrimSpace(p.Name())
	result = regexp.MustCompile(`[^\w\\-\\._]+`).ReplaceAllString(result, "-")
	result = regexp.MustCompile(`-{2,}`).ReplaceAllString(result, "-")

	return strings.Trim(result, "-")
}
