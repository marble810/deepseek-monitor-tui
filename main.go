package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"deepseek-monitor-tui/api"
	"deepseek-monitor-tui/config"
	"deepseek-monitor-tui/ui"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---- Auth state ----

type authState int

const (
	authReady      authState = iota // has a valid bearer token
	authNeedToken                   // showing token prompt
	authValidating                  // testing pasted token
	authSuccess                     // brief success flash
)

// ---- Messages ----

type fetchDoneMsg struct {
	result *api.FetchResult
}

type tickMsg time.Time

type errMsg struct{ err error }

type tokenPastedMsg struct {
	token string
}

type authSuccessDoneMsg struct{}

// ---- Model ----

type model struct {
	scraper     *api.Scraper // nil if no platform token configured
	config      *config.Config
	usageView   ui.UsageViewOptions
	result      *api.FetchResult
	lastFetch   time.Time
	nextRefresh time.Duration
	loading     bool
	selectedTab int // 0: 30d, 1: 1d
	err         error
	width       int
	height      int
	countdown   int
	auth        authState
	authMsg     string
}

func initialModel(cfg *config.Config) model {
	m := model{
		config: cfg,
		usageView: ui.UsageViewOptions{
			ShowRequests:  cfg.RequestsVisible(),
			ShowCached:    cfg.CachedVisible(),
			ShowNonCached: cfg.NonCachedVisible(),
		},
		loading:     true,
		nextRefresh: cfg.QueryInterval(),
		countdown:   int(cfg.QueryInterval().Seconds()),
		selectedTab: cfg.SelectedTab(),
	}

	if cfg.PlatformToken != "" {
		m.scraper = api.NewScraper(cfg.PlatformToken)
		m.auth = authReady
	} else {
		m.auth = authNeedToken
		m.loading = false
	}

	return m
}

// ---- Init ----

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{tickCmd()}
	if m.auth == authReady && m.scraper != nil {
		cmds = append(cmds, doFetch(m.scraper))
	}
	return tea.Batch(cmds...)
}

func doFetch(scraper *api.Scraper) tea.Cmd {
	return func() tea.Msg {
		if scraper == nil {
			return errMsg{err: fmt.Errorf("missing platform bearer token")}
		}
		return fetchDoneMsg{result: scraper.FetchAll()}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ---- Helpers ----

// looksLikeCookie checks if a string looks like an HTTP cookie string
// rather than a Bearer token.
func looksLikeCookie(s string) bool {
	return strings.Contains(s, "smidV2=") ||
		strings.Contains(s, ".thumbcache_") ||
		strings.Contains(s, "HWWAFSESID") ||
		strings.Contains(s, "workspace_ref=") ||
		(strings.Count(s, "=") > 2 && strings.Count(s, ";") > 0)
}

// ---- Update ----

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab":
			// Toggle between 30d and 1d tabs
			m.selectedTab = (m.selectedTab + 1) % 2
			m.config.SetSelectedTab(m.selectedTab)
			if err := m.config.Save(); err != nil {
				m.err = fmt.Errorf("save config: %w", err)
			}
			return m, nil

		case "t":
			// Paste platform Bearer token from clipboard
			token, err := clipboard.ReadAll()
			if err != nil {
				m.authMsg = "无法读取剪贴板，请重试"
				return m, nil
			}
			token = strings.TrimSpace(token)
			if token == "" {
				m.authMsg = "剪贴板为空，请先复制 Bearer Token"
				return m, nil
			}
			// Detect if user accidentally pasted a cookie instead of Bearer token
			if looksLikeCookie(token) {
				m.auth = authNeedToken
				m.authMsg = "检测到粘贴的是 Cookie，不是 Bearer Token！\n\n请从 DevTools → Network → 任意 XHR 请求 → Request Headers →\nAuthorization 中复制 Bearer 后面的值（不含 Bearer 前缀）。\n\nToken 示例：nCot7qMNK77gxM48ZOiIB...\nCookie 示例：smidV2=...（不对！）"
				return m, nil
			}
			m.auth = authValidating
			m.authMsg = "校验 Token 中..."
			return m, func() tea.Msg {
				s := api.NewScraper(token)
				if err := s.Validate(); err != nil {
					return errMsg{err: fmt.Errorf("token 校验失败: %w", err)}
				}
				return tokenPastedMsg{token: token}
			}

		case "r":
			// Manual refresh (only when not in auth flow)
			if m.auth != authReady || m.scraper == nil {
				return m, nil
			}
			m.loading = true
			m.err = nil
			m.countdown = int(m.config.QueryInterval().Seconds())
			m.nextRefresh = m.config.QueryInterval()
		default:
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case fetchDoneMsg:
		m.loading = false
		m.result = msg.result
		m.lastFetch = msg.result.FetchedAt
		m.countdown = int(m.config.QueryInterval().Seconds())
		m.nextRefresh = m.config.QueryInterval()
		m.err = nil

		// Check if any scraper error is ErrAuthFailed and switch to authNeedToken
		if m.scraper != nil {
			for _, errStr := range msg.result.Errors {
				if strings.Contains(errStr, api.ErrAuthFailed.Error()) {
					m.auth = authNeedToken
					m.authMsg = "Token 已过期，请重新粘贴 Bearer Token（按 t）"
					m.loading = false
					break
				}
			}
		}

	case tickMsg:
		if m.auth != authReady || m.scraper == nil {
			return m, tickCmd()
		}
		m.countdown--
		if m.countdown < 0 {
			m.countdown = 0
		}
		m.nextRefresh = time.Duration(m.countdown) * time.Second
		if m.countdown <= 0 {
			m.countdown = int(m.config.QueryInterval().Seconds())
			m.nextRefresh = m.config.QueryInterval()
			m.loading = true
			m.err = nil
			return m, tea.Batch(doFetch(m.scraper), tickCmd())
		}
		return m, tickCmd()

	case tokenPastedMsg:
		// Token is valid, save to config
		if err := m.config.SetPlatformToken(msg.token); err != nil {
			m.err = fmt.Errorf("save token: %w", err)
			return m, nil
		}
		m.scraper = api.NewScraper(msg.token)
		m.auth = authSuccess
		m.authMsg = ""
		m.loading = true
		m.countdown = int(m.config.QueryInterval().Seconds())
		m.nextRefresh = m.config.QueryInterval()
		return m, tea.Tick(1200*time.Millisecond, func(time.Time) tea.Msg {
			return authSuccessDoneMsg{}
		})

	case authSuccessDoneMsg:
		if m.auth != authSuccess {
			return m, nil
		}
		m.auth = authReady
		m.loading = true
		m.err = nil
		return m, doFetch(m.scraper)

	case errMsg:
		m.err = msg.err
		m.loading = false
		// If we were validating, go back to authNeedToken so user can retry
		if m.auth == authValidating {
			m.auth = authNeedToken
			m.authMsg = "Token 校验失败，请检查是否复制了正确的值。\n\n提示：Bearer Token 在 DevTools Network 标签页的 Authorization 请求头中，\n复制 Bearer 后面的部分（不含 Bearer 前缀）。"
		}
	}

	return m, nil
}

