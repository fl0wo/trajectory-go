package gui

import (
	"bytes"
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	Models "github.com/you/trajectory/space/model"
)

type LevelSquare struct {
	X, Y, Size float32
	LevelNum   int
	IsLocked   bool
}

type HomeScreenImpl struct {
	starCount                                            int
	screenManager                                        *ScreenManager
	textFaceSource                                       *text.GoTextFaceSource
	layout                                               *ResponsiveLayout
	settingsButtonX, settingsButtonY, settingsButtonSize float32
	levelSquares                                         []LevelSquare
}

func NewHomeScreen(screenManager *ScreenManager) *HomeScreenImpl {
	// Create text face source
	textFaceSource, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}

	homeScreen := &HomeScreenImpl{
		starCount:      12, // Example star count
		screenManager:  screenManager,
		textFaceSource: textFaceSource,
		layout:         NewResponsiveLayout(1920, 1080), // Default, updated in Layout
	}

	homeScreen.generateLevelSquares()
	return homeScreen
}

func (h *HomeScreenImpl) generateLevelSquares() {
	// Get total number of levels from PredefinedLevels
	totalLevels := len(Models.PredefinedLevels)
	h.levelSquares = make([]LevelSquare, totalLevels)

	// For now, all levels are unlocked (you can add logic for locked levels later)
	for i := 1; i <= totalLevels; i++ {
		h.levelSquares[i-1] = LevelSquare{
			LevelNum: i,
			IsLocked: false, // All unlocked for now
		}
	}
}

func (h *HomeScreenImpl) updateLevelSquarePositions() {
	if len(h.levelSquares) == 0 {
		return
	}

	squareSize := float32(h.layout.GetButtonSize()) * 1.2
	margin := float32(h.layout.GetMargin())

	// Calculate grid layout (3 columns for 9 levels)
	cols := 3
	rows := (len(h.levelSquares) + cols - 1) / cols

	// Calculate starting position to center the grid
	totalWidth := float32(cols)*squareSize + float32(cols-1)*margin
	totalHeight := float32(rows)*squareSize + float32(rows-1)*margin

	startX := (float32(h.layout.Width) - totalWidth) / 2
	startY := (float32(h.layout.Height) - totalHeight) / 2

	// Position each square
	for i := range h.levelSquares {
		row := i / cols
		col := i % cols

		h.levelSquares[i].X = startX + float32(col)*(squareSize+margin)
		h.levelSquares[i].Y = startY + float32(row)*(squareSize+margin)
		h.levelSquares[i].Size = squareSize
	}
}

func (h *HomeScreenImpl) Update() error {
	// Update level square positions based on current layout
	h.updateLevelSquarePositions()

	// Handle touch/click input
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		fx, fy := float32(x), float32(y)

		// Check if settings button was clicked
		if IsPointInCircle(fx, fy, h.settingsButtonX+h.settingsButtonSize/2, h.settingsButtonY+h.settingsButtonSize/2, h.settingsButtonSize/2) {
			h.screenManager.SetScreen(SettingsScreen)
			return nil
		}

		// Check if any level square was clicked
		for _, square := range h.levelSquares {
			if !square.IsLocked && IsPointInRect(fx, fy, square.X, square.Y, square.Size, square.Size) {
				// Load the selected level and go to game screen
				h.screenManager.SetScreenWithLevel(GameScreen, square.LevelNum)
				return nil
			}
		}
	}

	return nil
}

func (h *HomeScreenImpl) Draw(screen *ebiten.Image) {
	// Fill screen with black background
	screen.Fill(color.RGBA{A: 255})

	margin := h.layout.GetMargin()

	// Draw star count in top-left corner
	starText := fmt.Sprintf("⭐ %d", h.starCount)
	starFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetBodyFontSize(),
	}

	starOp := &text.DrawOptions{}
	starOp.GeoM.Translate(float64(margin), float64(margin))
	starOp.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, 255}) // Gold color
	text.Draw(screen, starText, starFace, starOp)

	// Draw settings icon in top-right corner
	h.settingsButtonSize = float32(h.layout.GetButtonSize())
	h.settingsButtonX = float32(h.layout.Width-margin) - h.settingsButtonSize
	h.settingsButtonY = float32(margin)

	DrawSettingsIcon(screen, h.settingsButtonX, h.settingsButtonY, h.settingsButtonSize, color.RGBA{255, 255, 255, 255})

	// Draw game title at top
	titleText := "TRAJECTORY"
	titleFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetTitleFontSize(),
	}

	titleWidthF, _ := text.Measure(titleText, titleFace, 0)
	titleWidth := int(titleWidthF)

	titleX := (h.layout.Width - titleWidth) / 2
	titleY := margin * 3

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(titleX), float64(titleY))
	titleOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
	text.Draw(screen, titleText, titleFace, titleOp)

	// Draw subtitle
	var subtitleText string
	if h.layout.IsMobile() {
		subtitleText = "Select a level to play"
	} else {
		subtitleText = "Select a level to play"
	}

	subtitleFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetBodyFontSize(),
	}

	subtitleWidthF, _ := text.Measure(subtitleText, subtitleFace, 0)
	subtitleWidth := int(subtitleWidthF)
	subtitleX := (h.layout.Width - subtitleWidth) / 2
	subtitleY := titleY + int(h.layout.GetTitleFontSize()) + margin/2

	subtitleOp := &text.DrawOptions{}
	subtitleOp.GeoM.Translate(float64(subtitleX), float64(subtitleY))
	subtitleOp.ColorScale.ScaleWithColor(color.RGBA{192, 192, 192, 255})
	text.Draw(screen, subtitleText, subtitleFace, subtitleOp)

	// Draw level squares
	levelNumberFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetBodyFontSize(),
	}

	for _, square := range h.levelSquares {
		DrawLevelSquare(screen, square.X, square.Y, square.Size, square.LevelNum, levelNumberFace, square.IsLocked)
	}
}

func (h *HomeScreenImpl) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	h.layout = NewResponsiveLayout(outsideWidth, outsideHeight)
	return outsideWidth, outsideHeight
}

func (h *HomeScreenImpl) SetStarCount(count int) {
	h.starCount = count
}
