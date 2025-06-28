package Models

import (
	"math"

	"github.com/you/trajectory/space/resources"
	"golang.org/x/image/math/f32"
)

var PredefinedLevels = map[int]*Level{

	1: NewLevel("Level 1 - First Steps",
		// spawn near left edge
		f32.Vec2{0.2, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Earth",
				Position:    f32.Vec2{0.6, 0.5},
				Radius:      0.08,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08),
				OrbitRadius: 0.25,
			},
			&WhiteHole{ // win-point
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
				// ImagePath:   resources.BlackHoleImage,
			},
		},
	),

	2: NewLevel("Level 2 - White Hole Detour",
		// spawn low left
		f32.Vec2{0.2, 0.3},
		[]CelestialBody{
			&Planet{
				Name:        "Mars",
				Position:    f32.Vec2{0.5, 0.6},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06),
				OrbitRadius: 0.20,
			},
			&WhiteHole{ // win-point
				Position:    f32.Vec2{0.8, 0.4},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	3: func() *Level {
		jupiter := &Planet{
			Name:        "Jupiter",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.12,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.12 * 0.12 * 0.12 * 1.2),
			OrbitRadius: 0.40,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(jupiter, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(jupiter, 0.22, 0.012, 0.8, 0.7, math.Pi, false, resources.Asteroid2Image),
			NewRingAsteroid(jupiter, 0.26, 0.018, 1.2, 0.5, math.Pi/2, true, resources.Asteroid1Image),
		}
		return NewLevelWithAsteroids("Level 3 - Asteroid Belt",
			// spawn above left
			f32.Vec2{0.1, 0.8},
			[]CelestialBody{
				jupiter,
				&WhiteHole{ // win-point
					Position:    f32.Vec2{0.85, 0.2},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			asteroids,
		)
	}(),

	4: NewLevel("Level 4 - Binary Stars",
		// spawn mid-left
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Star A",
				Position:    f32.Vec2{0.45, 0.55},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 5),
				OrbitRadius: 0.25,
			},
			&Planet{
				Name:        "Star B",
				Position:    f32.Vec2{0.55, 0.45},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 5),
				OrbitRadius: 0.25,
			},
			&WhiteHole{ // win-point
				Position:    f32.Vec2{0.85, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.20,
			},
		},
	),

	5: NewLevel("Level 5 - Black Hole Slingshot",
		// spawn left-center
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Moon",
				Position:    f32.Vec2{0.3, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 0.5),
				OrbitRadius: 0.15,
			},
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.30,
			},
			&WhiteHole{ // win-point
				Position:    f32.Vec2{0.8, 0.2},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	6: NewLevel("Level 6 - Asteroid Maze",
		// spawn bottom-left
		f32.Vec2{0.1, 0.1},
		[]CelestialBody{
			&Planet{
				Name:        "Wall A",
				Position:    f32.Vec2{0.3, 0.3},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
			},
			&Planet{
				Name:        "Wall B",
				Position:    f32.Vec2{0.6, 0.25},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
			},
			&Planet{
				Name:        "Wall C",
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
			},
			&Planet{
				Name:        "Wall D",
				Position:    f32.Vec2{0.8, 0.7},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
			},
			&WhiteHole{ // win-point
				Position:    f32.Vec2{0.9, 0.9},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	7: NewLevel("Level 7 - Cluster Challenge",
		// spawn mid-left
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Cluster 1",
				Position:    f32.Vec2{0.4, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 5),
				OrbitRadius: 0.12,
			},
			&Planet{
				Name:        "Cluster 2",
				Position:    f32.Vec2{0.5, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 5),
				OrbitRadius: 0.12,
			},
			&Planet{
				Name:        "Cluster 3",
				Position:    f32.Vec2{0.6, 0.3},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 5),
				OrbitRadius: 0.12,
			},
			&BlackHole{
				Position:    f32.Vec2{0.8, 0.5},
				Radius:      0.03,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.03 * 0.03 * 0.03 * 25),
				OrbitRadius: 0.20,
			},
			&WhiteHole{ // win-point
				Position:    f32.Vec2{0.9, 0.7},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	8: NewLevel("Level 8 - Deep Gravity Well",
		// spawn top-left
		f32.Vec2{0.1, 0.9},
		[]CelestialBody{
			&Planet{
				Name:        "Mars",
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.15,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.15 * 0.15 * 0.15 * 5),
				OrbitRadius: 0.45,
			},
			&WhiteHole{ // win-point
				Position:    f32.Vec2{0.5, 0.25},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.10,
			},
		},
	),

	9: NewLevel("Level 9 - Grand Finale",
		// spawn mid-left
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Guard 1",
				Position:    f32.Vec2{0.6, 0.3},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 10),
				OrbitRadius: 0.18,
			},
			&Planet{
				Name:        "Guard 2",
				Position:    f32.Vec2{0.6, 0.7},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 10),
				OrbitRadius: 0.18,
			},
			&BlackHole{
				Position:    f32.Vec2{0.8, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 30),
				OrbitRadius: 0.25,
			},
			&BlackHole{
				Position:    f32.Vec2{0.8, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 30),
				OrbitRadius: 0.25,
			},
			&Planet{
				Name:        "Core",
				Position:    f32.Vec2{1.3, 0.5},
				Radius:      0.10,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.10 * 0.10 * 0.10 * 40),
				OrbitRadius: 0.30,
			},
			&WhiteHole{ // win-point
				Position:    f32.Vec2{1.45, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.12,
			},
		},
	),
}

func GetLevel(levelNumber int) *Level {
	if lvl, ok := PredefinedLevels[levelNumber]; ok {
		return lvl
	}
	return PredefinedLevels[1]
}
