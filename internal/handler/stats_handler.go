package handler

import (
	"net/http"

	"diary/internal/domain"
	"diary/internal/handler/dto"
)

type StatsHandler struct {
	diaryRepo domain.DiaryRepository
	todoRepo  domain.TodoRepository
}

func NewStatsHandler(diaryRepo domain.DiaryRepository, todoRepo domain.TodoRepository) *StatsHandler {
	return &StatsHandler{
		diaryRepo: diaryRepo,
		todoRepo:  todoRepo,
	}
}

func (h *StatsHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)

	// 1. Diary Count
	diaryCount, err := h.diaryRepo.CountByUserID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取日记统计失败", err.Error())
		return
	}

	// 2. Todo Stats
	todoCount, err := h.todoRepo.CountByUserID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取待办统计失败", err.Error())
		return
	}

	pendingCount, err := h.todoRepo.CountPending(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取待办详情统计失败", err.Error())
		return
	}

	var todoCompletedRate float64
	if todoCount > 0 {
		todoCompletedRate = float64(todoCount-pendingCount) / float64(todoCount)
	}

	// 3. Monthly Trend
	monthlyTrend, err := h.diaryRepo.GetMonthlyTrend(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取月度趋势失败", err.Error())
		return
	}

	var dtoMonthlyTrend []dto.MonthlyTrendItem
	if monthlyTrend != nil {
		dtoMonthlyTrend = make([]dto.MonthlyTrendItem, len(monthlyTrend))
		for i, item := range monthlyTrend {
			dtoMonthlyTrend[i] = dto.MonthlyTrendItem{
				Month: item.Month,
				Count: item.Count,
			}
		}
	} else {
		dtoMonthlyTrend = []dto.MonthlyTrendItem{}
	}

	// 4. Top Tags
	topTags, err := h.diaryRepo.GetTopTags(r.Context(), userID, 6)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取热门标签失败", err.Error())
		return
	}

	var dtoTopTags []dto.TopTagItem
	if topTags != nil {
		dtoTopTags = make([]dto.TopTagItem, len(topTags))
		for i, item := range topTags {
			dtoTopTags[i] = dto.TopTagItem{
				Tag:   item.Tag,
				Count: item.Count,
			}
		}
	} else {
		dtoTopTags = []dto.TopTagItem{}
	}

	response := dto.DashboardStatsResponse{
		DiaryCount:        diaryCount,
		TodoCompletedRate: todoCompletedRate,
		MonthlyTrend:      dtoMonthlyTrend,
		TopTags:           dtoTopTags,
	}

	respondSuccess(w, http.StatusOK, "获取统计成功", response)
}
