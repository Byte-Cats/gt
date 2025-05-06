package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"gt/buffer"
	"gt/config"
	"gt/render"

	"github.com/creack/pty"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
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
	if err := runApp(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func runApp() error {
	runtime.LockOSThread() // SDL requires the main loop on the main thread

	// --- SDL/TTF Initialization ---
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return fmt.Errorf("could not initialize SDL: %w", err)
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		return fmt.Errorf("could not initialize SDL_ttf: %w", err)
	}
	defer ttf.Quit()

	// --- Configuration Loading ---
	theme := config.LoadTheme() // LoadTheme only returns the theme
	// Removed error check here as LoadTheme handles fallback internally and logs warnings.

	// --- Font Loading ---
	// Ensure FreeType library is available for bold variants
	// This might only be relevant on certain platforms or build configurations.
	// Consider if explicit FreeType linking/handling is needed.

	// Try loading the regular font first
	font, err := ttf.OpenFont(theme.FontPath, theme.FontSize)
	if err != nil {
		return fmt.Errorf("failed to open font at %s: %w", theme.FontPath, err)
	}
	log.Printf("Loaded font: %s, size: %d", theme.FontPath, theme.FontSize)
	defer font.Close()

	// Try loading the bold font variant
	var boldFont *ttf.Font
	// Specific handling for Space Mono in user library
	spaceMonoRegularPath := "/Users/fource/Library/Fonts/SpaceMono-Regular.ttf"
	spaceMonoBoldPath := "/Users/fource/Library/Fonts/SpaceMono-Bold.ttf"

	if theme.FontPath == spaceMonoRegularPath {
		boldFont, err = ttf.OpenFont(spaceMonoBoldPath, theme.FontSize)
		if err != nil {
			log.Printf("Warning: Could not load Space Mono Bold variant from %s: %v. Proceeding without bold.", spaceMonoBoldPath, err)
			boldFont = nil
		} else {
			log.Printf("Loaded bold font variant: %s", spaceMonoBoldPath)
		}
	} else {
		// Fallback for other fonts (e.g., TTC collections like Menlo or SFMono if path was different)
		boldFont, err = ttf.OpenFontIndex(theme.FontPath, theme.FontSize, 1)
		if err != nil {
			log.Printf("Warning: Could not load bold font variant from %s (index 1): %v. Proceeding without bold.", theme.FontPath, err)
			boldFont = nil // Proceed without bold variant
		} else {
			log.Printf("Loaded bold font variant from %s (index 1)", theme.FontPath)
		}
	}

	if boldFont != nil {
		defer boldFont.Close()
	}

	// Calculate cell size based on font
	glyphWidth, glyphHeight, err := font.SizeUTF8("W") // Use a wide char for estimate
	if err != nil {
		return fmt.Errorf("failed to get font glyph size: %w", err)
	}
	// --- DEBUG LOG: Compare heights ---
	fontHeightFromHeightFunc := font.Height()
	log.Printf("[Metrics] font.SizeUTF8('W') -> Width: %d, Height: %d", glyphWidth, glyphHeight)
	log.Printf("[Metrics] font.Height() -> Height: %d", fontHeightFromHeightFunc)
	// Use font.Height() for consistency with renderer? Let's try it.
	glyphHeight = fontHeightFromHeightFunc // Ensure glyphHeight is set to the actual font height used by renderer
	log.Printf("[Metrics] Using glyphHeight = %d for PTY calculation", glyphHeight)
	// --- END DEBUG LOG ---

	// Define maximum initial window size in pixels
	const maxInitialWidthPx = 900
	const maxInitialHeightPx = 600

	// Initial desired size based on cols/rows (or defaults)
	desiredInitialCols := 80
	desiredInitialRows := 24
	calculatedWidth := glyphWidth * desiredInitialCols
	calculatedHeight := glyphHeight * desiredInitialRows // Padding will be added by renderer logic

	// Clamp calculated size to maximums
	initialWindowWidth := calculatedWidth
	if initialWindowWidth > maxInitialWidthPx {
		initialWindowWidth = maxInitialWidthPx
	}
	initialWindowHeight := calculatedHeight
	if initialWindowHeight > maxInitialHeightPx {
		initialWindowHeight = maxInitialHeightPx
	}

	log.Printf("Requesting Initial Window Size (capped at %dx%d): %dx%d",
		maxInitialWidthPx, maxInitialHeightPx, initialWindowWidth, initialWindowHeight)

	window, err := sdl.CreateWindow("gt Terminal", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(initialWindowWidth), int32(initialWindowHeight),
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE|sdl.WINDOW_ALLOW_HIGHDPI)
	if err != nil {
		return fmt.Errorf("failed to create window: %w", err)
	}
	defer window.Destroy()

	// Apply theme opacity if set and valid
	if theme.WindowOpacity > 0 && theme.WindowOpacity < 1.0 {
		err = window.SetWindowOpacity(theme.WindowOpacity)
		if err != nil {
			log.Printf("Warning: Failed to set window opacity to %.2f: %v", theme.WindowOpacity, err)
			// Continue even if opacity setting fails, as it might not be supported on all platforms/WMs
		} else {
			log.Printf("Set window opacity to %.2f", theme.WindowOpacity)
		}
	}

	// Call platform-specific window adjustments and get top padding
	topPadding := customizeWindow(window)
	log.Printf("[WindowSetup] Top padding from customizeWindow: %dpx", topPadding)

	rendererSDL, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}
	defer rendererSDL.Destroy()

	// Get the actual drawable size in pixels (important for HighDPI)
	drawableW, drawableH, err := rendererSDL.GetOutputSize()
	if err != nil {
		// Fallback or error handling if we can't get the size
		log.Printf("Warning: Could not get renderer output size: %v. Using window size.", err)
		// Use the capped window size as a fallback for drawable size
		drawableW = int32(initialWindowWidth)
		drawableH = int32(initialWindowHeight)
	}
	log.Printf("Initial Drawable Size: %dx%d pixels", drawableW, drawableH)

	// Calculate initial rows/cols based on *actual drawable* size and glyph *pixel* size
	if glyphWidth <= 0 || glyphHeight <= 0 {
		return fmt.Errorf("invalid glyph dimensions: %dx%d", glyphWidth, glyphHeight)
	}
	drawableContentHeight := int(drawableH) - topPadding
	// Ensure effective content height is a multiple of glyphHeight
	effectiveContentHeight := drawableContentHeight - (drawableContentHeight % glyphHeight)
	initialCols := int(drawableW) / glyphWidth
	initialRows := effectiveContentHeight / glyphHeight

	log.Printf("[WindowSetup] Initial PTY Calc: DrawableH=%d, TopPadding=%d, GlyphHeight=%d, CalculatedContentH=%d, EffectiveContentH=%d => Rows=%d, Cols=%d",
		drawableH, topPadding, glyphHeight, drawableContentHeight, effectiveContentHeight, initialRows, initialCols)

	if initialCols <= 0 {
		initialCols = 80
	} // Fallback if calculation yields zero/negative
	if initialRows <= 0 {
		initialRows = 24
	} // Fallback if calculation yields zero/negative
	log.Printf("Calculated Initial Grid based on Drawable Size: %d cols, %d rows", initialCols, initialRows)

	// --- PTY Setup ---
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}
	c := exec.Command(shell)
	// Set TERM environment variable for the PTY process
	c.Env = append(os.Environ(), "TERM=xterm-256color")

	// Start PTY with initial calculated size
	winSize := &pty.Winsize{Rows: uint16(initialRows), Cols: uint16(initialCols)}
	ptmx, err := pty.StartWithSize(c, winSize)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}
	defer func() { _ = ptmx.Close() }()

	// --- Buffer & Renderer Setup ---
	outBuffer := buffer.NewOutputBuffer(initialRows, initialCols)
	// Pass both fonts, theme, and top padding to the renderer
	termRenderer := render.NewSDLRenderer(rendererSDL, font, boldFont, theme, topPadding)
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
						// Ignore event data (ev.Data1, ev.Data2) as it might be in points.
						// Get the new drawable size directly from the renderer.
						newDrawableW, newDrawableH, err := rendererSDL.GetOutputSize()
						if err != nil {
							log.Printf("Error getting new renderer size on resize: %v", err)
							continue // Skip resize if we can't get the size
						}

						log.Printf("Window Resized Event. New Drawable Size: %dx%d", newDrawableW, newDrawableH)

						// Recalculate cols/rows based on new *drawable* size and glyph *pixel* size
						newDrawableContentHeight := int(newDrawableH) - topPadding
						// Ensure effective content height is a multiple of glyphHeight
						effectiveNewContentHeight := newDrawableContentHeight - (newDrawableContentHeight % glyphHeight)
						newCols := int(newDrawableW) / glyphWidth
						newRows := effectiveNewContentHeight / glyphHeight

						log.Printf("[WindowResize] PTY Calc: NewDrawableH=%d, TopPadding=%d, GlyphHeight=%d, CalculatedContentH=%d, EffectiveContentH=%d => Rows=%d, Cols=%d",
							newDrawableH, topPadding, glyphHeight, newDrawableContentHeight, effectiveNewContentHeight, newRows, newCols)

						if newCols > 0 && newRows > 0 {
							log.Printf("Recalculated Grid Size: %d cols, %d rows", newCols, newRows)
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
					// Positive Y is scrolling away (Scroll Up / Shift+Scroll Right)
					// Negative Y is scrolling towards (Scroll Down / Shift+Scroll Left)
					// Positive X is scrolling right (not always available, use Shift+Y)
					// Negative X is scrolling left (not always available, use Shift+Y)
					const scrollAmount = 3 // Lines per tick for vertical scroll
					var scrolled bool

					// Check for Shift modifier for horizontal scrolling
					modState := sdl.GetModState()
					isShift := modState&sdl.KMOD_SHIFT != 0

					if isShift {
						// Use ev.Y for horizontal scroll amount when Shift is pressed
						// Adjust delta: Positive Y (scroll away) means scroll Right (+deltaX)
						// Negative Y (scroll towards) means scroll Left (-deltaX)
						// We can scale ev.Y if needed, but let's pass it directly for now
						horizDelta := int(ev.Y)
						if ev.Direction == sdl.MOUSEWHEEL_FLIPPED { // Handle natural scrolling direction
							horizDelta *= -1
						}
						scrolled = termRenderer.ScrollImageHorizontal(horizDelta)
						log.Printf("[Scroll] Horizontal attempt (Shift + Y=%d). Scrolled: %v", ev.Y, scrolled)
					} else {
						// Vertical scroll (no shift)
						// Try scrolling the image first (ScrollImage expects positive delta for UP)
						vertDelta := int(ev.Y)
						if ev.Direction == sdl.MOUSEWHEEL_FLIPPED {
							vertDelta *= -1
						}
						scrolled = termRenderer.ScrollImage(vertDelta)

						// If image wasn't scrolled, scroll the buffer
						if !scrolled {
							if vertDelta > 0 { // Scroll Up (show older content)
								outBuffer.ScrollUp(scrollAmount)
								scrolled = true // Mark that buffer scrolling happened
							} else if vertDelta < 0 { // Scroll Down (show newer content)
								outBuffer.ScrollDown(scrollAmount)
								scrolled = true // Mark that buffer scrolling happened
							}
						}
						log.Printf("[Scroll] Vertical attempt (Y=%d). Scrolled: %v", ev.Y, scrolled)
					}

					// If any scrolling happened (image or buffer), trigger redraw
					if scrolled {
						needsRedraw = true
					}
				}
			}
		}

		// Check if buffer has changed (e.g., due to PTY output)
		if outBuffer.HasChanged() { // Add HasChanged() method to Output buffer
			needsRedraw = true
			outBuffer.ResetChanged() // Reset the flag after checking
		}

		// --- Rendering ---
		if needsRedraw {
			// Remove redundant clear; termRenderer.Draw handles the background
			// rendererSDL.SetDrawColor(0, 0, 0, 255) // Black background
			// rendererSDL.Clear()

			// Call the SDL renderer - it handles background drawing internally
			if err := termRenderer.Draw(outBuffer); err != nil {
				log.Printf("Error drawing buffer: %v", err)
				// Decide if error is fatal
			}

			rendererSDL.Present()
			needsRedraw = false
		}

		// Prevent busy-waiting if no events/redraws
		if !needsRedraw {
			runtime.Gosched() // Yield processor briefly
			// Alternatively, introduce a small delay if CPU usage is high
			// time.Sleep(1 * time.Millisecond)
		}
	}

	log.Println("Application loop finished cleanly.")
	return nil
}

// customizeWindow is defined in platform-specific files (main_darwin.go or main_other.go)
// func customizeWindow(window *sdl.Window) {}
