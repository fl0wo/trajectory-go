package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type ScreenType int

const (
	HomeScreen ScreenType = iota
	GameScreen
	SettingsScreen
)

type Screen interface {
	Update() error
	Draw(screen *ebiten.Image)
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}

type ScreenManager struct {
	currentScreen  Screen
	currentType    ScreenType
	gameScreen     Screen
	homeScreen     Screen
	settingsScreen Screen
}

func NewScreenManager() *ScreenManager {
	return &ScreenManager{
		currentType: HomeScreen,
	}
}

func (sm *ScreenManager) SetScreen(screenType ScreenType) {
	sm.currentType = screenType

	switch screenType {
	case HomeScreen:
		sm.currentScreen = sm.homeScreen
	case GameScreen:
		sm.currentScreen = sm.gameScreen
	case SettingsScreen:
		sm.currentScreen = sm.settingsScreen
	}
}

func (sm *ScreenManager) SetHomeScreen(screen Screen) {
	sm.homeScreen = screen
	if sm.currentType == HomeScreen {
		sm.currentScreen = screen
	}
}

func (sm *ScreenManager) SetGameScreen(screen Screen) {
	sm.gameScreen = screen
}

func (sm *ScreenManager) SetSettingsScreen(screen Screen) {
	sm.settingsScreen = screen
}

func (sm *ScreenManager) Update() error {
	if sm.currentScreen != nil {
		return sm.currentScreen.Update()
	}
	return nil
}

func (sm *ScreenManager) Draw(screen *ebiten.Image) {
	if sm.currentScreen != nil {
		sm.currentScreen.Draw(screen)
	}
}

func (sm *ScreenManager) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if sm.currentScreen != nil {
		return sm.currentScreen.Layout(outsideWidth, outsideHeight)
	}
	return 1920, 1080
}

func (sm *ScreenManager) GetCurrentType() ScreenType {
	return sm.currentType
}

func (sm *ScreenManager) SetScreenWithLevel(screenType ScreenType, levelNum int) {
	sm.currentType = screenType

	switch screenType {
	case HomeScreen:
		sm.currentScreen = sm.homeScreen
	case GameScreen:
		// Create a new game screen with the specified level
		if gameScreen, ok := sm.gameScreen.(GameScreenWithLevel); ok {
			gameScreen.LoadLevel(levelNum)
			sm.currentScreen = gameScreen
		} else {
			sm.currentScreen = sm.gameScreen
		}
	case SettingsScreen:
		sm.currentScreen = sm.settingsScreen
	}
}

// Interface for game screens that support level loading
type GameScreenWithLevel interface {
	Screen
	LoadLevel(levelNum int) error
}
