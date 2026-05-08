package ui

import (
	"fmt"
	"strings"
	"time"

	"deepseek-monitor-tui/api"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	accent    = lipgloss.Color("#6366f1") // indigo
	positive  = lipgloss.Color("#22c55e") // green
	negative  = lipgloss.Color("#ef4444") // red
	warning   = lipgloss.Color("#f59e0b") // amber
	subtle    = lipgloss.Color("#6b7280") // gray
	muted     = lipgloss.Color("#9ca3af")
	bg        = lipgloss.Color("#1f2937")
	borderCol = lipgloss.Color("#374151")
	text      = lipgloss.Color("#e5e7eb")
	textDim   = lipgloss.Color("#9ca3af")

	// Base - no fixed width, adapts to terminal
	BaseStyle = lipgloss.NewStyle().Padding(0, 0)

	// Header
	HeaderStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Padding(0, 1).
			Margin(0, 0, 0, 0)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(textDim).
			Padding(0, 0).
			Margin(0, 0, 0, 0).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(borderCol)

	// Section header (used by auth.go)
	SectionStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Padding(0, 0).
			Margin(0, 0, 0, 0)

	// Legacy width constant (used by auth.go layout)
	Width = 80

	// Error
	ErrorStyle = lipgloss.NewStyle().
			Foreground(negative).
			Padding(0, 1)

	// Loading
	LoadingStyle = lipgloss.NewStyle().
			Foreground(warning).
			Padding(0, 1).
			Italic(true)

	// Endpoint status indicators
	StatusOkStyle  = lipgloss.NewStyle().Foreground(positive).Bold(true)
	StatusBadStyle = lipgloss.NewStyle().Foreground(negative).Bold(true)

	// Card styles
	cardTitleStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Padding(0, 0)

	cardSectionStyle = lipgloss.NewStyle().
				Foreground(accent).
				Bold(true).
				Padding(0, 0).
				Margin(0, 0, 0, 0)

	cardKeyStyle = lipgloss.NewStyle().
			Foreground(textDim)

	cardValStyle = lipgloss.NewStyle().
			Foreground(text)
)

type UsageViewOptions struct {
	ShowRequests  bool
	ShowCached    bool
	ShowNonCached bool
}

// mkRow renders a key-value row.
func mkRow(key, value string, cardW int) string {
	keyR := cardKeyStyle.Render(key)
	valR := lipgloss.NewStyle().
		Foreground(text).
		Width(cardW - lipgloss.Width(key) - 4).
		Align(lipgloss.Right).
		Render(value)
	return lipgloss.JoinHorizontal(lipgloss.Top, keyR, valR)
}

// mkSection renders a sub-section header.
func mkSection(name string, cardW int) string {
	return cardSectionStyle.Render(name)
}

// mkCardParts renders components of a card for merging.
func mkCardParts(title string, cardW int, rows ...string) (top, body, bottom string) {
	bCol := lipgloss.NewStyle().Foreground(borderCol)
	titleText := " " + title + " "
	titleR := cardTitleStyle.Render(titleText)
	calcWidth := cardW - lipgloss.Width(titleText) - 3
	if calcWidth < 0 {
		calcWidth = 0
	}
	top = bCol.Render("╭─") + titleR + bCol.Render(strings.Repeat("─", calcWidth)+"╮")

	var bodyRows []string
	for _, r := range rows {
		padding := cardW - lipgloss.Width(r) - 2
		if padding < 0 {
			padding = 0
		}
		bodyRows = append(bodyRows, bCol.Render("│")+" "+r+strings.Repeat(" ", padding-1)+bCol.Render("│"))
	}
	body = strings.Join(bodyRows, "\n")
	bottom = bCol.Render("╰" + strings.Repeat("─", cardW-2) + "╯")
	return
}

// calcCardWidths returns the number of columns and inner width per card.
func calcCardWidths(totalW int) (nCols int, cardW int) {
	nCols = 1
	cardW = totalW - 2
	if cardW > 60 {
		cardW = 60
	}
	if cardW < 20 {
		cardW = 20
	}
	return nCols, cardW
}

