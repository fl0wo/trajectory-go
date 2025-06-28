//go:build ignore
// +build ignore

//kage:unit pixels

package main

// Neutral eye parameters (calm and balanced)
const EyeSeparation = 0.30
const EyeVertical = -0.38
const EyeRadius = 0.060
const PupilRadius = 0.035 // Moderate pupil size for calm look

// Neutral mouth parameters (straight line)
const MouthVertical = -0.50
const MouthRadius = 0.022
const MouthHalfWidth = 0.15 // Half-width of straight mouth line

// Uniforms from Go
var PlayerPos vec2             // [0..1] world-space player center
var WhiteHolePos vec2          // [0..1] world-space white hole center
var WhiteHoleOrbitRadius float // [0..1] orbit radius (for dynamic mouth)
var CameraPos vec2             // [0..1]
var Zoom float
var Radius float // white hole radius in world units [0..1]
var Time float
var ScreenSize vec2 // [width, height] in pixels

// SDF for a circle at c with radius r
func sdfCircle(p, c vec2, r float) float {
	return length(p-c) - r
}

// SDF for a capsule (line segment with rounded ends)
func sdfCapsule(p, a, b vec2, r float) float {
	pa := p - a
	ba := b - a
	h := clamp(dot(pa, ba)/dot(ba, ba), 0.0, 1.0)
	return length(pa-ba*h) - r
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
	return (world - WhiteHolePos) / radiusSafe
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 1) Into local circle coords & flip 180°
	p := toLocal(dstPos)
	p = -p

	// 2) Compute alpha for the circle with anti-aliasing
	distWhiteHole := length(p)
	sdfWhiteHole := distWhiteHole - 1.0
	wWhiteHole := fwidth(sdfWhiteHole)
	aWhiteHole := smoothstep(wWhiteHole, -wWhiteHole, sdfWhiteHole)

	// 3) Early exit if fully outside
	if aWhiteHole < 0.001 {
		return vec4(0.0)
	}

	// 4) Build neutral facial features
	outCol := vec4(0.0)

	// 5) Look-direction axes
	delta := PlayerPos - WhiteHolePos
	dir := vec2(0.0, 1.0)
	if length(delta) > 0.0001 {
		asp := ScreenSize.x / ScreenSize.y
		dir = normalize(vec2(delta.x*asp, delta.y))
	}
	rightDir := vec2(dir.y, -dir.x)
	leftDir := -rightDir

	// 6) Draw neutral eyes (simple circles)
	{
		// Left eye (white sclera)
		leftPos := leftDir*EyeSeparation + dir*EyeVertical
		dL := sdfCircle(p, leftPos, EyeRadius)
		wL := fwidth(dL)
		aL := smoothstep(-wL, wL, -dL) * aWhiteHole
		outCol = mix(outCol, vec4(1.0, 1.0, 1.0, 1.0), aL) // White eyes for neutral look

		// Right eye (white sclera)
		rightPos := rightDir*EyeSeparation + dir*EyeVertical
		dR := sdfCircle(p, rightPos, EyeRadius)
		wR := fwidth(dR)
		aR := smoothstep(-wR, wR, -dR) * aWhiteHole
		outCol = mix(outCol, vec4(1.0, 1.0, 1.0, 1.0), aR) // White eyes for neutral look

		// Moderate black pupils (looking at player)
		leftPupilPos := leftPos + dir*(EyeRadius-PupilRadius)*0.7
		dLP := sdfCircle(p, leftPupilPos, PupilRadius)
		wLP := fwidth(dLP)
		aLP := smoothstep(-wLP, wLP, -dLP) * aWhiteHole
		outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), aLP)

		rightPupilPos := rightPos + dir*(EyeRadius-PupilRadius)*0.7
		dRP := sdfCircle(p, rightPupilPos, PupilRadius)
		wRP := fwidth(dRP)
		aRP := smoothstep(-wRP, wRP, -dRP) * aWhiteHole
		outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), aRP)
	}

	// 7) Dynamic mouth-scaling (neutral expression gets slightly wider when closer)
	playerDist := length(delta)
	mouthScale := 1.0
	if playerDist < WhiteHoleOrbitRadius {
		// 1.0 at orbit edge → 2.5 at center (more subtle than black holes)
		frac := 1.0 - (playerDist / WhiteHoleOrbitRadius)
		mouthScale = 1.0 + frac*1.5
	}
	scaledMouthWidth := MouthHalfWidth * mouthScale

	// 8) Draw neutral straight mouth line
	{
		mouthCenter := dir * MouthVertical

		// Create horizontal line endpoints relative to the face direction
		leftEnd := mouthCenter + leftDir*scaledMouthWidth
		rightEnd := mouthCenter + rightDir*scaledMouthWidth

		// Use capsule SDF for straight line with rounded ends
		dMouth := sdfCapsule(p, leftEnd, rightEnd, MouthRadius)
		wMouth := fwidth(dMouth)
		aMouth := smoothstep(-wMouth, wMouth, -dMouth) * aWhiteHole

		outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), aMouth)
	}

	return outCol
}
