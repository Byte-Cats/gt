package render

import (
	"gt/buffer"
	"log"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Extended 256 color palette (standard xterm)
// Source: https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
var xterm256Palette = []sdl.Color{
	// Standard colors (0-7)
	{0, 0, 0, 255}, {128, 0, 0, 255}, {0, 128, 0, 255}, {128, 128, 0, 255},
	{0, 0, 128, 255}, {128, 0, 128, 255}, {0, 128, 128, 255}, {192, 192, 192, 255},
	// Bright colors (8-15)
	{128, 128, 128, 255}, {255, 0, 0, 255}, {0, 255, 0, 255}, {255, 255, 0, 255},
	{0, 0, 255, 255}, {255, 0, 255, 255}, {0, 255, 255, 255}, {255, 255, 255, 255},
	// Color cube (16-231) 6x6x6
	// Levels: 0, 95, 135, 175, 215, 255
	{0, 0, 0, 255}, {0, 0, 95, 255}, {0, 0, 135, 255}, {0, 0, 175, 255}, {0, 0, 215, 255}, {0, 0, 255, 255},
	{0, 95, 0, 255}, {0, 95, 95, 255}, {0, 95, 135, 255}, {0, 95, 175, 255}, {0, 95, 215, 255}, {0, 95, 255, 255},
	{0, 135, 0, 255}, {0, 135, 95, 255}, {0, 135, 135, 255}, {0, 135, 175, 255}, {0, 135, 215, 255}, {0, 135, 255, 255},
	{0, 175, 0, 255}, {0, 175, 95, 255}, {0, 175, 135, 255}, {0, 175, 175, 255}, {0, 175, 215, 255}, {0, 175, 255, 255},
	{0, 215, 0, 255}, {0, 215, 95, 255}, {0, 215, 135, 255}, {0, 215, 175, 255}, {0, 215, 215, 255}, {0, 215, 255, 255},
	{0, 255, 0, 255}, {0, 255, 95, 255}, {0, 255, 135, 255}, {0, 255, 175, 255}, {0, 255, 215, 255}, {0, 255, 255, 255},
	{95, 0, 0, 255}, {95, 0, 95, 255}, {95, 0, 135, 255}, {95, 0, 175, 255}, {95, 0, 215, 255}, {95, 0, 255, 255},
	{95, 95, 0, 255}, {95, 95, 95, 255}, {95, 95, 135, 255}, {95, 95, 175, 255}, {95, 95, 215, 255}, {95, 95, 255, 255},
	{95, 135, 0, 255}, {95, 135, 95, 255}, {95, 135, 135, 255}, {95, 135, 175, 255}, {95, 135, 215, 255}, {95, 135, 255, 255},
	{95, 175, 0, 255}, {95, 175, 95, 255}, {95, 175, 135, 255}, {95, 175, 175, 255}, {95, 175, 215, 255}, {95, 175, 255, 255},
	{95, 215, 0, 255}, {95, 215, 95, 255}, {95, 215, 135, 255}, {95, 215, 175, 255}, {95, 215, 215, 255}, {95, 215, 255, 255},
	{95, 255, 0, 255}, {95, 255, 95, 255}, {95, 255, 135, 255}, {95, 255, 175, 255}, {95, 255, 215, 255}, {95, 255, 255, 255},
	{135, 0, 0, 255}, {135, 0, 95, 255}, {135, 0, 135, 255}, {135, 0, 175, 255}, {135, 0, 215, 255}, {135, 0, 255, 255},
	{135, 95, 0, 255}, {135, 95, 95, 255}, {135, 95, 135, 255}, {135, 95, 175, 255}, {135, 95, 215, 255}, {135, 95, 255, 255},
	{135, 135, 0, 255}, {135, 135, 95, 255}, {135, 135, 135, 255}, {135, 135, 175, 255}, {135, 135, 215, 255}, {135, 135, 255, 255},
	{135, 175, 0, 255}, {135, 175, 95, 255}, {135, 175, 135, 255}, {135, 175, 175, 255}, {135, 175, 215, 255}, {135, 175, 255, 255},
	{135, 215, 0, 255}, {135, 215, 95, 255}, {135, 215, 135, 255}, {135, 215, 175, 255}, {135, 215, 215, 255}, {135, 215, 255, 255},
	{135, 255, 0, 255}, {135, 255, 95, 255}, {135, 255, 135, 255}, {135, 255, 175, 255}, {135, 255, 215, 255}, {135, 255, 255, 255},
	{175, 0, 0, 255}, {175, 0, 95, 255}, {175, 0, 135, 255}, {175, 0, 175, 255}, {175, 0, 215, 255}, {175, 0, 255, 255},
	{175, 95, 0, 255}, {175, 95, 95, 255}, {175, 95, 135, 255}, {175, 95, 175, 255}, {175, 95, 215, 255}, {175, 95, 255, 255},
	{175, 135, 0, 255}, {175, 135, 95, 255}, {175, 135, 135, 255}, {175, 135, 175, 255}, {175, 135, 215, 255}, {175, 135, 255, 255},
	{175, 175, 0, 255}, {175, 175, 95, 255}, {175, 175, 135, 255}, {175, 175, 175, 255}, {175, 175, 215, 255}, {175, 175, 255, 255},
	{175, 215, 0, 255}, {175, 215, 95, 255}, {175, 215, 135, 255}, {175, 215, 175, 255}, {175, 215, 215, 255}, {175, 215, 255, 255},
	{175, 255, 0, 255}, {175, 255, 95, 255}, {175, 255, 135, 255}, {175, 255, 175, 255}, {175, 255, 215, 255}, {175, 255, 255, 255},
	{215, 0, 0, 255}, {215, 0, 95, 255}, {215, 0, 135, 255}, {215, 0, 175, 255}, {215, 0, 215, 255}, {215, 0, 255, 255},
	{215, 95, 0, 255}, {215, 95, 95, 255}, {215, 95, 135, 255}, {215, 95, 175, 255}, {215, 95, 215, 255}, {215, 95, 255, 255},
	{215, 135, 0, 255}, {215, 135, 95, 255}, {215, 135, 135, 255}, {215, 135, 175, 255}, {215, 135, 215, 255}, {215, 135, 255, 255},
	{215, 175, 0, 255}, {215, 175, 95, 255}, {215, 175, 135, 255}, {215, 175, 175, 255}, {215, 175, 215, 255}, {215, 175, 255, 255},
	{215, 215, 0, 255}, {215, 215, 95, 255}, {215, 215, 135, 255}, {215, 215, 175, 255}, {215, 215, 215, 255}, {215, 215, 255, 255},
	{215, 255, 0, 255}, {215, 255, 95, 255}, {215, 255, 135, 255}, {215, 255, 175, 255}, {215, 255, 215, 255}, {215, 255, 255, 255},
	{255, 0, 0, 255}, {255, 0, 95, 255}, {255, 0, 135, 255}, {255, 0, 175, 255}, {255, 0, 215, 255}, {255, 0, 255, 255},
	{255, 95, 0, 255}, {255, 95, 95, 255}, {255, 95, 135, 255}, {255, 95, 175, 255}, {255, 95, 215, 255}, {255, 95, 255, 255},
	{255, 135, 0, 255}, {255, 135, 95, 255}, {255, 135, 135, 255}, {255, 135, 175, 255}, {255, 135, 215, 255}, {255, 135, 255, 255},
	{255, 175, 0, 255}, {255, 175, 95, 255}, {255, 175, 135, 255}, {255, 175, 175, 255}, {255, 175, 215, 255}, {255, 175, 255, 255},
	{255, 215, 0, 255}, {255, 215, 95, 255}, {255, 215, 135, 255}, {255, 215, 175, 255}, {255, 215, 215, 255}, {255, 215, 255, 255},
	{255, 255, 0, 255}, {255, 255, 95, 255}, {255, 255, 135, 255}, {255, 255, 175, 255}, {255, 255, 215, 255}, {255, 255, 255, 255},
	// Grayscale ramp (232-255)
	{8, 8, 8, 255}, {18, 18, 18, 255}, {28, 28, 28, 255}, {38, 38, 38, 255}, {48, 48, 48, 255}, {58, 58, 58, 255},
	{68, 68, 68, 255}, {78, 78, 78, 255}, {88, 88, 88, 255}, {98, 98, 98, 255}, {108, 108, 108, 255}, {118, 118, 118, 255},
	{128, 128, 128, 255}, {138, 138, 138, 255}, {148, 148, 148, 255}, {158, 158, 158, 255}, {168, 168, 168, 255}, {178, 178, 178, 255},
	{188, 188, 188, 255}, {198, 198, 198, 255}, {208, 208, 208, 255}, {218, 218, 218, 255}, {228, 228, 228, 255}, {238, 238, 238, 255},
}

// Helper to generate the 6x6x6 color cube part of the palette
// func init() {
// 	levels := []uint8{0, 95, 135, 175, 215, 255}
// 	idx := 16
// 	for r := 0; r < 6; r++ {
// 		for g := 0; g < 6; g++ {
// 			for b := 0; b < 6; b++ {
// 				xterm256Palette[idx] = sdl.Color{R: levels[r], G: levels[g], B: levels[b], A: 255}
// 				idx++
// 			}
// 		}
// 	}
// }

// glyphCacheKey uniquely identifies a glyph variation (char + color for now)
// Note: SDL Color does not work directly as a map key, need comparable representation.
// Using Fg code from buffer might be simpler if mapBufferColorToSDL is deterministic.
// Let's try using the buffer code directly for simplicity.
// Alternatively, use a struct { char rune; r, g, b, a uint8 } or string representation.
type glyphCacheKey struct {
	char        rune
	fg          int // Buffer's color code/index
	fgColorType string
	bold        bool // Add bold to cache key
	// Bg needed if rendering involves it directly (e.g., specific blend modes)
	// For now, assume Bg is handled by background rect fill.
}

// SDLRenderer handles drawing the terminal buffer state using SDL.
type SDLRenderer struct {
	sdlRenderer *sdl.Renderer
	font        *ttf.Font
	boldFont    *ttf.Font // Add bold font variant
	glyphWidth  int
	glyphHeight int

	glyphCache map[glyphCacheKey]*sdl.Texture // Cache for rendered glyphs
}

// NewSDLRenderer creates a new renderer that draws to the given SDL renderer using the specified font.
func NewSDLRenderer(renderer *sdl.Renderer, font, boldFont *ttf.Font) *SDLRenderer {
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
		boldFont:    boldFont, // Store bold font
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
func (r *SDLRenderer) Draw(buf *buffer.Output) error {
	// Use GetVisibleGrid which accounts for scrollback offset
	grid := buf.GetVisibleGrid()
	rows := len(grid)
	cols := 0
	if rows > 0 {
		cols = len(grid[0])
	}

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			cell := grid[y][x]

			// Skip rendering continuation cells of wide characters
			if cell.Width == 0 {
				continue
			}

			// --- Determine Colors & Draw Background ---
			fgCode := cell.Fg
			bgCode := cell.Bg
			fgType := cell.FgColorType
			bgType := cell.BgColorType

			if cell.Reverse {
				fgCode, bgCode = bgCode, fgCode
				fgType, bgType = bgType, fgType
			}
			fgColorSDL := mapBufferColorToSDL(fgCode, fgType)
			bgColorSDL := mapBufferColorToSDL(bgCode, bgType)

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
				// Select font based on bold attribute
				activeFont := r.font
				if cell.Bold && r.boldFont != nil {
					activeFont = r.boldFont
				}

				// Update cache key to include bold status
				key := glyphCacheKey{char: cell.Char, fg: fgCode, fgColorType: fgType, bold: cell.Bold}
				texture, found := r.glyphCache[key]

				if !found {
					// Not cached: Render, create texture, add to cache
					fnt := activeFont
					charStr := string(cell.Char)
					log.Printf("Rendering uncached char: '%s' (Rune: %U, Int: %d) Fg: %d (%s)", charStr, cell.Char, cell.Char, fgCode, fgType) // More detailed log
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
					surface.Free()
					r.glyphCache[key] = texture
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
	// Only draw the cursor if we are viewing the live screen (offset == 0)
	if buf.IsLiveView() { // Need to add IsLiveView() to buffer
		cursorX, cursorY := buf.GetCursorPos()
		// Ensure cursor pos is valid for the current grid dimensions
		if cursorY >= 0 && cursorY < rows && cursorX >= 0 && cursorX < cols {
			cursorRect := sdl.Rect{
				X: int32(cursorX * r.glyphWidth),
				Y: int32(cursorY * r.glyphHeight),
				W: int32(r.glyphWidth),
				H: int32(r.glyphHeight),
			}
			cursorColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
			r.sdlRenderer.SetDrawColor(cursorColor.R, cursorColor.G, cursorColor.B, cursorColor.A)
			r.sdlRenderer.FillRect(&cursorRect)
		}
	}

	return nil
}

// mapBufferColorToSDL converts buffer color codes/indices to SDL colors based on type.
func mapBufferColorToSDL(value int, colorType string) sdl.Color {
	switch colorType {
	case buffer.ColorTypeStandard:
		// Handle default codes
		if value == buffer.FgDefault {
			return sdl.Color{R: 204, G: 204, B: 204, A: 255} // Default FG
		}
		if value == buffer.BgDefault {
			return sdl.Color{R: 0, G: 0, B: 0, A: 255} // Default BG
		}
		// Map standard 16 color codes (30-37, 40-47) - Assume bright codes map to bright colors if added later
		var index int
		if value >= buffer.BgBlack {
			index = value - buffer.BgBlack
		} else {
			index = value - buffer.FgBlack
		}
		if index >= 0 && index < 16 {
			// Use first 16 entries of the 256 palette for standard/bright mapping
			return xterm256Palette[index]
		}

	case buffer.ColorType256:
		if value >= 0 && value <= 255 {
			return xterm256Palette[value]
		}

	case buffer.ColorTypeTrue:
		// Unpack RGB from the integer value
		r := uint8((value >> 16) & 0xFF)
		g := uint8((value >> 8) & 0xFF)
		b := uint8(value & 0xFF)
		return sdl.Color{R: r, G: g, B: b, A: 255}
	}

	// Fallback / Unknown: return default foreground color
	return sdl.Color{R: 204, G: 204, B: 204, A: 255}
}

// ClearScreen is no longer needed here as SDL clearing is handled in main loop.
// func (r *SDLRenderer) ClearScreen() error { ... }
