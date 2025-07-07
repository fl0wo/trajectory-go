package gui

import (
	"bytes"
	"github.com/you/trajectory/constants"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type SettingsScreenImpl struct {
	screenManager *ScreenManager
	textFace      *text.GoTextFace
	screenWidth   int
	screenHeight  int
}

func NewSettingsScreen(screenManager *ScreenManager) *SettingsScreenImpl {
	// Create text face
	textFaceSource, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		log.Fatal(err)
	}

	textFace := &text.GoTextFace{
		Source: textFaceSource,
		Size:   48,
	}

	return &SettingsScreenImpl{
		screenManager: screenManager,
		textFace:      textFace,
		screenWidth:   constants.ScreenWidth,
		screenHeight:  constants.ScreenHeight,
	}
}

func (s *SettingsScreenImpl) Update() error {
	// Handle touch/click to go back to home screen
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		s.screenManager.SetScreen(HomeScreen)
	}

	return nil
}

func (s *SettingsScreenImpl) Draw(screen *ebiten.Image) {
	// Fill screen with black background
	screen.Fill(color.RGBA{A: 255})

	// Draw "Settings" text in center
	settingsText := "Settings"

	// Get text bounds for centering
	textWidthF, textHeightF := text.Measure(settingsText, s.textFace, 0)
	textWidth := int(textWidthF)
	textHeight := int(textHeightF)

	// Calculate center position
	x := (s.screenWidth - textWidth) / 2
	y := (s.screenHeight - textHeight) / 2

	// Draw the text
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
	text.Draw(screen, settingsText, s.textFace, op)

	// Draw back instruction
	backText := "Click anywhere to go back"
	backFace := &text.GoTextFace{
		Source: s.textFace.Source,
		Size:   24,
	}

	backWidthF, _ := text.Measure(backText, backFace, 0)
	backWidth := int(backWidthF)
	backX := (s.screenWidth - backWidth) / 2
	backY := y + textHeight + 50

	backOp := &text.DrawOptions{}
	backOp.GeoM.Translate(float64(backX), float64(backY))
	backOp.ColorScale.ScaleWithColor(color.RGBA{192, 192, 192, 255})
	text.Draw(screen, backText, backFace, backOp)
}

func (s *SettingsScreenImpl) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	screen.DrawImage(offscreen, &ebiten.DrawImageOptions{
		GeoM: geoM,
	})
}

func (s *SettingsScreenImpl) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	s.screenWidth = outsideWidth
	s.screenHeight = outsideHeight
	return outsideWidth, outsideHeight
}
