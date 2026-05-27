package history

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"iam-advisor/backend/internal/db"
	"iam-advisor/backend/internal/models"
)

// historyRecord is the JSON shape returned to the client.
// Using an explicit struct avoids leaking GORM internals (DeletedAt, etc.).
type historyRecord struct {
	ID                  uint      `json:"id"`
	UserID              uint      `json:"user_id"`
	Type                string    `json:"type"`
	InputPromptOrPolicy string    `json:"input_prompt_or_policy"`
	GeneratedPolicy     string    `json:"generated_policy"`
	AnalysisReport      string    `json:"analysis_report"`
	CreatedAt           time.Time `json:"created_at"`
}

// GetHistory handles GET /api/history.
// It returns all PolicyHistory records for the authenticated user, ordered by
// created_at descending. An empty list is returned (not null) when no records exist.
func GetHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var records []models.PolicyHistory
	result := db.DB.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&records)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Map to the response shape; ensure we return [] not null for empty results.
	response := make([]historyRecord, 0, len(records))
	for _, r := range records {
		response = append(response, historyRecord{
			ID:                  r.ID,
			UserID:              r.UserID,
			Type:                r.Type,
			InputPromptOrPolicy: r.InputPromptOrPolicy,
			GeneratedPolicy:     r.GeneratedPolicy,
			AnalysisReport:      r.AnalysisReport,
			CreatedAt:           r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}
