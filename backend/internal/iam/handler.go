package iam

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"iam-advisor/backend/internal/db"
	"iam-advisor/backend/internal/models"
)

// Handler holds the AI client dependency for IAM endpoints.
type Handler struct {
	aiClient AIClient
}

// NewHandler creates a new Handler with the given AIClient.
func NewHandler(aiClient AIClient) *Handler {
	return &Handler{aiClient: aiClient}
}

type adviseRequest struct {
	Prompt string `json:"prompt"`
}

type auditRequest struct {
	Policy string `json:"policy"`
}

// Advise handles POST /api/iam/advise.
// It validates the prompt, proxies to the AI service, persists history, and returns the result.
func (h *Handler) Advise(c *gin.Context) {
	var req adviseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	userID, _ := c.Get("user_id")

	result, err := h.aiClient.GeneratePolicy(c.Request.Context(), req.Prompt)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI service unavailable"})
		return
	}

	// Encode policy map to JSON string for storage.
	policyJSON, _ := json.Marshal(result.Policy)

	record := models.PolicyHistory{
		UserID:              userID.(uint),
		Type:                "advise",
		InputPromptOrPolicy: req.Prompt,
		GeneratedPolicy:     string(policyJSON),
		AnalysisReport:      "",
	}
	db.DB.Create(&record)

	c.JSON(http.StatusOK, result)
}

// Audit handles POST /api/iam/audit.
// It validates the policy JSON, proxies to the AI service, persists history, and returns findings.
func (h *Handler) Audit(c *gin.Context) {
	var req auditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if req.Policy == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "policy is required"})
		return
	}
	if !json.Valid([]byte(req.Policy)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	userID, _ := c.Get("user_id")

	result, err := h.aiClient.AuditPolicy(c.Request.Context(), req.Policy)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "AI service unavailable"})
		return
	}

	// Encode findings to JSON string for storage.
	findingsJSON, _ := json.Marshal(result.Findings)

	record := models.PolicyHistory{
		UserID:              userID.(uint),
		Type:                "audit",
		InputPromptOrPolicy: req.Policy,
		GeneratedPolicy:     "",
		AnalysisReport:      string(findingsJSON),
	}
	db.DB.Create(&record)

	c.JSON(http.StatusOK, result)
}
