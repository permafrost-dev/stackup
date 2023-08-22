package app

import "github.com/stackup-app/stackup/lib/settings"

type IncludedTemplate struct {
	Name          string                  `yaml:"name"`
	Version       string                  `yaml:"version"`
	Checksum      string                  `yaml:"checksum"`
	LastModified  string                  `yaml:"last-modified"`
	Author        string                  `yaml:"author"`
	Description   string                  `yaml:"description"`
	Settings      *settings.Settings      `yaml:"settings"`
	Init          string                  `yaml:"init"`
	Tasks         []*Task                 `yaml:"tasks"`
	Preconditions []*WorkflowPrecondition `yaml:"preconditions"`
	Startup       []*TaskReference        `yaml:"startup"`
	Shutdown      []*TaskReference        `yaml:"shutdown"`
	Servers       []*TaskReference        `yaml:"servers"`
}
