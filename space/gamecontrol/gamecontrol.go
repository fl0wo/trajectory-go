package gamecontrol

import (
	"github.com/you/trajectory/space/input"
	Models "github.com/you/trajectory/space/model"
	"github.com/you/trajectory/space/physics"
)

// GameController handles game state management and input processing
type GameController struct {
	physicsSystem *physics.PhysicsSystem
}

// NewGameController creates a new game controller
func NewGameController() *GameController {
	return &GameController{
		physicsSystem: physics.NewPhysicsSystem(),
	}
}

// ProcessInput handles all input processing for the game
func (gc *GameController) ProcessInput(model *Models.SpaceGame, inputHandler *input.Input) error {
	// Handle restart key
	if inputHandler.IsRestartPressed() {
		err := model.Reset()
		if err != nil {
			return err
		}
		return nil // Exit early since game was reset
	}

	// Handle camera toggle (C key)
	if inputHandler.IsCameraTogglePressed() {
		model.ToggleCameraMode()
	}

	// Handle shadow toggle (S key)
	if inputHandler.IsShadowTogglePressed() {
		model.ToggleShadows()
	}

	// Handle level selection keys (1-9)
	levelKey := inputHandler.GetLevelKeyPressed()
	if levelKey > 0 {
		err := model.LoadLevel(levelKey)
		if err != nil {
			return err
		}
		return nil // Exit early since level was changed
	}

	// Handle scroll zoom
	scrollDelta := inputHandler.GetScrollDelta()
	if scrollDelta != 0 {
		// Zoom sensitivity
		zoomSpeed := float32(0.1)
		model.Camera.AdjustZoom(scrollDelta * zoomSpeed)
	}

	// Handle drag and throw mechanics
	dragInfo := inputHandler.GetDragInfo()

	// Apply zoom-out effect when dragging (only if player is not moving)
	if dragInfo.IsDragging && !model.Player.IsMoving() {
		// Calculate zoom-out based on drag distance
		const maxDragDistance = float32(150.0) // Max drag distance for full zoom effect
		const maxZoomOut = float32(-0.6)       // Maximum zoom-out amount

		// Normalize drag distance to [0, 1] range
		normalizedDistance := dragInfo.DragDistance / maxDragDistance
		if normalizedDistance > 1.0 {
			normalizedDistance = 1.0
		}

		// Apply eased zoom-out (smooth curve)
		zoomOutFactor := normalizedDistance * normalizedDistance * maxZoomOut / 10.0
		model.Camera.SetDragZoom(zoomOutFactor)
	} else {
		// Reset drag zoom when not dragging
		model.Camera.SetDragZoom(0.0)
	}

	if dragInfo.IsReleased {
		gc.physicsSystem.ProcessThrow(model, dragInfo.StartPos, dragInfo.CurrentPos)
	}

	return nil
}

// UpdatePhysics updates all physics systems
func (gc *GameController) UpdatePhysics(model *Models.SpaceGame, deltaTime float32, timeScale float32) error {
	// Update physics
	err := gc.physicsSystem.UpdateGame(model, deltaTime, timeScale)
	if err != nil {
		return err
	}

	// Update camera physics
	gc.physicsSystem.UpdateCameraPhysics(model, deltaTime)

	return nil
}

// CalculateTimeDilation calculates time dilation and proximity zoom effects
func (gc *GameController) CalculateTimeDilation(model *Models.SpaceGame, baseDeltaTime float32) {
	// Calculate time dilation based on player proximity to celestial bodies
	model.UpdateTimeDilation(baseDeltaTime)

	// Calculate proximity zoom based on player proximity to celestial bodies
	model.UpdateProximityZoom(baseDeltaTime)

	// Apply proximity zoom to camera
	model.Camera.SetProximityZoom(model.ProximityZoom)
}
