package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type App struct {
	screenManager *ScreenManager
}

func NewApp() (*App, error) {
	screenManager := NewScreenManager()

	// Create home screen
	homeScreen := NewHomeScreen(screenManager)
	screenManager.SetHomeScreen(homeScreen)

	// Create settings screen
	settingsScreen := NewSettingsScreen(screenManager)
	screenManager.SetSettingsScreen(settingsScreen)

	// Create game screen wrapper with level support
	gameScreen, err := NewGameScreenWrapper(screenManager)
	if err != nil {
		return nil, err
	}
	screenManager.SetGameScreen(gameScreen)

	// Start with home screen
	screenManager.SetScreen(HomeScreen)

	return &App{
		screenManager: screenManager,
	}, nil
}

func (a *App) Update() error {
	return a.screenManager.Update()
}

func (a *App) Draw(screen *ebiten.Image) {
	a.screenManager.Draw(screen)
}

func (a *App) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	//if g.screenManager. == nil {
	//	s, err := ebiten.NewShader(crtGo)
	//	if err != nil {
	//		panic(fmt.Sprintf("flappy: failed to compiled the CRT shader: %v", err))
	//	}
	//	g.crtShader = s
	//}
	//
	//os := offscreen.Bounds().Size()
	//
	//op := &ebiten.DrawRectShaderOptions{}
	//op.Images[0] = offscreen
	//op.GeoM = geoM
	//screen.DrawRectShader(os.X, os.Y, g.crtShader, op)

	a.screenManager.DrawFinalScreen(screen, offscreen, geoM)

}

func (a *App) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return a.screenManager.Layout(outsideWidth, outsideHeight)
}