// ColumnsLayout builds the main content area with stacked cards sharing borders.
func ColumnsLayout(balance []api.BalanceInfo, usage *api.UsageResponse, cost *api.CostResponse, totalW int, selectedTab int, usageView UsageViewOptions) string {
	_, cardW := calcCardWidths(totalW)

	type item struct {
		title string
		rows  []string
		has   bool
	}
	var items []item

	var bRows []string
	for _, b := range balance {
		bRows = append(bRows, mkRow("Total", b.TotalBalance, cardW))
	}
	items = append(items, item{"Balance", bRows, len(balance) > 0})

	// Tab-aware titles
	usageTitle := "Usage (30d)"
	costTitle := "Cost (30d)"
	if selectedTab == 1 {
		usageTitle = "Usage (1d)"
		costTitle = "Cost (1d)"
	}

	var cRows []string
	if cost != nil {
		cRows = append(cRows, mkRow("Total", fmt.Sprintf("%s %.2f", cost.Currency, cost.TotalCost), cardW))
	}
	items = append(items, item{costTitle, cRows, cost != nil})

	var uRows []string
	if usage != nil {
		for _, m := range usage.ModelBreakdown {
			if m.Requests == 0 && m.TotalTokens == 0 {
				continue
			}

			nonCachedTokens := m.InputTokens - m.CachedTokens
			if nonCachedTokens < 0 {
				nonCachedTokens = 0
			}

			var modelRows []string
			if usageView.ShowRequests {
				modelRows = append(modelRows, mkRow("Requests", FormatNumber(m.Requests), cardW))
			}
			if usageView.ShowCached {
				modelRows = append(modelRows, mkRow("Cached", FormatTokens(m.CachedTokens), cardW))
			}
			if usageView.ShowNonCached {
				modelRows = append(modelRows, mkRow("Non-Cached", FormatTokens(nonCachedTokens), cardW))
			}
			modelRows = append(modelRows, mkRow("Output", FormatTokens(m.OutputTokens), cardW))

			uRows = append(uRows, mkSection(strings.TrimPrefix(m.Model, "deepseek-"), cardW))
			uRows = append(uRows, modelRows...)
		}
	}
	items = append(items, item{usageTitle, uRows, len(uRows) > 0})

	var result []string
	var active []item
	for _, it := range items {
		if it.has {
			active = append(active, it)
		}
	}

	for i, it := range active {
		top, body, bottom := mkCardParts(it.title, cardW, it.rows...)
		if i > 0 {
			// Shared separator with Title: ├─ TITLE ====┤
			bCol := lipgloss.NewStyle().Foreground(borderCol)
			titleText := " " + it.title + " "
			titleR := cardTitleStyle.Render(titleText)
			titleW := lipgloss.Width(titleText)

			// Subtract horizontal border parts: ├─ (2 chars) and ┤ (1 char)
			calcWidth := cardW - titleW - 3
			if calcWidth < 0 {
				calcWidth = 0
			}
			sep := bCol.Render("├─") + titleR + bCol.Render(strings.Repeat("─", calcWidth)+"┤")
			result = append(result, sep)
		} else {
			result = append(result, top)
		}
		result = append(result, body)
		if i == len(active)-1 {
			result = append(result, bottom)
		}
	}
	return strings.Join(result, "\n")
}

// StatusBarCompact renders a multi-line status bar with updated time and next refresh.
func StatusBarCompact(updatedAt time.Time, nextRefresh time.Duration, errCount int, totalW int) string {
	left := fmt.Sprintf("Updated: %s", updatedAt.Format("15:04:05"))
	right := fmt.Sprintf("Next: %ds", int(nextRefresh.Seconds()))

	if errCount > 0 {
		right += StatusBadStyle.Render(fmt.Sprintf("  ⚠ %d err", errCount))
	}

	leftR := lipgloss.NewStyle().Foreground(textDim).Render(left)
	rightR := lipgloss.NewStyle().Foreground(textDim).Render(right)

	return StatusBarStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, leftR, rightR),
	)
}

// --- Format helpers ---

// FormatNumber formats large numbers with commas.
func FormatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

// FormatTokens formats token counts with K/M suffix.
func FormatTokens(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.2fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// ErrorList renders a list of error messages.
func ErrorList(errors []string) string {
	if len(errors) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(ErrorStyle.Render("⚠ Errors:\n"))
	for _, e := range errors {
		sb.WriteString(ErrorStyle.Render("  • " + e + "\n"))
	}
	return sb.String()
}

// LoadingView renders the initial loading state.
func LoadingView(msg string) string {
	return lipgloss.NewStyle().
		Width(60).
		Align(lipgloss.Center).
		Padding(2, 0).
		Render(LoadingStyle.Render("⏳ " + msg))
}
