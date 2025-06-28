//go:build ignore
// +build ignore

//kage:unit pixels

package main

// — Compile‐time toggles —
const EnableCapsules = true        // master on/off switch for capsules
const AllCapsuleBelowPlayer = true // if true, *all* capsules draw beneath the player

// — Eye parameters (in circle‐local units) —
const EyeSeparation = 0.4 // how far apart the eyes are (x-axis)
const EyeVertical = 0.34  // how far up from center the eyes sit (y-axis)
const EyeRadius = 0.20    // radius of each white eye
const PupilRadius = 0.12  // radius of each black pupil

// — Mouth parameters (in circle‐local units) —
const MouthVertical = 0.5   // how far up (along dir) the mouth sits
const MouthRadius = 0.04    // thickness of the smile arc
const MouthArcRadius = 0.25 // radius of the smile circle
const MouthArcCos = 0.707   // cos(45°) → keeps ~90° arc opening up

// — Uniforms from Go —
var PlayerPos vec2   // [0..1] world‐space player center
var PlayerColor vec4 // player color (RGBA, 0–255)
var CameraPos vec2   // [0..1] world‐space camera center
var Zoom float       // zoom factor
var Radius float     // player radius [0..1]
var Velocity vec2    // world‐space units/sec
var Time float       // seconds
var ScreenSize vec2  // [width,height] in pixels

// trail parameters
var DropCount int
var TrailLength float
var DropSizeMin float
var DropSizeMax float
var JitterAmt float
var SpawnRate float
var Lifetime float

const MaxDrops = 16

func sdfCapsule(p, A, B vec2, r float) float {
	pa := p - A
	ba := B - A
	h := clamp(dot(pa, ba)/dot(ba, ba), 0.0, 1.0)
	return length(pa-ba*h) - r
}

func sdfCircle(p, c vec2, r float) float {
	return length(p-c) - r
}

func toLocal(dstPos vec4) vec2 {
	uv := dstPos.xy / ScreenSize
	c := uv - vec2(0.5, 0.5)
	asp := ScreenSize.x / ScreenSize.y
	c.x *= asp
	c /= Zoom
	world := CameraPos + c
	return (world - PlayerPos) / Radius
}

func hash(n float) float {
	return fract(sin(n) * 43758.5453)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 1) Convert to circle-local coords
	p := toLocal(dstPos)

	// 2) Transparent background
	var outCol vec4 = vec4(0.0, 0.0, 0.0, 0.0)

	// 3) Darkness shades for trail
	darkness := [4]float{1.0, 0.98, 0.95, 0.90}

	// 4) Accumulators
	var belowCol, aboveCol vec3 = vec3(0.0), vec3(0.0)
	var belowA, aboveA float = 0.0, 0.0

	// 5) Capsules (if enabled)
	if EnableCapsules && length(Velocity) > 0.01 {
		asp := ScreenSize.x / ScreenSize.y
		v := Velocity * vec2(asp, 1.0)
		dir := normalize(v)
		perp := vec2(-dir.y, dir.x)

		for i := 0; i < MaxDrops; i++ {
			if i >= DropCount {
				break
			}
			seed := float(i) + floor(Time*SpawnRate)
			t := fract(Time*SpawnRate + hash(float(i)*12.9898))
			if t < Lifetime/SpawnRate {
				life := 1.0 - t/(Lifetime/SpawnRate)
				jitter := (hash(seed*12.9898) - 0.5) * JitterAmt
				dist := t * TrailLength
				center := -dir*dist + perp*jitter

				var sdf, α float
				if life > 0.5 {
					size := mix(DropSizeMin, DropSizeMax, (life-0.5)*2.0)
					A, B := center, center-dir*size
					sdf = sdfCapsule(p, A, B, DropSizeMin)
					α = step(0.0, -sdf)
				} else {
					r := mix(0.05, DropSizeMin, life*2.0)
					sdf = sdfCircle(p, center, r)
					α = step(0.0, -sdf)
				}

				if α > 0.0 {
					idx := int(hash(seed*7.5625) * 4.0)
					col := (PlayerColor.rgb / 255.0) * darkness[idx]
					if !AllCapsuleBelowPlayer && hash(seed*3.14159) < 0.5 {
						aboveCol = col
						aboveA = max(aboveA, α)
					} else {
						belowCol = col
						belowA = max(belowA, α)
					}
				}
			}
		}
		// blend below capsules
		trailB := vec4(belowCol, min(belowA, PlayerColor.a/255.0))
		outCol = mix(outCol, trailB, trailB.a)
	}

	// 6) Draw the player circle
	if length(p) <= 1.0 {
		outCol = vec4(PlayerColor.rgb/255.0, PlayerColor.a/255.0)
	}

	// 7) Blend above capsules (if any)
	if EnableCapsules && !AllCapsuleBelowPlayer && length(Velocity) > 0.01 {
		trailA := vec4(aboveCol, min(aboveA, PlayerColor.a/255.0))
		outCol = mix(outCol, trailA, trailA.a)
	}

	// 8) Draw the eyes on top, looking toward CameraPos
	delta := CameraPos - PlayerPos
	var dir vec2
	if length(delta) > 0.0001 {
		asp := ScreenSize.x / ScreenSize.y
		dir = normalize(vec2(delta.x*asp, delta.y))
	} else {
		dir = vec2(0.0, 1.0)
	}

	// local axes for eye placement
	rightDir := vec2(dir.y, -dir.x)
	leftDir := -rightDir

	leftPos := leftDir*EyeSeparation + dir*EyeVertical
	rightPos := rightDir*EyeSeparation + dir*EyeVertical

	// white eyeballs
	aL := step(0.0, -sdfCircle(p, leftPos, EyeRadius))
	aR := step(0.0, -sdfCircle(p, rightPos, EyeRadius))
	outCol = mix(outCol, vec4(1.0, 1.0, 1.0, 1.0), aL)
	outCol = mix(outCol, vec4(1.0, 1.0, 1.0, 1.0), aR)

	// 9) Draw pupils (inner edge of each white eye)
	//    push inward by (EyeRadius - PupilRadius) along dir
	offset := dir * (EyeRadius - PupilRadius)
	pL := leftPos + offset
	pR := rightPos + offset

	pAL := step(0.0, -sdfCircle(p, pL, PupilRadius))
	pAR := step(0.0, -sdfCircle(p, pR, PupilRadius))

	outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), pAL)
	outCol = mix(outCol, vec4(0.0, 0.0, 0.0, 1.0), pAR)

	// 9) Cute mouth arc
	arcCenter := dir * MouthVertical
	q := p - arcCenter
	d := length(q)
	sdfArc := abs(d-MouthArcRadius) - MouthRadius
	mThick := step(0.0, -sdfArc)
	nq := q / d
	mAngle := step(MouthArcCos, dot(nq, dir))
	mMask := mThick * mAngle
	outCol = mix(outCol, vec4(0, 0, 0, 1), mMask)

	return outCol
}
