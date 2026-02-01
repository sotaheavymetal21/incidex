package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"incidex/internal/domain"
	"incidex/internal/testutil"
	"incidex/internal/testutil/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createTestStatsUsecase(
	incidentRepo *mocks.MockIncidentRepository,
	cacheRepo *mocks.MockCacheRepository,
) *StatsUsecase {
	return NewStatsUsecase(incidentRepo, cacheRepo)
}

func TestStatsUsecase_GetDashboardStats(t *testing.T) {
	t.Parallel()

	t.Run("returns correct stats with cache miss", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		// Mock cache miss
		cacheRepo.On("Get", ctx, "stats:dashboard:daily").Return("", errors.New("cache miss"))

		// Mock total count
		incidentRepo.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(0).(*int64)
			*count = 100
		}).Return(nil)

		// Mock severity counts
		incidentRepo.On("CountBySeverity", domain.SeverityCritical, mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 10
		}).Return(nil)
		incidentRepo.On("CountBySeverity", domain.SeverityHigh, mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 20
		}).Return(nil)
		incidentRepo.On("CountBySeverity", domain.SeverityMedium, mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 30
		}).Return(nil)
		incidentRepo.On("CountBySeverity", domain.SeverityLow, mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 40
		}).Return(nil)

		// Mock status counts
		incidentRepo.On("CountByStatus", domain.StatusOpen, mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 25
		}).Return(nil)
		incidentRepo.On("CountByStatus", domain.StatusInvestigating, mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 35
		}).Return(nil)
		incidentRepo.On("CountByStatus", domain.StatusResolved, mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 20
		}).Return(nil)
		incidentRepo.On("CountByStatus", domain.StatusClosed, mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 20
		}).Return(nil)

		// Mock recent incidents
		recentIncidents := []*domain.Incident{
			testutil.NewTestIncident(func(i *domain.Incident) { i.ID = 1 }),
			testutil.NewTestIncident(func(i *domain.Incident) { i.ID = 2 }),
		}
		incidentRepo.On("FindRecent", 10).Return(recentIncidents, nil)

		// Mock trend data
		incidentRepo.On("GetAllIncidents").Return([]*domain.Incident{}, nil)

		// Mock cache set
		cacheRepo.On("Set", ctx, "stats:dashboard:daily", mock.AnythingOfType("string"), 5*time.Minute).Return(nil)

		stats, err := usecase.GetDashboardStats("daily")

		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, int64(100), stats.TotalIncidents)
		assert.Equal(t, int64(10), stats.BySeverity["critical"])
		assert.Equal(t, int64(20), stats.BySeverity["high"])
		assert.Equal(t, int64(30), stats.BySeverity["medium"])
		assert.Equal(t, int64(40), stats.BySeverity["low"])
		assert.Equal(t, int64(25), stats.ByStatus["open"])
		assert.Equal(t, int64(35), stats.ByStatus["investigating"])
		assert.Equal(t, int64(20), stats.ByStatus["resolved"])
		assert.Equal(t, int64(20), stats.ByStatus["closed"])
		assert.Len(t, stats.RecentIncidents, 2)

		incidentRepo.AssertExpectations(t)
		cacheRepo.AssertExpectations(t)
	})

	t.Run("returns cached stats on cache hit", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cachedStats := &DashboardStats{
			TotalIncidents: 50,
			BySeverity: map[string]int64{
				"critical": 5,
				"high":     10,
				"medium":   15,
				"low":      20,
			},
			ByStatus: map[string]int64{
				"open":          10,
				"investigating": 15,
				"resolved":      15,
				"closed":        10,
			},
			RecentIncidents: []*domain.Incident{},
			TrendData:       []TrendDataPoint{},
		}
		cachedJSON, _ := json.Marshal(cachedStats)
		cacheRepo.On("Get", ctx, "stats:dashboard:weekly").Return(string(cachedJSON), nil)

		stats, err := usecase.GetDashboardStats("weekly")

		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, int64(50), stats.TotalIncidents)
		assert.Equal(t, int64(5), stats.BySeverity["critical"])

		cacheRepo.AssertExpectations(t)
		// Incident repository should not be called when cache hits
		incidentRepo.AssertNotCalled(t, "Count")
	})

	t.Run("handles empty data", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:dashboard:daily").Return("", errors.New("cache miss"))

		// Mock zero counts
		incidentRepo.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(0).(*int64)
			*count = 0
		}).Return(nil)

		incidentRepo.On("CountBySeverity", mock.AnythingOfType("domain.Severity"), mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 0
		}).Return(nil)

		incidentRepo.On("CountByStatus", mock.AnythingOfType("domain.Status"), mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 0
		}).Return(nil)

		incidentRepo.On("FindRecent", 10).Return([]*domain.Incident{}, nil)
		incidentRepo.On("GetAllIncidents").Return([]*domain.Incident{}, nil)

		cacheRepo.On("Set", ctx, "stats:dashboard:daily", mock.AnythingOfType("string"), 5*time.Minute).Return(nil)

		stats, err := usecase.GetDashboardStats("daily")

		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, int64(0), stats.TotalIncidents)
		assert.Equal(t, int64(0), stats.BySeverity["critical"])
		assert.Len(t, stats.RecentIncidents, 0)
	})

	t.Run("handles database error on Count", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:dashboard:daily").Return("", errors.New("cache miss"))
		incidentRepo.On("Count", mock.AnythingOfType("*int64")).Return(errors.New("database error"))

		stats, err := usecase.GetDashboardStats("daily")

		require.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("handles database error on CountBySeverity", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:dashboard:daily").Return("", errors.New("cache miss"))
		incidentRepo.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(0).(*int64)
			*count = 100
		}).Return(nil)
		incidentRepo.On("CountBySeverity", domain.SeverityCritical, mock.AnythingOfType("*int64")).Return(errors.New("database error"))

		stats, err := usecase.GetDashboardStats("daily")

		require.Error(t, err)
		assert.Nil(t, stats)
	})

	t.Run("handles database error on CountByStatus", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:dashboard:daily").Return("", errors.New("cache miss"))
		incidentRepo.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(0).(*int64)
			*count = 100
		}).Return(nil)
		incidentRepo.On("CountBySeverity", mock.AnythingOfType("domain.Severity"), mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 0
		}).Return(nil)
		incidentRepo.On("CountByStatus", domain.StatusOpen, mock.AnythingOfType("*int64")).Return(errors.New("database error"))

		stats, err := usecase.GetDashboardStats("daily")

		require.Error(t, err)
		assert.Nil(t, stats)
	})

	t.Run("handles database error on FindRecent", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:dashboard:daily").Return("", errors.New("cache miss"))
		incidentRepo.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(0).(*int64)
			*count = 100
		}).Return(nil)
		incidentRepo.On("CountBySeverity", mock.AnythingOfType("domain.Severity"), mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 0
		}).Return(nil)
		incidentRepo.On("CountByStatus", mock.AnythingOfType("domain.Status"), mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 0
		}).Return(nil)
		incidentRepo.On("FindRecent", 10).Return(nil, errors.New("database error"))

		stats, err := usecase.GetDashboardStats("daily")

		require.Error(t, err)
		assert.Nil(t, stats)
	})

	t.Run("continues on cache set error", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:dashboard:daily").Return("", errors.New("cache miss"))
		incidentRepo.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(0).(*int64)
			*count = 10
		}).Return(nil)
		incidentRepo.On("CountBySeverity", mock.AnythingOfType("domain.Severity"), mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 0
		}).Return(nil)
		incidentRepo.On("CountByStatus", mock.AnythingOfType("domain.Status"), mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			count := args.Get(1).(*int64)
			*count = 0
		}).Return(nil)
		incidentRepo.On("FindRecent", 10).Return([]*domain.Incident{}, nil)
		incidentRepo.On("GetAllIncidents").Return([]*domain.Incident{}, nil)

		// Cache set fails, but should not affect the result
		cacheRepo.On("Set", ctx, "stats:dashboard:daily", mock.AnythingOfType("string"), 5*time.Minute).Return(errors.New("cache error"))

		stats, err := usecase.GetDashboardStats("daily")

		require.NoError(t, err) // Should succeed despite cache error
		assert.NotNil(t, stats)
		assert.Equal(t, int64(10), stats.TotalIncidents)
	})
}

