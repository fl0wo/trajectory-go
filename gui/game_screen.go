package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/you/trajectory/space"
)

// GameScreenWrapper wraps the existing game with level loading capability
type GameScreenWrapper struct {
	game *space.Game
}

func NewGameScreenWrapper() (*GameScreenWrapper, error) {
	game, err := space.NewGame()
	if err != nil {
		return nil, err
	}

	return &GameScreenWrapper{
		game: game,
	}, nil
}

func (gsw *GameScreenWrapper) LoadLevel(levelNum int) error {
	// Create a fresh game instance to ensure clean state
	newGame, err := space.NewGame()
	if err != nil {
		return err
	}

	// Load the specific level into the fresh game instance
	err = newGame.LoadLevel(levelNum)
	if err != nil {
		return err
	}

	// Replace the existing game with the fresh instance
	gsw.game = newGame

	return nil
}

func (gsw *GameScreenWrapper) Update() error {
	return gsw.game.Update()
}

func (gsw *GameScreenWrapper) Draw(screen *ebiten.Image) {
	gsw.game.Draw(screen)
}

func (gsw *GameScreenWrapper) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return gsw.game.Layout(outsideWidth, outsideHeight)
}
