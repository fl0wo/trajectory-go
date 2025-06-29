// shaders/spiral_distortion.kage
//go:build ignore
// +build ignore

//kage:unit pixels
//kage:filter linear
//kage:address clamp-to-edge

package main

var Time float        // (unused here, but you can animate Strength if you like)
var NumBlackHoles int // number of active black holes (0-3)

// Black hole 1
var BHPos1 vec2        // black-hole center in pixels
var OrbitRadius1 float // effect OrbitRadius in pixels
var Strength1 float    // [0…1] how strongly to pull at the core (1.0 = full collapse)

// Black hole 2
var BHPos2 vec2        // black-hole center in pixels
var OrbitRadius2 float // effect OrbitRadius in pixels
var Strength2 float    // [0…1] how strongly to pull at the core (1.0 = full collapse)

// Black hole 3
var BHPos3 vec2        // black-hole center in pixels
var OrbitRadius3 float // effect OrbitRadius in pixels
var Strength3 float    // [0…1] how strongly to pull at the core (1.0 = full collapse)

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
	// Start with the original position
	samplePos := srcPos

	// Apply up to 3 black hole distortions sequentially
	if NumBlackHoles >= 1 {
		samplePos = applyBlackHoleDistortion(samplePos, BHPos1, OrbitRadius1, Strength1)
	}
	if NumBlackHoles >= 2 {
		samplePos = applyBlackHoleDistortion(samplePos, BHPos2, OrbitRadius2, Strength2)
	}
	if NumBlackHoles >= 3 {
		samplePos = applyBlackHoleDistortion(samplePos, BHPos3, OrbitRadius3, Strength3)
	}

	// Sample from the final distorted position
	return imageSrc0UnsafeAt(samplePos)
}
