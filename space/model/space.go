package Models

import (
	"golang.org/x/image/math/f32"
	"math"
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
	var numPlanets = rand.Intn(2) + 3 // Random number between 3 and 10'

	const margin = 0.1 // Margin for positioning planets

	planets := randomPlanets(numPlanets, margin)

	var player = &Player{
		Position:     f32.Vec2{margin, 0.5}, // Start near the left edge
		Radius:       1 / 100.0,             // Player radius for collision detection
		Velocity:     f32.Vec2{0, 0},        // No initial velocity
		Acceleration: f32.Vec2{0, 0},        // No initial acceleration
		State:        PlayerStateIdle,
		Mass:         10.0, // Default mass
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
	var numPlanets = rand.Intn(2) + 3 // Random number between 3 and 10
	const margin = 0.1                // Margin for positioning planets

	planets := randomPlanets(numPlanets, margin)

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

func randomPlanets(numPlanets int, margin float32) []Planet {
	var planets = make([]Planet, numPlanets)
	for i := 0; i < numPlanets; i++ {
		// Generate random position and radius for each planet
		x := margin + (rand.Float32() * (1 - margin*2)) // value between margin and 1-margin
		y := margin + (rand.Float32() * (1 - margin*2)) // value between margin and 1-margin

		// radius in [0, 1]
		radius := rand.Float32() // value between 0 and 1

		// for fun: make mass proportional to volume of a sphere
		// mass := (4.0 / 3.0) * math.Pi * radius * radius * radius

		mass := float32(math.Inf(1)) // infinite mass for simplicity

		// ensure orbitRadius > radius and ≤ 1

		planets[i] = Planet{
			Name:     "Planet " + string(rune(i+1)),
			Position: f32.Vec2{x, y},

			Mass: mass,

			Radius:      radius / 10.0,
			OrbitRadius: radius / 2.0,
		}
	}
	return planets
}
