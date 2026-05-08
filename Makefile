.PHONY: build run install clean config tidy fmt release-assets publish-release

# Binary name
BINARY=dpskmon

# Install path for cmux Dock
INSTALL_PATH?=$(HOME)/.local/bin
OUTPUT_DIR?=dist
TAG?=

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

install: build
	mkdir -p $(INSTALL_PATH)
	cp $(BINARY) $(INSTALL_PATH)/$(BINARY)
	@echo "Installed to $(INSTALL_PATH)/$(BINARY)"
	@echo "Make sure $(INSTALL_PATH) is in your PATH, or update .cmux/dock.json command path."

config:
	@if [ ! -f deepseek-monitor-tui.json ]; then \
		echo '{"platform_token": "your-platform-bearer-token-here"}' > deepseek-monitor-tui.json; \
		echo "Created deepseek-monitor-tui.json — edit it with your real Bearer token."; \
	else \
		echo "deepseek-monitor-tui.json already exists."; \
	fi

clean:
	rm -f $(BINARY)

tidy:
	go mod tidy

fmt:
	go fmt ./...

release-assets:
	@test -n "$(TAG)" || (echo "TAG is required, e.g. make release-assets TAG=v1.2.3" && exit 1)
	bash scripts/build-release-assets.sh "$(TAG)" "$(OUTPUT_DIR)"

publish-release:
	@test -n "$(TAG)" || (echo "TAG is required, e.g. make publish-release TAG=v1.2.3" && exit 1)
	bash scripts/publish-release.sh "$(TAG)" "$(OUTPUT_DIR)"
