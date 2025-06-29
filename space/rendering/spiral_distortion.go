// shaders/spiral_distortion.kage
//go:build ignore
// +build ignore

// **pixel mode**: dstPos and srcPos are in raw pixels
// address clamp-to-edge so unsafe sampler will clamp
// filter linear for smooth interpolation
// unit pixels
// kage:unit pixels
// kage:filter linear
// kage:address clamp-to-edge

package main

var Time float     // seconds since start
var BHPos vec2     // black-hole center, in pixels
var Radius float   // effect radius, in pixels
var Strength float // max twist (radians) at r=0

func Fragment(dstPos vec4, srcPos vec2, _ vec4) vec4 {
	// Get the original pixel color
	originalColor := imageSrc0At(srcPos)
	
	// Apply a red tint to the entire screen for testing
	// Increase red channel by 0.3, keep other channels as is
	return vec4(originalColor.r + 0.3, originalColor.g, originalColor.b, originalColor.a)
}
