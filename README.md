# dpskmon

> **[中文](#中文) | [English](#english)**

一个用于监控 [DeepSeek Platform](https://platform.deepseek.com) API 使用情况的终端 UI —— 余额、Token 消耗、请求次数，一目了然。

*A terminal UI for monitoring your [DeepSeek Platform](https://platform.deepseek.com) API usage — balance, token consumption, and request counts — all in one place.*

![demo](https://raw.githubusercontent.com/marble810/deepseek-monitor-tui/main/demo.gif)

---

## 目录 · Table of Contents

| 中文 | English |
|------|---------|
| [安装](#安装) | [Installation](#installation) |
| [配置](#配置) | [Configuration](#configuration) |
| [使用方法](#使用方法) | [Usage](#usage) |
| [卸载](#卸载) | [Uninstall](#uninstall) |
| [开发](#开发) | [Development](#development) |

---

<h2 id="中文"></h2>

## 中文

<h3 id="安装">安装</h3>

#### Homebrew（macOS / Linux）— 推荐

```sh
brew tap marble810/tap
brew install dpskmon
```

#### 预编译二进制

从 [Releases](https://github.com/marble810/deepseek-monitor-tui/releases) 页面下载对应平台的最新版本，解压后将二进制文件移动到 `PATH` 中的目录：

```sh
# macOS arm64 示例
curl -fsSL https://github.com/marble810/deepseek-monitor-tui/releases/latest/download/dpskmon_0.1.0_darwin_arm64.tar.gz \
  | tar xz
sudo mv dpskmon /usr/local/bin/
```

#### 从源码构建

需要 [Go 1.21+](https://go.dev/dl/)。

```sh
git clone https://github.com/marble810/deepseek-monitor-tui.git
cd deepseek-monitor-tui
make build          # 生成 ./dpskmon
make install        # 安装到 ~/.local/bin（可通过 INSTALL_PATH 覆盖路径）
```

<h3 id="配置">配置</h3>

`dpskmon` 需要一个来自 [platform.deepseek.com](https://platform.deepseek.com) 的 **平台 Bearer Token**。按以下优先级从最先找到的来源加载：

| 优先级 | 来源 |
|--------|------|
| 1 | `./deepseek-monitor-tui.json`（当前工作目录） |
| 2 | `~/.config/deepseek-monitor-tui/config.json` |
| 3 | `DEEPSEEK_PLATFORM_TOKEN` 环境变量 |

#### 获取 Bearer Token

1. 登录 [platform.deepseek.com](https://platform.deepseek.com)。
2. 打开开发者工具（`F12`）→ **网络（Network）** 标签页。
3. 点击任意 XHR 请求（如 `get_user_summary`）。
4. 在 **请求头（Request Headers）** 中找到 `Authorization` 头。
5. 复制 `Bearer ` **之后**的值（**不要**包含 `Bearer ` 这个词）。

#### 交互式配置（推荐）

直接运行 `dpskmon`，无需任何配置 — 程序会引导你粘贴 Token：

```
按  t  从剪贴板粘贴 Bearer Token
```

Token 会自动保存到 `~/.config/deepseek-monitor-tui/config.json`。

#### 手动配置文件

创建 `~/.config/deepseek-monitor-tui/config.json`（或 `./deepseek-monitor-tui.json`）：

```json
{
  "platform_token": "你的平台-bearer-token-在这里",
  "default_period": "30d",
  "query_interval_seconds": 30,
  "show_requests": true,
  "show_cached": true,
  "show_non_cached": true
}
```

#### 环境变量

```sh
export DEEPSEEK_PLATFORM_TOKEN="你的平台-bearer-token-在这里"
dpskmon
```

#### 配置项参考

| 字段 | 类型 | 默认值 | 说明 |
|-------|------|---------|------|
| `platform_token` | string | — | **必填。** platform.deepseek.com 的 Bearer Token |
| `default_period` | string | `"30d"` | 默认时间范围标签页（`"30d"` 或 `"1d"`） |
| `query_interval_seconds` | int | `30` | 自动刷新间隔（秒） |
| `show_requests` | bool | `true` | 显示请求总数列 |
| `show_cached` | bool | `true` | 显示缓存命中 Token 用量列 |
| `show_non_cached` | bool | `true` | 显示缓存未命中 Token 用量列 |

<h3 id="使用方法">使用方法</h3>

```sh
dpskmon
```

#### 键盘快捷键

| 按键 | 操作 |
|------|------|
| `Tab` | 在 30 天和 1 天视图之间切换 |
| `r` | 手动刷新数据 |
| `t` | 从剪贴板粘贴新的 Bearer Token |
| `q` / `Ctrl+C` | 退出 |

程序会按配置的间隔自动刷新（默认：30 秒）。倒计时显示在底部状态栏。

<h3 id="卸载">卸载</h3>

#### Homebrew

```sh
brew uninstall dpskmon
brew untap marble810/tap   # 可选，移除 tap
```

#### 手动卸载

```sh
rm "$(which dpskmon)"
```

#### 删除配置文件

```sh
rm -f ~/.config/deepseek-monitor-tui/config.json
rmdir ~/.config/deepseek-monitor-tui   # 仅当目录为空时
```

<h3 id="开发">开发</h3>

#### 前置条件

- [Go 1.21+](https://go.dev/dl/)
- [gh CLI](https://cli.github.com)（用于发布）

#### 常用命令

```sh
make build            # 构建 ./dpskmon
make run              # 构建并运行
make fmt              # gofmt 格式化所有包
make tidy             # go mod tidy
make clean            # 删除 ./dpskmon 二进制文件
make install          # 安装到 ~/.local/bin（通过 INSTALL_PATH=... 覆盖路径）
```

#### 运行测试

```sh
go test ./...
```

#### 项目结构

```
.
├── main.go               # 入口，TUI 模型（bubbletea）
├── api/                  # DeepSeek API 客户端与数据抓取
├── config/               # 配置加载、保存、默认值
├── ui/                   # Lipgloss 样式与视图组件
├── cmd/                  # （预留，用于未来子命令）
├── internal/             # 内部工具
├── scripts/
│   ├── build-release-assets.sh   # 交叉编译并打包归档
│   ├── publish-release.sh        # 通过 gh 创建 GitHub Release
│   └── update-homebrew-formula.sh # 生成 Formula/dpskmon.rb
└── .github/workflows/
    └── release.yml       # 标签触发 CI：构建 → 发布 → 更新 tap
```

#### 发布新版本

1. 将所有更改提交到 `main` 分支。
2. 创建并推送语义化版本标签：

```sh
git tag v1.2.3
git push origin v1.2.3
```

`release.yml` 工作流会自动触发并执行：
- 运行 `go test ./...`
- 交叉编译 `darwin/amd64`、`darwin/arm64`、`linux/amd64`、`linux/arm64` 平台
- 创建 GitHub Release 并附带归档文件和 `checksums.txt`
- 更新 `marble810/homebrew-tap` 中的 Homebrew formula

> **必需的 Secret：** `HOMEBREW_TAP_TOKEN` — 一个具有 `marble810/homebrew-tap` 写入权限的 GitHub PAT，需在源仓库的 Actions secrets 中配置。

#### 本地手动发布

```sh
make release-assets TAG=v1.2.3    # 构建 dist/ 归档文件
make publish-release TAG=v1.2.3   # 通过 gh 上传到 GitHub Releases
bash scripts/update-homebrew-formula.sh v1.2.3 ~/path/to/homebrew-tap
```

---

<h2 id="english"></h2>

## English

<h3 id="installation">Installation</h3>

#### Homebrew (macOS / Linux) — Recommended

```sh
brew tap marble810/tap
brew install dpskmon
```

#### Pre-built Binary

Download the latest release for your platform from the [Releases](https://github.com/marble810/deepseek-monitor-tui/releases) page, extract, and move the binary to a directory in your `PATH`:

```sh
# macOS arm64 example
curl -fsSL https://github.com/marble810/deepseek-monitor-tui/releases/latest/download/dpskmon_0.1.0_darwin_arm64.tar.gz \
  | tar xz
sudo mv dpskmon /usr/local/bin/
```

#### Build from Source

Requires [Go 1.21+](https://go.dev/dl/).

```sh
git clone https://github.com/marble810/deepseek-monitor-tui.git
cd deepseek-monitor-tui
make build          # produces ./dpskmon
make install        # copies to ~/.local/bin (set INSTALL_PATH to override)
```

<h3 id="configuration">Configuration</h3>

`dpskmon` requires a **Platform Bearer token** from [platform.deepseek.com](https://platform.deepseek.com). It is loaded from the first source found, in this order:

| Priority | Source |
|----------|--------|
| 1 | `./deepseek-monitor-tui.json` (working directory) |
| 2 | `~/.config/deepseek-monitor-tui/config.json` |
| 3 | `DEEPSEEK_PLATFORM_TOKEN` environment variable |

#### Getting Your Bearer Token

1. Log in to [platform.deepseek.com](https://platform.deepseek.com).
2. Open DevTools (`F12`) → **Network** tab.
3. Click any XHR request (e.g. `get_user_summary`).
4. In **Request Headers**, find the `Authorization` header.
5. Copy the value **after** `Bearer ` (do **not** include the word `Bearer`).

#### Interactive Setup (Recommended)

Just run `dpskmon` without any config — the app will guide you through pasting the token:

```
Press  t  to paste Bearer token from clipboard
```

The token is saved to `~/.config/deepseek-monitor-tui/config.json` automatically.

#### Manual Config File

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

#### Environment Variable

```sh
export DEEPSEEK_PLATFORM_TOKEN="your-platform-bearer-token-here"
dpskmon
```

#### Config Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `platform_token` | string | — | **Required.** Bearer token from platform.deepseek.com |
| `default_period` | string | `"30d"` | Default time period tab (`"30d"` or `"1d"`) |
| `query_interval_seconds` | int | `30` | Auto-refresh interval in seconds |
| `show_requests` | bool | `true` | Show total request count column |
| `show_cached` | bool | `true` | Show cached token usage column |
| `show_non_cached` | bool | `true` | Show non-cached token usage column |

<h3 id="usage">Usage</h3>

```sh
dpskmon
```

#### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Toggle between 30-day and 1-day views |
| `r` | Manually refresh data |
| `t` | Paste a new Bearer token from clipboard |
| `q` / `Ctrl+C` | Quit |

The app auto-refreshes on the configured interval (default: 30 seconds). The countdown is shown in the footer.

<h3 id="uninstall">Uninstall</h3>

#### Homebrew

```sh
brew uninstall dpskmon
brew untap marble810/tap   # optional, removes the tap
```

#### Manual

```sh
rm "$(which dpskmon)"
```

#### Remove Config

```sh
rm -f ~/.config/deepseek-monitor-tui/config.json
rmdir ~/.config/deepseek-monitor-tui   # only if empty
```

<h3 id="development">Development</h3>

#### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [gh CLI](https://cli.github.com) (for releases)

#### Common Tasks

```sh
make build            # build ./dpskmon
make run              # build and run
make fmt              # gofmt all packages
make tidy             # go mod tidy
make clean            # remove ./dpskmon binary
make install          # install to ~/.local/bin (override with INSTALL_PATH=...)
```

#### Running Tests

```sh
go test ./...
```

#### Project Structure

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

#### Releasing a New Version

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

#### Local Release (Manual)

```sh
make release-assets TAG=v1.2.3    # builds dist/ archives
make publish-release TAG=v1.2.3   # uploads to GitHub Releases via gh
bash scripts/update-homebrew-formula.sh v1.2.3 ~/path/to/homebrew-tap
```
