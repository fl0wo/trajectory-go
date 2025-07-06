package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/you/trajectory/constants"
)

type ScreenType int

const (
	InitScreenType ScreenType = iota
	HomeScreen
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
	initScreen     Screen
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
	case InitScreenType:
		sm.currentScreen = sm.initScreen
	case HomeScreen:
		sm.currentScreen = sm.homeScreen
	case GameScreen:
		sm.currentScreen = sm.gameScreen
	case SettingsScreen:
		sm.currentScreen = sm.settingsScreen
	}
}

func (sm *ScreenManager) SetInitScreen(screen Screen) {
	sm.initScreen = screen
	if sm.currentType == InitScreenType {
		sm.currentScreen = screen
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
	return constants.ScreenWidth, constants.ScreenHeight
}

func (sm *ScreenManager) GetCurrentType() ScreenType {
	return sm.currentType
}

func (sm *ScreenManager) SetScreenWithLevel(screenType ScreenType, levelNum int) {
	sm.currentType = screenType

	switch screenType {
	case InitScreenType:
		sm.currentScreen = sm.initScreen
	case HomeScreen:
		sm.currentScreen = sm.homeScreen
	case GameScreen:
		// Create a new game screen with the specified level
		if gameScreen, ok := sm.gameScreen.(GameScreenWithLevel); ok {
			err := gameScreen.LoadLevel(levelNum)
			if err != nil {
				// Handle error (e.g., log it, show an error message)
				print("Error loading level:", err)
				return
			}
			sm.currentScreen = gameScreen
		} else {
			sm.currentScreen = sm.gameScreen
		}
	case SettingsScreen:
		sm.currentScreen = sm.settingsScreen
	}
}

// GameScreenWithLevel Interface for game screens that support level loading
type GameScreenWithLevel interface {
	Screen
	LoadLevel(levelNum int) error
}
