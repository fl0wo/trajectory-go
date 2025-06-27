package shadows

import (
	_ "embed"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/math/f32"
	"image/color"
	"math"
	"sort"
)

//go:embed stepped_light.go
var steppedLightShader []byte

// ShadowSystem handles raycasting and shadow rendering for the space game
type ShadowSystem struct {
	shadowImage   *ebiten.Image
	triangleImage *ebiten.Image
	lightShader   *ebiten.Shader
	screenWidth   int
	screenHeight  int
}

type line struct{ X1, Y1, X2, Y2 float64 }

func (l *line) angle() float64 { return math.Atan2(l.Y2-l.Y1, l.X2-l.X1) }
func newRay(x, y, length, angle float64) line {
	return line{x, y, x + length*math.Cos(angle), y + length*math.Sin(angle)}
}

// clipRay fires a single ray at 'angle' and stops it at the first occluder (or lets it run full-length)
func clipRay(
	lightX, lightY, length, angle float64,
	occluderLines []line,
) line {
	r := newRay(lightX, lightY, length, angle)
	var hits [][2]float64
	for _, ol := range occluderLines {
		if px, py, ok := intersection(r, ol); ok {
			hits = append(hits, [2]float64{px, py})
		}
	}
	if len(hits) > 0 {
		minD := math.Inf(1)
		var closest [2]float64
		for _, p := range hits {
			d := (lightX-p[0])*(lightX-p[0]) + (lightY-p[1])*(lightY-p[1])
			if d < minD {
				minD = d
				closest = p
			}
		}
		return line{lightX, lightY, closest[0], closest[1]}
	}
	return r
}

func intersection(l1, l2 line) (float64, float64, bool) {
	den := (l1.X1-l1.X2)*(l2.Y1-l2.Y2) - (l1.Y1-l1.Y2)*(l2.X1-l2.X2)
	if den == 0 {
		return 0, 0, false
	}
	t := ((l1.X1-l2.X1)*(l2.Y1-l2.Y2) - (l1.Y1-l2.Y1)*(l2.X1-l2.X2)) / den
	if t < 0 || t > 1 {
		return 0, 0, false
	}
	u := -((l1.X1-l1.X2)*(l1.Y1-l2.Y1) - (l1.Y1-l1.Y2)*(l1.X1-l2.X1)) / den
	if u < 0 || u > 1 {
		return 0, 0, false
	}
	x := l1.X1 + t*(l1.X2-l1.X1)
	y := l1.Y1 + t*(l1.Y2-l1.Y1)
	return x, y, true
}

func circleToLines(center f32.Vec2, radius float32, segments int) []line {
	var lines []line
	step := 2 * math.Pi / float64(segments)
	for i := 0; i < segments; i++ {
		a1 := float64(i) * step
		a2 := float64(i+1) * step
		x1 := float64(center[0]) + float64(radius)*math.Cos(a1)
		y1 := float64(center[1]) + float64(radius)*math.Sin(a1)
		x2 := float64(center[0]) + float64(radius)*math.Cos(a2)
		y2 := float64(center[1]) + float64(radius)*math.Sin(a2)
		lines = append(lines, line{x1, y1, x2, y2})
	}
	return lines
}

func getAllOccluderPoints(
	celestialPositions []f32.Vec2, celestialRadii []float32,
	asteroidPositions []f32.Vec2, asteroidRadii []float32,
) [][2]float64 {
	var pts [][2]float64
	for i, pos := range celestialPositions {
		r := celestialRadii[i]
		for j := 0; j < 16; j++ {
			θ := 2 * math.Pi * float64(j) / 16
			pts = append(pts, [2]float64{
				float64(pos[0]) + float64(r)*math.Cos(θ),
				float64(pos[1]) + float64(r)*math.Sin(θ),
			})
		}
	}
	for i, pos := range asteroidPositions {
		r := asteroidRadii[i]
		for j := 0; j < 8; j++ {
			θ := 2 * math.Pi * float64(j) / 8
			pts = append(pts, [2]float64{
				float64(pos[0]) + float64(r)*math.Cos(θ),
				float64(pos[1]) + float64(r)*math.Sin(θ),
			})
		}
	}
	return pts
}

