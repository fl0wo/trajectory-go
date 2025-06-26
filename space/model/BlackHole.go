package Models

import "golang.org/x/image/math/f32"

type BlackHole struct {
	Position f32.Vec2
	Radius   float32
	Mass     float32

	// If the player enters this radius then we apply the gravity force
	OrbitRadius float32

	// Image asset path for rendering
	ImagePath string
}

func (bh *BlackHole) GetPosition() f32.Vec2 {
	return bh.Position
}

func (bh *BlackHole) GetRadius() float32 {
	return bh.Radius
}

func (bh *BlackHole) GetOrbitRadius() float32 {
	return bh.OrbitRadius
}

func (bh *BlackHole) GetMass() float32 {
	return bh.Mass * 250 * 3 // 3 Scale mass for black holes
}

func (bh *BlackHole) GetType() CelestialBodyType {
	return CelestialBodyTypeBlackHole
}
