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
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/github/fakegithub"
	"k8s.io/test-infra/prow/labels"
)

func formatLabels(labels ...string) []string {
	r := []string{}
	for _, l := range labels {
		r = append(r, formatSingleLabel(l))
	}
	if len(r) == 0 {
		return nil
	}
	return r
}

func formatSingleLabel(label string) string {
	return fmt.Sprintf("%s/%s#%d:%s", "org", "repo", 1, label)
}

// TestHandle tests that the handle function applies the correct default labels to a PR.
func TestHandle(t *testing.T) {
	type testCase struct {
		name              string
		expectedNewLabels []string
		neededLabels      sets.String
		repoLabels        []string
		issueLabels       []string
	}
	testcases := []testCase{
		{
			name:              "no labels",
			expectedNewLabels: []string{},
			neededLabels: sets.NewString(
				labels.LGTM,
			),
			repoLabels:  []string{},
			issueLabels: []string{},
		},
		{
			name:              "new valid label",
			expectedNewLabels: formatLabels(labels.LGTM),
			neededLabels:      sets.NewString(labels.LGTM),
			repoLabels:        []string{labels.LGTM},
			issueLabels:       []string{},
		},
		{
			name:              "new invalid label (not defined for repo)",
			expectedNewLabels: formatLabels(),
			neededLabels:      sets.NewString(labels.LGTM),
			repoLabels:        []string{},
			issueLabels:       []string{},
		},
		{
			name:              "existing valid label (already attached to PR)",
			expectedNewLabels: formatLabels(),
			neededLabels:      sets.NewString(labels.LGTM),
			repoLabels:        []string{labels.LGTM},
			issueLabels:       []string{labels.LGTM},
		},
		{
			name:              "multiple new valid labels",
			expectedNewLabels: formatLabels(labels.LGTM, labels.Approved),
			neededLabels:      sets.NewString(labels.LGTM, labels.Approved),
			repoLabels:        []string{labels.LGTM, labels.Approved},
			issueLabels:       []string{},
		},
		{
			name:              "mixed valid and invalid labels",
			expectedNewLabels: formatLabels(labels.LGTM),
			neededLabels:      sets.NewString(labels.LGTM, labels.Approved),
			repoLabels:        []string{labels.LGTM},
			issueLabels:       []string{},
		},
		{
			name:              "mixed new and existing labels",
			expectedNewLabels: formatLabels(labels.LGTM),
			neededLabels:      sets.NewString(labels.LGTM, labels.Approved),
			repoLabels:        []string{labels.LGTM, labels.Approved, "kind/docs"},
			issueLabels:       []string{labels.Approved, "kind/docs"},
		},
		{
			name:              "mixed new and existing valid, invalid labels",
			expectedNewLabels: formatLabels(labels.LGTM),
			neededLabels:      sets.NewString(labels.LGTM, labels.Approved, "kind/docs"),
			repoLabels:        []string{labels.LGTM, labels.Approved},
			issueLabels:       []string{labels.Approved},
		},
		{
			name:              "new valid label, existing unmanaged labels",
			expectedNewLabels: formatLabels(labels.Approved),
			neededLabels:      sets.NewString(labels.LGTM, labels.Approved),
			repoLabels:        []string{labels.LGTM, labels.Approved, "kind/docs"},
			issueLabels:       []string{labels.LGTM, "kind/docs"},
		},
		{
			name:              "new invalid label, existing unmanaged labels",
			expectedNewLabels: formatLabels(),
			neededLabels:      sets.NewString("kind/docs"),
			repoLabels:        []string{labels.LGTM, labels.Approved},
			issueLabels:       []string{labels.LGTM},
		},
		{
			name:              "existing valid labels, existing unmanaged labels",
			expectedNewLabels: formatLabels(),
			neededLabels:      sets.NewString(labels.LGTM, "kind/docs"),
			repoLabels:        []string{labels.LGTM, labels.Approved, "kind/docs"},
			issueLabels:       []string{labels.LGTM, labels.Approved, "kind/docs"},
		},
	}

	for _, tc := range testcases {
		basicPR := github.PullRequest{
			Number: 1,
			Base: github.PullRequestBranch{
				Repo: github.Repo{
					Owner: github.User{
						Login: "org",
					},
					Name: "repo",
				},
			},
			User: github.User{
				Login: "user",
			},
		}

		t.Logf("Running scenario %q", tc.name)
		sort.Strings(tc.expectedNewLabels)

		fghc := &fakegithub.FakeClient{
			PullRequests: map[int]*github.PullRequest{
				basicPR.Number: &basicPR,
			},
			PullRequestChanges: map[int][]github.PullRequestChange{},
			RepoLabelsExisting: tc.repoLabels,
			IssueLabelsAdded:   []string{},
		}
		// Add initial labels
		for _, label := range tc.issueLabels {
			_ = fghc.AddLabel(basicPR.Base.Repo.Owner.Login, basicPR.Base.Repo.Name, basicPR.Number, label)
		}
		pre := &github.PullRequestEvent{
			Action:      github.PullRequestActionOpened,
			Number:      basicPR.Number,
			PullRequest: basicPR,
			Repo:        basicPR.Base.Repo,
		}

		err := handle(tc.neededLabels, fghc, logrus.WithField("plugin", PluginName), pre)
		if err != nil {
			t.Errorf("[%s] unexpected error from handle: %v", tc.name, err)
			continue
		}

		// Check that all the correct labels (and only the correct labels) were added.
		expectLabels := append(formatLabels(tc.issueLabels...), tc.expectedNewLabels...)
		if expectLabels == nil {
			expectLabels = []string{}
		}
		sort.Strings(expectLabels)
		sort.Strings(fghc.IssueLabelsAdded)
		if !reflect.DeepEqual(expectLabels, fghc.IssueLabelsAdded) {
			t.Errorf("FAIL: test case %q expected the labels %q to be added, but %q were added.", tc.name, expectLabels, fghc.IssueLabelsAdded)
		}

	}
}
