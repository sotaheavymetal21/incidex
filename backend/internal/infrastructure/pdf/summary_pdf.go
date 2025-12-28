package pdf

import (
	"fmt"
	"incidex/internal/domain"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
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

	// Add statistics cards
	s.addStatisticsCards(m, stats)

	// Add incidents table
	s.addEnhancedIncidentsTable(m, incidents)

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
	// Title
	m.AddRow(20,
		col.New(12).Add(
			text.New("Incident Summary Report", props.Text{
				Size:  20,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: &props.Color{Red: 30, Green: 58, Blue: 138},
			}),
		),
	)

	// Period
	m.AddRow(12,
		col.New(12).Add(
			text.New(
				fmt.Sprintf("Period: %s - %s",
					startDate.Format("2006-01-02"),
					endDate.Format("2006-01-02")),
				props.Text{
					Size:  14,
					Align: align.Center,
					Color: &props.Color{Red: 75, Green: 85, Blue: 99},
				}),
		),
	)

	// Generated timestamp
	m.AddRow(8,
		col.New(12).Add(
			text.New(fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")), props.Text{
				Size:  9,
				Align: align.Center,
				Color: &props.Color{Red: 107, Green: 114, Blue: 128},
			}),
		),
	)

	// Separator
	m.AddRow(10,
		col.New(12).Add(
			text.New("═══════════════════════════════════════════════════════", props.Text{
				Size:  10,
				Align: align.Center,
				Color: &props.Color{Red: 200, Green: 200, Blue: 200},
			}),
		),
	)
}

