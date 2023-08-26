package utils_test

import (
	"testing"
	"time"

	"github.com/stackup-app/stackup/lib/utils"
	"github.com/stretchr/testify/assert"
)

func TestFsSafeName(t *testing.T) {
	assert.Equal(t, "testbinary", utils.FsSafeName("testbinary"), "fs safe name should be 'testbinary'")
	assert.Equal(t, "testbinary", utils.FsSafeName("testbinary!$"), "fs safe name should be 'testbinary'")
	assert.Equal(t, "test-binary", utils.FsSafeName("test:binary!$"), "fs safe name should be 'test-binary'")
}

func TestGetUniqueStrings(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, utils.GetUniqueStrings([]string{"a", "b", "c", "a", "b", "c"}))
	assert.Equal(t, []string{"a", "b", "c"}, utils.GetUniqueStrings([]string{"a", "b", "c"}))
}

func TestGenerateTaskUuid(t *testing.T) {
	uid := utils.GenerateTaskUuid()
	assert.LessOrEqual(t, 8, len(uid))
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

func TestArrayContains(t *testing.T) {
	ptrArr := utils.Ptrs(1, 2, 3)
	//  []*int{utils.IntPtr(1), utils.IntPtr(2), utils.IntPtr(3)}
	assert.True(t, utils.ArrayContains(ptrArr, ptrArr[1]))
	// assert.True(t, utils.ArrayContains(ptrArr, utils.Ptr(3)))

	assert.True(t, utils.ArrayContains([]string{"a", "b", "c"}, "a"))
	assert.True(t, utils.ArrayContains([]int{1, 2, 3}, 3))
	assert.True(t, utils.ArrayContains([]bool{true, true}, true))
	assert.True(t, utils.ArrayContains([]float32{1, 2, 3}, float32(1)))
	assert.True(t, utils.ArrayContains([]string{"1a", "2a", "3a"}, []string{"1a", "2a"}))
	assert.True(t, utils.ArrayContains([]int{1, 2, 3}, []int{1, 2}))

	assert.True(t, utils.ArrayContains([]string{"a", "bbb", "c"}, "bbb"))
	assert.False(t, utils.ArrayContains([]string{"a", "b", "c"}, "d"))
	assert.False(t, utils.ArrayContains([]string{"a", "b", "c"}, ""))
	assert.False(t, utils.ArrayContains([]string{"a", "b", "c"}, "aaa"))
}

func TestMatchesPattern(t *testing.T) {
	assert.True(t, utils.MatchesPattern("abc.test", "^abc\\..+$"))
	assert.False(t, utils.MatchesPattern("abc.def", "^test.+$"))
}

func TestBinaryExistsInPath(t *testing.T) {
	assert.True(t, utils.BinaryExistsInPath("go"))
	assert.False(t, utils.BinaryExistsInPath("missing-binary-asdf-1234"))
}

func TestFileSize(t *testing.T) {
	assert.Equal(t, int64(0), utils.FileSize("missing.txt"))
	assert.Less(t, int64(1024), utils.FileSize("./utils.go")) // filesize is gt 1kb
}

func TestFileExists(t *testing.T) {
	assert.False(t, utils.FileExists("missing.txt"))
	assert.True(t, utils.FileExists("./utils.go"))
}

func TestWaitForStartOfNextInterval(t *testing.T) {
	d, _ := time.ParseDuration("1s")
	currentTime := time.Now().UnixMicro()
	utils.WaitForStartOfNextInterval(d)
	assert.GreaterOrEqual(t, time.Now().UnixMicro(), currentTime)
}

func TestSaveStringToFile(t *testing.T) {
	fn := "./test.txt"
	defer utils.RemoveFile(fn)

	err := utils.SaveStringToFile("test", fn)
	assert.NoError(t, err)
	assert.True(t, utils.FileExists(fn))
}

func TestReplaceFilenameInUrl(t *testing.T) {
	assert.Equal(t, "https://test.com/test2.txt", utils.ReplaceFilenameInUrl("https://test.com/test.txt", "test2.txt"))
	assert.Equal(t, "https://test.com", utils.ReplaceFilenameInUrl("https://test.com/test.txt", ""))
	assert.Equal(t, "https://test.com/a/b/c/test2.txt", utils.ReplaceFilenameInUrl("https://test.com/a/b/c/test.txt", "test2.txt"))
	assert.Equal(t, "https://test.com/a/b/c", utils.ReplaceFilenameInUrl("https://test.com/a/b/c/", ""))
	assert.Equal(t, "https://test.com/a/b", utils.ReplaceFilenameInUrl("https://test.com/a/b/c", ""))
	assert.Equal(t, "https://test.com/a/b/c/test.txt", utils.ReplaceFilenameInUrl("https://test.com/a/b/c/", "test.txt"))
}

func TestUrlBasePath(t *testing.T) {
	assert.Equal(t, "https://test.com", utils.UrlBasePath("https://test.com/test.txt"))
	assert.Equal(t, "https://test.com/a", utils.UrlBasePath("https://test.com/a/test.txt?a=1"))
	assert.Equal(t, "https://test.com/a/b/c", utils.UrlBasePath("https://test.com/a/b/c/test.txt"))
	assert.Equal(t, "https://test.com/a/b", utils.UrlBasePath("https://test.com/a/b/c/"))
	assert.Equal(t, "https://test.com/a/b", utils.UrlBasePath("https://test.com/a/b/c"))
}

func TestEnsureEnsureConfigDirExists(t *testing.T) {
	dirs := []string{utils.WorkingDir("test")}
	result, err := utils.EnsureConfigDirExists(dirs[0], "configtest")

	dirs = append(dirs, result)
	for _, dir := range dirs {
		defer utils.RemoveFile(dir)
	}

	assert.NotEmpty(t, result)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(dirs), 2)

	for _, dir := range dirs {
		assert.True(t, utils.PathExists(dir), "dir should exist: "+dir)
	}
}

