/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package defaultlabels

import (
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/plugins"
)

const (
	// PluginName defines this plugin's registered name.
	PluginName = "default-labels"
)

func init() {
	plugins.RegisterPullRequestHandler(PluginName, handlePullRequest, helpProvider)
}

func helpProvider(config *plugins.Configuration, _ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
			Description: "The default-labels plugin automatically adds specified labels to PRs when they are opened.",
		},
		nil
}

type githubClient interface {
	AddLabel(org, repo string, number int, label string) error
	GetIssueLabels(org, repo string, number int) ([]github.Label, error)
	GetRepoLabels(owner, repo string) ([]github.Label, error)
}

func handlePullRequest(pc plugins.Agent, pre github.PullRequestEvent) error {
	if pre.Action != github.PullRequestActionOpened && pre.Action != github.PullRequestActionReopened {
		return nil
	}

	// Extract the default label set for the current repo, if one is set.
	defaultLabels := pc.PluginConfig.DefaultLabelsFor(pre.Repo.Owner.Login, pre.Repo.Name)
	neededLabels := sets.NewString()
	for _, label := range defaultLabels.Labels {
		neededLabels.Insert(label)
	}

	return handle(neededLabels, pc.GitHubClient, pc.Logger, &pre)
}

func handle(neededLabels sets.String, ghc githubClient, log *logrus.Entry, pre *github.PullRequestEvent) error {
	org := pre.Repo.Owner.Login
	repo := pre.Repo.Name
	number := pre.Number

	// Get the list of labels defined for the repo
	repoLabels, err := ghc.GetRepoLabels(org, repo)
	if err != nil {
		return err
	}
	repoLabelsExisting := sets.NewString()
	for _, label := range repoLabels {
		repoLabelsExisting.Insert(label.Name)
	}

	// Get the list of labels already applied to the PR
	issuelabels, err := ghc.GetIssueLabels(org, repo, number)
	if err != nil {
		return err
	}
	currentLabels := sets.NewString()
	for _, label := range issuelabels {
		currentLabels.Insert(label.Name)
	}

	nonexistent := sets.NewString()
	for _, labelToAdd := range neededLabels.Difference(currentLabels).List() {
		if !repoLabelsExisting.Has(labelToAdd) {
			nonexistent.Insert(labelToAdd)
			continue
		}
		if err := ghc.AddLabel(org, repo, number, labelToAdd); err != nil {
			log.WithError(err).Errorf("GitHub failed to add the following label: %s", labelToAdd)
		}
	}

	if nonexistent.Len() > 0 {
		log.Warnf("Unable to add nonexistent labels: %q", nonexistent.List())
	}
	return nil
}
