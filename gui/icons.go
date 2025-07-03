package gui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawSettingsIcon draws a circular settings icon (gear) at the specified position
func DrawSettingsIcon(screen *ebiten.Image, x, y, size float32, iconColor color.RGBA) {
	centerX := x + size/2
	centerY := y + size/2

	// Draw outer circle background
	bgColor := color.RGBA{40, 40, 40, 200} // Semi-transparent dark background
	vector.DrawFilledCircle(screen, centerX, centerY, size/2, bgColor, true)

	// Draw gear shape
	outerRadius := size * 0.35
	innerRadius := size * 0.2
	teethCount := 8

	// Draw gear teeth
	for i := 0; i < teethCount; i++ {
		angle1 := float64(i) * 2 * math.Pi / float64(teethCount)
		angle2 := float64(i+1) * 2 * math.Pi / float64(teethCount)

		// Outer points
		x1 := centerX + float32(math.Cos(angle1)*float64(outerRadius))
		y1 := centerY + float32(math.Sin(angle1)*float64(outerRadius))
		x2 := centerX + float32(math.Cos(angle2)*float64(outerRadius))
		y2 := centerY + float32(math.Sin(angle2)*float64(outerRadius))

		// Inner points
		midAngle := (angle1 + angle2) / 2
		x3 := centerX + float32(math.Cos(midAngle)*float64(innerRadius))
		y3 := centerY + float32(math.Sin(midAngle)*float64(innerRadius))

		// Draw tooth as a filled path
		vertices := []ebiten.Vertex{
			{DstX: x1, DstY: y1, SrcX: 0, SrcY: 0, ColorR: float32(iconColor.R) / 255, ColorG: float32(iconColor.G) / 255, ColorB: float32(iconColor.B) / 255, ColorA: float32(iconColor.A) / 255},
			{DstX: x2, DstY: y2, SrcX: 0, SrcY: 0, ColorR: float32(iconColor.R) / 255, ColorG: float32(iconColor.G) / 255, ColorB: float32(iconColor.B) / 255, ColorA: float32(iconColor.A) / 255},
			{DstX: x3, DstY: y3, SrcX: 0, SrcY: 0, ColorR: float32(iconColor.R) / 255, ColorG: float32(iconColor.G) / 255, ColorB: float32(iconColor.B) / 255, ColorA: float32(iconColor.A) / 255},
		}
		indices := []uint16{0, 1, 2}
		screen.DrawTriangles(vertices, indices, ebiten.NewImage(1, 1), nil)
	}

	// Draw center circle
	vector.DrawFilledCircle(screen, centerX, centerY, innerRadius*0.6, iconColor, true)

	// Draw inner hole
	vector.DrawFilledCircle(screen, centerX, centerY, innerRadius*0.3, bgColor, true)
}

// IsPointInCircle checks if a point (px, py) is inside a circle at (cx, cy) with radius r
func IsPointInCircle(px, py, cx, cy, r float32) bool {
	dx := px - cx
	dy := py - cy
	return dx*dx+dy*dy <= r*r
}
