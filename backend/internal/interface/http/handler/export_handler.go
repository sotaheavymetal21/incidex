package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"incidex/internal/domain"
	"incidex/internal/infrastructure/pdf"
	"incidex/internal/usecase"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ExportHandler はエクスポート関連の HTTP handler を提供します
type ExportHandler struct {
	incidentUsecase usecase.IncidentUsecase
	pdfService      *pdf.IncidentPDFService
}

// NewExportHandler は新しい ExportHandler を作成します
func NewExportHandler(incidentUsecase usecase.IncidentUsecase) *ExportHandler {
	return &ExportHandler{
		incidentUsecase: incidentUsecase,
		pdfService:      pdf.NewIncidentPDFService(),
	}
}

// ExportIncidentsCSV godoc
// @Summary インシデントを CSV にエクスポートします
// @Description すべてのインシデント（フィルタ付き）を CSV 形式でエクスポートします
// @Tags export
// @Accept json
// @Produce text/csv
// @Param severity query string false "重要度でフィルタ"
// @Param status query string false "ステータスでフィルタ"
// @Param tag_ids query string false "タグ ID でフィルタ（カンマ区切り）"
// @Param search query string false "タイトル/説明で検索"
// @Success 200 {file} file "CSV ファイル"
// @Failure 500 {object} map[string]string
// @Router /api/export/incidents [get]
// @Security BearerAuth
func (h *ExportHandler) ExportIncidentsCSV(c *gin.Context) {
	// クエリパラメータからフィルタをパース
	filters := domain.IncidentFilters{
		Severity: c.Query("severity"),
		Status:   c.Query("status"),
		Search:   c.Query("search"),
	}

	// タグ ID をパース（指定されている場合）
	if tagIDsStr := c.Query("tag_ids"); tagIDsStr != "" {
		tagIDsParts := strings.Split(tagIDsStr, ",")
		for _, part := range tagIDsParts {
			if tagID, err := strconv.ParseUint(part, 10, 32); err == nil {
				filters.TagIDs = append(filters.TagIDs, uint(tagID))
			}
		}
	}

	// ページネーションなしですべてのインシデントを取得（大きな上限値を設定）
	pagination := domain.Pagination{
		Page:  1,
		Limit: 10000, // すべてのインシデントを取得するための大きな数値
	}

	incidents, _, err := h.incidentUsecase.GetAllIncidents(c.Request.Context(), filters, pagination)
	if err != nil {
		HandleError(c, err)
		return
	}

	// CSV バッファを作成
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Excel 互換性のために UTF-8 BOM を書き込み
	buf.WriteString("\xEF\xBB\xBF")

	// CSV header を書き込み
	header := []string{
		"ID",
		"タイトル",
		"説明",
		"重要度",
		"ステータス",
		"影響範囲",
		"検出日時",
		"解決日時",
		"担当者",
		"作成者",
		"タグ",
		"作成日時",
	}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV header"})
		return
	}

	// インシデントデータを書き込み
	for _, incident := range incidents {
		var assigneeName string
		if incident.Assignee != nil {
			assigneeName = incident.Assignee.Name
		}

		var resolvedAt string
		if incident.ResolvedAt != nil {
			resolvedAt = incident.ResolvedAt.Format("2006-01-02 15:04:05")
		}

		// タグ名を収集
		var tagNames []string
		for _, tag := range incident.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		tagsStr := strings.Join(tagNames, ", ")

		row := []string{
			fmt.Sprintf("%d", incident.ID),
			incident.Title,
			incident.Description,
			string(incident.Severity),
			string(incident.Status),
			incident.ImpactScope,
			incident.DetectedAt.Format("2006-01-02 15:04:05"),
			resolvedAt,
			assigneeName,
			incident.Creator.Name,
			tagsStr,
			incident.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV row"})
			return
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize CSV"})
		return
	}

	// ファイルダウンロード用の header を設定
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=incidents.csv")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Length", strconv.Itoa(buf.Len()))

	c.Data(http.StatusOK, "text/csv", buf.Bytes())
}

// ExportIncidentPDF godoc
// @Summary 単一のインシデントを PDF にエクスポートします
// @Description 単一のインシデントの PDF レポートを生成します
// @Tags export
// @Produce application/pdf
// @Param id path int true "インシデント ID"
// @Success 200 {file} file "PDF ファイル"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/export/incidents/{id}/pdf [get]
// @Security BearerAuth
func (h *ExportHandler) ExportIncidentPDF(c *gin.Context) {
	// インシデント ID をパース
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	// インシデントを取得
	incident, err := h.incidentUsecase.GetIncidentByID(c.Request.Context(), uint(id))
	if err != nil {
		HandleError(c, err)
		return
	}

	// PDF を生成
	pdfBytes, err := h.pdfService.GenerateIncidentReport(incident)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate PDF: %v", err)})
		return
	}

	// ファイルダウンロード用の header を設定
	filename := fmt.Sprintf("incident_%d_%s.pdf", incident.ID, time.Now().Format("20060102"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
