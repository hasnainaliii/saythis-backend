package handler

import (
	"net/http"
	"time"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth"
	statsdomain "saythis-backend/internal/src/stats/domain"
	"saythis-backend/internal/src/stats/usecase"
)

type GetStatsHandler struct {
	usecase *usecase.StatsUseCase
}

func NewGetStatsHandler(uc *usecase.StatsUseCase) *GetStatsHandler {
	return &GetStatsHandler{usecase: uc}
}

func (h *GetStatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	from, err := parseOptionalDateQuery(r, "from")
	if err != nil {
		status, msg := mapStatsError(statsdomain.ErrInvalidDate)
		helper.Error(w, status, msg)
		return
	}
	to, err := parseOptionalDateQuery(r, "to")
	if err != nil {
		status, msg := mapStatsError(statsdomain.ErrInvalidDate)
		helper.Error(w, status, msg)
		return
	}

	stats, err := h.usecase.GetStats(r.Context(), claims.UserID, from, to)
	if err != nil {
		status, msg := mapStatsError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusOK, toStatsResponse(stats))
}

func parseOptionalDateQuery(r *http.Request, key string) (*time.Time, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
