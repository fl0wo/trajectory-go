package util

import "image/color"

// a function that gets 4 ints [0-255] and returns a color.Color
func RGBAColor(r, g, b, a int) color.RGBA {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 || a < 0 || a > 255 {
		panic("RGBAColor: values must be in range [0-255]")
	}

	alpha := float32(a) / 255.0

	return color.RGBA{
		R: uint8(float32(r) * alpha),
		G: uint8(float32(g) * alpha),
		B: uint8(float32(b) * alpha),
		A: uint8(a),
	}
}
