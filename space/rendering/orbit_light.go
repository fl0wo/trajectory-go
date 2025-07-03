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

// Current orbit circle being rendered
var CircleCenter vec2
var CircleRadius float

// All other orbit circles for overlap detection (up to 10 celestial bodies)
var NumOtherOrbits int          // number of other orbits to check (0-10)
var OtherOrbitCenters [20]float // x1,y1, x2,y2, ... (up to 10 orbits)
var OtherOrbitRadii [10]float   // radius1, radius2, ... (up to 10 orbits)

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Get pixel position relative to screen origin (same as shadow system)
	pos := dstPos.xy - imageDstOrigin()

	// Check if pixel is inside any other orbit
	for i := 0; i < 10; i++ {
		if i >= NumOtherOrbits {
			break // No more orbits to check
		}
		// Get the center of the i-th other orbit
		centerX := OtherOrbitCenters[i*2]
		centerY := OtherOrbitCenters[i*2+1]
		otherCenter := vec2(centerX, centerY)
		otherRadius := OtherOrbitRadii[i] * 1 // Reduce radius by 2%

		// Check if the current orbit is completely inside this other orbit
		distanceBetweenCenters := length(CircleCenter - otherCenter)
		if distanceBetweenCenters+CircleRadius <= otherRadius {
			continue // Current orbit is fully inside, skip discard check
		}

		// Calculate distance from current pixel to the center of the other orbit
		distanceToOther := length(pos - otherCenter)

		// If the pixel is inside another orbit's adjusted radius, discard it
		if distanceToOther <= otherRadius {
			discard()
		}
	}

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
		return vec4(1.0-OriginalColor.rgb, OriginalColor.a) // more transparent with zoom
	} else {
		// Outside light cone - use original color
		return OriginalColor
	}
}
