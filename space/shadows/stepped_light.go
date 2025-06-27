//go:build ignore

//kage:unit pixels

package main

// Uniform variables.
var LightPos vec2
var MaxDistance float
var Zoom float // zoom 1 = standard zoom, > 1 = zoomed in, < 1 = zoomed out

// IMPORTANT: Must return premultiplied alpha for proper blending!
func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Get the current pixel position relative to the destination origin
	// then convert from screen‐space to world‐space
	pos := (dstPos.xy - imageDstOrigin())
	// Calculate distance from light position to current pixel (in world units)
	distance := length(pos-LightPos) / Zoom
	// Normalize distance (0.0 to 1.0 based on MaxDistance in world units)
	normalizedDistance := min(distance/MaxDistance, 1.0)
	// Exponential‐spaced steps:
	//   widths are X, 2X, 4X, 8X, 16X  (for STEPS=5)
	//   sum = (2^5 - 1) * X  → X = 1/31
	const STEPS = 5
	const TOTAL_UNITS = 31.0 // 2^5 - 1
	const BASE_STEP = 1.0 / TOTAL_UNITS
	// brightness levels for each band
	var alphas = [STEPS]float{0.8, 0.6, 0.4, 0.25, 0.1}
	// walk outward through the bands
	var threshold = 0.0
	var width = BASE_STEP
	var alpha = 0.0
	for i := 0; i < STEPS; i++ {
		threshold += width
		if normalizedDistance < threshold {
			alpha = alphas[i]
			break
		}
		width *= 2.0
	}
	lightColor := vec3(1.0, 1.0, 1.0)
	return vec4(lightColor*alpha, alpha)
}
