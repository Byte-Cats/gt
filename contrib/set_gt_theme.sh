#!/bin/bash

# Script to easily apply a theme file to gt.
# Usage: ./set_gt_theme.sh /path/to/your_theme.toml

# Check if a theme file path is provided
if [ -z "$1" ]; then
  echo "Usage: $0 /path/to/your_theme.toml"
  exit 1
fi

SOURCE_THEME_FILE="$1"
TARGET_CONFIG_DIR="$HOME/.config/gt"
TARGET_THEME_FILE="$TARGET_CONFIG_DIR/theme.toml"

# Check if the source theme file exists
if [ ! -f "$SOURCE_THEME_FILE" ]; then
  echo "Error: Source theme file not found: $SOURCE_THEME_FILE"
  exit 1
fi

# Ensure the target directory exists
echo "Ensuring target directory exists: $TARGET_CONFIG_DIR"
mkdir -p "$TARGET_CONFIG_DIR"

# Copy the theme file
echo "Copying '$SOURCE_THEME_FILE' to '$TARGET_THEME_FILE' ..."
cp "$SOURCE_THEME_FILE" "$TARGET_THEME_FILE"

if [ $? -eq 0 ]; then
  echo "Theme applied successfully! Restart gt to see the changes."
else
  echo "Error: Failed to copy theme file."
  exit 1
fi

exit 0 