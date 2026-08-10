package wallpaper

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeAndScale(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.png")

	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			src.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := png.Encode(f, src); err != nil {
		_ = f.Close()
		t.Fatalf("failed to encode png: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	scaled, err := DecodeAndScale(filePath, 20, 10)
	if err != nil {
		t.Fatalf("DecodeAndScale failed: %v", err)
	}
	if scaled.Bounds().Dx() != 20 || scaled.Bounds().Dy() != 10 {
		t.Errorf("scaled dimensions = %dx%d, want 20x10", scaled.Bounds().Dx(), scaled.Bounds().Dy())
	}
}
