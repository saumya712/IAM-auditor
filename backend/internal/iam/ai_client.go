package iam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GeneratePolicyResponse is the response returned by the AI service's
// /ai/generate-policy endpoint.
type GeneratePolicyResponse struct {
	Policy      map[string]interface{} `json:"policy"`
	Explanation string                 `json:"explanation"`
}

// Finding represents a single security finding from a policy audit.
type Finding struct {
	Description string `json:"description"`
	RiskLevel   string `json:"risk_level"`
	Remediation string `json:"remediation"`
}

// AuditPolicyResponse is the response returned by the AI service's
// /ai/audit-policy endpoint.
type AuditPolicyResponse struct {
	Findings []Finding `json:"findings"`
}

// AIClient is the interface for communicating with the Python AI service.
type AIClient interface {
	GeneratePolicy(ctx context.Context, prompt string) (*GeneratePolicyResponse, error)
	AuditPolicy(ctx context.Context, policy string) (*AuditPolicyResponse, error)
}

// HTTPAIClient implements AIClient by making HTTP POST requests to the AI service.
type HTTPAIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPAIClient creates a new HTTPAIClient targeting the given base URL.
// The underlying HTTP client has a 30-second timeout.
func NewHTTPAIClient(baseURL string) *HTTPAIClient {
	return &HTTPAIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GeneratePolicy sends a POST request to {baseURL}/ai/generate-policy with the
// given prompt and decodes the response into a GeneratePolicyResponse.
// Returns an error if the request fails or the AI service returns a non-2xx status.
func (c *HTTPAIClient) GeneratePolicy(ctx context.Context, prompt string) (*GeneratePolicyResponse, error) {
	body, err := json.Marshal(map[string]string{"prompt": prompt})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal generate-policy request: %w", err)
	}

	resp, err := c.post(ctx, c.baseURL+"/ai/generate-policy", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var result GeneratePolicyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode generate-policy response: %w", err)
	}
	return &result, nil
}

// AuditPolicy sends a POST request to {baseURL}/ai/audit-policy with the given
// policy string and decodes the response into an AuditPolicyResponse.
// Returns an error if the request fails or the AI service returns a non-2xx status.
func (c *HTTPAIClient) AuditPolicy(ctx context.Context, policy string) (*AuditPolicyResponse, error) {
	body, err := json.Marshal(map[string]string{"policy": policy})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal audit-policy request: %w", err)
	}

	resp, err := c.post(ctx, c.baseURL+"/ai/audit-policy", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return nil, err
	}

	var result AuditPolicyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode audit-policy response: %w", err)
	}
	return &result, nil
}

// post is a helper that sends a JSON POST request and returns the raw response.
// The caller is responsible for closing resp.Body.
func (c *HTTPAIClient) post(ctx context.Context, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request to %s: %w", url, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI service unreachable at %s: %w", url, err)
	}
	return resp, nil
}

// checkStatus reads the response body and returns a descriptive error when the
// HTTP status code is outside the 2xx range.
func checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("AI service returned status %d (could not read body: %v)", resp.StatusCode, err)
	}
	return fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(bodyBytes))
}
