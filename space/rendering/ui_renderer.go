package rendering

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/colors"
	Models "github.com/you/trajectory/space/model"
	"golang.org/x/image/math/f32"
	"image/color"
	"math"
	"time"
)

// drawNebulaBackground renders the animated nebula background
func (r *Renderer) drawNebulaBackground(screen *ebiten.Image, model *Models.SpaceGame) {
	if r.nebulaShader == nil {
		// Fallback to solid background
		screen.Fill(colors.NebulaBackground)
		return
	}

	camera := model.Camera

	// Calculate elapsed time since game start for animations
	currentTime := float32(time.Since(r.startTime).Seconds())

	// Get player and camera positions for parallax
	playerPos := model.Player.Position
	cameraPos := camera.Position

	// Prepare shader uniforms
	uniforms := map[string]any{
		"Time":       currentTime,
		"PlayerPos":  []float32{playerPos[0], playerPos[1]}, // World coordinates
		"CameraPos":  []float32{cameraPos[0], cameraPos[1]}, // World coordinates
		"ScreenSize": []float32{float32(constants.ScreenWidth), float32(constants.ScreenHeight)},
		"Zoom":       camera.Zoom,
	}

	// Draw the nebula background shader
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = r.whiteTexture // Use white texture as dummy source

	screen.DrawRectShader(constants.ScreenWidth, constants.ScreenHeight, r.nebulaShader, op)
}

// drawFPS draws the current FPS in the top right corner
func (r *Renderer) drawFPS(screen *ebiten.Image) {
	const fontSize = 16

	// Calculate FPS
	fps := ebiten.ActualTPS()
	fpsText := fmt.Sprintf("FPS: %.0f", fps)

	// Create text face
	face := &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   fontSize,
	}

	// Measure text width to position it correctly in top right
	textWidth, _ := text.Measure(fpsText, face, 0)

	// Position in top right corner with some padding
	const padding = 10
	x := float64(constants.ScreenWidth) - textWidth - padding
	y := float64(padding + fontSize)

	// Draw the FPS text
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, fpsText, face, op)
}

// drawBorderIndicators draws "+" markers at each corner of the game border,
// stretching only the two interior-facing arms of each cross.
func (r *Renderer) drawBorderIndicators(screen *ebiten.Image, model *Models.SpaceGame) {
	// 1) Get border corners in screen space
	border := model.CalculateBorders()
	cam := model.Camera
	bl := cam.WorldToScreen(border.BottomLeft, constants.ScreenWidth, constants.ScreenHeight)
	br := cam.WorldToScreen(border.BottomRight, constants.ScreenWidth, constants.ScreenHeight)
	tl := cam.WorldToScreen(border.TopLeft, constants.ScreenWidth, constants.ScreenHeight)
	tr := cam.WorldToScreen(border.TopRight, constants.ScreenWidth, constants.ScreenHeight)

	// 2) Precompute stretched length
	stretchArm := baseArm * stretchFactor
	col := colors.BorderIndicator

	// 3) Check if we should apply light inversion effect
	if model.ShadowsEnabled && r.invertOnLightShader != nil {
		// Get light information for shader
		lightPos := cam.WorldToScreen(model.Player.Position, constants.ScreenWidth, constants.ScreenHeight)
		lightDirection := cam.WorldToScreen(cam.Position, constants.ScreenWidth, constants.ScreenHeight)

		// Calculate light direction vector
		lightDirVec := f32.Vec2{
			lightDirection[0] - lightPos[0],
			lightDirection[1] - lightPos[1],
		}

		maxDistance := math.Hypot(float64(constants.ScreenWidth), float64(constants.ScreenHeight))
		fov := r.getAdaptiveFov(lightDirection, lightPos)

		// Prepare uniforms for the invert shader
		uniforms := r.prepareInvertOnLightUniforms(
			lightPos, lightDirVec,
			float32(fov*math.Pi/180.0), // Convert to radians
			float32(maxDistance),
			cam.Zoom,
			col,
		)

		// Draw each corner with shader effect
		r.drawCornerPlusWithShader(screen, bl, baseArm, stretchArm, true, true, col, uniforms)
		r.drawCornerPlusWithShader(screen, br, baseArm, stretchArm, false, true, col, uniforms)
		r.drawCornerPlusWithShader(screen, tl, baseArm, stretchArm, true, false, col, uniforms)
		r.drawCornerPlusWithShader(screen, tr, baseArm, stretchArm, false, false, col, uniforms)
	} else {
		// Fallback to regular drawing without shader
		r.drawCornerPlus(screen, bl, baseArm, stretchArm, true, true, col)
		r.drawCornerPlus(screen, br, baseArm, stretchArm, false, true, col)
		r.drawCornerPlus(screen, tl, baseArm, stretchArm, true, false, col)
		r.drawCornerPlus(screen, tr, baseArm, stretchArm, false, false, col)
	}
}

