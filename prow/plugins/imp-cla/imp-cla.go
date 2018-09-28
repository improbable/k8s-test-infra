package imp_cla

import (
	"context"
	"github.com/pkg/errors"
	"k8s.io/test-infra/proto/cla"
	"regexp"

	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/plugins"
)

const (
	pluginName  = "imp-cla"
)

var (
	checkCLARegex = regexp.MustCompile(".* /check-cla .*")
)

type githubClient interface {
	AddLabel(owner, repo string, number int, label string) error
	CreateComment(owner, repo string, number int, comment string) error
	RemoveLabel(owner, repo string, number int, label string) error
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
}

func helpProvider(_ *plugins.Configuration, _ []string) (*pluginhelp.PluginHelp, error) {
	// The Config field is omitted because this plugin is not configurable.
	return &pluginhelp.PluginHelp{
		Description: "The imp-cla plugin checks whether the author of a PR is permitted to contribute to the respository (i.e. is an Improbable employee or has signed the Contributor License Agreement).",
	}, nil
}

func init() {
	conn, err := GetApiConnection("my service account file here")
	if err != nil {
		panic(errors.Wrap(err, "failed to get Improbable api connection"))
	}

	handler := &claStatusHandler{
		claService: improbable_cla.NewCLAServiceClient(conn),
	}

	plugins.RegisterPullRequestHandler(pluginName, handler.pullRequestHandler, helpProvider)
	plugins.RegisterReviewCommentEventHandler(pluginName, handler.commentHandler, helpProvider)
}

type claStatusHandler struct {
	claService improbable_cla.CLAServiceClient
}

func (c *claStatusHandler) pullRequestHandler(pc plugins.PluginClient, event github.PullRequestEvent) error {
	if event.Action != github.PullRequestActionOpened {
		return nil
	}
	return c.updateCLAStatus(event.Repo, event.PullRequest, pc.GitHubClient)
}

func (c *claStatusHandler) commentHandler(pc plugins.PluginClient, event github.ReviewCommentEvent) error {
	if event.Action != github.ReviewCommentActionCreated && event.Action != github.ReviewCommentActionEdited {
		return nil
	}

	if checkCLARegex.MatchString(event.Comment.Body) {
		return c.updateCLAStatus(event.Repo, event.PullRequest, pc.GitHubClient)
	}

	return nil
}

func (c *claStatusHandler) updateCLAStatus(repo github.Repo, pr github.PullRequest, github githubClient) error {
	org := repo.Owner.Login
	repoName := repo.Name
	prNumber := pr.Number
	author := pr.User.Login
	resp, err := c.claService.GetSignedCLA(context.TODO(), &improbable_cla.GetSignedCLARequest{
		GithubUsername: author,
	})
	if err != nil {
		return errors.Wrap(err, "failed to retrieve cla signed status")
	}
	if resp.HasSignedImprobableCLA {
		github.RemoveLabel(org, repoName, prNumber, "cla:no")
		github.RemoveLabel(org, repoName, prNumber, "do-not-merge/...")
		github.AddLabel(org, repoName, prNumber, "cla:yes")
	} else {
		github.RemoveLabel(org, repoName, prNumber,"cla:yes")
		github.AddLabel(org, repoName, prNumber,"cla:no")
		github.AddLabel(org, repoName, prNumber,"do-not-merge/...")
		github.CreateComment(org, repoName, prNumber,"please sign cla...")
	}

	return nil
}
