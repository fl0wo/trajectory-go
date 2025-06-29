// shaders/spiral_distortion.kage
//go:build ignore
// +build ignore

//kage:unit pixels
//kage:filter linear
//kage:address clamp-to-edge

package main

var Time float        // (unused here, but you can animate Strength if you like)
var BHPos vec2        // black-hole center in pixels
var OrbitRadius float // effect OrbitRadius in pixels
var Strength float    // [0…1] how strongly to pull at the core (1.0 = full collapse)

func Fragment(dstPos vec4, srcPos vec2, _ vec4) vec4 {
	// original pixel
	col := imageSrc0At(srcPos)

	// vector from center → this pixel
	delta := srcPos - BHPos
	d := length(delta)

	if d < OrbitRadius {
		// falloff: 1.0 at center → 0.0 at OrbitRadius
		falloff := smoothstep(OrbitRadius, 0.0, d)

		// compute how far to pull:
		// `Strength` is a fraction [0…1], so at the center
		// we pull full `Strength`, and at the edge none.
		pullFactor := falloff * Strength

		// new sample position: move srcPos toward BHPos by that fraction
		// i.e. samplePos = mix(srcPos, BHPos, pullFactor)
		samplePos := BHPos + delta*(1.0-pullFactor)

		return imageSrc0UnsafeAt(samplePos)
	}

	// outside the circle → unchanged
	return col
}
