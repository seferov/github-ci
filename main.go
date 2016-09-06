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

// Configuration jj
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

		// Protected files & folders
		if in(folder, []string{"post-clone.sh"}) {
			break
		}

		// if folder not found in open branches, delete it
		if !found {
			// delete
			fmt.Printf("Deleting " + folder + "...\n")
			os.RemoveAll(configuration.Directory + folder)
		}
	}

	var branchFolder string
	for _, branch := range openBranches {
		branchFolder = branchToFolderName(branch)
		if !in(branchFolder, clonedBranchFolders) {
			cloneBranch(branch)

			if configuration.Hooks.PostClone != "" {
				runHook(configuration.Hooks.PostClone, branchFolder)
			}
		}

		if configuration.Hooks.PostUpdate != "" {
			runHook(configuration.Hooks.PostUpdate, branchFolder)
		}
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
	fmt.Printf("Cloning " + branch + "...\n")

	cmd := exec.Command("git", "clone", "-b", branch, getCloneURL(), configuration.Directory+branchToFolderName(branch))
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func runHook(executable string, dir string) {
	fmt.Printf("/bin/sh " + executable + " " + configuration.Directory + dir + "\n")
	cmd := exec.Command("/bin/sh", executable, configuration.Directory+dir)
	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func getCloneURL() string {
	return "https://" + configuration.Github.Username + ":" + configuration.Github.Password + "@github.com/" + configuration.Github.Organization + "/" + configuration.Github.Repo + ".git"
}

func in(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
