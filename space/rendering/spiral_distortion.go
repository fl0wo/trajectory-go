// shaders/spiral_distortion.kage
//go:build ignore
// +build ignore

//kage:unit pixels
//kage:filter linear
//kage:address clamp-to-edge

package main

var Time float // (unused here, but you can animate Strength if you like)

// Single black hole uniforms
var BH_Pos vec2    // Black hole position
var BH_Radius float // Black hole orbit radius
var BH_Strength float // Black hole strength

func applyBlackHoleDistortion(srcPos vec2, bhPos vec2, orbitRadius float, strength float) vec2 {
	// vector from center → this pixel
	delta := srcPos - bhPos
	d := length(delta)

	if d < orbitRadius && d > 0.0 {
		// falloff: 1.0 at center → 0.0 at orbitRadius
		falloff := smoothstep(orbitRadius, 0.0, d)

		// compute how far to pull:
		// `strength` is a fraction [0…1], so at the center
		// we pull full `strength`, and at the edge none.
		pullFactor := falloff * strength

		// new sample position: move srcPos toward bhPos by that fraction
		// i.e. samplePos = mix(srcPos, bhPos, pullFactor)
		return bhPos + delta*(1.0-pullFactor)
	}

	// outside the circle → unchanged
	return srcPos
}

func Fragment(dstPos vec4, srcPos vec2, _ vec4) vec4 {
	// Apply single black hole distortion
	samplePos := applyBlackHoleDistortion(srcPos, BH_Pos, BH_Radius, BH_Strength)

	// Sample from the distorted position
	return imageSrc0UnsafeAt(samplePos)
}
