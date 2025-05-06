package render

import (
	"fmt"
	"gt/buffer"
	"image"
	"image/draw" // Standard Go draw package
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"gt/config"

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

// glyphCacheKey for text glyphs (now local or adjusted)
type glyphCacheKey struct {
	char        rune
	fg          int
	fgColorType string
	bold        bool
}

// imageCacheKey is the key for the image texture cache.
// It includes the buffer key (coordinates) and the unique image ID.
type imageCacheKey struct {
	BufKey buffer.ImageKey
	ImgID  int
}

// SDLRenderer handles drawing the terminal buffer state using SDL.
type SDLRenderer struct {
	sdlRenderer       *sdl.Renderer
	font              *ttf.Font
	boldFont          *ttf.Font
	glyphWidth        int
	glyphHeight       int
	glyphCache        map[glyphCacheKey]*sdl.Texture
	imageTextureCache map[imageCacheKey]*sdl.Texture // Cache for image textures
	noiseTexture      *sdl.Texture                   // For subtle background noise
	renderer          *sdl.Renderer
	cellWidth         int
	cellHeight        int

	// Image scrolling state
	imgScrollOffsetX     int32        // Horizontal scroll offset (pixels)
	imgScrollOffsetY     int32        // Vertical scroll offset (pixels)
	scrollableImgTargetW int32        // Target width of the last potentially scrollable image
	scrollableImgTargetH int32        // Target height of the last potentially scrollable image drawn
	scrollableImgAnchorX int32        // X position (pixels) where the left of the scrollable image is anchored
	scrollableImgAnchorY int32        // Y position (pixels) where the top of the scrollable image is anchored
	lastWindowWidthPx    int32        // Last known window width in pixels
	lastWindowHeightPx   int32        // Last known window height in pixels
	theme                config.Theme // Store the loaded theme
	topPaddingPx         int          // Top padding (e.g., for macOS title bar) in pixels
}

// NewSDLRenderer creates a new renderer that draws to the given SDL renderer using the specified font.
func NewSDLRenderer(renderer *sdl.Renderer, font, boldFont *ttf.Font, theme config.Theme, topPaddingPx int) *SDLRenderer {
	// Calculate glyph dimensions (assuming monospace)
	// Error handling should ideally happen before calling NewSDLRenderer
	width, height, _ := font.SizeUTF8("W")
	// Use GlyphMetrics for potentially more accurate height
	metrics, err := font.GlyphMetrics('W')
	if err == nil {
		height = metrics.MaxY - metrics.MinY // Or just use font.Height()?
	}
	height = font.Height() // This is often the most reliable

	// Initialize random seed for noise generation
	rand.Seed(time.Now().UnixNano())

	// --- Create Noise Texture (if enabled) ---
	var noiseTex *sdl.Texture
	if theme.Noise.Enabled {
		noiseTex, err = createNoiseTexture(renderer, 64, 64) // Create a 64x64 noise texture
		if err != nil {
			log.Printf("Warning: Failed to create noise texture: %v", err)
			noiseTex = nil // Proceed without noise if creation fails
		} else if noiseTex != nil {
			// Apply opacity (alpha modulation)
			noiseOpacity := uint8(theme.Noise.Opacity * 255)
			if err := noiseTex.SetAlphaMod(noiseOpacity); err != nil {
				log.Printf("Warning: Failed to set alpha modulation on noise texture: %v", err)
			}
			if err := noiseTex.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
				log.Printf("Warning: Failed to set blend mode on noise texture: %v", err)
			}
			log.Printf("Created noise texture with opacity %d/255", noiseOpacity)
		}
	}

	return &SDLRenderer{
		sdlRenderer:       renderer,
		font:              font,
		boldFont:          boldFont,
		glyphWidth:        width,
		glyphHeight:       height,
		glyphCache:        make(map[glyphCacheKey]*sdl.Texture),
		imageTextureCache: make(map[imageCacheKey]*sdl.Texture), // Initialize image cache
		noiseTexture:      noiseTex,                             // Store the noise texture
		renderer:          renderer,
		cellWidth:         width,
		cellHeight:        height,
		// Initialize image scroll state
		imgScrollOffsetX:     0,
		imgScrollOffsetY:     0,
		scrollableImgTargetW: -1, // Indicate no scrollable image initially
		scrollableImgTargetH: -1,
		scrollableImgAnchorX: -1,
		scrollableImgAnchorY: -1,
		lastWindowWidthPx:    -1,
		lastWindowHeightPx:   -1,
		theme:                theme,        // Store the theme
		topPaddingPx:         topPaddingPx, // Store top padding
	}
}

// Destroy frees resources used by the renderer, including cached textures.
func (r *SDLRenderer) Destroy() {
	// Destroy cached glyph textures
	for _, texture := range r.glyphCache {
		texture.Destroy()
	}
	// Destroy cached image textures
	for _, texture := range r.imageTextureCache { // Iterate over the new cache type
		texture.Destroy()
	}
	// Destroy noise texture
	if r.noiseTexture != nil {
		r.noiseTexture.Destroy()
	}
	// Destroy fonts
	r.glyphCache = nil
	r.imageTextureCache = nil
}

