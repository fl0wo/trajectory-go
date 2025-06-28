package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/colors"
	Models "github.com/you/trajectory/space/model"
	"github.com/you/trajectory/space/resources"
	"golang.org/x/image/math/f32"
	"image/color"
	"math"
	"time"
)

// drawPlayer renders the player
func (r *Renderer) drawPlayer(screen *ebiten.Image, model *Models.SpaceGame) {
	camera := model.Camera
	player := model.Player
	playerScreenPos := camera.WorldToScreen(player.Position, constants.ScreenWidth, constants.ScreenHeight)
	playerRadius := camera.RadiusToScreen(player.Radius, constants.ScreenWidth, constants.ScreenHeight)

	// Draw alien trail effect if shader is available
	if r.alienTrailShader != nil {
		r.drawPlayerWithAlienTrail(screen, model)
	}

	// Render player with image if ImagePath is provided, otherwise use green circle
	if player.ImagePath != "" {
		playerImg := resources.LoadImage(player.ImagePath)
		if playerImg != nil {
			// Calculate rotation based on velocity direction
			var rotation float64 = 0
			if player.IsMoving() {
				rotation = math.Atan2(float64(player.Velocity[1]), float64(player.Velocity[0]))
			}
			r.drawPlayerWithImage(screen, playerScreenPos, playerRadius, player.ImagePath, rotation)
		} else {
			// Fallback to green circle if image loading fails
			vector.DrawFilledCircle(screen, playerScreenPos[0], playerScreenPos[1], playerRadius, colors.PlayerBody, true)
		}
	} else {
		// Draw simple green circle when no ImagePath is provided
		vector.DrawFilledCircle(screen, playerScreenPos[0], playerScreenPos[1], playerRadius, colors.PlayerBody, true)
	}
}

// drawPlayerTrail draws the player's movement trail with fading effect
func (r *Renderer) drawPlayerTrail(screen *ebiten.Image, model *Models.SpaceGame) {
	trailPoints := model.Player.GetTrailPoints()
	if len(trailPoints) < 2 {
		return // Need at least 2 points to draw lines
	}

	camera := model.Camera
	now := time.Now()

	// If shadows are enabled and we have the player trail shader, apply light effects
	if model.ShadowsEnabled && r.playerTrailShader != nil {
		// Get light information (same as shadow system)
		lightPos := camera.WorldToScreen(model.Player.Position, constants.ScreenWidth, constants.ScreenHeight)
		lightDirection := camera.WorldToScreen(camera.Position, constants.ScreenWidth, constants.ScreenHeight)

		// Calculate light direction vector (from light pos to camera/target)
		lightDirVec := f32.Vec2{
			lightDirection[0] - lightPos[0],
			lightDirection[1] - lightPos[1],
		}

		// Calculate max distance for the light cone (same as shadow system)
		maxDistance := math.Hypot(float64(constants.ScreenWidth), float64(constants.ScreenHeight))
		fov := r.getAdaptiveFov(lightDirection, lightPos)

		// Draw lines between consecutive trail points with shader
		for i := 1; i < len(trailPoints); i++ {
			prevPoint := trailPoints[i-1]
			currPoint := trailPoints[i]

			// Calculate age of the current point
			age := now.Sub(currPoint.Timestamp)
			trailDuration := 3.0 * time.Second // Should match the constant in Player.go

			// Calculate alpha based on age (newer = more opaque)
			ageRatio := float64(age) / float64(trailDuration)
			alpha := uint8(255 * (1.0 - ageRatio))
			if ageRatio >= 1.0 {
				continue // Skip expired points
			}

			// Convert world positions to screen coordinates
			prevScreen := camera.WorldToScreen(prevPoint.Position, constants.ScreenWidth, constants.ScreenHeight)
			currScreen := camera.WorldToScreen(currPoint.Position, constants.ScreenWidth, constants.ScreenHeight)

			// Trail color with fading alpha
			trailColor := color.RGBA{R: colors.PlayerTrail.R, G: colors.PlayerTrail.G, B: colors.PlayerTrail.B, A: alpha}

			// Calculate line width based on age (newer = thicker)
			width := float32(1 + (1.0-ageRatio)*2) // Width from 1 to 3

			// Prepare shader uniforms
			uniforms := map[string]any{
				"LightPos":       []float32{lightPos[0], lightPos[1]},
				"LightDirection": []float32{lightDirVec[0], lightDirVec[1]},
				"FOVAngle":       float32(fov * math.Pi / 180.0), // Convert to radians
				"MaxDistance":    float32(maxDistance),
				"Zoom":           camera.GetTotalZoom(),
				"OriginalColor": []float32{
					float32(trailColor.R) / 255.0,
					float32(trailColor.G) / 255.0,
					float32(trailColor.B) / 255.0,
					float32(trailColor.A) / 255.0,
				},
			}

			// Draw the trail segment with shader effect
			r.drawLineWithShader(screen, prevScreen, currScreen, width, trailColor, r.playerTrailShader, uniforms)
		}
	}
}

