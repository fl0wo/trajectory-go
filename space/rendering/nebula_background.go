//go:build ignore

//kage:unit pixels

package main

// Uniforms
var Time float
var CameraPos vec2  // in normalized UV [0..1]
var ScreenSize vec2 // in pixels (only for initial UV calc)
var Zoom float      // <1 = zoomed out, >1 = zoomed in

// Pseudo-random [0..1)
func hash(p vec2) float {
	return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453)
}

// Zoom + camera parallax helper
func worldUV(uv vec2, parallax float) vec2 {
	uv = (uv-vec2(0.5, 0.5))/Zoom + vec2(0.5, 0.5)
	off := vec2(0.5, 0.5) - CameraPos
	return uv - off*parallax
}

// Particle layers: circles only, no more “cut” edges
func particles(uv vec2, layer float) vec4 {
	// per-layer tuning
	// per-layer tuning
	var scale, speed, bright, size, parallax, zc float
	if layer < 1.0 {
		// farthest: big, faint, hardly moves
		scale, speed, bright, size, parallax, zc = 1.0, 0.3, 0.3, .6, 0.2, 0.1
	} else if layer < 2.0 {
		// middle: medium size, medium brightness & parallax
		scale, speed, bright, size, parallax, zc = 2.0, 0.5, 0.6, 1.2, 0.3, 0.3
	} else {
		// nearest: small, bright, moves most
		scale, speed, bright, size, parallax, zc = 3.0, 0.7, 1.0, 1.6, 0.5, 0.5
	}

	// world-space UV
	wuv := worldUV(uv, parallax)

	// grid cell
	scaled := wuv * scale
	cell := floor(scaled)
	frac := fract(scaled)

	// RNG seeds
	rX := hash(cell)
	rY := hash(cell + vec2(5.2, 1.3))
	t := hash(cell + vec2(8.7, 3.4))

	// compute radius (cell units)
	radius := 0.008 * size / ((1 + Zoom) * zc)

	// wingling animation
	ang := Time*speed + (rX+rY)*6.2831

	// margin to keep circle + wiggle inside cell
	pad := radius + 0.2

	// safe jitter center in [pad..1-pad]
	jX := rX*(1.0-2.0*pad) + pad
	jY := rY*(1.0-2.0*pad) + pad

	// final particle position, then fract to remain in [0,1]
	pos := fract(
		vec2(
			jX+cos(ang)*0.2,
			jY+sin(ang)*0.2,
		),
	)

	// local offset within cell
	local := frac - pos

	// aspect-correct for non-square screens
	aspectInv := ScreenSize.y / ScreenSize.x
	localCorr := vec2(local.x, local.y*aspectInv)

	// circle mask
	mask := step(length(localCorr), radius)

	// flat palette
	var col vec3
	if t < 0.02 {
		col = vec3(1.0, 1.0, 0.992)
	} else if t < 0.40 {
		col = vec3(0.129, 0.059, 0.314)
	} else if t < 0.75 {
		col = vec3(0.278, 0.102, 0.459)
	} else if t < 0.90 {
		col = vec3(0.282, 0.110, 0.459)
	} else if t < 0.96 {
		col = vec3(0.220, 0.008, 0.635)
	} else {
		col = vec3(0.031, 0.008, 0.106)
	}

	alpha := bright * mask
	return vec4(col, alpha)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// normalized UV
	uv := (dstPos.xy - imageDstOrigin()) / ScreenSize

	// base rich-black background
	result := vec4(11/255.0, 2/255.0, 43/255.0, 1.0)

	// three layers of circular particles
	for layer := 0.0; layer < 3.0; layer++ {
		p := particles(uv, layer)
		result.rgb += p.rgb * p.a
	}

	// clamp & return
	result.rgb = min(result.rgb, vec3(1.0))
	return result
}
