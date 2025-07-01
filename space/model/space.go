package Models

import (
	"github.com/you/trajectory/constants"
	"github.com/you/trajectory/space/resources"
	"golang.org/x/image/math/f32"
	"image/color"
	"math"
	"math/rand"
)

// CameraMode defines how the camera follows entities
type CameraMode int

const (
	CameraModeCenter CameraMode = iota // Follow center of all entities (default)
	CameraModePlayer                   // Follow player only when moving
)

const (
	MaxCollisionHistory = 3    // Maximum number of collision records to keep
	ShakeDuration       = 1.0  // Duration of shake effect in seconds
	MarkerDuration      = 30.0 // Duration of collision marker visibility in seconds
)

type CollisionRecord struct {
	Position       f32.Vec2 // Position where collision occurred
	BodyIdx        int      // Index of the celestial body that was hit
	ShakeTimer     float32  // Timer for shake effect (0 = no shake)
	ShakeIntensity float32  // Current shake intensity
	MarkerTimer    float32  // Timer for collision marker visibility (independent of shake)
}

type SpaceGame struct {
	CelestialBodies     []CelestialBody
	RingAsteroids       []*RingAsteroid // Separate list for asteroids that need updates
	Portals             []*Portal       // List of portals for teleportation
	Player              *Player
	Camera              *Camera2D
	CurrentLevel        *Level
	CurrentLevelNum     int
	CameraMode          CameraMode        // Camera follow mode setting
	TimeScale           float32           // Current time dilation scale (1.0 = normal, 0.1 = 10x slower)
	TargetTimeScale     float32           // Target time scale for smooth interpolation
	ProximityZoom       float32           // Current proximity zoom multiplier (1.0 = normal, 1.25 = zoomed in)
	TargetProximityZoom float32           // Target proximity zoom for smooth interpolation
	ShadowsEnabled      bool              // Toggle for shadow rendering system
	CollisionHistory    []CollisionRecord // Array of recent collision records
}

type Border struct {
	BottomLeft, BottomRight, TopLeft, TopRight f32.Vec2
}

// NewSpaceGame creates a new SpaceGame starting with level 1.
func NewSpaceGame() (*SpaceGame, error) {
	return NewSpaceGameWithLevel(1)
}

// NewSpaceGameWithLevel creates a new SpaceGame with a specific level
func NewSpaceGameWithLevel(levelNum int) (*SpaceGame, error) {
	level := GetLevel(levelNum)

	var player = &Player{
		Position:     level.PlayerStart,
		Radius:       0.02,           // Player radius for collision detection
		Velocity:     f32.Vec2{0, 0}, // No initial velocity
		Acceleration: f32.Vec2{0, 0}, // No initial acceleration
		State:        PlayerStateIdle,
		Mass:         8.0, // Default mass
	}

	// Create camera
	camera := NewCamera2D()

	// Create temporary SpaceGame to calculate level center
	tempGame := &SpaceGame{
		CelestialBodies:     level.CelestialBodies,
		RingAsteroids:       level.RingAsteroids,
		Portals:             level.Portals,
		Player:              player,
		Camera:              camera,
		CurrentLevel:        level,
		CurrentLevelNum:     levelNum,
		CameraMode:          CameraModeCenter,                                // Default to center mode
		TimeScale:           1.0,                                             // Normal time initially
		TargetTimeScale:     1.0,                                             // Target matches current initially
		ProximityZoom:       1.0,                                             // Normal zoom initially
		TargetProximityZoom: 1.0,                                             // Target matches current initially
		ShadowsEnabled:      true,                                            // Enable shadows by default
		CollisionHistory:    make([]CollisionRecord, 0, MaxCollisionHistory), // Initialize empty collision history
	}

	// Calculate center of all entities and set camera position
	levelCenter := tempGame.CalculateLevelCenter()
	camera.Position = levelCenter
	camera.SetTarget(levelCenter)

	return tempGame, nil
}

