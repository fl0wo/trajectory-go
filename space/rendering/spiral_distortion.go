// shaders/spiral_distortion.kage
//go:build ignore
// +build ignore

// kage:unit pixels
// kage:filter linear
// kage:address clamp-to-edge

package main

var Time float             // seconds since start
var ScreenSize vec2        // [width, height] in pixels
var CameraPos vec2         // world-space center in normalized UV [0..1]
var Zoom float             // camera zoom
var BlackHolePosition vec2 // world-space UV of black hole center
var OrbitRadius float      // world-space UV radius
var Strength float         // max radians of twist at center

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Get pixel position relative to destination origin
	pos := dstPos.xy - imageDstOrigin()

	// Convert to normalized UV coordinates
	uv := pos / ScreenSize

	// Transform to world space with camera and zoom
	c := uv - vec2(0.5, 0.5)
	aspect := ScreenSize.x / ScreenSize.y
	c.x *= aspect
	z := max(0.01, min(Zoom, 100.0))
	c /= z
	worldPos := CameraPos + c

	// Calculate distance from black hole
	delta := worldPos - BlackHolePosition
	d := length(delta)

	// Apply spiral distortion if within orbit radius
	if d > 0.0001 && d < OrbitRadius {
		// Smooth falloff from center to edge
		falloff := smoothstep(OrbitRadius, 0.0, d)

		// Create spiral rotation with time animation
		angle := falloff*Strength + Time*0.3
		s := sin(angle)
		co := cos(angle)

		// Rotate the delta vector
		rotatedDelta := vec2(
			delta.x*co-delta.y*s,
			delta.x*s+delta.y*co,
		)

		// Transform back to UV space
		worldPos = BlackHolePosition + rotatedDelta
		c = (worldPos - CameraPos) * z
		c.x /= aspect
		distortedUV := vec2(0.5, 0.5) + c

		// Check bounds and sample
		if distortedUV.x >= 0.0 && distortedUV.x <= 1.0 &&
			distortedUV.y >= 0.0 && distortedUV.y <= 1.0 {
			samplePos := distortedUV*ScreenSize + imageSrc0Origin()
			return vec4(imageSrc0At(samplePos).rgb, 1.0)
		}
	}

	// Outside orbit or out of bounds - sample original position
	return vec4(imageSrc0At(srcPos).rgb, 1.0)
}
