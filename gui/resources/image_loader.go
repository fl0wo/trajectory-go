package resources

import (
	"embed"
	"fmt"
	"image"
	_ "image/png" // For PNG support
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed *.png
var levelPreviewFiles embed.FS

// ImageCache stores loaded level preview images to avoid reloading
var levelPreviewCache = make(map[int]*ebiten.Image)

// LoadLevelPreview loads a level preview image from the embedded resources
func LoadLevelPreview(levelNum int) *ebiten.Image {
	// Check if image is already cached
	if img, exists := levelPreviewCache[levelNum]; exists {
		return img
	}

	// Determine the filename
	filename := fmt.Sprintf("level%d.png", levelNum)

	// Try to load the file
	file, err := levelPreviewFiles.Open(filename)
	if err != nil {
		log.Printf("Failed to open level preview file %s: %v", filename, err)
		return nil
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Failed to decode level preview %s: %v", filename, err)
		return nil
	}

	// Convert to Ebitengine image
	ebitenImg := ebiten.NewImageFromImage(img)

	// Cache the image
	levelPreviewCache[levelNum] = ebitenImg

	return ebitenImg
}

// GetLevelPreviewSize returns the size of a level preview image without loading it into cache
func GetLevelPreviewSize(levelNum int) (int, int) {
	img := LoadLevelPreview(levelNum)
	if img == nil {
		return 0, 0
	}
	return img.Bounds().Size().X, img.Bounds().Size().Y
}
