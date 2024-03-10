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
	"bytes"
	"net/http"
	"testing"

	"github.com/palantir/policy-bot/pull"
	"github.com/stretchr/testify/assert"
)

func TestOptionsFromRequest(t *testing.T) {
	body := `
	{
		"ignore_comments":{
			"users":["iignore"],
			"teams":[""],
			"organizations":[""],
			"permissions":[""]
		},
		"ignore_reviews":{
			"users":["iignore"],
			"teams":[""],
			"organizations":[""],
			"permissions":[""]
		},
		"add_comments":[
			{"author":"iignore", "body":":+1:"}
		],
		"add_reviews":[
			{"author":"iignore", "body":":+1:", "state":"approved"}
		],
		"base_branch":"test-base"
	}`

	req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer([]byte(body)))
	assert.NoError(t, err)

	opt, err := NewOptionsFromRequest(req)
	assert.NoError(t, err)

	assert.Equal(t, []string{"iignore"}, opt.IgnoreComments.Users)
	assert.Equal(t, []string{"iignore"}, opt.IgnoreReviews.Users)

	assert.Equal(t, "iignore", opt.AddComments[0].Author)
	assert.Equal(t, ":+1:", opt.AddComments[0].Body)

	assert.Equal(t, "iignore", opt.AddReviews[0].Author)
	assert.Equal(t, ":+1:", opt.AddReviews[0].Body)

	assert.Equal(t, pull.ReviewApproved, opt.AddReviews[0].toPullReview("test-id", "test-sha").State)
	assert.Equal(t, "test-base", opt.BaseBranch)
}

func TestOptionDefaults(t *testing.T) {
	options := Options{
		AddComments: []Comment{
			{Author: "aperson", Body: ":+1:"},
			{Author: "otherperson", Body: ":+1:"},
		},
		AddReviews: []Review{
			{Author: "aperson", Body: ":+1:"},
			{Author: "otherperson", Body: ":+1:"},
		},
	}

	options.setDefaults()
	for _, comment := range options.AddComments {
		assert.False(t, comment.CreatedAt.IsZero())
		assert.False(t, comment.LastEditedAt.IsZero())
	}

	for _, review := range options.AddReviews {
		assert.False(t, review.CreatedAt.IsZero())
		assert.False(t, review.LastEditedAt.IsZero())
	}
}
