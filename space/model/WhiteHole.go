package Models

import "golang.org/x/image/math/f32"

type WhiteHole struct {
	Position f32.Vec2
	Radius   float32
	Mass     float32

	// If the player enters this radius then we apply the gravity force
	OrbitRadius float32

	// Image asset path for rendering
	ImagePath string
}

func (wh *WhiteHole) GetPosition() f32.Vec2 {
	return wh.Position
}

func (wh *WhiteHole) GetRadius() float32 {
	return wh.Radius
}

func (wh *WhiteHole) GetOrbitRadius() float32 {
	return wh.OrbitRadius
}

func (wh *WhiteHole) GetMass() float32 {
	return wh.Mass * 100 // Scale mass for white holes (less than black holes)
}

func (wh *WhiteHole) GetType() CelestialBodyType {
	return CelestialBodyTypeWhiteHole
}
