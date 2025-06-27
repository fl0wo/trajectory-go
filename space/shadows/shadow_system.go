package shadows

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/math/f32"
	"image/color"
	"math"
	"sort"
)

// ShadowSystem handles raycasting and shadow rendering for the space game
type ShadowSystem struct {
	shadowImage   *ebiten.Image
	triangleImage *ebiten.Image
	screenWidth   int
	screenHeight  int
}

// NewShadowSystem creates a new shadow system
func NewShadowSystem(screenWidth, screenHeight int) *ShadowSystem {
	shadowImage := ebiten.NewImage(screenWidth, screenHeight)
	triangleImage := ebiten.NewImage(screenWidth, screenHeight)
	triangleImage.Fill(color.White)

	return &ShadowSystem{
		shadowImage:   shadowImage,
		triangleImage: triangleImage,
		screenWidth:   screenWidth,
		screenHeight:  screenHeight,
	}
}

// line represents a line segment for raycasting calculations
type line struct {
	X1, Y1, X2, Y2 float64
}

// angle calculates the angle of the line from start to end point
func (l *line) angle() float64 {
	return math.Atan2(l.Y2-l.Y1, l.X2-l.X1)
}

// newRay creates a ray starting from x,y with given length and angle
func newRay(x, y, length, angle float64) line {
	return line{
		X1: x,
		Y1: y,
		X2: x + length*math.Cos(angle),
		Y2: y + length*math.Sin(angle),
	}
}

// intersection calculates the intersection of two lines
func intersection(l1, l2 line) (float64, float64, bool) {
	// https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection#Given_two_points_on_each_line
	denom := (l1.X1-l1.X2)*(l2.Y1-l2.Y2) - (l1.Y1-l1.Y2)*(l2.X1-l2.X2)
	tNum := (l1.X1-l2.X1)*(l2.Y1-l2.Y2) - (l1.Y1-l2.Y1)*(l2.X1-l2.X2)
	uNum := -((l1.X1-l1.X2)*(l1.Y1-l2.Y1) - (l1.Y1-l1.Y2)*(l1.X1-l2.X1))

	if denom == 0 {
		return 0, 0, false
	}

	t := tNum / denom
	if t > 1 || t < 0 {
		return 0, 0, false
	}

	u := uNum / denom
	if u > 1 || u < 0 {
		return 0, 0, false
	}

	x := l1.X1 + t*(l1.X2-l1.X1)
	y := l1.Y1 + t*(l1.Y2-l1.Y1)
	return x, y, true
}

// circleToLines converts a circle (celestial body) to line segments for raycasting
func circleToLines(center f32.Vec2, radius float32, segments int) []line {
	var lines []line
	angleStep := 2 * math.Pi / float64(segments)

	for i := 0; i < segments; i++ {
		angle1 := float64(i) * angleStep
		angle2 := float64(i+1) * angleStep

		x1 := float64(center[0]) + float64(radius)*math.Cos(angle1)
		y1 := float64(center[1]) + float64(radius)*math.Sin(angle1)
		x2 := float64(center[0]) + float64(radius)*math.Cos(angle2)
		y2 := float64(center[1]) + float64(radius)*math.Sin(angle2)

		lines = append(lines, line{X1: x1, Y1: y1, X2: x2, Y2: y2})
	}

	return lines
}

// getAllOccluderPoints extracts all relevant points from celestial bodies and asteroids for raycasting
func getAllOccluderPoints(celestialPositions []f32.Vec2, celestialRadii []float32, asteroidPositions []f32.Vec2, asteroidRadii []float32) [][2]float64 {
	var points [][2]float64

	// Add points from celestial bodies (use fewer segments for performance)
	for i, pos := range celestialPositions {
		radius := celestialRadii[i]
		segments := 16 // 16 segments per circle for good balance of quality vs performance

		for j := 0; j < segments; j++ {
			angle := 2 * math.Pi * float64(j) / float64(segments)
			x := float64(pos[0]) + float64(radius)*math.Cos(angle)
			y := float64(pos[1]) + float64(radius)*math.Sin(angle)
			points = append(points, [2]float64{x, y})
		}
	}

	// Add points from asteroids (fewer segments since they're smaller)
	for i, pos := range asteroidPositions {
		radius := asteroidRadii[i]
		segments := 8 // Fewer segments for smaller asteroids

		for j := 0; j < segments; j++ {
			angle := 2 * math.Pi * float64(j) / float64(segments)
			x := float64(pos[0]) + float64(radius)*math.Cos(angle)
			y := float64(pos[1]) + float64(radius)*math.Sin(angle)
			points = append(points, [2]float64{x, y})
		}
	}

	return points
}

