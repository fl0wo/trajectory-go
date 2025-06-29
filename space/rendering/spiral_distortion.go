// shaders/spiral_distortion.kage
//go:build ignore
// +build ignore

//kage:unit pixels
//kage:filter linear
//kage:address clamp-to-edge

package main

var Time float        // (unused here, but you can animate Strength if you like)
var NumBlackHoles int // number of active black holes (0-3)

// Arrays for up to 3 black holes (6 floats for positions, 3 for radii, 3 for strengths)
var BHPositions [6]float    // x1,y1, x2,y2, x3,y3 
var OrbitRadii [3]float     // radius1, radius2, radius3
var Strengths [3]float      // strength1, strength2, strength3

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

	// Apply black hole distortions using constant loop with early break
	// This is more flexible than separate if statements for each black hole
	for i := 0; i < 3; i++ {
		if i >= NumBlackHoles {
			break
		}
		
		// Extract position from flattened array (x,y pairs)
		bhPos := vec2(BHPositions[i*2], BHPositions[i*2+1])
		orbitRadius := OrbitRadii[i]
		strength := Strengths[i]
		
		samplePos = applyBlackHoleDistortion(samplePos, bhPos, orbitRadius, strength)
	}

	// Sample from the final distorted position
	return imageSrc0UnsafeAt(samplePos)
}
