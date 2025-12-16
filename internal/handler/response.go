package handler

import (
	"encoding/json"
	"net/http"

	"diary/internal/handler/dto"
)

// respondSuccess 返回成功响应
func respondSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.Response{
		Code:    statusCode,
		Message: message,
		Data:    data,
	})
}

// respondError 返回错误响应
func respondError(w http.ResponseWriter, statusCode int, message string, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.ErrorResponse{
		Code:    statusCode,
		Message: message,
		Error:   err,
	})
}