// getAllOccluderLines creates line segments from all celestial bodies and asteroids
func getAllOccluderLines(celestialPositions []f32.Vec2, celestialRadii []float32, asteroidPositions []f32.Vec2, asteroidRadii []float32) []line {
	var lines []line

	// Add lines from celestial bodies
	for i, pos := range celestialPositions {
		radius := celestialRadii[i]
		bodyLines := circleToLines(pos, radius, 16) // 16 segments per celestial body
		lines = append(lines, bodyLines...)
	}

	// Add lines from asteroids
	for i, pos := range asteroidPositions {
		radius := asteroidRadii[i]
		asteroidLines := circleToLines(pos, radius, 8) // 8 segments per asteroid
		lines = append(lines, asteroidLines...)
	}

	return lines
}

// rayCasting performs raycasting from a light source to create shadow rays
func rayCasting(lightX, lightY float64, celestialPositions []f32.Vec2, celestialRadii []float32, asteroidPositions []f32.Vec2, asteroidRadii []float32) []line {
	const rayLength = 3000 // Large enough to reach screen edges

	// Get all occluder points
	occluderPoints := getAllOccluderPoints(celestialPositions, celestialRadii, asteroidPositions, asteroidRadii)
	occluderLines := getAllOccluderLines(celestialPositions, celestialRadii, asteroidPositions, asteroidRadii)

	var rays []line

	// Cast rays toward each occluder point (with small angle offsets for better shadows)
	for _, point := range occluderPoints {
		l := line{lightX, lightY, point[0], point[1]}
		angle := l.angle()

		// Cast two rays with slight angle offsets to capture shadow edges better
		for _, offset := range []float64{-0.002, 0.002} {
			ray := newRay(lightX, lightY, rayLength, angle+offset)

			// Find closest intersection with any occluder
			var intersectionPoints [][2]float64
			for _, occluderLine := range occluderLines {
				if px, py, ok := intersection(ray, occluderLine); ok {
					intersectionPoints = append(intersectionPoints, [2]float64{px, py})
				}
			}

			// Find the closest intersection point
			if len(intersectionPoints) > 0 {
				minDist := math.Inf(1)
				var closestPoint [2]float64

				for _, p := range intersectionPoints {
					dist := (lightX-p[0])*(lightX-p[0]) + (lightY-p[1])*(lightY-p[1])
					if dist < minDist {
						minDist = dist
						closestPoint = p
					}
				}

				rays = append(rays, line{lightX, lightY, closestPoint[0], closestPoint[1]})
			} else {
				// No intersection, ray goes to full length
				rays = append(rays, ray)
			}
		}
	}

	// Sort rays by angle for proper triangle rendering
	sort.Slice(rays, func(i, j int) bool {
		return rays[i].angle() < rays[j].angle()
	})

	return rays
}

// rayVertices creates vertices for shadow triangles
func rayVertices(x1, y1, x2, y2, x3, y3 float64) []ebiten.Vertex {
	return []ebiten.Vertex{
		{DstX: float32(x1), DstY: float32(y1), SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(x2), DstY: float32(y2), SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(x3), DstY: float32(y3), SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
}

// RenderShadows renders the shadow system to the screen
func (ss *ShadowSystem) RenderShadows(screen *ebiten.Image, lightPos f32.Vec2, celestialPositions []f32.Vec2, celestialRadii []float32, asteroidPositions []f32.Vec2, asteroidRadii []float32, showRays bool) {
	// Skip rendering if no occluders
	if len(celestialPositions) == 0 && len(asteroidPositions) == 0 {
		return
	}

	// Reset shadow image to black (full shadow)
	ss.shadowImage.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 255})

	// Perform raycasting
	rays := rayCasting(float64(lightPos[0]), float64(lightPos[1]), celestialPositions, celestialRadii, asteroidPositions, asteroidRadii)

	// Create illuminated areas by subtracting light triangles from shadow
	opt := &ebiten.DrawTrianglesOptions{}
	opt.Address = ebiten.AddressRepeat
	opt.Blend = ebiten.BlendSourceOut // Subtract mode - removes shadow where light hits

	for i, ray := range rays {
		nextRay := rays[(i+1)%len(rays)]

		// Create triangle between light source and two consecutive rays
		vertices := rayVertices(
			float64(lightPos[0]), float64(lightPos[1]), // Light source
			nextRay.X2, nextRay.Y2, // End of next ray
			ray.X2, ray.Y2, // End of current ray
		)

		// Draw the illuminated triangle (removes shadow)
		ss.shadowImage.DrawTriangles(vertices, []uint16{0, 1, 2}, ss.triangleImage, opt)
	}

	// Render shadow overlay on screen with transparency
	shadowOpt := &ebiten.DrawImageOptions{}
	shadowOpt.ColorScale.ScaleAlpha(0.7) // 70% opacity for shadows
	screen.DrawImage(ss.shadowImage, shadowOpt)

	// Optionally draw rays for debugging
	if showRays {
		for _, ray := range rays {
			vector.StrokeLine(screen,
				float32(ray.X1), float32(ray.Y1),
				float32(ray.X2), float32(ray.Y2),
				1, color.RGBA{R: 255, G: 255, B: 0, A: 150}, true)
		}
	}
}
