package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"gt/cmd/gt/config"
	"gt/cmd/gt/logger"
	"gt/cmd/gt/terminal"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Application constants
const (
	AppName    = "gt Terminal"
	AppVersion = "1.2.0"
)

var (
	log                *logger.Logger
	showVersionAndExit = flag.Bool("version", false, "Show version information and exit")
	configFile         = flag.String("config", "", "Path to configuration file")
	fullscreen         = flag.Bool("fullscreen", false, "Start in fullscreen mode")
	maximized          = flag.Bool("maximized", false, "Start with maximized window")
	fontSizeFlag       = flag.Int("font-size", 0, "Font size to use (overrides config)")
	themeFlag          = flag.String("theme", "", "Theme to use (overrides config)")
	debugFlag          = flag.Bool("debug", false, "Enable debug logging")
	executeCommand     = flag.String("e", "", "Execute command instead of shell")
	initialWidth       = flag.Int("width", 0, "Initial window width (overrides config)")
	initialHeight      = flag.Int("height", 0, "Initial window height (overrides config)")
)

func main() {
	// Parse command-line flags
	flag.Parse()

	// Set up logger
	log = logger.GetLogger()
	if *debugFlag {
		log.SetLevel(logger.DebugLevel)
	} else {
		log.SetLevel(logger.InfoLevel)
	}

	// Show version if requested
	if *showVersionAndExit {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		os.Exit(0)
	}

	log.Info("%s v%s starting...", AppName, AppVersion)
	
	// Lock main thread for SDL
	runtime.LockOSThread()
	
	// Initialize SDL
	if err := initSDL(); err != nil {
		log.Fatal("Failed to initialize SDL: %v", err)
	}
	defer cleanupSDL()

	// Load configuration
	cfg := loadConfiguration()
	
	// Apply command-line overrides
	applyCommandLineOverrides(cfg)

	// Create and initialize terminal
	term := terminal.NewTerminal(cfg, cfg.GetTheme())
	
	// Determine initial window size
	width, height := determineInitialWindowSize(cfg)
	
	// Initialize the terminal
	if err := term.Initialize(width, height); err != nil {
		log.Fatal("Failed to initialize terminal: %v", err)
	}
	defer term.Cleanup()

	// Set up signal handling for clean shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Main event loop
	mainLoop(term, signalChan)
	
	exitCode := term.GetExitCode()
	log.Info("%s terminated with exit code: %d", AppName, exitCode)
	os.Exit(exitCode)
}

// initSDL initializes SDL and TTF
func initSDL() error {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return fmt.Errorf("could not initialize SDL: %w", err)
	}
	
	if err := ttf.Init(); err != nil {
		sdl.Quit()
		return fmt.Errorf("could not initialize SDL_ttf: %w", err)
	}
	
	log.Info("SDL and TTF initialized successfully")
	return nil
}

// cleanupSDL releases SDL resources
func cleanupSDL() {
	ttf.Quit()
	sdl.Quit()
	log.Info("SDL resources released")
}

// loadConfiguration loads the terminal configuration
func loadConfiguration() *config.TerminalConfig {
	var cfg *config.TerminalConfig
	
	// Use specified config file if provided
	if *configFile != "" {
		configPath := *configFile
		// If the path is not absolute, make it relative to the executable
		if !filepath.IsAbs(configPath) {
			execPath, err := os.Executable()
			if err == nil {
				execDir := filepath.Dir(execPath)
				configPath = filepath.Join(execDir, configPath)
			}
		}
		
		var err error
		cfg, err = config.LoadFromFile(configPath)
		if err != nil {
			log.Warn("Failed to load config from %s: %v", configPath, err)
			cfg = config.GetConfig() // Fall back to default config
		} else {
			log.Info("Loaded configuration from %s", configPath)
		}
	} else {
		// Load default configuration
		cfg = config.GetConfig()
	}
	
	return cfg
}

// applyCommandLineOverrides applies command-line flag overrides to the config
func applyCommandLineOverrides(cfg *config.TerminalConfig) {
	// Apply fullscreen if specified
	if *fullscreen {
		cfg.StartFullscreen = true
	}
	
	// Apply maximized if specified
	if *maximized {
		cfg.StartMaximized = true
	}
	
	// Apply font size if specified
	if *fontSizeFlag > 0 {
		theme := cfg.GetTheme()
		theme.FontSize = *fontSizeFlag
		cfg.SetTheme(theme)
	}
	
	// Apply theme if specified
	if *themeFlag != "" {
		if err := cfg.LoadTheme(*themeFlag); err != nil {
			log.Warn("Failed to load theme %s: %v", *themeFlag, err)
		} else {
			log.Info("Applied theme: %s", *themeFlag)
		}
	}
	
	// Apply initial width/height if specified
	if *initialWidth > 0 {
		cfg.LastWidth = *initialWidth
		cfg.RememberSize = true
	}
	
	if *initialHeight > 0 {
		cfg.LastHeight = *initialHeight
		cfg.RememberSize = true
	}
	
	// Apply execute command if specified
	if *executeCommand != "" {
		cfg.ShellCommand = *executeCommand
	}
}

// determineInitialWindowSize calculates the initial window size
func determineInitialWindowSize(cfg *config.TerminalConfig) (int, int) {
	// Default size if nothing is specified
	defaultWidth := 900
	defaultHeight := 600
	
	if cfg.RememberSize && cfg.LastWidth > 0 && cfg.LastHeight > 0 {
		return cfg.LastWidth, cfg.LastHeight
	}
	
	// Check if command-line overrides were provided
	if *initialWidth > 0 && *initialHeight > 0 {
		return *initialWidth, *initialHeight
	}
	
	return defaultWidth, defaultHeight
}

// mainLoop handles the main application event loop
func mainLoop(term *terminal.Terminal, signalChan chan os.Signal) {
	// Frame rate limiter
	frameDelay := sdl.GetPerformanceFrequency() / 60 // Target 60 FPS
	lastFrame := sdl.GetPerformanceCounter()
	
	for term.IsRunning() {
		select {
		case sig := <-signalChan:
			log.Info("Received signal: %v, initiating shutdown...", sig)
			return
			
		default:
			// Process events (returns false when quitting)
			if !term.ProcessEvents() {
				return
			}
			
			// Render the terminal
			term.Render()
			
			// Frame rate limiting
			now := sdl.GetPerformanceCounter()
			elapsed := now - lastFrame
			if elapsed < frameDelay {
				// Sleep to maintain target frame rate
				sdl.Delay(uint32((frameDelay - elapsed) * 1000 / sdl.GetPerformanceFrequency()))
			}
			lastFrame = sdl.GetPerformanceCounter()
		}
	}
}