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

	"github.com/google/go-github/v58/github"
	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
	"github.com/rs/zerolog"
)

const (
	LogKeyDryRun = "dry_run"
)

// DryRun performs an evaluation of of a specific pull request, similar to if the evaluation how been triggered by a
// GitHub event. However instead of writing the result as a status back to the pr, it returns a response indicating
// what the status would be if the pr was triggered from a GitHub event as normal.
type DryRun struct {
	Base
}

type DryRunResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

func (h *DryRun) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)
	var response DryRunResponse

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

	ctx, logger = h.prepareDryRunContext(ctx, installation.ID, pr)
	evalCtx, err := h.NewEvalContext(ctx, installation.ID, pull.Locator{
		Owner:  owner,
		Repo:   repo,
		Number: number,
		Value:  pr,
	})

	if err != nil {
		logger.Error().Err(err).Msg("failed to generate eval context")
		baseapp.WriteJSON(w, http.StatusInternalServerError, &response)
		return
	}

	evaluator, err := evalCtx.ParseConfig(ctx, common.TriggerAll)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get evaluator")
		baseapp.WriteJSON(w, http.StatusInternalServerError, &response)
		return
	}

	if evaluator == nil {
		logger.Error().Msg("evaluator was nil")
		baseapp.WriteJSON(w, http.StatusInternalServerError, &response)
		return
	}

	result := evaluator.Evaluate(ctx, evalCtx.PullContext)
	if result.Error != nil {
		logger.Error().Err(result.Error).Msgf("error evaluating policy in %s: %s", evalCtx.Config.Source, evalCtx.Config.Path)
		baseapp.WriteJSON(w, http.StatusInternalServerError, &response)
		return
	}

	response.Status = result.Status.String()
	response.Description = result.StatusDescription
	baseapp.WriteJSON(w, http.StatusOK, &response)
}

func (h *DryRun) prepareDryRunContext(ctx context.Context, installationID int64, pr *github.PullRequest) (context.Context, *zerolog.Logger) {
	ctx, logger := githubapp.PreparePRContext(ctx, installationID, pr.GetBase().GetRepo(), pr.GetNumber())

	logger = logger.With().Bool(LogKeyDryRun, true).Str(LogKeyGitHubSHA, pr.GetHead().GetSHA()).Logger()
	ctx = logger.WithContext(ctx)

	return ctx, &logger
}
