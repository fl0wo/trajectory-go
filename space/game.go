package space

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/gamecontrol"
	"github.com/you/trajectory/space/input"
	Models "github.com/you/trajectory/space/model"
	"github.com/you/trajectory/space/rendering"
)

type Game struct {
	input      *input.Input
	model      *Models.SpaceGame
	lastTime   int64 // For delta time calculation
	renderer   *rendering.Renderer
	controller *gamecontrol.GameController
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return constants.ScreenWidth, constants.ScreenHeight
}

// Draw draws the current game to the given screen.
func (g *Game) Draw(screen *ebiten.Image) {
	// Use the renderer to draw everything
	g.renderer.Draw(screen, g.model)

	// Draw trajectory arrow if dragging (needs input info)
	dragInfo := g.input.GetDragInfo()
	if dragInfo.IsDragging {
		g.renderer.DrawTrajectoryArrow(screen, g.model, dragInfo.StartPos, dragInfo.CurrentPos, true)
	}
}

// NewGame generates a new Game object.
func NewGame() (*Game, error) {
	g := &Game{
		input:      input.NewInput(),
		renderer:   rendering.NewRenderer(),
		controller: gamecontrol.NewGameController(),
	}
	var err error
	g.model, err = Models.NewSpaceGame()
	if err != nil {
		return nil, err
	}

	return g, nil
}

// LoadLevel loads a specific level into the game
func (g *Game) LoadLevel(levelNum int) error {
	// Use the existing LoadLevel method from the model
	return g.model.LoadLevel(levelNum)
}

// Update updates the current game state.
func (g *Game) Update() error {
	// Calculate delta time
	baseDeltaTime := float32(1.0 / 60.0) // Assume 60 FPS for simplicity

	g.input.Update()

	// Calculate time dilation and proximity effects
	g.controller.CalculateTimeDilation(g.model, baseDeltaTime)
	timeScale := g.model.TimeScale

	// Apply time dilation to delta time
	deltaTime := baseDeltaTime * timeScale

	// Process input
	err := g.controller.ProcessInput(g.model, g.input)
	if err != nil {
		return err
	}

	// Update physics
	err = g.controller.UpdatePhysics(g.model, deltaTime, timeScale)
	if err != nil {
		return err
	}

	return nil
}
