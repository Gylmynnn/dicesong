APP_NAME = dicesong
BUILD_DIR = build
BIN_DIR = $(HOME)/.local/bin
SRC_DIR = .

.PHONY: all build install clean uninstall

all: build

build:
	@echo "🔧 Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(SRC_DIR)

install: build
	@echo "📦 Installing $(APP_NAME) to $(BIN_DIR)..."
	@mkdir -p $(BIN_DIR)
	@cp $(BUILD_DIR)/$(APP_NAME) $(BIN_DIR)/
	@echo "✅ Installed $(APP_NAME) to $(BIN_DIR)"

uninstall:
	@echo "🗑️ Uninstalling $(APP_NAME) from $(BIN_DIR)..."
	@rm -f $(BIN_DIR)/$(APP_NAME)
	@echo "✅ Uninstalled $(APP_NAME)"

clean:
	@echo "🧹 Cleaning build files..."
	@rm -rf $(BUILD_DIR)
