package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_Liveness(t *testing.T) {
	t.Parallel()

	t.Run("returns ok status", func(t *testing.T) {
		t.Parallel()

		// HealthHandler only needs db for Readiness, Liveness doesn't use it
		handler := NewHealthHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/health/liveness", nil)

		handler.Liveness(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
	})
}

// Note: Readiness tests require a real database connection or sqlmock.
// For unit testing purposes, we focus on Liveness which doesn't require DB.
// Integration tests should cover Readiness with actual database connectivity.
