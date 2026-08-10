package app

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestRasterCanvasHalfBlockFallback(t *testing.T) {
	// When backend is HalfBlock, RenderImage must match Render().
	red := lipgloss.Color("#ff0000")
	blue := lipgloss.Color("#0000ff")

	rc := NewRasterCanvas(3, 1, RasterHalfBlock)
	rc.Set(0, 0, red)
	rc.Set(1, 1, blue)
	rc.Set(2, 0, red)
	rc.Set(2, 1, blue)

	img := rc.RenderImage()
	plain := stripAnsi(img)
	if plain != "▀▄▀" {
		t.Errorf("half-block RenderImage = %q, want half-block glyphs", plain)
	}
}

func TestRasterKittyEncodesRGBA(t *testing.T) {
	green := lipgloss.Color("#00ff00")
	rc := NewRasterCanvas(2, 1, RasterKitty) // 2x2 pixels
	rc.Set(0, 0, green)
	rc.Set(1, 1, green)

	out := rc.RenderKitty()
	if !strings.HasPrefix(out, "\x1b_G") {
		t.Errorf("kitty output should start with \\x1b_G, got %q", out)
	}
	if !strings.HasSuffix(out, "\x1b\\") {
		t.Errorf("kitty output should end with \\x1b\\\\, got %q", out)
	}
	if !strings.Contains(out, "f=32") {
		t.Error("kitty output should use RGBA format f=32")
	}
	if !strings.Contains(out, "s=2,v=2") {
		t.Error("kitty output must advertise pixel dimensions")
	}
}

func TestRasterSixelEncodesData(t *testing.T) {
	green := lipgloss.Color("#00ff00")
	rc := NewRasterCanvas(2, 1, RasterSixel)
	rc.Set(0, 0, green)
	rc.Set(1, 1, green)

	out := rc.RenderSixel()
	if !strings.HasPrefix(out, "\x1bPq") {
		t.Errorf("sixel output should start with \\x1bPq, got %q", out)
	}
	if !strings.HasSuffix(out, "\x1b\\") {
		t.Errorf("sixel output should end with \\x1b\\\\, got %q", out)
	}
	// Must contain at least one palette definition.
	if !strings.Contains(out, ";2;") {
		t.Error("sixel output should contain palette definitions")
	}
}

func TestRasterEmptyCanvas(t *testing.T) {
	for _, backend := range []RasterBackend{RasterHalfBlock, RasterKitty, RasterSixel} {
		rc := NewRasterCanvas(3, 1, backend)
		img := rc.RenderImage()
		if img == "" && backend != RasterHalfBlock {
			// An empty kitty/sixel canvas may return "" (no data),
			// which is fine — it just means there is nothing to paint.
			continue
		}
		_ = img // must not panic
	}
}

func TestRenderImageWithTextHalfBlock(t *testing.T) {
	green := lipgloss.Color("#00ff00")
	rc := NewRasterCanvas(2, 1, RasterHalfBlock)
	rc.Set(0, 0, green)

	out := RenderImageWithText(rc, "header", "footer")
	if !strings.Contains(out, "header") || !strings.Contains(out, "footer") {
		t.Errorf("RenderImageWithText should include header and footer: %q", out)
	}
	// The canvas render should be between them.
	hi := strings.Index(out, "header")
	fi := strings.Index(out, "footer")
	ci := strings.Index(out, "▀")
	if !(hi < ci && ci < fi) {
		t.Errorf("expected header < canvas < footer order: %q", out)
	}
}

func TestRasterImageRows(t *testing.T) {
	// 4 cell rows -> 8 pixels tall
	rc := NewRasterCanvas(10, 4, RasterKitty)
	if rc.ImageRows() != 8 {
		t.Errorf("kitty ImageRows should be pixel height (8), got %d", rc.ImageRows())
	}
	rc.Backend = RasterSixel
	if rc.ImageRows() != 1 { // ceil(8/12) = 1
		t.Errorf("sixel ImageRows should be ceil(h/12)=1, got %d", rc.ImageRows())
	}
	rc.Backend = RasterHalfBlock
	if rc.ImageRows() != 4 {
		t.Errorf("half-block ImageRows should be cell rows (4), got %d", rc.ImageRows())
	}
}

func TestDetectRasterBackendDefault(t *testing.T) {
	// Without any terminal hint, should fall back to HalfBlock.
	// We cannot reset os.Environ in this test (other tests may depend on
	// the real env), but we can at least assert that DefaultRasterBackend
	// is a valid enum value.
	if DefaultRasterBackend < RasterHalfBlock || DefaultRasterBackend > RasterKitty {
		t.Errorf("DefaultRasterBackend out of range: %d", DefaultRasterBackend)
	}
}

func TestRasterCanvasDelegatesSetAndBlit(t *testing.T) {
	green := lipgloss.Color("#00ff00")
	for _, backend := range []RasterBackend{RasterHalfBlock, RasterKitty, RasterSixel} {
		rc := NewRasterCanvas(4, 2, backend)
		rc.Blit([]string{"####"}, map[byte]color.Color{'#': green}, 0, 0)
		// Must not panic.
		_ = rc.RenderImage()
	}
}
