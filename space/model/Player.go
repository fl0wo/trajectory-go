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

// Player.ApplyGravitationalForce applies gravity only if within planet.OrbitRadius
func (p *Player) ApplyGravitationalForce(planet Planet) {
	if p.State != PlayerStateMoving {
		return
	}

	// vector from player → planet
	dx := planet.Position[0] - p.Position[0]
	dy := planet.Position[1] - p.Position[1]
	distSq := dx*dx + dy*dy
	if distSq == 0 || math.IsNaN(float64(distSq)) {
		return
	}

	// if you're outside the planet's "orbit radius", skip it:
	orbitR := planet.OrbitRadius
	if distSq > (orbitR * orbitR) {
		return
	}

	// F = G m1 m2 / r^2
	const G = 0.01
	force := G * p.Mass * planet.Mass / distSq

	// now turn that into an accel vector: a = F/m1 * unitDir
	// we'll need the actual distance for the unit vector:
	invDist := 1.0 / float32(math.Sqrt(float64(distSq)))
	ux := dx * invDist
	uy := dy * invDist

	ax := force * ux / p.Mass
	ay := force * uy / p.Mass

	p.Acceleration[0] += ax
	p.Acceleration[1] += ay
}

// CheckCollisionWithPlanet is now just a squared‐radius test:
func (p *Player) CheckCollisionWithPlanet(planet Planet) bool {
	dx := planet.Position[0] - p.Position[0]
	dy := planet.Position[1] - p.Position[1]
	distSq := dx*dx + dy*dy

	r := planet.Radius + p.Radius
	return distSq <= r*r
}

// ResetAcceleration clears all accumulated accelerations
func (p *Player) ResetAcceleration() {
	p.Acceleration = f32.Vec2{0, 0}
}
