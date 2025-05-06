#!/bin/bash

# Script to prepare icon resources for a Windows Go build.

# --- Configuration ---
ICON_FILE_NAME="gt_icon.ico" # Your .ico icon file
RC_FILE_NAME="gt.rc"
SYSO_FILE_NAME="gt.syso"
MAIN_PACKAGE_DIR="cmd/gt" # Directory containing your main Go package

# --- Paths ---
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )" # Root of your project
ICON_PATH="${SCRIPT_DIR}/${MAIN_PACKAGE_DIR}/${ICON_FILE_NAME}"
RC_FILE_PATH="${SCRIPT_DIR}/${MAIN_PACKAGE_DIR}/${RC_FILE_NAME}"
SYSO_FILE_PATH="${SCRIPT_DIR}/${MAIN_PACKAGE_DIR}/${SYSO_FILE_NAME}"

# --- Pre-flight Checks ---
if ! command -v windres &> /dev/null
then
    echo "Error: windres command not found. Please install MinGW or a similar toolchain."
    exit 1
fi

if [ ! -f "${ICON_PATH}" ]; then
    echo "Error: Icon file not found at ${ICON_PATH}"
    echo "Please create an .ico file and place it there."
    exit 1
fi

# --- Create .rc file ---
echo "Creating resource file (${RC_FILE_PATH})..."
RC_CONTENT="// ${RC_FILE_NAME}
1 ICON \"${ICON_FILE_NAME}\""

echo "${RC_CONTENT}" > "${RC_FILE_PATH}"

# --- Compile .rc to .syso ---
echo "Compiling resource file to .syso (${SYSO_FILE_PATH})..."
windres -i "${RC_FILE_PATH}" -o "${SYSO_FILE_PATH}"

if [ $? -eq 0 ]; then
    echo ""
    echo "${SYSO_FILE_NAME} created successfully in ${MAIN_PACKAGE_DIR}/."
    echo "You can now build your Windows executable using Go:"
    echo "  GOOS=windows GOARCH=amd64 go build -o gt.exe ./cmd/gt"
    echo "The Go build tool will automatically include the .syso file."
    echo ""
else
    echo "Error: Failed to compile resource file with windres."
    rm -f "${RC_FILE_PATH}" # Clean up .rc file on error
    exit 1
fi

exit 0 