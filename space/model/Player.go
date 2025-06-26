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
	Size     f32.Vec2

	Velocity f32.Vec2 // current velocity vector (units/sec)

	// Player state
	State PlayerState

	// Visual properties
	Mass float32 // mass for visual effects
}

// Update updates the player's position based on velocity
func (p *Player) Update(deltaTime float32) {
	if p.State == PlayerStateMoving {
		// Simple linear movement - no friction or gravity
		p.Position[0] += p.Velocity[0] * deltaTime
		p.Position[1] += p.Velocity[1] * deltaTime

		// Stop if velocity is very small
		speed := float32(math.Sqrt(float64(p.Velocity[0]*p.Velocity[0] + p.Velocity[1]*p.Velocity[1])))
		if speed < 0.01 {
			p.Velocity = f32.Vec2{0, 0}
			p.State = PlayerStateIdle
		}
	}
}

// Throw launches the player with the given velocity
func (p *Player) Throw(velocity f32.Vec2) {
	p.Velocity = velocity
	p.State = PlayerStateMoving
}

// Stop stops the player movement
func (p *Player) Stop() {
	p.Velocity = f32.Vec2{0, 0}
	p.State = PlayerStateIdle
}

// IsMoving returns true if the player is currently moving
func (p *Player) IsMoving() bool {
	return p.State == PlayerStateMoving
}
