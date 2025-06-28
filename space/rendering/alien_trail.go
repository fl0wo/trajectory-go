//go:build ignore
// +build ignore

//kage:unit pixels

package main

// — Compile-time toggle —
// Set to false to completely disable drawing capsules
const EnableCapsules = true
const AllCapsuleBelowPlayer = true // if true, *all* capsules draw beneath the player

// — Uniforms from Go —
var PlayerPos vec2   // [0..1] world‐space player center
var PlayerColor vec4 // player color (RGBA, 0-255)
var CameraPos vec2   // [0..1] world‐space camera center
var Zoom float       // zoom factor
var Radius float     // player radius [0..1]
var Velocity vec2    // world‐space units/sec
var Time float       // seconds
var ScreenSize vec2  // [width,height] in pixels

// trail parameters
var DropCount int     // max number of capsules
var TrailLength float // max distance from player (in radii)
var DropSizeMin float // capsule thickness (in radii)
var DropSizeMax float // capsule length (in radii)
var JitterAmt float   // max side‐to‐side offset (in radii)
var SpawnRate float   // capsules spawned per second
var Lifetime float    // how long each capsule lives (seconds)

const MaxDrops = 16

// SDF for capsule segment A→B radius r
func sdfCapsule(p, A, B vec2, r float) float {
	pa := p - A
	ba := B - A
	h := clamp(dot(pa, ba)/dot(ba, ba), 0.0, 1.0)
	return length(pa-ba*h) - r
}

// SDF for circle at center c with radius r
func sdfCircle(p, c vec2, r float) float {
	return length(p-c) - r
}

// Convert pixel → circle-local coords
func toLocal(dstPos vec4) vec2 {
	uv := dstPos.xy / ScreenSize // [0..1]
	c := uv - vec2(0.5, 0.5)     // [-.5..+.5]
	asp := ScreenSize.x / ScreenSize.y
	c.x *= asp                          // keep circles round
	c /= Zoom                           // apply zoom
	world := CameraPos + c              // back to [0..1]
	return (world - PlayerPos) / Radius // circle-local
}

// Hash function for pseudo-random numbers
func hash(n float) float {
	return fract(sin(n) * 43758.5453)
}
func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 1) Convert to circle-local coordinates
	p := toLocal(dstPos)

	// 2) Initialize output color (transparent background)
	var outCol vec4 = vec4(0.0, 0.0, 0.0, 0.0)

	// 3) Darkness factors
	darknessFactors := [4]float{1.0, 0.98, 0.95, 0.90}

	// 4) Accumulators
	var belowCol, aboveCol vec3 = vec3(0.0), vec3(0.0)
	var belowA, aboveA float = 0.0, 0.0

	// 5) Process capsules (only if enabled)
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
					A := center
					B := center - dir*size
					sdf = sdfCapsule(p, A, B, DropSizeMin)
					α = step(0.0, -sdf)
				} else {
					radius := mix(0.05, DropSizeMin, life*2.0)
					sdf = sdfCircle(p, center, radius)
					α = step(0.0, -sdf)
				}

				if α > 0.0 {
					shadeIndex := int(hash(seed*7.5625) * 4.0)
					darkness := darknessFactors[shadeIndex]
					capsuleColor := (PlayerColor.rgb / 255.0) * darkness

					// Decide above vs below:
					if !AllCapsuleBelowPlayer && hash(seed*3.14159) < 0.5 {
						aboveCol = capsuleColor
						aboveA = max(aboveA, α)
					} else {
						belowCol = capsuleColor
						belowA = max(belowA, α)
					}
				}
			}
		}

		// Blend all below-player capsules
		belowTrail := vec4(belowCol, min(belowA, PlayerColor.a/255.0))
		outCol = mix(outCol, belowTrail, belowTrail.a)
	}

	// 6) Draw the player (always on top of below‐capsules)
	if length(p) <= 1.0 {
		outCol = vec4(PlayerColor.rgb/255.0, PlayerColor.a/255.0)
	}

	// 7) Optionally blend above-player capsules
	if EnableCapsules && !AllCapsuleBelowPlayer && length(Velocity) > 0.01 {
		aboveTrail := vec4(aboveCol, min(aboveA, PlayerColor.a/255.0))
		outCol = mix(outCol, aboveTrail, aboveTrail.a)
	}

	return outCol
}
