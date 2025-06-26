package Models

import (
	"golang.org/x/image/math/f32"
	"math/rand"
)

type SpaceGame struct {
	Planets   []Planet
	BlackHole *BlackHole
	Player    *Player
	Camera    *Camera2D
}

// NewSpaceGame creates a new SpaceGame with the given size.
func NewSpaceGame() (*SpaceGame, error) {
	// Define some planets with their names, positions, and radii.
	// pick a random number between 3 and 10 for the number of planets
	var numPlanets = rand.Intn(8) + 3 // Random number between 3 and 10'

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

			OrbitRadius: radius * (3 + (rand.Float32() * 7)),
		}
	}

	var player = &Player{
		Position:     f32.Vec2{margin, 0.5}, // Start near the left edge
		Radius:       1.0,                   // Player radius for collision detection
		Velocity:     f32.Vec2{0, 0},        // No initial velocity
		Acceleration: f32.Vec2{0, 0},        // No initial acceleration
		State:        PlayerStateIdle,
		Mass:         1.0, // Default mass
	}

	// Create camera and set initial target to player
	camera := NewCamera2D()
	camera.SetTarget(player.Position)

	return &SpaceGame{
		Planets: planets,
		Player:  player,
		Camera:  camera,
	}, nil
}

// Reset regenerates the entire game with new planets and resets player
func (sg *SpaceGame) Reset() error {
	// Generate new planets
	var numPlanets = rand.Intn(8) + 3 // Random number between 3 and 10
	const margin = 0.1                // Margin for positioning planets

	var planets = make([]Planet, numPlanets)
	for i := 0; i < numPlanets; i++ {
		// Generate random position and radius for each planet
		x := margin + (rand.Float32() * (1 - margin*2))
		y := margin + (rand.Float32() * (1 - margin*2))
		radius := rand.Float32() + 1 // Random radius between 1 and 2

		planets[i] = Planet{
			Name:        "Planet " + string(rune(i+1)),
			Position:    f32.Vec2{x, y},
			Radius:      radius,
			OrbitRadius: radius * (3 + (rand.Float32() * 7)),
		}
	}

	// Reset player to starting position
	sg.Player.Position = f32.Vec2{margin, 0.5}
	sg.Player.Velocity = f32.Vec2{0, 0}
	sg.Player.Acceleration = f32.Vec2{0, 0}
	sg.Player.State = PlayerStateIdle

	// Reset camera
	sg.Camera.SetTarget(sg.Player.Position)
	sg.Camera.Position = f32.Vec2{0.5, 0.5}

	// Update planets
	sg.Planets = planets

	return nil
}
