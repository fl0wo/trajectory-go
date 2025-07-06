package main

import (
	"fmt"

	"github.com/you/trajectory/gui/resources"
	Models "github.com/you/trajectory/space/model"
)

func main() {
	fmt.Println("Testing actual level preview loading...")

	// Get total number of levels
	totalLevels := len(Models.PredefinedLevels)
	fmt.Printf("Testing %d actual level previews\n", totalLevels)

	// Test loading each level preview
	for levelNum := 1; levelNum <= totalLevels; levelNum++ {
		img := resources.LoadLevelPreview(levelNum)
		if img != nil {
			width, height := img.Bounds().Dx(), img.Bounds().Dy()
			fmt.Printf("✓ Level %d: %dx%d pixels\n", levelNum, width, height)
		} else {
			fmt.Printf("✗ Level %d: Failed to load\n", levelNum)
		}
	}

	fmt.Println("Actual level preview loading test completed!")
}
