package commands

import (
	"encoding/json"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
)

func (c *GitCommand) GithubMostRecentPRs() ([]*models.GithubPullRequest, error) {
	commandOutput, err := c.OSCommand.RunCommandWithOutput("gh pr list --limit 50 --state all --json state,url,number,headRefName,headRepositoryOwner")
	if err != nil {
		return nil, err
	}

	prs := []*models.GithubPullRequest{}
	err = json.Unmarshal([]byte(commandOutput), &prs)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (c *GitCommand) GenerateGithubPullRequestMap(prs []*models.GithubPullRequest, branches []*models.Branch) (map[*models.Branch]*models.GithubPullRequest, bool) {
	res := map[*models.Branch]*models.GithubPullRequest{}

	if len(prs) == 0 {
		return res, false
	}

	remotesToOwnersMap, _ := c.GetRemotesToOwnersMap()
	if len(remotesToOwnersMap) == 0 {
		return res, false
	}

	foundBranchWithGithubPullRequest := false

	prWithStringKey := map[string]models.GithubPullRequest{}

	for _, pr := range prs {
		prWithStringKey[pr.UserName()+":"+pr.BranchName()] = *pr
	}

	for _, branch := range branches {
		if !branch.IsTrackingRemote() {
			continue
		}

		remoteAndName := strings.SplitN(branch.UpstreamName, "/", 2)
		owner, foundRemoteOwner := remotesToOwnersMap[remoteAndName[0]]
		if len(remoteAndName) != 2 || !foundRemoteOwner {
			continue
		}

		pr, hasPr := prWithStringKey[owner+":"+remoteAndName[1]]
		if !hasPr {
			continue
		}

		foundBranchWithGithubPullRequest = true

		res[branch] = &pr
	}

	return res, foundBranchWithGithubPullRequest
}
