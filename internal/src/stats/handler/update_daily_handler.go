package handler

import (
	"net/http"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth"
	"saythis-backend/internal/src/stats/usecase"
)

type UpdateDailyHandler struct {
	usecase *usecase.StatsUseCase
}

func NewUpdateDailyHandler(uc *usecase.StatsUseCase) *UpdateDailyHandler {
	return &UpdateDailyHandler{usecase: uc}
}

type updateDailyResponse struct {
	DailyStat dailyStatResponse `json:"daily_stat"`
}

func (h *UpdateDailyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	patch, err := decodeDailyStatPatch(r)
	if err != nil {
		status, msg := mapStatsError(err)
		helper.Error(w, status, msg)
		return
	}

	stat, err := h.usecase.UpdateDailyStat(r.Context(), claims.UserID, patch)
	if err != nil {
		status, msg := mapStatsError(err)
		helper.Error(w, status, msg)
		return
	}

	helper.JSON(w, http.StatusOK, updateDailyResponse{DailyStat: toDailyStatResponse(stat)})
}