func TestStatsUsecase_generateTrendData(t *testing.T) {
	t.Parallel()

	t.Run("generates daily trend data correctly", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		now := time.Now()
		incidents := []*domain.Incident{
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.DetectedAt = now.Add(-1 * 24 * time.Hour)
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.DetectedAt = now.Add(-1 * 24 * time.Hour)
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.DetectedAt = now.Add(-5 * 24 * time.Hour)
			}),
		}

		incidentRepo.On("GetAllIncidents").Return(incidents, nil)

		trendData, err := usecase.generateTrendData("daily")

		require.NoError(t, err)
		assert.Len(t, trendData, 30) // 30 days of data
		assert.NotNil(t, trendData)
	})

	t.Run("generates weekly trend data correctly", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		now := time.Now()
		incidents := []*domain.Incident{
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.DetectedAt = now.Add(-3 * 24 * time.Hour)
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.DetectedAt = now.Add(-10 * 24 * time.Hour)
			}),
		}

		incidentRepo.On("GetAllIncidents").Return(incidents, nil)

		trendData, err := usecase.generateTrendData("weekly")

		require.NoError(t, err)
		assert.Len(t, trendData, 12) // 12 weeks of data
	})

	t.Run("generates monthly trend data correctly", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		now := time.Now()
		incidents := []*domain.Incident{
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.DetectedAt = now.Add(-30 * 24 * time.Hour)
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.DetectedAt = now.Add(-60 * 24 * time.Hour)
			}),
		}

		incidentRepo.On("GetAllIncidents").Return(incidents, nil)

		trendData, err := usecase.generateTrendData("monthly")

		require.NoError(t, err)
		assert.Len(t, trendData, 12) // 12 months of data
	})

	t.Run("handles invalid period by defaulting to daily", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		incidentRepo.On("GetAllIncidents").Return([]*domain.Incident{}, nil)

		trendData, err := usecase.generateTrendData("invalid")

		require.NoError(t, err)
		assert.Len(t, trendData, 30) // Defaults to 30 days
	})

	t.Run("handles empty incidents list", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		incidentRepo.On("GetAllIncidents").Return([]*domain.Incident{}, nil)

		trendData, err := usecase.generateTrendData("daily")

		require.NoError(t, err)
		assert.Len(t, trendData, 30)
		// All counts should be zero
		for _, point := range trendData {
			assert.Equal(t, int64(0), point.Count)
		}
	})

	t.Run("handles database error", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		incidentRepo.On("GetAllIncidents").Return(nil, errors.New("database error"))

		trendData, err := usecase.generateTrendData("daily")

		require.Error(t, err)
		assert.Nil(t, trendData)
	})

	t.Run("filters out old incidents", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		now := time.Now()
		incidents := []*domain.Incident{
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.DetectedAt = now.Add(-1 * 24 * time.Hour) // Within range
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.DetectedAt = now.Add(-100 * 24 * time.Hour) // Outside daily range
			}),
		}

		incidentRepo.On("GetAllIncidents").Return(incidents, nil)

		trendData, err := usecase.generateTrendData("daily")

		require.NoError(t, err)
		// Count total incidents in trend data
		totalCount := int64(0)
		for _, point := range trendData {
			totalCount += point.Count
		}
		assert.Equal(t, int64(1), totalCount) // Only the recent one
	})
}

