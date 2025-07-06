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
	// Level 1: Simple straight shot to white hole
	1: NewLevel("Level 1 - Welcome to Space",
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

	// Level 2: Introduce a single planet with light gravity
	2: NewLevel("Level 2 - Gentle Pull",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Earth",
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.06,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 2),
				OrbitRadius: 0.2,
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

	// Level 3: Two planets, player must curve around them
	3: NewLevel("Level 3 - Double Orbit",
		f32.Vec2{0.1, 0.3},
		[]CelestialBody{
			&Planet{
				Name:        "Mars",
				Position:    f32.Vec2{0.4, 0.4},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        2.0,
			},
			&Planet{
				Name:        "Venus",
				Position:    f32.Vec2{0.6, 0.6},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        3.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.5},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 4: Planet blocking direct path, requires slight curve
	4: NewLevel("Level 4 - Obstacle Course",
		f32.Vec2{0.1, 0.5},
		[]CelestialBody{
			&Planet{
				Name:        "Jupiter",
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.08,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08 * 3),
				OrbitRadius: 0.25,
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

	// Level 5: Three planets in a triangle, teaching precision
	5: NewLevel("Level 5 - Triangle Trap",
		f32.Vec2{0.1, 0.3},
		[]CelestialBody{
			&Planet{
				Name:        "Mercury",
				Position:    f32.Vec2{0.4, 0.3},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        5.0,
			},
			&Planet{
				Name:        "Saturn",
				Position:    f32.Vec2{0.6, 0.3},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        6.0,
			},
			&Planet{
				Name:        "Neptune",
				Position:    f32.Vec2{0.5, 0.5},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        7.0,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.3},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 6: Introduce a single portal pair, no rotation
	6: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.5}, f32.Vec2{0.7, 0.5},
			0.0, 0.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 6 - Portal Jump",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&Planet{
					Name:        "Uranus",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.06,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 2),
					OrbitRadius: 0.2,
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
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 7: Portal pair with 90-degree rotation
	7: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.3}, f32.Vec2{0.7, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 7 - Twist and Turn",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				&Planet{
					Name:        "Pluto",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        9.0,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.7},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 8: Two portal pairs, no rotation
	8: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.3}, f32.Vec2{0.7, 0.7},
			0.0, 0.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.3, 0.7}, f32.Vec2{0.7, 0.3},
			0.0, 0.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 8 - Portal Crossroads",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				&Planet{
					Name:        "Eris",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.06,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 2),
					OrbitRadius: 0.2,
					BaseColor:   nextFlatColor(),
					Seed:        10.0,
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

	// Level 9: Two portal pairs with rotations
	9: func() *Level {
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
		return NewLevelWithPortals("Level 9 - Portal Spin",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				&Planet{
					Name:        "Ceres",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.06,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 2.5),
					OrbitRadius: 0.2,
					BaseColor:   nextFlatColor(),
					Seed:        11.0,
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

	// Level 10: Portals with planets blocking paths
	10: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.4}, f32.Vec2{0.8, 0.6},
			0.0, 180.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 10 - Portal Gauntlet",
			f32.Vec2{0.1, 0.4},
			[]CelestialBody{
				&Planet{
					Name:        "Titan",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.07,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.07 * 0.07 * 0.07 * 3),
					OrbitRadius: 0.22,
					BaseColor:   nextFlatColor(),
					Seed:        12.0,
				},
				&Planet{
					Name:        "Ganymede",
					Position:    f32.Vec2{0.6, 0.4},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        13.0,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.6},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 11: Introduce a single black hole
	11: NewLevel("Level 11 - Black Hole Bend",
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

	// Level 12: Black hole with a planet
	12: NewLevel("Level 12 - Gravitational Dance",
		f32.Vec2{0.1, 0.4},
		[]CelestialBody{
			&Planet{
				Name:        "Callisto",
				Position:    f32.Vec2{0.3, 0.5},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        14.0,
			},
			&BlackHole{
				Position:    f32.Vec2{0.6, 0.5},
				Radius:      0.04,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
				OrbitRadius: 0.3,
			},
			&WhiteHole{
				Position:    f32.Vec2{0.9, 0.4},
				Radius:      0.03,
				Mass:        0.1,
				OrbitRadius: 0.15,
			},
		},
	),

	// Level 13: Two black holes, simple path
	13: NewLevel("Level 13 - Dual Vortex",
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

	// Level 14: Black hole with portal pair
	14: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.3}, f32.Vec2{0.7, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 14 - Black Hole Portal",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				&BlackHole{
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
					OrbitRadius: 0.3,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.7},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 15: Black hole and two planets
	15: NewLevel("Level 15 - Triple Threat",
		f32.Vec2{0.1, 0.4},
		[]CelestialBody{
			&Planet{
				Name:        "Io",
				Position:    f32.Vec2{0.3, 0.4},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2.5),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        15.0,
			},
			&Planet{
				Name:        "Europa",
				Position:    f32.Vec2{0.4, 0.6},
				Radius:      0.05,
				Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2.5),
				OrbitRadius: 0.18,
				BaseColor:   nextFlatColor(),
				Seed:        16.0,
			},
			&BlackHole{
				Position:    f32.Vec2{0.6, 0.5},
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

	// Level 16: Introduce asteroids around a planet
	16: func() *Level {
		planet := &Planet{
			Name:        "Asteroid Hub",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.07,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.07 * 0.07 * 0.07 * 3),
			OrbitRadius: 0.22,
			BaseColor:   nextFlatColor(),
			Seed:        17.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(planet, 0.22, 0.012, 0.8, 0.7, math.Pi, false, resources.Asteroid2Image),
		}
		return NewLevelWithAsteroids("Level 16 - Asteroid Ring",
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

	// Level 17: Asteroids with a portal
	17: func() *Level {
		planet := &Planet{
			Name:        "Asteroid Base",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.07,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.07 * 0.07 * 0.07 * 3),
			OrbitRadius: 0.22,
			BaseColor:   nextFlatColor(),
			Seed:        18.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
		}
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.4}, f32.Vec2{0.7, 0.6},
			0.0, 90.0,
			0.025, 0.025,
		)
		return NewLevelWithAsteroidsAndPortals("Level 17 - Asteroid Portal",
			f32.Vec2{0.1, 0.4},
			[]CelestialBody{
				planet,
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.6},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			asteroids,
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 18: Dense asteroid field
	18: func() *Level {
		planet := &Planet{
			Name:        "Asteroid Core",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.08,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08 * 3),
			OrbitRadius: 0.25,
			BaseColor:   nextFlatColor(),
			Seed:        19.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(planet, 0.22, 0.012, 0.8, 0.7, math.Pi/2, false, resources.Asteroid2Image),
			NewRingAsteroid(planet, 0.26, 0.018, 1.2, 0.5, math.Pi, true, resources.Asteroid1Image),
		}
		return NewLevelWithAsteroids("Level 18 - Asteroid Maze",
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

	// Level 19: Asteroids and two planets
	19: func() *Level {
		planet := &Planet{
			Name:        "Asteroid Anchor",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.07,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.07 * 0.07 * 0.07 * 3),
			OrbitRadius: 0.22,
			BaseColor:   nextFlatColor(),
			Seed:        20.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(planet, 0.22, 0.012, 0.8, 0.7, math.Pi, false, resources.Asteroid2Image),
		}
		return NewLevelWithAsteroids("Level 19 - Double Anchor",
			f32.Vec2{0.1, 0.4},
			[]CelestialBody{
				planet,
				&Planet{
					Name:        "Companion",
					Position:    f32.Vec2{0.6, 0.4},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        21.0,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.4},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			asteroids,
		)
	}(),

	// Level 20: Asteroids, portal, and planet
	20: func() *Level {
		planet := &Planet{
			Name:        "Asteroid Gate",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.07,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.07 * 0.07 * 0.07 * 3),
			OrbitRadius: 0.22,
			BaseColor:   nextFlatColor(),
			Seed:        22.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
		}
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.3}, f32.Vec2{0.7, 0.7},
			0.0, 180.0,
			0.025, 0.025,
		)
		return NewLevelWithAsteroidsAndPortals("Level 20 - Cosmic Shortcut",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				planet,
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.7},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			asteroids,
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 21: Black hole and portal combination
	21: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.4}, f32.Vec2{0.8, 0.6},
			0.0, 90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 21 - Vortex Portal",
			f32.Vec2{0.1, 0.4},
			[]CelestialBody{
				&BlackHole{
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
					OrbitRadius: 0.3,
				},
				&Planet{
					Name:        "Protector",
					Position:    f32.Vec2{0.6, 0.6},
					Radius:      0.05,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.05 * 0.05 * 0.05 * 2.5),
					OrbitRadius: 0.18,
					BaseColor:   nextFlatColor(),
					Seed:        23.0,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.6},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 22: Two black holes and a portal
	22: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.3, 0.3}, f32.Vec2{0.7, 0.7},
			0.0, 180.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 22 - Dual Vortex Portal",
			f32.Vec2{0.1, 0.3},
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
					Position:    f32.Vec2{0.9, 0.7},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 23: Black holes, portals, and a planet
	23: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 23 - Gravitational Maze",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				&Planet{
					Name:        "Defender",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.06,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 3),
					OrbitRadius: 0.2,
					BaseColor:   nextFlatColor(),
					Seed:        24.0,
				},
				&BlackHole{
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
					OrbitRadius: 0.3,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.7},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 24: Multiple portals and black holes
	24: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.3, 0.7}, f32.Vec2{0.7, 0.3},
			0.0, -90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 24 - Portal Vortex",
			f32.Vec2{0.1, 0.3},
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

	// Level 25: Black holes, portals, and asteroids
	25: func() *Level {
		planet := &Planet{
			Name:        "Chaos Core",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.07,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.07 * 0.07 * 0.07 * 3),
			OrbitRadius: 0.22,
			BaseColor:   nextFlatColor(),
			Seed:        25.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(planet, 0.22, 0.012, 0.8, 0.7, math.Pi, false, resources.Asteroid2Image),
		}
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.4}, f32.Vec2{0.8, 0.6},
			0.0, 90.0,
			0.025, 0.025,
		)
		return NewLevelWithAsteroidsAndPortals("Level 25 - Cosmic Chaos",
			f32.Vec2{0.1, 0.4},
			[]CelestialBody{
				planet,
				&BlackHole{
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 20),
					OrbitRadius: 0.3,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.6},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			asteroids,
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 26: Dense black hole and portal challenge
	26: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 180.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.3, 0.7}, f32.Vec2{0.7, 0.3},
			0.0, -90.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 26 - Vortex Maze",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				&BlackHole{
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 25),
					OrbitRadius: 0.3,
				},
				&BlackHole{
					Position:    f32.Vec2{0.6, 0.6},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 25),
					OrbitRadius: 0.3,
				},
				&Planet{
					Name:        "Blocker",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.06,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 3),
					OrbitRadius: 0.2,
					BaseColor:   nextFlatColor(),
					Seed:        26.0,
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

	// Level 27: All mechanics combined
	27: func() *Level {
		planet := &Planet{
			Name:        "Chaos Center",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.08,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08 * 3),
			OrbitRadius: 0.25,
			BaseColor:   nextFlatColor(),
			Seed:        27.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(planet, 0.22, 0.012, 0.8, 0.7, math.Pi/2, false, resources.Asteroid2Image),
			NewRingAsteroid(planet, 0.26, 0.018, 1.2, 0.5, math.Pi, true, resources.Asteroid1Image),
		}
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		return NewLevelWithAsteroidsAndPortals("Level 27 - Stellar Storm",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				planet,
				&BlackHole{
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 25),
					OrbitRadius: 0.3,
				},
				&BlackHole{
					Position:    f32.Vec2{0.6, 0.6},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 25),
					OrbitRadius: 0.3,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.7},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			asteroids,
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 28: Tight asteroid field with portals
	28: func() *Level {
		planet := &Planet{
			Name:        "Asteroid Fortress",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.08,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08 * 3),
			OrbitRadius: 0.25,
			BaseColor:   nextFlatColor(),
			Seed:        28.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.15, 0.015, 1.2, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(planet, 0.18, 0.012, 1.0, 0.8, math.Pi/4, false, resources.Asteroid2Image),
			NewRingAsteroid(planet, 0.22, 0.018, 0.8, 0.6, math.Pi/2, true, resources.Asteroid1Image),
		}
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.4}, f32.Vec2{0.8, 0.6},
			0.0, 180.0,
			0.025, 0.025,
		)
		return NewLevelWithAsteroidsAndPortals("Level 28 - Asteroid Breach",
			f32.Vec2{0.1, 0.4},
			[]CelestialBody{
				planet,
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.6},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			asteroids,
			[]*Portal{portal1, portal2},
		)
	}(),

	// Level 29: Multiple black holes and asteroids
	29: func() *Level {
		planet := &Planet{
			Name:        "Dark Cluster",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.07,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.07 * 0.07 * 0.07 * 3),
			OrbitRadius: 0.22,
			BaseColor:   nextFlatColor(),
			Seed:        29.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.18, 0.015, 1.0, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(planet, 0.22, 0.012, 0.8, 0.7, math.Pi, false, resources.Asteroid2Image),
		}
		return NewLevelWithAsteroids("Level 29 - Black Hole Belt",
			f32.Vec2{0.1, 0.5},
			[]CelestialBody{
				planet,
				&BlackHole{
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 25),
					OrbitRadius: 0.3,
				},
				&BlackHole{
					Position:    f32.Vec2{0.6, 0.6},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 25),
					OrbitRadius: 0.3,
				},
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

	// Level 30: Complex portal network
	30: func() *Level {
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.3, 0.7}, f32.Vec2{0.7, 0.3},
			0.0, -90.0,
			0.025, 0.025,
		)
		portal5, portal6 := NewPortalPair(3,
			f32.Vec2{0.4, 0.4}, f32.Vec2{0.6, 0.6},
			0.0, 180.0,
			0.025, 0.025,
		)
		return NewLevelWithPortals("Level 30 - Portal Network",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				&Planet{
					Name:        "Nexus",
					Position:    f32.Vec2{0.5, 0.5},
					Radius:      0.06,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 3),
					OrbitRadius: 0.2,
					BaseColor:   nextFlatColor(),
					Seed:        30.0,
				},
				&BlackHole{
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 25),
					OrbitRadius: 0.3,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			[]*Portal{portal1, portal2, portal3, portal4, portal5, portal6},
		)
	}(),

	// Level 31: High difficulty with all mechanics
	31: func() *Level {
		planet := &Planet{
			Name:        "Final Bastion",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.08,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08 * 3.5),
			OrbitRadius: 0.25,
			BaseColor:   nextFlatColor(),
			Seed:        31.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet, 0.15, 0.015, 1.2, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(planet, 0.18, 0.012, 1.0, 0.8, math.Pi/4, false, resources.Asteroid2Image),
			NewRingAsteroid(planet, 0.22, 0.018, 0.8, 0.6, math.Pi/2, true, resources.Asteroid1Image),
		}
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.3, 0.7}, f32.Vec2{0.7, 0.3},
			0.0, -90.0,
			0.025, 0.025,
		)
		return NewLevelWithAsteroidsAndPortals("Level 31 - Cosmic Gauntlet",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				planet,
				&BlackHole{
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 25),
					OrbitRadius: 0.3,
				},
				&BlackHole{
					Position:    f32.Vec2{0.6, 0.6},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 25),
					OrbitRadius: 0.3,
				},
				&WhiteHole{
					Position:    f32.Vec2{0.9, 0.5},
					Radius:      0.03,
					Mass:        0.1,
					OrbitRadius: 0.15,
				},
			},
			asteroids,
			[]*Portal{portal1, portal2, portal3, portal4},
		)
	}(),

	// Level 32: Ultimate challenge with all mechanics
	32: func() *Level {
		planet1 := &Planet{
			Name:        "Final Core",
			Position:    f32.Vec2{0.5, 0.5},
			Radius:      0.08,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.08 * 0.08 * 0.08 * 4),
			OrbitRadius: 0.25,
			BaseColor:   nextFlatColor(),
			Seed:        32.0,
		}
		planet2 := &Planet{
			Name:        "Guardian",
			Position:    f32.Vec2{0.6, 0.4},
			Radius:      0.06,
			Mass:        float32((4.0 / 3.0) * math.Pi * 0.06 * 0.06 * 0.06 * 3),
			OrbitRadius: 0.2,
			BaseColor:   nextFlatColor(),
			Seed:        33.0,
		}
		asteroids := []*RingAsteroid{
			NewRingAsteroid(planet1, 0.15, 0.015, 1.2, 1.0, 0, true, resources.Asteroid1Image),
			NewRingAsteroid(planet1, 0.18, 0.012, 1.0, 0.8, math.Pi/4, false, resources.Asteroid2Image),
			NewRingAsteroid(planet1, 0.22, 0.018, 0.8, 0.6, math.Pi/2, true, resources.Asteroid1Image),
			NewRingAsteroid(planet1, 0.26, 0.015, 1.0, 0.7, math.Pi, false, resources.Asteroid2Image),
		}
		portal1, portal2 := NewPortalPair(1,
			f32.Vec2{0.2, 0.3}, f32.Vec2{0.8, 0.7},
			0.0, 90.0,
			0.025, 0.025,
		)
		portal3, portal4 := NewPortalPair(2,
			f32.Vec2{0.3, 0.7}, f32.Vec2{0.7, 0.3},
			0.0, -90.0,
			0.025, 0.025,
		)
		portal5, portal6 := NewPortalPair(3,
			f32.Vec2{0.4, 0.4}, f32.Vec2{0.6, 0.6},
			0.0, 180.0,
			0.025, 0.025,
		)
		return NewLevelWithAsteroidsAndPortals("Level 32 - Galactic Finale",
			f32.Vec2{0.1, 0.3},
			[]CelestialBody{
				planet1,
				planet2,
				&BlackHole{
					Position:    f32.Vec2{0.4, 0.4},
					Radius:      0.04,
					Mass:        float32((4.0 / 3.0) * math.Pi * 0.04 * 0.04 * 0.04 * 30),
					OrbitRadius: 0.3,
				},
				&BlackHole{
					Position:    f32.Vec2{0.6, 0.6},
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
			asteroids,
			[]*Portal{portal1, portal2, portal3, portal4, portal5, portal6},
		)
	}(),
}

func GetLevel(levelNumber int) *Level {
	if lvl, ok := PredefinedLevels[levelNumber]; ok {
		return lvl
	}
	return PredefinedLevels[1] // Fallback to Level 1 if invalid
}
