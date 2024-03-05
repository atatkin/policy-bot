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

package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/palantir/policy-bot/policy"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
	"github.com/palantir/policy-bot/pull/simulated"
	"github.com/pkg/errors"
)

const (
	ignoreCommentsParam = "ignore_comments"
	addCommentsParam    = "add_comments"
)

// Simulate provides a baseline for handlers to perform simulated pull request evaluations and
// either return the result or display it in the ui.
type Simulate struct {
	Base
}

func getSimulatedOptions(r *http.Request) simulated.Options {
	var options simulated.Options
	if r.URL.Query().Has(ignoreCommentsParam) {
		options.IgnoreCommentsFrom = strings.Split(r.URL.Query().Get(ignoreCommentsParam), ",")
	}

	if r.URL.Query().Has(addCommentsParam) {
		options.AddApprovalCommentsFrom = strings.Split(r.URL.Query().Get(addCommentsParam), ",")
	}

	return options
}

func (h *Simulate) getApprovalResult(ctx context.Context, installation githubapp.Installation, loc pull.Locator, options simulated.Options) (*common.Result, error) {
	evalCtx, err := h.NewEvalContext(ctx, installation.ID, loc)
	switch {
	case err != nil:
		return nil, errors.Wrap(err, "failed to generate eval context")
	case evalCtx.Config.LoadError != nil:
		return nil, errors.Wrap(evalCtx.Config.LoadError, "failed to load policy file")
	case evalCtx.Config.ParseError != nil:
		return nil, errors.Wrap(evalCtx.Config.ParseError, "failed to parse policy")
	case evalCtx.Config.Config == nil:
		return nil, errors.New("no policy file found in repo")
	}

	evaluator, err := policy.ParsePolicy(evalCtx.Config.Config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get policy evaluator")
	}

	simulatedCtx := simulated.NewContext(evalCtx.PullContext, options)
	result := evaluator.Evaluate(ctx, simulatedCtx)
	if result.Error != nil {
		return nil, errors.Wrapf(err, "error evaluating policy in %s: %s", evalCtx.Config.Source, evalCtx.Config.Path)
	}

	return &result, nil
}
