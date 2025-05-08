# GT Terminal Image API Reference

## Overview

GT Terminal supports rich image rendering capabilities via ANSI escape sequences compatible with the iTerm2 inline image protocol. This document explains how to leverage these capabilities in CLI applications.

## Basic Usage

To display an image in GT Terminal, send a special escape sequence to stdout:

```sh
echo -e "\033]1337;File=inline=1:$(base64 < image.png)\a"
```

## Enhanced Image Protocol

GT Terminal extends the iTerm2 protocol with additional options for more control over image rendering:

### Basic Options

| Option | Description | Example | Default |
|--------|-------------|---------|---------|
| `inline` | Display image at current cursor position | `inline=1` | Required |
| `width` | Image width | `width=10` (cells), `width=100px`, `width=50%` | `auto` |
| `height` | Image height | `height=5` (cells), `height=80px`, `height=30%` | `auto` |
| `preserveaspectratio` | Maintain aspect ratio | `preserveaspectratio=0` | `1` (true) |

### Extended Options

| Option | Description | Example | Default |
|--------|-------------|---------|---------|
| `max-width` | Maximum width in pixels | `max-width=800` | `0` (unlimited) |
| `max-height` | Maximum height in pixels | `max-height=600` | `0` (unlimited) |
| `z-index` | Layering order (higher = in front) | `z-index=10` | `0` |
| `align` | Horizontal alignment | `align=left`, `align=center`, `align=right` | `left` |
| `name` | Named reference for later manipulation | `name=logo` | (none) |
| `persistent` | Keep image when cursor moves | `persistent=1` | `0` (false) |

## Example Usage

### Display an image with constraints

```sh
# Display at half terminal width, 300px height maximum, centered
echo -e "\033]1337;File=inline=1;width=50%;max-height=300;align=center:$(base64 < image.png)\a"
```

### Display overlapping images with z-index

```sh
# Background image
echo -e "\033]1337;File=inline=1;width=100%;height=100%;z-index=1;name=bg:$(base64 < background.jpg)\a"

# Foreground image (will render on top)  
echo -e "\033]1337;File=inline=1;width=20%;align=center;z-index=2;name=logo:$(base64 < logo.png)\a"
```

## Dimensions and Scaling

- **Cell-based**: `width=10` means 10 character cells wide
- **Pixel-based**: `width=100px` specifies exact pixel dimensions
- **Percentage**: `width=50%` means 50% of terminal width
- **Auto**: `width=auto` uses image's native dimensions

## Multi-Image Management

When displaying multiple images, use these strategies:

1. Use `name` to give each image a unique identifier
2. Use `z-index` to control layering order
3. Use `persistent=1` to prevent images from disappearing when cursor moves

## Working with Images in Python

```python
import base64
import sys

def display_image(path, **options):
    with open(path, 'rb') as f:
        image_data = base64.b64encode(f.read()).decode('ascii')
    
    # Build options string
    opts = ['inline=1']
    for key, value in options.items():
        opts.append(f"{key}={value}")
    
    # Construct the escape sequence  
    seq = f"\033]1337;File={';'.join(opts)}:{image_data}\a"
    sys.stdout.write(seq)
    sys.stdout.flush()

# Examples
display_image('photo.jpg', width='50%', align='center')
display_image('icon.png', width='32px', height='32px', name='status_icon', persistent='1')
```

## Working with Images in Go

```go
package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func displayImage(path string, options map[string]string) error {
	// Read image file
	imgData, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	// Encode as base64
	b64Data := base64.StdEncoding.EncodeToString(imgData)
	
	// Build options string
	opts := []string{"inline=1"}
	for k, v := range options {
		opts = append(opts, fmt.Sprintf("%s=%s", k, v))
	}
	
	// Output the escape sequence
	fmt.Printf("\033]1337;File=%s:%s\a", strings.Join(opts, ";"), b64Data)
	return nil
}

func main() {
	displayImage("image.png", map[string]string{
		"width":        "40%",
		"max-height":   "300",
		"align":        "center",
		"name":         "main_image",
		"z-index":      "5",
		"persistent":   "1",
	})
}
```

## Performance Considerations

- Large images can consume significant memory
- Consider using `max-width` and `max-height` to limit resource usage
- Use image compression appropriately before sending to the terminal
- The terminal will cache images with the same content

## Compatibility

This extended protocol is backward compatible with terminals that support the iTerm2 image protocol, but extended features will only work in GT Terminal.