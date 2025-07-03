package gui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/you/trajectory/constants"
)

type HomeScreenImpl struct {
	starCount     int
	screenManager *ScreenManager
}

func NewHomeScreen(screenManager *ScreenManager) *HomeScreenImpl {
	return &HomeScreenImpl{
		starCount:     12, // Example star count
		screenManager: screenManager,
	}
}

func (h *HomeScreenImpl) Update() error {
	// Handle input for settings button (top-right corner click)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		// Settings button area (top-right corner, roughly 100x100 pixels)
		if x >= constants.ScreenWidth-100 && y <= 100 {
			// TODO: Switch to settings screen when implemented
			// h.screenManager.SetScreen(SettingsScreen)
		}

		// Click anywhere else to go to game (temporary for testing)
		if x < constants.ScreenWidth-100 || y > 100 {
			h.screenManager.SetScreen(GameScreen)
		}
	}

	return nil
}

func (h *HomeScreenImpl) Draw(screen *ebiten.Image) {
	// Fill screen with black background
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw star count in top-left corner
	starText := fmt.Sprintf("Stars: %d", h.starCount)
	ebitenutil.DebugPrintAt(screen, starText, 20, 20)

	// Draw settings icon in top-right corner (simple text for now)
	settingsText := "⚙️"
	ebitenutil.DebugPrintAt(screen, settingsText, constants.ScreenWidth-80, 20)

	// Draw instructions in center
	instructionsText := "Click anywhere to start game\nClick settings (⚙️) for options"
	ebitenutil.DebugPrintAt(screen, instructionsText, constants.ScreenWidth/2-100, constants.ScreenHeight/2)
}

func (h *HomeScreenImpl) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return constants.ScreenWidth, constants.ScreenHeight
}

func (h *HomeScreenImpl) SetStarCount(count int) {
	h.starCount = count
}
