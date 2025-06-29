package Models

import (
	"golang.org/x/image/math/f32"
	"image/color"
)

type Planet struct {
	Name     string
	Position f32.Vec2
	Radius   float32
	Mass     float32

	// If the player enters this radius then we apply the gravity force
	OrbitRadius float32

	// Image asset path for rendering
	ImagePath string

	// Base color for lava effect
	BaseColor color.RGBA

	// Random seed for lava animation
	Seed float32
}

func (p *Planet) GetPosition() f32.Vec2 {
	return p.Position
}

func (p *Planet) GetRadius() float32 {
	return p.Radius
}

func (p *Planet) GetOrbitRadius() float32 {
	return p.OrbitRadius
}

func (p *Planet) GetMass() float32 {
	return p.Mass * 250 // Scale mass for planets
}

func (p *Planet) GetType() CelestialBodyType {
	return CelestialBodyTypePlanet
}
