package space

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/you/trajectory/space/input"
	Models "github.com/you/trajectory/space/model"
	"golang.org/x/image/math/f32"
	"image/color"
	"math"
)

const (
	ScreenWidth  = 2080
	ScreenHeight = 1080
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
	return ScreenWidth, ScreenHeight
}

// Draw draws the current game to the given screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)

	camera := g.model.Camera

	// Draw planets with camera transform
	for _, planet := range g.model.Planets {
		screenPos := camera.WorldToScreen(planet.Position, ScreenWidth, ScreenHeight)
		radius := planet.Radius * ScreenWidth / 100 * camera.GetZoom()

		// Only draw if on screen (simple culling)
		if screenPos[0] > -radius && screenPos[0] < ScreenWidth+radius && screenPos[1] > -radius && screenPos[1] < ScreenHeight+radius {
			vector.DrawFilledCircle(
				screen,
				screenPos[0],
				screenPos[1],
				radius,
				color.White, true,
			)
			// Draw planet's orbit radius as a dashed circle
			orbitRadius := planet.OrbitRadius * ScreenWidth / 100 * camera.GetZoom()
			if orbitRadius > 0 {
				vector.StrokeCircle(
					screen,
					screenPos[0],
					screenPos[1],
					orbitRadius,
					2,                                        // Line width
					color.RGBA{R: 255, G: 255, B: 0, A: 128}, // Yellow with some transparency
					true,                                     // Closed
				)
			}
		}
	}

	// Draw the player as a green circle with camera transform
	player := g.model.Player
	playerScreenPos := camera.WorldToScreen(player.Position, ScreenWidth, ScreenHeight)
	playerRadius := player.Radius * ScreenWidth / 100 * camera.GetZoom()

	vector.DrawFilledCircle(screen,
		playerScreenPos[0],
		playerScreenPos[1],
		playerRadius,
		color.RGBA{G: 255, A: 255}, true, // Green circle
	)

	// Draw drag line if dragging
	dragInfo := g.input.GetDragInfo()
	if dragInfo.IsDragging {
		// Convert screen drag to world space for visual feedback
		startWorld := camera.ScreenToWorld(dragInfo.StartPos, ScreenWidth, ScreenHeight)
		currentWorld := camera.ScreenToWorld(dragInfo.CurrentPos, ScreenWidth, ScreenHeight)

		// Draw trajectory line (inverted direction - opposite of drag)
		trajectoryVector := f32.Vec2{
			startWorld[0] - currentWorld[0],
			startWorld[1] - currentWorld[1],
		}

		endWorld := f32.Vec2{
			startWorld[0] + trajectoryVector[0],
			startWorld[1] + trajectoryVector[1],
		}

		startScreen := camera.WorldToScreen(startWorld, ScreenWidth, ScreenHeight)
		endScreen := camera.WorldToScreen(endWorld, ScreenWidth, ScreenHeight)

		vector.StrokeLine(screen,
			startScreen[0], startScreen[1],
			endScreen[0], endScreen[1],
			2, color.RGBA{R: 255, A: 255}, true,
		)
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
		startWorld := camera.ScreenToWorld(dragInfo.StartPos, ScreenWidth, ScreenHeight)
		endWorld := camera.ScreenToWorld(dragInfo.CurrentPos, ScreenWidth, ScreenHeight)

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

	// Apply gravitational forces from all planets
	for _, planet := range g.model.Planets {
		g.model.Player.ApplyGravitationalForce(planet, deltaTime)

		// Check for collision with planet
		if g.model.Player.CheckCollisionWithPlanet(planet) {
			// Player hit a planet - reset the game
			err := g.model.Reset()
			if err != nil {
				return err
			}
			return nil // Exit early since game was reset
		}
	}

	// Update player physics
	g.model.Player.Update(deltaTime)

	// Update camera to follow player
	g.model.Camera.SetTarget(g.model.Player.Position)
	g.model.Camera.Update(deltaTime)

	return nil
}
