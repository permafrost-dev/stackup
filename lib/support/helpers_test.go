package support

import (
	"testing"
)

func TestFindExistingFile(t *testing.T) {
	filenames := []string{"helpers.go", "file2.txt", "file3.txt"}
	defaultFilename := "default.txt"

	// Positive test case: existing file found
	existingFile := FindExistingFile(filenames, defaultFilename)
	if existingFile != "helpers.go" {
		t.Errorf("Expected 'file1.txt', but got '%s'", existingFile)
	}

	// Negative test case: no existing file found, default filename used
	noExistingFile := FindExistingFile([]string{}, defaultFilename)
	if noExistingFile != defaultFilename {
		t.Errorf("Expected '%s', but got '%s'", defaultFilename, noExistingFile)
	}
}

func TestFindExistingFileWithEmptyFilenames(t *testing.T) {
	filenames := []string{}
	defaultFilename := "default.txt"

	// Negative test case: empty filenames slice
	emptyFilenames := FindExistingFile(filenames, defaultFilename)
	if emptyFilenames != defaultFilename {
		t.Errorf("Expected '%s', but got '%s'", defaultFilename, emptyFilenames)
	}
}

func TestFindExistingFileWithNoDefaultFilename(t *testing.T) {
	filenames := []string{"file1.txt", "file2.txt", "file3.txt"}

	// Negative test case: no default filename provided
	noDefaultFilename := FindExistingFile(filenames, "")
	if noDefaultFilename != "" {
		t.Errorf("Expected empty string, but got '%s'", noDefaultFilename)
	}
}

func TestFindExistingFileWithExistingDefaultFilename(t *testing.T) {
	filenames := []string{"file1.txt", "file2.txt", "file3.txt"}
	defaultFilename := "file2.txt"

	// Positive test case: existing default filename found
	existingDefaultFile := FindExistingFile(filenames, defaultFilename)
	if existingDefaultFile != defaultFilename {
		t.Errorf("Expected '%s', but got '%s'", defaultFilename, existingDefaultFile)
	}
}
