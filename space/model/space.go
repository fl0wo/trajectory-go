package Models

import (
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/resources"
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
	CelestialBodies     []CelestialBody
	RingAsteroids       []*RingAsteroid // Separate list for asteroids that need updates
	Player              *Player
	Camera              *Camera2D
	CurrentLevel        *Level
	CurrentLevelNum     int
	CameraMode          CameraMode // Camera follow mode setting
	TimeScale           float32    // Current time dilation scale (1.0 = normal, 0.1 = 10x slower)
	TargetTimeScale     float32    // Target time scale for smooth interpolation
	ProximityZoom       float32    // Current proximity zoom multiplier (1.0 = normal, 1.25 = zoomed in)
	TargetProximityZoom float32    // Target proximity zoom for smooth interpolation
	ShadowsEnabled      bool       // Toggle for shadow rendering system
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
		Mass:         8.0, // Default mass
	}

	// Create camera
	camera := NewCamera2D()

	// Create temporary SpaceGame to calculate level center
	tempGame := &SpaceGame{
		CelestialBodies:     level.CelestialBodies,
		RingAsteroids:       level.RingAsteroids,
		Player:              player,
		Camera:              camera,
		CurrentLevel:        level,
		CurrentLevelNum:     levelNum,
		CameraMode:          CameraModeCenter, // Default to center mode
		TimeScale:           1.0,              // Normal time initially
		TargetTimeScale:     1.0,              // Target matches current initially
		ProximityZoom:       1.0,              // Normal zoom initially
		TargetProximityZoom: 1.0,              // Target matches current initially
		ShadowsEnabled:      true,             // Enable shadows by default
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
	sg.Player.ClearTrail() // Clear the movement trail

	// Update level data first
	sg.CurrentLevel = level
	sg.CurrentLevelNum = levelNum
	sg.CelestialBodies = level.CelestialBodies
	sg.RingAsteroids = level.RingAsteroids

	// Calculate center of all entities and reset camera to that center
	levelCenter := sg.CalculateLevelCenter()
	sg.Camera.Position = levelCenter
	sg.Camera.SetTarget(levelCenter)

	return nil
}

