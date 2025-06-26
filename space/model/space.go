package Models

import (
	"github.com/you/trajectory/constants"
	"golang.org/x/image/math/f32"
	"math"
	"math/rand"
)

// CameraMode defines how the camera follows entities
type CameraMode int

const (
	CameraModeCenter CameraMode = iota // Follow center of all entities (default)
	CameraModePlayer                   // Follow player only when moving
)

type SpaceGame struct {
	CelestialBodies []CelestialBody
	Player          *Player
	Camera          *Camera2D
	CurrentLevel    *Level
	CurrentLevelNum int
	CameraMode      CameraMode // Camera follow mode setting
}

// NewSpaceGame creates a new SpaceGame starting with level 1.
func NewSpaceGame() (*SpaceGame, error) {
	return NewSpaceGameWithLevel(1)
}

// NewSpaceGameWithLevel creates a new SpaceGame with a specific level
func NewSpaceGameWithLevel(levelNum int) (*SpaceGame, error) {
	level := GetLevel(levelNum)

	var player = &Player{
		Position:     level.PlayerStart,
		Radius:       0.02,           // Player radius for collision detection
		Velocity:     f32.Vec2{0, 0}, // No initial velocity
		Acceleration: f32.Vec2{0, 0}, // No initial acceleration
		State:        PlayerStateIdle,
		Mass:         10.0, // Default mass
	}

	// Create camera
	camera := NewCamera2D()

	// Create temporary SpaceGame to calculate level center
	tempGame := &SpaceGame{
		CelestialBodies: level.CelestialBodies,
		Player:          player,
		Camera:          camera,
		CurrentLevel:    level,
		CurrentLevelNum: levelNum,
		CameraMode:      CameraModeCenter, // Default to center mode
	}

	// Calculate center of all entities and set camera position
	levelCenter := tempGame.CalculateLevelCenter()
	camera.Position = levelCenter
	camera.SetTarget(levelCenter)

	return tempGame, nil
}

// Reset resets the current level
func (sg *SpaceGame) Reset() error {
	return sg.LoadLevel(sg.CurrentLevelNum)
}

// LoadLevel loads a specific level
func (sg *SpaceGame) LoadLevel(levelNum int) error {
	level := GetLevel(levelNum)

	// Reset player to level's starting position
	sg.Player.Position = level.PlayerStart
	sg.Player.Velocity = f32.Vec2{0, 0}
	sg.Player.Acceleration = f32.Vec2{0, 0}
	sg.Player.State = PlayerStateIdle

	// Update level data first
	sg.CurrentLevel = level
	sg.CurrentLevelNum = levelNum
	sg.CelestialBodies = level.CelestialBodies

	// Calculate center of all entities and reset camera to that center
	levelCenter := sg.CalculateLevelCenter()
	sg.Camera.Position = levelCenter
	sg.Camera.SetTarget(levelCenter)

	return nil
}

// CalculateLevelCenter calculates the center position of all entities in the level (player + celestial bodies)
func (sg *SpaceGame) CalculateLevelCenter() f32.Vec2 {
	if len(sg.CelestialBodies) == 0 {
		return sg.Player.Position
	}

	// Initialize bounds with player position
	minX := sg.Player.Position[0]
	maxX := sg.Player.Position[0]
	minY := sg.Player.Position[1]
	maxY := sg.Player.Position[1]

	// Expand bounds to include all celestial bodies
	for _, body := range sg.CelestialBodies {
		pos := body.GetPosition()
		if pos[0] < minX {
			minX = pos[0]
		}
		if pos[0] > maxX {
			maxX = pos[0]
		}
		if pos[1] < minY {
			minY = pos[1]
		}
		if pos[1] > maxY {
			maxY = pos[1]
		}
	}

	// Return center of bounding box
	return f32.Vec2{
		(minX + maxX) / 2.0,
		(minY + maxY) / 2.0,
	}
}

// ToggleCameraMode switches between camera modes
func (sg *SpaceGame) ToggleCameraMode() {
	if sg.CameraMode == CameraModeCenter {
		sg.CameraMode = CameraModePlayer
	} else {
		sg.CameraMode = CameraModeCenter
	}
}

// GetCameraModeString returns a string representation of the current camera mode
func (sg *SpaceGame) GetCameraModeString() string {
	switch sg.CameraMode {
	case CameraModeCenter:
		return "Center View"
	case CameraModePlayer:
		return "Player Follow"
	default:
		return "Unknown"
	}
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