// drawPlayerWithImage renders the player using an image texture with rotation based on movement
func (r *Renderer) drawPlayerWithImage(screen *ebiten.Image, screenPos f32.Vec2, radius float32, imagePath string, rotation float64) {
	// Load the image
	img := resources.LoadImage(imagePath)
	if img == nil {
		// Fallback to circle if image loading fails
		vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], radius, colors.PlayerBody, true)
		return
	}

	// Calculate scaling to fit the desired radius
	imgSize := img.Bounds().Size()
	imgRadius := float32(imgSize.X) / 2.0 // Assume square images
	if imgSize.Y > imgSize.X {
		imgRadius = float32(imgSize.Y) / 2.0
	}

	scale := (radius * 2.0) / (imgRadius * 2.0) // Scale to fit diameter

	// Create draw options
	op := &ebiten.DrawImageOptions{}

	// Move image center to origin for rotation/scaling
	op.GeoM.Translate(-float64(imgSize.X)/2, -float64(imgSize.Y)/2)

	// Apply rotation (astronaut points in direction of movement)
	op.GeoM.Rotate(rotation)

	// Scale the image to the desired size
	op.GeoM.Scale(float64(scale), float64(scale))

	// Move to final screen position
	op.GeoM.Translate(float64(screenPos[0]), float64(screenPos[1]))

	// Draw the image
	screen.DrawImage(img, op)
}

