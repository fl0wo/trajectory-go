package shadows

import (
	_ "embed"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/you/trajectory/space/colors"
	"golang.org/x/image/math/f32"
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

// getAllOccluderPoints includes only celestials and asteroids (not portals)
func getAllOccluderPoints(
	celestialPositions []f32.Vec2, celestialRadii []float32,
	asteroidPositions []f32.Vec2, asteroidRadii []float32,
	portalPositions []f32.Vec2, portalRadii []float32,
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

// getAllOccluderLines includes only celestials and asteroids
func getAllOccluderLines(
	celestialPositions []f32.Vec2, celestialRadii []float32,
	asteroidPositions []f32.Vec2, asteroidRadii []float32,
	portalPositions []f32.Vec2, portalRadii []float32,
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

// getPortalLines returns line segments for portal 0 only
func getPortalLines(portalPositions []f32.Vec2, portalRadii []float32) []line {
	if len(portalPositions) == 0 {
		return nil
	}
	return circleToLines(portalPositions[0], portalRadii[0], 12)
}

// findPortalHit checks if a ray hits portal 0
func findPortalHit(r line, portalLines []line) (float64, float64, bool) {
	var minD float64 = math.Inf(1)
	var hitX, hitY float64
	hit := false
	for _, pl := range portalLines {
		if x, y, ok := intersection(r, pl); ok {
			d := (r.X1-x)*(r.X1-x) + (r.Y1-y)*(r.Y1-y)
			if d < minD {
				minD = d
				hitX, hitY = x, y
				hit = true
			}
		}
	}
	return hitX, hitY, hit
}

// rayCastingWithPortals casts rays from the light source, conducting light from portal 0 to portal 1
func rayCastingWithPortals(
	lightX, lightY float64,
	celestialPositions []f32.Vec2, celestialRadii []float32,
	asteroidPositions []f32.Vec2, asteroidRadii []float32,
	portalPositions []f32.Vec2, portalRadii []float32,
) (playerRays []line, portalRays []line) {
	const rayLength = 3000
	occluderLines := getAllOccluderLines(celestialPositions, celestialRadii, asteroidPositions, asteroidRadii, portalPositions, portalRadii)
	portalLines := getPortalLines(portalPositions, portalRadii)
	points := getAllOccluderPoints(celestialPositions, celestialRadii, asteroidPositions, asteroidRadii, portalPositions, portalRadii)

	// Cast rays from the player
	for _, pt := range points {
		ang := math.Atan2(float64(pt[1])-lightY, float64(pt[0])-lightX)
		for _, off := range []float64{-0.002, 0.002} {
			rayAngle := ang + off
			r := newRay(lightX, lightY, rayLength, rayAngle)

			// Find intersections with occluders
			var occluderHits [][2]float64
			for _, ol := range occluderLines {
				if x, y, ok := intersection(r, ol); ok {
					occluderHits = append(occluderHits, [2]float64{x, y})
				}
			}

			// Find intersection with portal 0
			var portalHit [2]float64
			var portalHitOk bool
			var portalHitD float64
			if len(portalLines) > 0 {
				px, py, hit := findPortalHit(r, portalLines)
				if hit {
					portalHit = [2]float64{px, py}
					portalHitOk = true
					portalHitD = (lightX-px)*(lightX-px) + (lightY-py)*(lightY-py)
				}
			}

			if portalHitOk {
				// Determine if portal hit is closer than occluder hits
				closestOccluderD := math.Inf(1)
				for _, oh := range occluderHits {
					d := (lightX-oh[0])*(lightX-oh[0]) + (lightY-oh[1])*(lightY-oh[1])
					if d < closestOccluderD {
						closestOccluderD = d
					}
				}
				if portalHitD < closestOccluderD {
					// Ray hits portal 0 first; clip it there and propagate from portal 1
					playerRays = append(playerRays, line{lightX, lightY, portalHit[0], portalHit[1]})
					if len(portalPositions) > 1 {
						portal1X, portal1Y := float64(portalPositions[1][0]), float64(portalPositions[1][1])
						portalRay := clipRay(portal1X, portal1Y, rayLength, rayAngle, occluderLines)
						portalRays = append(portalRays, portalRay)
					}
				} else {
					// Occluder is closer; clip ray there
					minD := math.Inf(1)
					var closest [2]float64
					for _, oh := range occluderHits {
						d := (lightX-oh[0])*(lightX-oh[0]) + (lightY-oh[1])*(lightY-oh[1])
						if d < minD {
							minD = d
							closest = oh
						}
					}
					playerRays = append(playerRays, line{lightX, lightY, closest[0], closest[1]})
				}
			} else if len(occluderHits) > 0 {
				// No portal hit; clip against occluders
				minD := math.Inf(1)
				var closest [2]float64
				for _, oh := range occluderHits {
					d := (lightX-oh[0])*(lightX-oh[0]) + (lightY-oh[1])*(lightY-oh[1])
					if d < minD {
						minD = d
						closest = oh
					}
				}
				playerRays = append(playerRays, line{lightX, lightY, closest[0], closest[1]})
			} else {
				playerRays = append(playerRays, r)
			}
		}
	}

	// Sort rays by angle for proper polygon construction
	sort.Slice(playerRays, func(i, j int) bool {
		return playerRays[i].angle() < playerRays[j].angle()
	})
	if len(portalRays) > 0 {
		sort.Slice(portalRays, func(i, j int) bool {
			return portalRays[i].angle() < portalRays[j].angle()
		})
	}

	return playerRays, portalRays
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

func rayVertices(x1, y1, x2, y2, x3, y3 float64) []ebiten.Vertex {
	return []ebiten.Vertex{
		{DstX: float32(x1), DstY: float32(y1), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(x2), DstY: float32(y2), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(x3), DstY: float32(y3), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
}

// RenderShadows renders a spotlight with light propagation from portal 0 to portal 1
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
	portalPositions []f32.Vec2,
	portalRadii []float32,
	showRays bool,
) {
	if len(celestialPositions) == 0 && len(asteroidPositions) == 0 {
		return
	}

	// 1) Fill screen with shadow
	ss.shadowImage.Fill(colors.ShadowBackground)

	lightX, lightY := float64(lightPos[0]), float64(lightPos[1])

	// 2) Get rays for player and portal
	playerRays, portalRays := rayCastingWithPortals(
		lightX, lightY,
		celestialPositions, celestialRadii,
		asteroidPositions, asteroidRadii,
		portalPositions, portalRadii,
	)

	// 3) Calculate spotlight center and half-FOV
	base := math.Atan2(float64(winPos[1])-lightY, float64(winPos[0])-lightX)
	half := (fovAngle * math.Pi / 180) / 2

	// 4) Filter player rays within the FOV cone
	var interiorPlayer []line
	for _, r := range playerRays {
		if math.Abs(normalizeAngle(r.angle()-base)) <= half {
			interiorPlayer = append(interiorPlayer, r)
		}
	}
	sort.Slice(interiorPlayer, func(i, j int) bool {
		return normalizeAngle(interiorPlayer[i].angle()-base) < normalizeAngle(interiorPlayer[j].angle()-base)
	})

	// 5) Build occluder lines for boundary clipping
	occluderLines := getAllOccluderLines(celestialPositions, celestialRadii, asteroidPositions, asteroidRadii, portalPositions, portalRadii)
	diag := math.Hypot(float64(ss.screenWidth), float64(ss.screenHeight))

	// 6) Clip boundary rays for player
	left := clipRay(lightX, lightY, diag, base-half, occluderLines)
	right := clipRay(lightX, lightY, diag, base+half, occluderLines)

	// 7) Assemble final player ray list
	finalPlayerRays := []line{left}
	finalPlayerRays = append(finalPlayerRays, interiorPlayer...)
	finalPlayerRays = append(finalPlayerRays, right)

	// 8) Carve out player light polygon
	triOpt := &ebiten.DrawTrianglesOptions{Address: ebiten.AddressRepeat}
	triOpt.Blend = ebiten.BlendSourceOut
	for i := 0; i < len(finalPlayerRays)-1; i++ {
		a, b := finalPlayerRays[i], finalPlayerRays[i+1]
		vs := rayVertices(lightX, lightY, b.X2, b.Y2, a.X2, a.Y2)
		ss.shadowImage.DrawTriangles(vs, []uint16{0, 1, 2}, ss.triangleImage, triOpt)
	}

	// 9) Carve out portal light polygon if applicable
	if len(portalRays) > 0 && len(portalPositions) > 1 {
		portal1X, portal1Y := float64(portalPositions[1][0]), float64(portalPositions[1][1])
		for i := 0; i < len(portalRays)-1; i++ {
			a, b := portalRays[i], portalRays[i+1]
			vs := rayVertices(portal1X, portal1Y, b.X2, b.Y2, a.X2, a.Y2)
			ss.shadowImage.DrawTriangles(vs, []uint16{0, 1, 2}, ss.triangleImage, triOpt)
		}
	}

	// 10) Overlay the shadow on the screen
	shadowOpt := &ebiten.DrawImageOptions{}
	shadowOpt.ColorScale.ScaleAlpha(0.7)
	screen.DrawImage(ss.shadowImage, shadowOpt)

	// 11) Draw light cones with shader (if available)
	if ss.lightShader != nil {
		maxDistance := float32(math.Hypot(float64(ss.screenWidth), float64(ss.screenHeight)))
		lightOpt := &ebiten.DrawTrianglesShaderOptions{}
		lightOpt.Uniforms = map[string]any{
			"LightPos":    []float32{float32(lightX), float32(lightY)},
			"MaxDistance": maxDistance,
			"Zoom":        cameraZoom,
			"Fov":         fovAngle,
		}
		lightOpt.Images[0] = ss.triangleImage
		lightOpt.Blend = ebiten.BlendSourceOver

		// Player light cone
		for i := 0; i < len(finalPlayerRays)-1; i++ {
			a, b := finalPlayerRays[i], finalPlayerRays[i+1]
			vs := rayVertices(lightX, lightY, b.X2, b.Y2, a.X2, a.Y2)
			screen.DrawTrianglesShader(vs, []uint16{0, 1, 2}, ss.lightShader, lightOpt)
		}

		// Portal light cone
		if len(portalRays) > 0 && len(portalPositions) > 1 {
			portal1X, portal1Y := float64(portalPositions[1][0]), float64(portalPositions[1][1])
			lightOpt.Uniforms["LightPos"] = []float32{float32(portal1X), float32(portal1Y)}
			lightOpt.Uniforms["Fov"] = float64(360) // Full illumination from portal
			for i := 0; i < len(portalRays)-1; i++ {
				a, b := portalRays[i], portalRays[i+1]
				vs := rayVertices(portal1X, portal1Y, b.X2, b.Y2, a.X2, a.Y2)
				screen.DrawTrianglesShader(vs, []uint16{0, 1, 2}, ss.lightShader, lightOpt)
			}
		}
	}

	// 12) Debug rays
	if showRays {
		for _, r := range finalPlayerRays {
			vector.StrokeLine(screen, float32(r.X1), float32(r.Y1), float32(r.X2), float32(r.Y2), 1, colors.DebugRay, true)
		}
		if len(portalRays) > 0 {
			for _, r := range portalRays {
				vector.StrokeLine(screen, float32(r.X1), float32(r.Y1), float32(r.X2), float32(r.Y2), 1, colors.DebugRay, true)
			}
		}
	}
}

// NewShadowSystem creates a new shadow system
func NewShadowSystem(screenWidth, screenHeight int) *ShadowSystem {
	shadowImage := ebiten.NewImage(screenWidth, screenHeight)
	triangleImage := ebiten.NewImage(screenWidth, screenHeight)
	triangleImage.Fill(colors.TriangleImage)

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
