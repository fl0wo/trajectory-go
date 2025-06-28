//go:build ignore
// +build ignore

//kage:unit pixels

package main

// Evil eye parameters (angular and menacing)
const EyeSeparation = 0.32
const EyeVertical = -0.35
const EyeRadius = 0.070
const PupilRadius = 0.055  // Larger pupils for more intense look

// Evil mouth parameters (frown instead of smile)
const MouthVertical = -0.52
const MouthRadius = 0.028
const MouthArcRadius = 0.048
const MouthArcCos = 0.15  // Wider frown arc

// Uniforms from Go
var PlayerPos vec2          // [0..1] world-space player center
var BlackHolePos vec2       // [0..1] world-space black hole center
var BlackHoleOrbitRadius float // [0..1] orbit radius (for dynamic mouth)
var CameraPos vec2          // [0..1]
var Zoom float
var Radius float            // black hole radius in world units [0..1]
var Time float
var ScreenSize vec2         // [width, height] in pixels

// SDF for a circle at c with radius r
func sdfCircle(p, c vec2, r float) float {
	return length(p-c) - r
}

// Exact coordinate mapping as the working planet shader
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
	return (world - BlackHolePos) / radiusSafe
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 1) Into local circle coords & flip 180°
	p := toLocal(dstPos)
	p = -p

	// 2) Compute alpha for the circle with anti-aliasing
	distBlackHole := length(p)
	sdfBlackHole := distBlackHole - 1.0
	wBlackHole := fwidth(sdfBlackHole)
	aBlackHole := smoothstep(wBlackHole, -wBlackHole, sdfBlackHole)

	// 3) Early exit if fully outside
	if aBlackHole < 0.001 {
		return vec4(0.0)
	}

	// 4) Build evil facial features
	outCol := vec4(0.0)

	// 5) Look-direction axes
	delta := PlayerPos - BlackHolePos
	dir := vec2(0.0, 1.0)
	if length(delta) > 0.0001 {
		asp := ScreenSize.x / ScreenSize.y
		dir = normalize(vec2(delta.x*asp, delta.y))
	}
	rightDir := vec2(dir.y, -dir.x)
	leftDir := -rightDir

	// 6) Draw evil eyes (angular and intense)
	{
		// Left eye (white sclera)
		leftPos := leftDir*EyeSeparation + dir*EyeVertical
		dL := sdfCircle(p, leftPos, EyeRadius)
		wL := fwidth(dL)
		aL := smoothstep(-wL, wL, -dL) * aBlackHole
		outCol = mix(outCol, vec4(1.0, 0.0, 0.0, 1.0), aL) // Red eyes for evil effect

		// Right eye (white sclera)
		rightPos := rightDir*EyeSeparation + dir*EyeVertical
		dR := sdfCircle(p, rightPos, EyeRadius)
		wR := fwidth(dR)
		aR := smoothstep(-wR, wR, -dR) * aBlackHole
		outCol = mix(outCol, vec4(1.0, 0.0, 0.0, 1.0), aR) // Red eyes for evil effect

		// Large black pupils for intensity
		leftPupilPos := leftPos + dir * (EyeRadius - PupilRadius) * 0.6
		dLP := sdfCircle(p, leftPupilPos, PupilRadius)
		wLP := fwidth(dLP)
		aLP := smoothstep(-wLP, wLP, -dLP) * aBlackHole
		outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), aLP)

		rightPupilPos := rightPos + dir * (EyeRadius - PupilRadius) * 0.6
		dRP := sdfCircle(p, rightPupilPos, PupilRadius)
		wRP := fwidth(dRP)
		aRP := smoothstep(-wRP, wRP, -dRP) * aBlackHole
		outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), aRP)
	}

	// 7) Dynamic mouth-scaling (gets more evil when player is closer)
	playerDist := length(delta)
	mouthScale := 1.0
	if playerDist < BlackHoleOrbitRadius {
		// 1.0 at orbit edge → 6.0 at center (more dramatic than planets)
		frac := 1.0 - (playerDist / BlackHoleOrbitRadius)
		mouthScale = 1.0 + frac*5.0
	}
	scaledArcR := MouthArcRadius * mouthScale

	// 8) Draw evil frown (inverted smile)
	{
		arcCenter := dir * MouthVertical
		q := p - arcCenter
		d := length(q)

		// arc strip
		sdfArc := abs(d-scaledArcR) - MouthRadius
		wM := fwidth(sdfArc)
		mThick := smoothstep(-wM, wM, -sdfArc)

		// angular mask for frown (using positive dir for downward curve)
		nq := q / max(d, 0.0001)
		mAngle := step(MouthArcCos, dot(nq, dir)) // Note: using positive dir for frown
		mArc := mThick * mAngle * aBlackHole

		// end-caps
		sinArc := sqrt(max(0.0, 1.0-MouthArcCos*MouthArcCos))

		vL := dir*MouthArcCos + leftDir*sinArc  // Flipped for frown
		epL := arcCenter + vL*scaledArcR
		dE1 := sdfCircle(p, epL, MouthRadius)
		wE1 := fwidth(dE1)
		aE1 := smoothstep(-wE1, wE1, -dE1) * aBlackHole

		vR := dir*MouthArcCos + rightDir*sinArc  // Flipped for frown
		epR := arcCenter + vR*scaledArcR
		dE2 := sdfCircle(p, epR, MouthRadius)
		wE2 := fwidth(dE2)
		aE2 := smoothstep(-wE2, wE2, -dE2) * aBlackHole

		mMask := max(mArc, max(aE1, aE2))
		outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), mMask)
	}

	return outCol
}