func TestDomainGlobMatch(t *testing.T) {
	assert.True(t, utils.DomainGlobMatch("*.test", "abc.test"))
	assert.True(t, utils.DomainGlobMatch("*.test", "abc.def.test"))
	assert.False(t, utils.DomainGlobMatch("*.test", "abc.def"))
	assert.False(t, utils.DomainGlobMatch("*.test", "abc.def.ghi"))
}

func TestGlobMatch(t *testing.T) {
	assert.True(t, utils.GlobMatch("*.test", "abc.test", false))
	assert.True(t, utils.GlobMatch("**.test", "abc.test", true))
	assert.False(t, utils.GlobMatch("*test", "abc.def", false))
	assert.False(t, utils.GlobMatch("?\\*\test\\", "abc.def", false))
}

func TestEnforceSuffix(t *testing.T) {
	assert.Equal(t, "test.txt", utils.EnforceSuffix("test.txt", ".txt"))
	assert.Equal(t, "test.txt", utils.EnforceSuffix("test", ".txt"))
	assert.Equal(t, "test.txt.txt", utils.EnforceSuffix("test.txt.txt", ".txt"))
	assert.Equal(t, "test.txt", utils.EnforceSuffix("test.txt", ""))
}

func TestReverseArray(t *testing.T) {
	assert.Equal(t, []string{"c", "b", "a"}, utils.ReverseArray([]string{"a", "b", "c"}))
	assert.Equal(t, []int{3, 2, 1}, utils.ReverseArray([]int{1, 2, 3}))
}

func TestCombineArrays(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, utils.CombineArrays([]string{"a", "b"}, []string{"c"}))
	assert.Equal(t, []string{"a", "b", "c"}, utils.CombineArrays([]string{"a"}, []string{"b"}, []string{"c"}))
	assert.Equal(t, []int{1, 2, 3}, utils.CombineArrays([]int{1, 2}, []int{3}))
}