// DrawTrajectoryArrow draws trajectory arrow (called from main game loop with drag info)
func (r *Renderer) DrawTrajectoryArrow(screen *ebiten.Image, model *Models.SpaceGame, startPos, currentPos f32.Vec2, isDragging bool) {
	if !isDragging || model.Player.IsMoving() {
		return
	}

	camera := model.Camera

	// Convert screen drag to world space for trajectory calculation
	startWorld := camera.ScreenToWorld(startPos, constants.ScreenWidth, constants.ScreenHeight)
	currentWorld := camera.ScreenToWorld(currentPos, constants.ScreenWidth, constants.ScreenHeight)

	// Calculate trajectory vector (opposite of drag direction)
	trajectoryVector := f32.Vec2{
		startWorld[0] - currentWorld[0],
		startWorld[1] - currentWorld[1],
	}

	// Calculate drag distance and clamp it to maximum allowed (same as throw logic)
	const maxDragDistance = float32(0.25)
	const maxVelocity = float32(0.5)

	trajectoryDistance := float32(math.Sqrt(float64(trajectoryVector[0]*trajectoryVector[0] + trajectoryVector[1]*trajectoryVector[1])))
	if trajectoryDistance > maxDragDistance {
		// Normalize and scale to max distance
		trajectoryVector[0] = (trajectoryVector[0] / trajectoryDistance) * maxDragDistance
		trajectoryVector[1] = (trajectoryVector[1] / trajectoryDistance) * maxDragDistance
	}

	// Scale trajectory for visualization (same as throw logic)
	velocityMultiplier := float32(2.0)
	trajectoryVector[0] *= velocityMultiplier
	trajectoryVector[1] *= velocityMultiplier

	// Clamp final visualization vector to match max velocity
	visualMagnitude := float32(math.Sqrt(float64(trajectoryVector[0]*trajectoryVector[0] + trajectoryVector[1]*trajectoryVector[1])))
	if visualMagnitude > maxVelocity {
		trajectoryVector[0] = (trajectoryVector[0] / visualMagnitude) * maxVelocity
		trajectoryVector[1] = (trajectoryVector[1] / visualMagnitude) * maxVelocity
	}

	// Start arrow from player center
	playerWorldPos := model.Player.Position
	endWorld := f32.Vec2{
		playerWorldPos[0] + trajectoryVector[0],
		playerWorldPos[1] + trajectoryVector[1],
	}

	// Convert to screen coordinates
	playerScreen := camera.WorldToScreen(playerWorldPos, constants.ScreenWidth, constants.ScreenHeight)
	endScreen := camera.WorldToScreen(endWorld, constants.ScreenWidth, constants.ScreenHeight)

	// Only draw if trajectory has significant length
	dragDistance := float32(math.Sqrt(float64(trajectoryVector[0]*trajectoryVector[0] + trajectoryVector[1]*trajectoryVector[1])))
	if dragDistance > 0.01 {
		// If shadows are enabled and we have the trajectory arrow shader, apply light effects
		if model.ShadowsEnabled && r.trajectoryArrowShader != nil {
			// Get light information (same as shadow system)
			lightPos := camera.WorldToScreen(model.Player.Position, constants.ScreenWidth, constants.ScreenHeight)
			lightDirection := camera.WorldToScreen(camera.Position, constants.ScreenWidth, constants.ScreenHeight)

			// Calculate light direction vector (from light pos to camera/target)
			lightDirVec := f32.Vec2{
				lightDirection[0] - lightPos[0],
				lightDirection[1] - lightPos[1],
			}

			// Calculate max distance for the light cone (same as shadow system)
			maxDistance := math.Hypot(float64(constants.ScreenWidth), float64(constants.ScreenHeight))

			fov := r.getAdaptiveFov(lightDirection, lightPos)

			// Prepare shader uniforms
			uniforms := map[string]any{
				"LightPos":       []float32{lightPos[0], lightPos[1]},
				"LightDirection": []float32{lightDirVec[0], lightDirVec[1]},
				"FOVAngle":       float32(fov * math.Pi / 180.0), // Convert to radians
				"MaxDistance":    float32(maxDistance),
				"Zoom":           camera.GetTotalZoom(),
				"OriginalColor": []float32{
					float32(colors.TrajectoryArrow.R) / 255.0,
					float32(colors.TrajectoryArrow.G) / 255.0,
					float32(colors.TrajectoryArrow.B) / 255.0,
					float32(colors.TrajectoryArrow.A) / 255.0,
				},
			}

			// Draw main arrow line with shader effect
			r.drawLineWithShader(screen, playerScreen, endScreen, 4, colors.TrajectoryArrow, r.trajectoryArrowShader, uniforms)

			// Draw arrowhead with shader effect
			r.drawArrowHeadWithShader(screen, playerScreen, endScreen, colors.TrajectoryArrow, r.trajectoryArrowShader, uniforms)
		} else {
			// Fallback to regular trajectory arrow rendering
			// Draw main arrow line (white, thick)
			vector.StrokeLine(screen, playerScreen[0], playerScreen[1], endScreen[0], endScreen[1], 4, colors.TrajectoryArrow, true)

			// Draw arrowhead
			r.drawArrowHead(screen, playerScreen, endScreen, colors.TrajectoryArrow)
		}
	}
}
func (r *Renderer) drawPlayerWithAlienTrail(screen *ebiten.Image, model *Models.SpaceGame) {
	player := model.Player
	cam := model.Camera

	// world‐space in [0..1]
	playerWorldPos := player.Position
	playerWorldRad := player.Radius

	// seconds since start
	currentTime := float32(time.Since(r.startTime).Seconds())

	print("colors.PlayerBody: ", colors.PlayerBody.R, "\n")

	// build the uniform map
	uniforms := map[string]interface{}{
		// core camera/player transforms
		"PlayerPos":   []float32{playerWorldPos[0], playerWorldPos[1]},
		"PlayerColor": []float32{float32(colors.PlayerBody.R), float32(colors.PlayerBody.G), float32(colors.PlayerBody.B), float32(colors.PlayerBody.A)},
		"CameraPos":   []float32{cam.Position[0], cam.Position[1]},
		"Zoom":        cam.GetTotalZoom(),
		"Radius":      playerWorldRad,

		// velocity & time
		"Velocity": []float32{player.Velocity[0], player.Velocity[1]},
		"Time":     currentTime,

		// screen geometry
		"ScreenSize": []float32{float32(constants.ScreenWidth), float32(constants.ScreenHeight)},

		// capsule-trail parameters
		"DropCount":   8,            // how many capsules at once
		"Lifetime":    float32(2),   // each lives 1.2s
		"TrailLength": float32(2),   // 2.5× the player radius
		"DropSizeMin": float32(0.4), // thickness = 10% of player radius
		"DropSizeMax": float32(0.6), // length = 30% of player radius
		"JitterAmt":   float32(0.8), // up to 20% sideways
		"SpawnRate":   float32(3.0), // 2 capsules per second
	}

	// hook up your Ebiten shader
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = r.whiteTexture
	screen.DrawRectShader(
		constants.ScreenWidth,
		constants.ScreenHeight,
		r.alienTrailShader,
		op,
	)
}

// drawArrowHead draws an arrowhead at the end of a line
func (r *Renderer) drawArrowHead(screen *ebiten.Image, start, end f32.Vec2, color color.RGBA) {
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

	// Draw arrow head lines
	vector.StrokeLine(screen, end[0], end[1], leftX, leftY, 3, color, true)
	vector.StrokeLine(screen, end[0], end[1], rightX, rightY, 3, color, true)
}
