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

// GenerateSummaryReport generates a PDF summary report for a date range
func (s *IncidentPDFService) GenerateSummaryReport(
	incidents []*domain.Incident,
	startDate, endDate time.Time,
	stats *SummaryStats,
) ([]byte, error) {
	cfg := config.NewBuilder().Build()
	m := maroto.New(cfg)

	// Add header
	s.addSummaryHeader(m, startDate, endDate)

	// Add statistics overview
	s.addStatisticsOverview(m, stats)

	// Add incidents by severity
	s.addIncidentsBySeverity(m, incidents)

	// Add incidents table
	s.addIncidentsTable(m, incidents)

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary PDF: %w", err)
	}

	return document.GetBytes(), nil
}

type SummaryStats struct {
	TotalIncidents   int
	BySeverity       map[string]int
	ByStatus         map[string]int
	ResolvedCount    int
	AverageMTTR      float64
	SLAViolatedCount int
}

func (s *IncidentPDFService) addSummaryHeader(m core.Maroto, startDate, endDate time.Time) {
	m.AddRows(
		row.New(12).Add(
			col.New(12).Add(
				text.New("インシデントサマリーレポート", props.Text{
					Size:  16,
					Style: fontstyle.Bold,
					Align: align.Center,
				}),
			),
		),
		row.New(8).Add(
			col.New(12).Add(
				text.New(
					fmt.Sprintf("%s 〜 %s",
						startDate.Format("2006-01-02"),
						endDate.Format("2006-01-02")),
					props.Text{
						Size:  12,
						Align: align.Center,
					}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("生成日時: %s", time.Now().Format("2006-01-02 15:04:05")), props.Text{
					Size:  9,
					Align: align.Center,
				}),
			),
		),
	)
}

