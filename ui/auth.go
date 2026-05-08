package ui

import "github.com/charmbracelet/lipgloss"

// stepStyle wraps individual step items with subtle foreground and left padding.
var (
	stepStyle = lipgloss.NewStyle().
			Foreground(textDim).
			PaddingLeft(2)

	stepNumStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	codeStyle = lipgloss.NewStyle().
			Foreground(warning).
			Background(bg).
			Padding(0, 1).
			Italic(true)

	keyHintStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Padding(0, 1).
			Background(lipgloss.Color("#2d2d5e"))

	infoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderCol).
			Padding(1, 2).
			Margin(0, 0, 1, 0)

	instructionsBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(1, 2).
			Width(Width - 6)

	successBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(positive).
			Padding(1, 2).
			Margin(0, 0, 1, 0)

	hintBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(warning).
			Padding(1, 2).
			Margin(0, 0, 1, 0)

	hintTextStyle = lipgloss.NewStyle().
			Foreground(warning).
			PaddingLeft(2)
)

// TokenPromptView renders the UI when a platform Bearer token is needed.
// hint is an optional message shown at the top (e.g., error guidance).
func TokenPromptView(hint string) string {
	title := SectionStyle.Render("Token Authentication Required")

	var hintBlock string
	if hint != "" {
		hintBlock = hintBoxStyle.Render(
			hintTextStyle.Render(hint),
		) + "\n"
	}

	desc := lipgloss.NewStyle().
		Foreground(textDim).
		PaddingLeft(2).
		Render("Paste your platform.deepseek.com Bearer token to authenticate API requests.")

	step1 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stepNumStyle.Render("1."),
		stepStyle.Render("Open "+codeStyle.Render("platform.deepseek.com")+" and log in to your account"),
	)

	step2 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stepNumStyle.Render("2."),
		stepStyle.Render("Open DevTools ("+codeStyle.Render("F12")+") and switch to the Network tab"),
	)

	step3 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stepNumStyle.Render("3."),
		stepStyle.Render("Find any XHR request ("+codeStyle.Render("get_user_summary")+", etc.) and click it"),
	)

	step4 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stepNumStyle.Render("4."),
		stepStyle.Render("In the Request Headers section, find the Authorization header"),
	)

	step5 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stepNumStyle.Render("5."),
		stepStyle.Render("Copy the value after "+codeStyle.Render("Bearer ")+" (do NOT include the word Bearer)"),
	)

	step6 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stepNumStyle.Render("6."),
		stepStyle.Render("Return here and press:"),
	)
	step6hint := lipgloss.JoinHorizontal(
		lipgloss.Top,
		stepStyle.Render(""),
		keyHintStyle.Render(" t "),
		lipgloss.NewStyle().Foreground(text).Render(" to paste the token from clipboard"),
	)

	steps := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		step1,
		"",
		step2,
		"",
		step3,
		"",
		step4,
		"",
		step5,
		"",
		step6,
		step6hint,
		"",
	)

	instructions := instructionsBox.Render(steps)

	note := lipgloss.NewStyle().
		Foreground(subtle).
		PaddingLeft(2).
		Render("Your token is stored locally and never sent outside this app.")

	keybinds := lipgloss.NewStyle().
		Foreground(text).
		Padding(1, 2).
		Render(infoBoxStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				lipgloss.NewStyle().Foreground(accent).Bold(true).Render("Keyboard Shortcuts"),
				"",
				keyHintStyle.Render(" t ")+"  Paste Bearer token from clipboard",
				"",
				keyHintStyle.Render(" q ")+"  Quit",
			),
		))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		hintBlock,
		desc,
		"",
		instructions,
		"",
		note,
		"",
		keybinds,
	)

	return BaseStyle.Render(content)
}

// AuthSuccessView renders a brief confirmation after successful authentication.
func AuthSuccessView(hasToken bool) string {
	title := SectionStyle.Render("Authentication Complete")

	var tokenStatus string

	if hasToken {
		tokenStatus = StatusOkStyle.Render("  Bearer token set")
	} else {
		tokenStatus = StatusBadStyle.Render("  Bearer token missing")
	}

	box := successBoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(positive).Bold(true).Render("Ready"),
			"",
			tokenStatus,
		),
	)

	nextHint := lipgloss.NewStyle().Foreground(textDim).PaddingLeft(2).Render("Starting monitoring...")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		box,
		"",
		nextHint,
	)

	return BaseStyle.Render(content)
}
