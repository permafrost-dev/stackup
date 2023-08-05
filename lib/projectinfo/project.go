package projectinfo

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Project struct {
	cachedName *string
	cachedPath *string
}

func New() *Project {
	return &Project{}
}

func NewNamed(name string) *Project {
	binDir, _ := filepath.Abs(path.Dir(os.Args[0]))
	return &Project{cachedName: &name, cachedPath: &binDir}
}

func (p *Project) ClearCache() {
	p.cachedName = nil
	p.cachedPath = nil
}

func (p *Project) cacheName(value string) string {
	p.cachedName = &value
	return *p.cachedName
}

func (p *Project) cachePath(value string) string {
	p.cachedPath = &value
	return *p.cachedPath
}

func (p *Project) Name() string {
	if p.cachedName != nil && *p.cachedName != "" {
		return *p.cachedName
	}

	var cwd string
	var err error

	if cwd, err = os.Getwd(); err != nil {
		cwd = "."
	}

	baseDir, err := findProjectBaseDir(cwd)

	fmt.Printf("baseDir: %s\n", baseDir)

	if err != nil {
		if absCwd, err := filepath.Abs(cwd); err == nil {
			return p.cacheName(path.Base(absCwd))
		}
		return p.cacheName(path.Base(cwd))
	}

	if nameFromGit, err := extractRepositoryNameFromGitConfig(path.Join(baseDir, ".git/config")); err == nil {
		return p.cacheName(nameFromGit)
	}

	if packageName, err := getProjectNameFromPackageJson(path.Join(baseDir, "package.json")); err == nil {
		return p.cacheName(packageName)
	}

	if moduleName, err := getRepoNameFromGoMod(path.Join(baseDir, "go.mod")); err == nil {
		return p.cacheName(moduleName)
	}

	return p.cacheName(path.Base(baseDir))
}

func (p *Project) Path() string {
	if p.cachedPath != nil && *p.cachedPath != "" {
		return *p.cachedPath
	}

	var cwd string
	var err error

	if cwd, err = os.Getwd(); err != nil {
		cwd = "."
	}

	baseDir, err := findProjectBaseDir(cwd)
	if err != nil {
		if absCwd, err := filepath.Abs(cwd); err == nil {
			return p.cachePath(absCwd)
		}
		return p.cachePath(cwd)
	}

	if absBaseDir, err := filepath.Abs(baseDir); err == nil {
		return p.cachePath(absBaseDir)
	}

	return p.cachePath(baseDir)
}

func (p *Project) FsSafeName() string {
	result := p.Name()

	result = regexp.MustCompile(`[^\w]+`).ReplaceAllString(result, "-")
	result = regexp.MustCompile(`-{2,}`).ReplaceAllString(result, "-")
	result = strings.Trim(result, "-")

	return result
}
