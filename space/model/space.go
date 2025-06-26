package Models

import (
	"golang.org/x/image/math/f32"
	"math/rand"
)

type SpaceGame struct {
	Planets   []Planet
	BlackHole *BlackHole
}

// NewSpaceGame creates a new SpaceGame with the given size.
func NewSpaceGame() (*SpaceGame, error) {
	// Define some planets with their names, positions, and radii.
	// pick a random number between 3 and 10 for the number of planets
	var numPlanets = rand.Intn(8) + 15 // Random number between 3 and 10'

	const margin = 0.1 // Margin for positioning planets

	var planets = make([]Planet, numPlanets)
	for i := 0; i < numPlanets; i++ {
		// Generate random position and radius for each planet
		x := margin + (rand.Float32() * (1 - margin*2))
		y := margin + (rand.Float32() * (1 - margin*2))
		radius := rand.Float32() + 1 // Random radius between 20 and 50

		planets[i] = Planet{
			Name:     "Planet " + string(rune(i+1)),
			Position: f32.Vec2{x, y},
			Radius:   radius,
		}
	}
	return &SpaceGame{
		Planets: planets,
	}, nil
}
