package projectinfo

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func extractRepositoryNameFromGitConfig(gitConfigPath string) (string, error) {
	file, err := os.Open(gitConfigPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inOriginSection := false
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "[remote \"origin\"]") {
			inOriginSection = true
			continue
		}

		if inOriginSection && strings.HasPrefix(line, "url =") {
			urlParts := strings.Split(line, "/")
			repoNameWithExtension := urlParts[len(urlParts)-1]
			repoName := strings.TrimSuffix(repoNameWithExtension, ".git")
			return repoName, nil
		}

		if inOriginSection && strings.HasPrefix(line, "[") {
			// Exit the origin section if another section starts
			inOriginSection = false
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("origin URL not found in %s", gitConfigPath)
}

// findProjectBaseDir searches the current directory for a directory named `.git`,
// and if not found, searches its parent directory recursively until `.git` is found or the root is reached.
func findProjectBaseDir(dirName string) (string, error) {
	dir, err := filepath.Abs(dirName)

	if err != nil {
		dir = dirName
	}

	// Check if the directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("directory does not exist: %s", dir)
	}

	// Check if .git directory exists in the current directory
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return dir, nil
	}

	// If reached the root, return an error
	if dir == "/" || dir == "" || path.Dir(dir) == "" {
		return "", fmt.Errorf(".git directory not found")
	}

	// Recursively search the parent directory
	parentDir := filepath.Dir(dir)

	return findProjectBaseDir(parentDir)
}

// Read a package.json file and returns the project name.
func getProjectNameFromPackageJson(packageJSONPath string) (string, error) {
	file, err := os.Open(packageJSONPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var packageData struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(file).Decode(&packageData); err != nil {
		return "", fmt.Errorf("error decoding package.json: %v", err)
	}

	if packageData.Name == "" {
		return "", fmt.Errorf("package.json does not contain a name")
	}

	return packageData.Name, nil
}

// getRepoNameFromGoMod reads a go.mod file and returns the repository name from the package name.
func getRepoNameFromGoMod(goModPath string) (string, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimPrefix(line, "module ")
			parts := strings.Split(modulePath, "/")
			if len(parts) > 1 {
				repoName := parts[len(parts)-1]
				return repoName, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("repository name not found in %s", goModPath)
}
