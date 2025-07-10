package Models

import (
	"image/color"
	"math"

	"github.com/you/trajectory/space/colors"
	"github.com/you/trajectory/space/resources"
	"golang.org/x/image/math/f32"
)

var colorCursor int

// nextFlatColor returns the next color from FlatColors in sequence, wrapping around if necessary.
func nextFlatColor() color.RGBA {
	c := colors.FlatColors[colorCursor%len(colors.FlatColors)]
	colorCursor++
	return c
}

var PredefinedLevels = map[int]*Level{
	// **Level 1: Whitehole only**
	// A simple introduction: direct shot to the white hole.
	1: NewLevel("Level 1 - First Step",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// **Levels 2-6: Planets + Whitehole**
	// Introduce gravitational obstacles that require curving trajectories.

	// Level 2: One planet slightly off-center, gentle curve needed.
	2: NewLevel("Level 2 - Gravity's Touch",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Earth",
				Position:    f32.Vec2{0.5, 0.55},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 2),
				OrbitRadius: 0.2,
				BaseColor:   nextFlatColor(),
				Seed:        2.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 3: Two planets creating a narrow path.
	3: NewLevel("Level 3 - Twin Barriers",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Mars",
				Position:    f32.Vec2{0.5, 0.4},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        3.0,
			},
			&Planet{
				Name:        "Venus",
				Position:    f32.Vec2{0.5, 0.6},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        4.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 4: Larger planet directly in the path, forcing a wider curve.
	4: NewLevel("Level 4 - Big Blocker",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Jupiter",
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.08,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08 * 3),
				OrbitRadius: 0.25,
				BaseColor:   nextFlatColor(),
				Seed:        5.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 5: Three planets in a triangle, tight navigation required.
	5: NewLevel("Level 5 - Triangle Puzzle",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Mercury",
				Position:    f32.Vec2{0.4, 0.4},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        6.0,
			},
			&Planet{
				Name:        "Saturn",
				Position:    f32.Vec2{0.6, 0.4},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        7.0,
			},
			&Planet{
				Name:        "Neptune",
				Position:    f32.Vec2{0.5, 0.6},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        8.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 6: Four planets forming a square, player must thread the needle.
	6: NewLevel("Level 6 - Quadrant Challenge",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Alpha",
				Position:    f32.Vec2{0.4, 0.4},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        9.0,
			},
			&Planet{
				Name:        "Beta",
				Position:    f32.Vec2{0.4, 0.6},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        10.0,
			},
			&Planet{
				Name:        "Gamma",
				Position:    f32.Vec2{0.6, 0.4},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        11.0,
			},
			&Planet{
				Name:        "Delta",
				Position:    f32.Vec2{0.6, 0.6},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        12.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// **Levels 7-8: Planets + Ring Asteroids + Whitehole**
	// Add timing challenges with moving asteroids.

	// Level 7: Planet with one asteroid ring, timing required.
	7: func() *Level {
		planet := &Planet{
			Name:        "Ringed Haven",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.07,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.07 * 0.07 * 0.07 * 3),
			OrbitRadius: 0.22,
			BaseColor:   nextFlatColor(),
			Seed:        13.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
		}
		return NewLevelWithAsteroids("Level 7 - Asteroid Dance",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				planet,
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			asteroids,
		)
	}(),

	// Level 8: Planet with two asteroid rings, increased difficulty.
	8: func() *Level {
		planet := &Planet{
			Name:        "Asteroid Fortress",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.07,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.07 * 0.07 * 0.07 * 3),
			OrbitRadius: 0.22,
			BaseColor:   nextFlatColor(),
			Seed:        14.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(planet, 0.22, 0.012, 0.8, 0.7, math.Pi, false, resources.Asteroid2Image),
		}
		return NewLevelWithAsteroids("Level 8 - Double Ring Puzzle",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				planet,
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			asteroids,
		)
	}(),

	// **Levels 9-14: Blackholes + Whiteholes**
	// Introduce strong gravitational pulls for slingshot mechanics.

	// Level 9: One black hole to curve around.
	9: NewLevel("Level 9 - Gravity Bend",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 10: Two black holes, simple slingshot path.
	10: NewLevel("Level 10 - Dual Pull",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&BlackHole{
				Position:    f32.Vec2{0.4, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.6, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 11: Three black holes in a triangle.
	11: NewLevel("Level 11 - Triple Vortex",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&BlackHole{
				Position:    f32.Vec2{0.4, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.6, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 12: One black hole near white hole, sharp curve needed.
	12: NewLevel("Level 12 - Close Encounter",
		f32.Vec2{0.25, 0.5},
		[]CelestialBody{
			&BlackHole{
				Position:    f32.Vec2{0.8, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 13: Two black holes for a slingshot effect.
	13: NewLevel("Level 13 - Slingshot Puzzle",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&BlackHole{
				Position:    f32.Vec2{0.3, 0.3},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.7, 0.7},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 14: Three black holes in a line, precision required.
	14: NewLevel("Level 14 - Gravity Gauntlet",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			//&BlackHole{
			//	Position:    f32.Vec2{0.3, 0.5},
			//	Radius:      0.04,
			//	Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
			//	OrbitRadius: 0.3,
			//},
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.7, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// **Levels 15-18: Blackholes + Whiteholes (Increased Complexity)**
	// More black holes or tighter arrangements.

	// Level 15: Four black holes in a square.
	15: NewLevel("Level 15 - Vortex Square",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&BlackHole{
				Position:    f32.Vec2{0.4, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.4, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.6, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.6, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 16: Five black holes, central obstacle.
	16: NewLevel("Level 16 - Black Hole Maze",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&BlackHole{
				Position:    f32.Vec2{0.4, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.4, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.6, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.6, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 17: Black holes in a vertical line, narrow corridor.
	17: NewLevel("Level 17 - Gravity Corridor",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.3},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.7},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 18: Black holes with varying masses for unpredictable pulls.
	18: NewLevel("Level 18 - Mass Mayhem",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&BlackHole{
				Position:    f32.Vec2{0.3, 0.3},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 25),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.7, 0.7},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 30),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// **Levels 19-20: Blackholes + Planets + Whiteholes**
	// Combine gravitational obstacles and pulls.

	// Level 19: One black hole and one planet.
	19: NewLevel("Level 19 - Mixed Forces",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Titan",
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 2),
				OrbitRadius: 0.2,
				BaseColor:   nextFlatColor(),
				Seed:        19.0,
			},
			&BlackHole{
				Position:    f32.Vec2{0.4, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 20: Two black holes and two planets.
	20: NewLevel("Level 20 - Cosmic Clash",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Io",
				Position:    f32.Vec2{0.3, 0.3},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        20.0,
			},
			&Planet{
				Name:        "Europa",
				Position:    f32.Vec2{0.7, 0.7},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        21.0,
			},
			&BlackHole{
				Position:    f32.Vec2{0.4, 0.6},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&BlackHole{
				Position:    f32.Vec2{0.6, 0.4},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// **Levels 21-24: Portals + Whiteholes**
	// Introduce teleportation mechanics.

	// Level 21: One portal pair, straightforward jump.
	21: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.5}, f32.Vec2{0.7, 0.5},
			0.0, 0.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 21 - Portal Leap",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 22: Two portal pairs, choice of path.
	22: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 0.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.2, 0.7}, f32.Vec2{0.8, 0.3},
			0.0, 0.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 22 - Portal Choice",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2, portal3, portal4},
		)
	}(),

	// Level 23: Portal with 90-degree rotation, direction matters.
	23: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.3}, f32.Vec2{0.7, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 23 - Twist Portal",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 24: Two portal pairs with rotations, complex navigation.
	24: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.2, 0.7}, f32.Vec2{0.8, 0.3},
			0.0, -90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 24 - Portal Maze",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2, portal3, portal4},
		)
	}(),

	// **Levels 25-30: Planets + Portals + Whiteholes**
	// Combine obstacles with teleportation.

	// Level 25: One planet, one portal pair.
	25: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.5}, f32.Vec2{0.8, 0.5},
			0.0, 0.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 25 - Planetary Portal",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&Planet{
					Name:        "Callisto",
					Position:    f32.Vec2{0.7, 0.5},
					Radius:      0.06,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 2),
					OrbitRadius: 0.2,
					BaseColor:   nextFlatColor(),
					Seed:        25.0,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 26: Two planets, one portal pair.
	26: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 0.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 26 - Dual Orbit Portal",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&Planet{
					Name:        "Ganymede",
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        26.0,
				},
				&Planet{
					Name:        "Enceladus",
					Position:    f32.Vec2{0.6, 0.6},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        27.0,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 27: One planet, two portal pairs.
	27: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.3}, f32.Vec2{0.7, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.3, 0.7}, f32.Vec2{0.7, 0.3},
			0.0, -90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 27 - Portal Options",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&Planet{
					Name:        "Triton",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.06,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 2),
					OrbitRadius: 0.2,
					BaseColor:   nextFlatColor(),
					Seed:        28.0,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2, portal3, portal4},
		)
	}(),

	// Level 28: Two planets blocking paths, portal required.
	28: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.5}, f32.Vec2{0.8, 0.5},
			0.0, 0.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 28 - Blocked Path",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&Planet{
					Name:        "Phobos",
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        29.0,
				},
				&Planet{
					Name:        "Deimos",
					Position:    f32.Vec2{0.6, 0.6},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        30.0,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 29: Three planets, two portal pairs.
	29: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.2, 0.7}, f32.Vec2{0.8, 0.3},
			0.0, -90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 29 - Planetary Maze",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&Planet{
					Name:        "Ceres",
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        31.0,
				},
				&Planet{
					Name:        "Pallas",
					Position:    f32.Vec2{0.6, 0.6},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        32.0,
				},
				&Planet{
					Name:        "Vesta",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        33.0,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2, portal3, portal4},
		)
	}(),

	// Level 30: Large planet with two portal pairs, complex layout.
	30: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.3}, f32.Vec2{0.7, 0.7},
			0.0, 180.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.3, 0.7}, f32.Vec2{0.7, 0.3},
			0.0, 0.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 30 - Giant Gateway",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&Planet{
					Name:        "Uranus",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.08,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08 * 3),
					OrbitRadius: 0.25,
					BaseColor:   nextFlatColor(),
					Seed:        34.0,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2, portal3, portal4},
		)
	}(),

	// **Levels 31-32: Portals + Blackholes + Whiteholes**
	// Add strong gravity to portal challenges.

	// Level 31: One black hole, one portal pair.
	31: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.5}, f32.Vec2{0.7, 0.5},
			0.0, 0.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 31 - Vortex Jump",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&BlackHole{
					Position:    f32.Vec2{0.8, 0.5},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
					OrbitRadius: 0.3,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 32: Two black holes, two portal pairs.
	32: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.2, 0.7}, f32.Vec2{0.8, 0.3},
			0.0, -90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 32 - Black Hole Portals",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&BlackHole{
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
					OrbitRadius: 0.3,
				},
				&BlackHole{
					Position:    f32.Vec2{0.6, 0.6},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
					OrbitRadius: 0.3,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2, portal3, portal4},
		)
	}(),

	// **Levels 33-34: Portals + Blackholes + Whiteholes + Planets**
	// Ultimate challenges combining all elements.

	// Level 33: One of each element.
	33: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.5}, f32.Vec2{0.8, 0.5},
			0.0, 0.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 33 - Cosmic Combo",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&Planet{
					Name:        "Pluto",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.06,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 2),
					OrbitRadius: 0.2,
					BaseColor:   nextFlatColor(),
					Seed:        35.0,
				},
				&BlackHole{
					Position:    f32.Vec2{0.7, 0.5},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
					OrbitRadius: 0.3,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 34: Two of each element, ultimate puzzle.
	34: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.2}, f32.Vec2{0.8, 0.8},
			0.0, 90.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.2, 0.8}, f32.Vec2{0.8, 0.2},
			0.0, -90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 34 - Grand Finale",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&Planet{
					Name:        "Eris",
					Position:    f32.Vec2{0.3, 0.3},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        36.0,
				},
				&Planet{
					Name:        "Haumea",
					Position:    f32.Vec2{0.7, 0.7},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        37.0,
				},
				&BlackHole{
					Position:    f32.Vec2{0.4, 0.6},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
					OrbitRadius: 0.3,
				},
				&BlackHole{
					Position:    f32.Vec2{0.6, 0.4},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
					OrbitRadius: 0.3,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2, portal3, portal4},
		)
	}(),
}

func GetLevel(levelNumber int) *Level {
	if lvl, ok := PredefinedLevels[levelNumber]; ok {
		return lvl
	}
	return PredefinedLevels[1] // Fallback to Level 1 if invalid
}

func GetNLevels() int {
	return len(PredefinedLevels)
}
