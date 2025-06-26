package Models

import "golang.org/x/image/math/f32"

type CelestialBodyType int

const (
	CelestialBodyTypePlanet CelestialBodyType = iota
	CelestialBodyTypeBlackHole
	CelestialBodyTypeWhiteHole
	CelestialBodyTypeAsteroid
)

type CelestialBody interface {
	GetPosition() f32.Vec2
	GetRadius() float32
	GetOrbitRadius() float32
	GetMass() float32
	GetType() CelestialBodyType
}
