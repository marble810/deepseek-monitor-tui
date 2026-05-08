package api

import (
	"fmt"
	"time"
)

// ---- Shared response types ----

type BalanceInfo struct {
	Currency        string `json:"currency"`
	TotalBalance    string `json:"total_balance"`
	GrantedBalance  string `json:"granted_balance"`
	ToppedUpBalance string `json:"topped_up_balance"`
}

type BalanceResponse struct {
	IsAvailable  bool          `json:"is_available"`
	BalanceInfos []BalanceInfo `json:"balance_infos"`
}

// ---- Usage (third-party claimed endpoint) ----

type UsageResponse struct {
	TotalRequests      int     `json:"total_requests"`
	SuccessfulRequests int     `json:"successful_requests"`
	FailedRequests     int     `json:"failed_requests"`
	SuccessRate        float64 `json:"success_rate"`
	TotalTokens        int     `json:"total_tokens"`
	InputTokens        int                  `json:"input_tokens"`
	OutputTokens       int                  `json:"output_tokens"`
	CachedTokens       int                  `json:"cached_tokens"`
	ModelBreakdown    []ModelUsageBreakdown `json:"model_breakdown,omitempty"`
}

// ---- Cost (third-party claimed endpoint) ----

type CostByModel map[string]float64

// ModelUsageBreakdown holds per-model usage data for TUI display.
type ModelUsageBreakdown struct {
	Model        string `json:"model"`
	Requests     int    `json:"requests"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	CachedTokens int    `json:"cached_tokens"`
	TotalTokens  int    `json:"total_tokens"`
}

type CostResponse struct {
	TotalCost  float64     `json:"total_cost"`
	Currency   string      `json:"currency"`
	CostByModel CostByModel `json:"cost_by_model,omitempty"`
}

// ---- Combined result ----

type FetchResult struct {
	Balance   *BalanceResponse
	Usage     *UsageResponse
	Usage1d   *UsageResponse
	Cost      *CostResponse
	Cost1d    *CostResponse
	Errors    []string         // non-fatal per-endpoint errors
	Endpoints map[string]bool  // which endpoints succeeded
	FetchedAt time.Time
}

// StatusSummary returns a one-line summary of which endpoints worked.
func (r *FetchResult) StatusSummary() string {
	ok, fail := 0, 0
	for _, v := range r.Endpoints {
		if v {
			ok++
		} else {
			fail++
		}
	}
	if fail == 0 {
		return fmt.Sprintf("✅ all %d endpoints", ok)
	}
	return fmt.Sprintf("✅ %d  ❌ %d", ok, fail)
}
