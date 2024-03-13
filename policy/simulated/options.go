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
	"encoding/json"
	"net/http"
	"time"

	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
	"github.com/pkg/errors"
)

// Options should contain optional data that can be used to modify the results of the methods on the simulated Context.
type Options struct {
	IgnoreComments *Actors   `json:"ignore_comments"`
	IgnoreReviews  *Actors   `json:"ignore_reviews"`
	AddComments    []Comment `json:"add_comments"`
	AddReviews     []Review  `json:"add_reviews"`
	BaseBranch     string    `json:"base_branch"`
}

func NewOptionsFromRequest(r *http.Request) (Options, error) {
	var o Options
	if r.Body == nil {
		return o, nil
	}

	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		return o, errors.Wrap(err, "failed to unmarshal body into options")
	}

	o.setDefaults()
	return o, nil
}

// setDefaults sets any values for the options that were not intentionally set in the request body but which should have
// consistent values for the length of the simulation, such as the created time for a comment or review.
func (o *Options) setDefaults() {
	for i, review := range o.AddReviews {
		review.setDefaults()
		o.AddReviews[i] = review
	}

	for i, comment := range o.AddComments {
		comment.setDefaults()
		o.AddComments[i] = comment
	}
}

type Actors struct {
	Users         []string `json:"users"`
	Teams         []string `json:"teams"`
	Organizations []string `json:"organizations"`
	Permissions   []string `json:"permissions"`
}

func (a *Actors) toCommonActors() (common.Actors, error) {
	var permissions []pull.Permission
	for _, p := range a.Permissions {
		permission, err := pull.ParsePermission(p)
		if err != nil {
			return common.Actors{}, err
		}

		permissions = append(permissions, permission)
	}

	return common.Actors{
		Users:         a.Users,
		Teams:         a.Teams,
		Organizations: a.Organizations,
		Permissions:   permissions,
	}, nil
}

type Comment struct {
	CreatedAt    *time.Time `json:"created_at"`
	LastEditedAt *time.Time `json:"last_edited_at"`
	Author       string     `json:"author"`
	Body         string     `json:"body"`
}

// setDefaults sets the createdAt and lastEdtedAt values to time.Now() if they are otherwise unset
func (c *Comment) setDefaults() {
	now := time.Now()
	if c.CreatedAt == nil {
		c.CreatedAt = &now
	}

	if c.LastEditedAt == nil {
		c.LastEditedAt = &now
	}
}

func (c *Comment) toPullComment() *pull.Comment {
	return &pull.Comment{
		CreatedAt:    *c.CreatedAt,
		LastEditedAt: *c.LastEditedAt,
		Author:       c.Author,
		Body:         c.Body,
	}
}

type Review struct {
	CreatedAt    *time.Time `json:"created_at"`
	LastEditedAt *time.Time `json:"last_edited_at"`
	Author       string     `json:"author"`
	Body         string     `json:"body"`
	State        string     `json:"state"`
	Teams        []string   `json:"teams"`
}

// setDefaults sets the createdAt and lastEdtedAt values to time.Now() if they are otherwise unset
func (r *Review) setDefaults() {
	now := time.Now()
	if r.CreatedAt == nil {
		r.CreatedAt = &now
	}

	if r.LastEditedAt == nil {
		r.LastEditedAt = &now
	}
}

func (r *Review) toPullReview(id, sha string) *pull.Review {
	return &pull.Review{
		ID:           id,
		SHA:          sha,
		CreatedAt:    *r.CreatedAt,
		LastEditedAt: *r.LastEditedAt,
		Author:       r.Author,
		State:        pull.ReviewState(r.State),
		Body:         r.Body,
		Teams:        r.Teams,
	}
}
