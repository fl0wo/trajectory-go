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
	ZoomSpeed   float32 // How fast the zoom interpolates (0-1)

	// View properties
	Zoom       float32  // Current camera zoom level (1.0 = normal)
	TargetZoom float32  // Target zoom level for smooth interpolation
	Offset     f32.Vec2 // Offset from target position
}

// NewCamera2D creates a new camera with default settings
func NewCamera2D() *Camera2D {
	return &Camera2D{
		Position:    f32.Vec2{0.5, 0.5}, // Start at center
		Target:      f32.Vec2{0.5, 0.5},
		SmoothSpeed: 2.0, // Adjust this for faster/slower following
		ZoomSpeed:   5.0, // Adjust this for faster/slower zoom
		Zoom:        1.0,
		TargetZoom:  1.0,
		Offset:      f32.Vec2{0, 0},
	}
}

// Update updates the camera position and zoom to smoothly follow the target
func (c *Camera2D) Update(deltaTime float32) {
	// Calculate desired position (target + offset)
	desiredPos := f32.Vec2{
		c.Target[0] + c.Offset[0],
		c.Target[1] + c.Offset[1],
	}

	// Smooth interpolation towards desired position
	positionT := c.SmoothSpeed * deltaTime
	if positionT > 1.0 {
		positionT = 1.0 // Clamp to prevent overshooting
	}

	// Linear interpolation for position
	c.Position[0] = c.Position[0] + (desiredPos[0]-c.Position[0])*positionT
	c.Position[1] = c.Position[1] + (desiredPos[1]-c.Position[1])*positionT

	// Smooth interpolation towards target zoom
	zoomT := c.ZoomSpeed * deltaTime
	if zoomT > 1.0 {
		zoomT = 1.0 // Clamp to prevent overshooting
	}

	// Linear interpolation for zoom
	c.Zoom = c.Zoom + (c.TargetZoom-c.Zoom)*zoomT
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

// RadiusToScreen converts a radius (in world units) to pixels.
// It scales the radius by the camera zoom, then maps that
// normalized size into screen pixels by averaging the X and Y scales.
func (c *Camera2D) RadiusToScreen(radius float32, screenWidth, screenHeight float32) float32 {
	// 1) apply zoom in world space
	scaled := radius * c.Zoom

	// 2) compute how many pixels one normalized unit is in X and Y
	pxPerUnitX := screenWidth
	pxPerUnitY := screenHeight

	// 3) map the zoomed radius into pixels on each axis
	pixelRadiusX := scaled * pxPerUnitX
	pixelRadiusY := scaled * pxPerUnitY

	// 4) average them to keep the circle isotropic in pixel space
	return (pixelRadiusX + pixelRadiusY) * 0.5
}

// SetZoom sets the target zoom level (will be smoothly interpolated)
func (c *Camera2D) SetZoom(zoom float32) {
	c.TargetZoom = clamp(zoom, 0.1, 3.0) // Clamp zoom to a reasonable range
}

// SetZoomImmediate sets the zoom level immediately without interpolation
func (c *Camera2D) SetZoomImmediate(zoom float32) {
	c.Zoom = zoom
	c.TargetZoom = zoom
}

// GetZoom returns the current zoom level
func (c *Camera2D) GetZoom() float32 {
	return c.Zoom
}

// AdjustZoom adjusts the target camera zoom by the given delta
func (c *Camera2D) AdjustZoom(delta float32) {
	newZoom := c.TargetZoom + delta
	c.SetZoom(newZoom)
}

// ZoomIn increases the zoom level
func (c *Camera2D) ZoomIn(amount float32) {
	c.AdjustZoom(amount)
}

// ZoomOut decreases the zoom level
func (c *Camera2D) ZoomOut(amount float32) {
	c.AdjustZoom(-amount)
}

// clamp restricts a value to a specified range
func clamp(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
