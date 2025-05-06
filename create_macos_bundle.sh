#!/bin/bash

# Script to create a macOS .app bundle for the 'gt' application.

# --- Configuration ---
APP_NAME="gt"
VERSION="1.0.0" # Your application's version
IDENTIFIER="com.bytecats.gt" # Updated identifier
GO_BINARY_NAME="gt" # The name of your compiled Go binary
ICON_FILE_NAME="gt_icon.icns" # Your .icns icon file

# --- Paths ---
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )" # Root of your project
BUILD_DIR="${SCRIPT_DIR}/build" # Where the .app bundle will be created
APP_BUNDLE_PATH="${BUILD_DIR}/${APP_NAME}.app"
CONTENTS_DIR="${APP_BUNDLE_PATH}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"
GO_BINARY_PATH="${SCRIPT_DIR}/${GO_BINARY_NAME}" # Assumes 'gt' binary is in project root
ICON_PATH="${SCRIPT_DIR}/${ICON_FILE_NAME}"     # Assumes icon is in project root

# --- Pre-flight Checks ---
if [ ! -f "${GO_BINARY_PATH}" ]; then
    echo "Error: Go binary not found at ${GO_BINARY_PATH}"
    echo "Please build your Go application first (e.g., go build -o ${GO_BINARY_NAME} ./cmd/gt)"
    exit 1
fi

if [ ! -f "${ICON_PATH}" ]; then
    echo "Error: Icon file not found at ${ICON_PATH}"
    echo "Please create an .icns file and place it there."
    exit 1
fi

# --- Create Bundle Structure ---
echo "Creating bundle structure at ${APP_BUNDLE_PATH}..."
rm -rf "${APP_BUNDLE_PATH}" # Remove old bundle if it exists
mkdir -p "${MACOS_DIR}"
mkdir -p "${RESOURCES_DIR}"

# --- Copy Binary ---
echo "Copying Go binary..."
cp "${GO_BINARY_PATH}" "${MACOS_DIR}/"
chmod +x "${MACOS_DIR}/${GO_BINARY_NAME}"

# --- Copy Icon ---
echo "Copying icon file..."
cp "${ICON_PATH}" "${RESOURCES_DIR}/"

# --- Create Info.plist ---
echo "Creating Info.plist..."
PLIST_CONTENT="<?xml version=\"1.0\" encoding=\"UTF-8\"?>
<!DOCTYPE plist PUBLIC \"-//Apple//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\">
<plist version=\"1.0\">
<dict>
    <key>CFBundleExecutable</key>
    <string>${GO_BINARY_NAME}</string>
    <key>CFBundleIconFile</key>
    <string>${ICON_FILE_NAME}</string> <!-- .icns is implied -->
    <key>CFBundleIdentifier</key>
    <string>${IDENTIFIER}</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.12</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key> <!-- Set to true if it's a background app without a dock icon by default -->
    <false/> <!-- Assuming gt is a foreground app that appears in Dock -->
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string> <!-- Required for foreground apps -->
    <key>NSAppleEventsUsageDescription</key>
    <string>This app does not require AppleEvents.</string>
</dict>
</plist>"

echo "${PLIST_CONTENT}" > "${CONTENTS_DIR}/Info.plist"

echo ""
echo "${APP_NAME}.app bundle created successfully at ${APP_BUNDLE_PATH}"
echo "You may need to codesign it for distribution or to run on newer macOS versions without warnings."
echo "To run: open ${APP_BUNDLE_PATH}"
echo ""

exit 0 