func (s *IncidentPDFService) addStatisticsCards(m core.Maroto, stats *SummaryStats) {
	m.AddRow(5)

	// Main stats row
	m.AddRow(30,
		col.New(4).Add(
			text.New("Total Incidents", props.Text{
				Size:  9,
				Align: align.Center,
				Color: &props.Color{Red: 107, Green: 114, Blue: 128},
			}),
			text.New(fmt.Sprintf("%d", stats.TotalIncidents), props.Text{
				Size:  24,
				Style: fontstyle.Bold,
				Align: align.Center,
				Top:   5,
				Color: &props.Color{Red: 59, Green: 130, Blue: 246},
			}),
		),
		col.New(4).Add(
			text.New("Resolved", props.Text{
				Size:  9,
				Align: align.Center,
				Color: &props.Color{Red: 107, Green: 114, Blue: 128},
			}),
			text.New(fmt.Sprintf("%d", stats.ResolvedCount), props.Text{
				Size:  24,
				Style: fontstyle.Bold,
				Align: align.Center,
				Top:   5,
				Color: &props.Color{Red: 34, Green: 197, Blue: 94},
			}),
		),
		col.New(4).Add(
			text.New("Avg Resolution Time", props.Text{
				Size:  9,
				Align: align.Center,
				Color: &props.Color{Red: 107, Green: 114, Blue: 128},
			}),
			text.New(fmt.Sprintf("%.1fh", stats.AverageMTTR), props.Text{
				Size:  20,
				Style: fontstyle.Bold,
				Align: align.Center,
				Top:   5,
				Color: &props.Color{Red: 168, Green: 85, Blue: 247},
			}),
		),
	)

	m.AddRow(8)

	// Severity breakdown header
	m.AddRow(15,
		col.New(12).Add(
			text.New("Severity Breakdown", props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 30, Green: 58, Blue: 138},
			}),
		),
	)

	critical := stats.BySeverity["critical"]
	high := stats.BySeverity["high"]
	medium := stats.BySeverity["medium"]
	low := stats.BySeverity["low"]

	m.AddRow(10,
		col.New(3).Add(
			text.New("Critical:", props.Text{
				Size:  11,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%d", critical), props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 220, Green: 38, Blue: 38},
			}),
		),
		col.New(3).Add(
			text.New("High:", props.Text{
				Size:  11,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%d", high), props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 249, Green: 115, Blue: 22},
			}),
		),
	)

	m.AddRow(10,
		col.New(3).Add(
			text.New("Medium:", props.Text{
				Size:  11,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%d", medium), props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 251, Green: 191, Blue: 36},
			}),
		),
		col.New(3).Add(
			text.New("Low:", props.Text{
				Size:  11,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%d", low), props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 34, Green: 197, Blue: 94},
			}),
		),
	)

	m.AddRow(8)

	// Status breakdown header
	m.AddRow(15,
		col.New(12).Add(
			text.New("Status Breakdown", props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 30, Green: 58, Blue: 138},
			}),
		),
	)

	resolved := stats.ByStatus["resolved"]
	open := stats.ByStatus["open"]
	investigating := stats.ByStatus["investigating"]
	closed := stats.ByStatus["closed"]

	m.AddRow(10,
		col.New(3).Add(
			text.New("Resolved:", props.Text{
				Size:  11,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%d", resolved), props.Text{
				Size:  14,
				Color: &props.Color{Red: 34, Green: 197, Blue: 94},
			}),
		),
		col.New(3).Add(
			text.New("Open:", props.Text{
				Size:  11,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%d", open), props.Text{
				Size:  14,
				Color: &props.Color{Red: 239, Green: 68, Blue: 68},
			}),
		),
	)

	m.AddRow(10,
		col.New(3).Add(
			text.New("Investigating:", props.Text{
				Size:  11,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%d", investigating), props.Text{
				Size:  14,
				Color: &props.Color{Red: 251, Green: 191, Blue: 36},
			}),
		),
		col.New(3).Add(
			text.New("Closed:", props.Text{
				Size:  11,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New(fmt.Sprintf("%d", closed), props.Text{
				Size:  14,
				Color: &props.Color{Red: 107, Green: 114, Blue: 128},
			}),
		),
	)

	m.AddRow(8)
}

func (s *IncidentPDFService) addEnhancedIncidentsTable(m core.Maroto, incidents []*domain.Incident) {
	// Table header
	m.AddRow(18,
		col.New(12).Add(
			text.New("Incident Details", props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 30, Green: 58, Blue: 138},
			}),
		),
	)

	// Separator
	m.AddRow(5,
		col.New(12).Add(
			text.New("───────────────────────────────────────────────────────", props.Text{
				Size:  10,
				Align: align.Center,
				Color: &props.Color{Red: 200, Green: 200, Blue: 200},
			}),
		),
	)

	// Column headers
	m.AddRow(10,
		col.New(1).Add(
			text.New("ID", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
		col.New(4).Add(
			text.New("Title", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
			}),
		),
		col.New(2).Add(
			text.New("Severity", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
		col.New(2).Add(
			text.New("Status", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
		col.New(3).Add(
			text.New("Detected At", props.Text{
				Size:  10,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
	)

	// Separator under header
	m.AddRow(3,
		col.New(12).Add(
			text.New("═══════════════════════════════════════════════════════", props.Text{
				Size:  8,
				Align: align.Center,
				Color: &props.Color{Red: 156, Green: 163, Blue: 175},
			}),
		),
	)

	// Table rows
	for _, incident := range incidents {
		// Handle empty title
		displayTitle := incident.Title
		if displayTitle == "" {
			displayTitle = "(No Title)"
		}

		m.AddRow(10,
			col.New(1).Add(
				text.New(fmt.Sprintf("%d", incident.ID), props.Text{
					Size:  9,
					Align: align.Center,
				}),
			),
			col.New(4).Add(
				text.New(truncateString(displayTitle, 35), props.Text{
					Size: 9,
				}),
			),
			col.New(2).Add(
				text.New(string(incident.Severity), props.Text{
					Size:  9,
					Align: align.Center,
					Style: fontstyle.Bold,
					Color: s.getSeverityColor(incident.Severity),
				}),
			),
			col.New(2).Add(
				text.New(formatStatus(string(incident.Status)), props.Text{
					Size:  9,
					Align: align.Center,
					Color: s.getStatusColor(incident.Status),
				}),
			),
			col.New(3).Add(
				text.New(incident.DetectedAt.Format("2006-01-02 15:04"), props.Text{
					Size:  9,
					Align: align.Center,
				}),
			),
		)

		// Light separator
		m.AddRow(2,
			col.New(12).Add(
				text.New("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -", props.Text{
					Size:  6,
					Align: align.Center,
					Color: &props.Color{Red: 229, Green: 231, Blue: 235},
				}),
			),
		)
	}
}

func formatStatus(status string) string {
	switch status {
	case "open":
		return "Open"
	case "investigating":
		return "Investigating"
	case "resolved":
		return "Resolved"
	case "closed":
		return "Closed"
	default:
		return status
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
