//go:build ignore

//kage:unit pixels

package main

// Uniform variables.
var LightPos vec2
var MaxDistance float

// Fragment is the entry point of the fragment shader.
// Fragment returns the color value for the current position.
// IMPORTANT: Must return premultiplied alpha for proper blending!
func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Get the current pixel position relative to the destination origin
	pos := dstPos.xy - imageDstOrigin()

	// Calculate distance from light position to current pixel
	distance := length(pos - LightPos)

	// Normalize distance (0.0 to 1.0 based on MaxDistance)
	normalizedDistance := min(distance/MaxDistance, 1.0)

	// Exponential‐spaced steps:
	//   widths are X, 2X, 4X, 8X, 16X  (for STEPS=5)
	//   sum = (2^5 - 1) * X  → we choose X so that sum==1.0 → X = 1/31
	const STEPS = 5
	const TOTAL_UNITS = 31.0 // (2^5 - 1)
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
	// beyond threshold = fully dark (alpha==0)

	// Light color (white)
	lightColor := vec3(1.0, 1.0, 1.0)

	// CRUCIAL: Return premultiplied alpha (RGB * alpha, alpha)
	return vec4(lightColor*alpha, alpha)
}
