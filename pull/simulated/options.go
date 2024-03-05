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
	"time"

	"github.com/palantir/policy-bot/pull"
)

// Options should contain optional data that can be used to modify the results of the methods on the simulated Context.
type Options struct {
	IgnoreCommentsFrom      []string
	AddApprovalCommentsFrom []string
}

func (o *Options) filterIgnoredComments(comments []*pull.Comment) []*pull.Comment {
	if len(o.IgnoreCommentsFrom) <= 0 {
		return comments
	}

	ignoreCommentFromMap := make(map[string]bool)
	for _, name := range o.IgnoreCommentsFrom {
		ignoreCommentFromMap[name] = true
	}

	var filteredComments []*pull.Comment
	for _, comment := range comments {
		if ignoreCommentFromMap[comment.Author] {
			continue
		}

		filteredComments = append(filteredComments, comment)
	}

	return filteredComments
}

func (o *Options) addApprovalComments(comments []*pull.Comment) []*pull.Comment {
	for _, author := range o.AddApprovalCommentsFrom {
		comments = append(comments, &pull.Comment{
			CreatedAt:    time.Now(),
			LastEditedAt: time.Now(),
			Author:       author,
			Body:         ":+1:",
		})
	}

	return comments
}