func TestMax(t *testing.T) {
	assert.Equal(t, 3, utils.Max(1, 2, 3))
	assert.Equal(t, 3, utils.Max(3, 2, 1))
	assert.Equal(t, 3, utils.Max(3))
	assert.Equal(t, 0, utils.Max())
}

func TestMin(t *testing.T) {
	assert.Equal(t, 1, utils.Min(1, 2, 3))
	assert.Equal(t, 1, utils.Min(3, 2, 1))
	assert.Equal(t, -3, utils.Min(1, 2, -3))
	assert.Equal(t, 3, utils.Min(3))
	assert.Equal(t, 0, utils.Min())
}

func TestCastAndCombineArrays(t *testing.T) {
	var t1 interface{} = "a"
	var t2 interface{} = "b"

	assert.Equal(t, []string{"a", "b", "c"}, utils.CastAndCombineArrays([]string{t1.(string), t2.(string)}, []interface{}{"c"}), "should combine arrays")
	assert.Equal(t, []string{"a", "b", "2"}, utils.CastAndCombineArrays([]string{"a", "b"}, []interface{}{"2"}))
	assert.Equal(t, []any{1, 2, "5"}, utils.CastAndCombineArrays([]interface{}{1, 2}, []interface{}{"5"}))
}

func TestGetUrlJson(t *testing.T) {
	var result interface{}

	err := utils.GetUrlJson("https://api.github.com/repos/permafrost-dev/stackup", &result, nil)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.github.com/repos/permafrost-dev/stackup", result.(map[string]interface{})["url"])
	assert.Equal(t, "https://api.github.com/repos/permafrost-dev/stackup/commits{/sha}", result.(map[string]interface{})["commits_url"])
}

func TestAbsoluteFilePath(t *testing.T) {
	assert.Equal(t, utils.WorkingDir("test.txt"), utils.AbsoluteFilePath("test.txt"))
	assert.Equal(t, utils.WorkingDir("test.txt"), utils.AbsoluteFilePath("./test.txt"))
	assert.Equal(t, utils.WorkingDir("/a/test.txt"), utils.AbsoluteFilePath("./a/test.txt"))
}

func TestRunCommandInPath(t *testing.T) {
	_, err := utils.RunCommandInPath("ls -la", ".", true)
	// _, _ := output.CombinedOutput()
	assert.NoError(t, err)
	// assert.Contains(t, string(str), "utils.go")

	// output, err = utils.RunCommandInPath("ls -la", ".", true)
	// pipe, _ = output.CombinedOutput()
	// str = string(pipe)
	// assert.NoError(t, err)
	// assert.Contains(t, str, "utils.go")
}

func TestExcept(t *testing.T) {
	assert.Equal(t, []string{"a", "c"}, utils.Except([]string{"a", "b", "c"}, []string{"b"}))
	assert.Equal(t, []string{"b"}, utils.Except([]string{"a", "b", "c"}, []string{"a", "c"}))
	assert.Equal(t, []string{}, utils.Except([]string{"a", "b", "c"}, []string{"a", "b", "c"}))
	assert.Equal(t, []string{"a", "b", "c"}, utils.Except([]string{"a", "b", "c"}, []string{"d", "e"}))
	assert.Equal(t, []string{"a", "b", "c"}, utils.Except([]string{"a", "b", "c"}, []string{}))
}

func TestOnly(t *testing.T) {
	assert.Equal(t, []string{"a", "c"}, utils.Only([]string{"a", "b", "c"}, []string{"a", "c"}))
	assert.Equal(t, []string{"a", "b", "c"}, utils.Only([]string{"a", "b", "c"}, []string{"a", "b", "c"}))
	assert.Equal(t, []string{}, utils.Only([]string{"a", "b", "c"}, []string{"d", "e"}))
	assert.Equal(t, []string{}, utils.Only([]string{"a", "b", "c"}, []string{}))
}
