package Models

import "golang.org/x/image/math/f32"

type Planet struct {
	Name     string
	Position f32.Vec2
	// The mass is calculated based on how visually big this planet is (it's radius)
	Radius float32

	// If the player enters this radius then we apply the gravity force
	OrbitRadius float32
}
