package Models

import (
	"golang.org/x/image/math/f32"
	"math"
)

// PlayerState represents the current state of the player
type PlayerState int

const (
	PlayerStateIdle PlayerState = iota
	PlayerStateMoving
)

// Player represents a projectile object in 2D space that can be thrown
type Player struct {
	// kinematic state
	Position f32.Vec2 // current world position (normalized 0-1)
	Radius   float32  // player radius for circular collision detection

	Velocity     f32.Vec2 // current velocity vector (units/sec)
	Acceleration f32.Vec2 // current acceleration from gravitational forces

	// Player state
	State PlayerState

	// Physics properties
	Mass float32 // mass for gravitational calculations
}

// Update updates the player's position based on velocity and acceleration
func (p *Player) Update(deltaTime float32) {
	if p.State == PlayerStateMoving {
		// Apply acceleration to velocity
		p.Velocity[0] += p.Acceleration[0] * deltaTime
		p.Velocity[1] += p.Acceleration[1] * deltaTime

		// Update position based on velocity
		p.Position[0] += p.Velocity[0] * deltaTime
		p.Position[1] += p.Velocity[1] * deltaTime

		// Stop if velocity is very small
		speed := float32(math.Sqrt(float64(p.Velocity[0]*p.Velocity[0] + p.Velocity[1]*p.Velocity[1])))
		if speed < 0.01 {
			p.Velocity = f32.Vec2{0, 0}
			p.Acceleration = f32.Vec2{0, 0}
			p.State = PlayerStateIdle
		}
	}
}

// Throw launches the player with the given velocity
func (p *Player) Throw(velocity f32.Vec2) {
	p.Velocity = velocity
	p.Acceleration = f32.Vec2{0, 0} // Reset acceleration
	p.State = PlayerStateMoving
}

// Stop stops the player movement
func (p *Player) Stop() {
	p.Velocity = f32.Vec2{0, 0}
	p.Acceleration = f32.Vec2{0, 0}
	p.State = PlayerStateIdle
}

// IsMoving returns true if the player is currently moving
func (p *Player) IsMoving() bool {
	return p.State == PlayerStateMoving
}

// ApplyGravitationalForce calculates and applies gravitational force from a planet
func (p *Player) ApplyGravitationalForce(planet Planet) {
	if p.State != PlayerStateMoving {
		return
	}

	// Calculate distance vector from player to planet in normalized coordinates
	dx := planet.Position[0] - p.Position[0]
	dy := planet.Position[1] - p.Position[1]
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	// Only apply gravity if player is within orbit radius
	// Both distance and OrbitRadius are now in the same normalized coordinate system
	if distance > planet.OrbitRadius {
		return
	}

	// Prevent division by zero and extreme forces when very close
	if distance < 0.01 {
		distance = 0.01
	}

	// Calculate gravitational force using F = G * m1 * m2 / r^2
	// We'll use planet radius as mass proxy and normalize constants
	const gravitationalConstant = 0.01
	planetMass := planet.Mass
	force := gravitationalConstant * p.Mass * planetMass / (distance * distance)

	// Calculate unit vector pointing toward planet
	unitX := dx / distance
	unitY := dy / distance

	// Calculate acceleration (F = m * a, so a = F / m)
	accelerationX := force * unitX / p.Mass
	accelerationY := force * unitY / p.Mass

	// Apply the acceleration
	p.Acceleration[0] += accelerationX
	p.Acceleration[1] += accelerationY
}

// CheckCollisionWithPlanet checks if player collides with planet surface
func (p *Player) CheckCollisionWithPlanet(planet Planet) bool {
	dx := planet.Position[0] - p.Position[0]
	dy := planet.Position[1] - p.Position[1]
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	// Convert radii to normalized coordinates
	planetRadius := planet.Radius
	playerRadius := p.Radius

	return distance < (planetRadius + playerRadius)
}

// ResetAcceleration clears all accumulated accelerations
func (p *Player) ResetAcceleration() {
	p.Acceleration = f32.Vec2{0, 0}
}
