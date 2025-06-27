package util

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// StrokeDashedCircleTrianglesWithShader draws a dashed circle whose dashes
// are rotated by `rotation` (in radians) around the center.
func StrokeDashedCircleTrianglesWithShader(
	dst *ebiten.Image,
	cx, cy, r float32,
	strokeWidth float32,
	clr color.Color,
	dashLength, gapLength float32,
	antialias bool,
	shader *ebiten.Shader,
	uniforms map[string]any,
	rotation float32, // Amount to rotate the dash pattern
) {
	if shader == nil {
		// Fallback to CPU‐only version if no shader
		StrokeDashedCircle(dst, cx, cy, r, strokeWidth, clr, dashLength, gapLength, antialias)
		return
	}

	// (Optional) pass rotation into the shader too, for any per‐fragment effects
	uniforms["Rotation"] = rotation

	// Compute the dash/gap angles
	circ := 2 * math.Pi * float64(r)
	dashAngle := dashLength / float32(circ) * 2 * math.Pi
	gapAngle := gapLength / float32(circ) * 2 * math.Pi

	strokeOp := &vector.StrokeOptions{Width: strokeWidth}

	useCachedVerticesAndIndices(func(vs []ebiten.Vertex, is []uint16) ([]ebiten.Vertex, []uint16) {
		// Build each dash, but shift its start/end by `rotation`
		angle := rotation
		for angle < rotation+2*math.Pi {
			start := angle
			end := angle + dashAngle
			if end > rotation+2*math.Pi {
				end = rotation + 2*math.Pi
			}
			var path vector.Path
			path.Arc(cx, cy, r, start, end, vector.Clockwise)
			path.Close()
			vs, is = path.AppendVerticesAndIndicesForStroke(vs, is, strokeOp)
			angle += dashAngle + gapAngle
		}

		// Color all vertices
		rr, gg, bb, aa := clr.RGBA()
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
		op.Images[0] = whiteSubImage
		dst.DrawTrianglesShader(vs, is, shader, op)

		return vs, is
	})
}
