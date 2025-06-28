//go:build ignore

//kage:unit pixels

package main

// Uniforms
var LightPos vec2
var LightDirection vec2
var FOVAngle float     // FOV in radians
var MaxDistance float  // Maximum light distance
var Zoom float         // Camera zoom factor
var OriginalColor vec4 // The original color of the dash

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Get pixel position relative to screen origin (same as shadow system)
	pos := dstPos.xy - imageDstOrigin()

	// Vector from light position to current pixel
	lightToPixel := pos - LightPos

	// Calculate normalized distance (same as shadow system)
	dist := length(lightToPixel) / Zoom
	normalizedDistance := min(dist/(MaxDistance/3), 1.0)

	// First check: Is pixel within the maximum light distance?
	if normalizedDistance >= 1.0 {
		// Outside light range - use original color
		return OriginalColor
	}

	// Normalize the light direction vector
	lightDirLength := length(LightDirection)
	if lightDirLength == 0.0 {
		// If no direction, just return original color
		return OriginalColor
	}
	normalizedLightDir := LightDirection / lightDirLength

	// Calculate the angle between light direction and light-to-pixel vector
	lightToPixelLength := length(lightToPixel)
	if lightToPixelLength == 0.0 {
		// At light source, definitely in cone and within range
		return vec4(1.0-OriginalColor.rgb, OriginalColor.a)
	}

	normalizedLightToPixel := lightToPixel / lightToPixelLength

	// Calculate angle between vectors using dot product
	dotProduct := dot(normalizedLightDir, normalizedLightToPixel)
	// Clamp to avoid numerical errors
	dotProduct = clamp(dotProduct, -1.0, 1.0)
	angle := acos(dotProduct)

	// Second check: Is pixel within the light cone (half FOV)?
	halfFOV := FOVAngle / 2.0
	if angle <= halfFOV {
		// Inside light cone AND within range - invert the color
		return vec4(1.0-OriginalColor.rgb, OriginalColor.a/Zoom) // more transparent with zoom
	} else {
		// Outside light cone - use original color
		return OriginalColor
	}
}
