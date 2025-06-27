package util

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	whiteImage    = ebiten.NewImage(3, 3)
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

var (
	cachedVertices []ebiten.Vertex
	cachedIndices  []uint16
	cacheM         sync.Mutex
)

func useCachedVerticesAndIndices(fn func([]ebiten.Vertex, []uint16) (vs []ebiten.Vertex, is []uint16)) {
	cacheM.Lock()
	defer cacheM.Unlock()
	cachedVertices, cachedIndices = fn(cachedVertices[:0], cachedIndices[:0])
}

func init() {
	b := whiteImage.Bounds()
	pix := make([]byte, 4*b.Dx()*b.Dy())
	for i := range pix {
		pix[i] = 0xff
	}
	// This is hacky, but WritePixels is better than Fill in term of automatic texture packing.
	whiteImage.WritePixels(pix)
}

func drawVerticesForUtil(dst *ebiten.Image, vs []ebiten.Vertex, is []uint16, clr color.Color, antialias bool) {
	r, g, b, a := clr.RGBA()
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(r) / 0xffff
		vs[i].ColorG = float32(g) / 0xffff
		vs[i].ColorB = float32(b) / 0xffff
		vs[i].ColorA = float32(a) / 0xffff
	}

	op := &ebiten.DrawTrianglesOptions{}
	op.ColorScaleMode = ebiten.ColorScaleModePremultipliedAlpha
	op.AntiAlias = antialias
	dst.DrawTriangles(vs, is, whiteSubImage, op)
}

// StrokeDashedCircle strokes a circle with center (cx,cy), radius r,
// strokeWidth, clr, and a repeating [dashLength on, gapLength off] pattern.
// dashLength & gapLength are in the same units as r (world‐space).
func StrokeDashedCircle(
	dst *ebiten.Image,
	cx, cy, r float32,
	strokeWidth float32,
	clr color.Color,
	dashLength, gapLength float32,
	antialias bool,
) {
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
			// next dash will start at end + gapAngle automatically
		}
		drawVerticesForUtil(dst, vs, is, clr, antialias)
		return vs, is
	})
}
