package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func testPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 100, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func decodeSize(t *testing.T, data []byte) (int, int) {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	b := img.Bounds()
	return b.Dx(), b.Dy()
}

func TestFitMaxDimension(t *testing.T) {
	tests := []struct {
		name  string
		w, h  int
		max   int
		wantW int
		wantH int
	}{
		{"under cap untouched", 100, 50, 8192, 100, 50},
		{"exact cap untouched", 8192, 1000, 8192, 8192, 1000},
		{"tiny max clamps to 1px", 100, 100, 1, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := testPNG(t, tt.w, tt.h)
			got, err := fitMaxDimension(in, tt.max)
			if err != nil {
				t.Fatalf("fitMaxDimension: %v", err)
			}
			w, h := decodeSize(t, got)
			if w != tt.wantW || h != tt.wantH {
				t.Errorf("size = %dx%d, want %dx%d", w, h, tt.wantW, tt.wantH)
			}
		})
	}

	// Downscaled cases: the contract is (a) no side exceeds max, (b) the
	// aspect ratio is preserved, (c) the longer side lands exactly on max.
	downscaled := []struct {
		name string
		w, h int
		max  int
	}{
		{"tall page downscaled", 1125, 17679, 8192},
		{"wide page downscaled", 20000, 500, 8192},
	}
	for _, tt := range downscaled {
		t.Run(tt.name, func(t *testing.T) {
			in := testPNG(t, tt.w, tt.h)
			got, err := fitMaxDimension(in, tt.max)
			if err != nil {
				t.Fatalf("fitMaxDimension: %v", err)
			}
			w, h := decodeSize(t, got)
			if w > tt.max || h > tt.max || w < 1 || h < 1 {
				t.Fatalf("size = %dx%d, must stay within 1..%d", w, h, tt.max)
			}
			if w != tt.max && h != tt.max {
				t.Errorf("size = %dx%d, longer side must land on %d", w, h, tt.max)
			}
			wantRatio := float64(tt.w) / float64(tt.h)
			gotRatio := float64(w) / float64(h)
			// ±1px truncation on a short side (e.g. 204px) is ~0.5%; allow 1%.
			if rel := (gotRatio - wantRatio) / wantRatio; rel < -0.01 || rel > 0.01 {
				t.Errorf("aspect ratio = %.4f, want ~%.4f (rel err %.4f)", gotRatio, wantRatio, rel)
			}
		})
	}
}

func TestFitMaxDimensionInvalidPNG(t *testing.T) {
	if _, err := fitMaxDimension([]byte("not a png"), 8192); err == nil {
		t.Fatal("fitMaxDimension accepted invalid PNG, want error")
	}
}
