//go:build ignore
// +build ignore

//kage:unit pixels

package main

// Base eye parameters
const EyeSeparation = 0.32
const EyeVertical = -0.35
const EyeRadius = 0.070
const PupilRadius = 0.055

// How much bigger when “evil” (inside orbit)
const EyeRadiusScaleFactor = 0.5   // +50%
const PupilRadiusScaleFactor = 0.2 // +20%
// Inset so we cut slightly less than half
const EyeCutOffset = 0.016

// Evil mouth parameters (unchanged)
const MouthVertical = -0.52
const MouthRadius = 0.028
const MouthArcRadius = 0.048
const MouthArcCos = 0.15

// Uniforms
var PlayerPos vec2
var BlackHolePos vec2
var BlackHoleOrbitRadius float
var CameraPos vec2
var Zoom float
var Radius float
var Time float
var ScreenSize vec2

// sdfCircle: distance field for circle at c with radius r
func sdfCircle(p, c vec2, r float) float {
	return length(p-c) - r
}

// toLocal: map screen pixel → unit‐circle coords around black hole
func toLocal(dstPos vec4) vec2 {
	uv := dstPos.xy / ScreenSize
	c := uv - vec2(0.5)
	asp := ScreenSize.x / ScreenSize.y
	c.x *= asp
	zoom := max(0.01, min(Zoom, 100.0))
	c /= zoom
	world := CameraPos + c
	rSafe := max(0.0001, min(Radius, 1.0))
	return (world - BlackHolePos) / rSafe
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 1) Into local coords & flip 180°
	p := toLocal(dstPos)
	p = -p

	// 2) AA‐alpha for the black‐hole circle
	dBH := length(p)
	sdfBH := dBH - 1.0
	wBH := fwidth(sdfBH)
	aBH := smoothstep(wBH, -wBH, sdfBH)
	if aBH < 0.001 {
		return vec4(0.0) // fully transparent outside
	}

	// 3) Vector to player + distance
	delta := PlayerPos - BlackHolePos
	distWorld := length(delta)

	// 4) Compute look‐direction axes
	dir := vec2(0.0, 1.0)
	if distWorld > 0.0001 {
		asp := ScreenSize.x / ScreenSize.y
		dir = normalize(vec2(delta.x*asp, delta.y))
	}
	rightDir := vec2(dir.y, -dir.x)
	leftDir := -rightDir

	// 5) Instant‐evil flag
	inOrbit := step(distWorld, BlackHoleOrbitRadius)

	// 6) Dynamic radii
	eyeRad := EyeRadius * (1.0 + inOrbit*EyeRadiusScaleFactor)
	pupRad := PupilRadius * (1.0 + inOrbit*PupilRadiusScaleFactor)

	// 7) Face‐rotation matrix coeffs
	c := dir.y
	s := -dir.x

	// 8) Local cut‐normals (baked with 180° inversion)
	const c45 = 0.70710678
	localL := vec2(-c45, -c45) // left eye
	localR := vec2(c45, -c45)  // right eye

	// 9) Rotate into world‐space
	cutL := vec2(c*localL.x-s*localL.y, s*localL.x+c*localL.y)
	cutR := vec2(c*localR.x-s*localR.y, s*localR.x+c*localR.y)

	// 10) Draw features
	outCol := vec4(0.0)

	// LEFT EYE
	{
		eyePos := leftDir*EyeSeparation + dir*EyeVertical
		q := p - eyePos

		// sclera
		dE := sdfCircle(p, eyePos, eyeRad)
		wE := fwidth(dE)
		aE0 := smoothstep(-wE, wE, -dE) * aBH

		// apply cut instantly
		mask := step(-EyeCutOffset, dot(cutL, q))
		aE := aE0 * (inOrbit*mask + (1.0 - inOrbit))
		outCol = mix(outCol, vec4(1, 0, 0, 1), aE)

		// pupil
		pupilPos := eyePos + dir*(eyeRad-pupRad)*0.6
		dP := sdfCircle(p, pupilPos, pupRad)
		wP := fwidth(dP)
		aP0 := smoothstep(-wP, wP, -dP) * aBH
		aP := aP0 * (inOrbit*mask + (1.0 - inOrbit))
		outCol = mix(outCol, vec4(0, 0, 0, 1), aP)
	}

	// RIGHT EYE
	{
		eyePos := rightDir*EyeSeparation + dir*EyeVertical
		q := p - eyePos

		dE := sdfCircle(p, eyePos, eyeRad)
		wE := fwidth(dE)
		aE0 := smoothstep(-wE, wE, -dE) * aBH
		mask := step(-EyeCutOffset, dot(cutR, q))
		aE := aE0 * (inOrbit*mask + (1.0 - inOrbit))
		outCol = mix(outCol, vec4(1, 0, 0, 1), aE)

		pupilPos := eyePos + dir*(eyeRad-pupRad)*0.6
		dP := sdfCircle(p, pupilPos, pupRad)
		wP := fwidth(dP)
		aP0 := smoothstep(-wP, wP, -dP) * aBH
		aP := aP0 * (inOrbit*mask + (1.0 - inOrbit))
		outCol = mix(outCol, vec4(0, 0, 0, 1), aP)
	}

	// MOUTH (dynamic frown, unchanged)
	{
		frac := clamp((BlackHoleOrbitRadius-distWorld)/BlackHoleOrbitRadius, 0.0, 1.0)
		mouthScale := 1.0 + frac*5.0
		scaledR := MouthArcRadius * mouthScale

		arcCenter := dir * MouthVertical
		qM := p - arcCenter
		dM := length(qM)

		sdfArc := abs(dM-scaledR) - MouthRadius
		wM := fwidth(sdfArc)
		mThick := smoothstep(-wM, wM, -sdfArc)

		nq := qM / max(dM, 0.0001)
		mAngle := step(MouthArcCos, dot(nq, dir))
		mArc := mThick * mAngle * aBH

		sinA := sqrt(max(0.0, 1.0-MouthArcCos*MouthArcCos))
		vL := dir*MouthArcCos + leftDir*sinA
		epL := arcCenter + vL*scaledR
		dE1 := sdfCircle(p, epL, MouthRadius)
		wE1 := fwidth(dE1)
		aE1 := smoothstep(-wE1, wE1, -dE1) * aBH

		vR := dir*MouthArcCos + rightDir*sinA
		epR := arcCenter + vR*scaledR
		dE2 := sdfCircle(p, epR, MouthRadius)
		wE2 := fwidth(dE2)
		aE2 := smoothstep(-wE2, wE2, -dE2) * aBH

		mMask := max(mArc, max(aE1, aE2))
		outCol = mix(outCol, vec4(0, 0, 0, 1), mMask)
	}

	return outCol
}
