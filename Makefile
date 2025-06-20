# Makefile for SpoofDPI
# macOS launchd service management with dynamic user configuration

# Variables
BINARY_NAME = spoofdpi
GO_MODULE = github.com/bariiss/SpoofDPI
CMD_PATH = ./cmd/spoofdpi
BUILD_DIR = build
INSTALL_DIR = $(HOME)/go/bin

# Dynamic user information
USERNAME := $(shell whoami)
USER_ID := $(shell id -u)
USER_HOME := $(HOME)
SERVICE_NAME = com.$(USERNAME).spoofdpi
PLIST_FILE = $(SERVICE_NAME).plist
LAUNCH_AGENTS_DIR = $(USER_HOME)/Library/LaunchAgents
PLIST_PATH = $(LAUNCH_AGENTS_DIR)/$(PLIST_FILE)

# Default configuration for plist
DEFAULT_PORT = 8080
DEFAULT_DNS = 1.1.1.1
DEFAULT_ADDR = 127.0.0.1
DEFAULT_WINDOW_SIZE = 1
DEFAULT_ENABLE_DOH = false
DEFAULT_SYSTEM_PROXY = false

# Colors for output
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[0;33m
BLUE = \033[0;34m
NC = \033[0m

.PHONY: help build install clean service-install service-start service-stop service-restart service-uninstall service-status service-logs service-reload service-config all

# Default target
all: build install service-install service-start

help: ## Show this help message
	@echo "$(BLUE)SpoofDPI Makefile Commands:$(NC)"
	@echo ""
	@echo "$(YELLOW)Build Commands:$(NC)"
	@echo "  build                Build the binary"
	@echo "  install              Install binary to $(INSTALL_DIR)"
	@echo "  clean                Clean build artifacts"
	@echo ""
	@echo "$(YELLOW)Service Management Commands:$(NC)"
	@echo "  service-install      Create and install launchd service"
	@echo "  service-start        Start the service"
	@echo "  service-stop         Stop the service"
	@echo "  service-restart      Restart the service"
	@echo "  service-uninstall    Uninstall and remove the service"
	@echo "  service-status       Show service status"
	@echo "  service-logs         Show service logs"
	@echo "  service-reload       Reload service configuration"
	@echo ""
	@echo "$(YELLOW)Configuration Commands:$(NC)"
	@echo "  show-config          Show current service configuration"
	@echo "  service-config       Configure service with custom parameters"
	@echo ""
	@echo "$(YELLOW)Combined Commands:$(NC)"
	@echo "  all                  Build, install, and start service"
	@echo ""
	@echo "$(BLUE)Current Configuration:$(NC)"
	@echo "  Username: $(USERNAME)"
	@echo "  Service Name: $(SERVICE_NAME)"
	@echo "  Plist Path: $(PLIST_PATH)"
	@echo "  Binary Path: $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "$(BLUE)Configuration Parameters (can be overridden):$(NC)"
	@echo "  PORT=$(DEFAULT_PORT)             - Proxy port"
	@echo "  DNS=$(DEFAULT_DNS)   - DNS server address"
	@echo "  ADDR=$(DEFAULT_ADDR)             - Bind address"
	@echo "  WINDOW_SIZE=$(DEFAULT_WINDOW_SIZE)             - Window size"
	@echo "  ENABLE_DOH=$(DEFAULT_ENABLE_DOH)           - Enable DNS over HTTPS"
	@echo "  SYSTEM_PROXY=$(DEFAULT_SYSTEM_PROXY)         - Enable system proxy"
	@echo ""
	@echo "$(YELLOW)Example:$(NC) make service-config PORT=8080 ENABLE_DOH=false SYSTEM_PROXY=true"

build: ## Build the binary
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags '-w -s' -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

install: build ## Install binary to system
	@echo "$(BLUE)Installing $(BINARY_NAME) to $(INSTALL_DIR)...$(NC)"
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "$(GREEN)Installation completed: $(INSTALL_DIR)/$(BINARY_NAME)$(NC)"

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)Clean completed$(NC)"

service-install: install ## Create and install launchd service
	@echo "$(BLUE)Creating launchd service configuration...$(NC)"
	@mkdir -p $(LAUNCH_AGENTS_DIR)
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > $(PLIST_PATH)
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"' >> $(PLIST_PATH)
	@echo '  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> $(PLIST_PATH)
	@echo '<plist version="1.0">' >> $(PLIST_PATH)
	@echo '<dict>' >> $(PLIST_PATH)
	@echo '  <key>Label</key>' >> $(PLIST_PATH)
	@echo '  <string>$(SERVICE_NAME)</string>' >> $(PLIST_PATH)
	@echo '' >> $(PLIST_PATH)
	@echo '  <key>ProgramArguments</key>' >> $(PLIST_PATH)
	@echo '  <array>' >> $(PLIST_PATH)
	@echo '    <string>$(INSTALL_DIR)/$(BINARY_NAME)</string>' >> $(PLIST_PATH)
	@echo '    <string>-addr=$(DEFAULT_ADDR)</string>' >> $(PLIST_PATH)
	@echo '    <string>-dns-addr=$(DEFAULT_DNS)</string>' >> $(PLIST_PATH)
	@echo '    <string>-enable-doh=$(DEFAULT_ENABLE_DOH)</string>' >> $(PLIST_PATH)
	@echo '    <string>-window-size=$(DEFAULT_WINDOW_SIZE)</string>' >> $(PLIST_PATH)
	@echo '    <string>-port=$(DEFAULT_PORT)</string>' >> $(PLIST_PATH)
	@echo '    <string>-system-proxy=$(DEFAULT_SYSTEM_PROXY)</string>' >> $(PLIST_PATH)
	@echo '  </array>' >> $(PLIST_PATH)
	@echo '' >> $(PLIST_PATH)
	@echo '  <key>RunAtLoad</key>' >> $(PLIST_PATH)
	@echo '  <true/>' >> $(PLIST_PATH)
	@echo '  <key>KeepAlive</key>' >> $(PLIST_PATH)
	@echo '  <true/>' >> $(PLIST_PATH)
	@echo '' >> $(PLIST_PATH)
	@echo '  <key>StandardOutPath</key>' >> $(PLIST_PATH)
	@echo '  <string>/tmp/$(BINARY_NAME).log</string>' >> $(PLIST_PATH)
	@echo '  <key>StandardErrorPath</key>' >> $(PLIST_PATH)
	@echo '  <string>/tmp/$(BINARY_NAME).err</string>' >> $(PLIST_PATH)
	@echo '</dict>' >> $(PLIST_PATH)
	@echo '</plist>' >> $(PLIST_PATH)
	@echo "$(GREEN)Service configuration created: $(PLIST_PATH)$(NC)"

service-start: ## Start the service
	@echo "$(BLUE)Starting $(SERVICE_NAME) service...$(NC)"
	@if [ -f "$(PLIST_PATH)" ]; then \
		launchctl bootstrap gui/$(USER_ID) $(PLIST_PATH) 2>/dev/null || true; \
		echo "$(GREEN)Service started$(NC)"; \
	else \
		echo "$(RED)Service not installed. Run 'make service-install' first.$(NC)"; \
		exit 1; \
	fi

service-stop: ## Stop the service
	@echo "$(BLUE)Stopping $(SERVICE_NAME) service...$(NC)"
	@launchctl bootout gui/$(USER_ID) $(PLIST_PATH) 2>/dev/null || true
	@echo "$(GREEN)Service stopped$(NC)"

service-restart: service-stop service-start ## Restart the service
	@echo "$(GREEN)Service restarted$(NC)"

service-uninstall: service-stop ## Uninstall and remove the service
	@echo "$(BLUE)Uninstalling $(SERVICE_NAME) service...$(NC)"
	@launchctl remove $(SERVICE_NAME) 2>/dev/null || true
	@if [ -f "$(PLIST_PATH)" ]; then \
		rm -f $(PLIST_PATH); \
		echo "$(GREEN)Service configuration removed: $(PLIST_PATH)$(NC)"; \
	fi
	@echo "$(GREEN)Service uninstalled$(NC)"

service-status: ## Show service status
	@echo "$(BLUE)Service Status for $(SERVICE_NAME):$(NC)"
	@if launchctl print gui/$(USER_ID)/$(SERVICE_NAME) >/dev/null 2>&1; then \
		echo "$(GREEN)Service is running$(NC)"; \
		launchctl print gui/$(USER_ID)/$(SERVICE_NAME) | grep -E "(state|pid)" || true; \
	else \
		echo "$(YELLOW)Service is not running$(NC)"; \
	fi
	@echo ""
	@if [ -f "$(PLIST_PATH)" ]; then \
		echo "$(BLUE)Configuration file exists: $(PLIST_PATH)$(NC)"; \
	else \
		echo "$(RED)Configuration file not found: $(PLIST_PATH)$(NC)"; \
	fi

service-logs: ## Show service logs
	@echo "$(BLUE)Showing logs for $(SERVICE_NAME):$(NC)"
	@echo "$(YELLOW)Standard Output:$(NC)"
	@if [ -f "/tmp/$(BINARY_NAME).log" ]; then \
		tail -n 50 /tmp/$(BINARY_NAME).log; \
	else \
		echo "No standard output log found"; \
	fi
	@echo ""
	@echo "$(YELLOW)Standard Error:$(NC)"
	@if [ -f "/tmp/$(BINARY_NAME).err" ]; then \
		tail -n 50 /tmp/$(BINARY_NAME).err; \
	else \
		echo "No error log found"; \
	fi

service-reload: ## Reload service configuration
	@echo "$(BLUE)Reloading $(SERVICE_NAME) service configuration...$(NC)"
	@$(MAKE) service-stop
	@$(MAKE) service-start
	@echo "$(GREEN)Service configuration reloaded$(NC)"

service-config: ## Configure service with custom parameters (e.g., make service-config PORT=8080 ENABLE_DOH=false)
	@echo "$(BLUE)Configuring service with custom parameters...$(NC)"
	@$(MAKE) service-install \
		DEFAULT_PORT=$(or $(PORT),$(DEFAULT_PORT)) \
		DEFAULT_DNS=$(or $(DNS),$(DEFAULT_DNS)) \
		DEFAULT_ADDR=$(or $(ADDR),$(DEFAULT_ADDR)) \
		DEFAULT_WINDOW_SIZE=$(or $(WINDOW_SIZE),$(DEFAULT_WINDOW_SIZE)) \
		DEFAULT_ENABLE_DOH=$(or $(ENABLE_DOH),$(DEFAULT_ENABLE_DOH)) \
		DEFAULT_SYSTEM_PROXY=$(or $(SYSTEM_PROXY),$(DEFAULT_SYSTEM_PROXY))
	@echo "$(GREEN)Service configured with custom parameters$(NC)"
	@echo "$(YELLOW)Use 'make service-restart' to apply changes$(NC)"

# Development targets
dev-run: build ## Run the binary directly (for development)
	@echo "$(BLUE)Running $(BINARY_NAME) in development mode...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME)

dev-test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -v ./...

# Docker targets
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build -t spoofdpi:latest .
	@echo "$(GREEN)Docker image built: spoofdpi:latest$(NC)"

docker-run: docker-build ## Run Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	@docker run -p 8371:8371 spoofdpi:latest

# Show current configuration
show-config: ## Show current service configuration
	@echo "$(BLUE)Current Service Configuration:$(NC)"
	@echo "Username: $(USERNAME)"
	@echo "User ID: $(USER_ID)"
	@echo "Home Directory: $(USER_HOME)"
	@echo "Service Name: $(SERVICE_NAME)"
	@echo "Plist File: $(PLIST_FILE)"
	@echo "Launch Agents Directory: $(LAUNCH_AGENTS_DIR)"
	@echo "Plist Path: $(PLIST_PATH)"
	@echo "Binary Install Path: $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "Default Port: $(DEFAULT_PORT)"
	@echo "Default DNS: $(DEFAULT_DNS)"
	@echo "Default Address: $(DEFAULT_ADDR)"
	@echo "Default Enable DoH: $(DEFAULT_ENABLE_DOH)"
	@echo "Default System Proxy: $(DEFAULT_SYSTEM_PROXY)"
