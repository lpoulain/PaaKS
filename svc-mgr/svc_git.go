package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/lpoulain/PaaKS/paaks"
)

func gitCreateService(serviceName string, cmd *GitCommand) error {
	fmt.Printf("Create service off repo [%s], branch [%s]\n", cmd.Command, cmd.Parameter)

	_, err := git.PlainClone("/tmp/storage/"+serviceName+"/", false, &git.CloneOptions{
		URL:           cmd.Command,
		ReferenceName: plumbing.NewBranchReferenceName(cmd.Parameter),
		Progress:      os.Stdout,
	})
	if err != nil {
		fmt.Printf("Error creating Git repo: %s\n", err)
		return err
	}

	return nil
}

func gitCommand(w http.ResponseWriter, r *http.Request, serviceName string, tenant string, cmd *GitCommand) {
	switch cmd.Command {
	case "branch":
		gitBranch(w, tenant, serviceName, cmd.Parameter)
		return
	case "pull":
		gitPull(w, tenant, serviceName)
		return
	default:
		paaks.IssueError(w, "Invalid command: "+cmd.Command, http.StatusBadRequest)
	}
}

func gitBranch(w http.ResponseWriter, tenant string, serviceName string, branch string) {
	repo, wt := getGitRepo(w, tenant, serviceName)
	if repo == nil {
		return
	}

	b, err := repo.Branch(branch)

	if b == nil {
		fmt.Printf("Branch not existing, creating a new one...\n")
		// we want to create a branch 'issues/166' that'd track origin/master
		var remote = "origin"

		// we resolve origin/master to a hash
		var remoteRef = plumbing.NewRemoteReferenceName(remote, branch)
		var ref, _ = repo.Reference(remoteRef, true)

		// create a new "tracking config"
		var mergeRef = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
		_ = repo.CreateBranch(&config.Branch{Name: branch, Remote: remote, Merge: mergeRef})

		// and finally create an "actual branch"
		var localRef = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
		_ = repo.Storer.SetReference(plumbing.NewHashReference(localRef, ref.Hash()))
	}

	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		fmt.Printf("Error switching Git branch: %s\n", err)
		return
	}
}

func gitPull(w http.ResponseWriter, tenant string, serviceName string) {
	repo, wt := getGitRepo(w, tenant, serviceName)
	if repo == nil {
		return
	}

	err := wt.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		paaks.IssueError(w, "Error pulling: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func getGitRepo(w http.ResponseWriter, tenant string, serviceName string) (*git.Repository, *git.Worktree) {
	repo, err := git.PlainOpen(fmt.Sprintf("/tmp/storage/tnt-%s-%s", tenant[:8], serviceName))
	if err != nil {
		paaks.IssueError(w, "Error opening Git repo", http.StatusInternalServerError)
		return nil, nil
	}

	wt, err := repo.Worktree()
	if err != nil {
		paaks.IssueError(w, "Error getting Git worktree: "+err.Error(), http.StatusInternalServerError)
		return nil, nil
	}

	return repo, wt
}
