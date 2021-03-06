// Copyright 2018 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"context"
	"strings"
	"time"

	"github.com/palantir/policy-bot/pull"
)

type Methods struct {
	Comments     []string `yaml:"comments,omitempty"`
	GithubReview bool     `yaml:"github_review,omitempty"`

	// If GithubReview is true, GithubReviewState is the state a review must
	// have to be considered a candidated. It is currently excluded from
	// serialized forms and should be set by the application.
	GithubReviewState pull.ReviewState `yaml:"-" json:"-"`
}

type Candidate struct {
	Order        int
	User         string
	LastModified time.Time
}

type CandidatesByModifiedTime []*Candidate

func (cs CandidatesByModifiedTime) Len() int      { return len(cs) }
func (cs CandidatesByModifiedTime) Swap(i, j int) { cs[i], cs[j] = cs[j], cs[i] }
func (cs CandidatesByModifiedTime) Less(i, j int) bool {
	return cs[i].LastModified.Before(cs[j].LastModified)
}

func (m *Methods) Candidates(ctx context.Context, prctx pull.Context) ([]*Candidate, error) {
	var candidates []*Candidate

	if len(m.Comments) > 0 {
		comments, err := prctx.Comments()
		if err != nil {
			return nil, err
		}

		for _, c := range comments {
			if m.CommentMatches(c.Body) {
				candidates = append(candidates, &Candidate{
					Order:        c.Order,
					User:         c.Author,
					LastModified: c.LastModified,
				})
			}
		}
	}

	if m.GithubReview {
		reviews, err := prctx.Reviews()
		if err != nil {
			return nil, err
		}

		for _, r := range reviews {
			if r.State == m.GithubReviewState {
				candidates = append(candidates, &Candidate{
					Order:        r.Order,
					User:         r.Author,
					LastModified: r.LastModified,
				})
			}
		}
	}

	return candidates, nil
}

func (m *Methods) CommentMatches(commentBody string) bool {
	for _, comment := range m.Comments {
		if strings.Contains(commentBody, comment) {
			return true
		}
	}

	return false
}
