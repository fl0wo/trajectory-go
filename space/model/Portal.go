package Models

import (
	"golang.org/x/image/math/f32"
	"math"
)

// Portal represents a wormhole/portal entity that teleports players between paired coordinates
type Portal struct {
	ID                 int      // Unique identifier for this portal
	PairID             int      // ID of the paired portal (portals with same PairID are linked)
	Position           f32.Vec2 // Position of the portal in normalized coordinates
	Rotation           float32  // Rotation angle in radians for portal orientation
	Width              float32  // Width of the capsule portal
	Height             float32  // Height of the capsule portal
	Mass               float32  // Mass for physics interactions (typically very small)
	IsActive           bool     // Whether the portal is currently active
	CooldownTimer      float32  // Cooldown timer to prevent rapid teleportation loops
	PlayerInside       bool     // Whether the player is currently inside this portal's radius
	WasPlayerInside    bool     // Whether the player was inside in the previous frame
	DistanceToPlayer   float32  // Current distance from portal center to player
	MinDistanceReached float32  // Minimum distance reached while inside portal
	HasReachedCenter   bool     // Whether player has reached close enough to center
}

// GetPosition returns the portal's position (implementing a common interface pattern)
func (p *Portal) GetPosition() f32.Vec2 {
	return p.Position
}

// GetRadius returns an effective radius for collision detection (average of width and height)
func (p *Portal) GetRadius() float32 {
	return (p.Width + p.Height)
}

// GetOrbitRadius returns zero as portals don't have gravitational orbits
func (p *Portal) GetOrbitRadius() float32 {
	return p.GetRadius() * 2.0
}

// GetMass returns the portal's mass for physics calculations
func (p *Portal) GetMass() float32 {
	return p.Mass
}

// GetType returns the portal type (we'll need to add this to CelestialBodyType)
func (p *Portal) GetType() CelestialBodyType {
	return CelestialBodyTypePortal
}

// CheckCollisionWithPlayer checks if a player should teleport through this portal
// Returns true only when player has reached close to center and is now moving away
func (p *Portal) CheckCollisionWithPlayer(player *Player) bool {
	if !p.IsActive || p.CooldownTimer > 0 {
		return false
	}

	// Player must be inside the portal radius
	if !p.PlayerInside {
		return false
	}

	// Player must have reached close enough to the center
	if !p.HasReachedCenter {
		return false
	}

	// Teleport when player is moving away from center after reaching it
	// Simple check: if current distance is greater than minimum reached distance + small threshold
	threshold := p.GetRadius() * 0.1 // 10% of portal radius
	isMovingAway := p.DistanceToPlayer > (p.MinDistanceReached + threshold)

	return isMovingAway
}

// Update updates the portal's state (mainly cooldown timer)
func (p *Portal) Update(deltaTime float32) {
	if p.CooldownTimer > 0 {
		p.CooldownTimer -= deltaTime
		if p.CooldownTimer < 0 {
			p.CooldownTimer = 0
		}
	}
}

// UpdatePlayerInsideState updates whether the player is currently inside this portal
func (p *Portal) UpdatePlayerInsideState(player *Player) {
	// Store previous state
	p.WasPlayerInside = p.PlayerInside

	// Get player position and radius
	playerPos := player.Position
	playerRadius := player.Radius

	// Calculate distance from portal center to player center
	dx := playerPos[0] - p.Position[0]
	dy := playerPos[1] - p.Position[1]
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	p.DistanceToPlayer = distance

	// Use consistent radius calculation with GetRadius method
	portalRadius := p.GetRadius()
	combinedRadius := playerRadius + portalRadius

	isCurrentlyInside := distance <= combinedRadius

	// Update the current inside state
	p.PlayerInside = isCurrentlyInside

	// Track minimum distance reached while inside portal
	if p.PlayerInside {
		if !p.WasPlayerInside {
			// Just entered portal - reset tracking
			p.MinDistanceReached = distance
			p.HasReachedCenter = false
		} else {
			// Update minimum distance
			if distance < p.MinDistanceReached {
				p.MinDistanceReached = distance
			}

			// Helping the user a little bit entering the portal
			centerThreshold := portalRadius * 1.2
			if distance <= centerThreshold {
				p.HasReachedCenter = true
			}
		}
	} else if p.WasPlayerInside {
		// Just exited portal - reset tracking
		p.HasReachedCenter = false
		p.MinDistanceReached = math.MaxFloat32
	}
}

// StartCooldown starts the cooldown timer to prevent immediate re-entry
func (p *Portal) StartCooldown(duration float32) {
	p.CooldownTimer = duration
}

// IsReady returns true if the portal is active and not on cooldown
func (p *Portal) IsReady() bool {
	return p.IsActive && p.CooldownTimer <= 0
}

// NewPortal creates a new portal with default values
func NewPortal(id, pairID int, position f32.Vec2, rotation, width, height float32) *Portal {
	return &Portal{
		ID:                 id,
		PairID:             pairID,
		Position:           position,
		Rotation:           rotation,
		Width:              width,
		Height:             height,
		Mass:               0.01, // Very small mass to avoid affecting physics significantly
		IsActive:           true,
		CooldownTimer:      0.0,
		PlayerInside:       false,           // Player starts outside the portal
		WasPlayerInside:    false,           // Player was also outside in the previous frame
		DistanceToPlayer:   math.MaxFloat32, // Initialize with large distance
		MinDistanceReached: math.MaxFloat32, // Initialize with large distance
		HasReachedCenter:   false,           // Player hasn't reached center yet
	}
}

// NewPortalPair creates a pair of linked portals
func NewPortalPair(pairID int, pos1, pos2 f32.Vec2, rotation1, rotation2, width, height float32) (*Portal, *Portal) {
	portal1 := NewPortal(pairID*2, pairID, pos1, rotation1, width, height)
	portal2 := NewPortal(pairID*2+1, pairID, pos2, rotation2, width, height)
	return portal1, portal2
}
