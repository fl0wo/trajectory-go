package space

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/input"
	Models "github.com/you/trajectory/space/model"
	"github.com/you/trajectory/space/resources"
	"github.com/you/trajectory/space/shadows"
	"golang.org/x/image/math/f32"
	"image/color"
	"math"
	"time"
)

var (
	// #2F3262
	backgroundColor = color.RGBA{
		R: 6, G: 2, B: 25, A: 255,
	}
)

const (
	// Maximum drag distance in world units to limit throw power
	maxDragDistance = float32(0.25)
	// Maximum velocity magnitude to prevent excessive speeds
	maxVelocity = float32(0.5)
)

type Game struct {
	input        *input.Input
	model        *Models.SpaceGame
	lastTime     int64 // For delta time calculation
	shadowSystem *shadows.ShadowSystem
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return constants.ScreenWidth, constants.ScreenHeight
}

// Draw draws the current game to the given screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)

	camera := g.model.Camera

	// Render shadows if enabled
	if g.model.ShadowsEnabled && g.shadowSystem != nil {
		// Collect celestial bodies in screen coordinates
		var celestialPositions []f32.Vec2
		var celestialRadii []float32

		for _, body := range g.model.CelestialBodies {
			// if player x is < than celestial body x, then ignore
			if g.model.Player.IsBefore(body) || body.GetType() == Models.CelestialBodyTypeWhiteHole {
				screenPos := camera.WorldToScreen(body.GetPosition(), constants.ScreenWidth, constants.ScreenHeight)
				screenRadius := camera.RadiusToScreen(body.GetRadius(), constants.ScreenWidth, constants.ScreenHeight)
				celestialPositions = append(celestialPositions, screenPos)
				celestialRadii = append(celestialRadii, screenRadius)
			}
		}

		// Collect asteroids in screen coordinates
		var asteroidPositions []f32.Vec2
		var asteroidRadii []float32

		for _, asteroid := range g.model.RingAsteroids {

			if g.model.Player.IsBefore(asteroid) {
				screenPos := camera.WorldToScreen(asteroid.GetPosition(), constants.ScreenWidth, constants.ScreenHeight)
				screenRadius := camera.RadiusToScreen(asteroid.GetRadius(), constants.ScreenWidth, constants.ScreenHeight)
				asteroidPositions = append(asteroidPositions, screenPos)
				asteroidRadii = append(asteroidRadii, screenRadius)
			}
		}

		// Use player position as light source in screen coordinates
		lightPos := camera.WorldToScreen(g.model.Player.Position, constants.ScreenWidth, constants.ScreenHeight)

		// Render shadows using the new system
		g.shadowSystem.RenderShadows(screen, lightPos, celestialPositions, celestialRadii, asteroidPositions, asteroidRadii, false)
	}

	// Draw celestial bodies with camera transform
	for _, body := range g.model.CelestialBodies {
		bodyPos := body.GetPosition()
		screenPos := camera.WorldToScreen(bodyPos, constants.ScreenWidth, constants.ScreenHeight)
		radius := camera.RadiusToScreen(body.GetRadius(), constants.ScreenWidth, constants.ScreenHeight)

		// Choose colors based on celestial body type
		var bodyColor color.RGBA
		var orbitColor color.RGBA

		switch body.GetType() {
		case Models.CelestialBodyTypePlanet:
			bodyColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White for planets
			orbitColor = color.RGBA{R: 255, G: 255, B: 0, A: 128}  // Yellow orbit for planets
		case Models.CelestialBodyTypeBlackHole:
			bodyColor = color.RGBA{R: 128, G: 128, B: 128, A: 255} // Gray for blackholes
			orbitColor = color.RGBA{R: 128, G: 0, B: 128, A: 128}  // Purple orbit for blackholes
		case Models.CelestialBodyTypeWhiteHole:
			bodyColor = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White for white holes (victory)
			orbitColor = color.RGBA{R: 0, G: 255, B: 0, A: 128}    // Green orbit for white holes
		case Models.CelestialBodyTypeAsteroid:
			bodyColor = color.RGBA{R: 139, G: 69, B: 19, A: 255} // Brown for asteroids
			orbitColor = color.RGBA{R: 0, G: 0, B: 0, A: 0}      // No orbit for asteroids
		}

		// Only draw if on screen (simple culling)
		//if screenPos[0] > -radius && screenPos[0] < ScreenWidth+radius && screenPos[1] > -radius && screenPos[1] < ScreenHeight+radius {

		// Check if this celestial body has an image
		var imagePath string
		switch body.GetType() {
		case Models.CelestialBodyTypePlanet:
			if planet, ok := body.(*Models.Planet); ok && planet.ImagePath != "" {
				imagePath = planet.ImagePath
			}
		case Models.CelestialBodyTypeBlackHole:
			if blackHole, ok := body.(*Models.BlackHole); ok && blackHole.ImagePath != "" {
				imagePath = blackHole.ImagePath
			}
		case Models.CelestialBodyTypeWhiteHole:
			if whiteHole, ok := body.(*Models.WhiteHole); ok && whiteHole.ImagePath != "" {
				imagePath = whiteHole.ImagePath
			}
		}

		if imagePath != "" {
			// Render with image
			g.drawCelestialBodyWithImage(screen, screenPos, radius, imagePath)
		} else {
			// Fallback to circle rendering
			vector.DrawFilledCircle(
				screen,
				screenPos[0],
				screenPos[1],
				radius,
				bodyColor, true,
			)
		}
		// Draw celestial body's orbit radius as a circle
		orbitRadius := camera.RadiusToScreen(body.GetOrbitRadius(), constants.ScreenWidth, constants.ScreenHeight)
		if orbitRadius > 0 {
			vector.StrokeCircle(
				screen,
				screenPos[0],
				screenPos[1],
				orbitRadius,
				2,          // Line width
				orbitColor, // Color based on body type
				true,       // Closed
			)
		}
	}

	// Draw asteroids
	for _, asteroid := range g.model.RingAsteroids {
		asteroidPos := asteroid.GetPosition()
		screenPos := camera.WorldToScreen(asteroidPos, constants.ScreenWidth, constants.ScreenHeight)
		radius := camera.RadiusToScreen(asteroid.GetRadius(), constants.ScreenWidth, constants.ScreenHeight)

		if asteroid.ImagePath != "" {
			// Render with image
			g.drawCelestialBodyWithImage(screen, screenPos, radius, asteroid.ImagePath)
		} else {
			// Fallback to circle rendering
			asteroidColor := color.RGBA{R: 139, G: 69, B: 19, A: 255}
			vector.DrawFilledCircle(
				screen,
				screenPos[0],
				screenPos[1],
				radius,
				asteroidColor, true,
			)
		}
	}

	// Draw player trail
	g.drawPlayerTrail(screen)

	// Draw the player using astronaut image
	player := g.model.Player
	playerScreenPos := camera.WorldToScreen(player.Position, constants.ScreenWidth, constants.ScreenHeight)
	playerRadius := camera.RadiusToScreen(player.Radius, constants.ScreenWidth, constants.ScreenHeight)

	// Render player with image if ImagePath is provided, otherwise use green circle
	if player.ImagePath != "" {
		playerImg := resources.LoadImage(player.ImagePath)
		if playerImg != nil {
			// Calculate rotation based on velocity direction
			var rotation float64 = 0
			if player.IsMoving() {
				rotation = math.Atan2(float64(player.Velocity[1]), float64(player.Velocity[0]))
			}
			g.drawPlayerWithImage(screen, playerScreenPos, playerRadius, player.ImagePath, rotation)
		} else {
			// Fallback to green circle if image loading fails
			vector.DrawFilledCircle(screen,
				playerScreenPos[0],
				playerScreenPos[1],
				playerRadius,
				color.RGBA{G: 255, A: 255}, true, // Green circle
			)
		}
	} else {
		// Draw simple green circle when no ImagePath is provided
		vector.DrawFilledCircle(screen,
			playerScreenPos[0],
			playerScreenPos[1],
			playerRadius,
			color.RGBA{G: 255, A: 255}, true, // Green circle
		)
	}

	// Draw trajectory arrow if dragging
	dragInfo := g.input.GetDragInfo()
	if dragInfo.IsDragging && !g.model.Player.IsMoving() {
		// Convert screen drag to world space for trajectory calculation
		startWorld := camera.ScreenToWorld(dragInfo.StartPos, constants.ScreenWidth, constants.ScreenHeight)
		currentWorld := camera.ScreenToWorld(dragInfo.CurrentPos, constants.ScreenWidth, constants.ScreenHeight)

		// Calculate trajectory vector (opposite of drag direction)
		trajectoryVector := f32.Vec2{
			startWorld[0] - currentWorld[0],
			startWorld[1] - currentWorld[1],
		}

		// Calculate drag distance and clamp it to maximum allowed (same as throw logic)
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
		playerWorldPos := g.model.Player.Position
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
			// Draw main arrow line (white, thick)
			vector.StrokeLine(screen,
				playerScreen[0], playerScreen[1],
				endScreen[0], endScreen[1],
				4, color.RGBA{R: 255, G: 255, B: 255, A: 255}, true,
			)

			// Draw arrowhead
			g.drawArrowHead(screen, playerScreen, endScreen, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
}

// NewGame generates a new Game object.
func NewGame() (*Game, error) {
	g := &Game{
		input: input.NewInput(),
	}
	var err error
	g.model, err = Models.NewSpaceGame()
	if err != nil {
		return nil, err
	}

	// Initialize shadow system
	g.shadowSystem = shadows.NewShadowSystem(constants.ScreenWidth, constants.ScreenHeight)

	return g, nil
}

// Update updates the current game state.
func (g *Game) Update() error {
	// Calculate delta time
	baseDeltaTime := float32(1.0 / 60.0) // Assume 60 FPS for simplicity

	g.input.Update()

	// Calculate time dilation based on player proximity to celestial bodies
	// Use base delta time for the interpolation calculation
	g.model.UpdateTimeDilation(baseDeltaTime)
	timeScale := g.model.TimeScale

	// Calculate proximity zoom based on player proximity to celestial bodies
	g.model.UpdateProximityZoom(baseDeltaTime)

	// Apply proximity zoom to camera
	g.model.Camera.SetProximityZoom(g.model.ProximityZoom)

	// Apply time dilation to delta time
	deltaTime := baseDeltaTime * timeScale

	// Handle restart key
	if g.input.IsRestartPressed() {
		err := g.model.Reset()
		if err != nil {
			return err
		}
		return nil // Exit early since game was reset
	}

	// Handle camera toggle (C key)
	if g.input.IsCameraTogglePressed() {
		g.model.ToggleCameraMode()
	}

	// Handle shadow toggle (S key)
	if g.input.IsShadowTogglePressed() {
		g.model.ToggleShadows()
	}

	// Handle level selection keys (1-9)
	levelKey := g.input.GetLevelKeyPressed()
	if levelKey > 0 {
		err := g.model.LoadLevel(levelKey)
		if err != nil {
			return err
		}
		return nil // Exit early since level was changed
	}

	// Handle scroll zoom
	scrollDelta := g.input.GetScrollDelta()
	if scrollDelta != 0 {
		// Zoom sensitivity
		zoomSpeed := float32(0.1)
		g.model.Camera.AdjustZoom(scrollDelta * zoomSpeed)
	}

	// Handle drag and throw mechanics
	dragInfo := g.input.GetDragInfo()
	if dragInfo.IsReleased && !g.model.Player.IsMoving() {
		// Convert screen drag to world velocity
		camera := g.model.Camera
		startWorld := camera.ScreenToWorld(dragInfo.StartPos, constants.ScreenWidth, constants.ScreenHeight)
		endWorld := camera.ScreenToWorld(dragInfo.CurrentPos, constants.ScreenWidth, constants.ScreenHeight)

		// Calculate throw vector (opposite direction of drag)
		throwVector := f32.Vec2{
			startWorld[0] - endWorld[0],
			startWorld[1] - endWorld[1],
		}

		// Calculate drag distance and limit it to maximum allowed
		dragDistance := float32(math.Sqrt(float64(throwVector[0]*throwVector[0] + throwVector[1]*throwVector[1])))

		// Only throw if drag distance is significant
		if dragDistance > 0.01 {
			// Clamp drag distance to maximum allowed
			if dragDistance > maxDragDistance {
				// Normalize the vector and scale it to max distance
				throwVector[0] = (throwVector[0] / dragDistance) * maxDragDistance
				throwVector[1] = (throwVector[1] / dragDistance) * maxDragDistance
				dragDistance = maxDragDistance
			}

			// Scale velocity based on clamped drag distance
			velocityMultiplier := float32(2.0)
			throwVelocity := f32.Vec2{
				throwVector[0] * velocityMultiplier,
				throwVector[1] * velocityMultiplier,
			}

			// Additional safety: clamp final velocity magnitude
			velocityMagnitude := float32(math.Sqrt(float64(throwVelocity[0]*throwVelocity[0] + throwVelocity[1]*throwVelocity[1])))
			if velocityMagnitude > maxVelocity {
				throwVelocity[0] = (throwVelocity[0] / velocityMagnitude) * maxVelocity
				throwVelocity[1] = (throwVelocity[1] / velocityMagnitude) * maxVelocity
			}

			g.model.Player.Throw(throwVelocity)
		}
	}

	// Reset player acceleration for this frame
	g.model.Player.ResetAcceleration()

	// Update asteroid positions
	g.model.UpdateAsteroids(deltaTime)

	// Apply gravitational forces from all celestial bodies
	for _, body := range g.model.CelestialBodies {
		g.model.Player.ApplyGravitationalForce(body)

		// Check for collision with celestial body
		if g.model.Player.CheckCollisionWithCelestialBody(body) {
			// Check if it's a white hole (victory condition)
			if body.GetType() == Models.CelestialBodyTypeWhiteHole {
				// Victory! Move to next level
				nextLevel := g.model.CurrentLevelNum + 1
				if nextLevel > 9 {
					// If beyond level 9, restart from level 1
					nextLevel = 1
				}
				err := g.model.LoadLevel(nextLevel)
				if err != nil {
					return err
				}
				return nil // Exit early since level was changed
			} else {
				// Player hit a planet or black hole - reset the level
				err := g.model.Reset()
				if err != nil {
					return err
				}
				return nil // Exit early since game was reset
			}
		}
	}

	// Check for collisions with asteroids
	for _, asteroid := range g.model.RingAsteroids {
		if g.model.Player.CheckCollisionWithCelestialBody(asteroid) {
			// Player hit an asteroid - reset the level
			err := g.model.Reset()
			if err != nil {
				return err
			}
			return nil // Exit early since game was reset
		}
	}

	// Update player physics
	g.model.Player.Update(deltaTime, timeScale)

	// Update camera behavior based on camera mode setting
	switch g.model.CameraMode {
	case Models.CameraModeCenter:
		// Always follow the center of all entities
		levelCenter := g.model.CalculateLevelCenter()
		g.model.Camera.SetTarget(levelCenter)
	case Models.CameraModePlayer:
		// Follow player when moving, center view when idle
		if g.model.Player.IsMoving() {
			g.model.Camera.SetTarget(g.model.Player.Position)
		} else {
			levelCenter := g.model.CalculateLevelCenter()
			g.model.Camera.SetTarget(levelCenter)
		}
	}
	g.model.Camera.Update(deltaTime)

	return nil
}

// drawArrowHead draws an arrowhead at the end of a line
func (g *Game) drawArrowHead(screen *ebiten.Image, start, end f32.Vec2, color color.RGBA) {
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

// drawPlayerTrail draws the player's movement trail with fading effect
func (g *Game) drawPlayerTrail(screen *ebiten.Image) {
	trailPoints := g.model.Player.GetTrailPoints()
	if len(trailPoints) < 2 {
		return // Need at least 2 points to draw lines
	}

	camera := g.model.Camera
	now := time.Now()

	// Draw lines between consecutive trail points
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

		// Trail color with fading alpha (cyan-ish color)
		trailColor := color.RGBA{R: 0, G: 255, B: 255, A: alpha}

		// Calculate line width based on age (newer = thicker)
		width := float32(1 + (1.0-ageRatio)*2) // Width from 1 to 3

		// Draw the trail segment
		vector.StrokeLine(screen,
			prevScreen[0], prevScreen[1],
			currScreen[0], currScreen[1],
			width, trailColor, true,
		)
	}
}

// drawCelestialBodyWithImage renders a celestial body using an image texture
func (g *Game) drawCelestialBodyWithImage(screen *ebiten.Image, screenPos f32.Vec2, radius float32, imagePath string) {
	// Load the image
	img := resources.LoadImage(imagePath)
	if img == nil {
		// Fallback to circle if image loading fails
		vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], radius, color.RGBA{R: 255, G: 255, B: 255, A: 255}, true)
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

	// Scale the image to the desired size
	op.GeoM.Scale(float64(scale), float64(scale))

	// Move to final screen position
	op.GeoM.Translate(float64(screenPos[0]), float64(screenPos[1]))

	// Draw the image
	screen.DrawImage(img, op)
}

// drawPlayerWithImage renders the player using an image texture with rotation based on movement
func (g *Game) drawPlayerWithImage(screen *ebiten.Image, screenPos f32.Vec2, radius float32, imagePath string, rotation float64) {
	// Load the image
	img := resources.LoadImage(imagePath)
	if img == nil {
		// Fallback to circle if image loading fails
		vector.DrawFilledCircle(screen, screenPos[0], screenPos[1], radius, color.RGBA{G: 255, A: 255}, true)
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
