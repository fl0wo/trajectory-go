package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/math/f32"
	"image/color"
	"math"
)

// getAdaptiveFov calculates adaptive field of view based on distance
func (r *Renderer) getAdaptiveFov(lightDirection f32.Vec2, lightPos f32.Vec2) float64 {
	lightDistanceFromCamera := math.Hypot(float64(lightDirection[0]-lightPos[0]), float64(lightDirection[1]-lightPos[1]))
	// Clamp FOV to reasonable range
	const minFov = 10.0
	const maxFov = 90.0
	fov := fovLight / (lightDistanceFromCamera / 250.0)
	if fov < minFov {
		return minFov
	}
	if fov > maxFov {
		return maxFov
	}
	return fov
}

// drawLineWithShader draws a line using triangles with the specified shader
func (r *Renderer) drawLineWithShader(screen *ebiten.Image, start, end f32.Vec2, width float32, color color.RGBA, shader *ebiten.Shader, uniforms map[string]any) {
	// Create line using vector path and triangulation
	var path vector.Path
	path.MoveTo(start[0], start[1])
	path.LineTo(end[0], end[1])

	strokeOp := &vector.StrokeOptions{Width: width}

	// Create vertices and indices for the line
	var vs []ebiten.Vertex
	var is []uint16
	vs, is = path.AppendVerticesAndIndicesForStroke(vs, is, strokeOp)

	// Set color for all vertices
	rr, gg, bb, aa := color.RGBA()
	cr := float32(rr) / 0xffff
	cg := float32(gg) / 0xffff
	cb := float32(bb) / 0xffff
	ca := float32(aa) / 0xffff
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = cr
		vs[i].ColorG = cg
		vs[i].ColorB = cb
		vs[i].ColorA = ca
	}

	// Draw with shader
	op := &ebiten.DrawTrianglesShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = r.whiteTexture
	screen.DrawTrianglesShader(vs, is, shader, op)
}

// drawArrowHeadWithShader draws an arrowhead using triangles with the specified shader
func (r *Renderer) drawArrowHeadWithShader(screen *ebiten.Image, start, end f32.Vec2, color color.RGBA, shader *ebiten.Shader, uniforms map[string]any) {
	// Calculate arrow direction
	dx := end[0] - start[0]
	dy := end[1] - start[1]
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if length == 0 {
		return
	}

	// Normalize direction vector
	unitX := dx / length
	unitY := dy / length

	// Arrow head size
	arrowLength := float32(15.0)
	arrowWidth := float32(8.0)

	// Calculate perpendicular vector
	perpX := -unitY
	perpY := unitX

	// Calculate arrow head points
	backX := end[0] - unitX*arrowLength
	backY := end[1] - unitY*arrowLength

	leftX := backX + perpX*arrowWidth
	leftY := backY + perpY*arrowWidth

	rightX := backX - perpX*arrowWidth
	rightY := backY - perpY*arrowWidth

	// Draw arrow head lines with shader
	r.drawLineWithShader(screen, end, f32.Vec2{leftX, leftY}, 3, color, shader, uniforms)
	r.drawLineWithShader(screen, end, f32.Vec2{rightX, rightY}, 3, color, shader, uniforms)
}
