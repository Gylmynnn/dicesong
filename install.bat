@echo off
setlocal

set "APP_NAME=dicesong"
set "BUILD_DIR=build"
set "BIN_DIR=%USERPROFILE%\go\bin"
set "SRC_DIR=."

if /I "%1"=="build"     goto :build
if /I "%1"=="install"   goto :install
if /I "%1"=="uninstall" goto :uninstall
if /I "%1"=="clean"     goto :clean

goto :usage

:build
echo "🔧 Building %APP_NAME%..."
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"
go build -o "%BUILD_DIR%\%APP_NAME%.exe" "%SRC_DIR%"
echo "✅ Build complete."
goto :eof

:install
call :build
echo "📦 Installing %APP_NAME% to %BIN_DIR%..."
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"
copy "%BUILD_DIR%\%APP_NAME%.exe" "%BIN_DIR%\"
echo "✅ Installed %APP_NAME% to %BIN_DIR%"
goto :eof

:uninstall
echo "🗑️ Uninstalling %APP_NAME% from %BIN_DIR%..."
if exist "%BIN_DIR%\%APP_NAME%.exe" del "%BIN_DIR%\%APP_NAME%.exe"
echo "✅ Uninstalled %APP_NAME%."
goto :eof

:clean
echo "🧹 Cleaning build files..."
if exist "%BUILD_DIR%" rmdir /S /Q "%BUILD_DIR%"
echo "✅ Clean complete."
goto :eof

:usage
echo Usage: %0 {build^|install^|uninstall^|clean}

echo   build      - Compile the application.

echo   install    - Build and install the application to %BIN_DIR%.

echo   uninstall  - Remove the application from %BIN_DIR%.

echo   clean      - Remove build artifacts.
goto :eof
