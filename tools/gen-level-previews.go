package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space"
	Models "github.com/you/trajectory/space/model"
)

const (
	// Preview image dimensions (16:9 ratio)
	PreviewHeight = 256
	PreviewWidth  = int(PreviewHeight * 16 / 9) // 455 pixels wide

	// Output directory (relative to project root)
	OutputDir = "../gui/resources"
)

func main() {
	fmt.Println("Generating level preview images...")

	// Ensure output directory exists
	if err := os.MkdirAll(OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Get total number of levels
	totalLevels := len(Models.PredefinedLevels)
	fmt.Printf("Found %d levels to generate previews for\n", totalLevels)

	// Create a batch preview generator that handles all levels in one game loop
	generator := &BatchPreviewGenerator{
		totalLevels:  totalLevels,
		currentLevel: 1,
		outputDir:    OutputDir,
		state:        "loading",
	}

	// Run the batch generator
	err := ebiten.RunGame(generator)
	if err != nil && err.Error() != "all previews generated" {
		log.Fatalf("Failed to generate previews: %v", err)
	}

	fmt.Printf("Preview generation completed! Generated %d level previews in %s\n", totalLevels, OutputDir)
}

// BatchPreviewGenerator generates all level previews in a single game loop
type BatchPreviewGenerator struct {
	totalLevels   int
	currentLevel  int
	outputDir     string
	state         string // "loading", "capturing", "saving", "done"
	game          *space.Game
	frameCount    int
	targetFrames  int
	capturedImage *ebiten.Image
}

// Update handles the state machine for generating previews
func (b *BatchPreviewGenerator) Update() error {
	switch b.state {
	case "loading":
		// Load the current level
		fmt.Printf("Generating preview for level %d...\n", b.currentLevel)

		// Create new game instance
		game, err := space.NewGame()
		if err != nil {
			log.Printf("Failed to create game for level %d: %v", b.currentLevel, err)
			b.moveToNextLevel()
			return nil
		}

		// Load the specific level
		err = game.LoadLevel(b.currentLevel)
		if err != nil {
			log.Printf("Failed to load level %d: %v", b.currentLevel, err)
			b.moveToNextLevel()
			return nil
		}

		b.game = game
		b.frameCount = 0
		b.targetFrames = 5 // Let the game stabilize for 5 frames
		b.state = "capturing"

	case "capturing":
		if b.frameCount < b.targetFrames {
			// Update the game to let systems stabilize
			err := b.game.Update()
			if err != nil {
				log.Printf("Failed to update game for level %d: %v", b.currentLevel, err)
				b.moveToNextLevel()
				return nil
			}
			b.frameCount++
		} else {
			b.state = "saving"
		}

	case "saving":
		// Save the captured image
		if b.capturedImage != nil {
			b.saveCurrentPreview()
		}
		b.moveToNextLevel()

	case "done":
		return fmt.Errorf("all previews generated")
	}

	return nil
}

// Draw renders the current level and captures the frame when ready
func (b *BatchPreviewGenerator) Draw(screen *ebiten.Image) {
	if b.state == "capturing" && b.game != nil {
		// Draw the game
		b.game.Draw(screen)

		buf := ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
		// Draw the game state to a buffer
		b.game.Draw(buf)

		b.game.DrawFinalScreen(screen, buf, ebiten.GeoM{})

		// Capture the frame when we've reached the target
		if b.frameCount >= b.targetFrames && b.capturedImage == nil {
			// Create a copy of the screen
			b.capturedImage = ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
			b.capturedImage.DrawImage(screen, nil)
		}
	}
}

// Layout returns the game's layout
func (b *BatchPreviewGenerator) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return constants.ScreenWidth, constants.ScreenHeight
}

// moveToNextLevel advances to the next level or finishes
func (b *BatchPreviewGenerator) moveToNextLevel() {
	b.currentLevel++
	b.game = nil
	b.capturedImage = nil

	if b.currentLevel > b.totalLevels {
		b.state = "done"
	} else {
		b.state = "loading"
	}
}

// saveCurrentPreview saves the captured image as PNG
func (b *BatchPreviewGenerator) saveCurrentPreview() {
	if b.capturedImage == nil {
		log.Printf("No image captured for level %d", b.currentLevel)
		return
	}

	// Convert Ebitengine image to standard image
	bounds := b.capturedImage.Bounds()

	// Scale down to preview size
	previewImg := image.NewRGBA(image.Rect(0, 0, PreviewWidth, PreviewHeight))

	// Simple scaling by sampling
	scaleX := float64(bounds.Dx()) / float64(PreviewWidth)
	scaleY := float64(bounds.Dy()) / float64(PreviewHeight)

	for y := 0; y < PreviewHeight; y++ {
		for x := 0; x < PreviewWidth; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)

			if srcX < bounds.Max.X && srcY < bounds.Max.Y {
				r, g, b, a := b.capturedImage.At(srcX, srcY).RGBA()
				previewImg.SetRGBA(x, y, color.RGBA{
					R: uint8(r >> 8),
					G: uint8(g >> 8),
					B: uint8(b >> 8),
					A: uint8(a >> 8),
				})
			}
		}
	}

	// Save as PNG
	filename := fmt.Sprintf("level%d.png", b.currentLevel)
	filePath := filepath.Join(b.outputDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create file for level %d: %v", b.currentLevel, err)
		return
	}
	defer file.Close()

	err = png.Encode(file, previewImg)
	if err != nil {
		log.Printf("Failed to encode PNG for level %d: %v", b.currentLevel, err)
		return
	}

	fmt.Printf("Saved preview: %s\n", filePath)
}
