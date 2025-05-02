package main

import (
	"gt/buffer"
	"gt/config" // Import config package
	"gt/render" // Will be refactored later
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	// "golang.org/x/term" // No longer needed
)

/* // Constants moved to theme config
const (
	defaultWidth  = 800
	defaultHeight = 600
	fontPath      = "/System/Library/Fonts/Menlo.ttc" // Using Menlo found on macOS
	fontSize      = 14
)
*/

func main() {
	runtime.LockOSThread() // SDL requires the main loop on the main thread

	// --- Load Theme ---
	theme := config.LoadTheme()

	// --- Initialize SDL ---
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		log.Fatalf("Failed to initialize SDL: %v", err)
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		log.Fatalf("Failed to initialize SDL_ttf: %v", err)
	}
	defer ttf.Quit()

	// --- Load Fonts ---
	log.Printf("Attempting to load font: %s (Size: %d)", theme.FontPath, theme.FontSize)
	font, err := ttf.OpenFont(theme.FontPath, theme.FontSize)
	if err != nil {
		log.Printf("Warning: Failed to open configured font %s: %v", theme.FontPath, err)
		// Fallback 1: Try default Menlo
		defaultFontPath := "/System/Library/Fonts/Menlo.ttc"
		log.Printf("Attempting fallback font: %s", defaultFontPath)
		font, err = ttf.OpenFont(defaultFontPath, theme.FontSize)
		if err != nil {
			// Fallback 2: Give up?
			log.Fatalf("Failed to open configured font and fallback font %s: %v", defaultFontPath, err)
		}
		theme.FontPath = defaultFontPath // Update theme if we used fallback
	}
	defer font.Close()

	// Attempt to load bold font variant from the same file as the primary font
	var boldFont *ttf.Font
	boldFont, err = ttf.OpenFontIndex(theme.FontPath, theme.FontSize, 1) // Use theme.FontPath
	if err != nil {
		log.Printf("Warning: Could not load bold font variant from %s (index 1): %v", theme.FontPath, err)
		boldFont = nil // Proceed without bold variant
	} else {
		log.Printf("Loaded bold font variant from %s (index 1)", theme.FontPath)
		defer boldFont.Close()
	}

	// Calculate cell size based on font
	glyphWidth, glyphHeight, err := font.SizeUTF8("W") // Use a wide char for estimate
	if err != nil {
		log.Fatalf("Failed to get font glyph size: %v", err)
	}

	// Initial window size based on cols/rows (or defaults)
	// TODO: Allow configuration of initial cols/rows
	initialCols := 80
	initialRows := 24
	windowWidth := glyphWidth * initialCols
	windowHeight := glyphHeight * initialRows

	window, err := sdl.CreateWindow("gt Terminal", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(windowWidth), int32(windowHeight), sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}
	defer window.Destroy()

	rendererSDL, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}
	defer rendererSDL.Destroy()

	// --- PTY Setup ---
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}
	c := exec.Command(shell)

	// Start PTY with initial calculated size
	winSize := &pty.Winsize{Rows: uint16(initialRows), Cols: uint16(initialCols)}
	ptmx, err := pty.StartWithSize(c, winSize)
	if err != nil {
		log.Fatalf("Failed to start pty: %v", err)
	}
	defer func() { _ = ptmx.Close() }()

	// --- Buffer & Renderer Setup ---
	outBuffer := buffer.NewOutputBuffer(initialRows, initialCols)
	// Pass both fonts to the renderer
	// termRenderer := render.NewSDLRenderer(rendererSDL, font, boldFont)
	// defer termRenderer.Destroy() // Defer cleanup of the glyph cache

	// Pass both fonts and theme to the renderer
	termRenderer := render.NewSDLRenderer(rendererSDL, font, boldFont, theme)
	defer termRenderer.Destroy()

	// --- Input/Output Goroutines ---
	// PTY Output -> Buffer
	ptyDone := make(chan struct{})
	go func() {
		defer close(ptyDone)
		// No longer trigger render here, just update buffer
		// Need a way to signal main loop to redraw? Maybe via SDL event?
		if _, err := io.Copy(outBuffer, ptmx); err != nil {
			// Handle expected EOF etc.
			if err != io.EOF && !strings.Contains(err.Error(), "file already closed") {
				log.Printf("Error copying PTY output to buffer: %v", err)
			}
		}
		log.Println("PTY read loop finished.")
	}()

	// Stdin equivalent (SDL Keyboard) -> PTY handled in main loop

	// --- Signal Handling (for OS signals, not window resize) ---
	// SIGWINCH is handled by SDL events now
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// --- Main Event Loop ---
	running := true
	needsRedraw := true // Initial draw

	// Ticker for minimum redraw rate (e.g., 60fps) even if no events
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for running {
		select {
		case <-ticker.C:
			// Minimum redraw trigger
			needsRedraw = true
		case <-sigChan:
			log.Println("Received OS signal, exiting.")
			running = false
		case <-ptyDone:
			log.Println("PTY closed, exiting.")
			running = false
		default:
			// Non-blocking check for SDL events
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch ev := event.(type) {
				case *sdl.QuitEvent:
					running = false
				case *sdl.WindowEvent:
					if ev.Event == sdl.WINDOWEVENT_RESIZED {
						newWidth := int(ev.Data1)
						newHeight := int(ev.Data2)
						newCols := newWidth / glyphWidth
						newRows := newHeight / glyphHeight
						if newCols > 0 && newRows > 0 {
							winSize := &pty.Winsize{Rows: uint16(newRows), Cols: uint16(newCols)}
							if err := pty.Setsize(ptmx, winSize); err != nil {
								log.Printf("error resizing pty: %s", err)
							}
							outBuffer.Resize(newRows, newCols)
							needsRedraw = true
						}
					}
				case *sdl.TextInputEvent:
					// Handle text input (regular characters)
					inputBytes := []byte(ev.GetText())
					if _, err := ptmx.Write(inputBytes); err != nil {
						log.Printf("Error writing text input to pty: %v", err)
					}
				case *sdl.KeyboardEvent:
					// Handle key presses (especially special keys, modifiers)
					if ev.State == sdl.PRESSED {
						var seq []byte
						mod := ev.Keysym.Mod
						sym := ev.Keysym.Sym

						// Check for Ctrl combinations first
						if mod&sdl.KMOD_CTRL != 0 {
							if sym >= sdl.K_a && sym <= sdl.K_z {
								// Ctrl+A to Ctrl+Z map to 0x01 to 0x1A
								seq = []byte{byte(sym - sdl.K_a + 1)}
							} else {
								// Handle other specific Ctrl combinations if needed
								// e.g., Ctrl+[, Ctrl+\\, Ctrl+], Ctrl+^, Ctrl+_
								// We already handle Ctrl+C implicitly above (K_c -> 0x03)
							}
						} else {
							// Not Ctrl modifier, handle other keys
							switch sym {
							case sdl.K_ESCAPE:
								seq = []byte("\x1b") // Send ASCII ESC byte
							case sdl.K_RETURN, sdl.K_RETURN2:
								seq = []byte("\r")
							case sdl.K_BACKSPACE:
								seq = []byte("\x08") // BS (Backspace control code)
							case sdl.K_TAB:
								seq = []byte("\t")
							case sdl.K_UP:
								seq = []byte("\x1b[A")
							case sdl.K_DOWN:
								seq = []byte("\x1b[B")
							case sdl.K_RIGHT:
								seq = []byte("\x1b[C")
							case sdl.K_LEFT:
								seq = []byte("\x1b[D")
								// TODO: Handle Home, End, PgUp, PgDown, Delete, Insert, F-keys
								// TODO: Handle Alt modifier (often sends ESC prefix)
								// TODO: Handle Shift modifier (for keys that change, e.g., Shift+Tab)
							}
						}

						// Send sequence if one was generated
						if len(seq) > 0 {
							if _, err := ptmx.Write(seq); err != nil {
								log.Printf("Error writing key sequence to pty: %v", err)
							}
						} else {
							// If no special sequence, maybe log unhandled key press for debugging
							// log.Printf("Unhandled Key Press: Scancode=%d, Keycode=%d, Mod=%d", ev.Keysym.Scancode, ev.Keysym.Sym, ev.Keysym.Mod)
						}
					}
				// TODO: Handle Mouse events? (Scrolling, Selection)
				case *sdl.MouseWheelEvent:
					// Positive Y is scrolling away (Scroll Up)
					// Negative Y is scrolling towards (Scroll Down)
					const scrollAmount = 3 // Lines per tick
					var scrolled bool

					// Try scrolling the image first (ScrollImage expects positive delta for UP)
					scrolled = termRenderer.ScrollImage(int(ev.Y))

					// If image wasn't scrolled, scroll the buffer
					if !scrolled {
						if ev.Y > 0 {
							outBuffer.ScrollUp(scrollAmount)
							scrolled = true // Mark that buffer scrolling happened
						} else if ev.Y < 0 {
							outBuffer.ScrollDown(scrollAmount)
							scrolled = true // Mark that buffer scrolling happened
						}
					}

					// Request redraw if any scrolling occurred
					if scrolled {
						needsRedraw = true
					}
					// ev.X could be used for horizontal scrolling if needed
				}
			}
		}

		// --- Rendering ---
		if needsRedraw {
			rendererSDL.SetDrawColor(0, 0, 0, 255) // Black background
			rendererSDL.Clear()

			// Call the new SDL renderer
			if err := termRenderer.Draw(outBuffer); err != nil {
				log.Printf("Error drawing buffer: %v", err)
			}

			rendererSDL.Present()
			needsRedraw = false
		}
	}

	log.Println("Exiting gt.")
}
