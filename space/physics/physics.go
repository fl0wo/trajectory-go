package physics

import (
	Models "github.com/you/trajectory/space/model"
	"golang.org/x/image/math/f32"
	"math"
)

const (
	// Maximum drag distance in world units to limit throw power
	MaxDragDistance = float32(0.25)
	// Maximum velocity magnitude to prevent excessive speeds
	MaxVelocity = float32(0.5)
)

// PhysicsSystem handles all physics calculations and updates
type PhysicsSystem struct{}

// NewPhysicsSystem creates a new physics system
func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{}
}

// UpdateGame updates the complete game physics state
func (p *PhysicsSystem) UpdateGame(model *Models.SpaceGame, deltaTime float32, timeScale float32) error {
	// Reset player acceleration for this frame
	model.Player.ResetAcceleration()

	// Update asteroid positions
	model.UpdateAsteroids(deltaTime)

	// Update portal states (cooldown timers)
	model.UpdatePortals(deltaTime)

	// Check for portal collisions and handle teleportation
	model.CheckPortalCollisions()

	// Apply gravitational forces from all celestial bodies
	for i, body := range model.CelestialBodies {
		model.Player.ApplyGravitationalForce(body)

		// Check for collision with celestial body
		if model.Player.CheckCollisionWithCelestialBody(body) {
			// Calculate the exact collision point
			playerPos := model.Player.Position
			bodyPos := body.GetPosition()
			playerRadius := model.Player.Radius
			bodyRadius := body.GetRadius()

			// Distance between centers
			delta := f32.Vec2{bodyPos[0] - playerPos[0], bodyPos[1] - playerPos[1]}
			dist := float32(math.Sqrt(float64(delta[0]*delta[0] + delta[1]*delta[1])))

			var collisionPos f32.Vec2

			// Avoid division by zero
			if dist < 1e-6 {
				// If centers are nearly coincident, use player position as fallback
				collisionPos = playerPos
			} else {
				// Compute collision point: interpolate along the line from player to body
				// The collision point is at the edge of the player's circle, scaled by the ratio of radii
				totalRadius := playerRadius + bodyRadius
				t := playerRadius / totalRadius // Fraction along the line where circles touch
				collisionPos = f32.Vec2{
					playerPos[0] + t*delta[0],
					playerPos[1] + t*delta[1],
				}
			}

			// Check if it's a white hole (victory condition)
			if body.GetType() == Models.CelestialBodyTypeWhiteHole {
				// Victory! Move to next level
				nextLevel := model.CurrentLevelNum + 1
				if nextLevel > 9 {
					// If beyond level 9, restart from level 1
					nextLevel = 1
				}
				err := model.LoadLevel(nextLevel)
				if err != nil {
					return err
				}
				return nil // Exit early since level was changed
			} else {
				// Player hit a planet or black hole - reset the level with collision info
				err := model.ResetWithCollision(collisionPos, i)
				if err != nil {
					return err
				}
				return nil // Exit early since game was reset
			}
		}
	}

	// Check for collisions with asteroids
	for _, asteroid := range model.RingAsteroids {
		if model.Player.CheckCollisionWithCelestialBody(asteroid) {
			// Player hit an asteroid - reset the level
			err := model.Reset()
			if err != nil {
				return err
			}
			return nil // Exit early since game was reset
		}
	}

	// Update shake effect
	model.UpdateShake(deltaTime)

	// Update player physics
	model.Player.Update(deltaTime, timeScale)

	return nil
}

// ProcessThrow processes the throw mechanics from drag input
func (p *PhysicsSystem) ProcessThrow(model *Models.SpaceGame, startScreenPos, endScreenPos f32.Vec2) {
	if model.Player.IsMoving() {
		return // Don't allow throwing while moving
	}

	// Convert screen drag to world velocity
	camera := model.Camera
	startWorld := camera.ScreenToWorld(startScreenPos, 2080, 1080) // Use game screen constants
	endWorld := camera.ScreenToWorld(endScreenPos, 2080, 1080)

	// Calculate throw vector (opposite direction of drag)
	throwVector := f32.Vec2{
		startWorld[0] - endWorld[0],
		startWorld[1] - endWorld[1],
	}

	// Calculate drag distance and limit it to maximum allowed
	dragDistance := float32(math.Sqrt(float64(throwVector[0]*throwVector[0] + throwVector[1]*throwVector[1])))

	// Only throw if drag distance is significant
	if dragDistance > 0.01 {
		// Clamp drag distance to maximum allowed
		if dragDistance > MaxDragDistance {
			// Normalize the vector and scale it to max distance
			throwVector[0] = (throwVector[0] / dragDistance) * MaxDragDistance
			throwVector[1] = (throwVector[1] / dragDistance) * MaxDragDistance
			dragDistance = MaxDragDistance
		}

		// Scale velocity based on clamped drag distance
		velocityMultiplier := float32(2.0)
		throwVelocity := f32.Vec2{
			throwVector[0] * velocityMultiplier,
			throwVector[1] * velocityMultiplier,
		}

		// Additional safety: clamp final velocity magnitude
		velocityMagnitude := float32(math.Sqrt(float64(throwVelocity[0]*throwVelocity[0] + throwVelocity[1]*throwVelocity[1])))
		if velocityMagnitude > MaxVelocity {
			throwVelocity[0] = (throwVelocity[0] / velocityMagnitude) * MaxVelocity
			throwVelocity[1] = (throwVelocity[1] / velocityMagnitude) * MaxVelocity
		}

		model.Player.Throw(throwVelocity)
	}
}

// UpdateCameraPhysics updates camera behavior based on camera mode
func (p *PhysicsSystem) UpdateCameraPhysics(model *Models.SpaceGame, deltaTime float32) {
	// camera follows the level center always
	model.Camera.SetTarget(model.CalculateLevelCenter())

	model.Camera.Update(deltaTime)
}
