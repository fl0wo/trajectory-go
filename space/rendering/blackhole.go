//go:build ignore
// +build ignore

//kage:unit pixels

package main

// Evil eye parameters
const EyeSeparation = 0.32
const EyeVertical = -0.35
const EyeRadius = 0.070
const PupilRadius = 0.055
const EyeCutOffset = 0.02 // inset so we cut slightly less than half

// Evil mouth parameters
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

	zoomSafe := max(0.01, min(Zoom, 100.0))
	c /= zoomSafe

	world := CameraPos + c
	radiusSafe := max(0.0001, min(Radius, 1.0))
	return (world - BlackHolePos) / radiusSafe
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 1) Into local coords & flip face
	p := toLocal(dstPos)
	p = -p

	// 2) AA‐alpha for the black hole circle
	distBH := length(p)
	sdfBH := distBH - 1.0
	wBH := fwidth(sdfBH)
	aBH := smoothstep(wBH, -wBH, sdfBH)
	if aBH < 0.001 {
		return vec4(0.0)
	}

	// 3) Compute look‐direction axes
	delta := PlayerPos - BlackHolePos
	dir := vec2(0.0, 1.0)
	if length(delta) > 0.0001 {
		asp := ScreenSize.x / ScreenSize.y
		dir = normalize(vec2(delta.x*asp, delta.y))
	}
	rightDir := vec2(dir.y, -dir.x)
	leftDir := -rightDir

	// 4) Are we inside the orbit?
	distWorld := length(delta)
	inOrbit := step(distWorld, BlackHoleOrbitRadius)

	// 5) Face‐rotation matrix (cosφ, sinφ)
	c := dir.y
	s := -dir.x

	// 6) Local face‐space cut normals at ±45°
	const c45 = 0.70710678
	localL := vec2(c45, c45)  // +45° for left
	localR := vec2(-c45, c45) // +135° (−45°) for right

	// 7) Rotate into world‐space, then rotate 180° more by negating
	cutL := vec2(c*localL.x-s*localL.y, s*localL.x+c*localL.y)
	cutR := vec2(c*localR.x-s*localR.y, s*localR.x+c*localR.y)
	cutL = -cutL
	cutR = -cutR

	outCol := vec4(0.0)

	// 8) Left eye (sclera + pupil), bottom‐half kept when angry
	{
		leftPos := leftDir*EyeSeparation + dir*EyeVertical
		qL := p - leftPos

		// sclera
		dL := sdfCircle(p, leftPos, EyeRadius)
		wL := fwidth(dL)
		aL0 := smoothstep(-wL, wL, -dL) * aBH
		maskL := step(-EyeCutOffset, dot(cutL, qL))
		aL := aL0 * (inOrbit*maskL + (1.0 - inOrbit))
		outCol = mix(outCol, vec4(1, 0, 0, 1), aL)

		// pupil
		pupilL := leftPos + dir*(EyeRadius-PupilRadius)*0.6
		dPL := sdfCircle(p, pupilL, PupilRadius)
		wPL := fwidth(dPL)
		aPL0 := smoothstep(-wPL, wPL, -dPL) * aBH
		aPL := aPL0 * (inOrbit*maskL + (1.0 - inOrbit))
		outCol = mix(outCol, vec4(0, 0, 0, 1), aPL)
	}

	// 9) Right eye
	{
		rightPos := rightDir*EyeSeparation + dir*EyeVertical
		qR := p - rightPos

		// sclera
		dR := sdfCircle(p, rightPos, EyeRadius)
		wR := fwidth(dR)
		aR0 := smoothstep(-wR, wR, -dR) * aBH
		maskR := step(-EyeCutOffset, dot(cutR, qR))
		aR := aR0 * (inOrbit*maskR + (1.0 - inOrbit))
		outCol = mix(outCol, vec4(1, 0, 0, 1), aR)

		// pupil
		pupilR := rightPos + dir*(EyeRadius-PupilRadius)*0.6
		dPR := sdfCircle(p, pupilR, PupilRadius)
		wPR := fwidth(dPR)
		aPR0 := smoothstep(-wPR, wPR, -dPR) * aBH
		aPR := aPR0 * (inOrbit*maskR + (1.0 - inOrbit))
		outCol = mix(outCol, vec4(0, 0, 0, 1), aPR)
	}

	// 10) Dynamic frown mouth (unchanged)
	{
		frac := clamp(1.0-distWorld/BlackHoleOrbitRadius, 0.0, 1.0)
		mouthScale := 1.0 + frac*5.0
		scaledR := MouthArcRadius * mouthScale

		arcCenter := dir * MouthVertical
		q := p - arcCenter
		d := length(q)

		sdfArc := abs(d-scaledR) - MouthRadius
		wM := fwidth(sdfArc)
		mThick := smoothstep(-wM, wM, -sdfArc)

		nq := q / max(d, 0.0001)
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
