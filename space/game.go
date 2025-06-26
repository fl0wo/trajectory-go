package space

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/you/trajectory/space/input"
	Models "github.com/you/trajectory/space/model"
	"image/color"
)

const (
	ScreenWidth  = 2080
	ScreenHeight = 1080
)

var (
	backgroundColor = color.Black
)

type Game struct {
	input *input.Input
	model *Models.SpaceGame
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

// Draw draws the current game to the given screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(backgroundColor)
	// vector.DrawFilledCircle(screen, 100, 100, 50, color.White, true)

	for _, planet := range g.model.Planets {
		// Draw each planet as a filled circle
		vector.DrawFilledCircle(screen,
			planet.Position[0]*ScreenWidth,
			planet.Position[1]*ScreenHeight,
			planet.Radius*ScreenWidth/100, // Scale radius to screen size
			color.White, true,
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
	g.input.Update()
	// Update game logic here
	// For example, move planets, check collisions, etc.
	return nil
}