// Reset resets the current level
func (sg *SpaceGame) Reset() error {
	return sg.ResetWithCollision(f32.Vec2{0, 0}, -1)
}

// ResetWithCollision resets the level with collision state information
func (sg *SpaceGame) ResetWithCollision(collisionPos f32.Vec2, bodyIdx int) error {
	// Add collision to history if valid
	if bodyIdx >= 0 && bodyIdx < len(sg.CelestialBodies) {
		newCollision := CollisionRecord{
			Position:       collisionPos,
			BodyIdx:        bodyIdx,
			ShakeTimer:     ShakeDuration,  // Duration of shake effect
			ShakeIntensity: 0.0003,         // Shake intensity
			MarkerTimer:    MarkerDuration, // Duration of marker visibility
		}

		// Add to front of history (most recent first)
		sg.CollisionHistory = append([]CollisionRecord{newCollision}, sg.CollisionHistory...)

		// Keep only the last MaxCollisionHistory records
		if len(sg.CollisionHistory) > MaxCollisionHistory {
			sg.CollisionHistory = sg.CollisionHistory[:MaxCollisionHistory]
		}
	}

	return sg.LoadLevel(sg.CurrentLevelNum)
}

// LoadLevel loads a specific level
func (sg *SpaceGame) LoadLevel(levelNum int) error {
	level := GetLevel(levelNum)

	// Reset player to level's starting position
	sg.Player.Position = level.PlayerStart
	sg.Player.Velocity = f32.Vec2{0, 0}
	sg.Player.Acceleration = f32.Vec2{0, 0}
	sg.Player.State = PlayerStateIdle
	sg.Player.ClearTrail() // Clear the movement trail

	// Clear collision history only when switching to a different level
	if levelNum != sg.CurrentLevelNum {
		sg.CollisionHistory = sg.CollisionHistory[:0] // Clear collision history
	}

	// Update level data first
	sg.CurrentLevel = level
	sg.CurrentLevelNum = levelNum
	sg.CelestialBodies = level.CelestialBodies
	sg.RingAsteroids = level.RingAsteroids
	sg.Portals = level.Portals

	// Calculate center of all entities and reset camera to that center
	levelCenter := sg.CalculateLevelCenter()
	sg.Camera.Position = levelCenter
	sg.Camera.SetTarget(levelCenter)

	return nil
}

func (sg *SpaceGame) NumNonAsteroids() int {
	count := 0
	for _, b := range sg.CelestialBodies {
		if b.GetType() != CelestialBodyTypeAsteroid {
			count++
		}
	}
	return count
}

