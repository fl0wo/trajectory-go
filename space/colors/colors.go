package colors

import "image/color"

// Game Background Colors
var (
	// Primary background color for space (#060219)
	Background = color.RGBA{R: 6, G: 2, B: 25, A: 255}

	// Shadow system background
	ShadowBackground = color.RGBA{R: 6, G: 2, B: 25, A: 255}
)

// Celestial Body Colors
var (
	// Planet colors
	PlanetBody  = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White
	PlanetOrbit = color.RGBA{R: 255, G: 255, B: 255, A: 255} // Black

	// Black hole colors
	BlackHoleBody  = color.RGBA{R: 255, G: 0, B: 0, A: 50} // Transparent red
	BlackHoleOrbit = color.RGBA{R: 0, G: 0, B: 0, A: 255}  // Black

	// White hole colors
	WhiteHoleBody  = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White
	WhiteHoleOrbit = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White

	// Asteroid colors
	AsteroidBody    = color.RGBA{R: 139, G: 69, B: 19, A: 255}   // Brown
	AsteroidBodyAlt = color.RGBA{R: 255, G: 50, B: 255, A: 50}   // Transparent magenta
	AsteroidOrbit   = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White
)

// Player Colors
var (
	// Player body color (green circle fallback)
	PlayerBody = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Green

	// Player trail color (fading white)
	PlayerTrail = color.RGBA{R: 254, G: 254, B: 254, A: 255} // Nearly white
)

// UI and Visual Effects Colors
var (
	// Trajectory arrow color
	TrajectoryArrow = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White

	// Debug ray color
	DebugRay = color.RGBA{R: 255, G: 255, B: 0, A: 150} // Yellow with transparency
)

// Light System Colors (for shader reference)
var (
	// Light step colors for shader (RGB values normalized to 0-1)
	LightStep1 = [3]float32{254.0 / 255.0, 254.0 / 255.0, 254.0 / 255.0} // rgb(254,254,254)
	LightStep2 = [3]float32{253.0 / 255.0, 218.0 / 255.0, 145.0 / 255.0} // rgb(253,218,145)
	LightStep3 = [3]float32{254.0 / 255.0, 166.0 / 255.0, 59.0 / 255.0}  // rgb(254,166,59)
	LightStep4 = [3]float32{0.0, 0.0, 0.0}                               // transparent
	LightStep5 = [3]float32{245.0 / 255.0, 57.0 / 255.0, 68.0 / 255.0}   // rgb(245,57,68)

	// Triangle image color for shadow system
	TriangleImage = color.RGBA{R: 254, G: 254, B: 254, A: 0} // Fully transparent white
)

// Nebula Background Colors
var (
	// Main background color - darkest midnight
	NebulaBackground = color.RGBA{R: 5, G: 1, B: 20, A: 255} // #050114

	// Particle colors for the dynamic background
	RussianViolet = color.RGBA{R: 33, G: 15, B: 80, A: 255}    // #210F50
	RichIndigo    = color.RGBA{R: 71, G: 26, B: 117, A: 255}   // #471A75
	Tekhelet      = color.RGBA{R: 72, G: 28, B: 117, A: 255}   // #481C75
	DeepBlue1     = color.RGBA{R: 56, G: 2, B: 162, A: 255}    // #3802A2
	DeepBlue2     = color.RGBA{R: 37, G: 2, B: 86, A: 255}     // #250256
	DeepBlue3     = color.RGBA{R: 8, G: 2, B: 27, A: 255}      // #08021B
	BabyWhite     = color.RGBA{R: 255, G: 255, B: 253, A: 255} // #FFFFFD
)
