package projectinfo_test

import (
	"testing"

	"github.com/stackup-app/stackup/lib/projectinfo"
	"github.com/stretchr/testify/assert"
)

func TestProjectName(t *testing.T) {
	p := projectinfo.New("testbinary", "/tmp")
	assert.Equal(t, "testbinary", p.Name(), "project name should be 'testbinary'")
}

func TestProjectPath(t *testing.T) {
	p := projectinfo.New("testbinary", "/tmp")
	assert.Equal(t, "/tmp", p.Path(), "project path should be '/tmp'")
}

func TestProjectFsSafeName(t *testing.T) {
	p := projectinfo.New("testbinary", "/tmp")
	assert.Equal(t, "testbinary", p.FsSafeName(), "project fs safe name should be 'testbinary'")

	p = projectinfo.New("testbinary!$", "/tmp")
	assert.Equal(t, "testbinary", p.FsSafeName(), "project fs safe name should be 'testbinary'")
}