// CalculateBorders computes the current game border corners.
//
//	1 celestial body: square 50% bigger than its orbit radius
//	2 celestial bodies: rectangle using biggest orbit as one side, distance between orbits as other side, + 50% padding
//	3+ celestial bodies: build their AABB and pad it by 50%
//	0 celestial bodies: fall back to a fixed box: x∈[-0.5,1.5], y∈[-0.5,1.5]
func (sg *SpaceGame) CalculateBorders() Border {
	numBodies := sg.NumNonAsteroids()

	const extraBorderPadding float32 = 0.25 // 10% padding on each side

	switch numBodies {
	case 0:
		// Default fallback rect
		return Border{
			BottomLeft:  f32.Vec2{-0.5, -0.5},
			BottomRight: f32.Vec2{1.5, -0.5},
			TopLeft:     f32.Vec2{-0.5, 1.5},
			TopRight:    f32.Vec2{1.5, 1.5},
		}

	case 1:
		// Single body: square 50% bigger than its orbit radius
		body := sg.CelestialBodies[0]
		pos := body.GetPosition()
		orbitRadius := body.GetOrbitRadius()

		// Create square centered on the body, size = orbit diameter + 50%
		squareHalfSize := orbitRadius * (1.0 + extraBorderPadding) // 50% bigger than orbit radius

		minX := float64(pos[0]) - float64(squareHalfSize)
		maxX := float64(pos[0]) + float64(squareHalfSize)
		minY := float64(pos[1]) - float64(squareHalfSize)
		maxY := float64(pos[1]) + float64(squareHalfSize)

		return Border{
			BottomLeft:  f32.Vec2{float32(minX), float32(minY)},
			BottomRight: f32.Vec2{float32(maxX), float32(minY)},
			TopLeft:     f32.Vec2{float32(minX), float32(maxY)},
			TopRight:    f32.Vec2{float32(maxX), float32(maxY)},
		}

	case 2:
		// Two bodies: rectangle with biggest orbit as one dimension, distance between orbits as other
		body1 := sg.CelestialBodies[0]
		body2 := sg.CelestialBodies[1]

		pos1 := body1.GetPosition()
		pos2 := body2.GetPosition()
		orbit1 := body1.GetOrbitRadius()
		orbit2 := body2.GetOrbitRadius()

		// find smallest & largest X and Y coordinates
		minX := math.Min(float64(pos1[0]-orbit1), float64(pos2[0]-orbit2))
		maxX := math.Max(float64(pos1[0]+orbit1), float64(pos2[0]+orbit2))
		minY := math.Min(float64(pos1[1]-orbit1), float64(pos2[1]-orbit2))
		maxY := math.Max(float64(pos1[1]+orbit1), float64(pos2[1]+orbit2))

		pad := float64(extraBorderPadding) // 10% padding on each side
		minX -= pad * (maxX - minX)
		maxX += pad * (maxX - minX)
		minY -= pad * (maxY - minY)
		maxY += pad * (maxY - minY)

		return Border{
			BottomLeft:  f32.Vec2{float32(minX), float32(minY)},
			BottomRight: f32.Vec2{float32(maxX), float32(minY)},
			TopLeft:     f32.Vec2{float32(minX), float32(maxY)},
			TopRight:    f32.Vec2{float32(maxX), float32(maxY)},
		}

	default:
		// 3+ bodies: build AABB over bodies and pad by 50%
		p0 := sg.CelestialBodies[0].GetPosition()
		minX, maxX := p0[0], p0[0]
		minY, maxY := p0[1], p0[1]

		for _, b := range sg.CelestialBodies {
			p := b.GetPosition()
			r := b.GetOrbitRadius()
			minX = min(minX, p[0]-r)
			maxX = max(maxX, p[0]+r)
			minY = min(minY, p[1]-r)
			maxY = max(maxY, p[1]+r)
		}

		// Pad by 10% of the width and height
		halfW := (maxX - minX) * extraBorderPadding
		halfH := (maxY - minY) * extraBorderPadding
		minX -= halfW
		maxX += halfW
		minY -= halfH
		maxY += halfH

		return Border{
			BottomLeft:  f32.Vec2{float32(minX), float32(minY)},
			BottomRight: f32.Vec2{float32(maxX), float32(minY)},
			TopLeft:     f32.Vec2{float32(minX), float32(maxY)},
			TopRight:    f32.Vec2{float32(maxX), float32(maxY)},
		}
	}
}

// DistanceToBorder returns signed distance of pos to the nearest border edge.
//   - Inside the rect → negative, = −min(distance to any edge).
//   - Outside the rect → positive, = how far past the closest edge you are.
func DistanceToBorder(pos f32.Vec2, b Border) float32 {
	minX := b.BottomLeft[0]
	maxX := b.BottomRight[0]
	minY := b.BottomLeft[1]
	maxY := b.TopLeft[1]

	// compute how far outside on each axis
	var dxOut, dyOut float32
	if pos[0] < minX {
		dxOut = minX - pos[0]
	} else if pos[0] > maxX {
		dxOut = pos[0] - maxX
	}
	if pos[1] < minY {
		dyOut = minY - pos[1]
	} else if pos[1] > maxY {
		dyOut = pos[1] - maxY
	}

	// outside: take the larger of the two
	if dxOut > 0 || dyOut > 0 {
		return float32(math.Max(float64(dxOut), float64(dyOut)))
	}

	// inside: distance to nearest edge (negative)
	distToLeft := pos[0] - minX
	distToRight := maxX - pos[0]
	distToBottom := pos[1] - minY
	distToTop := maxY - pos[1]
	nearestEdge := float32(math.Min(
		math.Min(float64(distToLeft), float64(distToRight)),
		math.Min(float64(distToBottom), float64(distToTop)),
	))

	return -nearestEdge
}

