package Models

import (
	"github.com/you/trajectory/constants"
	"golang.org/x/image/math/f32"
	"math"
	"math/rand"
)

type SpaceGame struct {
	CelestialBodies []CelestialBody
	Player          *Player
	Camera          *Camera2D
}

// NewSpaceGame creates a new SpaceGame with the given size.
func NewSpaceGame() (*SpaceGame, error) {
	// Generate random celestial bodies (planets and blackholes)
	// pick a random number between 3 and 10 for total celestial bodies
	var numBodies = rand.Intn(8) + 3 // Random number between 3 and 10

	const margin = 0.1 // Margin for positioning celestial bodies

	celestialBodies := randomCelestialBodies(numBodies, margin)

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
		CelestialBodies: celestialBodies,
		Player:          player,
		Camera:          camera,
	}, nil
}

// Reset regenerates the entire game with new celestial bodies and resets player
func (sg *SpaceGame) Reset() error {
	// Generate new celestial bodies
	var numBodies = rand.Intn(8) + 3 // Random number between 3 and 10
	const margin = 0.1               // Margin for positioning celestial bodies

	celestialBodies := randomCelestialBodies(numBodies, margin)

	// Reset player to starting position
	sg.Player.Position = f32.Vec2{margin, 0.5}
	sg.Player.Velocity = f32.Vec2{0, 0}
	sg.Player.Acceleration = f32.Vec2{0, 0}
	sg.Player.State = PlayerStateIdle

	// Reset camera
	sg.Camera.SetTarget(sg.Player.Position)
	sg.Camera.Position = f32.Vec2{constants.AspectRatio / 2.0, 0.5}

	// Update celestial bodies
	sg.CelestialBodies = celestialBodies

	return nil
}

func randomCelestialBodies(numBodies int, margin float32) []CelestialBody {
	var bodies = make([]CelestialBody, numBodies)

	// Use shared aspect ratio constant
	aspectRatio := constants.AspectRatio

	for i := 0; i < numBodies; i++ {
		// Generate random position and radius for each celestial body
		// X coordinate spans [0, aspectRatio], Y coordinate spans [0, 1]
		x := margin + (rand.Float32() * (aspectRatio - margin*2)) // value between margin and aspectRatio-margin
		y := margin + (rand.Float32() * (1 - margin*2))           // value between margin and 1-margin

		// radius in [0, 1] in Y-coordinate scale
		radius := rand.Float32() / 5.0 // smaller celestial bodies for better gameplay

		// for fun: make mass proportional to volume of a sphere
		mass := (4.0 / 3.0) * math.Pi * radius * radius * radius

		// Randomly decide if this should be a planet or blackhole (70% planet, 30% blackhole)
		if rand.Float32() < 0.7 {
			bodies[i] = &Planet{
				Name:        "Planet " + string(rune(i+1)),
				Position:    f32.Vec2{x, y},
				Mass:        mass, // scale mass for planets
				Radius:      radius,
				OrbitRadius: radius * 3.0, // orbit radius is 3x the planet radius
			}
		} else {
			// Blackholes are typically more massive and have larger orbit radius
			bodies[i] = &BlackHole{
				Position:    f32.Vec2{x, y},
				Mass:        mass,
				Radius:      radius,
				OrbitRadius: radius * 4.0, // orbit radius is 4x the blackhole radius
			}
		}
	}
	return bodies
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
