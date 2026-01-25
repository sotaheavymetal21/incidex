package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultSLAHours(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity Severity
		want     int
	}{
		{
			name:     "critical severity returns 4 hours",
			severity: SeverityCritical,
			want:     4,
		},
		{
			name:     "high severity returns 24 hours",
			severity: SeverityHigh,
			want:     24,
		},
		{
			name:     "medium severity returns 72 hours",
			severity: SeverityMedium,
			want:     72,
		},
		{
			name:     "low severity returns 168 hours (1 week)",
			severity: SeverityLow,
			want:     168,
		},
		{
			name:     "unknown severity defaults to 72 hours",
			severity: Severity("unknown"),
			want:     72,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GetDefaultSLAHours(tt.severity)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIncident_CalculateSLADeadline(t *testing.T) {
	t.Parallel()

	detectedAt := time.Date(2025, 1, 25, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		incident       *Incident
		wantDeadline   *time.Time
		wantNil        bool
	}{
		{
			name: "calculates deadline for 4 hour SLA",
			incident: &Incident{
				DetectedAt:               detectedAt,
				SLATargetResolutionHours: 4,
			},
			wantDeadline: func() *time.Time {
				t := detectedAt.Add(4 * time.Hour)
				return &t
			}(),
			wantNil: false,
		},
		{
			name: "calculates deadline for 24 hour SLA",
			incident: &Incident{
				DetectedAt:               detectedAt,
				SLATargetResolutionHours: 24,
			},
			wantDeadline: func() *time.Time {
				t := detectedAt.Add(24 * time.Hour)
				return &t
			}(),
			wantNil: false,
		},
		{
			name: "returns nil for zero SLA hours",
			incident: &Incident{
				DetectedAt:               detectedAt,
				SLATargetResolutionHours: 0,
			},
			wantNil: true,
		},
		{
			name: "returns nil for negative SLA hours",
			incident: &Incident{
				DetectedAt:               detectedAt,
				SLATargetResolutionHours: -1,
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.incident.CalculateSLADeadline()
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, *tt.wantDeadline, *got)
			}
		})
	}
}

func TestIncident_CheckSLAViolation(t *testing.T) {
	t.Parallel()

	now := time.Now()
	pastDeadline := now.Add(-1 * time.Hour)
	futureDeadline := now.Add(1 * time.Hour)

	tests := []struct {
		name     string
		incident *Incident
		want     bool
	}{
		{
			name: "no SLA deadline set - no violation",
			incident: &Incident{
				SLADeadline: nil,
			},
			want: false,
		},
		{
			name: "open incident before deadline - no violation",
			incident: &Incident{
				SLADeadline: &futureDeadline,
				Status:      StatusOpen,
				ResolvedAt:  nil,
			},
			want: false,
		},
		{
			name: "open incident after deadline - violation",
			incident: &Incident{
				SLADeadline: &pastDeadline,
				Status:      StatusOpen,
				ResolvedAt:  nil,
			},
			want: true,
		},
		{
			name: "resolved before deadline - no violation",
			incident: &Incident{
				SLADeadline: &futureDeadline,
				Status:      StatusResolved,
				ResolvedAt:  &now,
			},
			want: false,
		},
		{
			name: "resolved after deadline - violation",
			incident: &Incident{
				SLADeadline: &pastDeadline,
				Status:      StatusResolved,
				ResolvedAt:  &now,
			},
			want: true,
		},
		{
			name: "resolved exactly at deadline - no violation",
			incident: func() *Incident {
				deadline := now
				return &Incident{
					SLADeadline: &deadline,
					Status:      StatusResolved,
					ResolvedAt:  &deadline,
				}
			}(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.incident.CheckSLAViolation()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIncident_GetResolutionTime(t *testing.T) {
	t.Parallel()

	detectedAt := time.Date(2025, 1, 25, 10, 0, 0, 0, time.UTC)
	resolvedAt := time.Date(2025, 1, 25, 14, 30, 0, 0, time.UTC) // 4.5 hours later

	tests := []struct {
		name         string
		incident     *Incident
		wantDuration *time.Duration
	}{
		{
			name: "resolved incident returns duration",
			incident: &Incident{
				DetectedAt: detectedAt,
				ResolvedAt: &resolvedAt,
			},
			wantDuration: func() *time.Duration {
				d := resolvedAt.Sub(detectedAt)
				return &d
			}(),
		},
		{
			name: "unresolved incident returns nil",
			incident: &Incident{
				DetectedAt: detectedAt,
				ResolvedAt: nil,
			},
			wantDuration: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.incident.GetResolutionTime()
			if tt.wantDuration == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, *tt.wantDuration, *got)
			}
		})
	}
}

func TestIncident_IsOpen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{
			name:   "open status is open",
			status: StatusOpen,
			want:   true,
		},
		{
			name:   "investigating status is open",
			status: StatusInvestigating,
			want:   true,
		},
		{
			name:   "resolved status is not open",
			status: StatusResolved,
			want:   false,
		},
		{
			name:   "closed status is not open",
			status: StatusClosed,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			incident := &Incident{Status: tt.status}
			got := incident.IsOpen()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSeverityConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Severity("critical"), SeverityCritical)
	assert.Equal(t, Severity("high"), SeverityHigh)
	assert.Equal(t, Severity("medium"), SeverityMedium)
	assert.Equal(t, Severity("low"), SeverityLow)
}

func TestStatusConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, Status("open"), StatusOpen)
	assert.Equal(t, Status("investigating"), StatusInvestigating)
	assert.Equal(t, Status("resolved"), StatusResolved)
	assert.Equal(t, Status("closed"), StatusClosed)
}