func TestStatsUsecase_GetSLAMetrics(t *testing.T) {
	t.Parallel()

	t.Run("returns SLA metrics with cache miss", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:sla").Return("", errors.New("cache miss"))

		expectedMetrics := &domain.SLAMetrics{
			TotalIncidents:    100,
			ResolvedIncidents: 80,
			SLAViolatedCount:  20,
			SLAComplianceRate: 75.0,
			AverageMTTR:       12.5,
			MedianMTTR:        10.0,
			CurrentlyOverdue:  5,
		}
		incidentRepo.On("GetSLAMetrics").Return(expectedMetrics, nil)

		cacheRepo.On("Set", ctx, "stats:sla", mock.AnythingOfType("string"), 5*time.Minute).Return(nil)

		metrics, err := usecase.GetSLAMetrics()

		require.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(100), metrics.TotalIncidents)
		assert.Equal(t, int64(80), metrics.ResolvedIncidents)
		assert.Equal(t, int64(20), metrics.SLAViolatedCount)
		assert.Equal(t, 75.0, metrics.SLAComplianceRate)
		assert.Equal(t, 12.5, metrics.AverageMTTR)
		assert.Equal(t, 10.0, metrics.MedianMTTR)
		assert.Equal(t, int64(5), metrics.CurrentlyOverdue)

		incidentRepo.AssertExpectations(t)
		cacheRepo.AssertExpectations(t)
	})

	t.Run("returns cached SLA metrics on cache hit", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cachedMetrics := &domain.SLAMetrics{
			TotalIncidents:    50,
			ResolvedIncidents: 45,
			SLAViolatedCount:  5,
			SLAComplianceRate: 88.9,
			AverageMTTR:       8.5,
			MedianMTTR:        7.0,
			CurrentlyOverdue:  2,
		}
		cachedJSON, _ := json.Marshal(cachedMetrics)
		cacheRepo.On("Get", ctx, "stats:sla").Return(string(cachedJSON), nil)

		metrics, err := usecase.GetSLAMetrics()

		require.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(50), metrics.TotalIncidents)
		assert.Equal(t, 88.9, metrics.SLAComplianceRate)

		cacheRepo.AssertExpectations(t)
		incidentRepo.AssertNotCalled(t, "GetSLAMetrics")
	})

	t.Run("handles database error", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:sla").Return("", errors.New("cache miss"))
		incidentRepo.On("GetSLAMetrics").Return(nil, errors.New("database error"))

		metrics, err := usecase.GetSLAMetrics()

		require.Error(t, err)
		assert.Nil(t, metrics)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("continues on cache set error", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:sla").Return("", errors.New("cache miss"))

		expectedMetrics := &domain.SLAMetrics{
			TotalIncidents: 10,
		}
		incidentRepo.On("GetSLAMetrics").Return(expectedMetrics, nil)

		cacheRepo.On("Set", ctx, "stats:sla", mock.AnythingOfType("string"), 5*time.Minute).Return(errors.New("cache error"))

		metrics, err := usecase.GetSLAMetrics()

		require.NoError(t, err) // Should succeed despite cache error
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(10), metrics.TotalIncidents)
	})
}

