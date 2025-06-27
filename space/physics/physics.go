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

	// Apply gravitational forces from all celestial bodies
	for _, body := range model.CelestialBodies {
		model.Player.ApplyGravitationalForce(body)

		// Check for collision with celestial body
		if model.Player.CheckCollisionWithCelestialBody(body) {
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
				// Player hit a planet or black hole - reset the level
				err := model.Reset()
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
	switch model.CameraMode {
	case Models.CameraModeCenter:
		// Always follow the center of all entities
		levelCenter := model.CalculateLevelCenter()
		model.Camera.SetTarget(levelCenter)
	case Models.CameraModePlayer:
		// Follow player when moving, center view when idle
		if model.Player.IsMoving() {
			model.Camera.SetTarget(model.Player.Position)
		} else {
			levelCenter := model.CalculateLevelCenter()
			model.Camera.SetTarget(levelCenter)
		}
	}
	model.Camera.Update(deltaTime)
}
