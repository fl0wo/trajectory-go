package Models

import (
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/resources"
	"golang.org/x/image/math/f32"
	"math"
)

var PredefinedLevels = map[int]*Level{

	1: NewLevel("Level 1 - Simple Shot",
		f32.Vec2{0.5, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Earth",
				Position:    f32.Vec2{.5, 0.2},
				Radius:      0.08,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08),
				OrbitRadius: 0.24,
				// ImagePath:   resources.EarthImage,
			},
			&WhiteHole{
				Position:    f32.Vec2{.5, 0.8},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		}),

	2: NewLevel("Level 2 - Earth Obstacle",
		f32.Vec2{0.4, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Earth",
				Position:    f32.Vec2{constants.AspectRatio / 2, 0.5},
				Radius:      0.08,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08),
				OrbitRadius: 0.24,
				// ImagePath:   resources.EarthImage,
			},
			&WhiteHole{
				Position:    f32.Vec2{constants.AspectRatio - 0.2, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
				// ImagePath:   resources.BlackHoleImage,
			},
		}),

	3: func() *Level {
		// Create the planet first
		jupiter := &Planet{
			Name:        "Jupiter",
			Position:    f32.Vec2{constants.AspectRatio / 3, 0.5},
			Radius:      0.12,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.12 * 0.12 * 0.12 * 1),
			OrbitRadius: 0.4,
			// ImagePath:   resources.JupiterImage,
		}

		// Create asteroids that orbit around Jupiter
		asteroids := []*RingAsteroid{
			NewRingAsteroid(jupiter, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),         // Fast clockwise asteroid
			NewRingAsteroid(jupiter, 0.22, 0.012, 0.8, 0.7, math.Pi, false, resources.Asteroid2Image),  // Slower counter-clockwise asteroid
			NewRingAsteroid(jupiter, 0.26, 0.018, 1.2, 0.5, math.Pi/2, true, resources.Asteroid1Image), // Medium speed asteroid
		}

		return NewLevelWithAsteroids("Level 3 - Asteroid Ring",
			f32.Vec2{0.1, 0.8},
			[]CelestialBody{
				jupiter,
				&WhiteHole{
					Position:    f32.Vec2{constants.AspectRatio - 0.2, 0.2},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
					// ImagePath:   resources.BlackHoleImage,
				},
			},
			asteroids,
		)
	}(),

	4: NewLevel("Level 4 - Binary System",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Star A",
				Position:    f32.Vec2{constants.AspectRatio/2 - 0.2, 0.4},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 5),
				OrbitRadius: 0.25,
			},
			&Planet{
				Name:        "Star B",
				Position:    f32.Vec2{constants.AspectRatio/2 + 0.2, 0.6},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 5),
				OrbitRadius: 0.25,
			},
			&WhiteHole{
				Position:    f32.Vec2{constants.AspectRatio - 0.2, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
				// ImagePath:   resources.BlackHoleImage,
			},
		}),

	5: NewLevel("Level 5 - Black Hole Danger",
		f32.Vec2{0.1, 0.2},
		[]CelestialBody{
			&BlackHole{
				Position:    f32.Vec2{constants.AspectRatio / 2, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
				// ImagePath:   resources.BlackHoleImage,
			},
			&WhiteHole{
				Position:    f32.Vec2{constants.AspectRatio - 0.15, 0.8},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
				// ImagePath:   resources.BlackHoleImage,
			},
		}),

	6: NewLevel("Level 6 - Maze",
		f32.Vec2{0.1, 0.1},
		[]CelestialBody{
			&Planet{
				Name:        "Wall 1",
				Position:    f32.Vec2{0.4, 0.3},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
			},
			&Planet{
				Name:        "Wall 2",
				Position:    f32.Vec2{0.8, 0.7},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
			},
			&Planet{
				Name:        "Wall 3",
				Position:    f32.Vec2{1.2, 0.4},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
			},
			&WhiteHole{
				Position:    f32.Vec2{constants.AspectRatio - 0.15, 0.9},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
				// ImagePath:   resources.BlackHoleImage,
			},
		}),

	7: NewLevel("Level 7 - Cluster Challenge",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Cluster 1",
				Position:    f32.Vec2{0.5, 0.3},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 5),
				OrbitRadius: 0.12,
			},
			&Planet{
				Name:        "Cluster 2",
				Position:    f32.Vec2{0.7, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 5),
				OrbitRadius: 0.12,
			},
			&Planet{
				Name:        "Cluster 3",
				Position:    f32.Vec2{0.9, 0.7},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 15),
				OrbitRadius: 0.12,
			},
			&BlackHole{
				Position:    f32.Vec2{1.1, 0.4},
				Radius:      0.03,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.03 * 0.03 * 0.03 * 25),
				OrbitRadius: 0.2,
				// ImagePath:   resources.BlackHoleImage,
			},
			&WhiteHole{
				Position:    f32.Vec2{constants.AspectRatio - 0.15, 0.6},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
				// ImagePath:   resources.BlackHoleImage,
			},
		}),

	8: NewLevel("Level 8 - Gravity Well",
		f32.Vec2{0.1, 0.9},
		[]CelestialBody{
			&Planet{
				Name:        "Mars",
				Position:    f32.Vec2{constants.AspectRatio / 2, 0.5},
				Radius:      0.15,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.15 * 0.15 * 0.15 * 5),
				OrbitRadius: 0.5,
				// ImagePath:   resources.MarsImage,
			},
			&WhiteHole{
				Position:    f32.Vec2{constants.AspectRatio / 2, 0.2},
				Radius:      0.025,
				Mass:        0.1,
				OrbitRadius: 0.1,
				// ImagePath:   resources.BlackHoleImage,
			},
		}),

	9: NewLevel("Level 9 - Final Challenge",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Guard 1",
				Position:    f32.Vec2{0.6, 0.2},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 10),
				OrbitRadius: 0.18,
			},
			&Planet{
				Name:        "Guard 2",
				Position:    f32.Vec2{0.6, 0.8},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 10),
				OrbitRadius: 0.18,
			},
			&BlackHole{
				Position:    f32.Vec2{1.0, 0.3},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 30),
				OrbitRadius: 0.25,
				// ImagePath:   resources.BlackHoleImage,
			},
			&BlackHole{
				Position:    f32.Vec2{1.0, 0.7},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 30),
				OrbitRadius: 0.25,
				// ImagePath:   resources.BlackHoleImage,
			},
			&Planet{
				Name:        "Central Mass",
				Position:    f32.Vec2{1.4, 0.5},
				Radius:      0.08,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08 * 40),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{constants.AspectRatio - 0.15, 0.5},
				Radius:      0.025,
				Mass:        0.1,
				OrbitRadius: 0.12,
				// ImagePath:   resources.BlackHoleImage,
			},
		}),
}

func GetLevel(levelNumber int) *Level {
	if level, exists := PredefinedLevels[levelNumber]; exists {
		return level
	}
	return PredefinedLevels[1] // Default to level 1
}
