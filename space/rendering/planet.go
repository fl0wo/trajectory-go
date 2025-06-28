//go:build ignore
// +build ignore

//kage:unit pixels

package main

// Eye parameters (in circle-local units)
const EyeSeparation = 0.28
const EyeVertical = -0.4
const EyeRadius = 0.064

// Mouth parameters
const MouthVertical = -0.48
const MouthRadius = 0.026
const MouthArcRadius = 0.045
const MouthArcCos = 0.2

// Uniforms from Go
var PlayerPos vec2          // [0..1] world-space player center
var PlanetPos vec2          // [0..1] world-space planet center
var PlanetOrbitRadius float // [0..1] orbit radius (for dynamic mouth)
var CameraPos vec2          // [0..1]
var Zoom float
var Radius float // planet radius in world units [0..1]
var Time float
var ScreenSize vec2 // [width, height] in pixels

// SDF for a circle at c with radius r
func sdfCircle(p, c vec2, r float) float {
	return length(p-c) - r
}

// Exactly the same mapping your vector circle uses:
func toLocal(dstPos vec4) vec2 {
	uv := dstPos.xy / ScreenSize
	c := uv - vec2(0.5)
	asp := ScreenSize.x / ScreenSize.y
	c.x *= asp

	zoomSafe := max(0.01, min(Zoom, 100.0))
	c /= zoomSafe

	// world-space
	world := CameraPos + c
	radiusSafe := max(0.0001, min(Radius, 1.0))
	return (world - PlanetPos) / radiusSafe
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 1) Into local circle coords & flip 180°
	p := toLocal(dstPos)
	p = -p

	// 2) Compute aPlanet = 1 inside your white circle, 0 outside, anti-aliased
	distPlanet := length(p)
	sdfPlanet := distPlanet - 1.0
	wPlanet := fwidth(sdfPlanet)
	aPlanet := smoothstep(wPlanet, -wPlanet, sdfPlanet)

	// 3) If we're fully outside the circle + its AA band, draw nothing
	if aPlanet < 0.001 {
		return vec4(0.0)
	}

	// 4) Build our feature-mask‐gated shader
	outCol := vec4(0.0)

	// 5) Look‐direction axes
	delta := PlayerPos - PlanetPos
	dir := vec2(0.0, 1.0)
	if length(delta) > 0.0001 {
		asp := ScreenSize.x / ScreenSize.y
		dir = normalize(vec2(delta.x*asp, delta.y))
	}
	rightDir := vec2(dir.y, -dir.x)
	leftDir := -rightDir

	// 6) Draw eyes (pixel‐aware AA *inside* the circle)
	{
		// left
		leftPos := leftDir*EyeSeparation + dir*EyeVertical
		dL := sdfCircle(p, leftPos, EyeRadius)
		wL := fwidth(dL)
		aL := smoothstep(-wL, wL, -dL) * aPlanet
		outCol = mix(outCol, vec4(0, 0, 0, 1), aL)

		// right
		rightPos := rightDir*EyeSeparation + dir*EyeVertical
		dR := sdfCircle(p, rightPos, EyeRadius)
		wR := fwidth(dR)
		aR := smoothstep(-wR, wR, -dR) * aPlanet
		outCol = mix(outCol, vec4(0, 0, 0, 1), aR)
	}

	// 7) Dynamic mouth‐scaling
	playerDist := length(delta)
	mouthScale := 1.0
	if playerDist < PlanetOrbitRadius {
		// 1.0 at orbit edge → 5.0 at center
		frac := 1.0 - (playerDist / PlanetOrbitRadius)
		mouthScale = 1.0 + frac*4.0
	}
	scaledArcR := MouthArcRadius * mouthScale

	// 8) Draw the capsule‐mouth (AA) *inside* the circle
	{
		arcCenter := dir * MouthVertical
		q := p - arcCenter
		d := length(q)

		// arc strip
		sdfArc := abs(d-scaledArcR) - MouthRadius
		wM := fwidth(sdfArc)
		mThick := smoothstep(-wM, wM, -sdfArc)

		// angular mask
		nq := q / max(d, 0.0001)
		mAngle := step(MouthArcCos, dot(nq, -dir))
		mArc := mThick * mAngle * aPlanet

		// end-caps
		sinArc := sqrt(max(0.0, 1.0-MouthArcCos*MouthArcCos))

		vL := -dir*MouthArcCos + leftDir*sinArc
		epL := arcCenter + vL*scaledArcR
		dE1 := sdfCircle(p, epL, MouthRadius)
		wE1 := fwidth(dE1)
		aE1 := smoothstep(-wE1, wE1, -dE1) * aPlanet

		vR := -dir*MouthArcCos + rightDir*sinArc
		epR := arcCenter + vR*scaledArcR
		dE2 := sdfCircle(p, epR, MouthRadius)
		wE2 := fwidth(dE2)
		aE2 := smoothstep(-wE2, wE2, -dE2) * aPlanet

		mMask := max(mArc, max(aE1, aE2))
		outCol = mix(outCol, vec4(0, 0, 0, 1), mMask)
	}

	return outCol
}
