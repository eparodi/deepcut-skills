package main

import (
	"bytes"
	"image"
	"image/png"
)

// maxVisionImageDimension is the DeepSeek vision API's per-side limit
// (verified 2026-08-27: 1125x17679 → 400 "unsupported image"; the docs pin
// "Max image dimension | 8192 px per side"). Full-page captures are
// downscaled to fit before the request is sent.
const maxVisionImageDimension = 8192

// fitMaxDimension downscales a PNG (nearest-neighbor, stdlib only) so neither
// side exceeds max. Images already within the limit are returned unchanged.
// Needed because full-page captures at mobile DPR 3 routinely exceed 8192px.
func fitMaxDimension(pngData []byte, maxDim int) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxDim && h <= maxDim {
		return pngData, nil
	}
	scale := float64(maxDim) / float64(max(w, h))
	nw := int(float64(w) * scale)
	nh := int(float64(h) * scale)
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		sy := b.Min.Y + y*h/nh
		for x := 0; x < nw; x++ {
			sx := b.Min.X + x*w/nw
			dst.Set(x, y, img.At(sx, sy))
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
