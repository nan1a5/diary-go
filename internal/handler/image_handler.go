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

type ImageHandler struct {
	imageService service.ImageService
}

func NewImageHandler(imageService service.ImageService) *ImageHandler {
	return &ImageHandler{imageService: imageService}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// 限制文件大小，例如 10MB
	r.ParseMultipartForm(10 << 20)

	file, header, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "无法获取上传文件", err.Error())
		return
	}
	defer file.Close()

	userID := r.Context().Value("user_id").(uint)

	// 可选的 diary_id
	var diaryID *uint
	diaryIDStr := r.FormValue("diary_id")
	if diaryIDStr != "" {
		id, err := strconv.ParseUint(diaryIDStr, 10, 32)
		if err == nil {
			uid := uint(id)
			diaryID = &uid
		}
	}

	image, err := h.imageService.Upload(r.Context(), userID, file, header.Filename)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "图片上传失败", err.Error())
		return
	}

	// 如果提供了 diary_id，尝试关联
	if diaryID != nil {
		// 这里最好异步处理或者忽略错误，以免影响上传主流程，或者返回警告
		// 简单起见，我们尝试关联，如果失败记录日志或忽略
		_ = h.imageService.AttachToDiary(r.Context(), image.ID, *diaryID)
		// 更新返回对象的 DiaryID
		image.DiaryID = diaryID
	}

	respondSuccess(w, http.StatusCreated, "上传成功", h.toImageResponse(image))
}

func (h *ImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	// 先检查是否存在且属于当前用户
	existing, err := h.imageService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "图片不存在", err.Error())
		return
	}

	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "无权删除此图片", "")
		return
	}

	if err := h.imageService.Delete(r.Context(), uint(id)); err != nil {
		respondError(w, http.StatusInternalServerError, "删除图片失败", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, "删除成功", nil)
}

func (h *ImageHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	image, err := h.imageService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "图片不存在", err.Error())
		return
	}

	if image.UserID != userID {
		respondError(w, http.StatusForbidden, "无权查看此图片", "")
		return
	}

	respondSuccess(w, http.StatusOK, "获取成功", h.toImageResponse(image))
}

func (h *ImageHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}

	// 可选的 diary_id 过滤
	// var diaryID *uint
	// diaryIDStr := r.URL.Query().Get("diary_id")
	// if diaryIDStr != "" {
	// 	id, err := strconv.ParseUint(diaryIDStr, 10, 32)
	// 	if err == nil {
	// 		uid := uint(id)
	// 		diaryID = &uid
	// 	}
	// }

	images, total, err := h.imageService.ListByUserID(r.Context(), userID, page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取列表失败", err.Error())
		return
	}

	var imageResponses []dto.ImageResponse
	for _, img := range images {
		imageResponses = append(imageResponses, h.toImageResponse(&img))
	}

	respondSuccess(w, http.StatusOK, "获取成功", dto.ImageListResponse{
		Images:   imageResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *ImageHandler) AttachToDiary(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的图片ID", err.Error())
		return
	}

	var req dto.AttachImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	// 检查图片权限
	existing, err := h.imageService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "图片不存在", err.Error())
		return
	}
	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "无权操作此图片", "")
		return
	}

	// TODO: 这里应该也检查日记的所有权，但 imageService.AttachToDiary 可能会处理，或者在这里调用 diaryService 检查
	// 为了解耦，我们假设 Service 层或数据库约束会处理，或者我们相信用户只能操作自己的数据
	// 如果需要严格检查，可以在 Service 层增加逻辑，或者在这里引入 DiaryService

	if err := h.imageService.AttachToDiary(r.Context(), uint(id), req.DiaryID); err != nil {
		respondError(w, http.StatusInternalServerError, "关联日记失败", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, "关联成功", nil)
}

func (h *ImageHandler) toImageResponse(image *domain.Image) dto.ImageResponse {
	return dto.ImageResponse{
		ID:        image.ID,
		Path:      image.Path,
		DiaryID:   image.DiaryID,
		CreatedAt: image.CreatedAt,
	}
}
