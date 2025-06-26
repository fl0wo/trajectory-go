package space

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/input"
	Models "github.com/you/trajectory/space/model"
	"golang.org/x/image/math/f32"
	"image/color"
	"math"
)

var (
	backgroundColor = color.Black
)

type Game struct {
	input    *input.Input
	model    *Models.SpaceGame
	lastTime int64 // For delta time calculation
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return constants.ScreenWidth, constants.ScreenHeight
}

// Draw draws the current game to the given screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)

	camera := g.model.Camera

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
		}

		// Only draw if on screen (simple culling)
		//if screenPos[0] > -radius && screenPos[0] < ScreenWidth+radius && screenPos[1] > -radius && screenPos[1] < ScreenHeight+radius {
		vector.DrawFilledCircle(
			screen,
			screenPos[0],
			screenPos[1],
			radius,
			bodyColor, true,
		)
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
	//}

	// Draw the player as a green circle with camera transform
	player := g.model.Player
	playerScreenPos := camera.WorldToScreen(player.Position, constants.ScreenWidth, constants.ScreenHeight)
	playerRadius := camera.RadiusToScreen(player.Radius, constants.ScreenWidth, constants.ScreenHeight)

	vector.DrawFilledCircle(screen,
		playerScreenPos[0],
		playerScreenPos[1],
		playerRadius,
		color.RGBA{G: 255, A: 255}, true, // Green circle
	)

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

		// Scale trajectory for better visualization
		velocityMultiplier := float32(2.0)
		trajectoryVector[0] *= velocityMultiplier
		trajectoryVector[1] *= velocityMultiplier

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
	return g, nil
}

// Update updates the current game state.
func (g *Game) Update() error {
	// Calculate delta time
	deltaTime := float32(1.0 / 60.0) // Assume 60 FPS for simplicity

	g.input.Update()

	// Handle restart key
	if g.input.IsRestartPressed() {
		err := g.model.Reset()
		if err != nil {
			return err
		}
		return nil // Exit early since game was reset
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

		// Calculate throw velocity (opposite direction of drag)
		throwVector := f32.Vec2{
			startWorld[0] - endWorld[0],
			startWorld[1] - endWorld[1],
		}

		// Scale velocity based on drag distance (adjust multiplier as needed)
		velocityMultiplier := float32(2.0)
		throwVelocity := f32.Vec2{
			throwVector[0] * velocityMultiplier,
			throwVector[1] * velocityMultiplier,
		}

		// Only throw if drag distance is significant
		dragDistance := float32(math.Sqrt(float64(throwVector[0]*throwVector[0] + throwVector[1]*throwVector[1])))
		if dragDistance > 0.01 {
			g.model.Player.Throw(throwVelocity)
		}
	}

	// Reset player acceleration for this frame
	g.model.Player.ResetAcceleration()

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

	// Update player physics
	g.model.Player.Update(deltaTime)

	// Update camera to follow player
	g.model.Camera.SetTarget(g.model.Player.Position)
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