// CalculateLevelCenter ties it all together:
//   - if no entities, stick to player
//   - if distanceToBorder > 0, player “lost” → return player
//   - else fall back to weighted centroid.
func (sg *SpaceGame) CalculateLevelCenter() f32.Vec2 {
	// no entities → just point at player
	if len(sg.CelestialBodies) == 0 && len(sg.RingAsteroids) == 0 {
		return sg.Player.Position
	}

	// build border & measure
	border := sg.CalculateBorders()
	d := DistanceToBorder(sg.Player.Position, border)
	if d > 0 {
		// outside → lost
		return sg.Player.Position
	}

	// weighted centroid as before
	var sumX, sumY, totalW float32
	const (
		playerW   = 0.25
		bodyW     = 1.0
		asteroidW = 0.1
	)

	sumX += sg.Player.Position[0] * playerW
	sumY += sg.Player.Position[1] * playerW
	totalW += playerW

	for _, b := range sg.CelestialBodies {
		p := b.GetPosition()
		sumX += p[0] * bodyW
		sumY += p[1] * bodyW
		totalW += bodyW
	}
	for _, a := range sg.RingAsteroids {
		p := a.GetPosition()
		sumX += p[0] * asteroidW
		sumY += p[1] * asteroidW
		totalW += asteroidW
	}

	if totalW == 0 {
		return sg.Player.Position
	}
	return f32.Vec2{sumX / totalW, sumY / totalW}
}

// CalculateTimeDilation calculates the time scale based on player proximity to celestial bodies,
// using an ease-in curve (power law) instead of a linear interpolation.
func (sg *SpaceGame) CalculateTimeDilation() float32 {
	if sg.Player.State != PlayerStateMoving {
		return 1.0 // Normal time when not moving
	}

	minTimeScale := float32(1.0)
	playerPos := sg.Player.Position

	// Curve parameters
	const minTimeScaleAtCenter = 0.010 // 10x slower at the body’s surface
	const exponent = 1.15              // >1 = sharper ease-in

	// Check proximity to all celestial bodies
	for _, body := range sg.CelestialBodies {
		bodyPos := body.GetPosition()
		dx := bodyPos[0] - playerPos[0]
		dy := bodyPos[1] - playerPos[1]
		distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		orbitRadius := body.GetOrbitRadius()
		bodyRadius := body.GetRadius()

		// Only apply time dilation within orbit radius
		if distance <= orbitRadius {
			// Avoid division by zero by clamping at body radius
			minDistance := bodyRadius
			if distance < minDistance {
				distance = minDistance
			}

			// Normalize distance: 0 at surface, 1 at orbit edge
			normalizedDistance := (distance - minDistance) / (orbitRadius - minDistance)

			// Apply ease-in curve via power law
			curve := float32(math.Pow(float64(normalizedDistance), exponent))

			// Map curve [0→1] into timeScale [minTimeScaleAtCenter→1]
			timeScale := minTimeScaleAtCenter + (1.0-minTimeScaleAtCenter)*curve

			// Use the smallest timeScale (maximum slowdown)
			if timeScale < minTimeScale {
				minTimeScale = timeScale
			}
		}
	}

	// Check proximity to all asteroids (they don't have orbit radius but can cause time dilation on collision)
	for _, asteroid := range sg.RingAsteroids {
		bodyPos := asteroid.GetPosition()
		dx := bodyPos[0] - playerPos[0]
		dy := bodyPos[1] - playerPos[1]
		distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		// Only apply time dilation when very close to asteroids (collision imminent)
		asteroidInfluenceRadius := asteroid.GetRadius() * 2.0 // Small influence zone
		if distance <= asteroidInfluenceRadius {
			// Use asteroid radius as both min and max distance for simple collision-based effect
			normalizedDistance := distance / asteroidInfluenceRadius

			// Apply ease-in curve
			curve := float32(math.Pow(float64(normalizedDistance), exponent))

			// Time scale: more dramatic for close asteroid encounters
			timeScale := minTimeScaleAtCenter + (1.0-minTimeScaleAtCenter)*curve

			if timeScale < minTimeScale {
				minTimeScale = timeScale
			}
		}
	}

	return minTimeScale
}

