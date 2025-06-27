package util

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// StrokeDashedCircleWithShader strokes a circle with center (cx,cy), radius r,
// strokeWidth, clr, and a repeating [dashLength on, gapLength off] pattern.
// If shader is provided, it will be applied to the rendered circle.
func StrokeDashedCircleWithShader(
	dst *ebiten.Image,
	cx, cy, r float32,
	strokeWidth float32,
	clr color.Color,
	dashLength, gapLength float32,
	antialias bool,
	shader *ebiten.Shader,
	uniforms map[string]any,
) {
	if shader == nil {
		// Fallback to regular dashed circle if no shader
		StrokeDashedCircle(dst, cx, cy, r, strokeWidth, clr, dashLength, gapLength, antialias)
		return
	}

	// Create a temporary image to render the dashed circle
	bounds := dst.Bounds()
	tempImage := ebiten.NewImage(bounds.Dx(), bounds.Dy())
	defer tempImage.Deallocate()

	// Render the dashed circle to temporary image first
	StrokeDashedCircle(tempImage, cx, cy, r, strokeWidth, clr, dashLength, gapLength, antialias)

	// Apply shader to the temporary image and draw to destination
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = tempImage

	dst.DrawRectShader(bounds.Dx(), bounds.Dy(), shader, op)
}

// StrokeDashedCircleTrianglesWithShader creates triangle vertices for a dashed circle
// and applies a shader to them directly (more efficient approach)
func StrokeDashedCircleTrianglesWithShader(
	dst *ebiten.Image,
	cx, cy, r float32,
	strokeWidth float32,
	clr color.Color,
	dashLength, gapLength float32,
	antialias bool,
	shader *ebiten.Shader,
	uniforms map[string]any,
) {
	if shader == nil {
		// Fallback to regular dashed circle if no shader
		StrokeDashedCircle(dst, cx, cy, r, strokeWidth, clr, dashLength, gapLength, antialias)
		return
	}

	// total circumference
	circ := 2 * math.Pi * float64(r)
	// convert desired dash/gap lengths into angles:
	dashAngle := dashLength / float32(circ) * 2 * math.Pi
	gapAngle := gapLength / float32(circ) * 2 * math.Pi

	strokeOp := &vector.StrokeOptions{Width: strokeWidth}

	useCachedVerticesAndIndices(func(vs []ebiten.Vertex, is []uint16) ([]ebiten.Vertex, []uint16) {
		// for each dash, build a little arc and append its stroke
		for start := float32(0); start < 2*math.Pi; start += dashAngle + gapAngle {
			end := start + dashAngle
			if end > 2*math.Pi {
				end = 2 * math.Pi
			}
			var path vector.Path
			path.Arc(cx, cy, r, start, end, vector.Clockwise)
			path.Close()
			vs, is = path.AppendVerticesAndIndicesForStroke(vs, is, strokeOp)
		}

		// Apply color to vertices
		r, g, b, a := clr.RGBA()
		for i := range vs {
			vs[i].SrcX = 1
			vs[i].SrcY = 1
			vs[i].ColorR = float32(r) / 0xffff
			vs[i].ColorG = float32(g) / 0xffff
			vs[i].ColorB = float32(b) / 0xffff
			vs[i].ColorA = float32(a) / 0xffff
		}

		// Draw with shader
		op := &ebiten.DrawTrianglesShaderOptions{}
		op.Uniforms = uniforms
		op.Images[0] = whiteSubImage
		dst.DrawTrianglesShader(vs, is, shader, op)

		return vs, is
	})
}
