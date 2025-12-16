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

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	user, err := h.userService.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			respondError(w, http.StatusConflict, "用户已存在", err.Error())
		case service.ErrInvalidUsername:
			respondError(w, http.StatusBadRequest, "用户名格式不正确", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "注册失败", err.Error())
		}
		return
	}

	// 注册成功后自动登录
	_, token, err := h.userService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "注册成功但登录失败", err.Error())
		return
	}

	response := dto.LoginResponse{
		User:  *h.toUserResponse(user),
		Token: token,
	}

	respondSuccess(w, http.StatusCreated, "注册成功", response)
}

// Login 用户登录
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	user, token, err := h.userService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserNotFound, service.ErrInvalidPassword:
			respondError(w, http.StatusUnauthorized, "用户名或密码错误", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "登录失败", err.Error())
		}
		return
	}

	response := dto.LoginResponse{
		User:  *h.toUserResponse(user),
		Token: token,
	}

	respondSuccess(w, http.StatusOK, "登录成功", response)
}

// GetProfile 获取当前用户信息
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// 从上下文中获取用户ID（由中间件设置）
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "未授权", "")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "用户不存在", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, "获取成功", h.toUserResponse(user))
}

// GetUserByID 根据ID获取用户
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的用户ID", err.Error())
		return
	}

	user, err := h.userService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "用户不存在", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, "获取成功", h.toUserResponse(user))
}

// UpdateUsername 更新用户名
func (h *UserHandler) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "未授权", "")
		return
	}

	var req dto.UpdateUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	err := h.userService.UpdateUsername(r.Context(), userID, req.NewUsername)
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			respondError(w, http.StatusConflict, "用户名已被占用", err.Error())
		case service.ErrInvalidUsername:
			respondError(w, http.StatusBadRequest, "用户名格式不正确", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "更新失败", err.Error())
		}
		return
	}

	respondSuccess(w, http.StatusOK, "更新成功", nil)
}

// UpdatePassword 更新密码
func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "未授权", "")
		return
	}

	var req dto.UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	err := h.userService.UpdatePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		switch err {
		case service.ErrInvalidPassword:
			respondError(w, http.StatusUnauthorized, "旧密码错误", err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "更新失败", err.Error())
		}
		return
	}

	respondSuccess(w, http.StatusOK, "密码更新成功", nil)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "未授权", "")
		return
	}

	err := h.userService.Delete(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "删除失败", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, "删除成功", nil)
}

// ListUsers 获取用户列表（管理员功能）
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	users, total, err := h.userService.List(r.Context(), page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取用户列表失败", err.Error())
		return
	}

	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *h.toUserResponse(&user)
	}

	response := dto.UserListResponse{
		Users:    userResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	respondSuccess(w, http.StatusOK, "获取成功", response)
}

// 辅助方法：将领域模型转换为响应DTO
func (h *UserHandler) toUserResponse(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// 辅助方法：从上下文中获取用户ID
func (h *UserHandler) getUserIDFromContext(r *http.Request) uint {
	// 假设中间件会将用户ID存入context
	// 键名为 "user_id"
	if userID, ok := r.Context().Value("user_id").(uint); ok {
		return userID
	}
	return 0
}



