// Farhad Safarov <farhad.safarov@gmail.com>

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/github"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	confFile      = kingpin.Arg("confFile", "Config file path").Required().File()
	configuration Configuration
)

// Configuration example: conf.json
type Configuration struct {
	Directory string `json:"directory"`
	Github    struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		Organization string `json:"organization"`
		Repo         string `json:"repository"`
		BaseBranch   string `json:"base_branch"`
	} `json:"github"`
	Hooks struct {
		PostClone  string `json:"post-clone"`
		PostUpdate string `json:"post-update"`
		PostClose  string `json:"post-close"`
	} `json:"hooks"`
}

// Gets configuration file
func init() {
	kingpin.Parse()

	decoder := json.NewDecoder(*confFile)
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func main() {
	openBranches := getOpenBranches()

	files, _ := ioutil.ReadDir(configuration.Directory)
	var clonedBranchFolders []string
	for _, f := range files {
		clonedBranchFolders = append(clonedBranchFolders, f.Name())
	}

	// Delete created and closed branch folders
	for _, folder := range clonedBranchFolders {
		found := false
		for _, branch := range openBranches {
			if branchToFolderName(branch) == folder {
				found = true
				break
			}
		}

		// if folder not found in open branches, delete it
		if !found {
			// delete
			fmt.Println("Deleting " + folder)
			os.RemoveAll(configuration.Directory + folder)

			runHook(configuration.Hooks.PostClose, folder)
		}
	}

	var branchFolder string
	for _, branch := range openBranches {
		fmt.Println(branch)
		branchFolder = branchToFolderName(branch)
		if !in(branchFolder, clonedBranchFolders) {
			cloneBranch(branch)

			runHook(configuration.Hooks.PostClone, branchFolder)
		}

		runHook(configuration.Hooks.PostUpdate, branchFolder)
	}
}

func getOpenBranches() []string {
	tp := github.BasicAuthTransport{
		Username: configuration.Github.Username,
		Password: configuration.Github.Password,
	}

	client := github.NewClient(tp.Client())

	pullRequestListOptions := &github.PullRequestListOptions{
		Base:  configuration.Github.BaseBranch,
		State: "open",
	}
	pulls, _, err := client.PullRequests.List(configuration.Github.Organization, configuration.Github.Repo, pullRequestListOptions)

	if err != nil {
		panic(err)
	}

	var openBranches []string
	for _, pull := range pulls {
		openBranches = append(openBranches, pullToBranchName(pull))
	}

	return openBranches
}

// Splits organization name from branch using github.PullRequest
func pullToBranchName(pr *github.PullRequest) string {
	branch := strings.Replace(github.Stringify(pr.Head.Label), configuration.Github.Organization+":", "", -1)
	branch = strings.Replace(branch, "\"", "", -1)

	return branch
}

func branchToFolderName(branch string) string {
	branch = strings.Replace(branch, "/", "-", -1)
	branch = strings.ToLower(branch)

	return branch
}

func cloneBranch(branch string) {
	fmt.Println("Cloning " + branch)

	cmdArgs := []string{"clone", "-b", branch, getCloneURL(), configuration.Directory + branchToFolderName(branch)}
	runCmd("git", cmdArgs)
}

func runHook(executable string, folder string) {
	if executable == "" {
		return
	}

	name := strings.Replace(folder, "-", "", -1)

	cmdArgs := []string{executable, configuration.Directory + folder, name}
	runCmd("/bin/sh", cmdArgs)
}

func runCmd(cmdName string, cmdArgs []string) {
	var (
		cmdOut []byte
		err    error
	)

	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running command: ", err)
	}

	fmt.Println(string(cmdOut))
}

func getCloneURL() string {
	return fmt.Sprintf("https://%s:%s@github.com/%s/%s.git",
		configuration.Github.Username,
		configuration.Github.Password,
		configuration.Github.Organization,
		configuration.Github.Repo,
	)
}

func in(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
