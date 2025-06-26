package Models

import "golang.org/x/image/math/f32"

type Level struct {
	Name            string
	PlayerStart     f32.Vec2
	CelestialBodies []CelestialBody
}

func NewLevel(name string, playerStart f32.Vec2, bodies []CelestialBody) *Level {
	return &Level{
		Name:            name,
		PlayerStart:     playerStart,
		CelestialBodies: bodies,
	}
}
