package Models

import (
	"github.com/you/trajectory/constants"
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
	Zoom           float32  // Current camera zoom level (1.0 = normal)
	TargetZoom     float32  // Target zoom level for smooth interpolation
	ProximityZoom  float32  // Additional zoom multiplier from proximity effects
	DragZoom       float32  // Current drag zoom offset (negative values zoom out)
	TargetDragZoom float32  // Target drag zoom for smooth interpolation
	Offset         f32.Vec2 // Offset from target position
}

// NewCamera2D creates a new camera with default settings
func NewCamera2D() *Camera2D {
	// Calculate aspect ratio for default center position
	center := f32.Vec2{constants.AspectRatio / 2.0, 0.5}

	return &Camera2D{
		Position:       center, // Start at aspect-ratio-aware center
		Target:         center,
		SmoothSpeed:    2.0, // Adjust this for faster/slower following
		ZoomSpeed:      5.0, // Adjust this for faster/slower zoom
		Zoom:           1.0,
		TargetZoom:     1.0,
		ProximityZoom:  1.0, // Default proximity zoom multiplier
		DragZoom:       0.0, // Default no drag zoom
		TargetDragZoom: 0.0, // Default no target drag zoom
		Offset:         f32.Vec2{0, 0},
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

	// Smooth interpolation towards target drag zoom (using same speed as regular zoom)
	c.DragZoom = c.DragZoom + (c.TargetDragZoom-c.DragZoom)*zoomT
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
	// Calculate aspect ratio
	aspectRatio := screenWidth / screenHeight

	// Calculate effective zoom (user zoom * proximity zoom + drag zoom)
	effectiveZoom := (c.Zoom + c.DragZoom) * c.ProximityZoom

	// Apply camera offset and zoom
	relativePos := f32.Vec2{
		(worldPos[0] - c.Position[0]) * effectiveZoom,
		(worldPos[1] - c.Position[1]) * effectiveZoom,
	}

	// Convert to screen coordinates with aspect ratio correction
	// Y remains [0,1], X becomes [0, aspectRatio] in world space
	screenPos := f32.Vec2{
		(relativePos[0]/aspectRatio + 0.5) * screenWidth,
		(relativePos[1] + 0.5) * screenHeight,
	}

	return screenPos
}

// ScreenToWorld converts screen coordinates to world coordinates
func (c *Camera2D) ScreenToWorld(screenPos f32.Vec2, screenWidth, screenHeight float32) f32.Vec2 {
	// Calculate aspect ratio
	aspectRatio := screenWidth / screenHeight

	// Calculate effective zoom (user zoom * proximity zoom + drag zoom)
	effectiveZoom := (c.Zoom + c.DragZoom) * c.ProximityZoom

	// Convert screen to normalized coordinates
	normalizedPos := f32.Vec2{
		(screenPos[0] / screenWidth) - 0.5,
		(screenPos[1] / screenHeight) - 0.5,
	}

	// Apply inverse camera transform with aspect ratio correction
	worldPos := f32.Vec2{
		(normalizedPos[0] * aspectRatio / effectiveZoom) + c.Position[0],
		(normalizedPos[1] / effectiveZoom) + c.Position[1],
	}

	return worldPos
}

// RadiusToScreen converts a radius (in world units) to pixels.
// It scales the radius by the camera zoom, then maps that
// normalized size into screen pixels.
func (c *Camera2D) RadiusToScreen(radius float32, screenWidth, screenHeight float32) float32 {
	// Calculate effective zoom (user zoom * proximity zoom + drag zoom)
	effectiveZoom := (c.Zoom + c.DragZoom) * c.ProximityZoom

	// 1) apply zoom in world space
	scaled := radius * effectiveZoom

	// 2) convert normalized radius to pixels
	// Use the height as our reference scale since Y is always [0,1]
	// This ensures consistent circle sizes regardless of aspect ratio
	return scaled * screenHeight
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

// SetProximityZoom sets the proximity zoom multiplier
func (c *Camera2D) SetProximityZoom(proximityZoom float32) {
	c.ProximityZoom = proximityZoom
}

// SetDragZoom sets the target drag zoom offset (negative values zoom out)
func (c *Camera2D) SetDragZoom(dragZoom float32) {
	// Clamp drag zoom to prevent excessive zoom out
	const maxDragZoomOut = float32(-1.0) // Limit zoom out to prevent going too far
	if dragZoom < maxDragZoomOut {
		dragZoom = maxDragZoomOut
	}
	c.TargetDragZoom = dragZoom
}

// SetDragZoomImmediate sets the drag zoom immediately without interpolation
func (c *Camera2D) SetDragZoomImmediate(dragZoom float32) {
	const maxDragZoomOut = float32(-1.0)
	if dragZoom < maxDragZoomOut {
		dragZoom = maxDragZoomOut
	}
	c.DragZoom = dragZoom
	c.TargetDragZoom = dragZoom
}

// GetEffectiveZoom returns the combined zoom (user zoom * proximity zoom + drag zoom)
func (c *Camera2D) GetEffectiveZoom() float32 {
	return (c.Zoom + c.DragZoom) * c.ProximityZoom
}

// AdjustZoom adjusts the target camera zoom by the given delta
func (c *Camera2D) AdjustZoom(delta float32) {
	newZoom := c.TargetZoom + delta
	c.SetZoom(newZoom)
}

func (c *Camera2D) GetTotalZoom() float32 {
	// Returns the total zoom including proximity and drag effects
	return (c.Zoom + c.DragZoom) * c.ProximityZoom
}

func (c *Camera2D) SetTargetZoom(zoom float32) {
	// Clamp zoom to a reasonable range
	c.TargetZoom = zoom
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
