package colors

import (
	"github.com/you/trajectory/space/util"
	"image/color"
)

// Game Background Colors
var (
	// Background Primary background color for space (#060219)
	Background = color.RGBA{R: 6, G: 2, B: 25, A: 255}

	// ShadowBackground Shadow system background
	ShadowBackground = color.RGBA{R: 6, G: 2, B: 25, A: 255}

	FlatColors = []color.RGBA{
		{R: 255, G: 107, B: 129, A: 255}, // rgba(255,107,129,1.0)
		{R: 83, G: 82, B: 237, A: 255},   // rgba(83,82,237,1.0)
		{R: 55, G: 66, B: 250, A: 255},   // rgba(55,66,250,1.0)
		{R: 30, G: 144, B: 255, A: 255},  // rgba(30,144,255,1.0)
		{R: 123, G: 237, B: 159, A: 255}, // rgba(123,237,159,1.0)
		{R: 46, G: 213, B: 115, A: 255},  // rgba(46,213,115,1.0)
		{R: 255, G: 165, B: 2, A: 255},   // rgba(255,165,2,1.0)
		{R: 236, G: 204, B: 104, A: 255}, // rgba(236,204,104,1.0)
		{R: 179, G: 55, B: 113, A: 255},  // rgba(179,55,113,1.0)
		{R: 154, G: 236, B: 219, A: 255}, // rgba(154,236,219,1.0)
		{R: 109, G: 33, B: 79, A: 255},   // rgba(109,33,79,1.0)
		{R: 85, G: 230, B: 193, A: 255},  // rgba(85,230,193,1.0)
	}
)

// Celestial Body Colors
var (
	// PlanetBody Planet colors
	PlanetBody  = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White
	PlanetOrbit = util.RGBAColor(119, 140, 163, 255.0)

	// BlackHoleBody Black hole colors
	BlackHoleBody  = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Transparent red
	BlackHoleOrbit = util.RGBAColor(0, 0, 0, 0)             // Black

	// WhiteHoleBody White hole colors
	WhiteHoleBody  = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White
	WhiteHoleOrbit = util.RGBAColor(119, 140, 163, 255.0)

	// AsteroidBody Asteroid colors
	AsteroidBody    = util.RGBAColor(255, 255, 255, 100)
	AsteroidBodyAlt = color.RGBA{R: 255, G: 50, B: 255, A: 50} // Transparent magenta
	AsteroidOrbit   = util.RGBAColor(87.0, 96.0, 111.0, 255.0)

	BorderIndicator = util.RGBAColor(255, 255, 255, 255) // Border indicator color (purple-ish)

	TransparentColor        = color.RGBA{R: 0, G: 0, B: 0, A: 0}     // Fully transparent color
	LastCollisionCrossColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red cross for last collision

	PortalColors = FlatColors

	PortalColorInnerRing = NebulaBackground
)

// Player Colors
var (
	// PlayerBody Player body color (green circle fallback)
	PlayerBody = util.RGBAColor(123, 237, 159, 255) // Green

	// PlayerTrail Player trail color (fading white)
	PlayerTrail = color.RGBA{R: 254, G: 254, B: 254, A: 255} // Nearly white
)

// UI and Visual Effects Colors
var (
	// TrajectoryArrow Trajectory arrow color
	TrajectoryArrow = color.RGBA{R: 255, G: 255, B: 255, A: 255} // White

	// DebugRay Debug ray color
	DebugRay = color.RGBA{R: 255, G: 255, B: 0, A: 150} // Yellow with transparency
)

// Light System Colors (for shader reference)
var (
	// TriangleImage Triangle image color for shadow system
	TriangleImage = color.RGBA{R: 254, G: 254, B: 254, A: 0} // Fully transparent white
)

// Nebula Background Colors
var (
	// NebulaBackground Main background color - darkest midnight
	NebulaBackground = color.RGBA{R: 6, G: 2, B: 25, A: 255} // #050114

	// RussianViolet Particle colors for the dynamic background
	RussianViolet = color.RGBA{R: 33, G: 15, B: 80, A: 255}    // #210F50
	RichIndigo    = color.RGBA{R: 71, G: 26, B: 117, A: 255}   // #471A75
	Tekhelet      = color.RGBA{R: 72, G: 28, B: 117, A: 255}   // #481C75
	DeepBlue1     = color.RGBA{R: 56, G: 2, B: 162, A: 255}    // #3802A2
	DeepBlue2     = color.RGBA{R: 37, G: 2, B: 86, A: 255}     // #250256
	DeepBlue3     = color.RGBA{R: 8, G: 2, B: 27, A: 255}      // #08021B
	BabyWhite     = color.RGBA{R: 255, G: 255, B: 253, A: 255} // #FFFFFD
)
