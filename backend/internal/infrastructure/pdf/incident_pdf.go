package pdf

import (
	"fmt"
	"incidex/internal/domain"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

type IncidentPDFService struct{}

func NewIncidentPDFService() *IncidentPDFService {
	return &IncidentPDFService{}
}

// GenerateIncidentReport generates a PDF report for a single incident
func (s *IncidentPDFService) GenerateIncidentReport(incident *domain.Incident) ([]byte, error) {
	cfg := config.NewBuilder().Build()
	m := maroto.New(cfg)

	// Add header
	s.addHeader(m, incident)

	// Add basic information
	s.addBasicInfo(m, incident)

	// Add description and summary
	s.addDescriptionSection(m, incident)

	// Add tags
	if len(incident.Tags) > 0 {
		s.addTagsSection(m, incident.Tags)
	}

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return document.GetBytes(), nil
}

func (s *IncidentPDFService) addHeader(m core.Maroto, incident *domain.Incident) {
	m.AddRows(
		row.New(12).Add(
			col.New(12).Add(
				text.New("インシデントレポート", props.Text{
					Size:  16,
					Style: fontstyle.Bold,
					Align: align.Center,
				}),
			),
		),
		row.New(8).Add(
			col.New(12).Add(
				text.New(incident.Title, props.Text{
					Size:  14,
					Style: fontstyle.Bold,
					Align: align.Center,
				}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("ID: %d | 生成日時: %s", incident.ID, time.Now().Format("2006-01-02 15:04:05")), props.Text{
					Size:  9,
					Align: align.Center,
				}),
			),
		),
	)
}

func (s *IncidentPDFService) addBasicInfo(m core.Maroto, incident *domain.Incident) {
	m.AddRows(
		row.New(8).Add(
			col.New(12).Add(
				text.New("基本情報", props.Text{
					Size:  12,
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	// Severity
	severityColor := s.getSeverityColor(incident.Severity)
	m.AddRow(6,
		col.New(3).Add(
			text.New("重要度:", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
			}),
		),
		col.New(9).Add(
			text.New(string(incident.Severity), props.Text{
				Size:  10,
				Color: severityColor,
				Style: fontstyle.Bold,
			}),
		),
	)

	// Status
	statusColor := s.getStatusColor(incident.Status)
	m.AddRow(6,
		col.New(3).Add(
			text.New("ステータス:", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
			}),
		),
		col.New(9).Add(
			text.New(string(incident.Status), props.Text{
				Size:  10,
				Color: statusColor,
				Style: fontstyle.Bold,
			}),
		),
	)

	// Impact Scope
	m.AddRow(6,
		col.New(3).Add(
			text.New("影響範囲:", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
			}),
		),
		col.New(9).Add(
			text.New(incident.ImpactScope, props.Text{
				Size: 10,
			}),
		),
	)

	// Detected At
	m.AddRow(6,
		col.New(3).Add(
			text.New("検出日時:", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
			}),
		),
		col.New(9).Add(
			text.New(incident.DetectedAt.Format("2006-01-02 15:04:05"), props.Text{
				Size: 10,
			}),
		),
	)

	// Resolved At
	if incident.ResolvedAt != nil {
		m.AddRow(6,
			col.New(3).Add(
				text.New("解決日時:", props.Text{
					Size:  10,
					Style: fontstyle.Bold,
				}),
			),
			col.New(9).Add(
				text.New(incident.ResolvedAt.Format("2006-01-02 15:04:05"), props.Text{
					Size: 10,
				}),
			),
		)
	}

	// Creator
	if incident.Creator != nil {
		m.AddRow(6,
			col.New(3).Add(
				text.New("作成者:", props.Text{
					Size:  10,
					Style: fontstyle.Bold,
				}),
			),
			col.New(9).Add(
				text.New(incident.Creator.Name, props.Text{
					Size: 10,
				}),
			),
		)
	}

	// Assignee
	if incident.Assignee != nil {
		m.AddRow(6,
			col.New(3).Add(
				text.New("担当者:", props.Text{
					Size:  10,
					Style: fontstyle.Bold,
				}),
			),
			col.New(9).Add(
				text.New(incident.Assignee.Name, props.Text{
					Size: 10,
				}),
			),
		)
	}
}

func (s *IncidentPDFService) addDescriptionSection(m core.Maroto, incident *domain.Incident) {
	m.AddRows(
		row.New(4),
		row.New(8).Add(
			col.New(12).Add(
				text.New("詳細説明", props.Text{
					Size:  12,
					Style: fontstyle.Bold,
				}),
			),
		),
		row.New(0).Add(
			col.New(12).Add(
				text.New(incident.Description, props.Text{
					Size: 10,
				}),
			),
		),
	)

	// Add AI summary if exists
	if incident.Summary != "" {
		m.AddRows(
			row.New(4),
			row.New(8).Add(
				col.New(12).Add(
					text.New("AI要約", props.Text{
						Size:  12,
						Style: fontstyle.Bold,
					}),
				),
			),
			row.New(0).Add(
				col.New(12).Add(
					text.New(incident.Summary, props.Text{
						Size:  10,
						Color: &props.Color{Red: 50, Green: 50, Blue: 150},
					}),
				),
			),
		)
	}
}

func (s *IncidentPDFService) addTagsSection(m core.Maroto, tags []domain.Tag) {
	m.AddRows(
		row.New(4),
		row.New(8).Add(
			col.New(12).Add(
				text.New("タグ", props.Text{
					Size:  12,
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	tagNames := ""
	for i, tag := range tags {
		if i > 0 {
			tagNames += ", "
		}
		tagNames += tag.Name
	}

	m.AddRow(6,
		col.New(12).Add(
			text.New(tagNames, props.Text{
				Size: 10,
			}),
		),
	)
}

func (s *IncidentPDFService) getSeverityColor(severity domain.Severity) *props.Color {
	switch severity {
	case domain.SeverityCritical:
		return &props.Color{Red: 220, Green: 38, Blue: 38}
	case domain.SeverityHigh:
		return &props.Color{Red: 239, Green: 68, Blue: 68}
	case domain.SeverityMedium:
		return &props.Color{Red: 251, Green: 146, Blue: 60}
	case domain.SeverityLow:
		return &props.Color{Red: 34, Green: 197, Blue: 94}
	default:
		return &props.Color{Red: 0, Green: 0, Blue: 0}
	}
}

func (s *IncidentPDFService) getStatusColor(status domain.Status) *props.Color {
	switch status {
	case domain.StatusOpen:
		return &props.Color{Red: 239, Green: 68, Blue: 68}
	case domain.StatusInvestigating:
		return &props.Color{Red: 251, Green: 191, Blue: 36}
	case domain.StatusResolved:
		return &props.Color{Red: 34, Green: 197, Blue: 94}
	case domain.StatusClosed:
		return &props.Color{Red: 107, Green: 114, Blue: 128}
	default:
		return &props.Color{Red: 0, Green: 0, Blue: 0}
	}
}
