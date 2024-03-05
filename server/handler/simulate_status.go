package handler

import (
	"net/http"

	"github.com/palantir/go-baseapp/baseapp"
	"github.com/palantir/policy-bot/pull"
	"github.com/rs/zerolog"
)

type SimulateStatus struct {
	Simulate
}

type SimulatedStatusResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

func (h *SimulateStatus) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := *zerolog.Ctx(ctx)
	var response SimulatedStatusResponse

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