// UpdateTimeDilation updates the current time scale with smooth interpolation
func (sg *SpaceGame) UpdateTimeDilation(deltaTime float32) {
	// Calculate target time scale
	sg.TargetTimeScale = sg.CalculateTimeDilation()

	// Smooth interpolation towards target (faster when slowing down, slower when speeding up)
	timeDilationSpeed := float32(5.0) // Adjust this for faster/slower transitions
	if sg.TargetTimeScale < sg.TimeScale {
		// Slowing down - faster transition for dramatic effect
		timeDilationSpeed = 8.0
	}

	// Interpolate towards target
	t := timeDilationSpeed * deltaTime
	if t > 1.0 {
		t = 1.0 // Clamp to prevent overshooting
	}

	sg.TimeScale = sg.TimeScale + (sg.TargetTimeScale-sg.TimeScale)*t
}

// CalculateProximityZoom calculates a zoom‐out multiplier based on proximity.
// Closer → smaller than 1.0 (zoomed out), far → 1.0 (normal).
func (sg *SpaceGame) CalculateProximityZoom() float32 {
	if sg.Player.State != PlayerStateMoving {
		return 1.0 // Normal zoom when not moving
	}

	// Start at no zoom (1.0) and look for any closer bodies to pull it down.
	minZoomMultiplier := float32(1.0)
	playerPos := sg.Player.Position

	// Zoom parameters
	const minZoomAtCenter = 0.95 // 90% scale at surface (max zoom-out)
	const exponent = 1.05        // Same ease-in curve

	// Celestial bodies
	for _, body := range sg.CelestialBodies {
		bodyPos := body.GetPosition()
		dx := bodyPos[0] - playerPos[0]
		dy := bodyPos[1] - playerPos[1]
		distance := float32(math.Hypot(float64(dx), float64(dy)))

		orbitRadius := body.GetOrbitRadius()
		bodyRadius := body.GetRadius()

		if distance <= orbitRadius {
			// Clamp to avoid zero
			if distance < bodyRadius {
				distance = bodyRadius
			}
			// 0 at surface → 1 at orbit edge
			normalized := (distance - bodyRadius) / (orbitRadius - bodyRadius)
			curve := float32(math.Pow(float64(normalized), exponent))

			// Map [0→1] → [minZoomAtCenter→1]
			zoomMul := minZoomAtCenter + (1.0-minZoomAtCenter)*curve

			if zoomMul < minZoomMultiplier {
				minZoomMultiplier = zoomMul
			}
		}
	}

	// Asteroids
	for _, ast := range sg.RingAsteroids {
		bodyPos := ast.GetPosition()
		dx := bodyPos[0] - playerPos[0]
		dy := bodyPos[1] - playerPos[1]
		distance := float32(math.Hypot(float64(dx), float64(dy)))

		influence := ast.GetRadius() * 2.0
		if distance <= influence {
			normalized := distance / influence
			curve := float32(math.Pow(float64(normalized), exponent))
			zoomMul := minZoomAtCenter + (1.0-minZoomAtCenter)*curve

			if zoomMul < minZoomMultiplier {
				minZoomMultiplier = zoomMul
			}
		}
	}

	return minZoomMultiplier
}

