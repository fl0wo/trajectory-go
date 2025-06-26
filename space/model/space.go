package Models

import (
	"github.com/you/trajectory/constants"
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
		Radius:       0.02,                  // Player radius for collision detection
		Velocity:     f32.Vec2{0, 0},        // No initial velocity
		Acceleration: f32.Vec2{0, 0},        // No initial acceleration
		State:        PlayerStateIdle,
		Mass:         10.0, // Default mass
	}

	// Create camera and set initial target to player
	camera := NewCamera2D()
	// Set camera center to middle of aspect-ratio-aware world space
	camera.Position = f32.Vec2{constants.AspectRatio / 2.0, 0.5}
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
	sg.Camera.Position = f32.Vec2{constants.AspectRatio / 2.0, 0.5}

	// Update planets
	sg.Planets = planets

	return nil
}

func randomPlanets(numPlanets int, margin float32) []Planet {
	var planets = make([]Planet, numPlanets)

	// Use shared aspect ratio constant
	aspectRatio := constants.AspectRatio

	for i := 0; i < numPlanets; i++ {
		// Generate random position and radius for each planet
		// X coordinate spans [0, aspectRatio], Y coordinate spans [0, 1]
		x := margin + (rand.Float32() * (aspectRatio - margin*2)) // value between margin and aspectRatio-margin
		y := margin + (rand.Float32() * (1 - margin*2))           // value between margin and 1-margin

		// radius in [0, 1] in Y-coordinate scale
		radius := rand.Float32() / 5.0 // smaller planets for better gameplay

		// for fun: make mass proportional to volume of a sphere
		mass := (4.0 / 3.0) * math.Pi * radius * radius * radius
		mass *= 250

		// mass := float32(math.Inf(1)) // infinite mass for simplicity

		planets[i] = Planet{
			Name:     "Planet " + string(rune(i+1)),
			Position: f32.Vec2{x, y},

			Mass: float32(mass),

			Radius:      radius,
			OrbitRadius: radius * 3.0, // orbit radius is 3x the planet radius
		}
	}
	return planets
}
