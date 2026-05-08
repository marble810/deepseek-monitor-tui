.PHONY: build run install clean config

# Binary name
BINARY=deepseekMon

# Install path for cmux Dock
INSTALL_PATH?=$(HOME)/.local/bin

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
	@if [ ! -f deepseekMon.json ]; then \
		echo '{"platform_token": "your-platform-bearer-token-here"}' > deepseekMon.json; \
		echo "Created deepseekMon.json — edit it with your real Bearer token."; \
	else \
		echo "deepseekMon.json already exists."; \
	fi

clean:
	rm -f $(BINARY)

tidy:
	go mod tidy

fmt:
	go fmt ./...
