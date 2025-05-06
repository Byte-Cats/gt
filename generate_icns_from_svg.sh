#!/bin/bash

# Script to generate a .icns file from an SVG using rsvg-convert and iconutil.

# --- Configuration ---
SVG_SOURCE_FILE="gt_icon.svg"
ICNS_TARGET_FILE="gt_icon.icns"
ICONSET_NAME="gt.iconset"

# --- Paths ---
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )" # Root of your project
SVG_PATH="${SCRIPT_DIR}/${SVG_SOURCE_FILE}"
ICONSET_PATH="${SCRIPT_DIR}/${ICONSET_NAME}"
ICNS_PATH="${SCRIPT_DIR}/${ICNS_TARGET_FILE}"

# SVG rendering command (rsvg-convert is preferred if available)
# If rsvg-convert is not found, this script will exit with a message.
RSVG_CONVERT_CMD="rsvg-convert"

# --- Pre-flight Checks ---
if [ ! -f "${SVG_PATH}" ]; then
    echo "Error: SVG source file not found at ${SVG_PATH}"
    exit 1
fi

if ! command -v iconutil &> /dev/null; then
    echo "Error: iconutil command not found. This script is for macOS."
    exit 1
fi

if ! command -v ${RSVG_CONVERT_CMD} &> /dev/null; then
    echo "Error: ${RSVG_CONVERT_CMD} command not found."
    echo "Please install librsvg (e.g., 'brew install librsvg' on macOS) or ensure rsvg-convert is in your PATH."
    echo "Alternatively, you can modify this script to use another SVG to PNG converter like Inkscape CLI."
    exit 1
fi

# --- Create or clean iconset directory ---
echo "Preparing iconset directory: ${ICONSET_PATH}"
rm -rf "${ICONSET_PATH}"
mkdir -p "${ICONSET_PATH}"

# --- Define PNG sizes required for iconset ---
# Format: <base_size_for_filename> <actual_pixel_width> <actual_pixel_height>
# Example: icon_16x16.png will be 16x16, icon_16x16@2x.png will be 32x32
declare -a SIZES=(
    "16x16 16 16 icon_16x16.png"
    "16x16@2x 32 32 icon_16x16@2x.png"
    "32x32 32 32 icon_32x32.png"
    "32x32@2x 64 64 icon_32x32@2x.png"
    "128x128 128 128 icon_128x128.png"
    "128x128@2x 256 256 icon_128x128@2x.png"
    "256x256 256 256 icon_256x256.png"
    "256x256@2x 512 512 icon_256x256@2x.png"
    "512x512 512 512 icon_512x512.png"
    "512x512@2x 1024 1024 icon_512x512@2x.png" # Largest size for high-res displays
)

# --- Generate PNGs from SVG ---
echo "Generating PNGs from ${SVG_SOURCE_FILE}..."
for size_info in "${SIZES[@]}"; do
    read -r name width height filename <<<"${size_info}" # Bash specific way to split string
    echo "  Creating ${filename} (${width}x${height})"
    ${RSVG_CONVERT_CMD} -w "${width}" -h "${height}" "${SVG_PATH}" -o "${ICONSET_PATH}/${filename}"
    if [ $? -ne 0 ]; then
        echo "Error: Failed to generate ${filename}"
        echo "Cleaning up and exiting."
        rm -rf "${ICONSET_PATH}"
        exit 1
    fi
done

# --- Create .icns file using iconutil ---
echo "Creating ${ICNS_TARGET_FILE} using iconutil..."
iconutil -c icns "${ICONSET_PATH}" -o "${ICNS_PATH}"

if [ $? -eq 0 ]; then
    echo ""
    echo "${ICNS_TARGET_FILE} created successfully at ${ICNS_PATH}"
    echo "You can now use this .icns file with the create_macos_bundle.sh script."
else
    echo "Error: iconutil failed to create .icns file."
    echo "Check the output above for any errors from iconutil."
    # Keep iconset for debugging if iconutil fails
    exit 1
fi

# --- Clean up iconset directory ---
echo "Cleaning up temporary iconset directory: ${ICONSET_PATH}"
rm -rf "${ICONSET_PATH}"

echo "Done."
exit 0 