package render

import (
	"fmt"
	"gt/buffer"
	"image"
	"image/draw" // Standard Go draw package
	"log"
	"unsafe"

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
	renderer          *sdl.Renderer
	cellWidth         int
	cellHeight        int

	// Image scrolling state
	imgScrollOffsetY     int32 // Current scroll offset for the scrollable image (pixels)
	scrollableImgTargetH int32 // Target height of the last potentially scrollable image drawn
	scrollableImgAnchorY int32 // Y position (pixels) where the top of the scrollable image is anchored
	lastWindowHeightPx   int32 // Last known window height in pixels
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
		sdlRenderer:       renderer,
		font:              font,
		boldFont:          boldFont,
		glyphWidth:        width,
		glyphHeight:       height,
		glyphCache:        make(map[glyphCacheKey]*sdl.Texture),
		imageTextureCache: make(map[imageCacheKey]*sdl.Texture), // Initialize image cache
		renderer:          renderer,
		cellWidth:         width,
		cellHeight:        height,
		// Initialize image scroll state
		imgScrollOffsetY:     0,
		scrollableImgTargetH: -1, // Indicate no scrollable image initially
		scrollableImgAnchorY: -1,
		lastWindowHeightPx:   -1,
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
		wh = r.lastWindowHeightPx
		if wh <= 0 {
			wh = 600
		} // Fallback
	}
	r.lastWindowHeightPx = wh // Store current window height
	termW := ww               // Use pixel width for termW in calculations

	// Reset scrollable image tracking for this frame.
	// It will be set if a scrollable image is encountered and drawn.
	r.scrollableImgTargetH = -1
	// Don't reset r.imgScrollOffsetY here, preserve it between frames

	grid := buf.GetVisibleGrid()
	rows := len(grid)
	cols := 0
	if rows > 0 {
		cols = len(grid[0])
	}

	// Keep track of areas covered by images in the current row
	imageSkipUntil := make(map[int]int) // map[row] => skip rendering text cells until col X

	for y := 0; y < rows; y++ {
		skipUntilCol, rowHasSkip := imageSkipUntil[y]
		for x := 0; x < cols; x++ {
			// If we are within a skipped area from a previous image, continue
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
							// Calculate target dimensions based on constraints
							// termW is now window pixel width
							targetW, targetH := r.calculateTargetDimensions(
								wConstraint, hConstraint, preserveAspect,
								texW, texH, termW, r.lastWindowHeightPx)

							// Store details if this image is potentially scrollable
							if targetH > r.lastWindowHeightPx {
								r.scrollableImgTargetH = targetH
								r.scrollableImgAnchorY = int32(y * r.cellHeight)
								// Ensure scroll offset is still valid
								maxScroll := targetH - r.lastWindowHeightPx
								r.imgScrollOffsetY = max(0, min(r.imgScrollOffsetY, maxScroll))
							}

							// Destination rect: position includes scroll offset, size is calculated target
							imgDstRect := sdl.Rect{
								X: int32(x * r.cellWidth),
								Y: int32(y*r.cellHeight) - r.imgScrollOffsetY, // Apply scroll offset
								W: targetW,                                    // Use calculated width
								H: targetH,                                    // Use calculated height
							}
							log.Printf("   -> Drawing image texture at [%d, %d] W:%d H:%d (ScrollY: %d)",
								imgDstRect.X, imgDstRect.Y, imgDstRect.W, imgDstRect.H, r.imgScrollOffsetY)

							// Set clip rect for this draw call to window bounds
							clipRect := sdl.Rect{X: 0, Y: 0, W: ww, H: wh}
							r.renderer.SetClipRect(&clipRect)

							errCopy := r.renderer.Copy(imgTexture, nil, &imgDstRect)

							// Reset clip rect
							r.renderer.SetClipRect(nil)

							if errCopy != nil {
								log.Printf("   -> Error copying image texture: %v", errCopy) // LOG 8
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
			fgColorSDL := mapBufferColorToSDL(fgCode, fgType)
			bgColorSDL := mapBufferColorToSDL(bgCode, bgType)

			// --- Draw Background ---
			bgRect := sdl.Rect{
				X: int32(x * r.glyphWidth),
				Y: int32(y * r.glyphHeight),
				W: int32(r.glyphWidth),
				H: int32(r.glyphHeight),
			}
			r.renderer.SetDrawColor(bgColorSDL.R, bgColorSDL.G, bgColorSDL.B, bgColorSDL.A)
			r.renderer.FillRect(&bgRect)

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

				// Copy texture to renderer
				dstRect := sdl.Rect{
					X: int32(x * r.glyphWidth),
					Y: int32(y * r.glyphHeight),
					W: int32(r.glyphWidth),  // Use fixed grid width
					H: int32(r.glyphHeight), // Use fixed grid height
				}
				// Use SrcRect=nil to copy whole texture, DstRect defines position and size
				r.renderer.Copy(texture, nil, &dstRect)

				// --- Draw Underline ---
				if cell.Underline {
					lineY := int32((y+1)*r.glyphHeight - 1) // Bottom of the cell
					r.renderer.SetDrawColor(fgColorSDL.R, fgColorSDL.G, fgColorSDL.B, fgColorSDL.A)
					r.renderer.DrawLine(dstRect.X, lineY, dstRect.X+dstRect.W, lineY)
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
			r.renderer.SetDrawColor(cursorColor.R, cursorColor.G, cursorColor.B, cursorColor.A)
			r.renderer.FillRect(&cursorRect)
		}
	}

	return nil
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

	surface.Lock()
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
			offset := (y*int(surface.Pitch) + x*4) // 4 bytes per pixel
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

// calculateTargetDimensions determines the final pixel width and height based on constraints.
// NOTE: Temporarily simplified to ignore constraints and fit width, preserving aspect.
func (r *SDLRenderer) calculateTargetDimensions(wConstraint, hConstraint string, preserveAspect bool, nativeW, nativeH, termW, termH int32) (targetW, targetH int32) {
	log.Printf("[CalcDims] Input: WConstraint=%s, HConstraint=%s, Preserve=%v, Native=%dx%d, Term=%dx%d",
		wConstraint, hConstraint, preserveAspect, nativeW, nativeH, termW, termH)

	// --- Simplified Logic ---
	if nativeW <= 0 || nativeH <= 0 {
		log.Printf("[CalcDims] Invalid native dimensions, returning 1x1.")
		return 1, 1 // Cannot calculate aspect ratio
	}

	// 1. Start with native dimensions
	targetW = nativeW
	targetH = nativeH
	log.Printf("[CalcDims] Step 1 (Native): %dx%d", targetW, targetH)

	nativeAspect := float64(nativeW) / float64(nativeH)

	// 2. Clamp width to terminal width
	originalTargetW := targetW
	targetW = max(1, min(targetW, termW))
	log.Printf("[CalcDims] Step 2 (Clamp Width): %dx%d (TermW: %d)", targetW, targetH, termW)

	// 3. If width was clamped, adjust height to preserve aspect ratio
	if targetW != originalTargetW {
		targetH = int32(float64(targetW) / nativeAspect)
		log.Printf("[CalcDims] Step 3 (Adjust Height for Aspect): %dx%d (NativeAspect: %f)", targetW, targetH, nativeAspect)
	}

	// 4. Ensure height is at least 1
	targetH = max(1, targetH)
	log.Printf("[CalcDims] Step 4 (Ensure Min Height): %dx%d", targetW, targetH)

	// --- Original Logic (commented out for debugging) ---
	/*
		passeConstraint := func(constraint string, nativeDim int32, cellDim int, termDimPx int32) int32 {
			// ... (previous parsing logic) ...
		}

		initialW := parseConstraint(wConstraint, nativeW, r.cellWidth, termW)
		initialH := parseConstraint(hConstraint, nativeH, r.cellHeight, termH)

		if preserveAspect && nativeW > 0 && nativeH > 0 {
			// ... (previous aspect logic) ...
		} else {
			// ... (previous non-aspect logic) ...
		}

		widthBeforeClamp := targetW
		targetW = max(1, min(targetW, termW))
		if preserveAspect && nativeW > 0 && nativeH > 0 && targetW != widthBeforeClamp {
			targetH = int32(float64(targetW) / (float64(nativeW) / float64(nativeH)))
		}
		if !preserveAspect {
			targetH = max(1, min(targetH, termH))
		} else {
			targetH = max(1, targetH)
		}
	*/

	log.Printf("[CalcDims] Final Output: %dx%d", targetW, targetH)
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
	// Check if we have a scrollable image recorded from the last draw
	if r.scrollableImgTargetH <= r.lastWindowHeightPx || r.lastWindowHeightPx <= 0 {
		r.scrollableImgTargetH = -1 // Reset if not scrollable
		return false                // Not scrollable or window height unknown
	}

	maxScroll := r.scrollableImgTargetH - r.lastWindowHeightPx
	deltaPx := int32(deltaY * r.cellHeight) // Convert lines to pixels

	newOffsetY := r.imgScrollOffsetY - deltaPx // Subtract delta because positive deltaY means scroll UP (show earlier part of image)

	// Clamp the new offset
	clampedOffsetY := max(0, min(newOffsetY, maxScroll))

	if clampedOffsetY != r.imgScrollOffsetY {
		r.imgScrollOffsetY = clampedOffsetY
		log.Printf("[ScrollImage] Scrolled. OffsetY: %d (Max: %d)", r.imgScrollOffsetY, maxScroll)
		return true // Scrolling happened
	}

	return false // No change in scroll offset
}
