//go:build ignore
// +build ignore

//kage:unit pixels

package main

// — uniforms from Go —
var PlayerPos vec2  // [0..1] world-space player center
var CameraPos vec2  // [0..1] world-space camera center
var Zoom float      // zoom factor (1=no zoom)
var Radius float    // player radius in [0..1] units
var Velocity vec2   // world-space units/sec
var Time float      // seconds since start
var ScreenSize vec2 // [width,height] in pixels

// trail parameters
var CapsuleCount int    // how many capsules alive at once
var CapsuleLife float   // how many seconds each capsule lives
var TrailLength float   // how many radii from start→end
var CapsuleRadius float // capsule half-thickness in radii
var JitterAmt float     // max side-to-side offset in radii

const MaxCapsules = 16 // must ≥ max possible CapsuleCount

// SDF for capsule from A→B with radius r
func sdfCapsule(p, A, B vec2, r float) float {
	pa := p - A
	ba := B - A
	h := clamp(dot(pa, ba)/dot(ba, ba), 0.0, 1.0)
	return length(pa-ba*h) - r
}

// map pixel → “circle-local” co‐ords
func toLocal(dstPos vec4) vec2 {
	uv := dstPos.xy / ScreenSize    // [0..1]
	centered := uv - vec2(0.5, 0.5) // [-.5..+.5]
	aspect := ScreenSize.x / ScreenSize.y
	centered.x *= aspect                // correct aspect
	centered /= Zoom                    // apply zoom
	world := CameraPos + centered       // world-space [0..1]
	return (world - PlayerPos) / Radius // circle-local
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	p := toLocal(dstPos)
	// draw the player as a solid green circle
	d := length(p)
	var outCol vec4
	if d <= 1.0 {
		alpha := 1.0 - smoothstep(0.8, 1.0, d)
		outCol = vec4(0.0, 1.0, 0.0, alpha)
	}

	// only if moving
	if length(Velocity) > 0.01 {
		// compute normalized back-direction
		dir := normalize(-Velocity * vec2(aspectCorrection(), 1.0))

		// accumulate all capsules
		var accCol vec3 = vec3(0.0)
		var accA float = 0.0

		for i := 0; i < MaxCapsules; i++ {
			if i >= CapsuleCount {
				break
			}
			// each capsule has its own birth-time offset
			birth := float(i) * (CapsuleLife / float(CapsuleCount))
			age := mod(Time-birth, CapsuleLife)
			t := age / CapsuleLife // [0..1]
			// along-trail position in radii
			basePos := dir * (TrailLength * t)
			// little side jitter
			perp := vec2(-dir.y, dir.x)
			off := (hash(birth*3.14) - 0.5) * JitterAmt
			center := basePos + perp*off

			// capsule endpoints in circle-local space
			A := center
			B := center + dir*TrailLength*0.2 // capsules are 0.2× length
			// smooth fade-out as age→CapsuleLife
			fade := 1.0 - t

			// compute SDF & alpha
			dist := sdfCapsule(p, A, B, CapsuleRadius)
			α := smoothstep(0.0, CapsuleRadius*0.5, -dist) * fade

			accCol += vec3(0.0, 1.0, 0.0) * α
			accA += α
		}
		// blend capsule trail over player
		trailCol := vec4(accCol, min(accA, 1.0))
		outCol = mix(outCol, trailCol, trailCol.a)
	}

	return outCol
}

// returns aspect for Velocity.x scaling
func aspectCorrection() float {
	return ScreenSize.x / ScreenSize.y
}

// simple hash for jitter
func hash(n float) float {
	return fract(sin(n) * 43758.5453)
}

// smoothstep for fade
func smoothstep(e0, e1, x float) float {
	t := clamp((x-e0)/(e1-e0), 0.0, 1.0)
	return t * t * (3.0 - 2.0*t)
}
