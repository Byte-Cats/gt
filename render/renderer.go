package render

import (
	"gt/buffer"
	"log"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// glyphCacheKey uniquely identifies a glyph variation (char + color for now)
// Note: SDL Color does not work directly as a map key, need comparable representation.
// Using Fg code from buffer might be simpler if mapBufferColorToSDL is deterministic.
// Let's try using the buffer code directly for simplicity.
// Alternatively, use a struct { char rune; r, g, b, a uint8 } or string representation.
type glyphCacheKey struct {
	char rune
	fg   int // Using buffer's Fg code
	// TODO: Add other attributes like Bold, Underline if they affect glyph rendering itself
}

// SDLRenderer handles drawing the terminal buffer state using SDL.
type SDLRenderer struct {
	sdlRenderer *sdl.Renderer
	font        *ttf.Font
	glyphWidth  int
	glyphHeight int

	glyphCache map[glyphCacheKey]*sdl.Texture // Cache for rendered glyphs
}

// NewSDLRenderer creates a new renderer that draws to the given SDL renderer using the specified font.
func NewSDLRenderer(renderer *sdl.Renderer, font *ttf.Font) *SDLRenderer {
	// Calculate glyph dimensions (assuming monospace)
	// Error handling should ideally happen before calling NewSDLRenderer
	width, height, _ := font.SizeUTF8("W")
	// Use GlyphMetrics for potentially more accurate height
	metrics, err := font.GlyphMetrics('W')
	if err == nil {
		height = metrics.MaxY - metrics.MinY // Or just use font.Height()?
	}
	height = font.Height() // This is often the most reliable

	return &SDLRenderer{
		sdlRenderer: renderer,
		font:        font,
		glyphWidth:  width,
		glyphHeight: height,
		glyphCache:  make(map[glyphCacheKey]*sdl.Texture), // Initialize the cache
	}
}

// Destroy frees resources used by the renderer, including cached textures.
func (r *SDLRenderer) Destroy() {
	for _, texture := range r.glyphCache {
		texture.Destroy()
	}
	// Clear the map? Not strictly necessary if renderer is discarded.
	r.glyphCache = nil
}

// Draw renders the current state of the buffer to the SDL renderer.
// Assumes the caller handles Clear and Present.
func (r *SDLRenderer) Draw(buf *buffer.Output) error {
	grid := buf.GetGrid()
	rows := len(grid)
	cols := 0
	if rows > 0 {
		cols = len(grid[0])
	}

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			cell := grid[y][x]

			// Determine colors, handling reverse video
			fgCode := cell.Fg
			bgCode := cell.Bg
			if cell.Reverse {
				fgCode, bgCode = bgCode, fgCode
			}
			fgColorSDL := mapBufferColorToSDL(fgCode)
			bgColorSDL := mapBufferColorToSDL(bgCode)

			// --- Draw Background ---
			bgRect := sdl.Rect{
				X: int32(x * r.glyphWidth),
				Y: int32(y * r.glyphHeight),
				W: int32(r.glyphWidth),
				H: int32(r.glyphHeight),
			}
			r.sdlRenderer.SetDrawColor(bgColorSDL.R, bgColorSDL.G, bgColorSDL.B, bgColorSDL.A)
			r.sdlRenderer.FillRect(&bgRect)

			// --- Draw Character (if not blank) using Cache ---
			if cell.Char != ' ' {
				key := glyphCacheKey{char: cell.Char, fg: fgCode}
				texture, found := r.glyphCache[key]

				if !found {
					// Not cached: Render, create texture, add to cache
					fnt := r.font // Assign to local variable
					charStr := string(cell.Char)
					log.Printf("Rendering uncached char: '%s' (Rune: %U, Int: %d)", charStr, cell.Char, cell.Char) // DEBUG LOG
					surface, err := fnt.RenderUTF8Blended(charStr, fgColorSDL)
					if err != nil {
						log.Printf("  -> Failed to render char '%c' (%d): %v", cell.Char, cell.Char, err)
						continue
					}
					texture, err = r.sdlRenderer.CreateTextureFromSurface(surface)
					if err != nil {
						surface.Free()
						log.Printf("Failed to create texture for char '%c': %v", cell.Char, err)
						continue
					}
					surface.Free()              // Free surface, texture holds the data
					r.glyphCache[key] = texture // Add to cache
				}

				// Get texture dimensions (needed whether cached or newly created)
				// We don't strictly need W/H here anymore if using fixed grid size for drawing
				_, _, _, _, err := texture.Query() // Assign W and H to blank identifier
				if err != nil {
					log.Printf("Failed to query texture for char '%c': %v", cell.Char, err)
					continue // Should not happen often
				}

				// Copy texture to renderer
				dstRect := sdl.Rect{
					X: int32(x * r.glyphWidth),
					Y: int32(y * r.glyphHeight),
					W: int32(r.glyphWidth),  // Use fixed grid width
					H: int32(r.glyphHeight), // Use fixed grid height
				}
				// Use SrcRect=nil to copy whole texture, DstRect defines position and size
				r.sdlRenderer.Copy(texture, nil, &dstRect)

				// --- Draw Underline ---
				if cell.Underline {
					lineY := int32((y+1)*r.glyphHeight - 1) // Bottom of the cell
					r.sdlRenderer.SetDrawColor(fgColorSDL.R, fgColorSDL.G, fgColorSDL.B, fgColorSDL.A)
					r.sdlRenderer.DrawLine(dstRect.X, lineY, dstRect.X+dstRect.W, lineY)
				}

				// TODO: Handle Bold (maybe render again slightly offset, or use bold font variant if loaded)
			}
		}
	}

	// --- Draw Cursor ---
	cursorX, cursorY := buf.GetCursorPos()
	cursorRect := sdl.Rect{
		X: int32(cursorX * r.glyphWidth),
		Y: int32(cursorY * r.glyphHeight),
		W: int32(r.glyphWidth),
		H: int32(r.glyphHeight),
	}
	// Simple block cursor - use inverted color of the cell underneath or default white
	// For simplicity, just draw white block for now
	cursorColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	r.sdlRenderer.SetDrawColor(cursorColor.R, cursorColor.G, cursorColor.B, cursorColor.A)
	r.sdlRenderer.FillRect(&cursorRect)
	// TODO: Blinking cursor? Different cursor shapes?

	return nil
}

