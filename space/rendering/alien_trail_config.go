package rendering

import "image/color"

// AlienTrailConfig holds configuration for the alien trail effect
type AlienTrailConfig struct {
	DropCount      int
	TrailLength    float32
	SmearWidth     float32
	NoiseIntensity float32
	FadeStart      float32

	DropRate     float32    // Drops generated per second
	DropSizeMin  float32    // Minimum drop size (as fraction of player radius)
	DropSizeMax  float32    // Maximum drop size (as fraction of player radius)
	ColorInner   color.RGBA // Inner color of the effect
	ColorOuter   color.RGBA // Outer color of the effect
	NoiseScale   float32    // Scale of noise texture
	NoiseSpeed   float32    // Speed of noise animation
	Acceleration [2]float32 // Custom acceleration (e.g., gravity direction)
}

// Predefined trail configurations
var (
	// Green mucus alien trail (default)
	AlienTrailMucus = AlienTrailConfig{
		DropCount:      8,
		TrailLength:    2.0,
		SmearWidth:     0.5,
		NoiseIntensity: 0.5,
		FadeStart:      0.2,
		DropRate:       2.0,
		DropSizeMin:    0.2,
		DropSizeMax:    0.5,
		ColorInner:     color.RGBA{R: 102, G: 204, B: 51, A: 230}, // Bright green
		ColorOuter:     color.RGBA{R: 51, G: 153, B: 25, A: 179},  // Darker green
		NoiseScale:     10.0,
		NoiseSpeed:     0.5,
		Acceleration:   [2]float32{0.0, 0.3}, // Downward gravity
	}

	// Red fire trail
	AlienTrailFire = AlienTrailConfig{
		DropCount:      12,
		TrailLength:    1.5,
		SmearWidth:     0.3,
		NoiseIntensity: 0.8,
		FadeStart:      0.1,
		DropRate:       4.0,
		DropSizeMin:    0.15,
		DropSizeMax:    0.4,
		ColorInner:     color.RGBA{R: 255, G: 100, B: 0, A: 255}, // Bright orange
		ColorOuter:     color.RGBA{R: 200, G: 50, B: 0, A: 150},  // Dark red
		NoiseScale:     15.0,
		NoiseSpeed:     1.0,
		Acceleration:   [2]float32{0.0, -0.1}, // Upward (fire rises)
	}

	// Blue energy trail
	AlienTrailEnergy = AlienTrailConfig{
		DropCount:      6,
		TrailLength:    3.0,
		SmearWidth:     0.2,
		NoiseIntensity: 1.0,
		FadeStart:      0.3,
		DropRate:       3.0,
		DropSizeMin:    0.1,
		DropSizeMax:    0.3,
		ColorInner:     color.RGBA{R: 100, G: 200, B: 255, A: 200}, // Bright blue
		ColorOuter:     color.RGBA{R: 50, G: 100, B: 200, A: 100},  // Darker blue
		NoiseScale:     20.0,
		NoiseSpeed:     2.0,
		Acceleration:   [2]float32{0.0, 0.0}, // No gravity (floats)
	}

	// Purple poison trail
	AlienTrailPoison = AlienTrailConfig{
		DropCount:      4,
		TrailLength:    2.5,
		SmearWidth:     0.7,
		NoiseIntensity: 0.3,
		FadeStart:      0.4,
		DropRate:       1.5,
		DropSizeMin:    0.3,
		DropSizeMax:    0.7,
		ColorInner:     color.RGBA{R: 150, G: 50, B: 200, A: 220}, // Bright purple
		ColorOuter:     color.RGBA{R: 100, G: 25, B: 150, A: 120}, // Dark purple
		NoiseScale:     8.0,
		NoiseSpeed:     0.3,
		Acceleration:   [2]float32{0.0, 0.5}, // Heavy drops
	}
)

// ApplyTrailConfig applies a trail configuration to the shader uniforms
func (r *Renderer) ApplyTrailConfig(uniforms map[string]any, config AlienTrailConfig, playerRadius float32) {
	uniforms["DropRate"] = config.DropRate
	uniforms["DropSizeMin"] = playerRadius * config.DropSizeMin
	uniforms["DropSizeMax"] = playerRadius * config.DropSizeMax
	uniforms["ColorInner"] = []float32{
		float32(config.ColorInner.R) / 255.0,
		float32(config.ColorInner.G) / 255.0,
		float32(config.ColorInner.B) / 255.0,
		float32(config.ColorInner.A) / 255.0,
	}
	uniforms["ColorOuter"] = []float32{
		float32(config.ColorOuter.R) / 255.0,
		float32(config.ColorOuter.G) / 255.0,
		float32(config.ColorOuter.B) / 255.0,
		float32(config.ColorOuter.A) / 255.0,
	}
	uniforms["NoiseScale"] = config.NoiseScale
	uniforms["NoiseSpeed"] = config.NoiseSpeed
	uniforms["Acceleration"] = []float32{config.Acceleration[0], config.Acceleration[1]}

	uniforms["DropCount"] = config.DropCount
	uniforms["TrailLength"] = config.TrailLength
	uniforms["SmearWidth"] = config.SmearWidth
	uniforms["NoiseIntensity"] = config.NoiseIntensity
	uniforms["FadeStart"] = config.FadeStart
}
