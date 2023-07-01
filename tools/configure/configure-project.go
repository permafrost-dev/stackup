package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type GithubUser struct {
	Name string `json:"name"`
}

type GithubOrganization struct {
	Login                   string      `json:"login"`
	ID                      int         `json:"id"`
	NodeID                  string      `json:"node_id"`
	URL                     string      `json:"url"`
	ReposURL                string      `json:"repos_url"`
	EventsURL               string      `json:"events_url"`
	HooksURL                string      `json:"hooks_url"`
	IssuesURL               string      `json:"issues_url"`
	MembersURL              string      `json:"members_url"`
	PublicMembersURL        string      `json:"public_members_url"`
	AvatarURL               string      `json:"avatar_url"`
	Description             string      `json:"description"`
	Name                    string      `json:"name"`
	Company                 interface{} `json:"company"`
	Blog                    string      `json:"blog"`
	Location                string      `json:"location"`
	Email                   string      `json:"email"`
	TwitterUsername         string      `json:"twitter_username"`
	IsVerified              bool        `json:"is_verified"`
	HasOrganizationProjects bool        `json:"has_organization_projects"`
	HasRepositoryProjects   bool        `json:"has_repository_projects"`
	PublicRepos             int         `json:"public_repos"`
	PublicGists             int         `json:"public_gists"`
	Followers               int         `json:"followers"`
	Following               int         `json:"following"`
	HTMLURL                 string      `json:"html_url"`
	CreatedAt               time.Time   `json:"created_at"`
	UpdatedAt               time.Time   `json:"updated_at"`
	Type                    string      `json:"type"`
}

type GithubOrg struct {
	Login            string `json:"login"`
	ID               int    `json:"id"`
	NodeID           string `json:"node_id"`
	URL              string `json:"url"`
	ReposURL         string `json:"repos_url"`
	EventsURL        string `json:"events_url"`
	HooksURL         string `json:"hooks_url"`
	IssuesURL        string `json:"issues_url"`
	MembersURL       string `json:"members_url"`
	PublicMembersURL string `json:"public_members_url"`
	AvatarURL        string `json:"avatar_url"`
	Description      string `json:"description"`
}

func GetGithubOrganizationName(orgName string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s", orgName)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var org GithubOrganization
	err = json.Unmarshal(body, &org)
	if err != nil {
		return "", err
	}

	if org.Name == "" {
		return "", errors.New("no name found")
	}

	return org.Name, nil
}

