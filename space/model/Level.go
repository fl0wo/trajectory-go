package Models

import "golang.org/x/image/math/f32"

type Level struct {
	Name            string
	PlayerStart     f32.Vec2
	CelestialBodies []CelestialBody
	RingAsteroids   []*RingAsteroid // Asteroids that orbit around planets
}

func NewLevel(name string, playerStart f32.Vec2, bodies []CelestialBody) *Level {
	return &Level{
		Name:            name,
		PlayerStart:     playerStart,
		CelestialBodies: bodies,
		RingAsteroids:   []*RingAsteroid{}, // Empty by default
	}
}

func NewLevelWithAsteroids(name string, playerStart f32.Vec2, bodies []CelestialBody, asteroids []*RingAsteroid) *Level {
	return &Level{
		Name:            name,
		PlayerStart:     playerStart,
		CelestialBodies: bodies,
		RingAsteroids:   asteroids,
	}
}
