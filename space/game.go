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
		if screenPos[0] > -radius && screenPos[0] < ScreenWidth+radius &&
			screenPos[1] > -radius && screenPos[1] < ScreenHeight+radius {
			vector.DrawFilledCircle(screen,
				screenPos[0],
				screenPos[1],
				radius,
				color.White, true,
			)
		}
	}

	// Draw the player with camera transform
	player := g.model.Player
	playerScreenPos := camera.WorldToScreen(player.Position, ScreenWidth, ScreenHeight)
	playerSize := f32.Vec2{
		player.Size[0] * ScreenWidth / 100 * camera.GetZoom(),
		player.Size[1] * ScreenHeight / 100 * camera.GetZoom(),
	}

	vector.DrawFilledRect(screen,
		playerScreenPos[0]-playerSize[0]/2, // Center the rectangle
		playerScreenPos[1]-playerSize[1]/2, // Center the rectangle
		playerSize[0],
		playerSize[1],
		color.RGBA{G: 255, A: 255}, true,
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

	// Update player physics
	g.model.Player.Update(deltaTime)

	// Update camera to follow player
	g.model.Camera.SetTarget(g.model.Player.Position)
	g.model.Camera.Update(deltaTime)

	return nil
}
