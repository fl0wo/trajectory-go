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

// clipRay fires a single ray at 'angle' and stops at the first occluder (or full length), ignoring specified portal
func clipRay(
	lightX, lightY, length, angle float64,
	occluderLines []line,
	occluderTypes []string,
	occluderIndices []int,
	ignorePortalIdx int,
) line {
	r := newRay(lightX, lightY, length, angle)
	var hits [][2]float64
	for i, ol := range occluderLines {
		if occluderTypes[i] == "portal" && occluderIndices[i] == ignorePortalIdx {
			continue // Skip intersections with the ignored portal
		}
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

// getAllOccluderPoints includes celestials, asteroids, and all portals
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
	for i, pos := range portalPositions {
		r := portalRadii[i]
		for j := 0; j < 12; j++ {
			θ := 2 * math.Pi * float64(j) / 12
			pts = append(pts, [2]float64{
				float64(pos[0]) + float64(r)*math.Cos(θ),
				float64(pos[1]) + float64(r)*math.Sin(θ),
			})
		}
	}
	return pts
}

// getAllOccluderLines includes celestials, asteroids, and all portals, with type and index tracking
func getAllOccluderLines(
	celestialPositions []f32.Vec2, celestialRadii []float32,
	asteroidPositions []f32.Vec2, asteroidRadii []float32,
	portalPositions []f32.Vec2, portalRadii []float32,
) (lines []line, types []string, indices []int) {
	for i, pos := range celestialPositions {
		cLines := circleToLines(pos, celestialRadii[i], 16)
		lines = append(lines, cLines...)
		for range cLines {
			types = append(types, "celestial")
			indices = append(indices, i)
		}
	}
	for i, pos := range asteroidPositions {
		aLines := circleToLines(pos, asteroidRadii[i], 8)
		lines = append(lines, aLines...)
		for range aLines {
			types = append(types, "asteroid")
			indices = append(indices, i)
		}
	}
	for i, pos := range portalPositions {
		pLines := circleToLines(pos, portalRadii[i], 12)
		lines = append(lines, pLines...)
		for range pLines {
			types = append(types, "portal")
			indices = append(indices, i)
		}
	}
	return lines, types, indices
}

// getAllPortalLines returns line segments for all portals for propagation checks
func getAllPortalLines(portalPositions []f32.Vec2, portalRadii []float32) [][]line {
	var allPortalLines [][]line
	for i, pos := range portalPositions {
		lines := circleToLines(pos, portalRadii[i], 12)
		allPortalLines = append(allPortalLines, lines)
	}
	return allPortalLines
}

// findPortalHit checks if a ray hits any portal and returns the hit point and portal index
func findPortalHit(r line, allPortalLines [][]line) (float64, float64, int, bool) {
	var minD float64 = math.Inf(1)
	var hitX, hitY float64
	var hitPortal int = -1
	for i, portalLines := range allPortalLines {
		for _, pl := range portalLines {
			if x, y, ok := intersection(r, pl); ok {
				d := (r.X1-x)*(r.X1-x) + (r.Y1-y)*(r.Y1-y)
				if d < minD {
					minD = d
					hitX, hitY = x, y
					hitPortal = i
				}
			}
		}
	}
	if hitPortal != -1 {
		return hitX, hitY, hitPortal, true
	}
	return 0, 0, -1, false
}

// rayCastingWithPortals casts rays from the light source and conducts light through all portal pairs using XOR
func rayCastingWithPortals(
	lightX, lightY float64,
	celestialPositions []f32.Vec2, celestialRadii []float32,
	asteroidPositions []f32.Vec2, asteroidRadii []float32,
	portalPositions []f32.Vec2, portalRadii []float32,
) (playerRays []line, portalRaySets map[int][]line) {
	const rayLength = 3000
	occluderLines, occluderTypes, occluderIndices := getAllOccluderLines(celestialPositions, celestialRadii, asteroidPositions, asteroidRadii, portalPositions, portalRadii)
	allPortalLines := getAllPortalLines(portalPositions, portalRadii)
	points := getAllOccluderPoints(celestialPositions, celestialRadii, asteroidPositions, asteroidRadii, portalPositions, portalRadii)

	portalRaySets = make(map[int][]line)

	// Cast rays from the player to all occluder points
	for _, pt := range points {
		ang := math.Atan2(float64(pt[1])-lightY, float64(pt[0])-lightX)
		for _, off := range []float64{-0.002, 0.002} {
			rayAngle := ang + off
			r := newRay(lightX, lightY, rayLength, rayAngle)

			// Find closest occluder intersection (including portals)
			var closestOccluderD = math.Inf(1)
			var closestOccluder [2]float64
			var occluderHit bool
			for _, ol := range occluderLines {
				if x, y, ok := intersection(r, ol); ok {
					d := (lightX-x)*(lightX-x) + (lightY-y)*(lightY-y)
					if d < closestOccluderD {
						closestOccluderD = d
						closestOccluder = [2]float64{x, y}
						occluderHit = true
					}
				}
			}

			// Find closest portal intersection for propagation
			px, py, hitPortalIdx, hitPortal := findPortalHit(r, allPortalLines)
			var portalHitD float64
			if hitPortal {
				portalHitD = (lightX-px)*(lightX-px) + (lightY-py)*(lightY-py)
			}

			// Determine action based on closest hit
			if hitPortal && (!occluderHit || portalHitD <= closestOccluderD) {
				// Portal hit first; clip player ray and propagate to paired portal
				playerRays = append(playerRays, line{lightX, lightY, px, py})
				pairedIdx := hitPortalIdx ^ 1
				if pairedIdx < len(portalPositions) {
					pairedX, pairedY := float64(portalPositions[pairedIdx][0]), float64(portalPositions[pairedIdx][1])
					// Propagate ray from paired portal, ignoring its own lines
					portalRay := clipRay(pairedX, pairedY, rayLength, rayAngle, occluderLines, occluderTypes, occluderIndices, pairedIdx)
					portalRaySets[pairedIdx] = append(portalRaySets[pairedIdx], portalRay)
				}
			} else if occluderHit {
				// Occluder (or portal as occluder) hit first; clip player ray
				playerRays = append(playerRays, line{lightX, lightY, closestOccluder[0], closestOccluder[1]})
			} else {
				// No hits; use full ray length
				playerRays = append(playerRays, r)
			}
		}
	}

	// Sort player rays by angle
	sort.Slice(playerRays, func(i, j int) bool {
		return playerRays[i].angle() < playerRays[j].angle()
	})

	// Sort portal rays for each portal
	for _, rays := range portalRaySets {
		sort.Slice(rays, func(i, j int) bool {
			return rays[i].angle() < rays[j].angle()
		})
	}

	return playerRays, portalRaySets
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

// RenderShadows renders a spotlight with light propagation through all portal pairs
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
	portalRotations []float32, // how much the portal is rotated in radians
	showRays bool,
) {
	if len(celestialPositions) == 0 && len(asteroidPositions) == 0 && len(portalPositions) == 0 {
		return
	}

	// Fill screen with shadow
	ss.shadowImage.Fill(colors.ShadowBackground)

	lightX, lightY := float64(lightPos[0]), float64(lightPos[1])

	// Check if light source is inside any portal
	for i, pos := range portalPositions {
		dx := lightX - float64(pos[0])
		dy := lightY - float64(pos[1])
		distance := math.Sqrt(dx*dx + dy*dy)
		if distance <= float64(portalRadii[i]) {
			// Light source is inside a portal; render only the shadow (no light)
			shadowOpt := &ebiten.DrawImageOptions{}
			shadowOpt.ColorScale.ScaleAlpha(0.7)
			screen.DrawImage(ss.shadowImage, shadowOpt)
			return
		}
	}

	// Proceed with normal light rendering if not inside a portal
	// Get rays for player and portals
	playerRays, portalRaySets := rayCastingWithPortals(
		lightX, lightY,
		celestialPositions, celestialRadii,
		asteroidPositions, asteroidRadii,
		portalPositions, portalRadii,
	)

	// Calculate spotlight center and half-FOV
	base := math.Atan2(float64(winPos[1])-lightY, float64(winPos[0])-lightX)
	half := (fovAngle * math.Pi / 180) / 2

	// Filter player rays within the FOV cone
	var interiorPlayer []line
	for _, r := range playerRays {
		if math.Abs(normalizeAngle(r.angle()-base)) <= half {
			interiorPlayer = append(interiorPlayer, r)
		}
	}
	sort.Slice(interiorPlayer, func(i, j int) bool {
		return normalizeAngle(interiorPlayer[i].angle()-base) < normalizeAngle(interiorPlayer[j].angle()-base)
	})

	// Build occluder lines for boundary clipping
	occluderLines, occluderTypes, occluderIndices := getAllOccluderLines(celestialPositions, celestialRadii, asteroidPositions, asteroidRadii, portalPositions, portalRadii)
	diag := math.Hypot(float64(ss.screenWidth), float64(ss.screenHeight))

	// Clip boundary rays for player (no portal ignore for player rays)
	left := clipRay(lightX, lightY, diag, base-half, occluderLines, occluderTypes, occluderIndices, -1)
	right := clipRay(lightX, lightY, diag, base+half, occluderLines, occluderTypes, occluderIndices, -1)

	// Assemble final player ray list
	finalPlayerRays := []line{left}
	finalPlayerRays = append(finalPlayerRays, interiorPlayer...)
	finalPlayerRays = append(finalPlayerRays, right)

	// Carve out player light polygon
	triOpt := &ebiten.DrawTrianglesOptions{Address: ebiten.AddressRepeat}
	triOpt.Blend = ebiten.BlendSourceOut
	for i := 0; i < len(finalPlayerRays)-1; i++ {
		a, b := finalPlayerRays[i], finalPlayerRays[i+1]
		vs := rayVertices(lightX, lightY, b.X2, b.Y2, a.X2, a.Y2)
		ss.shadowImage.DrawTriangles(vs, []uint16{0, 1, 2}, ss.triangleImage, triOpt)
	}

	// Carve out portal light polygons for each portal with rays
	for portalIdx, rays := range portalRaySets {
		if len(rays) > 0 {
			portalX, portalY := float64(portalPositions[portalIdx][0]), float64(portalPositions[portalIdx][1])
			for i := 0; i < len(rays)-1; i++ {
				a, b := rays[i], rays[i+1]
				vs := rayVertices(portalX, portalY, b.X2, b.Y2, a.X2, a.Y2)
				ss.shadowImage.DrawTriangles(vs, []uint16{0, 1, 2}, ss.triangleImage, triOpt)
			}
		}
	}

	// Overlay the shadow on the screen
	shadowOpt := &ebiten.DrawImageOptions{}
	shadowOpt.ColorScale.ScaleAlpha(0.7)
	screen.DrawImage(ss.shadowImage, shadowOpt)

	// Draw light cones with shader (if available)
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

		// Portal light cones
		for portalIdx, rays := range portalRaySets {
			if len(rays) > 0 {
				portalX, portalY := float64(portalPositions[portalIdx][0]), float64(portalPositions[portalIdx][1])
				lightOpt.Uniforms["LightPos"] = []float32{float32(portalX), float32(portalY)}
				lightOpt.Uniforms["Fov"] = float64(360) // Full illumination from portal
				for i := 0; i < len(rays)-1; i++ {
					a, b := rays[i], rays[i+1]
					vs := rayVertices(portalX, portalY, b.X2, b.Y2, a.X2, a.Y2)
					screen.DrawTrianglesShader(vs, []uint16{0, 1, 2}, ss.lightShader, lightOpt)
				}
			}
		}
	}

	// Debug rays
	if showRays {
		for _, r := range finalPlayerRays {
			vector.StrokeLine(screen, float32(r.X1), float32(r.Y1), float32(r.X2), float32(r.Y2), 1, colors.DebugRay, true)
		}
		for _, rays := range portalRaySets {
			for _, r := range rays {
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