func GetGithubUsernameFromGithubCli() (string, error) {
	cmd := exec.Command("gh", "api", "user", "--jq", ".login")

	var out strings.Builder
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

func getGithubUsernameFromGitRemote() (string, error) {
	out, err := gitCommand("config remote.origin.url")
	if err != nil {
		return "", err
	}

	remoteUrlParts := strings.Split(strings.Replace(strings.TrimSpace(out), ":", "/", -1), "/")
	return remoteUrlParts[1], nil
}

func getGitLogLines() ([]string, error) {
	//"--author='@users.noreply.github.com'",
	cmd := exec.Command("git", "log", "--pretty='%an:%ae'", "--reverse")

	var out strings.Builder
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(out.String(), "\n")

	return lines, nil
}

func searchCommitsForGithubUsername() (string, error) {
	out, err := gitCommand(`config user.name`)

	if out == "" {
		return "", err
	}

	authorName := strings.ToLower(strings.TrimSpace(out))

	lines, _ := getGitLogLines()

	type committer struct {
		name, email string
		username    string
	}

	committers := make([]committer, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		username := strings.Split(parts[1], "@")[0]
		committers = append(committers, committer{parts[0], parts[1], username})
	}

	committerList := make([]committer, 0)

	for _, committer := range committers {
		if strings.Contains(committer.name, "[bot]") {
			continue
		}

		if strings.EqualFold(committer.name, authorName) {
			committerList = append(committerList, committer)
		}
	}

	if len(committerList) == 0 {
		return "", nil
	}

	return committerList[0].username, nil
}

func guessGithubUsername() (string, error) {
	result, err := searchCommitsForGithubUsername()

	if err != nil {
		return "", err
	}

	if result != "" {
		return result, nil
	}

	result, _ = GetGithubUsernameFromGithubCli()

	if result != "" {
		return result, nil
	}

	result, err = getGithubUsernameFromGitRemote()

	if err != nil {
		return "", err
	}

	return result, nil
}

func gitCommand(args string) (string, error) {
	cmd := exec.Command("git", strings.Split(args, " ")...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func GetGithubUserName(username string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var user GithubUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		return "", err
	}

	if user.Name == "" {
		return "", errors.New("no name found")
	}

	return user.Name, nil
}

func getGithubUserFirstOrg(username string) (GithubOrg, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s/orgs", username))

	if err != nil {
		return GithubOrg{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return GithubOrg{}, err
	}

	var orgs []GithubOrg
	err = json.Unmarshal(body, &orgs)
	if err != nil {
		return GithubOrg{}, err
	}

	if len(orgs) > 0 {
		return orgs[0], nil
	}

	return GithubOrg{}, fmt.Errorf("no organizations found")
}

func GetGithubVendorUsername(username string) (string, error) {
	org, _ := getGithubUserFirstOrg(username)

	if org != (GithubOrg{}) {
		return org.Login, nil
	}

	output, err := exec.Command("git", "remote", "get-url", "origin").Output()

	if err != nil {
		return "", err
	}

	url := strings.Trim(string(output), " \t\r\n")

	re := regexp.MustCompile(`(?i)(?:github\.com[:/])([\w-]+/[\w-]+)`)

	matches := re.FindStringSubmatch(url)
	var result string

	if len(matches) > 1 {
		result = strings.Split(matches[1], "/")[0]
		orgName, err := GetGithubOrganizationName(result)

		if err == nil {
			result = orgName
		}
	} else {
		return "", errors.New("could not find github username")
	}

	return result, nil
}

func promptUserForInput(prompt string, defaultValue string) string {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		if defaultValue != "" {
			fmt.Printf("%s (%s) ", prompt, defaultValue)
		} else {
			fmt.Printf("%s ", prompt)
		}

		scanner.Scan()

		input := strings.TrimSpace(scanner.Text())

		if input == "" && defaultValue != "" {
			return defaultValue
		}

		if input != "" {
			return input
		}
	}
}

func stringInArray(str string, arr []string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func removeBetween(str, start, end string) string {
	for {
		s := strings.Index(str, start)
		e := strings.Index(str, end)
		// If start or end string is not found, return the original string
		if s == -1 || e == -1 {
			return str
		}
		// Remove text between start and end string
		str = str[:s] + str[e+len(end):]
	}
}

func processReadmeFile() {
	content, err := os.ReadFile("README.md")
	if err != nil {
		return
	}

	str := removeBetween(string(content), "<!-- ==START TEMPLATE README== -->", "<!-- ==END TEMPLATE README== -->")

	os.WriteFile("README.md", []byte(str), 0644)
}

func installGitHooks() {
	bytes, err := os.ReadFile(".git/config")
	if err != nil {
		return
	}

	content := string(bytes)

	if strings.Contains(string(content), "hooksPath") {
		return
	}

	content = strings.Replace(content, "[core]", "[core]\n\thooksPath = .custom-hooks", 1)

	os.WriteFile(".git/config", []byte(content), 0644)
}

func processDirectoryFiles(dir string, varMap map[string]string) {
	// get the files in the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
		return
	}

	ignoreFiles := []string{
		".git",
		".gitattributes",
		".gitignore",
		"configure-project.go",
		"build-all.go",
		"build-version.go",
		"go.sum",
	}

	// loop through the files
	for _, file := range files {
		if stringInArray(strings.ToLower(file.Name()), ignoreFiles) {
			continue
		}

		filePath := dir + "/" + file.Name()

		if file.IsDir() {
			processDirectoryFiles(filePath, varMap)
			continue
		}

		bytes, err := os.ReadFile(filePath)

		if err != nil {
			fmt.Println(err)
			continue
		}

		content := string(bytes)

		for key, value := range varMap {
			if file.Name() == "go.mod" {
				tempKey := strings.ReplaceAll(key, ".", "-")
				content = strings.ReplaceAll(content, "/"+tempKey, "/"+value)
				continue
			}

			key = "{{" + key + "}}"
			content = strings.ReplaceAll(content, key, value)
		}

		if string(bytes) != content {
			fmt.Printf("Updating file: %s\n", filePath)
			os.WriteFile(filePath, []byte(content), 0644)
		}
	}
}

func main() {
	// get the current directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	projectDir, err := filepath.Abs(cwd)

	if err != nil {
		fmt.Println(err)
		return
	}

	varMap := make(map[string]string)

	githubNameBytes, err := exec.Command("git", "config", "--global", "user.name").Output()
	if err != nil {
		githubNameBytes = []byte("")
	}

	githubEmailBytes, err := exec.Command("git", "config", "--global", "user.email").Output()
	if err != nil {
		githubEmailBytes = []byte("")
	}

	githubName := strings.Trim(string(githubNameBytes), " \r\n\t")
	githubEmail := strings.Trim(string(githubEmailBytes), " \r\n\t")
	githubUser, _ := guessGithubUsername()

	varMap["project.name.full"] = promptUserForInput("Project name: ", path.Base(projectDir))
	varMap["project.name"] = strings.ReplaceAll(varMap["project.name.full"], " ", "-")
	varMap["project.description"] = promptUserForInput("Project description: ", "")
	varMap["project.author.name"] = promptUserForInput("Your full name: ", githubName)
	varMap["project.author.email"] = promptUserForInput("Your email address: ", githubEmail)
	varMap["project.author.github"] = promptUserForInput("Your github username: ", githubUser)

	vendorUsername, _ := GetGithubVendorUsername(varMap["project.author.github"])
	varMap["project.vendor.github"] = promptUserForInput("User/org vendor github name: ", vendorUsername)

	vendorName, _ := GetGithubUserName(varMap["project.vendor.github"])
	varMap["project.vendor.name"] = promptUserForInput("User/org vendor name: ", vendorName)

	varMap["date.year"] = fmt.Sprintf("%d", time.Now().Local().Year())

	processDirectoryFiles(projectDir, varMap)
	processReadmeFile()

	// for key, value := range varMap {
	// 	fmt.Printf("varMap[%s]: %s\n", key, value)
	// }

	targetDir := projectDir + "/cmd/" + varMap["project.name"]
	os.MkdirAll(targetDir, 0755)
	os.WriteFile(targetDir+"/main.go", []byte("package main\n\n"), 0644)

	fmt.Println("Installing git hooks...")
	installGitHooks()

	fmt.Println("Done!")
}