// UpdateProximityZoom smoothly interpolates ProximityZoom → TargetProximityZoom
func (sg *SpaceGame) UpdateProximityZoom(deltaTime float32) {
	sg.TargetProximityZoom = sg.CalculateProximityZoom()

	// Base speed; speed up when zooming out (i.e. target < current) if you like
	zoomSpeed := float32(5.0)
	if sg.TargetProximityZoom < sg.ProximityZoom {
		zoomSpeed = 8.0
	}

	t := zoomSpeed * deltaTime
	if t > 1.0 {
		t = 1.0
	}

	sg.ProximityZoom += (sg.TargetProximityZoom - sg.ProximityZoom) * t
}

// UpdateAsteroids updates all ring asteroids' orbital positions
func (sg *SpaceGame) UpdateAsteroids(deltaTime float32) {
	for _, asteroid := range sg.RingAsteroids {
		asteroid.Update(deltaTime)
	}
}

// UpdatePortals updates all portals' states (cooldown timers and player inside tracking)
func (sg *SpaceGame) UpdatePortals(deltaTime float32) {
	for _, portal := range sg.Portals {
		portal.Update(deltaTime)

		// Update player inside state for each portal
		portal.UpdatePlayerInsideState(sg.Player)
	}

	// Update camera target based on portal proximity
	sg.UpdateCameraTargetForPortals(deltaTime)
}

// UpdateCameraTargetForPortals sets camera target to center between portal pairs when player is inside
func (sg *SpaceGame) UpdateCameraTargetForPortals(deltaTime float32) {
	// Check if player is inside any portal
	for _, portal := range sg.Portals {
		if portal.PlayerInside {
			// Find the paired portal
			var pairedPortal *Portal
			for _, otherPortal := range sg.Portals {
				if otherPortal.PairID == portal.PairID && otherPortal.ID != portal.ID {
					pairedPortal = otherPortal
					break
				}
			}

			if pairedPortal != nil {
				// Calculate center position between the two portals
				centerX := (portal.Position[0] + pairedPortal.Position[0]) / 2.0
				centerY := (portal.Position[1] + pairedPortal.Position[1]) / 2.0
				centerPos := f32.Vec2{centerX, centerY}

				// Set camera target to the center between portals
				sg.Camera.SetTarget(centerPos)
				return // Only need to handle one portal pair
			}
		}
	}

	// If player is not inside any portal, check if they recently exited a portal
	// In this case, smoothly transition to following the player
	// BUT: Don't consider it "recently exited" if portal is on cooldown (means teleportation just occurred)
	playerRecentlyExitedPortal := false
	for _, portal := range sg.Portals {
		if !portal.PlayerInside && portal.WasPlayerInside && portal.CooldownTimer <= 0 {
			playerRecentlyExitedPortal = true
			break
		}
	}

	// If player just exited a portal, smoothly transition camera target to player position
	if playerRecentlyExitedPortal {
		sg.Camera.SetTargetSmooth(sg.Player.Position, deltaTime)
	}
}

// CheckPortalCollisions checks for player collisions with portals and handles teleportation
func (sg *SpaceGame) CheckPortalCollisions() {
	if sg.Player.State != PlayerStateMoving {
		return
	}

	for _, portal := range sg.Portals {
		if portal.CheckCollisionWithPlayer(sg.Player) {
			sg.TeleportPlayer(portal)
			break // Only teleport through one portal per frame
		}
	}
}

