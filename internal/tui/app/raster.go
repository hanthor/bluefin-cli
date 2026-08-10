package app

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// RasterBackend enumerates the available pixel-rendering strategies, from
// lowest fidelity (half-blocks, works everywhere) to highest (true per-pixel
// rendering via terminal graphics protocols).
type RasterBackend int

const (
	RasterHalfBlock RasterBackend = iota // always-available ▀ half-block canvas
	RasterSixel                          // sixel graphics (foot, xterm, Windows Terminal)
	RasterKitty                          // kitty graphics protocol (kitty, ghostty, wezterm)
)

// DetectRasterBackend probes the running terminal for graphics protocol
// support and returns the best available backend.
//
// Detection order:
//   - Kitty: the KITTY_WINDOW_ID env var is set by kitty for every child
//     process; Ghostty and WezTerm set TERM to xterm-kitty / wezterm.
//   - Sixel: foot sets TERM=foot or TERM=foot-direct; xterm built with
//     --enable-sixel-graphics exposes it (TERM=xterm-256color is common but
//     we can only rely on foot's unambiguous TERM).
//   - Fallback: half-block PixelCanvas (always available).
func DetectRasterBackend() RasterBackend {
	// Kitty protocol: the definitive signal is KITTY_WINDOW_ID, which
	// kitty itself sets unconditionally. Ghostty sets TERM=xterm-kitty
	// (or xterm-ghostty). WezTerm sets TERM=wezterm.
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return RasterKitty
	}
	term := os.Getenv("TERM")
	if strings.HasSuffix(term, "-kitty") || strings.Contains(term, "kitty") {
		return RasterKitty
	}
	if strings.HasPrefix(term, "wezterm") {
		return RasterKitty
	}
	if strings.HasPrefix(term, "ghostty") {
		return RasterKitty
	}

	// Sixel: foot is the most reliably detectable terminal with sixel.
	// xterm may or may not have --enable-sixel-graphics compiled in, so
	// we do not assume it.
	if strings.Contains(term, "foot") {
		return RasterSixel
	}

	return RasterHalfBlock
}

// RasterEncoded is the output of a raster backend: the raw terminal bytes
// plus how many rows the image occupies in the terminal grid.
type RasterEncoded struct {
	Data string // terminal escape sequences (Kitty or Sixel)
	Rows int    // number of cell rows the image covers
	Cols int    // number of cell columns the image covers
}

// EncodeKitty produces a Kitty graphics-protocol transmission for the
// pixel canvas. The image is sent as 32-bit RGBA (f=32), which Kitty
// decodes directly.
//
// Protocol reference: https://sw.kovidgoyal.net/kitty/graphics-protocol/
//
// We use:
//
//	a=T  — transmit and display (no chunking for these small images)
//	f=32 — RGBA format (8 bits per channel)
//	s=w,v=h — dimensions in pixels
//	m=1  — display the image (0 = just cache)
//	q=2  — quiet mode (no response from terminal)
//
// The image data is base64-encoded raw RGBA pixels, row by row.
func (c *PixelCanvas) EncodeKitty() RasterEncoded {
	w, h := c.w, c.h
	if w <= 0 || h <= 0 {
		return RasterEncoded{Rows: h / 2, Cols: w}
	}

	rgba := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := (y*w + x) * 4
			col := c.px[y*c.w+x]
			if col == nil {
				// Transparent pixels become the terminal background.
				// Kitty's f=32 does not carry alpha for blending, so
				// transparent regions are sent as black with zero alpha
				// — kitty will use its own background.
				rgba[off+0] = 0
				rgba[off+1] = 0
				rgba[off+2] = 0
				rgba[off+3] = 0
				continue
			}
			r, g, b, a := col.RGBA()
			rgba[off+0] = byte(r >> 8)
			rgba[off+1] = byte(g >> 8)
			rgba[off+2] = byte(b >> 8)
			rgba[off+3] = byte(a >> 8)
		}
	}

	b64 := base64.StdEncoding.EncodeToString(rgba)
	// The Kitty graphics escape: \e_G ... \e\\
	// We use q=2 (quiet) to suppress the terminal response.
	// Cell pixel ratio: Kitty renders at native cell size without
	// scaling, so each pixel maps 1:1 onto the canvas.
	data := fmt.Sprintf("\x1b_Ga=T,f=32,s=%d,v=%d,m=1,q=2;%s\x1b\\", w, h, b64)

	// The image covers rows of terminal cells.  Kitty maps pixels to
	// cells using the cell size as reported by the terminal (typically
	// ~10x20 px).  We report a conservative estimate: one pixel-row per
	// cell-row for small canvases (game scenes); the caller can use
	// placement escape sequences to align the image.
	return RasterEncoded{Data: data, Rows: h, Cols: w}
}

