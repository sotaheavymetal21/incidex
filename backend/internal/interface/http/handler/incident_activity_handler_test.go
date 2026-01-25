package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"incidex/internal/domain"
	"incidex/internal/testutil/mocks"
	"incidex/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIncidentActivityHandler_AddComment(t *testing.T) {
	t.Parallel()

	t.Run("successfully adds comment", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		activityRepo.On("Create", mock.Anything).Return(nil)

		reqBody := AddCommentRequest{
			Comment: "This is a test comment",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/1/comments", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("userID", uint(1))

		handler.AddComment(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["message"], "Comment added")
	})

	t.Run("fails with invalid incident ID", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/invalid/comments", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}
		c.Set("userID", uint(1))

		handler.AddComment(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails without authentication", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		reqBody := AddCommentRequest{Comment: "Test comment"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/1/comments", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		// No userID set

		handler.AddComment(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("fails with empty comment", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		reqBody := AddCommentRequest{Comment: ""}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/1/comments", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("userID", uint(1))

		handler.AddComment(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestIncidentActivityHandler_GetActivities(t *testing.T) {
	t.Parallel()

	t.Run("successfully returns activities", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		activities := []*domain.IncidentActivity{
			{
				ID:           1,
				IncidentID:   1,
				UserID:       1,
				ActivityType: domain.ActivityTypeComment,
				Comment:      "Test comment",
				CreatedAt:    time.Now(),
			},
		}

		activityRepo.On("FindByIncidentID", uint(1), 50).Return(activities, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/incidents/1/activities", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetActivities(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []*domain.IncidentActivity
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 1)
	})

	t.Run("accepts custom limit", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		activityRepo.On("FindByIncidentID", uint(1), 10).Return([]*domain.IncidentActivity{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/incidents/1/activities?limit=10", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetActivities(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("fails with invalid incident ID", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/incidents/invalid/activities", nil)
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.GetActivities(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		activityRepo.On("FindByIncidentID", uint(1), 50).Return(nil, domain.ErrDatabase("database error", nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/incidents/1/activities", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetActivities(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestIncidentActivityHandler_AddTimelineEvent(t *testing.T) {
	t.Parallel()

	t.Run("successfully adds timeline event", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		activityRepo.On("Create", mock.Anything).Return(nil)

		eventTime := time.Now().Format(time.RFC3339)
		reqBody := AddTimelineEventRequest{
			EventType:   "detected",
			EventTime:   eventTime,
			Description: "Issue detected in production",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/1/timeline", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("userID", uint(1))

		handler.AddTimelineEvent(c)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("fails with invalid event time format", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		reqBody := AddTimelineEventRequest{
			EventType:   "detected",
			EventTime:   "invalid-time",
			Description: "Issue detected",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/1/timeline", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("userID", uint(1))

		handler.AddTimelineEvent(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid event_time format")
	})

	t.Run("fails without authentication", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		reqBody := AddTimelineEventRequest{
			EventType:   "detected",
			EventTime:   time.Now().Format(time.RFC3339),
			Description: "Test",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/1/timeline", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		// No userID set

		handler.AddTimelineEvent(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("fails with invalid event type", func(t *testing.T) {
		t.Parallel()

		activityRepo := mocks.NewMockIncidentActivityRepository()
		incidentRepo := mocks.NewMockIncidentRepository()
		userRepo := mocks.NewMockUserRepository()
		activityUsecase := usecase.NewIncidentActivityUsecase(activityRepo, incidentRepo, userRepo, nil)
		handler := NewIncidentActivityHandler(activityUsecase)

		reqBody := map[string]string{
			"event_type":  "invalid_type",
			"event_time":  time.Now().Format(time.RFC3339),
			"description": "Test",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/incidents/1/timeline", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		c.Set("userID", uint(1))

		handler.AddTimelineEvent(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
