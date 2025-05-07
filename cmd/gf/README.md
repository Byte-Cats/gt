# gf - Go File Navigator

`gf` is a Terminal User Interface (TUI) file navigator written in Go using the excellent `charmbracelet/bubbletea` and `charmbracelet/lipgloss` libraries. It allows you to navigate directories, select files (outputting their absolute path), filter directory entries, and preview images directly in the terminal (if a compatible terminal image viewer like `gt` is present).

## Features

*   **Directory Navigation**: Easily move up and down the directory tree.
*   **File Selection**: Select a file to print its absolute path to standard output, useful for piping to other commands.
*   **Directory Entry Filtering**: Quickly filter the list of files and directories by typing.
*   **Image Previews**: If a selected file is an image and a terminal image viewer is detected (e.g., `gt` through iTerm2's imgcat protocol), it will be previewed directly in the terminal.
*   **Customizable Keybindings**: Modify key actions directly in the source code (`main.go`).
*   **Customizable Styles**: Change colors and a few other visual elements via a `config.toml` file.

## Keybindings

The following are the default keybindings. These can be seen in the footer of the application and are defined in `main.go`.

*   **Navigation**:
    *   `↑` / `k`: Move cursor up
    *   `↓` / `j`: Move cursor down
    *   `pgup`: Page up
    *   `pgdown`: Page down
    *   `home` / `g`: Go to top of the list
    *   `end` / `G`: Go to bottom of the list
*   **Selection / Directory Entry**:
    *   `enter` / `l`: Select the current item. If it's a directory, navigate into it. If it's a file, confirm selection (and print path/preview image).
    *   `space`: Also select the current item (same behavior as `enter`/`l`).
*   **Directory Traversal**:
    *   `←` / `h` / `backspace`: Go to the parent directory.
*   **Filtering**:
    *   `/`: Start typing to filter the current directory view.
    *   `esc` (while filtering): Clear the current filter text and stop filtering.
    *   `enter` (while filtering): Apply the current filter text and exit filtering input mode.
*   **Application**:
    *   `q` / `ctrl+c`: Quit the application.
*   **Confirmation**:
    *   `y`: Confirm file selection.
    *   `n` / `esc`: Cancel file selection.

## Configuration

`gf` can be customized by creating a `config.toml` file in the same directory as the executable.

Example `config.toml`:

```toml
# config.toml - Configuration for gf TUI

[style]
# Colors can be hex codes (#RRGGBB) or lipgloss named colors (e.g., "blue", "21", "hotpink")

# Item Styles
dir_color = "#87CEFA"      # LightSkyBlue
file_color = "#D3D3D3"     # LightGray
selected_prefix = "➜ "     # Using a different arrow
selected_color = "#EE82EE" # Violet
error_color = "#FF6347"    # Tomato Red

# Layout Styles
header_color = "#5F9EA0" # CadetBlue
footer_color = "#777777" # Dim Gray

# Border Styles (optional)
border_color = "#FF00FF" # Fuchsia / Magenta
border_type = "rounded"  # Options: single, double, rounded, thick, hidden
```

If `config.toml` is not found, or if there's an error decoding it, default styles will be used.

## Building

To build `gf`, navigate to the `cmd/gf` directory (or wherever `main.go` is located) and run:

```bash
go build
```

This will produce an executable named `gf` (or `gf.exe` on Windows) in the current directory.

## Running

After building, you can run `gf` from the directory where it was built:

```bash
./gf
```

Or, you can move the executable to a directory in your system's `PATH` (e.g., `/usr/local/bin` or `~/bin`) to run it from anywhere by simply typing `gf`.

## Dependencies

*   [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
*   [github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
*   [github.com/charmbracelet/bubbles](https://github.com/charmbracelet/bubbles)
*   [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml)

These dependencies are managed by Go modules and will be downloaded automatically when you build the project. 