// CalculateLevelCenter calculates the center position of all entities in the level (player + celestial bodies + asteroids)
func (sg *SpaceGame) CalculateLevelCenter() f32.Vec2 {
	if len(sg.CelestialBodies) == 0 && len(sg.RingAsteroids) == 0 {
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

	// Expand bounds to include all asteroids
	for _, asteroid := range sg.RingAsteroids {
		pos := asteroid.GetPosition()
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

// CalculateTimeDilation calculates the time scale based on player proximity to celestial bodies,
// using an ease-in curve (power law) instead of a linear interpolation.
func (sg *SpaceGame) CalculateTimeDilation() float32 {
	if sg.Player.State != PlayerStateMoving {
		return 1.0 // Normal time when not moving
	}

	minTimeScale := float32(1.0)
	playerPos := sg.Player.Position

	// Curve parameters
	const minTimeScaleAtCenter = 0.025 // 10x slower at the body’s surface
	const exponent = 1.25              // >1 = sharper ease-in

	// Check proximity to all celestial bodies
	for _, body := range sg.CelestialBodies {
		bodyPos := body.GetPosition()
		dx := bodyPos[0] - playerPos[0]
		dy := bodyPos[1] - playerPos[1]
		distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		orbitRadius := body.GetOrbitRadius()
		bodyRadius := body.GetRadius()

		// Only apply time dilation within orbit radius
		if distance <= orbitRadius {
			// Avoid division by zero by clamping at body radius
			minDistance := bodyRadius
			if distance < minDistance {
				distance = minDistance
			}

			// Normalize distance: 0 at surface, 1 at orbit edge
			normalizedDistance := (distance - minDistance) / (orbitRadius - minDistance)

			// Apply ease-in curve via power law
			curve := float32(math.Pow(float64(normalizedDistance), exponent))

			// Map curve [0→1] into timeScale [minTimeScaleAtCenter→1]
			timeScale := minTimeScaleAtCenter + (1.0-minTimeScaleAtCenter)*curve

			// Use the smallest timeScale (maximum slowdown)
			if timeScale < minTimeScale {
				minTimeScale = timeScale
			}
		}
	}

	// Check proximity to all asteroids (they don't have orbit radius but can cause time dilation on collision)
	for _, asteroid := range sg.RingAsteroids {
		bodyPos := asteroid.GetPosition()
		dx := bodyPos[0] - playerPos[0]
		dy := bodyPos[1] - playerPos[1]
		distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		// Only apply time dilation when very close to asteroids (collision imminent)
		asteroidInfluenceRadius := asteroid.GetRadius() * 2.0 // Small influence zone
		if distance <= asteroidInfluenceRadius {
			// Use asteroid radius as both min and max distance for simple collision-based effect
			normalizedDistance := distance / asteroidInfluenceRadius

			// Apply ease-in curve
			curve := float32(math.Pow(float64(normalizedDistance), exponent))

			// Time scale: more dramatic for close asteroid encounters
			timeScale := minTimeScaleAtCenter + (1.0-minTimeScaleAtCenter)*curve

			if timeScale < minTimeScale {
				minTimeScale = timeScale
			}
		}
	}

	return minTimeScale
}

// UpdateTimeDilation updates the current time scale with smooth interpolation
func (sg *SpaceGame) UpdateTimeDilation(deltaTime float32) {
	// Calculate target time scale
	sg.TargetTimeScale = sg.CalculateTimeDilation()

	// Smooth interpolation towards target (faster when slowing down, slower when speeding up)
	timeDilationSpeed := float32(5.0) // Adjust this for faster/slower transitions
	if sg.TargetTimeScale < sg.TimeScale {
		// Slowing down - faster transition for dramatic effect
		timeDilationSpeed = 8.0
	}

	// Interpolate towards target
	t := timeDilationSpeed * deltaTime
	if t > 1.0 {
		t = 1.0 // Clamp to prevent overshooting
	}

	sg.TimeScale = sg.TimeScale + (sg.TargetTimeScale-sg.TimeScale)*t
}

// CalculateProximityZoom calculates a zoom‐out multiplier based on proximity.
// Closer → smaller than 1.0 (zoomed out), far → 1.0 (normal).
func (sg *SpaceGame) CalculateProximityZoom() float32 {
	if sg.Player.State != PlayerStateMoving {
		return 1.0 // Normal zoom when not moving
	}

	// Start at no zoom (1.0) and look for any closer bodies to pull it down.
	minZoomMultiplier := float32(1.0)
	playerPos := sg.Player.Position

	// Zoom parameters
	const minZoomAtCenter = 0.95 // 90% scale at surface (max zoom-out)
	const exponent = 1.05        // Same ease-in curve

	// Celestial bodies
	for _, body := range sg.CelestialBodies {
		bodyPos := body.GetPosition()
		dx := bodyPos[0] - playerPos[0]
		dy := bodyPos[1] - playerPos[1]
		distance := float32(math.Hypot(float64(dx), float64(dy)))

		orbitRadius := body.GetOrbitRadius()
		bodyRadius := body.GetRadius()

		if distance <= orbitRadius {
			// Clamp to avoid zero
			if distance < bodyRadius {
				distance = bodyRadius
			}
			// 0 at surface → 1 at orbit edge
			normalized := (distance - bodyRadius) / (orbitRadius - bodyRadius)
			curve := float32(math.Pow(float64(normalized), exponent))

			// Map [0→1] → [minZoomAtCenter→1]
			zoomMul := minZoomAtCenter + (1.0-minZoomAtCenter)*curve

			if zoomMul < minZoomMultiplier {
				minZoomMultiplier = zoomMul
			}
		}
	}

	// Asteroids
	for _, ast := range sg.RingAsteroids {
		bodyPos := ast.GetPosition()
		dx := bodyPos[0] - playerPos[0]
		dy := bodyPos[1] - playerPos[1]
		distance := float32(math.Hypot(float64(dx), float64(dy)))

		influence := ast.GetRadius() * 2.0
		if distance <= influence {
			normalized := distance / influence
			curve := float32(math.Pow(float64(normalized), exponent))
			zoomMul := minZoomAtCenter + (1.0-minZoomAtCenter)*curve

			if zoomMul < minZoomMultiplier {
				minZoomMultiplier = zoomMul
			}
		}
	}

	return minZoomMultiplier
}

// UpdateProximityZoom smoothly interpolates ProximityZoom → TargetProximityZoom
func (sg *SpaceGame) UpdateProximityZoom(deltaTime float32) {
	sg.TargetProximityZoom = sg.CalculateProximityZoom()

	// Base speed; speed up when zooming out (i.e. target < current) if you like
	zoomSpeed := float32(5.0)
	if sg.TargetProximityZoom < sg.ProximityZoom {
		zoomSpeed = 8.0
	}

	t := zoomSpeed * deltaTime
	if t > 1.0 {
		t = 1.0
	}

	sg.ProximityZoom += (sg.TargetProximityZoom - sg.ProximityZoom) * t
}

// UpdateAsteroids updates all ring asteroids' orbital positions
func (sg *SpaceGame) UpdateAsteroids(deltaTime float32) {
	for _, asteroid := range sg.RingAsteroids {
		asteroid.Update(deltaTime)
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

// ToggleShadows switches shadow rendering on/off
func (sg *SpaceGame) ToggleShadows() {
	sg.ShadowsEnabled = !sg.ShadowsEnabled
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

func (sg *SpaceGame) GetWinPosition() f32.Vec2 {
	for _, body := range sg.CelestialBodies {
		if body.GetType() == CelestialBodyTypeWhiteHole {
			return body.GetPosition()
		}
	}
	return f32.Vec2{0, 0} // Default if no white hole found
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
				ImagePath:   resources.BlackHoleImage,
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
