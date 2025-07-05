package gui

import (
	"fmt"
	"github.com/you/trajectory/space/colors"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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

// IsPointInRect checks if a point (px, py) is inside a rectangle at (x, y) with width and height
func IsPointInRect(px, py, x, y, width, height float32) bool {
	return px >= x && px <= x+width && py >= y && py <= y+height
}

// DrawLevelCircle draws a level selection circle with number inside (Super Mario Galaxy style)
func DrawLevelCircle(screen *ebiten.Image, centerX, centerY, radius float32, levelNum int, textFace *text.GoTextFace, isLocked bool) {
	// Choose colors based on lock state
	var bgColor, borderColor, textColor color.RGBA
	if isLocked {
		bgColor = color.RGBA{60, 60, 60, 255}        // Dark gray
		borderColor = color.RGBA{100, 100, 100, 255} // Lighter gray
		textColor = color.RGBA{150, 150, 150, 255}   // Gray text
	} else {
		bgColor = color.RGBA{255, 215, 0, 255}       // Gold background
		borderColor = color.RGBA{255, 255, 255, 255} // White border
		textColor = color.RGBA{0, 0, 0, 255}         // Black text
	}

	// Draw background circle
	vector.DrawFilledCircle(screen, centerX, centerY, radius, bgColor, true)

	// Draw border circle
	borderWidth := float32(3)
	vector.StrokeCircle(screen, centerX, centerY, radius, borderWidth, borderColor, true)

	// Draw level number in center
	levelText := fmt.Sprintf("%d", levelNum)
	textWidth, textHeight := text.Measure(levelText, textFace, 0)

	textX := centerX - float32(textWidth)/2
	textY := centerY - float32(textHeight)/2

	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(textX), float64(textY))
	op.ColorScale.ScaleWithColor(textColor)
	text.Draw(screen, levelText, textFace, op)
}

// DrawLevelPreviewRect draws a rectangular level preview with border and level number
func DrawLevelPreviewRect(screen *ebiten.Image, centerX, centerY, width, height float32, levelNum int, textFace *text.GoTextFace, isLocked bool, previewImage *ebiten.Image) {
	// Calculate rectangle coordinates (center-based to corner-based)
	x := centerX - width/2
	y := centerY - height/2

	// Choose colors based on lock state
	var borderColor, textColor, bgColor color.RGBA
	if isLocked {
		borderColor = color.RGBA{100, 100, 100, 255} // Gray border
		textColor = color.RGBA{150, 150, 150, 255}   // Gray text
		bgColor = color.RGBA{30, 30, 30, 100}        // Dark semi-transparent background
	} else {
		borderColor = colors.BorderPreviewLevel // Black border
		textColor = colors.BorderPreviewLevel   // White text
		bgColor = colors.BorderPreviewLevel     // Dark blue semi-transparent background
	}

	// Draw preview image if available, otherwise draw background rectangle
	if previewImage != nil && !isLocked {
		// Draw the level preview image, scaled to fit the rectangle
		op := &ebiten.DrawImageOptions{}

		// Calculate scaling to fit the preview in the rectangle
		previewWidth := float32(previewImage.Bounds().Dx())
		previewHeight := float32(previewImage.Bounds().Dy())
		scaleX := width / previewWidth
		scaleY := height / previewHeight

		// Use the smaller scale to maintain aspect ratio
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}

		// Center the scaled image
		scaledWidth := previewWidth * scale
		scaledHeight := previewHeight * scale
		offsetX := (width - scaledWidth) / 2
		offsetY := (height - scaledHeight) / 2

		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(x+offsetX), float64(y+offsetY))

		screen.DrawImage(previewImage, op)
	} else {
		// Draw background rectangle (filled) for locked levels or when no preview
		vertices := []ebiten.Vertex{
			{DstX: x, DstY: y, SrcX: 0, SrcY: 0, ColorR: float32(bgColor.R) / 255, ColorG: float32(bgColor.G) / 255, ColorB: float32(bgColor.B) / 255, ColorA: float32(bgColor.A) / 255},
			{DstX: x + width, DstY: y, SrcX: 0, SrcY: 0, ColorR: float32(bgColor.R) / 255, ColorG: float32(bgColor.G) / 255, ColorB: float32(bgColor.B) / 255, ColorA: float32(bgColor.A) / 255},
			{DstX: x, DstY: y + height, SrcX: 0, SrcY: 0, ColorR: float32(bgColor.R) / 255, ColorG: float32(bgColor.G) / 255, ColorB: float32(bgColor.B) / 255, ColorA: float32(bgColor.A) / 255},
			{DstX: x + width, DstY: y + height, SrcX: 0, SrcY: 0, ColorR: float32(bgColor.R) / 255, ColorG: float32(bgColor.G) / 255, ColorB: float32(bgColor.B) / 255, ColorA: float32(bgColor.A) / 255},
		}
		indices := []uint16{0, 1, 2, 1, 2, 3}
		screen.DrawTriangles(vertices, indices, ebiten.NewImage(1, 1), nil)
	}

	// Draw border (using vector lines)
	borderWidth := float32(2)
	// Top line
	vector.StrokeLine(screen, x, y, x+width, y, borderWidth, borderColor, true)
	// Bottom line
	vector.StrokeLine(screen, x, y+height, x+width, y+height, borderWidth, borderColor, true)
	// Left line
	vector.StrokeLine(screen, x, y, x, y+height, borderWidth, borderColor, true)
	// Right line
	vector.StrokeLine(screen, x+width, y, x+width, y+height, borderWidth, borderColor, true)

	// Draw level number overlay (only if no preview or locked)
	if previewImage == nil || isLocked {
		levelText := fmt.Sprintf("Level %d", levelNum)
		textWidth, textHeight := text.Measure(levelText, textFace, 0)

		textX := centerX - float32(textWidth)/2
		textY := centerY - float32(textHeight)/2

		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(textX), float64(textY))
		op.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, levelText, textFace, op)
	} else {
		// Draw small level number in corner for previews
		levelText := fmt.Sprintf("%d", levelNum)
		textWidth, textHeight := text.Measure(levelText, textFace, 0)

		// Position in top-left corner with small padding
		padding := float32(4)
		textX := x + padding
		textY := y + padding + float32(textHeight)

		// Draw semi-transparent background for the number
		bgVertices := []ebiten.Vertex{
			{DstX: textX - padding, DstY: textY - float32(textHeight) - padding, SrcX: 0, SrcY: 0, ColorR: 0, ColorG: 0, ColorB: 0, ColorA: 0.7},
			{DstX: textX + float32(textWidth) + padding, DstY: textY - float32(textHeight) - padding, SrcX: 0, SrcY: 0, ColorR: 0, ColorG: 0, ColorB: 0, ColorA: 0.7},
			{DstX: textX - padding, DstY: textY + padding, SrcX: 0, SrcY: 0, ColorR: 0, ColorG: 0, ColorB: 0, ColorA: 0.7},
			{DstX: textX + float32(textWidth) + padding, DstY: textY + padding, SrcX: 0, SrcY: 0, ColorR: 0, ColorG: 0, ColorB: 0, ColorA: 0.7},
		}
		bgIndices := []uint16{0, 1, 2, 1, 2, 3}
		screen.DrawTriangles(bgVertices, bgIndices, ebiten.NewImage(1, 1), nil)

		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(textX), float64(textY))
		op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255}) // White text
		text.Draw(screen, levelText, textFace, op)
	}
}

