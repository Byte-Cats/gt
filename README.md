<div align="center";>

# gt
Go Terminal 
## Minimalist Terminal emulator written in Go

</div>


### Desired Feature Set

- Image Preview (PNG, JPEG, GIF, BMP, TIFF, WebP)
- Supports 256 colors
- Scrollable Text Buffer
- UTF-8 support

### Features

*   **PTY Integration:** Runs shell commands via a pseudo-terminal.
*   **SDL2 Rendering:** Uses SDL2 and SDL_ttf for graphical output.
*   **Color Support:** Handles 16-color, 256-color, and TrueColor (24-bit) escape sequences.
*   **Text Attributes:** Supports Bold (via font variant), Underline, and Reverse Video.
*   **Scrollback:** Basic scrollback buffer navigable via Mouse Wheel.
*   **iTerm2 Inline Image Protocol:** Displays inline images using the iTerm2 protocol (see below).
*   **Wide Character Support:** Basic handling for wide characters.

### Image Preview Support

`gt` supports displaying inline images using the **iTerm2 Inline Image Protocol**.

To display an image, an application running inside `gt` needs to output a specific escape sequence to standard output:

```
ESC ] 1337 ; File=inline=1 [;options...] : BASE64DATA ST
```

Where:
*   `ESC` is the escape character (`\x1b`).
*   `ST` is the String Terminator, which can be BEL (`\a`, `\x07`) or `ESC \` (`\x1b\\`).
*   `BASE64DATA` is the base64-encoded content of the image file.
*   `[;options...]` are optional key-value pairs separated by semicolons. Supported options include:
    *   `width=N`: Specify width in character cells (`N`), pixels (`Npx`), or percentage of terminal width (`N%`). `auto` uses native width.
    *   `height=N`: Specify height similarly (cells, `Npx`, `N%`). `auto` uses native height.
    *   `preserveAspectRatio=1|0`: Whether to maintain aspect ratio (default is 1/true).

**Example Script:**

A Python script is provided in `contrib/load_image.py` to display images:

```bash
python contrib/load_image.py /path/to/your/image.png
```

This script reads the image file, encodes it, and prints the necessary escape sequence.

### Build from source
```bash
go mod init && go mod tidy && go build
```

### Reaserch References

[C in go slideshow](http://akrennmair.github.io/golang-cgo-slides/#3)

[Rewrite vs Revive](https://medium.com/mysterium-network/golang-c-interoperability-caf0ba9f7bf3)
