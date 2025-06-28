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

// Map dstPos→circle-local coordinates
func toLocal(dstPos vec4) vec2 {
	uv := dstPos.xy / ScreenSize
	c := uv - vec2(0.5)
	asp := ScreenSize.x / ScreenSize.y
	c.x *= asp

	zoomSafe := max(0.01, min(Zoom, 100.0))
	c /= zoomSafe

	world := CameraPos + c
	radiusSafe := max(0.0001, min(Radius, 1.0))
	return (world - PlanetPos) / radiusSafe
}

// ─── Bigger metaballs ────────────────────────────────────────────────
func metaballs(uv vec2) float {
	sum := 0.0
	const N = 10
	for i := 0; i < N; i++ {
		phase := Time*(0.2+float(i)*0.08) + float(i)*1.7
		c := vec2(cos(phase), sin(phase)) * 0.5
		r := 0.22 + 0.1*sin(Time*1.2+float(i)*2.3)
		sum += r / length(uv-c)
	}
	return sum
}

// Lift to hemisphere & rotate in 3D, project back to XY
func rotatedSphereUV(uv vec2) vec2 {
	d2 := dot(uv, uv)
	z := 0.0
	if d2 < 1.0 {
		z = sqrt(1.0 - d2)
	}
	p3 := vec3(uv.x, uv.y, z)

	ax := Time * 0.4
	ay := Time * 0.25

	rotX := mat3(
		1, 0, 0,
		0, cos(ax), -sin(ax),
		0, sin(ax), cos(ax),
	)
	rotY := mat3(
		cos(ay), 0, sin(ay),
		0, 1, 0,
		-sin(ay), 0, cos(ay),
	)

	rp := rotY * rotX * p3
	return rp.xy
}

// ────────────────────────────────────────────────────────────────────────

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 1) Into circle coords & flip
	p := toLocal(dstPos)
	p = -p

	// 2) Circle mask
	distPlanet := length(p)
	sdfPlanet := distPlanet - 1.0
	wPlanet := fwidth(sdfPlanet)
	aPlanet := smoothstep(wPlanet, -wPlanet, sdfPlanet)
	if aPlanet < 0.001 {
		return vec4(0.0) // outside
	}

	// 3) Precompute face axes so we can reuse for blobs + eyes/mouth
	delta := PlayerPos - PlanetPos
	dir := vec2(0.0, 1.0)
	if length(delta) > 0.0001 {
		asp := ScreenSize.x / ScreenSize.y
		dir = normalize(vec2(delta.x*asp, delta.y))
	}
	rightDir := vec2(dir.y, -dir.x)
	leftDir := -rightDir

	// build 2×2 rotation that maps (0,1)->dir and (1,0)->rightDir:
	faceRot := mat2(
		rightDir.x, dir.x,
		rightDir.y, dir.y,
	)

	outCol := vec4(0.0)

	// 4) Lava-lamp fill, _rotated_ by faceRot
	{
		uv2 := rotatedSphereUV(p)
		uv2 = faceRot * uv2 // ← align blobs to the face orientation
		m := metaballs(uv2)
		// threshold for bigger, gooey blobs
		lavaMask := smoothstep(2.5, 3.5, m) * aPlanet

		lavaCol := mix(
			vec3(0.9, 0.3, 0.1),
			vec3(1.0, 0.8, 0.2),
			lavaMask,
		)
		outCol = mix(outCol, vec4(lavaCol, 1.0), lavaMask)
	}

	// 5) Draw eyes (AA) inside circle
	{
		// left eye
		lp := leftDir*EyeSeparation + dir*EyeVertical
		dL := sdfCircle(p, lp, EyeRadius)
		aL := smoothstep(-fwidth(dL), fwidth(dL), -dL) * aPlanet
		outCol = mix(outCol, vec4(0, 0, 0, 1), aL)

		// right eye
		rp := rightDir*EyeSeparation + dir*EyeVertical
		dR := sdfCircle(p, rp, EyeRadius)
		aR := smoothstep(-fwidth(dR), fwidth(dR), -dR) * aPlanet
		outCol = mix(outCol, vec4(0, 0, 0, 1), aR)
	}

	// 6) Dynamic mouth
	playerDist := length(delta)
	mouthScale := 1.0
	if playerDist < PlanetOrbitRadius {
		frac := 1.0 - (playerDist / PlanetOrbitRadius)
		mouthScale = 1.0 + frac*4.0
	}
	scaledArcR := MouthArcRadius * mouthScale

	// 7) Capsule mouth
	{
		arcC := dir * MouthVertical
		q := p - arcC
		d := length(q)

		sdfArc := abs(d-scaledArcR) - MouthRadius
		mThick := smoothstep(-fwidth(sdfArc), fwidth(sdfArc), -sdfArc)
		nq := q / max(d, 1e-4)
		mAngle := step(MouthArcCos, dot(nq, -dir))
		mArc := mThick * mAngle * aPlanet

		sinA := sqrt(max(0.0, 1.0-MouthArcCos*MouthArcCos))
		vL := -dir*MouthArcCos + leftDir*sinA
		epL := arcC + vL*scaledArcR
		dE1 := sdfCircle(p, epL, MouthRadius)
		aE1 := smoothstep(-fwidth(dE1), fwidth(dE1), -dE1) * aPlanet

		vR := -dir*MouthArcCos + rightDir*sinA
		epR := arcC + vR*scaledArcR
		dE2 := sdfCircle(p, epR, MouthRadius)
		aE2 := smoothstep(-fwidth(dE2), fwidth(dE2), -dE2) * aPlanet

		mMask := max(mArc, max(aE1, aE2))
		outCol = mix(outCol, vec4(0, 0, 0, 1), mMask)
	}

	return outCol
}