// TeleportPlayer teleports the player through a portal while preserving momentum
func (sg *SpaceGame) TeleportPlayer(sourcePortal *Portal) {
	// Find the paired portal
	var targetPortal *Portal
	for _, portal := range sg.Portals {
		if portal.PairID == sourcePortal.PairID && portal.ID != sourcePortal.ID {
			targetPortal = portal
			break
		}
	}

	if targetPortal == nil || !targetPortal.IsReady() {
		return // No valid target portal found
	}

	// Calculate rotation difference between portals
	rotationDiff := targetPortal.Rotation - sourcePortal.Rotation

	// Store current velocity and acceleration
	currentVelocity := sg.Player.Velocity
	currentAcceleration := sg.Player.Acceleration

	// Apply rotation transformation to velocity if there's a rotation difference
	if rotationDiff != 0 {
		cos := float32(math.Cos(float64(rotationDiff)))
		sin := float32(math.Sin(float64(rotationDiff)))

		// Rotate velocity vector
		newVelX := currentVelocity[0]*cos - currentVelocity[1]*sin
		newVelY := currentVelocity[0]*sin + currentVelocity[1]*cos
		currentVelocity = f32.Vec2{newVelX, newVelY}

		// Rotate acceleration vector
		newAccelX := currentAcceleration[0]*cos - currentAcceleration[1]*sin
		newAccelY := currentAcceleration[0]*sin + currentAcceleration[1]*cos
		currentAcceleration = f32.Vec2{newAccelX, newAccelY}
	}

	// Calculate offset from source portal center to ensure player exits target portal properly
	playerOffsetX := sg.Player.Position[0] - sourcePortal.Position[0]
	playerOffsetY := sg.Player.Position[1] - sourcePortal.Position[1]

	// Apply rotation to the offset if needed
	if rotationDiff != 0 {
		cos := float32(math.Cos(float64(rotationDiff)))
		sin := float32(math.Sin(float64(rotationDiff)))

		newOffsetX := playerOffsetX*cos - playerOffsetY*sin
		newOffsetY := playerOffsetX*sin + playerOffsetY*cos
		playerOffsetX = newOffsetX
		playerOffsetY = newOffsetY
	}

	// Teleport player to target portal position with offset
	sg.Player.Position = f32.Vec2{
		targetPortal.Position[0] + playerOffsetX,
		targetPortal.Position[1] + playerOffsetY,
	}

	// Preserve momentum
	sg.Player.Velocity = currentVelocity
	sg.Player.Acceleration = currentAcceleration

	// Mark both portals as having the player inside to prevent immediate re-entry
	// Set both current and previous state to prevent re-triggering
	sourcePortal.PlayerInside = true
	sourcePortal.WasPlayerInside = true
	targetPortal.PlayerInside = true
	targetPortal.WasPlayerInside = true

	// Start cooldown on both portals to prevent immediate re-entry
	const portalCooldown = 2.0 // 2 seconds cooldown
	sourcePortal.StartCooldown(portalCooldown)
	targetPortal.StartCooldown(portalCooldown)
}

// FindPortalByID finds a portal by its ID
func (sg *SpaceGame) FindPortalByID(id int) *Portal {
	for _, portal := range sg.Portals {
		if portal.ID == id {
			return portal
		}
	}
	return nil
}

// FindPortalsByPairID finds all portals with the given pair ID
func (sg *SpaceGame) FindPortalsByPairID(pairID int) []*Portal {
	var portals []*Portal
	for _, portal := range sg.Portals {
		if portal.PairID == pairID {
			portals = append(portals, portal)
		}
	}
	return portals
}

// ToggleCameraMode switches between camera modes
func (sg *SpaceGame) ToggleCameraMode() {
	if sg.CameraMode == CameraModeCenter {
		sg.CameraMode = CameraModePlayer
	} else {
		sg.CameraMode = CameraModeCenter
	}
}

// ToggleShadows switches shadow rendering on/off
func (sg *SpaceGame) ToggleShadows() {
	sg.ShadowsEnabled = !sg.ShadowsEnabled
}

// UpdateShake updates the shake effect timer and marker timer for all collisions
func (sg *SpaceGame) UpdateShake(deltaTime float32) {
	// Update all collision records and remove expired ones (based on marker timer)
	activeCollisions := make([]CollisionRecord, 0, len(sg.CollisionHistory))

	for _, collision := range sg.CollisionHistory {
		// Update both timers
		collision.ShakeTimer -= deltaTime
		collision.MarkerTimer -= deltaTime

		// Keep collision record as long as marker timer is active
		if collision.MarkerTimer > 0 {
			// Ensure shake timer doesn't go negative
			if collision.ShakeTimer < 0 {
				collision.ShakeTimer = 0
			}
			activeCollisions = append(activeCollisions, collision)
		}
	}

	sg.CollisionHistory = activeCollisions
}

