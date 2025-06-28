//go:build ignore
// +build ignore

//kage:unit pixels

package main

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

	// 3) Draw the player with flat border
	d := length(p)
	if d <= 1.0 {
		outCol = vec4(PlayerColor.rgb/255.0, PlayerColor.a/255.0) // Flat player color
	}

	// 4) Draw capsules if player is moving
	if length(Velocity) > 0.01 {
		// Normalize velocity in circle-local space
		asp := ScreenSize.x / ScreenSize.y
		v := Velocity * vec2(asp, 1.0)
		dir := normalize(v)
		perp := vec2(-dir.y, dir.x) // Perpendicular vector

		var accCol vec3 = vec3(0.0)
		var accA float = 0.0

		// Spawn capsules based on time
		for i := 0; i < MaxDrops; i++ {
			if i >= DropCount {
				break
			}

			// Unique seed for each capsule
			seed := float(i) + floor(Time*SpawnRate)
			t := fract(Time*SpawnRate + hash(float(i)*12.9898))

			// Only draw if capsule is "alive" (within lifetime)
			if t < Lifetime/SpawnRate {
				// Capsule lifetime (1.0 to 0.0)
				life := 1.0 - t/(Lifetime/SpawnRate)

				// Randomize capsule properties
				jitter := (hash(seed*12.9898) - 0.5) * JitterAmt
				// Interpolate size from DropSizeMax to DropSizeMin
				size := mix(DropSizeMin, DropSizeMax, life)

				// Position capsule behind player with 2x radius offset
				dist := t * TrailLength           // Start 2x radius away
				center := -dir*dist + perp*jitter // Offset in opposite direction

				// Capsule endpoints
				A := center
				B := center - dir*size // Extend capsule in direction

				// SDF with flat border
				sdf := sdfCapsule(p, A, B, DropSizeMin)
				α := step(0.0, -sdf) // Flat border, no alpha fade

				// Use flat PlayerColor, avoid additive white
				if α > 0.0 {
					accCol = PlayerColor.rgb / 255.0 // Set to PlayerColor directly
					accA = max(accA, α)              // Take max alpha to avoid overlap
				}
			}
		}

		// Blend capsules over background/player
		trail := vec4(accCol, min(accA, PlayerColor.a/255.0))
		outCol = mix(outCol, trail, trail.a)
	}

	return outCol
}