// EncodeSixel produces a Sixel-encoded image for the pixel canvas.
// Sixel packs six vertical pixels into one character; the encoding
// includes a colour palette (up to 256 entries, shared across the image)
// and the pixel data in sixel-scanline format.
//
// Format: \ePq ... \e\\
// Inside: "palette#id!len~sixel_data
//
// For efficiency we quantise to a fixed palette by snapping each pixel
// to the closest of a small set of colour bands (6 levels per channel =
// 216 colours + 16 greys).  Transparent pixels are skipped (they show
// the terminal background).
func (c *PixelCanvas) EncodeSixel() RasterEncoded {
	w, h := c.w, c.h
	if w <= 0 || h <= 0 {
		return RasterEncoded{Rows: h / 2, Cols: w}
	}

	// Quantise to a 216-colour web-safe(-ish) palette: 6 levels per
	// channel (0, 51, 102, 153, 204, 255).  Map each colour to its index
	// and collect the used palette entries.
	type rgb struct{ r, g, b uint8 }
	paletteMap := map[rgb]int{}   // rgb -> sixel palette index
	pixelIdx := make([]int, w*h)  // palette index per pixel, -1 = transparent
	pixelIdxInit := false

	for i := range pixelIdx {
		col := c.px[i]
		if col == nil {
			pixelIdx[i] = -1
			continue
		}
		r32, g32, b32, _ := col.RGBA()
		// Snap each 8-bit channel to one of 6 levels.
		r8 := snapChannel(uint8(r32 >> 8))
		g8 := snapChannel(uint8(g32 >> 8))
		b8 := snapChannel(uint8(b32 >> 8))
		key := rgb{r8, g8, b8}
		idx, ok := paletteMap[key]
		if !ok {
			idx = len(paletteMap)
			if idx >= 256 {
				// Palette overflow — transparent fallback.
				pixelIdx[i] = -1
				continue
			}
			paletteMap[key] = idx
		}
		pixelIdx[i] = idx
		pixelIdxInit = true
	}
	if !pixelIdxInit || len(paletteMap) == 0 {
		return RasterEncoded{Rows: h / 2, Cols: w}
	}

	// Build the sixel palette definition string.
	// Format: #<idx>;2;<r>;<g>;<b>
	var palParts []string
	// Build reverse map: index -> rgb
	palRGB := make([]rgb, len(paletteMap))
	for c, idx := range paletteMap {
		palRGB[idx] = c
	}
	for idx := 0; idx < len(palRGB); idx++ {
		entry := palRGB[idx]
		// Sixel uses 0-100 range for RGB percentages.
		rPct := int(entry.r) * 100 / 255
		gPct := int(entry.g) * 100 / 255
		bPct := int(entry.b) * 100 / 255
		palParts = append(palParts, fmt.Sprintf("#%d;2;%d;%d;%d", idx, rPct, gPct, bPct))
	}

	// Build sixel data: process the image in 6-row bands.
	// Each sixel character encodes 6 vertical pixels in one column.
	// Bits are LSB = top pixel.
	var data strings.Builder
	bands := (h + 5) / 6

	for band := 0; band < bands; band++ {
		// For each band, emit a graphics carriage return ($) to reset
		// column position.
		if band > 0 {
			data.WriteByte('-') // move to next sixel row (next 6-pixel band)
		}

		lastColor := -1
		for col := 0; col < w; col++ {
			// Build the sixel byte for this column in this band.
			var sixel byte
			for sub := 0; sub < 6; sub++ {
				y := band*6 + sub
				if y >= h {
					break
				}
				pi := pixelIdx[y*w+col]
				if pi >= 0 {
					sixel |= 1 << sub
				}
			}
			if sixel == 0 {
				// No pixels in this column — advance to next column.
				// Emit a colour change to the first used palette entry
				// (so the cursor moves forward without drawing).
				continue
			}

			// Find the colour for this column (use the bottom-most
			// non-transparent pixel's colour index for the whole sixel).
			colorIdx := -1
			for sub := 0; sub < 6; sub++ {
				y := band*6 + sub
				if y >= h {
					break
				}
				pi := pixelIdx[y*w+col]
				if pi >= 0 {
					colorIdx = pi
				}
			}
			if colorIdx < 0 {
				continue
			}

			if colorIdx != lastColor {
				// Emit colour select: #<idx>
				data.WriteString(fmt.Sprintf("#%d", colorIdx))
				lastColor = colorIdx
			}
			// Emit the sixel character: the sixel value + 63 (0x3F) maps
			// to ASCII '?' (0) through '~' (63).
			data.WriteByte(sixelChar(sixel))
		}
	}

	sixelData := data.String()
	if sixelData == "" {
		return RasterEncoded{Rows: h / 2, Cols: w}
	}

	paletteStr := strings.Join(palParts, "")
	sixelOut := fmt.Sprintf("\x1bPq%s%s\x1b\\", paletteStr, sixelData)

	// Sixel images render at native resolution; the terminal determines
	// how many text rows they occupy.  A conservative estimate: one cell
	// row per 12 pixel rows.
	rows := (h + 11) / 12
	if rows < 1 {
		rows = 1
	}
	return RasterEncoded{Data: sixelOut, Rows: rows, Cols: w}
}