// Draw renders the current state of the buffer to the SDL renderer.
func (r *SDLRenderer) Draw(buf *buffer.Output) error {
	// Get current window dimensions
	ww, wh, err := r.renderer.GetOutputSize()
	if err != nil {
		log.Printf("Error getting renderer output size: %v", err)
		// Use last known size or default?
		ww = r.lastWindowWidthPx
		wh = r.lastWindowHeightPx
		if ww <= 0 {
			ww = 800
		} // Fallback
		if wh <= 0 {
			wh = 600
		} // Fallback
	}
	r.lastWindowWidthPx = ww  // Store current window width
	r.lastWindowHeightPx = wh // Store current window height
	termW := ww               // Use pixel width for termW in calculations
	termH := wh               // Use pixel height

	// log.Printf("[Renderer.Draw] Using topPaddingPx: %d", r.topPaddingPx)

	// Draw background (Solid or Gradient)
	if r.theme.Gradient.Enabled {
		err = r.drawGradientBackground(ww, wh, r.topPaddingPx)
		if err != nil {
			log.Printf("Error drawing gradient background: %v. Falling back to solid.", err)
			r.drawSolidBackground(ww, wh, r.topPaddingPx) // Fallback to solid on error
		}
	} else {
		r.drawSolidBackground(ww, wh, r.topPaddingPx)
	}

	// --- Draw Noise Texture (if enabled and created) ---
	if r.noiseTexture != nil && r.theme.Noise.Enabled {
		// Ensure blend mode is set for the renderer (might be redundant if texture has it, but good practice)
		r.renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

		drawableContentY := int32(r.topPaddingPx)
		drawableContentH := wh - drawableContentY
		if drawableContentH < 0 {
			drawableContentH = 0
		}

		_, _, noiseTexW, noiseTexH, _ := r.noiseTexture.Query()

		if noiseTexW > 0 && noiseTexH > 0 && drawableContentH > 0 {
			for tileY := int32(0); tileY*noiseTexH < drawableContentH; tileY++ {
				for tileX := int32(0); tileX*noiseTexW < ww; tileX++ {
					dstRect := sdl.Rect{
						X: tileX * noiseTexW,
						Y: drawableContentY + (tileY * noiseTexH),
						W: noiseTexW,
						H: noiseTexH,
					}
					// Clip the noise tile to the drawable area if it extends beyond
					if dstRect.X+dstRect.W > ww {
						dstRect.W = ww - dstRect.X
					}
					if dstRect.Y+dstRect.H > wh { // Compare with wh (total window height)
						dstRect.H = wh - dstRect.Y
					}
					if dstRect.W > 0 && dstRect.H > 0 {
						r.renderer.Copy(r.noiseTexture, nil, &dstRect)
					}
				}
			}
		}
		// It's generally good to reset blend mode if other operations don't expect it,
		// but for this terminal, BLENDMODE_BLEND is likely fine for subsequent text/cursor drawing.
		// r.renderer.SetDrawBlendMode(sdl.BLENDMODE_NONE) // Optionally reset
	}

	// --- Draw Inner Border/Shadow Effect (if enabled) ---
	if r.theme.Border.Enabled && r.theme.Border.Thickness > 0 {
		r.renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND) // Ensure blending is on

		drawableContentX := int32(0)
		drawableContentY := int32(r.topPaddingPx)
		drawableContentW := ww
		drawableContentH := wh - drawableContentY
		if drawableContentH < 0 {
			drawableContentH = 0
		}

		thickness := int32(r.theme.Border.Thickness)

		// Parse Highlight Color
		hlColor, errHl := parseHexColor(r.theme.Border.HighlightColor)
		if errHl != nil {
			log.Printf("Warning: Invalid border highlight color '%s': %v. Skipping highlight.", r.theme.Border.HighlightColor, errHl)
		} else {
			hlColor.A = uint8(r.theme.Border.HighlightOpacity * 255) // Apply opacity
			r.renderer.SetDrawColor(hlColor.R, hlColor.G, hlColor.B, hlColor.A)

			// Top highlight line (inside content area)
			if drawableContentH > 0 && thickness > 0 {
				topRect := sdl.Rect{X: drawableContentX, Y: drawableContentY, W: drawableContentW, H: thickness}
				r.renderer.FillRect(&topRect)
			}
			// Left highlight line (inside content area, below top highlight)
			if drawableContentW > 0 && thickness > 0 {
				leftRect := sdl.Rect{X: drawableContentX, Y: drawableContentY + thickness, W: thickness, H: drawableContentH - thickness}
				if leftRect.H < 0 {
					leftRect.H = 0
				}
				if leftRect.H > 0 {
					r.renderer.FillRect(&leftRect)
				}
			}
		}

		// Parse Shadow Color
		shColor, errSh := parseHexColor(r.theme.Border.ShadowColor)
		if errSh != nil {
			log.Printf("Warning: Invalid border shadow color '%s': %v. Skipping shadow.", r.theme.Border.ShadowColor, errSh)
		} else {
			shColor.A = uint8(r.theme.Border.ShadowOpacity * 255) // Apply opacity
			r.renderer.SetDrawColor(shColor.R, shColor.G, shColor.B, shColor.A)

			// Bottom shadow line (inside content area)
			if drawableContentH > 0 && thickness > 0 {
				bottomRect := sdl.Rect{X: drawableContentX, Y: drawableContentY + drawableContentH - thickness, W: drawableContentW, H: thickness}
				if bottomRect.Y < drawableContentY {
					bottomRect.Y = drawableContentY
					bottomRect.H = drawableContentH
				} // Clamp if too thick
				if bottomRect.H > 0 {
					r.renderer.FillRect(&bottomRect)
				}
			}
			// Right shadow line (inside content area, above bottom shadow)
			if drawableContentW > 0 && thickness > 0 {
				rightRect := sdl.Rect{X: drawableContentX + drawableContentW - thickness, Y: drawableContentY, W: thickness, H: drawableContentH - thickness}
				if rightRect.X < drawableContentX {
					rightRect.X = drawableContentX
					rightRect.W = drawableContentW
				} // Clamp if too thick
				if rightRect.H < 0 {
					rightRect.H = 0
				}
				if rightRect.H > 0 {
					r.renderer.FillRect(&rightRect)
				}
			}
		}
		// r.renderer.SetDrawBlendMode(sdl.BLENDMODE_NONE) // Optionally reset blend mode
	}

	// Reset scrollable image tracking for this frame.
	r.scrollableImgTargetW = -1
	r.scrollableImgTargetH = -1
	r.scrollableImgAnchorX = -1
	r.scrollableImgAnchorY = -1
	// Don't reset scroll offsets (imgScrollOffsetX, imgScrollOffsetY)

	grid := buf.GetVisibleGrid()
	rows := len(grid)
	cols := 0
	if rows > 0 {
		cols = len(grid[0])
	}

	// Keep track of areas covered by images in the current row
	imageSkipUntil := make(map[int]int) // map[row] => skip rendering text cells until col X

	// Set clip rect for all cell/cursor/image drawing below padding
	clipRect := sdl.Rect{X: 0, Y: int32(r.topPaddingPx), W: ww, H: wh - int32(r.topPaddingPx)}
	if clipRect.H < 0 {
		clipRect.H = 0
	}
	r.renderer.SetClipRect(&clipRect)

	for y := 0; y < rows; y++ {
		var skipUntilCol int
		var rowHasSkip bool
		if skip, ok := imageSkipUntil[y]; ok {
			skipUntilCol = skip
			rowHasSkip = true
		}

		for x := 0; x < cols; x++ {
			// Skip rendering if within an image area for this row
			if rowHasSkip && x < skipUntilCol {
				continue
			}

			cell := grid[y][x]

			// Skip rendering standard continuation cells of wide characters
			if cell.Width == 0 {
				continue
			}

			// --- Check for and Render Image Placeholder ---
			if cell.IsImagePlaceholder {
				imgKey := buffer.ImageKey{R: y, C: x}
				log.Printf("Found image placeholder at [%d, %d]", y, x) // LOG 1
				// Get image and its constraints
				img, imgID, wConstraint, hConstraint, preserveAspect := buf.GetImage(imgKey)

				if img != nil && imgID > 0 {
					cacheKey := imageCacheKey{BufKey: imgKey, ImgID: imgID}
					imgTexture, cached := r.imageTextureCache[cacheKey]
					var err error // Declare err here to be accessible later

					if !cached {
						log.Printf("   -> Image not in texture cache, creating...") // LOG 3
						imgBounds := img.Bounds()
						imgW, imgH := int32(imgBounds.Dx()), int32(imgBounds.Dy())
						var surface *sdl.Surface

						// Create an SDL surface based on the image type
						switch imgData := img.(type) {
						case *image.RGBA:
							surface, err = sdl.CreateRGBSurfaceWithFormatFrom(unsafe.Pointer(&imgData.Pix[0]), imgW, imgH, 32, int32(imgData.Stride), uint32(sdl.PIXELFORMAT_RGBA32))
						case *image.NRGBA:
							rgbaImg := image.NewRGBA(imgBounds)
							draw.Draw(rgbaImg, imgBounds, imgData, imgBounds.Min, draw.Src)
							surface, err = sdl.CreateRGBSurfaceWithFormatFrom(unsafe.Pointer(&rgbaImg.Pix[0]), imgW, imgH, 32, int32(rgbaImg.Stride), uint32(sdl.PIXELFORMAT_RGBA32))
						case *image.Gray:
							surface, err = sdl.CreateRGBSurfaceWithFormat(0, imgW, imgH, 8, uint32(sdl.PIXELFORMAT_INDEX8))
							if err == nil {
								var palette *sdl.Palette
								palette, err = sdl.AllocPalette(256)
								if err == nil {
									paletteColors := make([]sdl.Color, 256)
									for i := range paletteColors {
										paletteColors[i] = sdl.Color{R: uint8(i), G: uint8(i), B: uint8(i), A: 255}
									}
									palette.SetColors(paletteColors)
									surface.SetPalette(palette)
									palette.Free()
									pixels := surface.Pixels()
									copy(pixels, imgData.Pix)
								}
							}
						case *image.Paletted:
							surface, err = sdl.CreateRGBSurfaceWithFormat(0, imgW, imgH, 8, uint32(sdl.PIXELFORMAT_INDEX8))
							if err == nil {
								var palette *sdl.Palette
								palette, err = sdl.AllocPalette(len(imgData.Palette))
								if err == nil {
									paletteColors := make([]sdl.Color, len(imgData.Palette))
									for i, c := range imgData.Palette {
										r, g, b, a := c.RGBA()
										paletteColors[i] = sdl.Color{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
									}
									palette.SetColors(paletteColors)
									surface.SetPalette(palette)
									palette.Free()
									pixels := surface.Pixels()
									copy(pixels, imgData.Pix)
								}
							}
						default:
							log.Printf("   -> Unsupported image type for direct surface creation: %T. Converting to RGBA.", imgData)
							rgbaImg := image.NewRGBA(imgBounds)
							draw.Draw(rgbaImg, imgBounds, imgData, imgBounds.Min, draw.Src)
							surface, err = sdl.CreateRGBSurfaceWithFormatFrom(unsafe.Pointer(&rgbaImg.Pix[0]), imgW, imgH, 32, int32(rgbaImg.Stride), uint32(sdl.PIXELFORMAT_RGBA32))
						}

						// Now create the texture from the surface (if surface creation was successful)
						if err != nil {
							log.Printf("   -> Failed to create surface from image: %v", err)
							if surface != nil {
								surface.Free()
							}
						} else if surface == nil {
							log.Printf("   -> Surface is nil after image conversion attempts.")
						} else {
							log.Printf("   -> Created surface: %p (Format: %s)", surface, sdl.GetPixelFormatName(uint(surface.Format.Format))) // LOG 4
							newTexture, texErr := r.renderer.CreateTextureFromSurface(surface)
							surface.Free()
							if texErr != nil {
								log.Printf("   -> Failed to create texture from image surface: %v", texErr)
							} else {
								imgTexture = newTexture // Assign to the outer scope variable
								r.imageTextureCache[cacheKey] = imgTexture
								log.Printf("   -> Created and cached image texture: %p (ID: %d)", imgTexture, imgID) // LOG 5
							}
						}
					} else {
						log.Printf("   -> Found cached image texture: %p (ID: %d)", imgTexture, imgID) // LOG 6
					}

					// Draw the texture if we have one (either cached or newly created)
					if imgTexture != nil {
						_, _, texW, texH, queryErr := imgTexture.Query()
						if queryErr != nil {
							log.Printf("   -> Error querying image texture: %v", queryErr)
						} else {
							// Calculate target dimensions based on constraints (using termW and termH)
							targetW, targetH := r.calculateTargetDimensions(
								wConstraint, hConstraint, preserveAspect,
								texW, texH, termW, termH) // Pass termH now

							// Store details if this image is potentially scrollable (horizontal or vertical)
							windowVisibleWidth := termW // No horizontal padding assumed
							windowVisibleHeight := termH - int32(r.topPaddingPx)
							if windowVisibleHeight < 0 {
								windowVisibleHeight = 0
							}

							anchorX := int32(x * r.cellWidth)
							anchorY := int32(y*r.cellHeight) + int32(r.topPaddingPx) // Anchor includes padding

							if targetW > windowVisibleWidth {
								r.scrollableImgTargetW = targetW
								r.scrollableImgAnchorX = anchorX
								maxScrollX := targetW - windowVisibleWidth
								if maxScrollX < 0 {
									maxScrollX = 0
								}
								r.imgScrollOffsetX = max(0, min(r.imgScrollOffsetX, maxScrollX))
							}
							if targetH > windowVisibleHeight {
								r.scrollableImgTargetH = targetH
								r.scrollableImgAnchorY = anchorY
								// Add cellHeight buffer to maxScrollY
								maxScrollY := targetH - windowVisibleHeight + int32(r.cellHeight)
								if maxScrollY < 0 {
									maxScrollY = 0
								}
								r.imgScrollOffsetY = max(0, min(r.imgScrollOffsetY, maxScrollY))
							}

							// Destination rect: position includes top padding AND scroll offsets
							imgDstRect := sdl.Rect{
								X: anchorX - r.imgScrollOffsetX, // APPLY HORIZONTAL SCROLL
								Y: anchorY - r.imgScrollOffsetY, // APPLY VERTICAL SCROLL & PADDING
								W: targetW,
								H: targetH,
							}
							log.Printf("   -> Drawing image texture at Dst=[%d, %d] W:%d H:%d (Anchor=[%d,%d], Pad=%d, Scroll=[%d,%d])",
								imgDstRect.X, imgDstRect.Y, imgDstRect.W, imgDstRect.H,
								r.scrollableImgAnchorX, r.scrollableImgAnchorY,
								r.topPaddingPx, r.imgScrollOffsetX, r.imgScrollOffsetY)

							// Clip rect is already set outside the loop
							errCopy := r.renderer.Copy(imgTexture, nil, &imgDstRect)

							// Reset clip rect (done after loop)
							// r.renderer.SetClipRect(nil)

							if errCopy != nil {
								log.Printf("   -> Error copying image texture: %v", errCopy)
							}

							// Mark this cell as skipped for text rendering
							colsToSkip := (targetW + int32(r.cellWidth) - 1) / int32(r.cellWidth) // Round up based on target size
							rowsToSkip := (targetH + int32(r.cellHeight) - 1) / int32(r.cellHeight)
							for rowOffset := 0; rowOffset < int(rowsToSkip); rowOffset++ {
								currentSkipRow := y + rowOffset
								skipEndCol := x + int(colsToSkip)
								if existingSkip, ok := imageSkipUntil[currentSkipRow]; ok {
									if skipEndCol > existingSkip { // Extend skip if this image goes further
										imageSkipUntil[currentSkipRow] = skipEndCol
									}
								} else {
									imageSkipUntil[currentSkipRow] = skipEndCol
								}
							}
							// Restart inner loop to respect the new skip calculation immediately
							skipUntilCol = imageSkipUntil[y]
							rowHasSkip = true
							continue
						}
					}
				} else {
					log.Printf("  -> Image retrieved from buffer is nil") // LOG 10
				}
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
			// Use theme-aware color mapping
			fgColorSDL := r.mapBufferColorToSDL(fgCode, fgType)
			bgColorSDL := r.mapBufferColorToSDL(bgCode, bgType)

			// --- Draw Background --- APPLY PADDING
			bgRect := sdl.Rect{
				X: int32(x * r.glyphWidth),
				Y: int32(y*r.glyphHeight) + int32(r.topPaddingPx), // APPLY PADDING
				W: int32(r.glyphWidth),
				H: int32(r.glyphHeight),
			}
			// if y == 0 && x == 0 {
			// 	log.Printf("[Renderer.Draw] First cell bgRect.Y: %d (y=%d, glyphH=%d, pad=%d)",
			// 		bgRect.Y, y, r.glyphHeight, r.topPaddingPx)
			// }
			r.renderer.SetDrawColor(bgColorSDL.R, bgColorSDL.G, bgColorSDL.B, bgColorSDL.A)
			r.renderer.FillRect(&bgRect)

			// --- Draw Character (if not blank) using Cache --- APPLY PADDING
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
					texture, err = r.renderer.CreateTextureFromSurface(surface)
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

				// Copy texture to renderer - APPLY PADDING
				dstRect := sdl.Rect{
					X: int32(x * r.glyphWidth),
					Y: int32(y*r.glyphHeight) + int32(r.topPaddingPx), // APPLY PADDING
					W: int32(r.glyphWidth),
					H: int32(r.glyphHeight),
				}
				r.renderer.Copy(texture, nil, &dstRect)

				// --- Draw Underline --- APPLY PADDING
				if cell.Underline {
					// lineY := int32((y+1)*r.glyphHeight - 1) // Old calculation
					lineY := dstRect.Y + int32(r.glyphHeight) - 1 // Bottom of the padded cell
					r.renderer.SetDrawColor(fgColorSDL.R, fgColorSDL.G, fgColorSDL.B, fgColorSDL.A)
					r.renderer.DrawLine(dstRect.X, lineY, dstRect.X+dstRect.W, lineY)
				}

				// TODO: Handle Bold (maybe render again slightly offset, or use bold font variant if loaded)
			}
		}
	}

	// --- Draw Cursor --- APPLY PADDING
	if buf.IsLiveView() {
		cursorX, cursorY := buf.GetCursorPos()
		if cursorY >= 0 && cursorY < rows && cursorX >= 0 && cursorX < cols {
			cursorRect := sdl.Rect{
				X: int32(cursorX * r.glyphWidth),
				Y: int32(cursorY*r.glyphHeight) + int32(r.topPaddingPx), // APPLY PADDING
				W: int32(r.glyphWidth),
				H: int32(r.glyphHeight),
			}
			cursorColor, err := parseHexColor(r.theme.Colors.Cursor)
			if err != nil {
				log.Printf("Warning: Invalid cursor color in theme '%s': %v. Using white.", r.theme.Colors.Cursor, err)
				cursorColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
			}
			r.renderer.SetDrawColor(cursorColor.R, cursorColor.G, cursorColor.B, cursorColor.A)
			r.renderer.FillRect(&cursorRect)
		}
	}

	// Reset clip rect after drawing cells/cursor
	r.renderer.SetClipRect(nil)

	return nil
}

