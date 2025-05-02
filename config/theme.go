package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Theme holds the color configuration for the terminal.
type Theme struct {
	FontPath string      `toml:"font_path,omitempty"`
	FontSize int         `toml:"font_size,omitempty"`
	Colors   ThemeColors `toml:"colors"`
	Gradient Gradient    `toml:"gradient,omitempty"`
	// TODO: Add gradient settings later
}

// ThemeColors defines the specific color values.
type ThemeColors struct {
	Foreground string `toml:"foreground"`
	Background string `toml:"background"`
	Cursor     string `toml:"cursor"`

	// Standard 16 ANSI colors
	Black         string `toml:"black"`
	Red           string `toml:"red"`
	Green         string `toml:"green"`
	Yellow        string `toml:"yellow"`
	Blue          string `toml:"blue"`
	Magenta       string `toml:"magenta"`
	Cyan          string `toml:"cyan"`
	White         string `toml:"white"`
	BrightBlack   string `toml:"bright_black"`
	BrightRed     string `toml:"bright_red"`
	BrightGreen   string `toml:"bright_green"`
	BrightYellow  string `toml:"bright_yellow"`
	BrightBlue    string `toml:"bright_blue"`
	BrightMagenta string `toml:"bright_magenta"`
	BrightCyan    string `toml:"bright_cyan"`
	BrightWhite   string `toml:"bright_white"`
}

// Gradient defines the background gradient settings.
type Gradient struct {
	Enabled    bool   `toml:"enabled"`
	StartColor string `toml:"start_color"`
	EndColor   string `toml:"end_color"`
	Direction  string `toml:"direction"` // "vertical" or "horizontal"
}

// DefaultTheme provides sensible default colors.
func DefaultTheme() Theme {
	return Theme{
		FontPath: "/System/Library/Fonts/Supplemental/SFMono-Regular.otf",
		FontSize: 14,
		Colors: ThemeColors{
			Foreground:    "#cccccc", // Light gray
			Background:    "#1e1e1e", // Dark gray
			Cursor:        "#ffffff", // White
			Black:         "#000000",
			Red:           "#cd3131",
			Green:         "#0dbc79",
			Yellow:        "#e5e510",
			Blue:          "#2472c8",
			Magenta:       "#bc3fbc",
			Cyan:          "#11a8cd",
			White:         "#e5e5e5",
			BrightBlack:   "#666666",
			BrightRed:     "#f14c4c",
			BrightGreen:   "#23d18b",
			BrightYellow:  "#f5f543",
			BrightBlue:    "#3b8eea",
			BrightMagenta: "#d670d6",
			BrightCyan:    "#29b8db",
			BrightWhite:   "#ffffff",
		},
		Gradient: Gradient{
			Enabled:    false,
			StartColor: "#303030",
			EndColor:   "#101010",
			Direction:  "vertical",
		},
	}
}

// LoadTheme attempts to load a theme from ~/.config/gt/theme.toml
// and falls back to the DefaultTheme if not found or invalid.
func LoadTheme() Theme {
	theme := DefaultTheme() // Start with defaults

	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Warning: Could not get user home directory: %v. Using default theme.", err)
		return theme
	}

	configPath := filepath.Join(home, ".config", "gt", "theme.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Info: No theme file found at %s. Using default theme.", configPath)
		// Optionally, we could create a default theme file here
		return theme
	}

	// Keep existing theme colors if TOML only specifies font
	loadedTheme := DefaultTheme() // Use defaults as base
	if _, err := toml.DecodeFile(configPath, &loadedTheme); err != nil {
		log.Printf("Warning: Failed to decode theme file %s: %v. Using default theme.", configPath, err)
		return DefaultTheme() // Return defaults again on error
	}

	// Override font path/size only if they were actually present in the TOML
	if loadedTheme.FontPath != "" {
		theme.FontPath = loadedTheme.FontPath
	}
	if loadedTheme.FontSize > 0 {
		theme.FontSize = loadedTheme.FontSize
	}
	// Colors and Gradient are directly overridden by DecodeFile
	theme.Colors = loadedTheme.Colors
	theme.Gradient = loadedTheme.Gradient // Load gradient settings

	log.Printf("Loaded theme from %s (Font: %s, Size: %d)", configPath, theme.FontPath, theme.FontSize)
	// TODO: Validate loaded color strings?
	return theme
}