// snapChannel quantises an 8-bit channel value to one of 6 evenly-spaced
// values (0, 51, 102, 153, 204, 255).
func snapChannel(v uint8) uint8 {
	return uint8((int(v)*5 + 127) / 255 * 51)
}

// sixelChar returns the sixel printable character for the low 6 bits.
// Sixel values 0-63 map to ASCII chars '?' (0x3F) through '~' (0x7E).
func sixelChar(v byte) byte {
	return (v & 0x3F) + 0x3F
}

// RenderKitty returns the full Kitty graphics protocol image for this
// canvas, preserving the same interface shape as Render().
func (c *PixelCanvas) RenderKitty() string {
	enc := c.EncodeKitty()
	return enc.Data
}

// RenderSixel returns the full Sixel image for this canvas.
func (c *PixelCanvas) RenderSixel() string {
	enc := c.EncodeSixel()
	return enc.Data
}

// RasterCanvas wraps a PixelCanvas and auto-selects the best rendering
// backend (Kitty > Sixel > half-block) based on terminal capabilities.
type RasterCanvas struct {
	*PixelCanvas
	Backend RasterBackend
}

// NewRasterCanvas creates a canvas with the given backend strategy.
func NewRasterCanvas(w, rows int, backend RasterBackend) *RasterCanvas {
	return &RasterCanvas{
		PixelCanvas: NewCanvas(w, rows),
		Backend:     backend,
	}
}

// RenderImage returns the best available image representation for this
// canvas. When the backend is RasterHalfBlock it delegates to Render()
// (half-block ANSI).  Kitty and Sixel backends produce graphics protocol
// escape sequences.
func (rc *RasterCanvas) RenderImage() string {
	switch rc.Backend {
	case RasterKitty:
		return rc.RenderKitty()
	case RasterSixel:
		return rc.RenderSixel()
	default:
		return rc.Render()
	}
}

// ImageRows estimates the number of text rows the rendered image will
// occupy (conservative; the terminal may use more or less depending on
// cell size and scaling).
func (rc *RasterCanvas) ImageRows() int {
	switch rc.Backend {
	case RasterKitty:
		// Kitty renders at native pixel resolution within a placement
		// rectangle.  Conservative: 1 row per pixel row (small canvases).
		return rc.h
	case RasterSixel:
		return (rc.h + 11) / 12
	default:
		return rc.h / 2
	}
}

// RenderImageWithText returns a combined view that places the graphics
// image at the top and overlays the given text rows on top.  The text is
// rendered using the standard half-block Render() for the canvas region
// so that bubbletea's cell-diff has something to diff against for the
// text overlay portions.
//
// For graphics backends the image escape sequence is emitted with a
// cursor-save/restore pair and absolute positioning so the image
// stays anchored while text scrolls above and below.
func RenderImageWithText(cv *RasterCanvas, header, footer string) string {
	var b strings.Builder

	if cv.Backend == RasterHalfBlock {
		if header != "" {
			b.WriteString(header)
			b.WriteString("\n")
		}
		b.WriteString(cv.Render())
		if footer != "" {
			b.WriteString("\n")
			b.WriteString(footer)
		}
		return b.String()
	}

	// For Kitty/Sixel: emit the image with absolute positioning.
	// We save the cursor, move to the target row, emit the image,
	// then restore.  The header and footer surround the image and
	// are rendered as normal text.
	if header != "" {
		b.WriteString(header)
	}
	b.WriteString("\n")
	// Save cursor and position the image just below the header.
	b.WriteString("\x1b7") // DECSC — save cursor
	// The image will draw starting at the current cursor row.
	b.WriteString(cv.RenderImage())
	b.WriteString("\x1b8") // DECRC — restore cursor
	if footer != "" {
		b.WriteString("\n")
		b.WriteString(footer)
	}
	return b.String()
}

// DefaultRasterBackend is the singleton result of DetectRasterBackend(),
// called once during app initialisation so we do not probe the terminal
// on every frame.
var DefaultRasterBackend = DetectRasterBackend()
