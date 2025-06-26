package Models

import (
	"golang.org/x/image/math/f32"
	"math"
)

// RingAsteroid represents an asteroid that orbits around a host planet
type RingAsteroid struct {
	// Host planet that this asteroid orbits around
	HostPlanet CelestialBody

	// Physical properties
	Radius float32 // Asteroid radius
	Mass   float32 // Asteroid mass

	// Orbital properties
	OrbitRadius    float32 // Distance from host planet center
	OrbitSpeed     float32 // Angular velocity in radians per second
	CurrentAngle   float32 // Current orbital angle in radians
	OrbitDirection int     // 1 for clockwise, -1 for counter-clockwise

	// Cached position (updated each frame)
	Position f32.Vec2

	// For time-based updates
	LastUpdateTime float64
}

// NewRingAsteroid creates a new ring asteroid orbiting a host planet
func NewRingAsteroid(hostPlanet CelestialBody, orbitRadius, radius, mass, orbitSpeed float32, startAngle float32, clockwise bool) *RingAsteroid {
	direction := 1
	if !clockwise {
		direction = -1
	}

	asteroid := &RingAsteroid{
		HostPlanet:     hostPlanet,
		Radius:         radius,
		Mass:           mass,
		OrbitRadius:    orbitRadius,
		OrbitSpeed:     orbitSpeed,
		CurrentAngle:   startAngle,
		OrbitDirection: direction,
		LastUpdateTime: 0,
	}

	// Calculate initial position
	asteroid.updatePosition()
	return asteroid
}

// Update updates the asteroid's orbital position based on time
func (ra *RingAsteroid) Update(deltaTime float32) {
	// Update orbital angle based on speed and time
	ra.CurrentAngle += ra.OrbitSpeed * deltaTime * float32(ra.OrbitDirection)

	// Keep angle in [0, 2π] range
	for ra.CurrentAngle >= 2*math.Pi {
		ra.CurrentAngle -= 2 * math.Pi
	}
	for ra.CurrentAngle < 0 {
		ra.CurrentAngle += 2 * math.Pi
	}

	// Update position based on new angle
	ra.updatePosition()
}

// updatePosition calculates the asteroid's current position based on orbital parameters
func (ra *RingAsteroid) updatePosition() {
	hostPos := ra.HostPlanet.GetPosition()

	// Calculate position using polar coordinates
	x := hostPos[0] + ra.OrbitRadius*float32(math.Cos(float64(ra.CurrentAngle)))
	y := hostPos[1] + ra.OrbitRadius*float32(math.Sin(float64(ra.CurrentAngle)))

	ra.Position = f32.Vec2{x, y}
}

// CelestialBody interface implementation
func (ra *RingAsteroid) GetPosition() f32.Vec2 {
	return ra.Position
}

func (ra *RingAsteroid) GetRadius() float32 {
	return ra.Radius
}

func (ra *RingAsteroid) GetOrbitRadius() float32 {
	// Asteroids don't have their own gravity wells, so orbit radius is 0
	return 0
}

func (ra *RingAsteroid) GetMass() float32 {
	// Asteroids are much lighter than planets
	return ra.Mass
}

func (ra *RingAsteroid) GetType() CelestialBodyType {
	return CelestialBodyTypeAsteroid
}

// GetHostPlanet returns the planet this asteroid orbits around
func (ra *RingAsteroid) GetHostPlanet() CelestialBody {
	return ra.HostPlanet
}

// GetOrbitSpeed returns the orbital speed in radians per second
func (ra *RingAsteroid) GetOrbitSpeed() float32 {
	return ra.OrbitSpeed
}

// GetCurrentAngle returns the current orbital angle in radians
func (ra *RingAsteroid) GetCurrentAngle() float32 {
	return ra.CurrentAngle
}
