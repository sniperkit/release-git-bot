// Sniperkit - 2018
// Status: Analyzed

package ghclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
)

// Client is a github client used to get info from github.
type Client struct {
	owner string
	repo  string

	c *github.Client
}

// New creates a new client.
func New(tc *http.Client, owner, repo string) *Client {
	return &Client{
		owner: owner,
		repo:  repo,
		c:     github.NewClient(tc),
	}
}

// Owner returns the github user name this client was build with.
func (c *Client) Owner() string {
	return c.owner
}

// Repo returns the github repo name this client was build with.
func (c *Client) Repo() string {
	return c.repo
}

// GetMergedPRsForMilestone returns a list of github issues that are merged PRs
// for this milestone.
func (c *Client) GetMergedPRsForMilestone(milestone string) []*github.Issue {
	return c.getMergedPRsForMilestone(milestone)
}

// GetMergedPRsForLabels returns a list of github issues that are merged PRs
// with the given label.
func (c *Client) GetMergedPRsForLabels(labels []string) []*github.Issue {
	return c.getMergedPRsForLabels(labels)
}

// GetOrgMembers returns a set of names of members in the org.
func (c *Client) GetOrgMembers(org string) map[string]struct{} {
	return c.getOrgMembers(org)
}

// CommitIDForMergedPR returns the commit id for pr.
//
// It returns "" if pr is not a merged PR.
func (c *Client) CommitIDForMergedPR(pr *github.Issue) string {
	return c.commitIDForMergedPR(pr)
}

// NewBranchFromHead create a new branch with the current commit from head.
//
// It does nothing if the branch already exists.
func (c *Client) NewBranchFromHead(branchName string) error {
	log.Infof("creating branch: %v/%v/%v", c.owner, c.repo, branchName)
	ctx := context.Background()

	refName := "heads/" + branchName
	// Check if ref already exists.
	if ref, _, err := c.c.Git.GetRef(ctx, c.owner, c.repo, refName); err == nil {
		log.Infof("ref already exists: %v", ref)
		return nil
	}

	// Get head SHA.
	ref, _, err := c.c.Git.GetRef(ctx, c.owner, c.repo, "heads/master")
	if err != nil {
		return fmt.Errorf("failed to get master hash: %v", err)
	}
	log.Infof("hash for HEAD: %v", ref.GetObject().GetSHA())

	// Create new ref.
	newRef, _, err := c.c.Git.CreateRef(ctx, c.owner, c.repo, &github.Reference{
		Ref:    &refName,
		Object: ref.GetObject(),
	})
	if err != nil {
		return fmt.Errorf("failed to create ref: %v", err)
	}

	log.Infof("new ref created: %v", newRef.String())
	return nil
}

// NewPullRequest creates a pull request to the owner/repo pointed by this
// Client.
//
// headUser:headBranch specifies where the pull request is from.
func (c *Client) NewPullRequest(headUser, headBranch, base, title, body string) (string, error) {
	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(headUser + ":" + headBranch),
		Base:                github.String(base),
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := c.c.PullRequests.Create(context.Background(), c.owner, c.repo, newPR)
	if err != nil {
		return "", err
	}
	log.Infof("PR created: %s", pr.GetHTMLURL())
	return pr.GetHTMLURL(), nil
}

// NewDraftRelease creates a draft release.
func (c *Client) NewDraftRelease(tagName, targetBranch, title, body string) (string, error) {
	newRelease := &github.RepositoryRelease{
		TagName:         github.String(tagName),
		TargetCommitish: github.String(targetBranch),
		Name:            github.String(title),
		Body:            github.String(body),
		Draft:           github.Bool(true),
	}
	release, _, err := c.c.Repositories.CreateRelease(context.Background(), c.owner, c.repo, newRelease)
	if err != nil {
		return "", err
	}
	return release.GetHTMLURL(), nil
}

// GetPrimaryEmail returns the primary email of the token owner.
func (c *Client) GetPrimaryEmail() (string, error) {
	emails, _, err := c.c.Users.ListEmails(context.Background(), nil)
	if err != nil {
		return "", err
	}
	if len(emails) <= 0 {
		return "", fmt.Errorf("no email address found")
	}
	var e *github.UserEmail
	for _, e = range emails {
		if e.GetPrimary() {
			return e.GetEmail(), nil
		}
	}
	log.Warning("No primary email found, returning a random one")
	return e.GetEmail(), nil
}

// GetLogin returns the username of the token owner.
func (c *Client) GetLogin() (string, error) {
	// Passing the empty string will fetch the authenticated user.
	user, _, err := c.c.Users.Get(context.Background(), "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}