func getAllOccluderLines(
	celestialPositions []f32.Vec2, celestialRadii []float32,
	asteroidPositions []f32.Vec2, asteroidRadii []float32,
) []line {
	var lines []line
	for i, pos := range celestialPositions {
		lines = append(lines, circleToLines(pos, celestialRadii[i], 16)...)
	}
	for i, pos := range asteroidPositions {
		lines = append(lines, circleToLines(pos, asteroidRadii[i], 8)...)
	}
	return lines
}

func normalizeAngle(a float64) float64 {
	for a <= -math.Pi {
		a += 2 * math.Pi
	}
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	return a
}

func rayCasting(
	lightX, lightY float64,
	celestialPositions []f32.Vec2, celestialRadii []float32,
	asteroidPositions []f32.Vec2, asteroidRadii []float32,
) []line {
	const rayLength = 3000
	points := getAllOccluderPoints(celestialPositions, celestialRadii, asteroidPositions, asteroidRadii)
	olines := getAllOccluderLines(celestialPositions, celestialRadii, asteroidPositions, asteroidRadii)

	var rays []line
	for _, pt := range points {
		base := line{lightX, lightY, pt[0], pt[1]}
		ang := base.angle()
		for _, off := range []float64{-0.002, 0.002} {
			r := newRay(lightX, lightY, rayLength, ang+off)
			var hits [][2]float64
			for _, ol := range olines {
				if x, y, ok := intersection(r, ol); ok {
					hits = append(hits, [2]float64{x, y})
				}
			}
			if len(hits) > 0 {
				minD := math.Inf(1)
				var cpt [2]float64
				for _, h := range hits {
					d := (lightX-h[0])*(lightX-h[0]) + (lightY-h[1])*(lightY-h[1])
					if d < minD {
						minD = d
						cpt = h
					}
				}
				rays = append(rays, line{lightX, lightY, cpt[0], cpt[1]})
			} else {
				rays = append(rays, r)
			}
		}
	}
	sort.Slice(rays, func(i, j int) bool {
		return rays[i].angle() < rays[j].angle()
	})
	return rays
}

