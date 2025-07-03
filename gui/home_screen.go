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
)

type HomeScreenImpl struct {
	starCount                                            int
	screenManager                                        *ScreenManager
	textFaceSource                                       *text.GoTextFaceSource
	layout                                               *ResponsiveLayout
	settingsButtonX, settingsButtonY, settingsButtonSize float32
}

func NewHomeScreen(screenManager *ScreenManager) *HomeScreenImpl {
	// Create text face source
	textFaceSource, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}

	return &HomeScreenImpl{
		starCount:      12, // Example star count
		screenManager:  screenManager,
		textFaceSource: textFaceSource,
		layout:         NewResponsiveLayout(1920, 1080), // Default, updated in Layout
	}
}

func (h *HomeScreenImpl) Update() error {
	// Handle touch/click input
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		fx, fy := float32(x), float32(y)

		// Check if settings button was clicked
		if IsPointInCircle(fx, fy, h.settingsButtonX+h.settingsButtonSize/2, h.settingsButtonY+h.settingsButtonSize/2, h.settingsButtonSize/2) {
			h.screenManager.SetScreen(SettingsScreen)
			return nil
		}

		// Click anywhere else to go to game
		h.screenManager.SetScreen(GameScreen)
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

	// Draw game title in center
	titleText := "TRAJECTORY"
	titleFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetTitleFontSize(),
	}

	titleWidthF, titleHeightF := text.Measure(titleText, titleFace, 0)
	titleWidth := int(titleWidthF)
	titleHeight := int(titleHeightF)

	titleX := (h.layout.Width - titleWidth) / 2
	titleY := (h.layout.Height-titleHeight)/2 - int(h.layout.GetBodyFontSize())

	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(titleX), float64(titleY))
	titleOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
	text.Draw(screen, titleText, titleFace, titleOp)

	// Draw subtitle
	var subtitleText string
	if h.layout.IsMobile() {
		subtitleText = "Tap to start your space journey"
	} else {
		subtitleText = "Click to start your space journey"
	}

	subtitleFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetBodyFontSize(),
	}

	subtitleWidthF, _ := text.Measure(subtitleText, subtitleFace, 0)
	subtitleWidth := int(subtitleWidthF)
	subtitleX := (h.layout.Width - subtitleWidth) / 2
	subtitleY := titleY + titleHeight + int(h.layout.GetBodyFontSize())/2

	subtitleOp := &text.DrawOptions{}
	subtitleOp.GeoM.Translate(float64(subtitleX), float64(subtitleY))
	subtitleOp.ColorScale.ScaleWithColor(color.RGBA{192, 192, 192, 255})
	text.Draw(screen, subtitleText, subtitleFace, subtitleOp)

	// Draw instructions at bottom
	var instructionText string
	if h.layout.IsMobile() {
		instructionText = "Tap settings ⚙️ for options"
	} else {
		instructionText = "Click settings ⚙️ for options"
	}

	instructionFace := &text.GoTextFace{
		Source: h.textFaceSource,
		Size:   h.layout.GetSmallFontSize(),
	}

	instructionWidthF, _ := text.Measure(instructionText, instructionFace, 0)
	instructionWidth := int(instructionWidthF)
	instructionX := (h.layout.Width - instructionWidth) / 2
	instructionY := h.layout.Height - margin - int(h.layout.GetSmallFontSize())

	instructionOp := &text.DrawOptions{}
	instructionOp.GeoM.Translate(float64(instructionX), float64(instructionY))
	instructionOp.ColorScale.ScaleWithColor(color.RGBA{128, 128, 128, 255})
	text.Draw(screen, instructionText, instructionFace, instructionOp)
}

func (h *HomeScreenImpl) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	h.layout = NewResponsiveLayout(outsideWidth, outsideHeight)
	return outsideWidth, outsideHeight
}

func (h *HomeScreenImpl) SetStarCount(count int) {
	h.starCount = count
}
