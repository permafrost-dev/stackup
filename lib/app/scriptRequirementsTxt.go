package app

import (
	"bufio"
	"os"
	"strings"

	"github.com/stackup-app/stackup/lib/utils"
)

type RequirementsTxt struct {
	Requirements map[string]string
}

func LoadRequirementsTxt(filename string) (*RequirementsTxt, error) {
	reqs := &RequirementsTxt{
		Requirements: make(map[string]string),
	}

	if utils.IsDir(filename) {
		filename = filename + "/requirements.txt"
	}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return reqs, err
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore comments and empty lines
		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}

		// Split the line into a package name and version
		parts := strings.Split(line, "==")
		if len(parts) != 2 {
			return reqs, err
		}

		// Add the package name and version to the requirements map
		reqs.Requirements[parts[0]] = parts[1]
	}

	if err := scanner.Err(); err != nil {
		return reqs, err
	}

	return reqs, nil
}

func (reqs *RequirementsTxt) HasRequirement(name string) bool {
	_, ok := reqs.Requirements[name]
	return ok
}

func (reqs *RequirementsTxt) GetRequirement(name string) string {
	return reqs.Requirements[name]
}
