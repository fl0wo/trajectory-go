//go:build ignore
// +build ignore

//kage:unit pixels

package main

// Eye parameters (in circle-local units)
const EyeSeparation = 0.6
const EyeVertical = 0.4
const EyeRadius = 0.25

// Mouth parameters
const MouthVertical = -0.3
const MouthRadius = 0.08
const MouthArcRadius = 0.35
const MouthArcCos = 0.5 // cos(60°) ≈ 0.5

// Uniforms from Go
var PlayerPos vec2   // [0..1] world-space player center
var PlanetPos vec2   // [0..1] world-space planet center
var PlanetColor vec4 // Planet color (RGBA 0–255)
var CameraPos vec2   // [0..1] world-space camera center
var Zoom float       // Zoom factor
var Radius float     // Planet radius [0..1]
var Time float       // Seconds (unused here)
var ScreenSize vec2  // [width, height] in pixels

// SDF for circle at center c with radius r
func sdfCircle(p, c vec2, r float) float {
	return length(p-c) - r
}

// Map pixel to unit-circle local coords around the planet
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

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 1) Circle-local coords
	p := toLocal(dstPos)

	// 2) Planet mask
	distPlanet := length(p)
	inPlanet := step(distPlanet, 1.0)

	// 3) Start with transparent, then fill planet base
	outCol := vec4(0.0)
	outCol = mix(outCol, vec4(PlanetColor.rgb/255.0, 1.0), inPlanet)

	// 4) Compute look-direction toward the player
	delta := PlayerPos - PlanetPos
	dir := vec2(0.0, 1.0)
	if length(delta) > 0.0001 {
		asp := ScreenSize.x / ScreenSize.y
		dir = normalize(vec2(delta.x*asp, delta.y))
	}
	rightDir := vec2(dir.y, -dir.x)
	leftDir := -rightDir

	// 5) Eyes (only inside planet, with fwidth anti-alias)
	if inPlanet > 0.5 {
		// Left eye
		leftPos := leftDir*EyeSeparation + dir*EyeVertical
		dL := sdfCircle(p, leftPos, EyeRadius)
		wL := fwidth(dL)
		aL := smoothstep(-wL, wL, -dL)
		outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), aL)

		// Right eye
		rightPos := rightDir*EyeSeparation + dir*EyeVertical
		dR := sdfCircle(p, rightPos, EyeRadius)
		wR := fwidth(dR)
		aR := smoothstep(-wR, wR, -dR)
		outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), aR)
	}

	// 6) Mouth arc (only inside planet, with fwidth anti-alias)
	if inPlanet > 0.5 {
		arcCenter := dir * MouthVertical
		q := p - arcCenter
		d := length(q)
		sdfArc := abs(d-MouthArcRadius) - MouthRadius

		// thickness + anti-alias
		wM := fwidth(sdfArc)
		mThick := smoothstep(-wM, wM, -sdfArc)

		// arc only where angle < 120°
		nq := q / max(d, 0.0001)
		mAngle := step(MouthArcCos, dot(nq, -dir))

		mMask := mThick * mAngle
		outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), mMask)
	}

	return outCol
}
