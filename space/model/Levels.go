package Models

import (
	"image/color"
	"math"

	"github.com/you/trajectory/space/resources"
	"golang.org/x/image/math/f32"
)

// FlatColors is a reusable palette for all planets and stars,
// updated to your provided base colors.
var FlatColors = []color.RGBA{
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

var colorCursor int

// nextFlatColor returns the next color from FlatColors in sequence,
// wrapping around if necessary.
func nextFlatColor() color.RGBA {
	c := FlatColors[colorCursor%len(FlatColors)]
	colorCursor++
	return c
}

var PredefinedLevels = map[int]*Level{

	1: NewLevel("Level 1 - First Steps",
		f32.Vec2{0.2, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Earth",
				Position:    f32.Vec2{0.6, 0.5},
				Radius:      0.08,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08),
				OrbitRadius: 0.25,
				BaseColor:   nextFlatColor(),
				Seed:        1.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	2: NewLevel("Level 2 - White Hole Detour",
		f32.Vec2{0.2, 0.3},
		[]CelestialBody{
			&Planet{
				Name:        "Mars",
				Position:    f32.Vec2{0.5, 0.6},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06),
				OrbitRadius: 0.20,
				BaseColor:   nextFlatColor(),
				Seed:        2.0,
			},
			&WhiteHole{
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
			BaseColor:   nextFlatColor(),
			Seed:        3.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(jupiter, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(jupiter, 0.22, 0.012, 0.8, 0.7, math.Pi, false, resources.Asteroid2Image),
			NewRingAsteroid(jupiter, 0.26, 0.018, 1.2, 0.5, math.Pi/2, true, resources.Asteroid1Image),
		}
		return NewLevelWithAsteroids("Level 3 - Asteroid Belt",
			f32.Vec2{0.1, 0.8},
			[]CelestialBody{
				jupiter,
				&WhiteHole{
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
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Star A",
				Position:    f32.Vec2{0.45, 0.55},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 5),
				OrbitRadius: 0.25,
				BaseColor:   nextFlatColor(),
				Seed:        4.0,
			},
			&Planet{
				Name:        "Star B",
				Position:    f32.Vec2{0.55, 0.45},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 5),
				OrbitRadius: 0.25,
				BaseColor:   nextFlatColor(),
				Seed:        5.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.85, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.20,
			},
		},
	),

	5: NewLevel("Level 5 - Black Hole Slingshot",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Moon",
				Position:    f32.Vec2{0.3, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 0.5),
				OrbitRadius: 0.15,
				BaseColor:   nextFlatColor(),
				Seed:        6.0,
			},
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.30,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.8, 0.2},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	6: NewLevel("Level 6 - Asteroid Maze",
		f32.Vec2{0.1, 0.1},
		[]CelestialBody{
			&Planet{
				Name:        "Wall A",
				Position:    f32.Vec2{0.3, 0.3},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
				BaseColor:   nextFlatColor(),
				Seed:        7.0,
			},
			&Planet{
				Name:        "Wall B",
				Position:    f32.Vec2{0.6, 0.25},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
				BaseColor:   nextFlatColor(),
				Seed:        8.0,
			},
			&Planet{
				Name:        "Wall C",
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
				BaseColor:   nextFlatColor(),
				Seed:        9.0,
			},
			&Planet{
				Name:        "Wall D",
				Position:    f32.Vec2{0.8, 0.7},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 5),
				OrbitRadius: 0.15,
				BaseColor:   nextFlatColor(),
				Seed:        10.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.9},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	7: NewLevel("Level 7 - Cluster Challenge",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Cluster 1",
				Position:    f32.Vec2{0.4, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 5),
				OrbitRadius: 0.12,
				BaseColor:   nextFlatColor(),
				Seed:        11.0,
			},
			&Planet{
				Name:        "Cluster 2",
				Position:    f32.Vec2{0.5, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 5),
				OrbitRadius: 0.12,
				BaseColor:   nextFlatColor(),
				Seed:        12.0,
			},
			&Planet{
				Name:        "Cluster 3",
				Position:    f32.Vec2{0.6, 0.3},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 5),
				OrbitRadius: 0.12,
				BaseColor:   nextFlatColor(),
				Seed:        13.0,
			},
			&BlackHole{
				Position:    f32.Vec2{0.8, 0.5},
				Radius:      0.03,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.03 * 0.03 * 0.03 * 25),
				OrbitRadius: 0.20,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.7},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	8: NewLevel("Level 8 - Deep Gravity Well",
		f32.Vec2{0.1, 0.9},
		[]CelestialBody{
			&Planet{
				Name:        "Mars",
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.15,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.15 * 0.15 * 0.15 * 5),
				OrbitRadius: 0.45,
				BaseColor:   nextFlatColor(),
				Seed:        14.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.5, 0.25},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.10,
			},
		},
	),

	9: NewLevel("Level 9 - Grand Finale",
		f32.Vec2{0.2, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Guard 1",
				Position:    f32.Vec2{0.6, 0.3},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 10),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        15.0,
			},
			&Planet{
				Name:        "Guard 2",
				Position:    f32.Vec2{0.6, 0.7},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 10),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        16.0,
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
				BaseColor:   nextFlatColor(),
				Seed:        17.0,
			},
			&WhiteHole{
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
