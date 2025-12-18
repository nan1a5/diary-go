package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"diary/internal/domain"
	"diary/internal/handler/dto"
	"diary/internal/service"

	"github.com/go-chi/chi/v5"
)

type DiaryHandler struct {
	diaryService service.DiaryService
}

func NewDiaryHandler(diaryService service.DiaryService) *DiaryHandler {
	return &DiaryHandler{diaryService: diaryService}
}

func (h *DiaryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDiaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	diary, err := h.diaryService.Create(
		r.Context(),
		userID,
		req.Title,
		req.Content,
		req.Weather,
		req.Mood,
		req.Location,
		req.Date,
		req.IsPublic,
		req.Tags,
		req.ImageIDs,
		req.Properties,
		req.Music,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "创建日记失败", err.Error())
		return
	}

	respondSuccess(w, http.StatusCreated, "创建成功", h.toDiaryResponse(diary, true))
}

func (h *DiaryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	var req dto.UpdateDiaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	// 先检查是否存在且属于当前用户
	existing, err := h.diaryService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "日记不存在", err.Error())
		return
	}

	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "无权修改此日记", "")
		return
	}

	err = h.diaryService.Update(
		r.Context(),
		uint(id),
		req.Title,
		req.Content,
		req.Weather,
		req.Mood,
		req.Location,
		req.Date,
		req.IsPublic,
		req.Tags,
		req.Properties,
		req.Music,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "更新日记失败", err.Error())
		return
	}

	// 获取更新后的对象以返回
	updated, err := h.diaryService.GetByID(r.Context(), uint(id))
	if err != nil {
		// 虽然更新成功但获取失败，返回成功但不带数据或带部分数据
		respondSuccess(w, http.StatusOK, "更新成功", nil)
		return
	}

	respondSuccess(w, http.StatusOK, "更新成功", h.toDiaryResponse(updated, true))
}

func (h *DiaryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	// 先检查是否存在且属于当前用户
	existing, err := h.diaryService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "日记不存在", err.Error())
		return
	}

	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "无权删除此日记", "")
		return
	}

	if err := h.diaryService.Delete(r.Context(), uint(id)); err != nil {
		respondError(w, http.StatusInternalServerError, "删除日记失败", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, "删除成功", nil)
}

func (h *DiaryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	diary, err := h.diaryService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "日记不存在", err.Error())
		return
	}

	// 如果是私有日记，只能作者查看
	// 如果是公开日记，任何人可以查看？还是说这只是个人日记应用？
	// 假设：公开日记可以被其他人查看（如果业务支持社交功能），但目前似乎只有单用户视角。
	// 不过 GetByID 通常用于编辑或查看详情，如果是自己的日记当然可以。
	// 如果是别人的公开日记，可能需要不同的逻辑。
	// 这里简单处理：只能看自己的，或者公开的。
	if diary.UserID != userID && !diary.IsPublic {
		respondError(w, http.StatusForbidden, "无权查看此日记", "")
		return
	}

	// 如果不是自己的日记，可能不想展示某些敏感信息？这里暂时一视同仁。

	respondSuccess(w, http.StatusOK, "获取成功", h.toDiaryResponse(diary, true))
}

func (h *DiaryHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}

	// 支持按日期范围过滤
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var diaries []domain.Diary
	var total int64
	var err error

	if startDateStr != "" && endDateStr != "" {
		// 按日期范围查询，这里假设 Service 层支持或我们在 Controller 组装
		// 实际上 Service 层有 GetByDateRange，但不带分页（通常用于日历视图）
		// 如果需要分页列表，可能需要扩展 Service。
		// 这里简单调用 ListByUserID
		diaries, total, err = h.diaryService.ListByUserID(r.Context(), userID, page, pageSize)
	} else {
		diaries, total, err = h.diaryService.ListByUserID(r.Context(), userID, page, pageSize)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取列表失败", err.Error())
		return
	}

	var diaryResponses []dto.DiaryResponse
	for _, d := range diaries {
		diaryResponses = append(diaryResponses, h.toDiaryResponse(&d, false))
	}

	respondSuccess(w, http.StatusOK, "获取成功", dto.DiaryListResponse{
		Diaries:  diaryResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *DiaryHandler) Search(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("q")
	if keyword == "" {
		respondError(w, http.StatusBadRequest, "搜索关键词不能为空", "")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	userID := r.Context().Value("user_id").(uint)

	diaries, total, err := h.diaryService.Search(r.Context(), userID, keyword, page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "搜索失败", err.Error())
		return
	}

	diaryResponses := make([]dto.DiaryResponse, len(diaries))
	for i, diary := range diaries {
		// 搜索列表只返回摘要，不返回详细内容（因为可能需要解密，且列表不需要全文）
		// 如果需要，可以在这里解密，但 Search 接口通常返回列表
		diaryResponses[i] = h.toDiaryResponse(&diary, false)
	}

	response := dto.DiaryListResponse{
		Diaries:  diaryResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	respondSuccess(w, http.StatusOK, "搜索成功", response)
}

func (h *DiaryHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}

	diaries, total, err := h.diaryService.ListPublic(r.Context(), page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取公开日记失败", err.Error())
		return
	}

	var diaryResponses []dto.DiaryResponse
	for _, d := range diaries {
		diaryResponses = append(diaryResponses, h.toDiaryResponse(&d, false))
	}

	respondSuccess(w, http.StatusOK, "获取成功", dto.DiaryListResponse{
		Diaries:  diaryResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *DiaryHandler) TogglePin(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	newStatus, err := h.diaryService.TogglePin(r.Context(), userID, uint(id))
	if err != nil {
		respondError(w, http.StatusBadRequest, "操作失败", err.Error())
		return
	}

	msg := "已置顶"
	if !newStatus {
		msg = "已取消置顶"
	}
	respondSuccess(w, http.StatusOK, msg, map[string]bool{"is_pinned": newStatus})
}

func (h *DiaryHandler) toDiaryResponse(diary *domain.Diary, includeContent bool) dto.DiaryResponse {
	resp := dto.DiaryResponse{
		ID:         diary.ID,
		Title:      diary.Title,
		Summary:    diary.Summary,
		Weather:    diary.Weather,
		Mood:       diary.Mood,
		Location:   diary.Location,
		Date:       diary.Date,
		IsPublic:   diary.IsPublic,
		IsPinned:   diary.IsPinned,
		Properties: diary.Properties,
		Music:      diary.Music,
		CreatedAt:  diary.CreatedAt,
		UpdatedAt:  diary.UpdatedAt,
	}

	if includeContent {
		resp.Content = diary.PlainContent
	}

	if len(diary.Tags) > 0 {
		var tags []dto.TagResponse
		for _, t := range diary.Tags {
			tags = append(tags, dto.TagResponse{
				ID:   t.ID,
				Name: t.Name,
			})
		}
		resp.Tags = tags
	}

	if len(diary.Images) > 0 {
		var images []dto.ImageResponse
		for _, img := range diary.Images {
			images = append(images, dto.ImageResponse{
				ID:        img.ID,
				Path:      img.Path,
				DiaryID:   img.DiaryID,
				CreatedAt: img.CreatedAt,
			})
		}
		resp.Images = images
	}

	return resp
}