func rayVertices(x1, y1, x2, y2, x3, y3 float64) []ebiten.Vertex {
	return []ebiten.Vertex{
		{DstX: float32(x1), DstY: float32(y1), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(x2), DstY: float32(y2), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(x3), DstY: float32(y3), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
}

// RenderShadows renders a spotlight of fovAngle (degrees) aimed at winPos
func (ss *ShadowSystem) RenderShadows(
	screen *ebiten.Image,
	lightPos f32.Vec2,
	winPos f32.Vec2,
	fovAngle float64,
	cameraZoom float32,
	celestialPositions []f32.Vec2,
	celestialRadii []float32,
	asteroidPositions []f32.Vec2,
	asteroidRadii []float32,
	showRays bool,
) {
	if len(celestialPositions) == 0 && len(asteroidPositions) == 0 {
		return
	}

	// 1) Darken entire screen
	ss.shadowImage.Fill(color.RGBA{R: 6, G: 2, B: 25, A: 255})

	lightX, lightY := float64(lightPos[0]), float64(lightPos[1])

	// 2) Get all interior (already clipped) rays
	allRays := rayCasting(
		lightX, lightY,
		celestialPositions, celestialRadii,
		asteroidPositions, asteroidRadii,
	)

	// 3) Spotlight center & half‐FOV
	base := math.Atan2(
		float64(winPos[1])-lightY,
		float64(winPos[0])-lightX,
	)
	half := (fovAngle * math.Pi / 180) / 2

	// 4) Filter interior rays inside cone
	var interior []line
	for _, r := range allRays {
		if math.Abs(normalizeAngle(r.angle()-base)) <= half {
			interior = append(interior, r)
		}
	}
	sort.Slice(interior, func(i, j int) bool {
		return normalizeAngle(interior[i].angle()-base) <
			normalizeAngle(interior[j].angle()-base)
	})

	// 5) Build occluderLines for boundary clipping
	occluderLines := getAllOccluderLines(
		celestialPositions, celestialRadii,
		asteroidPositions, asteroidRadii,
	)
	diag := math.Hypot(float64(ss.screenWidth), float64(ss.screenHeight))

	// 6) Clip the two boundary rays
	left := clipRay(lightX, lightY, diag, base-half, occluderLines)
	right := clipRay(lightX, lightY, diag, base+half, occluderLines)

	// 7) Assemble final ray list: left boundary, interior, right boundary
	finalRays := make([]line, 0, len(interior)+2)
	finalRays = append(finalRays, left)
	finalRays = append(finalRays, interior...)
	finalRays = append(finalRays, right)

	// 8) Carve holes in the shadow image
	triOpt := &ebiten.DrawTrianglesOptions{Address: ebiten.AddressRepeat}
	triOpt.Blend = ebiten.BlendSourceOut
	for i := 0; i < len(finalRays)-1; i++ {
		a, b := finalRays[i], finalRays[i+1]
		vs := rayVertices(lightX, lightY, b.X2, b.Y2, a.X2, a.Y2)
		ss.shadowImage.DrawTriangles(vs, []uint16{0, 1, 2}, ss.triangleImage, triOpt)
	}

	// 9) Overlay the semi-transparent shadow
	shadowOpt := &ebiten.DrawImageOptions{}
	shadowOpt.ColorScale.ScaleAlpha(0.7)
	screen.DrawImage(ss.shadowImage, shadowOpt)

	// 10) Draw the light cone with stepped falloff shader
	if ss.lightShader != nil {
		// Calculate max distance for the light cone
		maxDistance := math.Hypot(float64(ss.screenWidth), float64(ss.screenHeight))

		lightOpt := &ebiten.DrawTrianglesShaderOptions{}
		lightOpt.Uniforms = map[string]any{
			"LightPos":    []float32{float32(lightX), float32(lightY)},
			"MaxDistance": float32(maxDistance),
			"Zoom":        cameraZoom,
		}
		lightOpt.Images[0] = ss.triangleImage
		lightOpt.Blend = ebiten.BlendSourceOver

		// Draw each triangle of the light cone with the shader applied
		for i := 0; i < len(finalRays)-1; i++ {
			a, b := finalRays[i], finalRays[i+1]
			vs := rayVertices(lightX, lightY, b.X2, b.Y2, a.X2, a.Y2)
			screen.DrawTrianglesShader(vs, []uint16{0, 1, 2}, ss.lightShader, lightOpt)
		}
	} else {
		// Fallback to original white cone rendering
		//lightOpt := &ebiten.DrawTrianglesOptions{Address: ebiten.AddressRepeat}
		//for i := 0; i < len(finalRays)-1; i++ {
		//	a, b := finalRays[i], finalRays[i+1]
		//	vs := rayVertices(lightX, lightY, b.X2, b.Y2, a.X2, a.Y2)
		//	for j := range vs {
		//		vs[j].ColorA = 0.3
		//	}
		//	screen.DrawTriangles(vs, []uint16{0, 1, 2}, ss.triangleImage, lightOpt)
		//}
	}

	// 11) Optional debug rays
	if showRays {
		for _, r := range finalRays {
			vector.StrokeLine(
				screen,
				float32(r.X1), float32(r.Y1),
				float32(r.X2), float32(r.Y2),
				1, color.RGBA{R: 255, G: 255, B: 0, A: 150},
				true,
			)
		}
	}
}

// NewShadowSystem creates a new shadow system
func NewShadowSystem(screenWidth, screenHeight int) *ShadowSystem {
	shadowImage := ebiten.NewImage(screenWidth, screenHeight)
	triangleImage := ebiten.NewImage(screenWidth, screenHeight)
	triangleImage.Fill(color.White)

	// Initialize the stepped light shader
	var lightShader *ebiten.Shader
	if shader, err := ebiten.NewShader(steppedLightShader); err == nil {
		lightShader = shader
	}

	return &ShadowSystem{
		shadowImage:   shadowImage,
		triangleImage: triangleImage,
		lightShader:   lightShader,
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
	}
}
