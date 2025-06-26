package Models

import (
	"golang.org/x/image/math/f32"
)

// Camera2D represents a 2D camera that can follow targets smoothly
type Camera2D struct {
	Position f32.Vec2 // Camera position in world space (normalized 0-1)
	Target   f32.Vec2 // Target position to follow

	// Smoothing parameters
	SmoothSpeed float32 // How fast the camera follows the target (0-1)

	// View properties
	Zoom   float32  // Camera zoom level (1.0 = normal)
	Offset f32.Vec2 // Offset from target position
}

// NewCamera2D creates a new camera with default settings
func NewCamera2D() *Camera2D {
	return &Camera2D{
		Position:    f32.Vec2{0.5, 0.5}, // Start at center
		Target:      f32.Vec2{0.5, 0.5},
		SmoothSpeed: 2.0, // Adjust this for faster/slower following
		Zoom:        1.0,
		Offset:      f32.Vec2{0, 0},
	}
}

// Update updates the camera position to smoothly follow the target
func (c *Camera2D) Update(deltaTime float32) {
	// Calculate desired position (target + offset)
	desiredPos := f32.Vec2{
		c.Target[0] + c.Offset[0],
		c.Target[1] + c.Offset[1],
	}

	// Smooth interpolation towards desired position
	t := c.SmoothSpeed * deltaTime
	if t > 1.0 {
		t = 1.0 // Clamp to prevent overshooting
	}

	// Linear interpolation
	c.Position[0] = c.Position[0] + (desiredPos[0]-c.Position[0])*t
	c.Position[1] = c.Position[1] + (desiredPos[1]-c.Position[1])*t
}

// SetTarget sets the camera's target position
func (c *Camera2D) SetTarget(target f32.Vec2) {
	c.Target = target
}

// GetTarget returns the current target position
func (c *Camera2D) GetTarget() f32.Vec2 {
	return c.Target
}

// WorldToScreen converts world coordinates to screen coordinates
func (c *Camera2D) WorldToScreen(worldPos f32.Vec2, screenWidth, screenHeight float32) f32.Vec2 {
	// Apply camera offset and zoom
	relativePos := f32.Vec2{
		(worldPos[0] - c.Position[0]) * c.Zoom,
		(worldPos[1] - c.Position[1]) * c.Zoom,
	}

	// Convert to screen coordinates
	screenPos := f32.Vec2{
		(relativePos[0] + 0.5) * screenWidth,
		(relativePos[1] + 0.5) * screenHeight,
	}

	return screenPos
}

// ScreenToWorld converts screen coordinates to world coordinates
func (c *Camera2D) ScreenToWorld(screenPos f32.Vec2, screenWidth, screenHeight float32) f32.Vec2 {
	// Convert screen to normalized coordinates
	normalizedPos := f32.Vec2{
		(screenPos[0] / screenWidth) - 0.5,
		(screenPos[1] / screenHeight) - 0.5,
	}

	// Apply inverse camera transform
	worldPos := f32.Vec2{
		(normalizedPos[0] / c.Zoom) + c.Position[0],
		(normalizedPos[1] / c.Zoom) + c.Position[1],
	}

	return worldPos
}

// SetZoom sets the camera zoom level
func (c *Camera2D) SetZoom(zoom float32) {
	if zoom <= 0 {
		zoom = 0.1 // Minimum zoom
	}
	c.Zoom = zoom
}

// GetZoom returns the current zoom level
func (c *Camera2D) GetZoom() float32 {
	return c.Zoom
}