// drawSolidBackground clears the screen with a solid color from the theme, considering top padding.
func (r *SDLRenderer) drawSolidBackground(w, h int32, topPadding int) {
	bgColor, err := parseHexColor(r.theme.Colors.Background)
	if err != nil {
		log.Printf("Warning: Invalid background color in theme '%s': %v. Using black.", r.theme.Colors.Background, err)
		bgColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
	}
	r.renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
	// Clear the whole window first (might be needed if padding != 0)
	r.renderer.Clear()
	// Optionally, only clear the area below padding:
	// clearRect := sdl.Rect{X: 0, Y: int32(topPadding), W: w, H: h - int32(topPadding)}
	// r.renderer.FillRect(&clearRect)
}

// drawGradientBackground draws a gradient background based on theme settings, considering top padding.
func (r *SDLRenderer) drawGradientBackground(w, h int32, topPadding int) error {
	startColor, err1 := parseHexColor(r.theme.Gradient.StartColor)
	endColor, err2 := parseHexColor(r.theme.Gradient.EndColor)
	if err1 != nil || err2 != nil {
		log.Printf("Warning: Invalid gradient colors in theme '%s': %v, %v. Using solid background.", r.theme.Gradient.StartColor, err1, err2)
		r.drawSolidBackground(w, h, topPadding)
		return fmt.Errorf("invalid gradient colors: start=%v, end=%v", err1, err2)
	}

	if r.theme.Gradient.Direction == "horizontal" {
		for x := int32(0); x < w; x++ {
			t := float32(x) / float32(w-1) // Interpolation factor (0.0 to 1.0)
			col := interpolateColor(startColor, endColor, t)
			r.renderer.SetDrawColor(col.R, col.G, col.B, col.A)
			// Draw full height line for horizontal gradient
			r.renderer.DrawLine(x, 0, x, h-1)
		}
	} else { // Default to vertical
		// Draw the theme's main background color in the padded area
		if topPadding > 0 {
			padColor, err := parseHexColor(r.theme.Colors.Background) // Get main background color
			if err != nil {
				log.Printf("Warning: Invalid background color '%s' for padding: %v. Using black.", r.theme.Colors.Background, err)
				padColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
			}
			r.renderer.SetDrawColor(padColor.R, padColor.G, padColor.B, padColor.A)
			padRect := sdl.Rect{X: 0, Y: 0, W: w, H: int32(topPadding)}
			r.renderer.FillRect(&padRect)
		}
		// Draw gradient below padding
		drawH := h - int32(topPadding)
		startY := int32(topPadding)
		if drawH <= 0 {
			return nil // Nothing to draw below padding
		}
		for yOffset := int32(0); yOffset < drawH; yOffset++ {
			y := startY + yOffset
			t := float32(yOffset) / float32(drawH-1) // Interpolation factor (0.0 to 1.0 over drawH)
			col := interpolateColor(startColor, endColor, t)
			r.renderer.SetDrawColor(col.R, col.G, col.B, col.A)
			r.renderer.DrawLine(0, y, w-1, y)
		}
	}
	return nil
}

