//go:build ignore

//kage:unit pixels

package main

// Uniforms
var LightPos vec2
var MaxDistance float
var Zoom float // 1 = standard, >1 = in, <1 = out

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// world-space pixel pos
	pos := dstPos.xy - imageDstOrigin()
	// normalized distance 0→1
	dist := length(pos-LightPos) / Zoom
	t := min(dist/MaxDistance, 1.0)

	const STEPS = 5
	const TOTAL_UNITS = 31.0

	// 1) per-band alpha
	var alphas = [STEPS]float{
		.8,  // band 1
		.2,  // band 2
		.2,  // band 3
		0.0, // band 4 (transparent)
		.1,  // band 5
	}

	// 2) per-band color
	var colors = [STEPS]vec3{
		vec3(254.0/255.0, 254.0/255.0, 254.0/255.0), // rgb(254,254,254)
		vec3(253.0/255.0, 218.0/255.0, 145.0/255.0), // rgb(253,218,145)
		vec3(254.0/255.0, 166.0/255.0, 59.0/255.0),  // rgb(254,166,59)
		vec3(0.0, 0.0, 0.0),                         // transparent
		vec3(245.0/255.0, 57.0/255.0, 68.0/255.0),   // rgb(245,57,68)
	}

	// 3) explicit spacing for each band (must sum ≤ 1.0)
	var spaces = [STEPS]float{
		10.0 / TOTAL_UNITS, // band 1 width
		0.5 / TOTAL_UNITS,  // band 2 width
		0.5 / TOTAL_UNITS,  // band 3 width
		1.0 / TOTAL_UNITS,  // band 4 width
		0.25 / TOTAL_UNITS, // band 5 width
	}

	// walk outward summing spaces[]
	var threshold = 0.0
	var alpha = 0.0
	var lightColor = vec3(0.0)
	for i := 0; i < STEPS; i++ {
		threshold += spaces[i]
		if t < threshold {
			alpha = alphas[i]
			lightColor = colors[i]
			break
		}
	}

	// premultiplied alpha
	return vec4(lightColor*alpha, alpha)
}
