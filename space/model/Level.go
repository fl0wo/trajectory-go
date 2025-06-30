package Models

import "golang.org/x/image/math/f32"

type Level struct {
	Name            string
	PlayerStart     f32.Vec2
	CelestialBodies []CelestialBody
	RingAsteroids   []*RingAsteroid // Asteroids that orbit around planets
	Portals         []*Portal       // Portals for teleportation
}

func NewLevel(name string, playerStart f32.Vec2, bodies []CelestialBody) *Level {
	return &Level{
		Name:            name,
		PlayerStart:     playerStart,
		CelestialBodies: bodies,
		RingAsteroids:   []*RingAsteroid{}, // Empty by default
		Portals:         []*Portal{},       // Empty by default
	}
}

func NewLevelWithAsteroids(name string, playerStart f32.Vec2, bodies []CelestialBody, asteroids []*RingAsteroid) *Level {
	return &Level{
		Name:            name,
		PlayerStart:     playerStart,
		CelestialBodies: bodies,
		RingAsteroids:   asteroids,
		Portals:         []*Portal{}, // Empty by default
	}
}

// NewLevelWithPortals creates a new level with portals
func NewLevelWithPortals(name string, playerStart f32.Vec2, bodies []CelestialBody, portals []*Portal) *Level {
	return &Level{
		Name:            name,
		PlayerStart:     playerStart,
		CelestialBodies: bodies,
		RingAsteroids:   []*RingAsteroid{}, // Empty by default
		Portals:         portals,
	}
}

// NewLevelWithAsteroidsAndPortals creates a new level with both asteroids and portals
func NewLevelWithAsteroidsAndPortals(name string, playerStart f32.Vec2, bodies []CelestialBody, asteroids []*RingAsteroid, portals []*Portal) *Level {
	return &Level{
		Name:            name,
		PlayerStart:     playerStart,
		CelestialBodies: bodies,
		RingAsteroids:   asteroids,
		Portals:         portals,
	}
}
