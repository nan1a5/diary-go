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

type TodoHandler struct {
	todoService service.TodoService
}

func NewTodoHandler(todoService service.TodoService) *TodoHandler {
	return &TodoHandler{todoService: todoService}
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	todo, err := h.todoService.Create(r.Context(), userID, req.Title, req.Description, req.DueDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "创建待办事项失败", err.Error())
		return
	}

	respondSuccess(w, http.StatusCreated, "创建成功", h.toTodoResponse(todo))
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	var req dto.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	// 先检查是否存在且属于当前用户
	existing, err := h.todoService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "待办事项不存在", err.Error())
		return
	}

	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "无权修改此待办事项", "")
		return
	}

	err = h.todoService.Update(r.Context(), uint(id), req.Title, req.Description, req.DueDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "更新待办事项失败", err.Error())
		return
	}

	// 重新获取以返回最新数据
	updated, err := h.todoService.GetByID(r.Context(), uint(id))
	if err != nil {
		// 更新成功但获取失败，返回成功响应但不带数据
		respondSuccess(w, http.StatusOK, "更新成功", nil)
		return
	}

	respondSuccess(w, http.StatusOK, "更新成功", h.toTodoResponse(updated))
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	// 先检查是否存在且属于当前用户
	existing, err := h.todoService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "待办事项不存在", err.Error())
		return
	}

	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "无权删除此待办事项", "")
		return
	}

	if err := h.todoService.Delete(r.Context(), uint(id)); err != nil {
		respondError(w, http.StatusInternalServerError, "删除待办事项失败", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, "删除成功", nil)
}

func (h *TodoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	todo, err := h.todoService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "待办事项不存在", err.Error())
		return
	}

	if todo.UserID != userID {
		respondError(w, http.StatusForbidden, "无权查看此待办事项", "")
		return
	}

	respondSuccess(w, http.StatusOK, "获取成功", h.toTodoResponse(todo))
}

func (h *TodoHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	
	// 可选的过滤参数
	var todos []domain.Todo
	var total int64
	var err error

	doneStr := r.URL.Query().Get("done")
	if doneStr != "" {
		done := doneStr == "true"
		todos, total, err = h.todoService.ListByStatus(r.Context(), userID, done, page, pageSize)
	} else {
		todos, total, err = h.todoService.ListByUserID(r.Context(), userID, page, pageSize)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取列表失败", err.Error())
		return
	}

	var todoResponses []dto.TodoResponse
	for _, t := range todos {
		todoResponses = append(todoResponses, h.toTodoResponse(&t))
	}

	respondSuccess(w, http.StatusOK, "获取成功", dto.TodoListResponse{
		Todos:    todoResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *TodoHandler) MarkAsDone(w http.ResponseWriter, r *http.Request) {
	h.updateStatus(w, r, true)
}

func (h *TodoHandler) MarkAsUndone(w http.ResponseWriter, r *http.Request) {
	h.updateStatus(w, r, false)
}

func (h *TodoHandler) updateStatus(w http.ResponseWriter, r *http.Request, done bool) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)

	// 先检查是否存在且属于当前用户
	existing, err := h.todoService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "待办事项不存在", err.Error())
		return
	}

	if existing.UserID != userID {
		respondError(w, http.StatusForbidden, "无权修改此待办事项", "")
		return
	}

	if done {
		err = h.todoService.MarkAsDone(r.Context(), uint(id))
	} else {
		err = h.todoService.MarkAsUndone(r.Context(), uint(id))
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "更新状态失败", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, "更新状态成功", nil)
}

func (h *TodoHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uint)

	total, pending, err := h.todoService.GetStats(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取统计信息失败", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, "获取成功", dto.TodoStatsResponse{
		Total:   total,
		Pending: pending,
	})
}

func (h *TodoHandler) toTodoResponse(todo *domain.Todo) dto.TodoResponse {
	return dto.TodoResponse{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Done:        todo.Done,
		DueDate:     todo.DueDate,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
}
