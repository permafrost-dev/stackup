package scripting

import (
	"encoding/json"
	"os"

	"github.com/stackup-app/stackup/lib/utils"
)

type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Keywords        []string          `json:"keywords"`
	License         string            `json:"license"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

func LoadPackageJson(filename string) (*PackageJSON, error) {
	pkg := &PackageJSON{}

	if utils.IsDir(filename) {
		filename = filename + "/package.json"
	}

	contents, err := os.ReadFile(filename)
	if err != nil {
		return pkg, err
	}

	err = json.Unmarshal(contents, &pkg)
	if err != nil {
		return pkg, err
	}

	return pkg, nil
}

func (pkg *PackageJSON) HasDependency(name string) bool {
	_, ok := pkg.Dependencies[name]
	return ok
}

func (pkg *PackageJSON) GetDependencies() []string {
	dependencies := []string{}
	for name := range pkg.Dependencies {
		dependencies = append(dependencies, name)
	}
	return dependencies
}

func (pkg *PackageJSON) HasDevDependency(name string) bool {
	_, ok := pkg.DevDependencies[name]
	return ok
}

func (pkg *PackageJSON) GetDevDependencies() []string {
	dependencies := []string{}
	for name := range pkg.DevDependencies {
		dependencies = append(dependencies, name)
	}
	return dependencies
}

func (pkg *PackageJSON) HasScript(name string) bool {
	_, ok := pkg.Scripts[name]
	return ok
}

func (pkg *PackageJSON) GetScript(name string) string {
	return pkg.Scripts[name]
}
