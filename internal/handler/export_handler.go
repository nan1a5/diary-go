package handler

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"diary/internal/domain"
	"diary/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type ExportHandler struct {
	diaryService service.DiaryService
}

func NewExportHandler(diaryService service.DiaryService) *ExportHandler {
	return &ExportHandler{
		diaryService: diaryService,
	}
}

// ExportSingle 导出单个日记
func (h *ExportHandler) ExportSingle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "无效的ID", err.Error())
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "txt"
	}

	diary, err := h.diaryService.GetByID(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "日记未找到", err.Error())
		return
	}

	filename := fmt.Sprintf("diary_%s_%d.%s", diary.Date.Format("20060102"), diary.ID, format)
	content := h.formatDiary(diary, format)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(content)
}

// ExportBatchRequest 批量导出请求
type ExportBatchRequest struct {
	Type      string `json:"type"`       // "all", "selected", "date_range"
	IDs       []uint `json:"ids"`        // 当 type="selected" 时使用
	StartDate string `json:"start_date"` // 当 type="date_range" 或 "all" 时使用
	EndDate   string `json:"end_date"`   // 当 type="date_range" 或 "all" 时使用
	Format    string `json:"format"`     // "md", "txt", "csv"
}

// ExportBatch 批量导出日记
func (h *ExportHandler) ExportBatch(w http.ResponseWriter, r *http.Request) {
	var req ExportBatchRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		respondError(w, http.StatusBadRequest, "无效的请求参数", err.Error())
		return
	}

	userID := r.Context().Value("user_id").(uint)
	var diaries []domain.Diary
	var err error

	switch req.Type {
	case "selected":
		if len(req.IDs) == 0 {
			respondError(w, http.StatusBadRequest, "未选择任何日记", "")
			return
		}
		diaries, err = h.diaryService.GetByIDs(r.Context(), userID, req.IDs)
	case "date_range", "all":
		var start, end time.Time
		if req.StartDate != "" {
			start, _ = time.Parse("2006-01-02", req.StartDate)
		} else {
			start = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
		}
		if req.EndDate != "" {
			end, _ = time.Parse("2006-01-02", req.EndDate)
			end = end.Add(24 * time.Hour) // 包含结束当天
		} else {
			end = time.Now().Add(24 * time.Hour)
		}
		diaries, err = h.diaryService.GetByDateRange(r.Context(), userID, start, end)
	default:
		respondError(w, http.StatusBadRequest, "未知的导出类型", "")
		return
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "获取日记失败", err.Error())
		return
	}

	if len(diaries) == 0 {
		respondError(w, http.StatusBadRequest, "没有可导出的日记", "")
		return
	}

	// 如果是 CSV 格式，返回单个 CSV 文件
	if req.Format == "csv" {
		filename := fmt.Sprintf("diaries_export_%s.csv", time.Now().Format("20060102150405"))
		content := h.generateCSV(diaries)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Write(content)
		return
	}

	// 否则返回 ZIP 包
	filename := fmt.Sprintf("diaries_export_%s.zip", time.Now().Format("20060102150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/zip")

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, diary := range diaries {
		fileExt := req.Format
		if fileExt == "" {
			fileExt = "txt"
		}
		entryName := fmt.Sprintf("%s_%s.%s", diary.Date.Format("20060102"), strings.ReplaceAll(diary.Title, " ", "_"), fileExt)

		f, err := zipWriter.Create(entryName)
		if err != nil {
			continue
		}
		f.Write(h.formatDiary(&diary, req.Format))
	}
}

func (h *ExportHandler) formatDiary(diary *domain.Diary, format string) []byte {
	var sb strings.Builder

	if format == "md" {
		sb.WriteString(fmt.Sprintf("# %s\n\n", diary.Title))
		sb.WriteString(fmt.Sprintf("> 日期：%s  |  天气：%s  |  心情：%s\n\n",
			diary.Date.Format("2006-01-02 15:04"), diary.Weather, diary.Mood))
		if diary.Location != "" {
			sb.WriteString(fmt.Sprintf("> 地点：%s\n\n", diary.Location))
		}
		if diary.Music != "" {
			sb.WriteString(fmt.Sprintf("> 音乐：%s\n\n", diary.Music))
		}
		sb.WriteString("---\n\n")
		sb.WriteString(diary.PlainContent)
	} else {
		// TXT format
		sb.WriteString(fmt.Sprintf("标题：%s\n", diary.Title))
		sb.WriteString(fmt.Sprintf("日期：%s\n", diary.Date.Format("2006-01-02 15:04")))
		sb.WriteString(fmt.Sprintf("天气：%s\n", diary.Weather))
		sb.WriteString(fmt.Sprintf("心情：%s\n", diary.Mood))
		if diary.Location != "" {
			sb.WriteString(fmt.Sprintf("地点：%s\n", diary.Location))
		}
		sb.WriteString("\n----------------------------------------\n\n")
		sb.WriteString(diary.PlainContent)
	}

	return []byte(sb.String())
}

func (h *ExportHandler) generateCSV(diaries []domain.Diary) []byte {
	buf := &bytes.Buffer{}
	// 写入 UTF-8 BOM，防止 Excel 打开乱码
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(buf)

	// Header
	w.Write([]string{"ID", "日期", "标题", "内容", "天气", "心情", "地点", "音乐"})

	for _, d := range diaries {
		w.Write([]string{
			fmt.Sprintf("%d", d.ID),
			d.Date.Format("2006-01-02 15:04:05"),
			d.Title,
			d.PlainContent,
			d.Weather,
			d.Mood,
			d.Location,
			d.Music,
		})
	}
	w.Flush()
	return buf.Bytes()
}
