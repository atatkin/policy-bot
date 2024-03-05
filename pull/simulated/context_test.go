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

package simulated

import (
	"testing"

	"github.com/palantir/policy-bot/pull"
	"github.com/stretchr/testify/assert"
)

func TestComments(t *testing.T) {
	tests := map[string]struct {
		Comments         []*pull.Comment
		Options          Options
		FilteredComments []*pull.Comment
	}{
		"ignore comments by iignore": {
			Comments: []*pull.Comment{
				{Author: "rrandom"},
				{Author: "iignore"},
			},
			FilteredComments: []*pull.Comment{
				{Author: "rrandom"},
			},
			Options: Options{
				IgnoreCommentsFrom: []string{"iignore"},
			},
		},
		"do not ignore any comments": {
			Comments: []*pull.Comment{
				{Author: "rrandom"},
				{Author: "iignore"},
			},
			FilteredComments: []*pull.Comment{
				{Author: "rrandom"},
				{Author: "iignore"},
			},
		},
		"ignore all comments": {
			Comments: []*pull.Comment{
				{Author: "rrandom"},
				{Author: "iignore"},
			},
			Options: Options{
				IgnoreCommentsFrom: []string{"iignore", "rrandom"},
			},
		},
		"add new comment by sperson": {
			Comments: []*pull.Comment{
				{Author: "rrandom"},
				{Author: "iignore"},
			},
			Options: Options{
				AddApprovalCommentsFrom: []string{"sperson"},
			},
			FilteredComments: []*pull.Comment{
				{Author: "rrandom"},
				{Author: "iignore"},
				{Author: "sperson"},
			},
		},
		"add new comment by sperson and ignore one from iignore": {
			Comments: []*pull.Comment{
				{Author: "rrandom"},
				{Author: "iignore"},
			},
			Options: Options{
				IgnoreCommentsFrom:      []string{"iignore"},
				AddApprovalCommentsFrom: []string{"sperson"},
			},
			FilteredComments: []*pull.Comment{
				{Author: "rrandom"},
				{Author: "sperson"},
			},
		},
	}

	for message, test := range tests {
		context := Context{
			Context: &testPullContext{comments: test.Comments},
			options: test.Options,
		}

		comments, err := context.Comments()
		assert.NoError(t, err, test, message)
		assert.Equal(t, authors(test.FilteredComments), authors(comments), message)
	}
}

func authors(comments []*pull.Comment) []string {
	var authors []string
	for _, c := range comments {
		authors = append(authors, c.Author)
	}

	return authors
}

type testPullContext struct {
	pull.Context
	comments []*pull.Comment
}

func (c *testPullContext) Comments() ([]*pull.Comment, error) {
	return c.comments, nil
}
