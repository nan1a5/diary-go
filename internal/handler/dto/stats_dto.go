package dto

type MonthlyTrendItem struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

type TopTagItem struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

type DashboardStatsResponse struct {
	DiaryCount        int64              `json:"diary_count"`
	TodoCompletedRate float64            `json:"todo_completed_rate"`
	MonthlyTrend      []MonthlyTrendItem `json:"monthly_trend"`
	TopTags           []TopTagItem       `json:"top_tags"`
}
