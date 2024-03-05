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

	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/palantir/policy-bot/policy"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/policy/simulator"
	"github.com/palantir/policy-bot/pull"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	paramIgnoreCommentsFrom = "ignore_comments_from"
)

type Simulate struct {
	Base
}

type SimulatedResultResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

func (h *Simulate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := *zerolog.Ctx(ctx)
	var response SimulatedResultResponse

	owner, repo, number, ok := parsePullParams(r)
	if !ok {
		logger.Error().Msg("failed to parse pull request parameters from request")
		baseapp.WriteJSON(w, http.StatusBadRequest, &response)
		return
	}

	installation, err := h.Installations.GetByOwner(ctx, owner)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get installation for org")
		baseapp.WriteJSON(w, http.StatusNotFound, &response)
		return
	}

	client, err := h.ClientCreator.NewInstallationClient(installation.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create github client")
		baseapp.WriteJSON(w, http.StatusInternalServerError, &response)
		return
	}

	pr, _, err := client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get pr")
		if isNotFound(err) {
			baseapp.WriteJSON(w, http.StatusNotFound, &response)
		} else {
			baseapp.WriteJSON(w, http.StatusInternalServerError, &response)
		}

		return
	}

	ctx, logger = h.PreparePRContext(ctx, installation.ID, pr)
	result, err := h.getApprovalResult(ctx, installation, pull.Locator{
		Owner:  owner,
		Repo:   repo,
		Number: number,
		Value:  pr,
	}, h.getOptionsFromQuery(r))

	if err != nil {
		logger.Error().Err(err).Msg("failed to get approval result for pull request")
		baseapp.WriteJSON(w, http.StatusInternalServerError, &response)
		return
	}

	response.Status = result.Status.String()
	response.Description = result.StatusDescription
	baseapp.WriteJSON(w, http.StatusOK, &response)
}

func (h *Simulate) getOptionsFromQuery(r *http.Request) *simulator.Options {
	var options simulator.Options
	if r.URL.Query().Has(paramIgnoreCommentsFrom) {
		options.IgnoreCommentsFrom = strings.Split(r.URL.Query().Get(paramIgnoreCommentsFrom), ",")
	}

	return &options
}

func (h *Simulate) getApprovalResult(ctx context.Context, installation githubapp.Installation, loc pull.Locator, options *simulator.Options) (*common.Result, error) {
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
	simulatedCtx := simulator.NewSimulatedContext(evalCtx.PullContext, options)

	evaluator, err := policy.ParsePolicy(evalCtx.Config.Config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get policy evaluator")
	}

	result := evaluator.Evaluate(ctx, simulatedCtx)
	if result.Error != nil {
		return nil, errors.Wrapf(err, "error evaluating policy in %s: %s", evalCtx.Config.Source, evalCtx.Config.Path)
	}

	return &result, nil
}