// mapBufferColorToSDL converts buffer color codes to SDL colors.
func mapBufferColorToSDL(code int) sdl.Color {
	// Handle default codes first
	if code == buffer.FgDefault {
		// TODO: Make default colors configurable
		return sdl.Color{R: 204, G: 204, B: 204, A: 255} // Light Gray
	}
	if code == buffer.BgDefault {
		return sdl.Color{R: 0, G: 0, B: 0, A: 255} // Black
	}

	// Map standard 16 colors
	switch code {
	// Foreground/Background normal intensity
	case buffer.FgBlack, buffer.BgBlack:
		return sdl.Color{R: 0, G: 0, B: 0, A: 255}
	case buffer.FgRed, buffer.BgRed:
		return sdl.Color{R: 205, G: 49, B: 49, A: 255}
	case buffer.FgGreen, buffer.BgGreen:
		return sdl.Color{R: 13, G: 188, B: 121, A: 255}
	case buffer.FgYellow, buffer.BgYellow:
		return sdl.Color{R: 229, G: 229, B: 16, A: 255}
	case buffer.FgBlue, buffer.BgBlue:
		return sdl.Color{R: 36, G: 114, B: 200, A: 255}
	case buffer.FgMagenta, buffer.BgMagenta:
		return sdl.Color{R: 188, G: 63, B: 188, A: 255}
	case buffer.FgCyan, buffer.BgCyan:
		return sdl.Color{R: 17, G: 168, B: 205, A: 255}
	case buffer.FgWhite, buffer.BgWhite:
		return sdl.Color{R: 229, G: 229, B: 229, A: 255}
	// TODO: Add bright colors (90-97, 100-107)
	// TODO: Add 256 / TrueColor mapping
	default:
		// Fallback to default foreground if code is unknown
		return sdl.Color{R: 204, G: 204, B: 204, A: 255}
	}
}

// ClearScreen is no longer needed here as SDL clearing is handled in main loop.
// func (r *SDLRenderer) ClearScreen() error { ... }
