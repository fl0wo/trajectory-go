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

func particles(uv vec2, layer float) vec4 {
	// per-layer tuning (unchanged)
	var scale, speed, bright, size, parallax, zc float
	if layer < 1.0 {
		scale, speed, bright, size, parallax, zc = 1.0, 0.3, 0.3, 0.6, 0.2, 0.1
	} else if layer < 2.0 {
		scale, speed, bright, size, parallax, zc = 2.0, 0.5, 0.6, 1.2, 0.3, 0.3
	} else {
		scale, speed, bright, size, parallax, zc = 3.0, 0.7, 1.0, 1.6, 0.5, 0.5
	}

	// common transforms (zoom + parallax)
	wuv := worldUV(uv, parallax)
	scaled := wuv * scale
	cell := floor(scaled)
	frac := fract(scaled)

	// per-cell randomness
	rX := hash(cell)
	rY := hash(cell + vec2(5.2, 1.3))
	t := hash(cell + vec2(8.7, 3.4))

	// base orbiting “wingle”
	ang := Time*speed + (rX+rY)*6.2831

	// compute a safe, inset center so wiggle never overflows cell
	radius := 0.008 * size / ((1 + Zoom) * zc)
	wiggleAmp := 0.0001 // ±0.10 UV units
	pad := radius + 0.2 + wiggleAmp
	jX := rX*(1.0-2.0*pad) + pad
	jY := rY*(1.0-2.0*pad) + pad
	basePos := vec2(jX, jY)

	// combine orbit + wiggle
	wiggleFreq := 2.0 // ≈3s per full cycle
	wiggle := vec2(
		sin(Time*wiggleFreq+rX*6.2831),
		cos(Time*wiggleFreq+rY*6.2831),
	) * wiggleAmp
	pos := fract(basePos + vec2(cos(ang), sin(ang))*0.2 + wiggle)

	// mask that circle
	local := frac - pos
	aspectInv := ScreenSize.y / ScreenSize.x
	localCorr := vec2(local.x, local.y*aspectInv)
	mask := step(length(localCorr), radius)

	// tiny opacity pulse
	pulse := 1.0 + 0.1*sin(Time*1.0+(rX+rY)*6.2831)

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

	alpha := bright * mask * pulse
	return vec4(col, alpha)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// normalized UV
	uv := (dstPos.xy - imageDstOrigin()) / ScreenSize

	// base rich-black background
	result := vec4(11/255.0, 2/255.0, 43/255.0, 1.0)

	// three animated circle layers
	for layer := 0.0; layer < 3.0; layer++ {
		p := particles(uv, layer)
		result.rgb += p.rgb * p.a
	}

	// clamp and return
	result.rgb = min(result.rgb, vec3(1.0))
	return result
}
