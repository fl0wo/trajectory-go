package Models

import "golang.org/x/image/math/f32"

type Planet struct {
	Name     string
	Position f32.Vec2
	Radius   float32
	Mass     float32

	// If the player enters this radius then we apply the gravity force
	OrbitRadius float32
}
