//go:build ignore
// +build ignore

//kage:unit pixels

package main

// Portal distortion parameters
const EdgeSoftness = 0.5    // Fraction of radius for soft mask
const SwirlStrength = 0.1   // radians of swirl at the rim
const RadialWarpPower = 0.6 // <1 warps more strongly inward

// Uniforms injected from Go
var Portal_Pos vec2
var Portal_Radius float
var Portal_Color vec3
var CameraPos vec2
var Zoom float
var Time float
var ScreenSize vec2
var IsActive float

func worldToScreen(wp vec2) vec2 {
	rel := wp - CameraPos
	zs := max(0.01, min(Zoom, 100.0))
	rel *= zs
	rel.x /= (ScreenSize.x / ScreenSize.y)
	return (rel + vec2(0.5, 0.5)) * ScreenSize
}

func screenToWorld(sp vec2) vec2 {
	uv := sp/ScreenSize - vec2(0.5, 0.5)
	uv.x *= (ScreenSize.x / ScreenSize.y)
	uv /= max(0.01, min(Zoom, 100.0))
	return CameraPos + uv
}

// Portal distortion helper for a “wormhole” effect:
func distortUV(pix vec2, center vec2, radius float) vec2 {
	// vector from portal center to this pixel
	off := pix - center
	d := length(off)
	// if outside the portal, don’t change
	if d > radius {
		return pix
	}

	// normalized distance [0=center → 1=edge]
	t := d / radius

	// swirl: strongest at the rim, none at the center
	// (SwirlStrength should be declared as a const float)
	angle := SwirlStrength * (1.0 - t) * IsActive

	// radial warp: compress outer pixels toward the center
	// (RadialWarpPower should be a const float <1 for aggressive pull)
	warped := pow(t, RadialWarpPower)
	newD := warped * radius

	// reconstruct the offset with swirl + radial warp
	a := atan2(off.y, off.x) + angle
	dir := vec2(cos(a), sin(a))

	return center + dir*newD
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// mask in world‐space
	wp := screenToWorld(dstPos.xy)
	dW := length(wp - Portal_Pos)
	if dW > Portal_Radius {
		return imageSrc0At(srcPos)
	}

	// exact pixel‐space center & radius
	cPx := worldToScreen(Portal_Pos)
	ePx := worldToScreen(Portal_Pos + vec2(Portal_Radius, 0))
	rPx := length(ePx - cPx)

	// distort & sample directly in pixels
	dp := distortUV(dstPos.xy, cPx, rPx)
	scene := imageSrc0At(dp)

	// soft mask & tint
	mask := 1.0 - smoothstep(
		Portal_Radius*(1.0-EdgeSoftness),
		Portal_Radius,
		dW,
	)
	tint := vec4(Portal_Color, 1.0) * 0.2 * IsActive

	return mix(scene, scene+tint, mask)
}