// GetCameraModeString returns a string representation of the current camera mode
func (sg *SpaceGame) GetCameraModeString() string {
	switch sg.CameraMode {
	case CameraModeCenter:
		return "Center View"
	case CameraModePlayer:
		return "Player Follow"
	default:
		return "Unknown"
	}
}

func (sg *SpaceGame) GetWinPosition() f32.Vec2 {
	for _, body := range sg.CelestialBodies {
		if body.GetType() == CelestialBodyTypeWhiteHole {
			return body.GetPosition()
		}
	}
	return f32.Vec2{0, 0} // Default if no white hole found
}

func randomCelestialBodies(numBodies int, margin float32) []CelestialBody {
	var bodies = make([]CelestialBody, numBodies)

	// Use shared aspect ratio constant
	aspectRatio := constants.AspectRatio

	for i := 0; i < numBodies; i++ {
		// Generate random position and radius for each celestial body
		// X coordinate spans [0, aspectRatio], Y coordinate spans [0, 1]
		x := margin + (rand.Float32() * (aspectRatio - margin*2)) // value between margin and aspectRatio-margin
		y := margin + (rand.Float32() * (1 - margin*2))           // value between margin and 1-margin

		// radius in [0, 1] in Y-coordinate scale
		radius := rand.Float32() / 5.0 // smaller celestial bodies for better gameplay

		// for fun: make mass proportional to volume of a sphere
		mass := (4.0 / 3.0) * math.Pi * radius * radius * radius

		// Randomly decide if this should be a planet or blackhole (70% planet, 30% blackhole)
		if rand.Float32() < 0.7 {
			bodies[i] = &Planet{
				Name:        "Planet " + string(rune(i+1)),
				Position:    f32.Vec2{x, y},
				Mass:        mass, // scale mass for planets
				Radius:      radius,
				OrbitRadius: radius * 3.0, // orbit radius is 3x the planet radius
				BaseColor:   color.RGBA{R: uint8(180 + rand.Intn(75)), G: uint8(60 + rand.Intn(40)), B: uint8(10 + rand.Intn(30)), A: 255},
				Seed:        rand.Float32() * 100.0,
			}
		} else {
			// Blackholes are typically more massive and have larger orbit radius
			bodies[i] = &BlackHole{
				Position:    f32.Vec2{x, y},
				Mass:        mass,
				Radius:      radius,
				OrbitRadius: radius * 4.0, // orbit radius is 4x the blackhole radius
				ImagePath:   resources.BlackHoleImage,
			}
		}
	}
	return bodies
}

func randomPlanets(numPlanets int, margin float32) []Planet {
	var planets = make([]Planet, numPlanets)

	// Use shared aspect ratio constant
	aspectRatio := constants.AspectRatio

	for i := 0; i < numPlanets; i++ {
		// Generate random position and radius for each planet
		// X coordinate spans [0, aspectRatio], Y coordinate spans [0, 1]
		x := margin + (rand.Float32() * (aspectRatio - margin*2)) // value between margin and aspectRatio-margin
		y := margin + (rand.Float32() * (1 - margin*2))           // value between margin and 1-margin

		// radius in [0, 1] in Y-coordinate scale
		radius := rand.Float32() / 5.0 // smaller planets for better gameplay

		// for fun: make mass proportional to volume of a sphere
		mass := (4.0 / 3.0) * math.Pi * radius * radius * radius
		mass *= 250

		// mass := float32(math.Inf(1)) // infinite mass for simplicity

		planets[i] = Planet{
			Name:     "Planet " + string(rune(i+1)),
			Position: f32.Vec2{x, y},

			Mass: float32(mass),

			Radius:      radius,
			OrbitRadius: radius * 3.0, // orbit radius is 3x the planet radius
			BaseColor:   color.RGBA{R: uint8(180 + rand.Intn(75)), G: uint8(60 + rand.Intn(40)), B: uint8(10 + rand.Intn(30)), A: 255},
			Seed:        rand.Float32() * 100.0,
		}
	}
	return planets
}
