package testutil

import (
	"testing"
	"time"

	"deepseek-monitor-tui/api"
)

func TestLatestDayOnOrBeforeIgnoresFuturePlaceholders(t *testing.T) {
	now := time.Date(2026, time.May, 8, 12, 0, 0, 0, time.UTC)
	days := []api.DayUsage{
		{
			Date: "2026-05-08",
			Data: []api.ModelUsage{{
				Model: "deepseek-v4-pro",
				Usage: []api.UsageItem{{Type: "REQUEST", Amount: "260"}},
			}},
		},
		{
			Date: "2026-05-31",
			Data: []api.ModelUsage{{
				Model: "deepseek-v4-pro",
				Usage: []api.UsageItem{{Type: "REQUEST", Amount: "0"}},
			}},
		},
	}

	got := api.LatestDayOnOrBefore(days, now)
	if got == nil {
		t.Fatal("LatestDayOnOrBefore returned nil")
	}
	if got.Date != "2026-05-08" {
		t.Fatalf("expected 2026-05-08, got %s", got.Date)
	}
}

func TestUsageResponse1dUsesCurrentDayData(t *testing.T) {
	now := time.Date(2026, time.May, 8, 12, 0, 0, 0, time.UTC)
	resp := &api.UsageAmountResponse{}
	resp.Data.BizData.Days = []api.DayUsage{
		{
			Date: "2026-05-08",
			Data: []api.ModelUsage{{
				Model: "deepseek-v4-pro",
				Usage: []api.UsageItem{
					{Type: "PROMPT_CACHE_HIT_TOKEN", Amount: "10"},
					{Type: "PROMPT_CACHE_MISS_TOKEN", Amount: "5"},
					{Type: "RESPONSE_TOKEN", Amount: "3"},
					{Type: "REQUEST", Amount: "2"},
				},
			}},
		},
		{
			Date: "2026-05-31",
			Data: []api.ModelUsage{{
				Model: "deepseek-v4-pro",
				Usage: []api.UsageItem{
					{Type: "PROMPT_CACHE_HIT_TOKEN", Amount: "0"},
					{Type: "PROMPT_CACHE_MISS_TOKEN", Amount: "0"},
					{Type: "RESPONSE_TOKEN", Amount: "0"},
					{Type: "REQUEST", Amount: "0"},
				},
			}},
		},
	}

	day := api.LatestDayOnOrBefore(resp.Data.BizData.Days, now)
	if day == nil {
		t.Fatal("selected day was nil")
	}

	got := api.UsageResponseFromModels(day.Data)
	if got.TotalRequests != 2 {
		t.Fatalf("expected 2 requests, got %d", got.TotalRequests)
	}
	if got.InputTokens != 15 {
		t.Fatalf("expected 15 input tokens, got %d", got.InputTokens)
	}
	if got.CachedTokens != 10 {
		t.Fatalf("expected 10 cached tokens, got %d", got.CachedTokens)
	}
	if got.OutputTokens != 3 {
		t.Fatalf("expected 3 output tokens, got %d", got.OutputTokens)
	}
	if got.TotalTokens != 18 {
		t.Fatalf("expected 18 total tokens, got %d", got.TotalTokens)
	}
}

func TestCostResponse1dUsesCurrentDayData(t *testing.T) {
	now := time.Date(2026, time.May, 8, 12, 0, 0, 0, time.UTC)
	resp := &api.UsageCostResponse{}
	resp.Data.BizData = []api.UsageAmountData{
		{
			Days: []api.DayUsage{
				{
					Date: "2026-05-08",
					Data: []api.ModelUsage{{
						Model: "deepseek-v4-pro",
						Usage: []api.UsageItem{
							{Type: "PROMPT_CACHE_HIT_TOKEN", Amount: "0.25"},
							{Type: "PROMPT_CACHE_MISS_TOKEN", Amount: "1.50"},
							{Type: "RESPONSE_TOKEN", Amount: "0.75"},
						},
					}},
				},
				{
					Date: "2026-05-31",
					Data: []api.ModelUsage{{
						Model: "deepseek-v4-pro",
						Usage: []api.UsageItem{
							{Type: "PROMPT_CACHE_HIT_TOKEN", Amount: "0"},
							{Type: "PROMPT_CACHE_MISS_TOKEN", Amount: "0"},
							{Type: "RESPONSE_TOKEN", Amount: "0"},
						},
					}},
				},
			},
		},
	}

	day := api.LatestDayOnOrBefore(resp.Data.BizData[0].Days, now)
	if day == nil {
		t.Fatal("selected cost day was nil")
	}

	got := &api.CostResponse{Currency: "CNY", CostByModel: make(api.CostByModel)}
	api.CostResponseFromModels(day.Data, got)
	if got.TotalCost != 2.5 {
		t.Fatalf("expected total cost 2.5, got %v", got.TotalCost)
	}
	if got.CostByModel["deepseek-v4-pro"] != 2.5 {
		t.Fatalf("expected model cost 2.5, got %v", got.CostByModel["deepseek-v4-pro"])
	}
}

func TestToBalanceResponseAggregatesNormalAndBonusWallets(t *testing.T) {
	resp := &api.UserSummaryResponse{}
	resp.Data.BizData.NormalWallets = []api.WalletInfo{{
		Currency: "CNY",
		Balance:  "12.5",
	}}
	resp.Data.BizData.BonusWallets = []api.WalletInfo{{
		Currency: "CNY",
		Balance:  "3.25",
	}}

	got := resp.ToBalanceResponse()
	if !got.IsAvailable {
		t.Fatal("expected balance response to be available")
	}
	if len(got.BalanceInfos) != 1 {
		t.Fatalf("expected 1 balance row, got %d", len(got.BalanceInfos))
	}
	info := got.BalanceInfos[0]
	if info.Currency != "CNY" {
		t.Fatalf("expected CNY currency, got %s", info.Currency)
	}
	if info.ToppedUpBalance != "12.5" {
		t.Fatalf("expected topped up balance 12.5, got %s", info.ToppedUpBalance)
	}
	if info.GrantedBalance != "3.25" {
		t.Fatalf("expected granted balance 3.25, got %s", info.GrantedBalance)
	}
	if info.TotalBalance != "15.75" {
		t.Fatalf("expected total balance 15.75, got %s", info.TotalBalance)
	}
}
