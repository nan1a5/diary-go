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

type TagHandler struct {
	tagService service.TagService
}

func NewTagHandler(tagService service.TagService) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	tag, err := h.tagService.Create(r.Context(), req.Name)
	if err != nil {
		if err == service.ErrTagAlreadyExists {
			h.respondError(w, http.StatusConflict, "标签已存在", err.Error())
		} else {
			h.respondError(w, http.StatusInternalServerError, "创建标签失败", err.Error())
		}
		return
	}

	h.respondSuccess(w, http.StatusCreated, "创建成功", h.toTagResponse(tag))
}

func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var req dto.UpdateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	if err := h.tagService.Update(r.Context(), uint(id), req.Name); err != nil {
		switch err {
		case service.ErrTagNotFound:
			h.respondError(w, http.StatusNotFound, "标签不存在", err.Error())
		case service.ErrTagAlreadyExists:
			h.respondError(w, http.StatusConflict, "标签名称已存在", err.Error())
		default:
			h.respondError(w, http.StatusInternalServerError, "更新失败", err.Error())
		}
		return
	}

	h.respondSuccess(w, http.StatusOK, "更新成功", nil)
}

func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := h.tagService.Delete(r.Context(), uint(id)); err != nil {
		h.respondError(w, http.StatusInternalServerError, "删除失败", err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, "删除成功", nil)
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	tags, total, err := h.tagService.List(r.Context(), page, pageSize)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "获取列表失败", err.Error())
		return
	}

	tagResponses := make([]dto.TagResponse, len(tags))
	for i, tag := range tags {
		tagResponses[i] = *h.toTagResponse(&tag)
	}

	h.respondSuccess(w, http.StatusOK, "获取成功", dto.TagListResponse{
		Tags:     tagResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *TagHandler) GetPopular(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	
	tags, err := h.tagService.GetPopularTags(r.Context(), limit)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "获取热门标签失败", err.Error())
		return
	}

	tagResponses := make([]dto.TagResponse, len(tags))
	for i, tag := range tags {
		tagResponses[i] = *h.toTagResponse(&tag)
	}

	h.respondSuccess(w, http.StatusOK, "获取成功", tagResponses)
}

func (h *TagHandler) toTagResponse(tag *domain.Tag) *dto.TagResponse {
	return &dto.TagResponse{
		ID:        tag.ID,
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt,
	}
}

func (h *TagHandler) respondSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.Response{
		Code:    statusCode,
		Message: message,
		Data:    data,
	})
}

func (h *TagHandler) respondError(w http.ResponseWriter, statusCode int, message string, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.ErrorResponse{
		Code:    statusCode,
		Message: message,
		Error:   err,
	})
}