func (s *IncidentPDFService) addStatisticsOverview(m core.Maroto, stats *SummaryStats) {
	m.AddRows(
		row.New(4),
		row.New(10).Add(
			col.New(12).Add(
				text.New("統計サマリー", props.Text{
					Size:  13,
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	// Total incidents
	m.AddRow(7,
		col.New(4).Add(
			text.New("総インシデント数:", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
			}),
		),
		col.New(8).Add(
			text.New(fmt.Sprintf("%d件", stats.TotalIncidents), props.Text{
				Size: 10,
			}),
		),
	)

	// Resolved count
	m.AddRow(7,
		col.New(4).Add(
			text.New("解決済み:", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
			}),
		),
		col.New(8).Add(
			text.New(fmt.Sprintf("%d件", stats.ResolvedCount), props.Text{
				Size:  10,
				Color: &props.Color{Red: 34, Green: 197, Blue: 94},
			}),
		),
	)

	// Average MTTR
	if stats.AverageMTTR > 0 {
		m.AddRow(7,
			col.New(4).Add(
				text.New("平均復旧時間:", props.Text{
					Size:  10,
					Style: fontstyle.Bold,
				}),
			),
			col.New(8).Add(
				text.New(fmt.Sprintf("%.1f時間", stats.AverageMTTR), props.Text{
					Size: 10,
				}),
			),
		)
	}

	// SLA violations
	if stats.SLAViolatedCount > 0 {
		m.AddRow(7,
			col.New(4).Add(
				text.New("SLA違反:", props.Text{
					Size:  10,
					Style: fontstyle.Bold,
				}),
			),
			col.New(8).Add(
				text.New(fmt.Sprintf("%d件", stats.SLAViolatedCount), props.Text{
					Size:  10,
					Color: &props.Color{Red: 220, Green: 38, Blue: 38},
				}),
			),
		)
	}

	// Severity breakdown
	m.AddRows(
		row.New(4),
		row.New(8).Add(
			col.New(12).Add(
				text.New("重要度別", props.Text{
					Size:  11,
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	for severity, count := range stats.BySeverity {
		m.AddRow(6,
			col.New(4).Add(
				text.New(fmt.Sprintf("  %s:", severity), props.Text{
					Size: 9,
				}),
			),
			col.New(8).Add(
				text.New(fmt.Sprintf("%d件", count), props.Text{
					Size: 9,
				}),
			),
		)
	}

	// Status breakdown
	m.AddRows(
		row.New(4),
		row.New(8).Add(
			col.New(12).Add(
				text.New("ステータス別", props.Text{
					Size:  11,
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	for status, count := range stats.ByStatus {
		m.AddRow(6,
			col.New(4).Add(
				text.New(fmt.Sprintf("  %s:", status), props.Text{
					Size: 9,
				}),
			),
			col.New(8).Add(
				text.New(fmt.Sprintf("%d件", count), props.Text{
					Size: 9,
				}),
			),
		)
	}
}

func (s *IncidentPDFService) addIncidentsBySeverity(m core.Maroto, incidents []*domain.Incident) {
	m.AddRows(
		row.New(6),
		row.New(10).Add(
			col.New(12).Add(
				text.New("重要度別インシデント概要", props.Text{
					Size:  13,
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	// Group by severity
	critical := []*domain.Incident{}
	high := []*domain.Incident{}

	for _, incident := range incidents {
		switch incident.Severity {
		case domain.SeverityCritical:
			critical = append(critical, incident)
		case domain.SeverityHigh:
			high = append(high, incident)
		}
	}

	// Add critical incidents
	if len(critical) > 0 {
		m.AddRow(7,
			col.New(12).Add(
				text.New(fmt.Sprintf("Critical (%d件)", len(critical)), props.Text{
					Size:  10,
					Style: fontstyle.Bold,
					Color: &props.Color{Red: 220, Green: 38, Blue: 38},
				}),
			),
		)
		for _, inc := range critical {
			m.AddRow(5,
				col.New(12).Add(
					text.New(fmt.Sprintf("  • %s (ID: %d)", inc.Title, inc.ID), props.Text{
						Size: 9,
					}),
				),
			)
		}
	}

	// Add high incidents
	if len(high) > 0 {
		m.AddRow(7,
			col.New(12).Add(
				text.New(fmt.Sprintf("High (%d件)", len(high)), props.Text{
					Size:  10,
					Style: fontstyle.Bold,
					Color: &props.Color{Red: 239, Green: 68, Blue: 68},
				}),
			),
		)
		for _, inc := range high {
			m.AddRow(5,
				col.New(12).Add(
					text.New(fmt.Sprintf("  • %s (ID: %d)", inc.Title, inc.ID), props.Text{
						Size: 9,
					}),
				),
			)
		}
	}
}

func (s *IncidentPDFService) addIncidentsTable(m core.Maroto, incidents []*domain.Incident) {
	m.AddRows(
		row.New(6),
		row.New(10).Add(
			col.New(12).Add(
				text.New("インシデント一覧", props.Text{
					Size:  13,
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	// Table header
	m.AddRow(7,
		col.New(1).Add(
			text.New("ID", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
			}),
		),
		col.New(5).Add(
			text.New("タイトル", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New("重要度", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New("ステータス", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New("検出日", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
			}),
		),
	)

	// Table rows
	for _, incident := range incidents {
		m.AddRow(6,
			col.New(1).Add(
				text.New(fmt.Sprintf("%d", incident.ID), props.Text{
					Size: 8,
				}),
			),
			col.New(5).Add(
				text.New(truncateString(incident.Title, 40), props.Text{
					Size: 8,
				}),
			),
			col.New(2).Add(
				text.New(string(incident.Severity), props.Text{
					Size:  8,
					Color: s.getSeverityColor(incident.Severity),
				}),
			),
			col.New(2).Add(
				text.New(string(incident.Status), props.Text{
					Size:  8,
					Color: s.getStatusColor(incident.Status),
				}),
			),
			col.New(2).Add(
				text.New(incident.DetectedAt.Format("01/02"), props.Text{
					Size: 8,
				}),
			),
		)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
