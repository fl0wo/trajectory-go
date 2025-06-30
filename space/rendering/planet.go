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
var ScreenSize vec2      // [width, height] in pixels
var BaseColor vec3       // base color for the planet
var Seed float           // random seed for lava animation
var ShakeTimer float     // shake timer (0 = no shake, >0 = shake remaining)
var ShakeIntensity float // shake intensity multiplier

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

	// Apply shake effect if timer is active
	shakeOffset := vec2(0.0)
	if ShakeTimer > 0.0 {
		// Exponential decay for shake intensity
		const initialShakeIntensity = 0.002               // Must match Go code's initial ShakeIntensity
		const totalShakeDuration = 2.0                    // Must match Go code's initial ShakeTimer
		const decayRate = 2.0                             // Controls speed of exponential decay
		normalizedTime := ShakeTimer / totalShakeDuration // [0,1]
		adjustedIntensity := initialShakeIntensity * exp(-decayRate*(1.0-normalizedTime))

		// Generate shake offset using sine waves with different frequencies
		shakeX := adjustedIntensity * sin(ShakeTimer*15.0)
		shakeY := adjustedIntensity * cos(ShakeTimer*18.0)
		shakeOffset = vec2(shakeX, shakeY)
	}

	// Apply shake offset to planet position
	adjustedPlanetPos := PlanetPos + shakeOffset

	radiusSafe := max(0.0001, min(Radius, 1.0))
	return (world - adjustedPlanetPos) / radiusSafe
}

// ─── Bigger metaballs ────────────────────────────────────────────────
func metaballs(uv vec2) float {
	sum := 0.0
	const N = 10
	for i := 0; i < N; i++ {
		phase := Time*(0.2+float(i)*0.08) + float(i)*1.7 + Seed*3.14159
		c := vec2(cos(phase+Seed), sin(phase+Seed*1.3)) * 0.5
		r := 0.28 + 0.1*sin(Time*1.2+float(i)*2.3+Seed*2.1)
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

	// spin all blobs one way around Y
	angle := Time*0.4 + Seed
	rotY := mat3(
		cos(angle), 0, sin(angle),
		0, 1, 0,
		-sin(angle), 0, cos(angle),
	)

	rp := rotY * p3
	return rp.xy
}

// Fragment shader
func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 1) Into circle coords & mask planet
	p := toLocal(dstPos)
	p = -p

	distPlanet := length(p)
	sdfPlanet := distPlanet - 1.0
	wPlanet := fwidth(sdfPlanet)
	aPlanet := smoothstep(wPlanet, -wPlanet, sdfPlanet)
	if aPlanet < 0.001 {
		return vec4(0.0) // outside
	}

	// 2) Face orientation (so “light” direction follows player)
	delta := PlayerPos - PlanetPos
	dir := vec2(0.0, 1.0)
	if length(delta) > 1e-4 {
		asp := ScreenSize.x / ScreenSize.y
		dir = normalize(vec2(delta.x*asp, delta.y))
	}
	rightDir := vec2(dir.y, -dir.x)
	leftDir := -rightDir
	faceRot := mat2(
		rightDir.x, dir.x,
		rightDir.y, dir.y,
	)

	// 3) Base fill (darkCol)
	ambient := vec3(0.2)
	darkCol := BaseColor*0.8 + ambient
	outCol := vec4(darkCol, 1.0) * aPlanet

	// 4) Gooey lava‐lamp blobs
	{
		uv2 := faceRot * rotatedSphereUV(p)
		m := metaballs(uv2)

		const LOW = 2.5
		const HIGH = 3.5
		lavaMask := smoothstep(LOW, HIGH, m)

		const VALLEY_LIGHT = 0.7 // 0 = old darkCol, 1 = pure BaseColor
		valleyCol := mix(darkCol, BaseColor, VALLEY_LIGHT)

		intenseCol := BaseColor*1.4 + vec3(0.1, 0.1, 0.0)
		mostIntenseCol := BaseColor*1.8 + vec3(0.2, 0.2, 0.1)

		lavaCol := mix(
			valleyCol,
			mix(mostIntenseCol, intenseCol, lavaMask),
			lavaMask,
		)

		outCol = mix(outCol, vec4(lavaCol, 1.0), lavaMask)
	}

	// 5) Static “shadow” rim opposite the light direction (more diffused)
	{
		// compute 2D light direction
		asp := ScreenSize.x / ScreenSize.y
		lightDir2D := normalize(vec2(delta.x*asp, delta.y))

		// place shadow opposite
		shadowPos := lightDir2D

		const HR = 1.0  // rim radius in circle‐local units
		const HW = 0.25 // increased softness half‐width for more diffusion

		dSH := length(p - shadowPos)
		raw := smoothstep(HR-HW, HR+HW, dSH)
		shadowMask := (1.0 - raw) * aPlanet

		// blend toward darkCol to simulate shadow
		outCol = mix(outCol, vec4(darkCol, 1.0), shadowMask)
	}

	// 6) Draw eyes (AA) on top
	{
		lp := leftDir*EyeSeparation + dir*EyeVertical
		dL := sdfCircle(p, lp, EyeRadius)
		aL := smoothstep(-fwidth(dL), fwidth(dL), -dL) * aPlanet
		outCol = mix(outCol, vec4(0, 0, 0, 1), aL)

		rp := rightDir*EyeSeparation + dir*EyeVertical
		dR := sdfCircle(p, rp, EyeRadius)
		aR := smoothstep(-fwidth(dR), fwidth(dR), -dR) * aPlanet
		outCol = mix(outCol, vec4(0, 0, 0, 1), aR)
	}

	// 7) Capsule mouth on top
	{
		playerDist := length(delta)
		mouthScale := 1.0
		if playerDist < PlanetOrbitRadius {
			frac := 1.0 - (playerDist / PlanetOrbitRadius)
			mouthScale = 1.0 + frac*4.0
		}
		scaledArcR := MouthArcRadius * mouthScale

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
