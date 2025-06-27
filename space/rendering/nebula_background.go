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

func particles(uv vec2, layer float) vec4 {
	// per-layer tuning
	var scale, speed, brightness, size, parallax, zoomConsider float
	if layer < 1.0 {
		scale, speed, brightness, size, parallax, zoomConsider = 10.0, 0.3, 0.1, 1.0, 0.2, 0.1
	} else if layer < 2.0 {
		scale, speed, brightness, size, parallax, zoomConsider = 20.0, 0.5, 0.3, 1.2, 0.4, 0.4
	} else {
		scale, speed, brightness, size, parallax, zoomConsider = 30.0, 0.7, 0.5, 1.5, 0.7, 0.6
	}

	// 1) apply zoom around screen center
	uv = (uv-vec2(0.5, 0.5))/Zoom + vec2(0.5, 0.5)

	// 2) compute camera-only parallax offset (CameraPos is already [0..1])
	//    subtract 0.5 to center into [-0.5..+0.5]
	camCenter := vec2(0.5, 0.5) - CameraPos
	uv += -camCenter * parallax

	// 3) build grid at given density
	scaled := uv * scale
	cell := floor(scaled)
	frac := fract(scaled)

	// 4) per-cell RNG
	rX := hash(cell)
	rY := hash(cell + vec2(5.2, 1.3))
	t := hash(cell + vec2(8.7, 3.4))

	// 5) wingling motion
	ang := Time*speed + (rX+rY)*6.2831
	pos := fract(vec2(rX, rY) + vec2(cos(ang), sin(ang))*0.2)

	// 6) flat dot
	dist := length(frac - pos)
	radius := 0.05 * size / ((1 + Zoom) * zoomConsider)
	alpha := brightness * step(dist, radius)

	// 7) flat palette
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

	return vec4(col, alpha)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// 0) normalized UV from pixel coords
	uv := (dstPos.xy - imageDstOrigin()) / ScreenSize

	// 1) base rich-black background
	result := vec4(0.02, 0.004, 0.078, 1.0)

	// 2) draw 3 layers of flat-shaded particles
	for layer := 0.0; layer < 3.0; layer++ {
		p := particles(uv, layer)
		result.rgb += p.rgb * p.a
	}

	// 3) clamp and output
	result.rgb = min(result.rgb, vec3(1.0))
	return result
}