// drawCornerPlus draws a "+" at center, with its two interior-facing arms
// stretched. If stretchRight is true, the right arm is long; otherwise left
// arm is long. If stretchUp is true, the up arm is long; otherwise down arm
// is long.
func (r *Renderer) drawCornerPlus(
	screen *ebiten.Image,
	center f32.Vec2,
	baseArm, stretchArm float32,
	stretchRight, stretchUp bool,
	clr color.RGBA,
) {
	// decide horizontal lengths
	var leftLen, rightLen = baseArm, baseArm
	if stretchRight {
		rightLen = stretchArm
	} else {
		leftLen = stretchArm
	}

	// decide vertical lengths
	var downLen, upLen = baseArm, baseArm
	if stretchUp {
		upLen = stretchArm
	} else {
		downLen = stretchArm
	}

	// draw horizontal stroke
	r.DrawLine(screen,
		center[0]-leftLen, center[1],
		center[0]+rightLen, center[1],
		lineWidth, clr,
	)
	// draw vertical stroke
	r.DrawLine(screen,
		center[0], center[1]-downLen,
		center[0], center[1]+upLen,
		lineWidth, clr,
	)
}

// DrawPlusMarker now takes two arm lengths (half-width & half-height)
func (r *Renderer) DrawPlusMarker(
	screen *ebiten.Image,
	center f32.Vec2,
	armX, armY,
	offX, offY float32,
	lineWidth float32,
	clr color.RGBA,
) {
	// draw a horizontal line of length 2*armX centered at `center`
	r.DrawLine(screen,
		center[0]-offX, center[1],
		center[0]+offX, center[1],
		lineWidth, clr,
	)
	// draw a vertical   line of length 2*armY centered at `center`
	r.DrawLine(screen,
		center[0], center[1]-armY,
		center[0], center[1]+armY,
		lineWidth, clr,
	)
}

// drawCornerPlusWithShader draws a "+" at center using the invert shader, with its two interior-facing arms stretched
func (r *Renderer) drawCornerPlusWithShader(
	screen *ebiten.Image,
	center f32.Vec2,
	baseArm, stretchArm float32,
	stretchRight, stretchUp bool,
	clr color.RGBA,
	uniforms map[string]any,
) {
	// decide horizontal lengths
	var leftLen, rightLen = baseArm, baseArm
	if stretchRight {
		rightLen = stretchArm
	} else {
		leftLen = stretchArm
	}

	// decide vertical lengths
	var downLen, upLen = baseArm, baseArm
	if stretchUp {
		upLen = stretchArm
	} else {
		downLen = stretchArm
	}

	// draw horizontal stroke with shader
	r.drawLineWithShader(screen,
		f32.Vec2{center[0] - leftLen, center[1]},
		f32.Vec2{center[0] + rightLen, center[1]},
		lineWidth, clr, r.invertOnLightShader, uniforms,
	)
	// draw vertical stroke with shader
	r.drawLineWithShader(screen,
		f32.Vec2{center[0], center[1] - downLen},
		f32.Vec2{center[0], center[1] + upLen},
		lineWidth, clr, r.invertOnLightShader, uniforms,
	)
}

// DrawLine is a helper function for drawing lines with vector graphics
func (r *Renderer) DrawLine(screen *ebiten.Image, x1, y1, x2, y2, width float32, clr color.RGBA) {
	vector.StrokeLine(screen, x1, y1, x2, y2, width, clr, true)
}
