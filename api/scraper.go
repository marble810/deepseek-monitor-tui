package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const PlatformBaseURL = "https://platform.deepseek.com"

// Real platform internal API endpoints discovered from browser network inspection.
const PlatformUserSummaryURL = PlatformBaseURL + "/api/v0/users/get_user_summary"
const PlatformUsageAmountURL = PlatformBaseURL + "/api/v0/usage/amount"
const PlatformUsageCostURL = PlatformBaseURL + "/api/v0/usage/cost"

// ErrAuthFailed is the sentinel error returned when the platform token is
// invalid or expired. Callers can check with errors.Is to prompt re-auth.
var ErrAuthFailed = errors.New("authentication failed: platform token expired or invalid")

// ---------------------------------------------------------------------------
// Platform-specific response types
// ---------------------------------------------------------------------------

// WalletInfo holds balance info for a single wallet (normal or bonus).
type WalletInfo struct {
	Currency        string `json:"currency"`
	Balance         string `json:"balance"`
	TokenEstimation string `json:"token_estimation"`
}

// MonthlyCost holds a monthly cost entry.
type MonthlyCost struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

// UserSummaryResponse wraps the /api/v0/users/get_user_summary response.
type UserSummaryResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		BizCode int    `json:"biz_code"`
		BizMsg  string `json:"biz_msg"`
		BizData struct {
			CurrentToken                  int           `json:"current_token"`
			MonthlyUsage                  string        `json:"monthly_usage"`
			TotalUsage                    int           `json:"total_usage"`
			NormalWallets                 []WalletInfo  `json:"normal_wallets"`
			BonusWallets                  []WalletInfo  `json:"bonus_wallets"`
			TotalAvailableTokenEstimation string        `json:"total_available_token_estimation"`
			MonthlyCosts                  []MonthlyCost `json:"monthly_costs"`
			MonthlyTokenUsage             string        `json:"monthly_token_usage"`
		} `json:"biz_data"`
	} `json:"data"`
}

// UsageItem holds a single usage/cost metric type and its value.
type UsageItem struct {
	Type   string `json:"type"`
	Amount string `json:"amount"`
}

// ModelUsage holds usage/cost breakdown for a single model.
type ModelUsage struct {
	Model string      `json:"model"`
	Usage []UsageItem `json:"usage"`
}

// DayUsage holds usage/cost data for a single day.
type DayUsage struct {
	Date string       `json:"date"`
	Data []ModelUsage `json:"data"`
}

// UsageAmountData holds total and per-day usage/cost data.
type UsageAmountData struct {
	Total []ModelUsage `json:"total"`
	Days  []DayUsage   `json:"days"`
}

// UsageAmountResponse wraps the /api/v0/usage/amount response.
type UsageAmountResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		BizCode int             `json:"biz_code"`
		BizMsg  string          `json:"biz_msg"`
		BizData UsageAmountData `json:"biz_data"`
	} `json:"data"`
}

// UsageCostResponse wraps the /api/v0/usage/cost response.
// Same structure as UsageAmountResponse but amounts are monetary costs.
// Note: the cost endpoint returns biz_data as an ARRAY, unlike the amount
// endpoint which returns it as a single object.
type UsageCostResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		BizCode int               `json:"biz_code"`
		BizMsg  string            `json:"biz_msg"`
		BizData []UsageAmountData `json:"biz_data"` // array, not single object
	} `json:"data"`
}

// ---------------------------------------------------------------------------
// Scraper
// ---------------------------------------------------------------------------

// Scraper fetches usage, cost, and balance data from platform.deepseek.com
// using Bearer token authentication.
type Scraper struct {
	PlatformToken string
	HTTPClient    *http.Client
}

// NewScraper creates a Scraper with a 15-second timeout and redirects disabled.
func NewScraper(platformToken string) *Scraper {
	return &Scraper{
		PlatformToken: platformToken,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
			// Do not follow redirects — treat them as errors.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return fmt.Errorf("unexpected redirect to %s", req.URL)
			},
		},
	}
}

// FetchAll collects balance, usage, and cost from the platform using a single
// Bearer token.
func (s *Scraper) FetchAll() *FetchResult {
	r := &FetchResult{
		Endpoints: make(map[string]bool),
		FetchedAt: time.Now(),
	}

	summary, err := s.FetchUserSummary()
	if err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("balance: %v", err))
		r.Endpoints["balance"] = false
		if errors.Is(err, ErrAuthFailed) {
			r.Endpoints["usage"] = false
			r.Endpoints["cost"] = false
			return r
		}
	} else {
		r.Balance = summary.ToBalanceResponse()
		r.Endpoints["balance"] = true
	}

	now := time.Now()
	if usageResp, err := s.FetchUsageAmount(now.Year(), int(now.Month())); err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("platform usage: %v", err))
		r.Endpoints["usage"] = false
	} else {
		r.Usage = usageResp.ToUsageResponse()
		r.Usage1d = usageResp.ToUsageResponse1d()
		r.Endpoints["usage"] = true
	}

	if costResp, err := s.FetchUsageCost(now.Year(), int(now.Month())); err != nil {
		r.Errors = append(r.Errors, fmt.Sprintf("platform cost: %v", err))
		r.Endpoints["cost"] = false
	} else {
		r.Cost = costResp.ToCostResponse()
		r.Cost1d = costResp.ToCostResponse1d()
		r.Endpoints["cost"] = true
	}

	return r
}

