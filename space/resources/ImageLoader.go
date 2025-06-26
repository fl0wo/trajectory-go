package resources

import (
	"embed"
	"image"
	_ "image/png" // For PNG support
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed *.png
var imageFiles embed.FS

// ImageCache stores loaded images to avoid reloading
var imageCache = make(map[string]*ebiten.Image)

// LoadImage loads an image from the embedded resources
func LoadImage(filename string) *ebiten.Image {
	// Check if image is already cached
	if img, exists := imageCache[filename]; exists {
		return img
	}

	// Load image from embedded filesystem
	file, err := imageFiles.Open(filename)
	if err != nil {
		log.Printf("Failed to open image file %s: %v", filename, err)
		return nil
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		log.Printf("Failed to decode image %s: %v", filename, err)
		return nil
	}

	// Convert to Ebitengine image
	ebitenImg := ebiten.NewImageFromImage(img)

	// Cache the image
	imageCache[filename] = ebitenImg

	return ebitenImg
}

// GetImageSize returns the size of an image without loading it into cache
func GetImageSize(filename string) (int, int) {
	img := LoadImage(filename)
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
