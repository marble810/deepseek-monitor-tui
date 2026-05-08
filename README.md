# dpskmon

A terminal UI for monitoring your [DeepSeek Platform](https://platform.deepseek.com) API usage — balance, token consumption, and request counts — all in one place.

![demo](https://raw.githubusercontent.com/marble810/deepseek-monitor-tui/main/demo.gif)

---

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Uninstall](#uninstall)
- [Development](#development)

---

## Installation

### Homebrew (macOS / Linux) — Recommended

```sh
brew tap marble810/tap
brew install dpskmon
```

### Pre-built Binary

Download the latest release for your platform from the [Releases](https://github.com/marble810/deepseek-monitor-tui/releases) page, extract, and move the binary to a directory in your `PATH`:

```sh
# macOS arm64 example
curl -fsSL https://github.com/marble810/deepseek-monitor-tui/releases/latest/download/dpskmon_0.1.0_darwin_arm64.tar.gz \
  | tar xz
sudo mv dpskmon /usr/local/bin/
```

### Build from Source

Requires [Go 1.21+](https://go.dev/dl/).

```sh
git clone https://github.com/marble810/deepseek-monitor-tui.git
cd deepseek-monitor-tui
make build          # produces ./dpskmon
make install        # copies to ~/.local/bin (set INSTALL_PATH to override)
```

---

## Configuration

`dpskmon` requires a **Platform Bearer token** from [platform.deepseek.com](https://platform.deepseek.com). It is loaded from the first source found, in this order:

| Priority | Source |
|----------|--------|
| 1 | `./deepseek-monitor-tui.json` (working directory) |
| 2 | `~/.config/deepseek-monitor-tui/config.json` |
| 3 | `DEEPSEEK_PLATFORM_TOKEN` environment variable |

### Getting Your Bearer Token

1. Log in to [platform.deepseek.com](https://platform.deepseek.com).
2. Open DevTools (`F12`) → **Network** tab.
3. Click any XHR request (e.g. `get_user_summary`).
4. In **Request Headers**, find the `Authorization` header.
5. Copy the value **after** `Bearer ` (do **not** include the word `Bearer`).

### Interactive Setup (Recommended)

Just run `dpskmon` without any config — the app will guide you through pasting the token:

```
Press  t  to paste Bearer token from clipboard
```

The token is saved to `~/.config/deepseek-monitor-tui/config.json` automatically.

### Manual Config File

Create `~/.config/deepseek-monitor-tui/config.json` (or `./deepseek-monitor-tui.json`):

```json
{
  "platform_token": "your-platform-bearer-token-here",
  "default_period": "30d",
  "query_interval_seconds": 30,
  "show_requests": true,
  "show_cached": true,
  "show_non_cached": true
}
```

### Environment Variable

```sh
export DEEPSEEK_PLATFORM_TOKEN="your-platform-bearer-token-here"
dpskmon
```

### Config Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `platform_token` | string | — | **Required.** Bearer token from platform.deepseek.com |
| `default_period` | string | `"30d"` | Default time period tab (`"30d"` or `"1d"`) |
| `query_interval_seconds` | int | `30` | Auto-refresh interval in seconds |
| `show_requests` | bool | `true` | Show total request count column |
| `show_cached` | bool | `true` | Show cached token usage column |
| `show_non_cached` | bool | `true` | Show non-cached token usage column |

---

## Usage

```sh
dpskmon
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Toggle between 30-day and 1-day views |
| `r` | Manually refresh data |
| `t` | Paste a new Bearer token from clipboard |
| `q` / `Ctrl+C` | Quit |

The app auto-refreshes on the configured interval (default: 30 seconds). The countdown is shown in the footer.

---

## Uninstall

### Homebrew

```sh
brew uninstall dpskmon
brew untap marble810/tap   # optional, removes the tap
```

### Manual

```sh
rm "$(which dpskmon)"
```

### Remove Config

```sh
rm -f ~/.config/deepseek-monitor-tui/config.json
rmdir ~/.config/deepseek-monitor-tui   # only if empty
```

---

## Development

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [gh CLI](https://cli.github.com) (for releases)

### Common Tasks

```sh
make build            # build ./dpskmon
make run              # build and run
make fmt              # gofmt all packages
make tidy             # go mod tidy
make clean            # remove ./dpskmon binary
make install          # install to ~/.local/bin (override with INSTALL_PATH=...)
```

### Running Tests

```sh
go test ./...
```

### Project Structure

```
.
├── main.go               # Entry point, TUI model (bubbletea)
├── api/                  # DeepSeek API client & scraper
├── config/               # Config loading, saving, defaults
├── ui/                   # Lipgloss styles and view components
├── cmd/                  # (reserved for future sub-commands)
├── internal/             # Internal utilities
├── scripts/
│   ├── build-release-assets.sh   # Cross-compile & package archives
│   ├── publish-release.sh        # Create GitHub Release via gh
│   └── update-homebrew-formula.sh # Generate Formula/dpskmon.rb
└── .github/workflows/
    └── release.yml       # Tag-triggered CI: build → release → tap update
```

### Releasing a New Version

1. Commit all changes to `main`.
2. Create and push a semver tag:

```sh
git tag v1.2.3
git push origin v1.2.3
```

The `release.yml` workflow fires automatically and:
- Runs `go test ./...`
- Cross-compiles for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`
- Creates a GitHub Release with archives and `checksums.txt`
- Updates the Homebrew formula in `marble810/homebrew-tap`

> **Required secret:** `HOMEBREW_TAP_TOKEN` — a GitHub PAT with write access to `marble810/homebrew-tap`, set in the source repo's Actions secrets.

### Local Release (Manual)

```sh
make release-assets TAG=v1.2.3    # builds dist/ archives
make publish-release TAG=v1.2.3   # uploads to GitHub Releases via gh
bash scripts/update-homebrew-formula.sh v1.2.3 ~/path/to/homebrew-tap
```