func TestStatsUsecase_GetTagStats(t *testing.T) {
	t.Parallel()

	t.Run("returns tag statistics with cache miss", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:tags").Return("", errors.New("cache miss"))

		tag1 := testutil.NewTestTag(func(t *domain.Tag) {
			t.ID = 1
			t.Name = "Production"
			t.Color = "#ff0000"
		})
		tag2 := testutil.NewTestTag(func(t *domain.Tag) {
			t.ID = 2
			t.Name = "Security"
			t.Color = "#00ff00"
		})

		incidents := []*domain.Incident{
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.Tags = []domain.Tag{*tag1, *tag2}
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.Tags = []domain.Tag{*tag1}
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.Tags = []domain.Tag{*tag2}
			}),
		}

		incidentRepo.On("GetAllIncidents").Return(incidents, nil)
		cacheRepo.On("Set", ctx, "stats:tags", mock.AnythingOfType("string"), 10*time.Minute).Return(nil)

		tagStats, err := usecase.GetTagStats()

		require.NoError(t, err)
		assert.Len(t, tagStats, 2)

		// Both tags have count=2, so order is not guaranteed when counts are equal
		// Verify both tags are present with correct data
		var productionTag, securityTag *TagStats
		for i := range tagStats {
			if tagStats[i].TagID == 1 {
				productionTag = &tagStats[i]
			} else if tagStats[i].TagID == 2 {
				securityTag = &tagStats[i]
			}
		}

		require.NotNil(t, productionTag, "Production tag should be present")
		assert.Equal(t, "Production", productionTag.TagName)
		assert.Equal(t, int64(2), productionTag.Count)
		assert.InDelta(t, 66.66, productionTag.Percentage, 0.1)

		require.NotNil(t, securityTag, "Security tag should be present")
		assert.Equal(t, "Security", securityTag.TagName)
		assert.Equal(t, int64(2), securityTag.Count)
		assert.InDelta(t, 66.66, securityTag.Percentage, 0.1)

		incidentRepo.AssertExpectations(t)
		cacheRepo.AssertExpectations(t)
	})

	t.Run("returns cached tag stats on cache hit", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cachedStats := []TagStats{
			{
				TagID:      1,
				TagName:    "Cached Tag",
				TagColor:   "#123456",
				Count:      10,
				Percentage: 100.0,
			},
		}
		cachedJSON, _ := json.Marshal(cachedStats)
		cacheRepo.On("Get", ctx, "stats:tags").Return(string(cachedJSON), nil)

		tagStats, err := usecase.GetTagStats()

		require.NoError(t, err)
		assert.Len(t, tagStats, 1)
		assert.Equal(t, "Cached Tag", tagStats[0].TagName)
		assert.Equal(t, int64(10), tagStats[0].Count)

		cacheRepo.AssertExpectations(t)
		incidentRepo.AssertNotCalled(t, "GetAllIncidents")
	})

	t.Run("handles empty incidents", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:tags").Return("", errors.New("cache miss"))
		incidentRepo.On("GetAllIncidents").Return([]*domain.Incident{}, nil)
		cacheRepo.On("Set", ctx, "stats:tags", mock.AnythingOfType("string"), 10*time.Minute).Return(nil)

		tagStats, err := usecase.GetTagStats()

		require.NoError(t, err)
		assert.Len(t, tagStats, 0)
	})

	t.Run("handles incidents without tags", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:tags").Return("", errors.New("cache miss"))

		incidents := []*domain.Incident{
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.Tags = []domain.Tag{}
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.Tags = []domain.Tag{}
			}),
		}

		incidentRepo.On("GetAllIncidents").Return(incidents, nil)
		cacheRepo.On("Set", ctx, "stats:tags", mock.AnythingOfType("string"), 10*time.Minute).Return(nil)

		tagStats, err := usecase.GetTagStats()

		require.NoError(t, err)
		assert.Len(t, tagStats, 0)
	})

	t.Run("calculates correct percentages", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:tags").Return("", errors.New("cache miss"))

		tag1 := testutil.NewTestTag(func(t *domain.Tag) {
			t.ID = 1
			t.Name = "Tag1"
		})

		// Create 10 incidents total, 3 with tag1
		incidents := []*domain.Incident{}
		for i := 0; i < 3; i++ {
			incidents = append(incidents, testutil.NewTestIncident(func(inc *domain.Incident) {
				inc.Tags = []domain.Tag{*tag1}
			}))
		}
		for i := 0; i < 7; i++ {
			incidents = append(incidents, testutil.NewTestIncident(func(inc *domain.Incident) {
				inc.Tags = []domain.Tag{}
			}))
		}

		incidentRepo.On("GetAllIncidents").Return(incidents, nil)
		cacheRepo.On("Set", ctx, "stats:tags", mock.AnythingOfType("string"), 10*time.Minute).Return(nil)

		tagStats, err := usecase.GetTagStats()

		require.NoError(t, err)
		assert.Len(t, tagStats, 1)
		assert.Equal(t, int64(3), tagStats[0].Count)
		assert.InDelta(t, 30.0, tagStats[0].Percentage, 0.01) // 3/10 = 30%
	})

	t.Run("sorts tags by count descending", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:tags").Return("", errors.New("cache miss"))

		tag1 := testutil.NewTestTag(func(t *domain.Tag) {
			t.ID = 1
			t.Name = "LowCount"
		})
		tag2 := testutil.NewTestTag(func(t *domain.Tag) {
			t.ID = 2
			t.Name = "HighCount"
		})

		incidents := []*domain.Incident{
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.Tags = []domain.Tag{*tag1}
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.Tags = []domain.Tag{*tag2}
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.Tags = []domain.Tag{*tag2}
			}),
			testutil.NewTestIncident(func(i *domain.Incident) {
				i.Tags = []domain.Tag{*tag2}
			}),
		}

		incidentRepo.On("GetAllIncidents").Return(incidents, nil)
		cacheRepo.On("Set", ctx, "stats:tags", mock.AnythingOfType("string"), 10*time.Minute).Return(nil)

		tagStats, err := usecase.GetTagStats()

		require.NoError(t, err)
		assert.Len(t, tagStats, 2)
		// First should be HighCount with 3 incidents
		assert.Equal(t, "HighCount", tagStats[0].TagName)
		assert.Equal(t, int64(3), tagStats[0].Count)
		// Second should be LowCount with 1 incident
		assert.Equal(t, "LowCount", tagStats[1].TagName)
		assert.Equal(t, int64(1), tagStats[1].Count)
	})

	t.Run("handles database error", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:tags").Return("", errors.New("cache miss"))
		incidentRepo.On("GetAllIncidents").Return(nil, errors.New("database error"))

		tagStats, err := usecase.GetTagStats()

		require.Error(t, err)
		assert.Nil(t, tagStats)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("continues on cache set error", func(t *testing.T) {
		t.Parallel()

		incidentRepo := mocks.NewMockIncidentRepository()
		cacheRepo := mocks.NewMockCacheRepository()
		usecase := createTestStatsUsecase(incidentRepo, cacheRepo)

		ctx := context.Background()

		cacheRepo.On("Get", ctx, "stats:tags").Return("", errors.New("cache miss"))
		incidentRepo.On("GetAllIncidents").Return([]*domain.Incident{}, nil)
		cacheRepo.On("Set", ctx, "stats:tags", mock.AnythingOfType("string"), 10*time.Minute).Return(errors.New("cache error"))

		tagStats, err := usecase.GetTagStats()

		require.NoError(t, err) // Should succeed despite cache error
		assert.NotNil(t, tagStats)
	})
}