// interpolateColor linearly interpolates between two SDL colors.
func interpolateColor(c1, c2 sdl.Color, t float32) sdl.Color {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	r := float32(c1.R) + t*(float32(c2.R)-float32(c1.R))
	g := float32(c1.G) + t*(float32(c2.G)-float32(c1.G))
	b := float32(c1.B) + t*(float32(c2.B)-float32(c1.B))
	a := float32(c1.A) + t*(float32(c2.A)-float32(c1.A))
	return sdl.Color{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}

// imageToSurface converts image.Image to *sdl.Surface (needs careful pixel format handling)
// This is a simplified version assuming RGBA source.
func imageToSurface(img image.Image) (*sdl.Surface, error) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Create an SDL surface (try ARGB8888 which is common)
	surface, err := sdl.CreateRGBSurfaceWithFormat(0, int32(width), int32(height), 32, sdl.PIXELFORMAT_ARGB8888)
	if err != nil {
		return nil, fmt.Errorf("failed to create surface: %w", err)
	}

	surface.Lock() // Lock the surface before accessing pixels
	pixels := surface.Pixels()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			// RGBA() returns 16-bit values, convert to 8-bit
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			// Calculate index in the byte slice (assuming ARGB8888)
			// SDL pixel order might vary by endianness! This assumes little-endian like x86.
			// ARGB = [B G R A] in memory on little-endian? Check SDL docs!
			// Let's assume standard RGBA packing for simplicity first, might need adjustment.
			offset := (y*int(surface.Pitch) + x*4) // Correct: y is int, surface.Pitch is int
			if offset+3 < len(pixels) {
				// Assuming ARGB8888 on little-endian = BGRA byte order? Let's try typical RGBA order. Need confirmation.
				pixels[offset+0] = r8 // R
				pixels[offset+1] = g8 // G
				pixels[offset+2] = b8 // B
				pixels[offset+3] = a8 // A
			}
		}
	}
	surface.Unlock()

	return surface, nil
}

