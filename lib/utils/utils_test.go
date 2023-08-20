package utils_test

import (
	"testing"

	"github.com/stackup-app/stackup/lib/utils"
	"github.com/stretchr/testify/assert"
)

func TestFsSafeName(t *testing.T) {
	assert.Equal(t, "testbinary", utils.FsSafeName("testbinary"), "fs safe name should be 'testbinary'")
	assert.Equal(t, "testbinary", utils.FsSafeName("testbinary!$"), "fs safe name should be 'testbinary'")
	assert.Equal(t, "test-binary", utils.FsSafeName("test:binary!$"), "fs safe name should be 'test-binary'")
}

func TestGlobMatch(t *testing.T) {
	assert.True(t, utils.GlobMatch("*.test", "abc.test", false))
	assert.True(t, utils.GlobMatch("**.test", "abc.test", true))
	assert.False(t, utils.GlobMatch("*test", "abc.def", false))
}

func TestGetUniqueStrings(t *testing.T) {
	arr := utils.GetUniqueStrings([]string{"a", "b", "c", "a", "b", "c"})
	assert.Equal(t, []string{"a", "b", "c"}, arr)
}

func TestGenerateShortID(t *testing.T) {
	id := utils.GenerateShortID()
	assert.Equal(t, 8, len(id))

	id = utils.GenerateShortID(12)
	assert.Equal(t, 12, len(id))
}

func TestStringArrayContains(t *testing.T) {
	arr := []string{"a", "b", "c"}
	assert.True(t, utils.StringArrayContains(arr, "a"))
	assert.False(t, utils.StringArrayContains(arr, "d"))
}

func TestMatchesPattern(t *testing.T) {
	assert.True(t, utils.MatchesPattern("abc.test", "^abc\\..+$"))
	assert.False(t, utils.MatchesPattern("abc.def", "^test.+$"))
}

func TestBinaryExistsInPath(t *testing.T) {
	assert.True(t, utils.BinaryExistsInPath("go"))
	assert.False(t, utils.BinaryExistsInPath("missing-binary-asdf-1234"))
}
