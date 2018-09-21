package prow_testing

import (
	"context"
	"github.com/google/go-github/github"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"golang.org/x/oauth2"
)

// I dumped this code during a hand-off. It's not productionised at all and will likely need significant changes before it can be merged.
// Apologies for any difficulties understanding it.

const (
	oauthTokenFile     = "PATH-TO-GITHUB-PERSONAL-OAUTH-TOKEN"
  organization       = "NAME OF GITHUB ORGANISATION TO TEST ON"
	repository         = "NAME OF GITHUB REPOSITORY TO TEST ON"
	existingCodeBranch = "NAME OF BRANCH TO OPEN PRS TO TEST ON ETC"
	expectedReviewer   = "REVIEWER WHO SHOULD BE ADDED"
)


func TestGithubOauth(t *testing.T) {
	ctx := context.Background()
	client := getGithubClient(ctx, t)

	resp, _, err := client.Octocat(ctx, "Hello Test")

	if err != nil {
		t.Error(err)
		t.Fail()
		t.Fatal()
	}

	t.Log(resp)
}

func TestXSmallCommitsGetLabelled(t *testing.T) {
	ctx := context.Background()
	client := getGithubClient(ctx, t)

	// Given we have a clean pull request
	closeExistingPullRequests(t, ctx, client)
	id := openNewPullRequest(t, ctx, client)

	// Then, if we wait a few seconds
	time.Sleep(10 * time.Second)

	// Then it should be labelled with size/XS
	assert.True(t, containsLabel(client, ctx, id, t, "size/XS"))
}

func TestOwnerGetsAddedAsReviewer(t *testing.T) {
	ctx := context.Background()
	client := getGithubClient(ctx, t)

	// Given we have a clean pull request
	closeExistingPullRequests(t, ctx, client)
	id := openNewPullRequest(t, ctx, client)

	// Then, if we wait a few seconds
	time.Sleep(10 * time.Second)

	// Then it should have the assigned reviewer as a reviewer
	assert.True(t, containsReviewer(client, ctx, id, t, expectedReviewer))
}

func getGithubClient(ctx context.Context, t *testing.T) *github.Client {
	oauthTokenB, err := ioutil.ReadFile(oauthTokenFile)
	if err != nil {
		t.Fatal("Unable to get GitHub oauth token from filesystem using path ", oauthTokenFile)
	}
	oauthToken := strings.TrimSpace(string(oauthTokenB))

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: oauthToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client
}

func containsReviewer(client *github.Client, ctx context.Context, id int, t *testing.T, expectedReviewer string) bool {
	pullReq := getPullReq(client, ctx, id, t)
	matched := false
	for _, reviewer := range pullReq.RequestedReviewers {
		matched = matched || reviewer.GetLogin() == expectedReviewer
	}
	return matched
}

func containsLabel(client *github.Client, ctx context.Context, id int, t *testing.T, expectedLabel string) bool {
	pullReq := getPullReq(client, ctx, id, t)
	matched := false
	for _, label := range pullReq.Labels {
		matched = matched || label.GetName() == expectedLabel
	}
	return matched
}

func getPullReq(client *github.Client, ctx context.Context, id int, t *testing.T) *github.PullRequest {
	pullReq, _, err := client.PullRequests.Get(ctx, organization, repository, id)
	if err != nil {
		t.Fatal("Unable to get pull request ", id)
	}
	return pullReq
}

func openNewPullRequest(t *testing.T, ctx context.Context, client *github.Client) int {
	title, body, base, head := "New pull request", "", "master", existingCodeBranch
	newPullReq := &github.NewPullRequest{
		Title: &title,
		Body: &body,
		Base: &base,
		Head: &head,
	}
	pullReq, _, err := client.PullRequests.Create(ctx, organization, repository, newPullReq)
	if err != nil {
		t.Fatal(err)
	}

	return *pullReq.Number
}

func closeExistingPullRequests(t *testing.T, ctx context.Context, client *github.Client) {
	requests, _, err := client.PullRequests.List(ctx, organization, repository, &github.PullRequestListOptions{
		State: "open",
	})
	if err != nil {
		t.Fatal("Unable to get repository")
	}

	for _, pullReq := range requests {
		closed := "closed"

		pullReq.State = &closed
		_, _, err = client.PullRequests.Edit(ctx, organization, repository, *pullReq.Number, pullReq)
		if err != nil {
			t.Fatal("Unable to close pull request", err)
		}
	}
}
