package handler

import (
	"errors"
	"net/http"
	"time"

	"saythis-backend/internal/helper"
	"saythis-backend/internal/src/auth"
	statsdomain "saythis-backend/internal/src/stats/domain"
	"saythis-backend/internal/src/stats/usecase"
)

type GetDailyHandler struct {
	usecase *usecase.StatsUseCase
}

func NewGetDailyHandler(uc *usecase.StatsUseCase) *GetDailyHandler {
	return &GetDailyHandler{usecase: uc}
}

type getDailyResponse struct {
	DailyStat *dailySnapshotResponse `json:"daily_stat"`
}

func (h *GetDailyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		helper.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	date, err := time.Parse(dateLayout, r.PathValue("date"))
	if err != nil {
		status, msg := mapStatsError(statsdomain.ErrInvalidDate)
		helper.Error(w, status, msg)
		return
	}

	stat, err := h.usecase.GetDailyStat(r.Context(), claims.UserID, date)
	if err != nil {
		if errors.Is(err, statsdomain.ErrDailyStatNotFound) {
			helper.JSON(w, http.StatusNotFound, getDailyResponse{DailyStat: nil})
			return
		}
		status, msg := mapStatsError(err)
		helper.Error(w, status, msg)
		return
	}

	response := toDailySnapshotResponse(stat)
	helper.JSON(w, http.StatusOK, getDailyResponse{DailyStat: &response})
}
