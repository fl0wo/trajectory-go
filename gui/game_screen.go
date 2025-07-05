package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/you/trajectory/space"
	"github.com/you/trajectory/space/colors"
)

// GameScreenWrapper wraps the existing game with level loading capability
type GameScreenWrapper struct {
	game                                     *space.Game
	screenManager                            *ScreenManager
	backButtonX, backButtonY, backButtonSize float32
}

func NewGameScreenWrapper(screenManager *ScreenManager) (*GameScreenWrapper, error) {
	game, err := space.NewGame()
	if err != nil {
		return nil, err
	}

	return &GameScreenWrapper{
		game:          game,
		screenManager: screenManager,
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
	// Handle back button click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		fx, fy := float32(x), float32(y)

		// Check if back button was clicked
		if IsPointInCircle(fx, fy, gsw.backButtonX+gsw.backButtonSize/2, gsw.backButtonY+gsw.backButtonSize/2, gsw.backButtonSize/2) {
			gsw.screenManager.SetScreen(HomeScreen)
			return nil
		}
	}

	return gsw.game.Update()
}

func (gsw *GameScreenWrapper) Draw(screen *ebiten.Image) {
	// Draw the game first
	gsw.game.Draw(screen)

	// Draw back button in top-left corner
	margin := float32(20)
	gsw.backButtonSize = 50
	gsw.backButtonX = margin
	gsw.backButtonY = margin

	DrawBackButton(screen, gsw.backButtonX, gsw.backButtonY, gsw.backButtonSize, colors.BorderPreviewLevel)
}

func (gsw *GameScreenWrapper) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return gsw.game.Layout(outsideWidth, outsideHeight)
}
