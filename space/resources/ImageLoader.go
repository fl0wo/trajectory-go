package resources

import (
	"embed"
	"fmt"
	"image"
	_ "image/png" // For PNG support
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed 256/*.png
var imageFiles embed.FS

// ImageCache stores loaded images to avoid reloading
var imageCache = make(map[string]*ebiten.Image)

// LoadImage loads an image from the embedded resources (uses default 256x256 size)
func LoadImage(filename string) *ebiten.Image {
	return LoadImageWithSize(filename, 256)
}

// LoadImageWithSize loads an image from the embedded resources with specified size
// size should be 32, 64, 128, or 256. Falls back to original image if size folder doesn't exist
func LoadImageWithSize(filename string, size int) *ebiten.Image {
	// Create cache key that includes size
	cacheKey := fmt.Sprintf("%s_%d", filename, size)

	// Check if image is already cached
	if img, exists := imageCache[cacheKey]; exists {
		return img
	}

	// Determine the file path based on size
	var filePath string
	switch size {
	case 32, 64, 128, 256:
		filePath = fmt.Sprintf("%d/%s", size, filename)
	default:
		// Default to 256 for invalid sizes
		filePath = fmt.Sprintf("256/%s", filename)
	}

	// Try to load from size-specific folder first
	file, err := imageFiles.Open(filePath)
	if err != nil {
		// Fallback to original image if size-specific version doesn't exist
		file, err = imageFiles.Open(filename)
		if err != nil {
			log.Printf("Failed to open image file %s (tried %s): %v", filename, filePath, err)
			return nil
		}
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Failed to decode image %s: %v", filePath, err)
		return nil
	}

	// Convert to Ebitengine image
	ebitenImg := ebiten.NewImageFromImage(img)

	// Cache the image with size-specific key
	imageCache[cacheKey] = ebitenImg

	return ebitenImg
}

// GetImageSize returns the size of an image without loading it into cache (uses default 256x256 size)
func GetImageSize(filename string) (int, int) {
	return GetImageSizeWithSize(filename, 256)
}

// GetImageSizeWithSize returns the size of an image with specified size without loading it into cache
func GetImageSizeWithSize(filename string, size int) (int, int) {
	img := LoadImageWithSize(filename, size)
	if img == nil {
		return 0, 0
	}
	return img.Bounds().Size().X, img.Bounds().Size().Y
}

// Planet image assets
const (
	EarthImage   = "earth.png"
	JupiterImage = "jupiter.png"
	MarsImage    = "mars.png"
	PlanetBImage = "planet_b.png"
)

// Asteroid image assets
const (
	Asteroid1Image = "asteroid1.png"
	Asteroid2Image = "asteroid2.png"
)

// Player image assets
const (
	AstronautImage = "astronaut.png"
)

// Black hole image assets
const (
	BlackHoleImage = "blackhole.png"
)

// Helper functions for loading different sized images
func LoadImageSmall(filename string) *ebiten.Image {
	return LoadImageWithSize(filename, 32)
}

func LoadImageMedium(filename string) *ebiten.Image {
	return LoadImageWithSize(filename, 64)
}

func LoadImageLarge(filename string) *ebiten.Image {
	return LoadImageWithSize(filename, 128)
}

func LoadImageXLarge(filename string) *ebiten.Image {
	return LoadImageWithSize(filename, 256)
}