// parseHexColor converts a hex color string (e.g., "#RRGGBB") to sdl.Color.
func parseHexColor(s string) (sdl.Color, error) {
	if len(s) != 7 || s[0] != '#' {
		return sdl.Color{}, fmt.Errorf("invalid hex color format: %s", s)
	}
	r, errR := strconv.ParseUint(s[1:3], 16, 8)
	g, errG := strconv.ParseUint(s[3:5], 16, 8)
	b, errB := strconv.ParseUint(s[5:7], 16, 8)
	if errR != nil || errG != nil || errB != nil {
		return sdl.Color{}, fmt.Errorf("invalid hex value in color: %s", s)
	}
	return sdl.Color{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}

// mapBufferColorToSDL converts buffer color codes/indices to SDL colors based on type and theme.
func (r *SDLRenderer) mapBufferColorToSDL(value int, colorType string) sdl.Color {
	defaultFgColor, err := parseHexColor(r.theme.Colors.Foreground)
	if err != nil {
		defaultFgColor = sdl.Color{R: 204, G: 204, B: 204, A: 255} // Fallback
	}
	defaultBgColor, err := parseHexColor(r.theme.Colors.Background)
	if err != nil {
		defaultBgColor = sdl.Color{R: 0, G: 0, B: 0, A: 255} // Fallback
	}

	switch colorType {
	case buffer.ColorTypeStandard:
		// Handle default codes using theme
		if value == buffer.FgDefault {
			return defaultFgColor
		}
		if value == buffer.BgDefault {
			return defaultBgColor
		}

		// Map standard 16 color codes (30-37 FG, 40-47 BG)
		var index int
		var isBg bool
		if value >= buffer.BgBlack && value <= buffer.BgWhite {
			index = value - buffer.BgBlack
			isBg = true
		} else if value >= buffer.FgBlack && value <= buffer.FgWhite {
			index = value - buffer.FgBlack
		} else if value >= buffer.BgBrightBlack && value <= buffer.BgBrightWhite { // Handle bright BG colors 100-107
			index = value - buffer.BgBrightBlack + 8 // Map 100-107 to 8-15
			isBg = true
		} else if value >= buffer.FgBrightBlack && value <= buffer.FgBrightWhite { // Handle bright FG colors 90-97
			index = value - buffer.FgBrightBlack + 8 // Map 90-97 to 8-15
		} else {
			// Unknown standard code, return default fg/bg
			if isBg {
				return defaultBgColor
			} else {
				return defaultFgColor
			}
		}

		// Use theme colors for the 16 ANSI colors
		var hexColor string
		switch index {
		case 0:
			hexColor = r.theme.Colors.Black
		case 1:
			hexColor = r.theme.Colors.Red
		case 2:
			hexColor = r.theme.Colors.Green
		case 3:
			hexColor = r.theme.Colors.Yellow
		case 4:
			hexColor = r.theme.Colors.Blue
		case 5:
			hexColor = r.theme.Colors.Magenta
		case 6:
			hexColor = r.theme.Colors.Cyan
		case 7:
			hexColor = r.theme.Colors.White
		case 8:
			hexColor = r.theme.Colors.BrightBlack
		case 9:
			hexColor = r.theme.Colors.BrightRed
		case 10:
			hexColor = r.theme.Colors.BrightGreen
		case 11:
			hexColor = r.theme.Colors.BrightYellow
		case 12:
			hexColor = r.theme.Colors.BrightBlue
		case 13:
			hexColor = r.theme.Colors.BrightMagenta
		case 14:
			hexColor = r.theme.Colors.BrightCyan
		case 15:
			hexColor = r.theme.Colors.BrightWhite
		default:
			if isBg {
				return defaultBgColor
			} else {
				return defaultFgColor
			}
		}
		sdlCol, err := parseHexColor(hexColor)
		if err != nil {
			log.Printf("Warning: Invalid theme color for index %d ('%s'): %v", index, hexColor, err)
			if isBg {
				return defaultBgColor
			} else {
				return defaultFgColor
			}
		}
		return sdlCol

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
	return defaultFgColor
}

// ClearScreen is no longer needed here as SDL clearing is handled in main loop.
// func (r *SDLRenderer) ClearScreen() error { ... }

// calculateTargetDimensions determines the final pixel width and height based on constraints.
func (r *SDLRenderer) calculateTargetDimensions(wConstraint, hConstraint string, preserveAspect bool, nativeW, nativeH, termW, termH int32) (targetW, targetH int32) {
	log.Printf("[CalcDims] Input: WConstraint=%s, HConstraint=%s, Preserve=%v, Native=%dx%d, Term=%dx%d, Cell=%dx%d",
		wConstraint, hConstraint, preserveAspect, nativeW, nativeH, termW, termH, r.cellWidth, r.cellHeight)

	// Helper function to parse a single dimension constraint string
	parseConstraint := func(constraint string, nativeDim int32, cellDim int, termDimPx int32) int32 {
		constraint = strings.ToLower(strings.TrimSpace(constraint))
		if constraint == "auto" {
			return nativeDim
		}
		if strings.HasSuffix(constraint, "px") {
			valStr := strings.TrimSuffix(constraint, "px")
			val, err := strconv.Atoi(valStr)
			if err == nil && val > 0 {
				return int32(val)
			}
		} else if strings.HasSuffix(constraint, "%") {
			valStr := strings.TrimSuffix(constraint, "%")
			val, err := strconv.Atoi(valStr)
			if err == nil && val > 0 && termDimPx > 0 {
				// Calculate percentage of the terminal's drawable dimension
				return int32(float64(termDimPx) * (float64(val) / 100.0))
			}
		} else { // Assume number represents character cells
			val, err := strconv.Atoi(constraint)
			if err == nil && val > 0 && cellDim > 0 {
				return int32(val * cellDim)
			}
		}
		// Fallback to native dimension if parsing fails or is invalid
		return nativeDim
	}

	// Parse constraints to get initial desired pixel dimensions
	initialW := parseConstraint(wConstraint, nativeW, r.cellWidth, termW)
	initialH := parseConstraint(hConstraint, nativeH, r.cellHeight, termH)
	log.Printf("[CalcDims] Parsed Constraints: Initial W=%dpx, Initial H=%dpx", initialW, initialH)

	// --- Start Calculation ---
	targetW = initialW // Start with parsed/native values
	targetH = initialH
	nativeAspect := float64(0)
	if nativeW > 0 && nativeH > 0 {
		nativeAspect = float64(nativeW) / float64(nativeH)
	}

	isWidthAuto := (wConstraint == "auto" || wConstraint == "")
	isHeightAuto := (hConstraint == "auto" || hConstraint == "")

	if isWidthAuto {
		log.Printf("[CalcDims] Width is auto, setting targetW to termW (%dpx)", termW)
		targetW = termW // Force full width
		if preserveAspect && nativeAspect > 0 {
			targetH = int32(float64(targetW) / nativeAspect) // Calculate H based on full width
			log.Printf("[CalcDims] Width auto, Preserve aspect -> Calculated H: %dpx", targetH)
		} else {
			// Width auto, no aspect OR invalid aspect. Height remains initialH (parsed from constraint or native)
			log.Printf("[CalcDims] Width auto, No Aspect -> H remains initial: %dpx", targetH)
		}
	} else {
		// Width is specific (not auto)
		log.Printf("[CalcDims] Width is specific: %dpx", targetW)
		if preserveAspect && nativeAspect > 0 {
			if isHeightAuto {
				// Width specific, height auto: Calculate H based on specific W
				targetH = int32(float64(targetW) / nativeAspect)
				log.Printf("[CalcDims] Width specific, Height auto -> Calculated H: %dpx", targetH)
			} else {
				// Both dimensions specific, preserve aspect. Adjust the one that deviates more.
				currentAspect := float64(targetW) / float64(targetH)
				if currentAspect > nativeAspect { // Target is wider than native; adjust width based on height
					targetW = int32(float64(targetH) * nativeAspect)
				} else if currentAspect < nativeAspect { // Target is narrower than native; adjust height based on width
					targetH = int32(float64(targetW) / nativeAspect)
				}
				log.Printf("[CalcDims] Both specific, Preserve aspect -> Corrected: W=%dpx, H=%dpx", targetW, targetH)
			}
		} else {
			// Width specific, No aspect preservation. Height remains initialH.
			log.Printf("[CalcDims] Width specific, No Aspect. W=%d, H=%d", targetW, targetH)
		}
	}

	// Final validation/adjustment (NO downward clamping)
	targetW = max(1, targetW) // Ensure at least 1px wide
	targetH = max(1, targetH) // Ensure at least 1px high

	log.Printf("[CalcDims] Final Output (No Clamping): %dx%d", targetW, targetH)
	return targetW, targetH
}

// Helper functions min/max for int32
func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// ScrollImage attempts to scroll the last rendered tall image.
// deltaY is the number of scroll *lines* (positive for up, negative for down).
// Returns true if image scrolling occurred, false otherwise.
func (r *SDLRenderer) ScrollImage(deltaY int) bool {
	// Calculate visible height considering padding
	visibleHeight := r.lastWindowHeightPx - int32(r.topPaddingPx)
	if visibleHeight <= 0 {
		return false // Cannot determine scroll bounds if window isn't tall enough
	}

	// Check if we have a scrollable image recorded from the last draw
	if r.scrollableImgTargetH <= visibleHeight {
		r.scrollableImgTargetH = -1 // Reset if not scrollable
		return false                // Not scrollable
	}

	// Add cellHeight buffer to maxScroll
	maxScroll := r.scrollableImgTargetH - visibleHeight + int32(r.cellHeight)
	if maxScroll < 0 {
		maxScroll = 0
	}
	deltaPx := int32(deltaY * r.cellHeight) // Convert lines to pixels

	newOffsetY := r.imgScrollOffsetY - deltaPx // Subtract delta because positive deltaY means scroll UP (show earlier part of image)

	// Clamp the new offset using the adjusted maxScroll
	clampedOffsetY := max(0, min(newOffsetY, maxScroll))

	if clampedOffsetY != r.imgScrollOffsetY {
		r.imgScrollOffsetY = clampedOffsetY
		log.Printf("[ScrollImage] Scrolled. OffsetY: %d (Max: %d)", r.imgScrollOffsetY, maxScroll)
		return true // Scrolling happened
	}

	return false // No change in scroll offset
}

// ScrollImageHorizontal attempts to scroll the last rendered wide image horizontally.
// deltaX is the number of scroll *units* (positive for right, negative for left).
// Returns true if image scrolling occurred, false otherwise.
func (r *SDLRenderer) ScrollImageHorizontal(deltaX int) bool {
	// Get current terminal width
	termW := r.lastWindowWidthPx
	if termW <= 0 {
		return false // Cannot determine scroll bounds
	}

	// Check if we have a horizontally scrollable image recorded from the last draw
	if r.scrollableImgTargetW <= termW {
		r.scrollableImgTargetW = -1 // Reset if not scrollable
		return false                // Not scrollable
	}

	maxScrollX := r.scrollableImgTargetW - termW
	// Arbitrarily map scroll delta units to pixels (e.g., multiply by cell width?)
	// Let's use a fraction of cell width for finer control
	deltaPx := int32(deltaX * (r.cellWidth / 2))

	newOffsetX := r.imgScrollOffsetX + deltaPx // Add delta for horizontal scroll

	// Clamp the new offset
	clampedOffsetX := max(0, min(newOffsetX, maxScrollX))

	if clampedOffsetX != r.imgScrollOffsetX {
		r.imgScrollOffsetX = clampedOffsetX
		log.Printf("[ScrollImageHorizontal] Scrolled. OffsetX: %d (Max: %d)", r.imgScrollOffsetX, maxScrollX)
		return true // Scrolling happened
	}

	return false // No change in scroll offset
}

// --- Helper function to create a procedural noise texture ---
func createNoiseTexture(renderer *sdl.Renderer, width, height int32) (*sdl.Texture, error) {
	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_STREAMING, width, height)
	if err != nil {
		return nil, fmt.Errorf("failed to create noise texture: %w", err)
	}

	// It's important to set the blend mode for the texture itself if you want its alpha to be used correctly during copy
	if err := texture.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		texture.Destroy() // Clean up if setting blend mode fails
		return nil, fmt.Errorf("failed to set blend mode on noise texture: %w", err)
	}

	pixels, pitch, err := texture.Lock(nil)
	if err != nil {
		texture.Destroy() // Clean up
		return nil, fmt.Errorf("failed to lock noise texture: %w", err)
	}

	bytesPerPixel := 4 // Assuming RGBA8888
	for y := int32(0); y < height; y++ {
		for x := int32(0); x < width; x++ {
			gray := uint8(rand.Intn(256)) // Simple random noise
			// offset calculation fixed: cast pitch to int32, ensure correct byte offset
			offset := (y * int32(pitch)) + (x * int32(bytesPerPixel))
			// Bounds check for pixel slice access
			if offset+int32(bytesPerPixel-1) < int32(len(pixels)) {
				pixels[offset+0] = gray // R
				pixels[offset+1] = gray // G
				pixels[offset+2] = gray // B
				pixels[offset+3] = 255  // Alpha of the noise pixel itself (texture's global alpha is set by SetAlphaMod)
			} else {
				// This should ideally not happen if width, height, and pitch are correct
				log.Printf("Warning: Noise texture pixel access out of bounds at (%d, %d), offset %d", x, y, offset)
			}
		}
	}

	texture.Unlock()
	return texture, nil
}