// DrawConnectionLine draws a line connecting two level circles
func DrawConnectionLine(screen *ebiten.Image, x1, y1, x2, y2 float32, lineColor color.RGBA) {
	// Calculate line direction and draw multiple small segments for a thicker line
	dx := x2 - x1
	dy := y2 - y1
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if length == 0 {
		return
	}

	// Normalize direction
	dx /= length
	dy /= length

	// Line thickness
	thickness := float32(4)

	// Draw thick line using multiple parallel thin lines
	for i := float32(-thickness / 2); i <= thickness/2; i += 0.5 {
		// Perpendicular offset
		perpX := -dy * i
		perpY := dx * i

		// Draw line with offset
		vector.StrokeLine(screen, x1+perpX, y1+perpY, x2+perpX, y2+perpY, 1, lineColor, true)
	}
}

// DrawBackButton draws a back arrow icon at the specified position
func DrawBackButton(screen *ebiten.Image, x, y, size float32, iconColor color.RGBA) {
	centerX := x + size/2
	centerY := y + size/2

	// Draw circular background
	bgColor := color.RGBA{40, 40, 40, 200} // Semi-transparent dark background
	vector.DrawFilledCircle(screen, centerX, centerY, size/2, bgColor, true)

	// Draw back arrow shape
	arrowSize := size * 0.3
	arrowThickness := size * 0.06

	// Arrow points left, so we draw from right to left
	// Arrow shaft (horizontal line)
	shaftY := centerY
	shaftStartX := centerX + arrowSize/2
	shaftEndX := centerX - arrowSize/2
	vector.StrokeLine(screen, shaftStartX, shaftY, shaftEndX, shaftY, arrowThickness, iconColor, true)

	// Arrow head (two lines forming a V pointing left)
	headSize := arrowSize * 0.4
	headStartX := shaftEndX
	headStartY := shaftY

	// Upper arrow head line
	headEndX1 := headStartX + headSize*0.7
	headEndY1 := headStartY - headSize*0.7
	vector.StrokeLine(screen, headStartX, headStartY, headEndX1, headEndY1, arrowThickness, iconColor, true)

	// Lower arrow head line
	headEndX2 := headStartX + headSize*0.7
	headEndY2 := headStartY + headSize*0.7
	vector.StrokeLine(screen, headStartX, headStartY, headEndX2, headEndY2, arrowThickness, iconColor, true)
}