// FetchUserSummary gets balance and monthly usage summary from the platform.
func (s *Scraper) FetchUserSummary() (*UserSummaryResponse, error) {
	resp, err := s.doGet(PlatformUserSummaryURL)
	if err != nil {
		return nil, fmt.Errorf("fetch user summary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user summary endpoint: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result UserSummaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode user summary response: %w", err)
	}
	return &result, nil
}

// FetchUsageAmount gets per-model token and request counts for a given month.
// month is 1-12, year e.g. 2026.
func (s *Scraper) FetchUsageAmount(year, month int) (*UsageAmountResponse, error) {
	url := fmt.Sprintf("%s?month=%d&year=%d", PlatformUsageAmountURL, month, year)
	resp, err := s.doGet(url)
	if err != nil {
		return nil, fmt.Errorf("fetch usage amount: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("usage amount endpoint: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result UsageAmountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode usage amount response: %w", err)
	}
	return &result, nil
}

// FetchUsageCost gets per-model costs for a given month.
// month is 1-12, year e.g. 2026.
func (s *Scraper) FetchUsageCost(year, month int) (*UsageCostResponse, error) {
	url := fmt.Sprintf("%s?month=%d&year=%d", PlatformUsageCostURL, month, year)
	resp, err := s.doGet(url)
	if err != nil {
		return nil, fmt.Errorf("fetch usage cost: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("usage cost endpoint: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result UsageCostResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode usage cost response: %w", err)
	}
	return &result, nil
}

// Validate checks whether the platform token is still accepted by the
// platform. It makes a lightweight authenticated request and returns
// ErrAuthFailed if the server responds with 401 or 403.
func (s *Scraper) Validate() error {
	resp, err := s.doGet(PlatformUserSummaryURL)
	if err != nil {
		// ErrAuthFailed is already returned by doGet for 401/403.
		// Other errors (network, DNS) are real failures.
		return fmt.Errorf("validate session: %w", err)
	}
	resp.Body.Close()

	// Any non-401/403 response (even 4xx/5xx) means the token was accepted.
	return nil
}

// ToUsageResponse converts platform usage amount data into the existing
// UsageResponse format used by the TUI.
func (r *UsageAmountResponse) ToUsageResponse() *UsageResponse {
	var totalRequests, inputTokens, outputTokens, cachedTokens int
	var breakdown []ModelUsageBreakdown
	for _, model := range r.Data.BizData.Total {
		var mRequests, mInput, mOutput, mCached int
		for _, item := range model.Usage {
			n := 0
			fmt.Sscanf(item.Amount, "%d", &n)
			switch item.Type {
			case "REQUEST":
				totalRequests += n
				mRequests += n
			case "PROMPT_TOKEN", "PROMPT_CACHE_MISS_TOKEN":
				inputTokens += n
				mInput += n
			case "PROMPT_CACHE_HIT_TOKEN":
				cachedTokens += n
				mCached += n
				inputTokens += n
				mInput += n
			case "RESPONSE_TOKEN":
				outputTokens += n
				mOutput += n
			}
		}
		breakdown = append(breakdown, ModelUsageBreakdown{
			Model:        model.Model,
			Requests:     mRequests,
			InputTokens:  mInput,
			OutputTokens: mOutput,
			CachedTokens: mCached,
			TotalTokens:  mInput + mOutput,
		})
	}
	return &UsageResponse{
		TotalRequests:      totalRequests,
		SuccessfulRequests: totalRequests,
		FailedRequests:     0,
		SuccessRate:        100.0,
		TotalTokens:        inputTokens + outputTokens,
		InputTokens:        inputTokens,
		OutputTokens:       outputTokens,
		CachedTokens:       cachedTokens,
		ModelBreakdown:     breakdown,
	}
}

func latestDayOnOrBefore(days []DayUsage, now time.Time) *DayUsage {
	if len(days) == 0 {
		return nil
	}

	cutoff := now.Format("2006-01-02")
	var selected *DayUsage
	for i := range days {
		day := &days[i]
		if day.Date > cutoff {
			continue
		}
		if selected == nil || day.Date > selected.Date {
			selected = day
		}
	}

	if selected != nil {
		return selected
	}

	return &days[len(days)-1]
}

func usageResponseFromModels(models []ModelUsage) *UsageResponse {
	var totalRequests, inputTokens, outputTokens, cachedTokens int
	var breakdown []ModelUsageBreakdown
	for _, model := range models {
		var mRequests, mInput, mOutput, mCached int
		for _, item := range model.Usage {
			n := 0
			fmt.Sscanf(item.Amount, "%d", &n)
			switch item.Type {
			case "REQUEST":
				totalRequests += n
				mRequests += n
			case "PROMPT_TOKEN", "PROMPT_CACHE_MISS_TOKEN":
				inputTokens += n
				mInput += n
			case "PROMPT_CACHE_HIT_TOKEN":
				cachedTokens += n
				mCached += n
				inputTokens += n
				mInput += n
			case "RESPONSE_TOKEN":
				outputTokens += n
				mOutput += n
			}
		}
		breakdown = append(breakdown, ModelUsageBreakdown{
			Model:        model.Model,
			Requests:     mRequests,
			InputTokens:  mInput,
			OutputTokens: mOutput,
			CachedTokens: mCached,
			TotalTokens:  mInput + mOutput,
		})
	}

	return &UsageResponse{
		TotalRequests:      totalRequests,
		SuccessfulRequests: totalRequests,
		FailedRequests:     0,
		SuccessRate:        100.0,
		TotalTokens:        inputTokens + outputTokens,
		InputTokens:        inputTokens,
		OutputTokens:       outputTokens,
		CachedTokens:       cachedTokens,
		ModelBreakdown:     breakdown,
	}
}

func costResponseFromModels(models []ModelUsage, result *CostResponse) {
	for _, model := range models {
		var modelCost float64
		for _, item := range model.Usage {
			var val float64
			fmt.Sscanf(item.Amount, "%f", &val)
			modelCost += val
		}
		result.TotalCost += modelCost
		result.CostByModel[model.Model] += modelCost
	}
}

// ToUsageResponse1d converts the latest day's usage into a UsageResponse.
func (r *UsageAmountResponse) ToUsageResponse1d() *UsageResponse {
	latest := latestDayOnOrBefore(r.Data.BizData.Days, time.Now())
	if latest == nil {
		return nil
	}

	return usageResponseFromModels(latest.Data)
}

// ToCostResponse1d converts the latest day's cost into a CostResponse.
func (r *UsageCostResponse) ToCostResponse1d() *CostResponse {
	result := &CostResponse{
		Currency:    "CNY",
		CostByModel: make(CostByModel),
	}

	for _, bizData := range r.Data.BizData {
		latest := latestDayOnOrBefore(bizData.Days, time.Now())
		if latest == nil {
			continue
		}
		costResponseFromModels(latest.Data, result)
	}

	return result
}

// ToCostResponse converts platform cost data into the existing CostResponse
// format used by the TUI.
func (r *UsageCostResponse) ToCostResponse() *CostResponse {
	result := &CostResponse{
		Currency:    "CNY",
		CostByModel: make(CostByModel),
	}
	for _, bizData := range r.Data.BizData {
		costResponseFromModels(bizData.Total, result)
	}
	return result
}

// ToBalanceResponse converts platform user summary data into the existing
// BalanceResponse format used by the TUI.
func (r *UserSummaryResponse) ToBalanceResponse() *BalanceResponse {
	type walletTotals struct {
		granted float64
		topped  float64
	}

	balances := make(map[string]*walletTotals)
	var order []string
	ensureTotals := func(currency string) *walletTotals {
		if totals, ok := balances[currency]; ok {
			return totals
		}
		totals := &walletTotals{}
		balances[currency] = totals
		order = append(order, currency)
		return totals
	}

	parseAmount := func(raw string) float64 {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0
		}
		return value
	}

	var infos []BalanceInfo
	for _, w := range r.Data.BizData.NormalWallets {
		totals := ensureTotals(w.Currency)
		totals.topped += parseAmount(w.Balance)
	}
	for _, w := range r.Data.BizData.BonusWallets {
		totals := ensureTotals(w.Currency)
		totals.granted += parseAmount(w.Balance)
	}
	for _, currency := range order {
		totals := balances[currency]
		infos = append(infos, BalanceInfo{
			Currency:        currency,
			TotalBalance:    strconv.FormatFloat(totals.topped+totals.granted, 'f', -1, 64),
			GrantedBalance:  strconv.FormatFloat(totals.granted, 'f', -1, 64),
			ToppedUpBalance: strconv.FormatFloat(totals.topped, 'f', -1, 64),
		})
	}
	return &BalanceResponse{
		IsAvailable:  len(infos) > 0,
		BalanceInfos: infos,
	}
}

// doGet performs an authenticated GET request using the Bearer token.
// It returns ErrAuthFailed immediately on 401 or 403 responses.
// Includes browser-like headers to avoid WAF blocking.
func (s *Scraper) doGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.PlatformToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-app-version", "20240425.0")
	req.Header.Set("Referer", "https://platform.deepseek.com/usage")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		return nil, ErrAuthFailed
	}

	// Detect WAF block page (returns 200 but content is HTML, not JSON)
	if resp.StatusCode == http.StatusOK {
		ct := resp.Header.Get("Content-Type")
		if strings.Contains(ct, "text/html") {
			_, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("request blocked by WAF (got HTML instead of JSON)")
		}
	}

	return resp, nil
}