// ---- View ----

func (m model) View() string {
	// Auth flow screens
	switch m.auth {
	case authNeedToken:
		return ui.TokenPromptView(m.authMsg)
	case authValidating:
		return ui.BaseStyle.Render("\n" + ui.LoadingView(m.authMsg) + "\n")
	case authSuccess:
		return ui.AuthSuccessView(m.config.PlatformToken != "")
	}

	// Loading state
	if m.loading && m.result == nil {
		return uiLoadingView()
	}

	// Error state (no result at all)
	if m.err != nil && m.result == nil {
		return uiErrorView(m.err)
	}

	// Main view
	return uiMainView(m)
}

// ---- Sub-views ----

func uiLoadingView() string {
	return ui.BaseStyle.Render(
		"\n" + ui.LoadingView("Fetching DeepSeek data...") + "\n",
	)
}

func uiErrorView(err error) string {
	return ui.BaseStyle.Render(
		ui.ErrorStyle.Render(fmt.Sprintf("Error: %v", err)) + "\n\n" +
			"Press 'r' to retry, 'q' to quit.",
	)
}

func uiMainView(m model) string {
	r := m.result

	var sections []string
	balance := []api.BalanceInfo{}
	if r.Balance != nil {
		balance = r.Balance.BalanceInfos
	}

	// Centered Header with Icon
	headerText := "🐳 DeepSeek Monitor"
	header := ui.HeaderStyle.Width(m.width).Align(lipgloss.Center).Render(headerText)
	sections = append(sections, header)

	// Cards in responsive columns
	usage := r.Usage
	cost := r.Cost
	if m.selectedTab == 1 {
		usage = r.Usage1d
		cost = r.Cost1d
	}

	content := ui.ColumnsLayout(
		balance,
		usage,
		cost,
		m.width,
		m.selectedTab,
		m.usageView,
	)
	if content != "" {
		sections = append(sections, content)
	} else {
		sections = append(sections, "\n"+ui.LoadingStyle.Render("No data available yet..."))
	}

	// Errors
	if m.err != nil {
		sections = append(sections, ui.ErrorList([]string{m.err.Error()}))
	}
	if len(r.Errors) > 0 {
		sections = append(sections, ui.ErrorList(r.Errors))
	}

	// Compact status bar
	status := ui.StatusBarCompact(m.lastFetch, m.nextRefresh, len(r.Errors), m.width)

	return ui.BaseStyle.Render(
		strings.Join(sections, "\n") + "\n" +
			status,
	)
}

// ---- Main ----

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		initialModel(cfg),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
